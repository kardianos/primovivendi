package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// A source is one input to the parser: a chapter or a reference sample. Its
// plain text is written for CoreNLP. For a chapter, paraLines maps the start
// offset (in UTF-16 units, as CoreNLP counts) of each paragraph to the
// paragraph's first line in the Markdown file.
type source struct {
	Name      string // "ch02" or "pg4351"
	File      string // chapter file name, for line references
	Label     string // what the report calls it
	Text      string
	ParaStart []int
	ParaLine  []int
	Ref       bool
}

var (
	reLink     = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	reFootRef  = regexp.MustCompile(`\[\^[^\]]*\]`)
	reHTML     = regexp.MustCompile(`<[^>]+>`)
	reListMark = regexp.MustCompile(`^\s*(?:[*+-]|\d+\.)\s+`)
	reBlank    = regexp.MustCompile(`\n[ \t]*\n`)
	reBracket  = regexp.MustCompile(`\[[^\]]*\]`)
	reSpaces   = regexp.MustCompile(`[ \t]+`)
)

// chapterSource keeps the chapter's own prose, quotations included, and drops
// headings, block quotations, tables, rules, and code. Markup is removed.
func chapterSource(path string) (source, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return source{}, err
	}
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".md")
	if i := strings.LastIndex(name, "_ch"); i >= 0 {
		n, _ := strconv.Atoi(name[i+3:])
		name = fmt.Sprintf("ch%02d", n)
	}
	type para struct {
		line int
		text []string
	}
	var paras []para
	var cur *para
	flush := func() {
		if cur != nil {
			paras = append(paras, *cur)
			cur = nil
		}
	}
	fence := false
	for i, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "```") {
			fence = !fence
			flush()
			continue
		}
		if fence {
			continue
		}
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, ">") || strings.HasPrefix(t, "|") ||
			strings.HasPrefix(t, "---") || strings.HasPrefix(t, "<!--") || strings.HasPrefix(t, ":::") {
			flush()
			continue
		}
		if reListMark.MatchString(line) {
			flush()
			line = reListMark.ReplaceAllString(line, "")
		}
		if cur == nil {
			cur = &para{line: i + 1}
		}
		cur.text = append(cur.text, strings.TrimSpace(line))
	}
	flush()
	rep := strings.NewReplacer("**", "", "*", "", "`", "", "\\", "")
	src := source{Name: name, File: base, Label: name}
	var sb strings.Builder
	units := 0
	for _, p := range paras {
		t := strings.Join(p.text, " ")
		t = reLink.ReplaceAllString(t, "$1")
		t = reFootRef.ReplaceAllString(t, "")
		t = reHTML.ReplaceAllString(t, " ")
		t = strings.TrimSpace(reSpaces.ReplaceAllString(rep.Replace(t), " "))
		if t == "" {
			continue
		}
		src.ParaStart = append(src.ParaStart, units)
		src.ParaLine = append(src.ParaLine, p.line)
		sb.WriteString(t)
		sb.WriteString("\n\n")
		units += utf16Len(t) + 2
	}
	src.Text = sb.String()
	return src, nil
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// lineAt returns the Markdown line of the paragraph holding offset off.
func (s source) lineAt(off int) int {
	i := sort.SearchInts(s.ParaStart, off+1) - 1
	if i < 0 || i >= len(s.ParaLine) {
		return 0
	}
	return s.ParaLine[i]
}

// paraIndex returns the index of the paragraph holding offset off.
func (s source) paraIndex(off int) int {
	return sort.SearchInts(s.ParaStart, off+1) - 1
}

// Reference samples: running prose from public-domain works, taken from 20
// percent into the text so that prefaces are skipped.

type refPick struct {
	ID    int
	Label string
}

var refPicks = []refPick{
	{4351, "Bagehot, The English Constitution (1867)"},
	{34901, "Mill, On Liberty (1859)"},
	{935, "Smiles, Self-Help (1859)"},
	{35916, "Thurston, The Growth of the Steam-Engine (1878)"},
	{4367, "Grant, Personal Memoirs (1885)"},
	{13529, "Mahan, The Influence of Sea Power upon History (1890)"},
	{26334, "Lecky, The Map of Life (1899)"},
	{6435, "Taylor, The Principles of Scientific Management (1911)"},
}

const refSampleWords = 6000

func refSource(data string, p refPick) (source, error) {
	b, err := os.ReadFile(filepath.Join(data, "ngram", "ref", fmt.Sprintf("pg%d.txt", p.ID)))
	if err != nil {
		return source{}, fmt.Errorf("%w (run civ/ngram: go run . fetch-ref)", err)
	}
	paras := gutenbergParagraphs(string(b))
	src := source{Name: fmt.Sprintf("pg%d", p.ID), Label: p.Label, Ref: true}
	var sb strings.Builder
	words := 0
	for _, t := range paras[len(paras)/5:] {
		t = strings.TrimSpace(reSpaces.ReplaceAllString(t, " "))
		sb.WriteString(t)
		sb.WriteString("\n\n")
		words += len(strings.Fields(t))
		if words >= refSampleWords {
			break
		}
	}
	src.Text = sb.String()
	return src, nil
}

func gutenbergParagraphs(txt string) []string {
	txt = strings.ReplaceAll(txt, "\r", "")
	if i := strings.Index(txt, "*** START OF"); i >= 0 {
		if j := strings.Index(txt[i:], "\n"); j >= 0 {
			txt = txt[i+j+1:]
		}
	}
	if i := strings.Index(txt, "*** END OF"); i >= 0 {
		txt = txt[:i]
	}
	var paras []string
	for _, block := range reBlank.Split(txt, -1) {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		for i := range lines {
			lines[i] = strings.TrimSpace(lines[i])
		}
		p := reBracket.ReplaceAllString(strings.Join(lines, " "), " ")
		if skipPara(p) {
			continue
		}
		paras = append(paras, strings.ReplaceAll(p, "_", ""))
	}
	return paras
}

func skipPara(p string) bool {
	letters, upper, digits := 0, 0, 0
	for _, r := range p {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
		if unicode.IsDigit(r) {
			digits++
		}
	}
	return letters < 40 || float64(upper) > 0.5*float64(letters) ||
		float64(digits) > 0.1*float64(letters) || !strings.ContainsAny(p, ".?!")
}

func loadSources(root, data string) ([]source, error) {
	files, err := filepath.Glob(filepath.Join(root, "civ", "chapters", "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []source
	for _, f := range files {
		if strings.HasPrefix(filepath.Base(f), "00_") {
			continue
		}
		s, err := chapterSource(f)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	for _, p := range refPicks {
		s, err := refSource(data, p)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
