package corenlp

import "testing"

func TestParseConstituencyAndFallbackTree(t *testing.T) {
	sexpr := `(ROOT (S (NP (PRP We)) (VP (VBP come) (PP (IN into) (NP (DT this) (NN world))) (ADVP (JJ naked))) (. .)))`
	c := ParseConstituency(sexpr)
	if c == nil || c.Label != "ROOT" || len(c.Children) != 1 {
		t.Fatalf("unexpected constituency root: %+v", c)
	}
	tr := ParseFallback("We come into this world naked.")
	if tr.N() != 7 {
		t.Fatalf("expected 7 tokens, got %d", tr.N())
	}
	if tr.Words() != 6 {
		t.Fatalf("expected 6 words, got %d", tr.Words())
	}
	if tr.Root == 0 {
		t.Fatalf("expected non-zero root")
	}
	if got := tr.Surface(1, tr.N()); got != "We come into this world naked." {
		t.Fatalf("surface mismatch: got %q", got)
	}
}

func TestUTF8ByteOffsetsAndLegacyUTF16Normalization(t *testing.T) {
	// Contains 3-byte UTF-8 smart quotes & em-dash plus a 4-byte astral rune (U+10437 = 2 UTF-16 units).
	text := "“Truth”—𐐷 stands."
	tr := ParseFallback(text)
	for i := 1; i <= tr.N(); i++ {
		tok := tr.Toks[i]
		if got := text[tok.Begin:tok.End]; got != tok.Orig {
			t.Fatalf("fallback tok %d: text[%d:%d]=%q, want %q", i, tok.Begin, tok.End, got, tok.Orig)
		}
	}

	// Simulate Java/CoreNLP UTF-16 code-unit offsets for "stands":
	// “ (1) + Truth (5) + ” (1) + — (1) + 𐐷 (2) + space (1) = 11..17 in UTF-16.
	doc := &Doc{
		Sentences: []Sent{{
			Toks: []Tok{{Index: 1, Word: "stands", Orig: "stands", Begin: 11, End: 17}},
		}},
	}
	NormalizeUTF16ToUTF8(doc, text)
	tok := doc.Sentences[0].Toks[0]
	if got := text[tok.Begin:tok.End]; got != "stands" {
		t.Fatalf("NormalizeUTF16ToUTF8: text[%d:%d]=%q, want %q", tok.Begin, tok.End, got, "stands")
	}
}
