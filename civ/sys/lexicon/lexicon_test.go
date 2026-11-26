package lexicon

import (
	"reflect"
	"strings"
	"testing"
)

func TestInflectAndArticles(t *testing.T) {
	b := &Bundle{
		IrregPlural: cloneMap(builtinIrregPlural),
		IrregPast:   cloneMap(builtinIrregPast),
		IrregPart:   cloneMap(builtinIrregPart),
		IrregComp:   cloneMap(builtinIrregComp),
		IrregSuper:  cloneMap(builtinIrregSuper),
	}
	cases := [][3]string{
		{"child", "NNS", "children"},
		{"city", "NNS", "cities"},
		{"strike", "VBD", "struck"},
		{"write", "VBN", "written"},
		{"make", "VBG", "making"},
		{"carry", "VBZ", "carries"},
		{"good", "JJR", "better"},
		{"clear", "JJS", "clearest"},
	}
	for _, tc := range cases {
		if got := b.Inflect(tc[0], tc[1]); got != tc[2] {
			t.Errorf("Inflect(%q, %q) = %q, want %q", tc[0], tc[1], got, tc[2])
		}
	}
	if got := PreserveCase("Broken", "shattered"); got != "Shattered" {
		t.Errorf("PreserveCase = %q, want Shattered", got)
	}
	fixed := FixArticles([]string{"A", "honest", "man", "built", "an", "bridge"})
	want := []string{"An", "honest", "man", "built", "a", "bridge"}
	if !reflect.DeepEqual(fixed, want) {
		t.Errorf("FixArticles = %v, want %v", fixed, want)
	}
}

func TestFallbackAlternatives(t *testing.T) {
	b := &Bundle{
		IrregPlural:  cloneMap(builtinIrregPlural),
		IrregPast:    cloneMap(builtinIrregPast),
		IrregPart:    cloneMap(builtinIrregPart),
		Concreteness: map[string]float64{"forge": 4.2, "craft": 3.9},
		AoA:          map[string]AoAEntry{},
		SenseCount:   map[string]int{},
		CMU:          map[string][]string{},
	}
	alts := b.Alternatives("build", "Built", "VBD", 5)
	if len(alts) == 0 {
		t.Fatal("expected fallback alternatives for build/VBD")
	}
	foundPast := false
	for _, a := range alts {
		if a.Surface == "Raised" || a.Surface == "Framed" || a.Surface == "Shaped" {
			foundPast = true
		}
	}
	if !foundPast {
		t.Fatalf("expected capitalized past-tense alternative, got %+v", alts)
	}
}

func TestNoAdjacentRepeatsInPhraseAlternatives(t *testing.T) {
	if !HasAdjacentRepeat("Carrying living living from generation to generation.") {
		t.Fatal("expected HasAdjacentRepeat to detect 'living living'")
	}
	if HasAdjacentRepeat("Carrying living truth from generation to generation.") {
		t.Fatal("unexpected adjacent repeat on clean sentence")
	}
	b := &Bundle{
		IrregPlural:  cloneMap(builtinIrregPlural),
		IrregPast:    cloneMap(builtinIrregPast),
		IrregPart:    cloneMap(builtinIrregPart),
		Concreteness: map[string]float64{},
		AoA:          map[string]AoAEntry{},
		SenseCount:   map[string]int{},
		CMU:          map[string][]string{},
	}
	// Even with overlapping fallback synonyms, no candidate may contain adjacent repeats.
	cands := b.PhraseAlternatives("living life", []string{"JJ", "NN"}, []string{"living", "life"}, 12)
	for _, c := range cands {
		if HasAdjacentRepeat(c.Phrase) {
			t.Fatalf("phrase alternative contained adjacent repeat: %+v", c)
		}
	}
}

func TestSenseToggleAndBoostDeweightSteering(t *testing.T) {
	b := &Bundle{
		IrregPlural:  cloneMap(builtinIrregPlural),
		IrregPast:    cloneMap(builtinIrregPast),
		IrregPart:    cloneMap(builtinIrregPart),
		Concreteness: map[string]float64{},
		AoA:          map[string]AoAEntry{},
		SenseCount:   map[string]int{},
		CMU:          map[string][]string{},
	}
	senses := b.Senses("living", "NN", 4)
	if len(senses) < 3 {
		t.Fatalf("expected >=3 fallback senses for 'living', got %+v", senses)
	}
	// Disable the "animation" sense (-103) and de-weight "firm", boost "steady".
	var animOff int
	for _, s := range senses {
		for _, w := range s.Synonyms {
			if w == "animation" {
				animOff = s.Offset
			}
		}
	}
	if animOff == 0 {
		t.Fatalf("expected to find animation sense in %+v", senses)
	}
	opts := PhraseOpts{
		DisabledSenses: map[int]bool{animOff: true},
		DeweightTerms:  []string{"firm", "animation"},
		BoostTerms:     []string{"steady"},
		Expand2Hop:     true,
	}
	cands := b.PhraseAlternativesWithOpts("stable living", []string{"JJ", "NN"}, []string{"stable", "living"}, 16, opts)
	if len(cands) == 0 {
		t.Fatal("expected phrase alternatives for 'stable living'")
	}
	hasSteady := false
	for _, c := range cands {
		low := strings.ToLower(c.Phrase)
		if strings.Contains(low, "animation") || strings.Contains(low, "firm") {
			t.Fatalf("disabled/de-weighted word leaked into candidate: %+v", c)
		}
		if strings.Contains(low, "steady") || strings.Contains(low, "settled") {
			hasSteady = true
		}
	}
	if !hasSteady {
		t.Fatalf("expected steady/settled phrase option, got %+v", cands)
	}
}

func TestEtymologyClassificationAndScoring(t *testing.T) {
	b := &Bundle{
		IrregPast:    cloneMap(builtinIrregPast),
		Concreteness: map[string]float64{},
		AoA:          map[string]AoAEntry{},
		CMU:          map[string][]string{},
		EtymMap:      map[string]Etymology{},
	}
	cases := []struct {
		word string
		want Etymology
	}{
		{"hearth", EtymAngloSaxon},
		{"steady", EtymAngloSaxon},
		{"livelihood", EtymAngloSaxon},
		{"strength", EtymAngloSaxon},
		{"philosophy", EtymGreek},
		{"rhythm", EtymGreek},
		{"character", EtymGreek},
		{"method", EtymGreek},
		{"implementation", EtymLatinate},
		{"demonstration", EtymLatinate},
		{"acquisition", EtymLatinate},
		{"animation", EtymLatinate},
	}
	for _, tc := range cases {
		if got := b.ClassifyEtymology(tc.word); got != tc.want {
			t.Errorf("ClassifyEtymology(%q) = %v, want %v", tc.word, got, tc.want)
		}
	}
	oeScore, oeDom := b.PhraseEtymologyScore("steady livelihood", 6, 4, 7)
	grkScore, grkDom := b.PhraseEtymologyScore("civic ethos", 6, 4, 7)
	latScore, latDom := b.PhraseEtymologyScore("permanent implementation", 6, 4, 7)
	if oeDom != EtymAngloSaxon || oeScore <= 0 {
		t.Fatalf("expected positive OE score for 'steady livelihood', got %.2f (%v)", oeScore, oeDom)
	}
	if grkDom != EtymGreek || grkScore <= 0 {
		t.Fatalf("expected positive Greek score for 'civic ethos', got %.2f (%v)", grkScore, grkDom)
	}
	if latDom != EtymLatinate || latScore >= 0 {
		t.Fatalf("expected negative Latinate score for 'permanent implementation', got %.2f (%v)", latScore, latDom)
	}
}
