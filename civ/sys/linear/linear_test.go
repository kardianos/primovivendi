package linear

import (
	"strings"
	"testing"

	"po/civ/sys/corenlp"
)

func TestGenerateStructuresAndLocks(t *testing.T) {
	tr := corenlp.ParseFallback("The bridge stands firm; the valley endures.")
	opts := GenerateStructures(tr, nil)
	hasSplit := false
	for _, o := range opts {
		if o.Kind == "split_mark" {
			hasSplit = true
		}
	}
	if !hasSplit {
		t.Fatalf("expected split_mark option for semicolon sentence, got %+v", opts)
	}

	// Lock across the semicolon: split_mark must be suppressed!
	locked := GenerateStructures(tr, []SpanLock{{Lo: 4, Hi: 6, Label: "set phrase"}})
	for _, o := range locked {
		if o.Kind == "split_mark" {
			t.Fatalf("split_mark violated span lock: %+v", o)
		}
	}

	asideTree := corenlp.ParseFallback("The court (which tests the witness directly) relies on evidence.")
	asideOpts := GenerateStructures(asideTree, nil)
	hasAside := false
	for _, o := range asideOpts {
		if o.Kind == "extract_aside" {
			hasAside = true
		}
	}
	if !hasAside {
		t.Fatalf("expected extract_aside option, got %+v", asideOpts)
	}
}

func TestSerialListDoesNotSplitAtAndButSplitsAtSubord(t *testing.T) {
	// 1. Exact CoreNLP parse shape from ch01 (including CoreNLP attaching does/raises/repairs as conj of ROOT).
	words := []struct {
		w, pos string
		head   int
		rel    string
	}{
		{"That", "DT", 2, "det"},
		{"learning", "NN", 3, "nsubj"},
		{"takes", "VBZ", 0, "ROOT"},
		{"place", "NN", 3, "obj"},
		{"in", "IN", 7, "case"},
		{"ordinary", "JJ", 7, "amod"},
		{"lives", "NNS", 3, "obl"},
		{",", ",", 3, "punct"},
		{"wherever", "WRB", 12, "advmod"},
		{"a", "DT", 11, "det"},
		{"person", "NN", 12, "nsubj"},
		{"tells", "VBZ", 3, "advcl"},
		{"the", "DT", 14, "det"},
		{"truth", "NN", 12, "obj"},
		{",", ",", 3, "punct"},
		{"does", "VBZ", 3, "conj"},
		{"good", "JJ", 18, "amod"},
		{"work", "NN", 16, "obj"},
		{",", ",", 3, "punct"},
		{"raises", "VBZ", 3, "conj"},
		{"children", "NNS", 20, "obj"},
		{",", ",", 3, "punct"},
		{"and", "CC", 24, "cc"},
		{"repairs", "NNS", 3, "conj"},
		{"what", "WP", 27, "nsubj:pass"},
		{"was", "VBD", 27, "aux:pass"},
		{"handed", "VBN", 24, "acl:relcl"},
		{"down", "RP", 27, "compound:prt"},
		{"to", "IN", 30, "case"},
		{"him", "PRP", 27, "obl"},
		{".", ".", 3, "punct"},
	}
	var s corenlp.Sent
	for i, item := range words {
		idx := i + 1
		s.Toks = append(s.Toks, corenlp.Tok{Index: idx, Word: item.w, Orig: item.w, POS: item.pos, Lemma: strings.ToLower(item.w), NER: "O"})
		s.Deps = append(s.Deps, corenlp.Dep{Dep: item.rel, Governor: item.head, Dependent: idx})
	}
	tr, ok := corenlp.BuildTree(s, "")
	if !ok {
		t.Fatal("BuildTree failed")
	}

	opts := GenerateStructures(tr, nil)
	var gotSubord, gotColon string
	for _, o := range opts {
		if o.Kind == "split_conj" {
			t.Fatalf("unexpected split_conj on list-final 'and': %+v", o)
		}
		if o.Kind == "swap_coord" {
			t.Fatalf("unexpected swap_coord on 3+ item serial list: %+v", o)
		}
		if o.Kind == "split_subord" {
			gotSubord = o.Text
		}
		if o.Kind == "split_colon_subord" {
			gotColon = o.Text
		}
	}
	wantSubord := "That learning takes place in ordinary lives. A person tells the truth, does good work, raises children, and repairs what was handed down to him."
	if gotSubord != wantSubord {
		t.Fatalf("split_subord mismatch:\n got: %q\nwant: %q", gotSubord, wantSubord)
	}
	wantColon := "That learning takes place in ordinary lives: a person tells the truth, does good work, raises children, and repairs what was handed down to him."
	if gotColon != wantColon {
		t.Fatalf("split_colon_subord mismatch:\n got: %q\nwant: %q", gotColon, wantColon)
	}

	// 2. Fallback parse on the edited sentence with ", as each person ...".
	fb := corenlp.ParseFallback("Learning how to replicate civilization happens in ordinary lives, as each person tells the truth, does good work, raises children, and fixes what was passed down to him.")
	fbOpts := GenerateStructures(fb, nil)
	var fbSubord, fbColon string
	for _, o := range fbOpts {
		if o.Kind == "split_conj" {
			t.Fatalf("fallback unexpectedly emitted split_conj on list-final 'and': %+v", o)
		}
		if o.Kind == "split_subord" {
			fbSubord = o.Text
		}
		if o.Kind == "split_colon_subord" {
			fbColon = o.Text
		}
	}
	wantFB := "Learning how to replicate civilization happens in ordinary lives. Each person tells the truth, does good work, raises children, and fixes what was passed down to him."
	if fbSubord != wantFB {
		t.Fatalf("fallback split_subord mismatch:\n got: %q\nwant: %q", fbSubord, wantFB)
	}
	wantFBColon := "Learning how to replicate civilization happens in ordinary lives: each person tells the truth, does good work, raises children, and fixes what was passed down to him."
	if fbColon != wantFBColon {
		t.Fatalf("fallback split_colon_subord mismatch:\n got: %q\nwant: %q", fbColon, wantFBColon)
	}

	// 3. Nested advcl attached below ROOT (e.g. gerund subject "Learning..." as ROOT,
	// "happens" under it, and ", as each person tells..." attached to "happens").
	nestedWords := []struct {
		w, pos string
		head   int
		rel    string
	}{
		{"Learning", "VBG", 0, "ROOT"},
		{"how", "WRB", 1, "obj"},
		{"to", "TO", 4, "mark"},
		{"replicate", "VB", 2, "acl"},
		{"civilization", "NN", 4, "obj"},
		{"happens", "VBZ", 2, "dep"},
		{"in", "IN", 9, "case"},
		{"ordinary", "JJ", 9, "amod"},
		{"lives", "NNS", 6, "obl"},
		{",", ",", 6, "punct"},
		{"as", "IN", 14, "mark"},
		{"each", "DT", 13, "det"},
		{"person", "NN", 14, "nsubj"},
		{"tells", "VBZ", 6, "advcl"},
		{"the", "DT", 16, "det"},
		{"truth", "NN", 14, "obj"},
		{",", ",", 6, "punct"},
		{"does", "VBZ", 6, "dep"},
		{"good", "JJ", 20, "amod"},
		{"work", "NN", 22, "nsubj"},
		{",", ",", 20, "punct"},
		{"raises", "VBZ", 18, "ccomp"},
		{"children", "NNS", 22, "obj"},
		{",", ",", 22, "punct"},
		{"and", "CC", 26, "cc"},
		{"fixes", "NNS", 22, "conj"},
		{"what", "WP", 29, "nsubj:pass"},
		{"was", "VBD", 29, "aux:pass"},
		{"passed", "VBN", 26, "acl:relcl"},
		{"down", "RP", 29, "compound:prt"},
		{"to", "IN", 32, "case"},
		{"him", "PRP", 29, "obl"},
		{".", ".", 1, "punct"},
	}
	var ns corenlp.Sent
	for i, item := range nestedWords {
		idx := i + 1
		ns.Toks = append(ns.Toks, corenlp.Tok{Index: idx, Word: item.w, Orig: item.w, POS: item.pos, Lemma: strings.ToLower(item.w), NER: "O"})
		ns.Deps = append(ns.Deps, corenlp.Dep{Dep: item.rel, Governor: item.head, Dependent: idx})
	}
	ntr, ok := corenlp.BuildTree(ns, "")
	if !ok {
		t.Fatal("BuildTree failed on nested advcl")
	}
	var nestedSubord, nestedColon string
	for _, o := range GenerateStructures(ntr, nil) {
		if o.Kind == "split_conj" {
			t.Fatalf("nested advcl tree unexpectedly emitted split_conj: %+v", o)
		}
		if o.Kind == "split_subord" {
			nestedSubord = o.Text
		}
		if o.Kind == "split_colon_subord" {
			nestedColon = o.Text
		}
	}
	if nestedSubord != wantFB {
		t.Fatalf("nested split_subord mismatch:\n got: %q\nwant: %q", nestedSubord, wantFB)
	}
	if nestedColon != wantFBColon {
		t.Fatalf("nested split_colon_subord mismatch:\n got: %q\nwant: %q", nestedColon, wantFBColon)
	}

	// Lock across ", as": must suppress split_subord and split_colon_subord.
	lockedOpts := GenerateStructures(fb, []SpanLock{{Lo: 8, Hi: 11, Label: "locked"}})
	for _, o := range lockedOpts {
		if o.Kind == "split_subord" || o.Kind == "split_colon_subord" {
			t.Fatalf("split_subord violated span lock: %+v", o)
		}
	}
}

func TestSplitConjStillAllowsTrueCompoundClauses(t *testing.T) {
	// Genuine 2-clause compound sentence with independent subjects on both sides.
	words := []struct {
		w, pos string
		head   int
		rel    string
	}{
		{"The", "DT", 2, "det"},
		{"bridge", "NN", 3, "nsubj"},
		{"held", "VBD", 0, "ROOT"},
		{"through", "IN", 6, "case"},
		{"the", "DT", 6, "det"},
		{"flood", "NN", 3, "obl"},
		{",", ",", 11, "punct"},
		{"and", "CC", 11, "cc"},
		{"the", "DT", 10, "det"},
		{"town", "NN", 11, "nsubj"},
		{"rebuilt", "VBD", 3, "conj"},
		{"the", "DT", 13, "det"},
		{"road", "NN", 11, "obj"},
		{"before", "IN", 15, "case"},
		{"winter", "NN", 11, "obl"},
		{".", ".", 3, "punct"},
	}
	var s corenlp.Sent
	for i, item := range words {
		idx := i + 1
		s.Toks = append(s.Toks, corenlp.Tok{Index: idx, Word: item.w, Orig: item.w, POS: item.pos, Lemma: strings.ToLower(item.w), NER: "O"})
		s.Deps = append(s.Deps, corenlp.Dep{Dep: item.rel, Governor: item.head, Dependent: idx})
	}
	tr, ok := corenlp.BuildTree(s, "")
	if !ok {
		t.Fatal("BuildTree failed")
	}
	opts := GenerateStructures(tr, nil)
	found := false
	for _, o := range opts {
		if o.Kind == "split_conj" {
			found = true
			want := "The bridge held through the flood. The town rebuilt the road before winter."
			if o.Text != want {
				t.Fatalf("split_conj mismatch:\n got: %q\nwant: %q", o.Text, want)
			}
		}
	}
	if !found {
		t.Fatalf("expected split_conj for true compound sentence, got %+v", opts)
	}

	// Prepositional ", as a man of honor," must NOT trigger split_subord.
	prepFB := corenlp.ParseFallback("He stood firm in the hall, as a man of honor, and spoke plainly to the council.")
	for _, o := range GenerateStructures(prepFB, nil) {
		if o.Kind == "split_subord" || o.Kind == "split_colon_subord" {
			t.Fatalf("prepositional 'as a man of honor' should not emit %s: %+v", o.Kind, o)
		}
	}

	// Adverb before 3sg verb in fallback (", as each person quietly tells the truth, ...") must still split.
	advFB := corenlp.ParseFallback("Learning how to replicate civilization happens in ordinary lives, as each person quietly tells the truth, does good work, raises children, and fixes what was passed down to him.")
	hasAdvSubord := false
	for _, o := range GenerateStructures(advFB, nil) {
		if o.Kind == "split_conj" {
			t.Fatalf("unexpected split_conj on adverb+list sentence: %+v", o)
		}
		if o.Kind == "split_subord" {
			hasAdvSubord = true
		}
	}
	if !hasAdvSubord {
		t.Fatalf("expected split_subord on adverb+3sg fallback sentence")
	}
}
