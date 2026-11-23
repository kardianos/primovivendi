package main

import (
	"strings"
	"testing"
)

func mockCadenceLexicons() *cadenceLexicons {
	cmuRaw := strings.Join([]string{
		"SCIENCE  S AY1 AH0 N S",
		"MARCHES  M AA1 R CH IH0 Z",
		"ON  AA1 N",
		"NUMEROUS  N UW1 M ER0 AH0 S",
		"STUDIES  S T AH1 D IY0 Z",
		"HAVE  HH AE1 V",
		"DEMONSTRATED  D EH1 M AH0 N S T R EY2 T AH0 D",
		"THAT  DH AE1 T",
		"MEN  M EH1 N",
		"ARE  AA1 R",
		"MORE  M AO1 R",
		"PSYCHOLOGICALLY  S AY2 K AH0 L AA1 JH IH0 K AH0 L IY0",
		"AND  AE1 N D",
		"PHYSIOLOGICALLY  F IH2 Z IY0 AH0 L AA1 JH IH0 K AH0 L IY0",
		"RESPONSIVE  R IH0 S P AA1 N S IH0 V",
		"TO  T UW1",
		"VISUAL  V IH1 ZH AH0 W AH0 L",
		"SEXUALLY  S EH1 K SH UW0 AH0 L IY0",
		"AROUSING  ER0 AW1 Z IH0 NG",
		"STIMULI  S T IH1 M Y AH0 L AY2",
		"THE  DH AH0",
		"OPERATIONAL  AA2 P ER0 EY1 SH AH0 N AH0 L",
		"IMPLEMENTATION  IH2 M P L AH0 M EH0 N T EY1 SH AH0 N",
		"OF  AH1 V",
		"PROCEDURAL  P R AH0 S IY1 JH ER0 AH0 L",
		"SAFEGUARDS  S EY1 F G AA2 R D Z",
		"FOR  F AO1 R",
		"MITIGATION  M IH2 T AH0 G EY1 SH AH0 N",
		"INSTITUTIONAL  IH2 N S T IH0 T UW1 SH AH0 N AH0 L",
		"FAILURE  F EY1 L Y ER0",
		"REQUIRES  R IH0 K W AY1 ER0 Z",
		"DISCONFIRMATION  D IH0 S K AA0 N F ER0 M EY1 SH AH0 N",
		"STONE  S T OW1 N",
		"WALL  W AO1 L",
		"STANDS  S T AE1 N D Z",
		"RAIN  R EY1 N",
		"CUTS  K AH1 T S",
		"CLAY  K L EY1",
		"BREAD  B R EH1 D",
		"FEEDS  F IY1 D Z",
		"HOUSE  HH AW1 S",
		"UNAFFECTED  AH2 N AH0 F EH1 K T IH0 D",
		"TECHNOLOGY  T EH0 K N AA1 L AH0 JH IY0",
	}, "\n")
	cmu, _ := loadCMU(strings.NewReader(cmuRaw))

	return &cadenceLexicons{
		cmu: cmu,
		conc: map[string]float64{
			"science":         2.90,
			"march":           3.70,
			"stone":           4.90,
			"wall":            4.85,
			"stand":           3.95,
			"rain":            4.90,
			"cut":             4.15,
			"clay":            4.75,
			"bread":           4.92,
			"feed":            3.90,
			"house":           4.95,
			"operational":     1.80,
			"implementation":  1.75,
			"procedural":      1.65,
			"mitigation":      1.55,
			"institutional":   1.90,
			"failure":         1.95,
			"disconfirmation": 1.40,
			"psychologically": 1.65,
			"physiologically": 2.10,
			"stimuli":         2.20,
		},
		aoa: map[string]aoaEntry{
			"science":         {aoa: 7.4, pos: "Noun", lemma: "science"},
			"marches":         {aoa: 6.8, pos: "Verb", lemma: "march"},
			"march":           {aoa: 6.8, pos: "Verb", lemma: "march"},
			"stone":           {aoa: 4.5, pos: "Noun", lemma: "stone"},
			"wall":            {aoa: 3.8, pos: "Noun", lemma: "wall"},
			"stands":          {aoa: 4.2, pos: "Verb", lemma: "stand"},
			"rain":            {aoa: 3.6, pos: "Noun", lemma: "rain"},
			"cuts":            {aoa: 4.4, pos: "Verb", lemma: "cut"},
			"clay":            {aoa: 5.5, pos: "Noun", lemma: "clay"},
			"bread":           {aoa: 3.7, pos: "Noun", lemma: "bread"},
			"feeds":           {aoa: 4.8, pos: "Verb", lemma: "feed"},
			"house":           {aoa: 3.5, pos: "Noun", lemma: "house"},
			"operational":     {aoa: 12.1, pos: "Adjective", lemma: "operational"},
			"implementation":  {aoa: 13.0, pos: "Noun", lemma: "implementation"},
			"procedural":      {aoa: 13.4, pos: "Adjective", lemma: "procedural"},
			"mitigation":      {aoa: 14.2, pos: "Noun", lemma: "mitigation"},
			"institutional":   {aoa: 12.6, pos: "Adjective", lemma: "institutional"},
			"failure":         {aoa: 8.6, pos: "Noun", lemma: "failure"},
			"requires":        {aoa: 9.8, pos: "Verb", lemma: "require"},
			"demonstrated":    {aoa: 10.6, pos: "Verb", lemma: "demonstrate"},
			"psychologically": {aoa: 12.8, pos: "Adverb", lemma: "psychological"},
			"physiologically": {aoa: 13.5, pos: "Adverb", lemma: "physiological"},
		},
		poly: map[string]int{
			"science":        2,
			"march":          14,
			"stone":          13,
			"wall":           9,
			"stand":          24,
			"rain":           4,
			"cut":            61,
			"clay":           5,
			"bread":          2,
			"feed":           12,
			"house":          14,
			"operational":    4,
			"implementation": 2,
			"procedural":     1,
			"mitigation":     2,
			"institutional":  2,
			"failure":        7,
			"psychological":  1,
			"physiological":  1,
		},
	}
}

func TestCadencePunchVsMud(t *testing.T) {
	lx := mockCadenceLexicons()

	punch := measureCadence("Science marches on.", lx)
	if !punch.ParaEndStr {
		t.Fatalf("expected 'Science marches on.' to end on a stressed masculine beat")
	}
	if punch.MaxWeak > 1 {
		t.Fatalf("expected MaxWeak <= 1 on 'Science marches on.', got %d", punch.MaxWeak)
	}
	if punch.StrPct < 55.0 {
		t.Fatalf("expected StrPct >= 55%% on 'Science marches on.', got %.1f%%", punch.StrPct)
	}

	mud := measureCadence("The operational implementation of procedural safeguards for the mitigation of institutional failure requires disconfirmation.", lx)
	crisp := measureCadence("The stone wall stands; rain cuts clay, and bread feeds the house.", lx)

	if crisp.Snap <= mud.Snap+30.0 {
		t.Fatalf("expected crisp concrete pivoted prose (%.1f) to beat bureaucratic mud (%.1f) by >30 points", crisp.Snap, mud.Snap)
	}
	if mud.Mud100 <= 0 {
		t.Fatalf("expected abstract prepositional mud chains > 0 on bureaucratic sentence, got %.2f", mud.Mud100)
	}
	if crisp.Mud100 != 0 {
		t.Fatalf("expected 0 mud chains on concrete sentence, got %.2f", crisp.Mud100)
	}
	if crisp.Pivot100 <= 0 {
		t.Fatalf("expected balanced pivots > 0 on semicolon/conjunction sentence, got %.2f", crisp.Pivot100)
	}
}

func TestCadenceFeminineEndingVsSetupPunch(t *testing.T) {
	lx := mockCadenceLexicons()

	fem := measureCadence("The stone wall is unaffected.", lx)
	if fem.ParaEndStr {
		t.Fatalf("expected 'unaffected.' to have a weak/feminine ending")
	}

	quip := measureCadence(
		"Numerous studies have demonstrated that men are more psychologically and physiologically responsive to visual sexually arousing stimuli. Science marches on.",
		lx,
	)
	if !quip.ParaEndStr {
		t.Fatalf("expected quip paragraph to end masculine")
	}
	if !quip.ShortPunch {
		t.Fatalf("expected setup + 'Science marches on.' to trigger ShortPunch")
	}
}

func TestExtractCadenceParasTracksHeadings(t *testing.T) {
	md := "# Chapter 1\n\n## First Section\n\nFirst paragraph line one\nline two.\n\n> Quoted line.\n\n### Sub\n\nSecond paragraph.\n"
	ps := extractCadenceParas("01_ch01.md", md)
	if len(ps) != 3 {
		t.Fatalf("got %d paragraphs, want 3", len(ps))
	}
	if ps[0].Heading != "First Section" || !strings.Contains(ps[0].Text, "line two.") {
		t.Fatalf("unexpected para 0: %+v", ps[0])
	}
	if ps[2].Heading != "Sub" {
		t.Fatalf("unexpected para 2 heading: %q", ps[2].Heading)
	}
}
