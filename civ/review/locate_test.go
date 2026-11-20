package main

import (
	"os"
	"strings"
	"testing"
)

func TestLocate(t *testing.T) {
	file, err := os.ReadFile("testdata/passage.md")
	if err != nil {
		t.Fatal(err)
	}
	household := "The household keeps its own children."
	fire := "Fire burns cotton because of what fire and cotton are."

	got := Locate(file, Item{Quote: household, ByteOffset: intPtr(0), ByteEnd: intPtr(len(household))})
	if got.Status != StatusOffset || got.Offset != 0 || got.Line != 1 || got.Col != 1 {
		t.Fatalf("offset hit: %+v", got)
	}
	if !CanAccept(got, false) {
		t.Fatal("an offset hit can be accepted")
	}

	got = Locate(file, Item{Quote: household, ByteOffset: intPtr(5), Before: "\n\n"})
	if got.Status != StatusMoved || got.Offset != 200 {
		t.Fatalf("moved hit: %+v", got)
	}

	got = Locate(file, Item{Quote: household, ByteOffset: intPtr(3)})
	if got.Status != StatusAmbiguous || len(got.Candidates) != 2 || got.Candidates[0] != 0 || got.Candidates[1] != 200 {
		t.Fatalf("ambiguous hit: %+v", got)
	}
	if CanAccept(got, false) {
		t.Fatal("an ambiguous hit cannot be accepted")
	}
	if !CanAccept(got, true) {
		t.Fatal("reaffirm allows an ambiguous hit")
	}

	got = Locate(file, Item{Quote: household, Before: "nope"})
	if got.Status != StatusAmbiguous || len(got.Candidates) != 2 {
		t.Fatalf("context that fits nothing stays ambiguous: %+v", got)
	}

	got = Locate(file, Item{Quote: fire})
	if got.Status != StatusMoved || got.Offset != 99 || got.End != 153 || got.Line != 3 {
		t.Fatalf("unique search: %+v", got)
	}

	got = Locate(file, Item{Quote: "This sentence is not in the passage.", ByteOffset: intPtr(0)})
	if got.Status != StatusMissing {
		t.Fatalf("missing: %+v", got)
	}

	got = Locate(file, Item{Quote: "", ByteOffset: intPtr(0)})
	if got.Status != StatusMissing {
		t.Fatalf("empty quote: %+v", got)
	}

	work := "The work — the real work — is keeping what was built."
	got = Locate(file, Item{Quote: work, ByteOffset: intPtr(318), ByteEnd: intPtr(375)})
	if got.Status != StatusOffset || got.Offset != 318 || got.End != 375 {
		t.Fatalf("em dash offset: %+v", got)
	}
	if _, _, _, ok := snippet(file, got, defaultContextLines, defaultContextLines); !ok {
		t.Fatal("snippet of an offset hit")
	}
}

func TestSnippetGrowsToTheChapter(t *testing.T) {
	file, err := os.ReadFile("testdata/passage.md")
	if err != nil {
		t.Fatal(err)
	}
	fire := Locate(file, Item{Quote: "Fire burns cotton because of what fire and cotton are.", ByteOffset: intPtr(99), ByteEnd: intPtr(153)})
	pre, match, post, ok := snippet(file, fire, 2, 2)
	if !ok || match != fireQuote(file, fire) {
		t.Fatal("tight snippet")
	}
	if !strings.Contains(pre, "continuance") {
		t.Fatal("two lines above include the previous paragraph")
	}
	if strings.Contains(post, "keeping what was built") {
		t.Fatal("two lines below stop short of the chapter's end")
	}
	if _, _, wider, _ := snippet(file, fire, 2, moreContext(2)); !strings.Contains(wider, "keeping what was built") {
		t.Fatal("the next step below reaches the end of this fixture")
	}
	pre, match, post, ok = snippet(file, fire, -1, -1)
	if !ok || pre+match+post != string(file) {
		t.Fatal("a negative count is the whole chapter")
	}
	work := Locate(file, Item{Quote: "The work — the real work — is keeping what was built.", ByteOffset: intPtr(318), ByteEnd: intPtr(375)})
	above, _, _, ok := snippet(file, work, 80, 0)
	if !ok || !strings.HasPrefix(above, "The household") {
		t.Fatal("opening above the last sentence reaches the start")
	}
}

func fireQuote(file []byte, h Hit) string { return string(file[h.Offset:h.End]) }

func intPtr(n int) *int { return &n }
