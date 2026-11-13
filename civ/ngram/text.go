package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// A Sentence is a list of segments. A segment is a run of words with no
// punctuation between them. N-grams are taken inside a segment only. Google
// Books counts punctuation as a token, so a phrase that crosses a comma in
// our text would not match the Google phrase without the comma.
//
// A hyphen between two words stays in the segment as the token "-", which is
// how Google tokenizes "self-interest" ("self - interest"). An apostrophe
// ends the segment and its suffix is dropped ("man's life" gives "man" and
// "life"), because the Ngram Viewer cannot query possessives reliably.
type Sentence struct {
	Segs  [][]string
	Words int
	First string
}

// A Doc is one text: a chapter, the whole book, or one reference work.
type Doc struct {
	Name  string
	Sents []Sentence
	Words int
	Cap   map[string]int // capitalized occurrences after the first word of a sentence
	Low   map[string]int // lowercase occurrences
}

func newDoc(name string) *Doc {
	return &Doc{Name: name, Cap: map[string]int{}, Low: map[string]int{}}
}

func (d *Doc) add(o *Doc) {
	d.Sents = append(d.Sents, o.Sents...)
	d.Words += o.Words
	for k, v := range o.Cap {
		d.Cap[k] += v
	}
	for k, v := range o.Low {
		d.Low[k] += v
	}
}

// Counts returns n-gram counts. Unigrams skip the hyphen token. An n-gram
// may contain a hyphen token but may not begin or end with one, and a phrase
// of two or more words may not stop or start inside a hyphenated compound
// ("build is self" from "build is self-anchored").
func (d *Doc) Counts(n int) map[string]int {
	m := map[string]int{}
	for _, s := range d.Sents {
		for _, seg := range s.Segs {
			for i := 0; i+n <= len(seg); i++ {
				g := seg[i : i+n]
				if g[0] == "-" || g[n-1] == "-" {
					continue
				}
				if n > 1 && ((i > 0 && seg[i-1] == "-") || (i+n < len(seg) && seg[i+n] == "-")) {
					continue
				}
				m[strings.Join(g, " ")]++
			}
		}
	}
	return m
}

// ProperNoun reports whether a word is written capitalized in most of its
// mid-sentence occurrences.
func (d *Doc) ProperNoun(w string) bool {
	c, l := d.Cap[w], d.Low[w]
	return c > 0 && c >= 4*l
}

var abbrev = map[string]bool{
	"mr": true, "mrs": true, "dr": true, "st": true, "vs": true, "etc": true, "no": true,
	"vol": true, "ch": true, "pp": true, "art": true, "arts": true, "ss": true, "cf": true,
	"ed": true, "eds": true, "gen": true, "gov": true, "col": true, "capt": true, "lt": true,
	"rev": true, "hon": true, "jr": true, "sr": true, "co": true, "inc": true, "ltd": true,
	"viz": true, "messrs": true, "esq": true, "sec": true, "fig": true, "ibid": true,
	"op": true, "cit": true, "ft": true, "lb": true, "lbs": true, "oz": true, "sq": true,
}

// splitSentences splits a paragraph at ".", "?", or "!" followed by space and
// a capital letter, a digit, or an opening mark. A period after an
// abbreviation, an initial, or a dotted form such as "U.S." does not end a
// sentence.
func splitSentences(p string) []string {
	var out []string
	rs := []rune(p)
	start := 0
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		if c != '.' && c != '?' && c != '!' {
			continue
		}
		j := i + 1
		for j < len(rs) && strings.ContainsRune("\"'”’)]*_", rs[j]) {
			j++
		}
		if j < len(rs) && !unicode.IsSpace(rs[j]) {
			continue
		}
		k := j
		for k < len(rs) && unicode.IsSpace(rs[k]) {
			k++
		}
		if k < len(rs) {
			n := rs[k]
			if !(unicode.IsUpper(n) || unicode.IsDigit(n) || strings.ContainsRune("\"“'‘(*[_", n)) {
				continue
			}
		}
		if c == '.' {
			b := i
			for b > start && (unicode.IsLetter(rs[b-1]) || rs[b-1] == '.') {
				b--
			}
			w := strings.ToLower(string(rs[b:i]))
			if abbrev[w] || len([]rune(w)) == 1 || strings.Contains(w, ".") {
				continue
			}
		}
		out = append(out, string(rs[start:j]))
		start = j
	}
	if start < len(rs) {
		out = append(out, string(rs[start:]))
	}
	return out
}

func isApostrophe(r rune) bool { return r == '\'' || r == '’' }

func tokenizeSentence(s string, d *Doc) Sentence {
	var sent Sentence
	var seg []string
	flush := func() {
		if len(seg) > 0 {
			sent.Segs = append(sent.Segs, seg)
			seg = nil
		}
	}
	rs := []rune(s)
	first := true
	for i := 0; i < len(rs); {
		r := rs[i]
		switch {
		case unicode.IsLetter(r):
			j := i
			digit := false
			for j < len(rs) && (unicode.IsLetter(rs[j]) || unicode.IsDigit(rs[j])) {
				if unicode.IsDigit(rs[j]) {
					digit = true
				}
				j++
			}
			if digit {
				flush()
				first = false
				i = j
				continue
			}
			w := string(rs[i:j])
			lw := strings.ToLower(w)
			if !first {
				if unicode.IsUpper(rs[i]) {
					d.Cap[lw]++
				} else {
					d.Low[lw]++
				}
			}
			first = false
			seg = append(seg, lw)
			sent.Words++
			if sent.First == "" {
				sent.First = lw
			}
			if j+1 < len(rs) && isApostrophe(rs[j]) && unicode.IsLetter(rs[j+1]) {
				k := j + 1
				for k < len(rs) && unicode.IsLetter(rs[k]) {
					k++
				}
				flush()
				i = k
				continue
			}
			if j+1 < len(rs) && rs[j] == '-' && unicode.IsLetter(rs[j+1]) {
				seg = append(seg, "-")
				i = j + 1
				continue
			}
			i = j
		case unicode.IsDigit(r):
			j := i
			for j < len(rs) && (unicode.IsLetter(rs[j]) || unicode.IsDigit(rs[j]) || rs[j] == '.' || rs[j] == ',') {
				j++
			}
			flush()
			first = false
			i = j
		case unicode.IsSpace(r):
			i++
		default:
			flush()
			i++
		}
	}
	flush()
	return sent
}

func buildDoc(name string, paras []string, maxWords int) *Doc {
	d := newDoc(name)
	for _, p := range paras {
		for _, s := range splitSentences(p) {
			st := tokenizeSentence(s, d)
			if st.Words == 0 {
				continue
			}
			d.Sents = append(d.Sents, st)
			d.Words += st.Words
			if maxWords > 0 && d.Words >= maxWords {
				return d
			}
		}
	}
	return d
}

var (
	reItalicQuote = regexp.MustCompile(`\*+["“][^"”\n]*["”]\*+`)
	reLongQuote   = regexp.MustCompile(`["“][^"”\n]{80,}["”]`)
	reLink        = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	reFootRef     = regexp.MustCompile(`\[\^[^\]]*\]`)
	reHTML        = regexp.MustCompile(`<[^>]+>`)
	reListMark    = regexp.MustCompile(`^\s*(?:[*+-]|\d+\.)\s+`)
	reBlank       = regexp.MustCompile(`\n[ \t]*\n`)
	reBracket     = regexp.MustCompile(`\[[^\]]*\]`)
)

// bookParagraphs turns a chapter's Markdown into plain paragraphs. It drops
// headings, block quotations, tables, rules, and code, and it replaces
// quotations set as *"..."* (and any quoted span of 80 characters or more)
// with a break, so that quoted sources are not counted as the book's prose.
func bookParagraphs(md string) []string {
	var paras, cur []string
	fence := false
	flush := func() {
		if len(cur) > 0 {
			paras = append(paras, strings.Join(cur, " "))
			cur = nil
		}
	}
	for _, line := range strings.Split(md, "\n") {
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
		cur = append(cur, strings.TrimSpace(line))
	}
	flush()
	rep := strings.NewReplacer("**", "", "*", "", "`", "", "\\", "")
	for i, p := range paras {
		p = reItalicQuote.ReplaceAllString(p, " ; ")
		p = reLongQuote.ReplaceAllString(p, " ; ")
		p = reLink.ReplaceAllString(p, "$1")
		p = reFootRef.ReplaceAllString(p, "")
		p = reHTML.ReplaceAllString(p, " ")
		paras[i] = rep.Replace(p)
	}
	return paras
}

// gutenbergParagraphs strips the Project Gutenberg header and footer and
// keeps paragraphs of running prose. It drops headings, short lines,
// all-capital lines, tables, index entries, and bracketed notes.
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
		p := strings.Join(lines, " ")
		p = reBracket.ReplaceAllString(p, " ")
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
	switch {
	case letters < 40:
		return true
	case float64(upper) > 0.5*float64(letters):
		return true
	case float64(digits) > 0.1*float64(letters):
		return true
	case !strings.ContainsAny(p, ".?!"):
		return true
	}
	return false
}

// loadBook reads civ/chapters/*.md except the title page. It returns one Doc
// per chapter, in file order.
func loadBook(root string) ([]*Doc, error) {
	files, err := filepath.Glob(filepath.Join(root, "civ", "chapters", "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var docs []*Doc
	for _, f := range files {
		base := filepath.Base(f)
		if strings.HasPrefix(base, "00_") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(base, ".md")
		if i := strings.LastIndex(name, "_ch"); i >= 0 {
			name = "ch" + strings.TrimLeft(name[i+3:], "0")
		}
		docs = append(docs, buildDoc(name, bookParagraphs(string(b)), 0))
	}
	return docs, nil
}

func mergeDocs(name string, docs []*Doc) *Doc {
	all := newDoc(name)
	for _, d := range docs {
		all.add(d)
	}
	return all
}
