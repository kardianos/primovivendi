package main

import (
	"strings"
	"testing"
)

func TestLoadCMUUppercase(t *testing.T) {
	d, err := loadCMU(strings.NewReader("stone S T OW1 N\nalone(2) AH0 L OW1 N\n"))
	if err != nil {
		t.Fatal(err)
	}
	suf, rhy := analyzeSentence("A stone stood alone.", d)
	if len(suf) != 0 {
		t.Fatalf("suffixes: %v", suf)
	}
	if len(rhy) != 1 || !strings.Contains(rhy[0], "stone") || !strings.Contains(rhy[0], "alone") {
		t.Fatalf("dict keys should match regardless of case: %v", rhy)
	}
}

func TestSentencesKeepAbbreviations(t *testing.T) {
	ss := sentencesOf("Mr. Smith left. Dr. Jones stayed.")
	if len(ss) != 2 {
		t.Fatalf("sentences: %+v", ss)
	}
	if !strings.Contains(ss[0].text, "Mr. Smith") {
		t.Fatalf("split on Mr.: %+v", ss)
	}
}

func TestHeavySuffixAndRhyme(t *testing.T) {
	d := dict{
		"NATION":   strings.Fields("N EY1 SH AH0 N"),
		"STATION":  strings.Fields("S T EY1 SH AH0 N"),
		"CREATION": strings.Fields("K R IY0 EY1 SH AH0 N"),
		"DOG":      strings.Fields("D AO1 G"),
		"LOG":      strings.Fields("L AO1 G"),
		"FROG":     strings.Fields("F R AO1 G"),
	}
	suf, rhy := analyzeSentence("The nation and the station followed the creation.", d)
	if len(suf) != 1 {
		t.Fatalf("suffixes: %v", suf)
	}
	if len(rhy) == 0 || !strings.Contains(rhy[0], "nation") || !strings.Contains(rhy[0], "station") {
		t.Fatalf("rhyme: %v", rhy)
	}
	_, near := analyzeSentence("The dog met one two three four five six seven eight log.", d)
	if len(near) != 1 {
		t.Fatalf("rhyme within 10 words: %v", near)
	}
	_, far := analyzeSentence("The dog met one two three four five six seven eight nine log.", d)
	if len(far) != 0 {
		t.Fatalf("rhyme past 10 words: %v", far)
	}
	_, same := analyzeSentence("The nation and the nations met.", d)
	if len(same) != 0 {
		t.Fatalf("identical lemma should not rhyme: %v", same)
	}
	d["GAIN"] = strings.Fields("G EY1 N")
	d["GRAIN"] = strings.Fields("G R EY1 N")
	_, twice := analyzeSentence("The gain of the grain and the grain.", d)
	if len(twice) != 1 || twice[0] != "gain / grain (EY1 N)" {
		t.Fatalf("a repeated lemma should be listed once: %v", twice)
	}
	// A stressed -ment is not the unstressed suffix.
	d["CEMENT"] = strings.Fields("S AH0 M EH1 N T")
	suf, _ = analyzeSentence("The document of cement and payment and treatment stood.", dict{
		"CEMENT":    strings.Fields("S AH0 M EH1 N T"),
		"PAYMENT":   strings.Fields("P EY1 M AH0 N T"),
		"TREATMENT": strings.Fields("T R IY1 T M AH0 N T"),
		"DOCUMENT":  strings.Fields("D AA1 K Y AH0 M AH0 N T"),
	})
	if len(suf) != 1 {
		t.Fatalf("cement should not count, three -ment words should: %v", suf)
	}
	if strings.Contains(suf[0], "cement") {
		t.Fatalf("stressed cement counted: %v", suf)
	}
}

func TestSentencesDoNotMergeAcrossParagraphs(t *testing.T) {
	md := "First paragraph ends here.\n\nSecond paragraph starts here.\n\n- Bullet one ends here.\n- Bullet two starts here.\n"
	ss := sentencesOf(md)
	if len(ss) != 4 {
		t.Fatalf("expected 4 sentences across paragraphs/bullets, got %d: %+v", len(ss), ss)
	}
	if ss[0].line != 1 || ss[1].line != 3 || ss[2].line != 5 || ss[3].line != 6 {
		t.Fatalf("unexpected line numbers: %+v", ss)
	}
}

func TestSuffixGroupedAndContrastRhymesSkipped(t *testing.T) {
	d := dict{
		"BLINDLY":      strings.Fields("B L AY1 N D L IY0"),
		"MERELY":       strings.Fields("M IH1 R L IY0"),
		"FACTION":      strings.Fields("F AE1 K SH AH0 N"),
		"ORDER":        strings.Fields("AO1 R D ER0"),
		"DISORDER":     strings.Fields("D IH0 S AO1 R D ER0"),
		"MORAL":        strings.Fields("M AO1 R AH0 L"),
		"IMMORAL":      strings.Fields("IH0 M AO1 R AH0 L"),
		"PHYSICAL":     strings.Fields("F IH1 Z IH0 K AH0 L"),
		"METAPHYSICAL": strings.Fields("M EH2 T AH0 F IH1 Z IH0 K AH0 L"),
		"STONE":        strings.Fields("S T OW1 N"),
		"THRONE":       strings.Fields("TH R OW1 N"),
		"ALONE":        strings.Fields("AH0 L OW1 N"),
	}
	suf, _ := analyzeSentence("He spoke blindly and merely to a faction.", d)
	if len(suf) != 0 {
		t.Fatalf("mixed suffix classes should not jingle: %v", suf)
	}
	_, rhy := analyzeSentence("Order beats disorder, moral beats immoral, physical beats metaphysical.", d)
	if len(rhy) != 0 {
		t.Fatalf("prefix contrast pairs should not rhyme: %v", rhy)
	}
	_, triple := analyzeSentence("The stone by the throne stood alone.", d)
	if len(triple) != 1 {
		t.Fatalf("triple rhyme should not also emit subset pair: %v", triple)
	}
}
