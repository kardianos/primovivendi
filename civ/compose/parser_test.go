package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParserBackendPriority checks the status badge and the parse order of
// Engine.ParseSentence: cache, then the Go parser, then the CoreNLP server
// (only when opted in), then the fallback.
func TestParserBackendPriority(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "civ", "chapters"), 0o755); err != nil {
		t.Fatal(err)
	}
	eng, err := NewEngine(root, filepath.Join(root, "civ", "notes", "compose"), false)
	if err != nil {
		t.Fatal(err)
	}
	if eng.Parser != nil {
		t.Fatal("an untrained root must not load the Go parser")
	}
	if got := eng.ParserStatus(); got != "Parser: fallback" {
		t.Fatalf("badge = %q, want %q", got, "Parser: fallback")
	}
	// The Go pipeline is absent here, so this exercises the fallback (or, if a
	// model appears, the Go parser): either way the tree is a valid tree.
	tr := eng.ParseSentence("We come into this world naked.")
	if tr == nil || tr.N() != 7 || tr.Root == 0 {
		t.Fatalf("unexpected tree: %+v", tr)
	}
	if tr.Surface(1, tr.N()) != "We come into this world naked." {
		t.Fatalf("surface = %q", tr.Surface(1, tr.N()))
	}
	// A second call is served from the in-memory cache.
	tr2 := eng.ParseSentence("  We   come into this world naked. ")
	if tr2 != tr {
		t.Fatal("whitespace-normalized repeats should hit the tree cache")
	}
	// A trained pipeline takes over: drop a stub under ext-ref/parser? Not
	// possible without weights, so only verify the server flag's badge.
	eng.UseCoreNLPServer = true
	if got := eng.ParserStatus(); got != "Parser: CoreNLP server" {
		t.Fatalf("badge = %q, want %q", got, "Parser: CoreNLP server")
	}
	eng.UseCoreNLPServer = false
	for _, empty := range []string{"", "   ", "\n\n"} {
		if got := eng.ParseSentence(empty); got == nil || got.N() != 0 {
			t.Fatalf("empty input %q should give an empty tree", empty)
		}
	}
	// Brackets get the Penn Treebank bracket tags whichever backend answered.
	tr = eng.ParseSentence("A (short) note.")
	if !strings.Contains(tr.Toks[2].POS, "LRB") || !strings.Contains(tr.Toks[4].POS, "RRB") {
		t.Fatalf("PTB bracket tags missing: %+v", tr.Toks)
	}
}
