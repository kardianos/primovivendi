package main

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Measures by section of a chapter. Distance in common-word habits is noisy
// for a stretch of a few hundred words, and it grows as the stretch gets
// shorter, so a section's Delta is judged against stretches of the
// reference works of the same length: the share of those stretches that lie
// nearer the period mean. Markup (bold type, lists, and cross-references to
// the principles and virtues) is counted as it stands in the Markdown.

// A bookSection is the text under one heading, up to the next heading of any
// level. The text between the chapter title and the first heading below it
// goes with the chapter title.
type bookSection struct {
	Chapter string // "ch2", as in loadBook
	Line    int    // line of the heading in the Markdown file
	Title   string
	Body    string // Markdown
}

func loadSections(root string) ([]bookSection, error) {
	files, err := filepath.Glob(filepath.Join(root, "civ", "chapters", "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []bookSection
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
		cur := bookSection{Chapter: name, Line: 1}
		var body []string
		fence := false
		for i, l := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(strings.TrimSpace(l), "```") {
				fence = !fence
			}
			if !fence && strings.HasPrefix(l, "#") {
				if cur.Title != "" || strings.TrimSpace(strings.Join(body, "")) != "" {
					cur.Body = strings.Join(body, "\n")
					out = append(out, cur)
				}
				cur = bookSection{Chapter: name, Line: i + 1, Title: strings.TrimSpace(strings.TrimLeft(l, "#"))}
				body = nil
				continue
			}
			body = append(body, l)
		}
		cur.Body = strings.Join(body, "\n")
		out = append(out, cur)
	}
	return out, nil
}

var (
	reBold     = regexp.MustCompile(`\*\*\*?([^*\n]+?)\*?\*\*`)
	reCrossRef = regexp.MustCompile(`\([^()]*\b(?:Principle|Virtue|Filter)s? \d[^()]*\)`)
	reIndent   = regexp.MustCompile(`^(?:  |\t)`)
)

// markupCount counts the words of a section as printed (quotations
// included), the words set in bold, the words in list items, and the
// parenthetical cross-references such as "(Principle 2)".
type markupCount struct{ Words, Bold, List, Tags int }

func countMarkup(md string) markupCount {
	var m markupCount
	strip := strings.NewReplacer("**", "", "*", "", "`", "", "\\", "")
	words := func(s string) int {
		s = reLink.ReplaceAllString(s, "$1")
		return tokenizeSentence(strip.Replace(s), newDoc("")).Words
	}
	fence, inList := false, false
	for _, line := range strings.Split(md, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "```") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		t, q := unquote(t)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "#") || strings.HasPrefix(t, "|") || strings.HasPrefix(t, "---") ||
			strings.HasPrefix(t, "<!--") || strings.HasPrefix(t, ":::") {
			inList = false
			continue
		}
		item := reListMark.MatchString(t)
		switch {
		case item:
			inList = true
		case !q && inList && reIndent.MatchString(line):
			item = true // a paragraph inside a list item
		default:
			inList = false
		}
		n := words(t)
		m.Words += n
		if item {
			m.List += n
		}
		for _, b := range reBold.FindAllStringSubmatch(t, -1) {
			m.Bold += words(b[1])
		}
		m.Tags += len(reCrossRef.FindAllString(t, -1))
	}
	return m
}

// periodWindows draws stretches of the reference works for comparing a text
// of a given length with the period.
type periodWindows struct {
	toks   [][]string    // each work's words
	loo    []deltaModel  // each work's model, built from the other works
	others [][][]float64 // for work i, the z-scores of the other works under loo[i]
	idx    map[string]int
	k      int
}

func newPeriodWindows(refs []*Doc, mfw []string) periodWindows {
	w := periodWindows{idx: map[string]int{}, k: len(mfw)}
	for j, x := range mfw {
		w.idx[x] = j
	}
	for i := range refs {
		rest := append(append([]*Doc{}, refs[:i]...), refs[i+1:]...)
		m := newDeltaModel(mfw, rest)
		var zs [][]float64
		for _, d := range rest {
			zs = append(zs, m.z(d))
		}
		w.toks = append(w.toks, tokens(refs[i]))
		w.loo = append(w.loo, m)
		w.others = append(w.others, zs)
	}
	return w
}

// windowsCloser returns the shares of n stretches of length words, drawn at
// random from the reference works, whose Delta and cosine Delta are smaller
// than delta and cos. A stretch is scored as the reference works were: its
// own work is left out of the mean and of the comparison. The draws depend
// only on the length, so a section's figures change only when it does.
func (w periodWindows) windowsCloser(length int, delta, cos float64, n int) (float64, float64) {
	rng := rand.New(rand.NewPCG(uint64(length), 1850))
	var fit []int
	for i, t := range w.toks {
		if len(t) >= length {
			fit = append(fit, i)
		}
	}
	if len(fit) == 0 || length == 0 {
		return 0, 0
	}
	nd, nc := 0, 0
	for range n {
		i := fit[rng.IntN(len(fit))]
		start := rng.IntN(len(w.toks[i]) - length + 1)
		z := w.loo[i].zFreq(relFreqTokens(w.toks[i][start:start+length], w.idx, w.k))
		if deltaCentroid(z) < delta {
			nd++
		}
		if nearestCosine(z, w.others[i]) < cos {
			nc++
		}
	}
	return float64(nd) / float64(n), float64(nc) / float64(n)
}
