package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVoiceMetrics(t *testing.T) {
	plain := `We come into this world naked. Beneath us stands reality. A wet child dies on bare ground in two days. Put him in a warm house with clean water and food, and he lives. The wall keeps out the wind. The roof keeps out the rain. None of that holds by itself.`
	m := measureVoice(plain)
	if m.Words < 40 {
		t.Fatalf("expected >=40 words, got %d", m.Words)
	}
	if m.ShortSentPct < 70 {
		t.Errorf("plain sample should have high short-sentence share, got %.1f%%", m.ShortSentPct)
	}
	if m.LatinatePct > 1.0 {
		t.Errorf("plain sample should have near-zero Latinate rate, got %.2f%%", m.LatinatePct)
	}
	if m.SignpostPct != 0 {
		t.Errorf("plain sample should have 0%% signposts, got %.1f%%", m.SignpostPct)
	}
	if m.Coaching100 != 0 {
		t.Errorf("plain sample should have 0 coaching markers, got %.2f", m.Coaching100)
	}
	if m.Score >= 0.0 {
		t.Errorf("plain sample should score below 0.0, got %+.2f", m.Score)
	}

	textbook := `First, institutional frameworks fundamentally and systematically require operational alignment, structural coherence, and procedural verification across every administrative dimension of your public organization. Second, consequently, when you explicitly abandon objective standards, meritocratic selection, and empirical accountability in favor of "compassion" or "equity," does your bureaucratic apparatus retain functional legitimacy? Third, therefore, you must deliberately enforce foundational boundaries, statutory clarity, and institutional discipline, because unconditional permissiveness inevitably produces systemic deterioration, administrative paralysis, and civilizational decline.`
	tm := measureVoice(textbook)
	if tm.LongWordPct < 40 {
		t.Errorf("textbook sample should have high long-word share, got %.1f%%", tm.LongWordPct)
	}
	if tm.LatinatePct < 15 {
		t.Errorf("textbook sample should have high Latinate rate, got %.1f%%", tm.LatinatePct)
	}
	if tm.SignpostPct < 90 {
		t.Errorf("textbook sample should have high signpost rate, got %.1f%%", tm.SignpostPct)
	}
	if tm.Triplets100 < 2.0 {
		t.Errorf("textbook sample should detect parallel triplets, got %.2f/100w", tm.Triplets100)
	}
	if tm.Coaching100 < 5.0 {
		t.Errorf("textbook sample should detect coaching markers, got %.2f/100w", tm.Coaching100)
	}
	if tm.Score < 2.0 {
		t.Errorf("textbook sample should score well above +2.0, got %+.2f", tm.Score)
	}
	if tm.Score-m.Score < 2.5 {
		t.Errorf("expected wide separation between plain (%+.2f) and textbook (%+.2f)", m.Score, tm.Score)
	}
}

func TestVoiceMarkdownStrippingAndSeams(t *testing.T) {
	doc := strings.Join([]string{
		"## Part One",
		"The wall holds. The roof keeps rain off the beam. A man cuts oak, drives pegs, and sets stone on stone. " +
			"Rain falls hard in winter. Frost splits wet rock. He mends the mortar in spring and checks the sill. " +
			"Nothing lasts by itself. Every house needs hands to keep it dry and sound year after year. " +
			"A farm works the same way. Clear the brush, plow the ground, sow clean seed, and mend the fence. " +
			"Leave the field alone and weeds take it back in five years. The iron rusts, the wood rots, and the ditch fills with mud. " +
			"Hands keep the work alive. When the hands stop, the work ends.",
		"",
		"## Part Two",
		"First, **institutional frameworks** fundamentally and systematically require operational alignment, structural coherence, and procedural verification across every administrative dimension of your organization. " +
			"Second, consequently, when you explicitly abandon objective standards, meritocratic selection, and empirical accountability in favor of \"equity\" or \"compassion,\" does your bureaucratic apparatus retain functional legitimacy? " +
			"Third, therefore, you must deliberately enforce foundational boundaries, statutory clarity, and institutional discipline, because unconditional permissiveness inevitably produces systemic deterioration, administrative paralysis, and civilizational decline. " +
			"Fourth, moreover, administrative centralization systematically undermines local accountability, institutional resilience, and operational competence across every municipal jurisdiction in your republic.",
	}, "\n")

	secs := parseVoiceSections("sample.md", doc, 80)
	if len(secs) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(secs))
	}
	if secs[0].Metrics.Score >= secs[1].Metrics.Score {
		t.Errorf("expected plain section (%+.2f) < textbook section (%+.2f)", secs[0].Metrics.Score, secs[1].Metrics.Score)
	}
	seams := findVoiceSeams(secs, 1.5)
	if len(seams) != 1 {
		t.Fatalf("expected 1 voice seam >= 1.5, got %d (scores %+.2f -> %+.2f)", len(seams), secs[0].Metrics.Score, secs[1].Metrics.Score)
	}
}

func TestVoiceChaptersCalibration(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Skip("not inside git repo")
	}
	lookup := func(chFile, headingSubstr string) float64 {
		t.Helper()
		p := filepath.Join(root, "civ", "chapters", chFile)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		for _, s := range parseVoiceSections(p, string(b), 100) {
			if strings.Contains(strings.ToLower(s.Heading), strings.ToLower(headingSubstr)) {
				return s.Metrics.Score
			}
		}
		t.Fatalf("section %q not found in %s", headingSubstr, chFile)
		return 0
	}

	plainAnchors := [][2]string{
		{"01_ch01.md", "The Immediate Fact of Our Existence"},
		{"02_ch02.md", "Know in Part"},
		{"03_ch03.md", "Ash'arite Occasionalism"},
		{"09_ch09.md", "Counterfeit 2: Value Symmetry"},
		{"10_ch10.md", `Testing "Not Ready"`},
		{"12_ch12.md", "B. Islam:"},
	}
	plainMax := -999.0
	for _, a := range plainAnchors {
		sc := lookup(a[0], a[1])
		if sc > 0.35 {
			t.Errorf("plain anchor %s (%s) scored too high: %+.2f", a[0], a[1], sc)
		}
		if sc > plainMax {
			plainMax = sc
		}
	}

	textbookAnchors := [][2]string{
		{"03_ch03.md", "3. Postmodern Constructivism"},
		{"11_ch11.md", "Repressive Tolerance"},
		{"13_ch13.md", "Five Mechanics of Modern Red-Green"},
	}
	for _, a := range textbookAnchors {
		sc := lookup(a[0], a[1])
		if sc <= plainMax {
			t.Errorf("textbook anchor %s (%s) scored %+.2f, not above highest plain anchor %+.2f", a[0], a[1], sc, plainMax)
		}
	}
}

func TestSpellingRaisingAndGrey(t *testing.T) {
	if hits := spellingHits("f.md", 1, "A neighborly barn-raising and raising children."); len(hits) != 0 {
		t.Errorf("raising should not be flagged as British: %s", joinTexts(hits))
	}
	if hits := spellingHits("f.md", 2, "A sour, grey sponge."); len(hits) == 0 || !strings.Contains(hits[0].text, "gray") {
		t.Errorf("grey should be flagged as British (American: gray), got %v", hits)
	}
}
