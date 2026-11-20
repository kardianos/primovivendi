package main

import (
	"strings"
	"testing"
)

// tw is one token of a synthetic CoreNLP sentence. head is the 1-based index
// of its governor, 0 for the root. rel of the root may be anything.
type tw struct {
	w, lem, pos, rel string
	head             int
}

func buildDoc(paragraphs ...[][]tw) (*source, *cnDoc, string) {
	var b strings.Builder
	doc := &cnDoc{Corefs: map[string][]cnMention{}}
	for pi, para := range paragraphs {
		if pi > 0 {
			b.WriteString("\n\n")
		}
		for si, words := range para {
			if si > 0 {
				b.WriteByte(' ')
			}
			var toks []cnTok
			var deps []cnDep
			for i, w := range words {
				if i > 0 {
					b.WriteByte(' ')
				}
				start := b.Len()
				b.WriteString(w.w)
				toks = append(toks, cnTok{
					Index: i + 1, Word: w.w, Orig: w.w, Lemma: w.lem,
					Begin: start, End: b.Len(), POS: w.pos,
				})
				deps = append(deps, cnDep{Dep: w.rel, Governor: w.head, Dependent: i + 1, DepGloss: w.w})
			}
			doc.Sentences = append(doc.Sentences, cnSent{Index: len(doc.Sentences), Toks: toks, Deps: deps})
		}
	}
	text := b.String()
	src := &source{Name: "t", Label: "t", Text: text, Ref: true}
	return src, doc, text
}

func kindCount(ss []*sentence, kind string) int {
	n := 0
	for _, s := range ss {
		for _, f := range s.Findings {
			if f.Kind == kind {
				n++
			}
		}
	}
	return n
}

func findingNotes(ss []*sentence, kind string) string {
	var b strings.Builder
	for _, s := range ss {
		for _, f := range s.Findings {
			if f.Kind == kind {
				b.WriteString(f.Note)
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

func dumpFindings(ss []*sentence) string {
	var b strings.Builder
	for _, s := range ss {
		b.WriteString(s.Text)
		b.WriteByte('\n')
		for _, f := range s.Findings {
			b.WriteString("  ")
			b.WriteString(f.Kind)
			b.WriteString(": ")
			b.WriteString(f.Note)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func TestItChainAndAbstractSubject(t *testing.T) {
	src, doc, text := theologyPair()
	ss := analyzeDoc(src, doc, text, map[string]float64{"theology": 1.93, "faith": 1.63})
	if kindCount(ss, "it chain") == 0 {
		t.Fatalf("it chain did not fire\n%s", dumpFindings(ss))
	}
	if kindCount(ss, "abstract subject") == 0 || !strings.Contains(findingNotes(ss, "abstract subject"), "seek") {
		t.Fatalf("abstract subject did not fire on seek\n%s", dumpFindings(ss))
	}
	for _, s := range ss {
		sum := 0.0
		for _, f := range s.Findings {
			if f.Part > 0 {
				sum += f.Weight
			}
		}
		if s.Score != sum {
			t.Fatalf("score %.1f counted a non-Rule-10 finding\n%s", s.Score, dumpFindings(ss))
		}
	}
}

func theologyPair() (*source, *cnDoc, string) {
	return buildDoc([][]tw{{
		{w: "Yet", lem: "yet", pos: "RB", rel: "advmod", head: 6},
		{w: "it", lem: "it", pos: "PRP", rel: "nsubj", head: 6},
		{w: "is", lem: "be", pos: "VBZ", rel: "cop", head: 6},
		{w: "not", lem: "not", pos: "RB", rel: "advmod", head: 6},
		{w: "a", lem: "a", pos: "DT", rel: "det", head: 6},
		{w: "theology", lem: "theology", pos: "NN", rel: "root", head: 0},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 6},
	}, {
		{w: "It", lem: "it", pos: "PRP", rel: "nsubj", head: 4},
		{w: "does", lem: "do", pos: "VBZ", rel: "aux", head: 4},
		{w: "not", lem: "not", pos: "RB", rel: "advmod", head: 4},
		{w: "seek", lem: "seek", pos: "VB", rel: "root", head: 0},
		{w: "to", lem: "to", pos: "TO", rel: "mark", head: 6},
		{w: "diminish", lem: "diminish", pos: "VB", rel: "xcomp", head: 4},
		{w: "any", lem: "any", pos: "DT", rel: "det", head: 8},
		{w: "faith", lem: "faith", pos: "NN", rel: "obj", head: 6},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 4},
	}})
}

func TestCloseEchoSet(t *testing.T) {
	// "Chapter 9 sets out the rule. This book does not set out to diminish any faith."
	src, doc, text := buildDoc([][]tw{{
		{w: "Chapter", lem: "chapter", pos: "NNP", rel: "compound", head: 3},
		{w: "9", lem: "9", pos: "CD", rel: "nummod", head: 3},
		{w: "sets", lem: "set", pos: "VBZ", rel: "root", head: 0},
		{w: "out", lem: "out", pos: "RP", rel: "compound:prt", head: 3},
		{w: "the", lem: "the", pos: "DT", rel: "det", head: 6},
		{w: "rule", lem: "rule", pos: "NN", rel: "obj", head: 3},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}, {
		{w: "This", lem: "this", pos: "DT", rel: "det", head: 2},
		{w: "book", lem: "book", pos: "NN", rel: "nsubj", head: 5},
		{w: "does", lem: "do", pos: "VBZ", rel: "aux", head: 5},
		{w: "not", lem: "not", pos: "RB", rel: "advmod", head: 5},
		{w: "set", lem: "set", pos: "VB", rel: "root", head: 0},
		{w: "out", lem: "out", pos: "RP", rel: "compound:prt", head: 5},
		{w: "to", lem: "to", pos: "TO", rel: "mark", head: 8},
		{w: "diminish", lem: "diminish", pos: "VB", rel: "xcomp", head: 5},
		{w: "any", lem: "any", pos: "DT", rel: "det", head: 10},
		{w: "faith", lem: "faith", pos: "NN", rel: "obj", head: 8},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 5},
	}})
	ss := analyzeDoc(src, doc, text, nil)
	if kindCount(ss, "close echo") == 0 || !strings.Contains(findingNotes(ss, "close echo"), `"set"`) {
		t.Fatalf("close echo on set did not fire\n%s", dumpFindings(ss))
	}
}

func TestDummyItChainSkipped(t *testing.T) {
	// "It is clear that rain falls. It is true that wood rots."
	src, doc, text := buildDoc([][]tw{{
		{w: "It", lem: "it", pos: "PRP", rel: "expl", head: 3},
		{w: "is", lem: "be", pos: "VBZ", rel: "cop", head: 3},
		{w: "clear", lem: "clear", pos: "JJ", rel: "root", head: 0},
		{w: "that", lem: "that", pos: "IN", rel: "mark", head: 6},
		{w: "rain", lem: "rain", pos: "NN", rel: "nsubj", head: 6},
		{w: "falls", lem: "fall", pos: "VBZ", rel: "csubj", head: 3},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}, {
		{w: "It", lem: "it", pos: "PRP", rel: "expl", head: 3},
		{w: "is", lem: "be", pos: "VBZ", rel: "cop", head: 3},
		{w: "true", lem: "true", pos: "JJ", rel: "root", head: 0},
		{w: "that", lem: "that", pos: "IN", rel: "mark", head: 6},
		{w: "wood", lem: "wood", pos: "NN", rel: "nsubj", head: 6},
		{w: "rots", lem: "rot", pos: "VBZ", rel: "csubj", head: 3},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}})
	ss := analyzeDoc(src, doc, text, map[string]float64{"rain": 4.97, "wood": 4.85})
	if kindCount(ss, "it chain") != 0 {
		t.Fatalf("dummy it chained\n%s", dumpFindings(ss))
	}
}

func TestConcreteSubjectsNotAbstract(t *testing.T) {
	// "A mason seeks sound stone." and "The father decides."
	src, doc, text := buildDoc([][]tw{{
		{w: "A", lem: "a", pos: "DT", rel: "det", head: 2},
		{w: "mason", lem: "mason", pos: "NN", rel: "nsubj", head: 3},
		{w: "seeks", lem: "seek", pos: "VBZ", rel: "root", head: 0},
		{w: "sound", lem: "sound", pos: "JJ", rel: "amod", head: 5},
		{w: "stone", lem: "stone", pos: "NN", rel: "obj", head: 3},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}}, [][]tw{{
		{w: "The", lem: "the", pos: "DT", rel: "det", head: 2},
		{w: "father", lem: "father", pos: "NN", rel: "nsubj", head: 3},
		{w: "decides", lem: "decide", pos: "VBZ", rel: "root", head: 0},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}})
	ss := analyzeDoc(src, doc, text, map[string]float64{"mason": 4.20, "father": 4.52, "stone": 4.72})
	if kindCount(ss, "abstract subject") != 0 {
		t.Fatalf("physical subjects flagged\n%s", dumpFindings(ss))
	}
}

func TestAmbiguousItSameSentenceCoref(t *testing.T) {
	// Sent 1: "The method holds."
	// Sent 2: "That life resists default; it needs work."
	// Coref links "it" to "life" in Sent 2 (only 1 intervening noun: default).
	// Must NOT fall back to "method" in Sent 1.
	src, doc, text := buildDoc([][]tw{{
		{w: "The", lem: "the", pos: "DT", rel: "det", head: 2},
		{w: "method", lem: "method", pos: "NN", rel: "nsubj", head: 3},
		{w: "holds", lem: "hold", pos: "VBZ", rel: "root", head: 0},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}, {
		{w: "That", lem: "that", pos: "DT", rel: "det", head: 2},
		{w: "life", lem: "life", pos: "NN", rel: "nsubj", head: 3},
		{w: "resists", lem: "resist", pos: "VBZ", rel: "root", head: 0},
		{w: "default", lem: "default", pos: "NN", rel: "obj", head: 3},
		{w: ";", lem: ";", pos: ":", rel: "punct", head: 3},
		{w: "it", lem: "it", pos: "PRP", rel: "nsubj", head: 7},
		{w: "needs", lem: "need", pos: "VBZ", rel: "parataxis", head: 3},
		{w: "work", lem: "work", pos: "NN", rel: "obj", head: 7},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}})
	doc.Corefs["1"] = []cnMention{
		{SentNum: 2, Start: 1, End: 3, Head: 2, Type: "NOMINAL", Text: "That life"},
		{SentNum: 2, Start: 6, End: 7, Head: 6, Type: "PRONOMINAL", Text: "it"},
	}
	ss := analyzeDoc(src, doc, text, nil)
	if kindCount(ss, "ambiguous it") != 0 {
		t.Fatalf("same-sentence coref with 1 intervening noun flagged\n%s", dumpFindings(ss))
	}
}

func TestAmbiguousItCrossSentenceSubjectFallback(t *testing.T) {
	// Sent 1: "The bridge carries traffic over the valley." (subject: bridge; intervening: traffic, valley)
	// Sent 2: "It stands." (no coref chain; no earlier noun in Sent 2)
	src, doc, text := buildDoc([][]tw{{
		{w: "The", lem: "the", pos: "DT", rel: "det", head: 2},
		{w: "bridge", lem: "bridge", pos: "NN", rel: "nsubj", head: 3},
		{w: "carries", lem: "carry", pos: "VBZ", rel: "root", head: 0},
		{w: "traffic", lem: "traffic", pos: "NN", rel: "obj", head: 3},
		{w: "over", lem: "over", pos: "IN", rel: "case", head: 7},
		{w: "the", lem: "the", pos: "DT", rel: "det", head: 7},
		{w: "valley", lem: "valley", pos: "NN", rel: "obl", head: 3},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}, {
		{w: "It", lem: "it", pos: "PRP", rel: "nsubj", head: 2},
		{w: "stands", lem: "stand", pos: "VBZ", rel: "root", head: 0},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 2},
	}})
	ss := analyzeDoc(src, doc, text, nil)
	if kindCount(ss, "ambiguous it") != 1 || !strings.Contains(findingNotes(ss, "ambiguous it"), `"bridge"`) {
		t.Fatalf("expected ambiguous it pointing to bridge\n%s", dumpFindings(ss))
	}
}

func TestAmbiguousItUnchainedMidSentenceNotBlamedOnPrev(t *testing.T) {
	// Sent 1: "The soul endures."
	// Sent 2: "This book rejects theology; it stays plain." (no coref chain)
	// Must NOT blame "soul" in Sent 1 for "it" in Sent 2.
	src, doc, text := buildDoc([][]tw{{
		{w: "The", lem: "the", pos: "DT", rel: "det", head: 2},
		{w: "soul", lem: "soul", pos: "NN", rel: "nsubj", head: 3},
		{w: "endures", lem: "endure", pos: "VBZ", rel: "root", head: 0},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}, {
		{w: "This", lem: "this", pos: "DT", rel: "det", head: 2},
		{w: "book", lem: "book", pos: "NN", rel: "nsubj", head: 3},
		{w: "rejects", lem: "reject", pos: "VBZ", rel: "root", head: 0},
		{w: "theology", lem: "theology", pos: "NN", rel: "obj", head: 3},
		{w: ";", lem: ";", pos: ":", rel: "punct", head: 3},
		{w: "it", lem: "it", pos: "PRP", rel: "nsubj", head: 7},
		{w: "stays", lem: "stay", pos: "VBZ", rel: "parataxis", head: 3},
		{w: "plain", lem: "plain", pos: "JJ", rel: "xcomp", head: 7},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 3},
	}})
	ss := analyzeDoc(src, doc, text, nil)
	if kindCount(ss, "ambiguous it") != 0 {
		t.Fatalf("unchained mid-sentence it blamed on prev\n%s", dumpFindings(ss))
	}
}

func TestItCleftSkipped(t *testing.T) {
	// "It is the mason or the smith who builds."
	src, doc, text := buildDoc([][]tw{{
		{w: "It", lem: "it", pos: "PRP", rel: "nsubj", head: 4},
		{w: "is", lem: "be", pos: "VBZ", rel: "cop", head: 4},
		{w: "the", lem: "the", pos: "DT", rel: "det", head: 4},
		{w: "mason", lem: "mason", pos: "NN", rel: "root", head: 0},
		{w: "or", lem: "or", pos: "CC", rel: "cc", head: 7},
		{w: "the", lem: "the", pos: "DT", rel: "det", head: 7},
		{w: "smith", lem: "smith", pos: "NN", rel: "conj", head: 4},
		{w: "who", lem: "who", pos: "WP", rel: "nsubj", head: 9},
		{w: "builds", lem: "build", pos: "VBZ", rel: "acl:relcl", head: 7},
		{w: ".", lem: ".", pos: ".", rel: "punct", head: 4},
	}})
	ss := analyzeDoc(src, doc, text, nil)
	if len(ss) != 1 || ss[0].Voice != "expletive" {
		t.Fatalf("it-cleft not classified as expletive: %+v", ss[0])
	}
}
