package main

import (
	"strings"
	"testing"
)

func TestSpellingAmericanAndQuotes(t *testing.T) {
	line := `The colour of honor is fine, and "colour" is quoted, and a promise to organise at the centre. We practise defence.`
	var got []string
	for _, h := range spellingHits("f.md", 1, line) {
		got = append(got, h.text)
	}
	joined := strings.Join(got, "\n")
	for _, want := range []string{"colour (American: color)", "organise", "centre", "practise", "defence"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s in %q", want, joined)
		}
	}
	if strings.Contains(joined, "honor") || strings.Contains(joined, "promise") {
		t.Errorf("flagged an American or allowed word: %s", joined)
	}
	n := strings.Count(joined, "colour (American: color)")
	if n != 1 {
		t.Errorf("quoted colour should be skipped once, findings: %s", joined)
	}
}

func TestSpellingCaseAndShortISE(t *testing.T) {
	hits := spellingHits("f.md", 3, "Colour rises. Organise the programme. NASA and McColour sit here.")
	joined := joinTexts(hits)
	if !strings.Contains(joined, "Color") {
		t.Errorf("Colour should suggest Color: %s", joined)
	}
	if strings.Contains(joined, "rises") || strings.Contains(joined, "rise (") {
		t.Errorf("rise is not a British -ise: %s", joined)
	}
	if !strings.Contains(joined, "Organize") {
		t.Errorf("Organise should suggest Organize: %s", joined)
	}
	if !strings.Contains(joined, "program") {
		t.Errorf("programme should be flagged: %s", joined)
	}
	if strings.Contains(joined, "NASA") || strings.Contains(joined, "McColour") {
		t.Errorf("interior capitals should be skipped: %s", joined)
	}
	kept := spellingHits("f.md", 4, "Likewise, the bruised and undisguised board was praised, praising an uncompromising and unsupervised crew. An uprising over merchandise, cruising, and disenfranchised voters was appraised.")
	if len(kept) != 0 {
		t.Errorf("American -ise words flagged: %s", joinTexts(kept))
	}
	extra := spellingHits("f.md", 5, "The labourer had an unfavourable and dishonourable habit, and we analyse the data.")
	ej := joinTexts(extra)
	for _, want := range []string{"laborer", "unfavorable", "dishonorable", "analyze"} {
		if !strings.Contains(ej, want) {
			t.Errorf("missing %s in %q", want, ej)
		}
	}
}

func TestPartAPatterns(t *testing.T) {
	if !partA[0].re.MatchString("a real pause — then") {
		t.Fatal("em dash")
	}
	if !partA[2].re.MatchString("the exact same stone") {
		t.Fatal("exact same")
	}
	if doubleHyphen("a -- b") != true {
		t.Fatal("double hyphen")
	}
	if doubleHyphen("---") || doubleHyphen("> ---") || doubleHyphen("| --- | --- |") {
		t.Fatal("a rule is not a double hyphen")
	}
}

func TestLockedLines(t *testing.T) {
	exact := []string{
		"1. Reality is real, knowable, and consequential. Truth is what corresponds to reality.",
		"2. Humans are self-interested, limited, and particular.",
		"3. Goodness must include human life continuing, and that life kept and built.",
		"4. What you keep and build is self-anchored, tiered, and outward. Enemies are real.",
		"5. Justice is scoped procedural symmetry.",
	}
	if hits := lockedHits("ch.md", exact); len(hits) != 0 {
		t.Fatalf("exact principles flagged: %+v", hits)
	}
	mut := append([]string(nil), exact...)
	mut[1] = "2. Humans are self-interested, limited, and general."
	hits := lockedHits("ch.md", mut)
	if len(hits) != 1 || !strings.Contains(hits[0].text, "Principle 2") {
		t.Fatalf("mutated principle: %+v", hits)
	}
	virtue := []string{
		"1. Align with truth.",
		"2. Take ownership, which is responsibility.",
		"3. Seek excellence.",
		"4. Admire beauty.",
		"5. Be humble.",
		"6. Have children.",
		"7. Form and listen.",
	}
	if hits := lockedHits("ch.md", virtue); len(hits) != 0 {
		t.Fatalf("virtues with periods flagged: %+v", hits)
	}
	quote := []string{
		"> **Principle 2 of Reality: Humans are self-interested, limited, and particular.**",
		"> **Virtue 1: Align with truth.**",
		"> **Virtue 5: Be humble.**",
	}
	if hits := lockedHits("ch.md", quote); len(hits) != 0 {
		t.Fatalf("blockquote statements flagged: %+v", hits)
	}
	bad := []string{"> **Virtue 1: Align with truthfulness.**"}
	hits = lockedHits("ch.md", bad)
	if len(hits) != 1 {
		t.Fatalf("bad virtue statement: %+v", hits)
	}
}

func TestCasing(t *testing.T) {
	hits := casingHits("f.md", 2, "See principle 2 and Virtue 5 and PRINCIPLE 2, then Virtues 6–7.")
	joined := joinTexts(hits)
	if !strings.Contains(joined, `"principle"`) || !strings.Contains(joined, `"PRINCIPLE"`) {
		t.Fatalf("casing: %s", joined)
	}
	if strings.Contains(joined, `"Virtue"`) || strings.Contains(joined, `"Virtues"`) {
		t.Fatalf("correct casing flagged: %s", joined)
	}
}

func joinTexts(hits []hit) string {
	var b strings.Builder
	for _, h := range hits {
		b.WriteString(h.text)
		b.WriteByte('\n')
	}
	return b.String()
}
