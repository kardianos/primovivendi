// stage4.go — diagnostics for Stage 4 (coreference, Rule 10 Part 4): run
// civ/linear's Rule 10 analysis under both engines on development chapters
// and list every sentence where the Part 4 findings differ, with the parse
// and the coreference chains a fix has to look at. ch13 and ch14 are never
// passed here during development (cmdPart4 refuses them).
package parser

import (
	"fmt"
	"sort"
	"strings"

	"po/civ/sys/corenlp"
)

// Part4Diff is one sentence whose Part 4 findings differ between engines.
type Part4Diff struct {
	Chapter                 string
	Num                     int
	Text                    string
	CoreNLP, Go             []string // finding kinds
	OnlyCoreNLP, OnlyGo     []string
	GoTree                  string // "i word POS head rel ner" lines
	GoChains, CoreNLPChains []string
	Why                     string // what the consumer saw: the conditions behind the findings
}

// Part4Report summarizes the differences over the chapters.
type Part4Report struct {
	Sentences int
	Diffs     []Part4Diff
	Parity    *Rule10Parity
}

// Part4Diffs runs the Rule 10 parity on the chapters and explains every
// sentence whose Part 4 finding kinds differ. It refuses held-out chapters.
func Part4Diffs(root, binary string, chapters []string, p *Pipeline) (*Part4Report, error) {
	for _, c := range chapters {
		if IsHeldout(c) {
			return nil, fmt.Errorf("parser: %s is held out; Stage 4 is tuned on ch01..ch12 only", c)
		}
	}
	rp, err := Rule10(root, binary, chapters)
	if err != nil {
		return nil, err
	}
	cn, err := LinearScores(root, binary, strings.Join(chapters, ","), "corenlp")
	if err != nil {
		return nil, err
	}
	goS, err := LinearScores(root, binary, strings.Join(chapters, ","), "go")
	if err != nil {
		return nil, err
	}
	byKey := map[string]linearSentence{}
	for _, s := range goS {
		byKey[s.Chapter+"|"+s.Text] = s
	}
	// The CoreNLP documents (for their chains) and the Go parses, per chapter.
	var docs = map[string]*corenlp.Doc{}
	var mine = map[string]*corenlp.Doc{}
	for _, c := range chapters {
		d, text, err := corenlp.LoadCached(root, c)
		if err != nil {
			return nil, err
		}
		docs[c] = d
		if p != nil {
			mine[c] = p.ParseDoc(text)
		}
	}
	rep := &Part4Report{Parity: rp}
	for _, g := range cn {
		m, ok := byKey[g.Chapter+"|"+g.Text]
		if !ok {
			continue
		}
		rep.Sentences++
		gk, mk := kindSet(g), kindSet(m)
		if sameKinds(gk, mk) {
			continue
		}
		d := Part4Diff{Chapter: g.Chapter, Num: g.Num, Text: g.Text, CoreNLP: sortedKinds(gk), Go: sortedKinds(mk)}
		for k := range gk {
			if !mk[k] {
				d.OnlyCoreNLP = append(d.OnlyCoreNLP, k)
			}
		}
		for k := range mk {
			if !gk[k] {
				d.OnlyGo = append(d.OnlyGo, k)
			}
		}
		sort.Strings(d.OnlyCoreNLP)
		sort.Strings(d.OnlyGo)
		// The Go tree and chains of the sentence, by matching its first
		// and last token text in the Go document of the chapter.
		if md := mine[g.Chapter]; md != nil {
			si := findSentence(md, g.Text)
			if si >= 0 {
				s := md.Sentences[si]
				var b strings.Builder
				for i, t := range s.Toks {
					b.WriteString(fmt.Sprintf("%d %s %s %s %s %s\n", i+1, t.Orig, t.POS, t.NER, depLabel(s.Deps, i+1), depHead(s.Deps, i+1)))
				}
				d.GoTree = b.String()
				d.GoChains = chainsOf(md, si+1)
			}
		}
		if cd := docs[g.Chapter]; cd != nil {
			if si := findSentence(cd, g.Text); si >= 0 {
				d.CoreNLPChains = chainsOf(cd, si+1)
			}
		}
		rep.Diffs = append(rep.Diffs, d)
	}
	return rep, nil
}

func sameKinds(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func sortedKinds(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// findSentence locates a sentence in a document by its token texts.
func findSentence(doc *corenlp.Doc, text string) int {
	want := normText(text)
	for i, s := range doc.Sentences {
		words := make([]string, len(s.Toks))
		for k, t := range s.Toks {
			words[k] = t.Orig
		}
		if normText(strings.Join(words, " ")) == want {
			return i
		}
	}
	tf := strings.Fields(want)
	for i, s := range doc.Sentences { // fallback: the first three words
		if len(s.Toks) >= 3 && len(tf) >= 3 && strings.ToLower(s.Toks[0].Orig) == tf[0] && strings.ToLower(s.Toks[1].Orig) == tf[1] && strings.ToLower(s.Toks[2].Orig) == tf[2] {
			return i
		}
	}
	return -1
}

// normText lowercases and collapses the whitespace of a sentence.
func normText(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }

func depLabel(deps []corenlp.Dep, i int) string {
	for _, d := range deps {
		if d.Dependent == i {
			return d.Dep
		}
	}
	return "?"
}

func depHead(deps []corenlp.Dep, i int) string {
	for _, d := range deps {
		if d.Dependent == i {
			return fmt.Sprint(d.Governor)
		}
	}
	return "?"
}

// chainsOf renders the coreference chains that have a mention in sentence
// sentNum (1-based): each mention with its sentence, head word, and type.
func chainsOf(doc *corenlp.Doc, sentNum int) []string {
	keys := make([]string, 0, len(doc.Corefs))
	for k := range doc.Corefs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []string
	for _, k := range keys {
		ms := doc.Corefs[k]
		has := false
		for _, m := range ms {
			if m.SentNum == sentNum {
				has = true
			}
		}
		if !has {
			continue
		}
		var parts []string
		for _, m := range ms {
			word := ""
			if m.SentNum-1 < len(doc.Sentences) && m.Head >= 1 && m.Head <= len(doc.Sentences[m.SentNum-1].Toks) {
				word = doc.Sentences[m.SentNum-1].Toks[m.Head-1].Orig
			}
			parts = append(parts, fmt.Sprintf("s%d:%q(%s,%d-%d)", m.SentNum, word, m.Type, m.Start, m.End))
		}
		out = append(out, "chain "+k+": "+strings.Join(parts, " "))
	}
	return out
}
