package lexicon

import (
	"strings"
	"unicode"
)

// Inflect inflects a base lemma into the requested Penn Treebank POS form
// (NNS, VBZ, VBD, VBN, VBG, JJR, JJS). Multi-word phrases inflect their head.
func (b *Bundle) Inflect(lemma, ptbPOS string) string {
	lemma = strings.TrimSpace(strings.ToLower(lemma))
	if lemma == "" {
		return ""
	}
	// Multi-word verb phrases ("break down", "give up") inflect the first word;
	// multi-word noun phrases ("social order") inflect the last word.
	if parts := strings.Fields(lemma); len(parts) > 1 {
		if strings.HasPrefix(ptbPOS, "VB") {
			parts[0] = b.Inflect(parts[0], ptbPOS)
		} else if ptbPOS == "NNS" {
			parts[len(parts)-1] = b.Inflect(parts[len(parts)-1], ptbPOS)
		}
		return strings.Join(parts, " ")
	}
	switch ptbPOS {
	case "NNS":
		return b.pluralize(lemma)
	case "VBZ":
		return thirdPersonSingular(lemma)
	case "VBD":
		return b.pastTense(lemma)
	case "VBN":
		return b.pastParticiple(lemma)
	case "VBG":
		return presentParticiple(lemma)
	case "JJR":
		return b.comparative(lemma)
	case "JJS":
		return b.superlative(lemma)
	default:
		return lemma
	}
}

// PreserveCase transfers capitalization style from orig onto replacement.
func PreserveCase(orig, replacement string) string {
	if orig == "" || replacement == "" {
		return replacement
	}
	or := []rune(orig)
	if len(or) > 1 && isAllUpper(or) {
		return strings.ToUpper(replacement)
	}
	if unicode.IsUpper(or[0]) {
		rr := []rune(replacement)
		rr[0] = unicode.ToUpper(rr[0])
		return string(rr)
	}
	return replacement
}

// FixArticles adjusts "a" / "an" in a token sequence when the following word changed.
func FixArticles(words []string) []string {
	out := append([]string(nil), words...)
	for i := 0; i < len(out)-1; i++ {
		lw := strings.ToLower(out[i])
		if lw != "a" && lw != "an" {
			continue
		}
		next := strings.ToLower(strings.TrimLeft(out[i+1], "\"'(["))
		if next == "" {
			continue
		}
		want := "a"
		if needsAn(next) {
			want = "an"
		}
		out[i] = PreserveCase(out[i], want)
	}
	return out
}

func needsAn(next string) bool {
	for _, p := range []string{"honest", "honor", "honour", "hour", "heir"} {
		if strings.HasPrefix(next, p) {
			return true
		}
	}
	for _, p := range []string{"uni", "use", "usu", "eur", "one", "once"} {
		if strings.HasPrefix(next, p) {
			return false
		}
	}
	return strings.ContainsRune("aeiou", rune(next[0]))
}

func isAllUpper(rs []rune) bool {
	hasLetter := false
	for _, r := range rs {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

func (b *Bundle) pluralize(w string) string {
	if b != nil {
		if irr, ok := b.IrregPlural[w]; ok {
			return irr
		}
	}
	switch {
	case strings.HasSuffix(w, "s"), strings.HasSuffix(w, "x"), strings.HasSuffix(w, "z"),
		strings.HasSuffix(w, "ch"), strings.HasSuffix(w, "sh"):
		return w + "es"
	case len(w) > 2 && strings.HasSuffix(w, "y") && !isVowelByte(w[len(w)-2]):
		return w[:len(w)-1] + "ies"
	case strings.HasSuffix(w, "fe") && len(w) > 2:
		return w[:len(w)-2] + "ves"
	default:
		return w + "s"
	}
}

func thirdPersonSingular(w string) string {
	switch w {
	case "be":
		return "is"
	case "have":
		return "has"
	case "do":
		return "does"
	case "go":
		return "goes"
	}
	switch {
	case strings.HasSuffix(w, "s"), strings.HasSuffix(w, "x"), strings.HasSuffix(w, "z"),
		strings.HasSuffix(w, "ch"), strings.HasSuffix(w, "sh"), strings.HasSuffix(w, "o"):
		return w + "es"
	case len(w) > 2 && strings.HasSuffix(w, "y") && !isVowelByte(w[len(w)-2]):
		return w[:len(w)-1] + "ies"
	default:
		return w + "s"
	}
}

func (b *Bundle) pastTense(w string) string {
	if b != nil {
		if irr, ok := b.IrregPast[w]; ok {
			return irr
		}
	}
	return regularEd(w)
}

func (b *Bundle) pastParticiple(w string) string {
	if b != nil {
		if irr, ok := b.IrregPart[w]; ok {
			return irr
		}
		if irr, ok := b.IrregPast[w]; ok {
			return irr
		}
	}
	return regularEd(w)
}

func regularEd(w string) string {
	switch {
	case strings.HasSuffix(w, "e"):
		return w + "d"
	case len(w) > 2 && strings.HasSuffix(w, "y") && !isVowelByte(w[len(w)-2]):
		return w[:len(w)-1] + "ied"
	case shouldDoubleFinalConsonant(w):
		return w + w[len(w)-1:] + "ed"
	default:
		return w + "ed"
	}
}

func presentParticiple(w string) string {
	switch w {
	case "be":
		return "being"
	case "die":
		return "dying"
	case "lie":
		return "lying"
	case "tie":
		return "tying"
	}
	switch {
	case strings.HasSuffix(w, "ie"):
		return w[:len(w)-2] + "ying"
	case strings.HasSuffix(w, "e") && !strings.HasSuffix(w, "ee") && !strings.HasSuffix(w, "ye") && !strings.HasSuffix(w, "oe"):
		return w[:len(w)-1] + "ing"
	case shouldDoubleFinalConsonant(w):
		return w + w[len(w)-1:] + "ing"
	default:
		return w + "ing"
	}
}

func (b *Bundle) comparative(w string) string {
	if b != nil {
		if irr, ok := b.IrregComp[w]; ok {
			return irr
		}
	}
	if approxSyllables(w) >= 3 {
		return "more " + w
	}
	switch {
	case strings.HasSuffix(w, "e"):
		return w + "r"
	case len(w) > 2 && strings.HasSuffix(w, "y") && !isVowelByte(w[len(w)-2]):
		return w[:len(w)-1] + "ier"
	case shouldDoubleFinalConsonant(w):
		return w + w[len(w)-1:] + "er"
	default:
		return w + "er"
	}
}

func (b *Bundle) superlative(w string) string {
	if b != nil {
		if irr, ok := b.IrregSuper[w]; ok {
			return irr
		}
	}
	if approxSyllables(w) >= 3 {
		return "most " + w
	}
	switch {
	case strings.HasSuffix(w, "e"):
		return w + "st"
	case len(w) > 2 && strings.HasSuffix(w, "y") && !isVowelByte(w[len(w)-2]):
		return w[:len(w)-1] + "iest"
	case shouldDoubleFinalConsonant(w):
		return w + w[len(w)-1:] + "est"
	default:
		return w + "est"
	}
}

func isVowelByte(c byte) bool {
	return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u'
}

func shouldDoubleFinalConsonant(w string) bool {
	n := len(w)
	if n < 3 || approxSyllables(w) > 1 {
		return false
	}
	last := w[n-1]
	if isVowelByte(last) || last == 'w' || last == 'x' || last == 'y' {
		return false
	}
	return isVowelByte(w[n-2]) && !isVowelByte(w[n-3])
}

var builtinIrregPlural = map[string]string{
	"man": "men", "woman": "women", "child": "children", "person": "people",
	"foot": "feet", "tooth": "teeth", "mouse": "mice", "goose": "geese",
	"ox": "oxen", "life": "lives", "knife": "knives", "wife": "wives",
	"leaf": "leaves", "half": "halves", "self": "selves", "wolf": "wolves",
	"criterion": "criteria", "phenomenon": "phenomena", "thesis": "theses",
	"basis": "bases", "crisis": "crises", "axis": "axes", "analysis": "analyses",
	"index": "indices", "medium": "media", "datum": "data", "sheep": "sheep",
	"fish": "fish", "series": "series", "species": "species",
}

var builtinIrregPast = map[string]string{
	"be": "was", "have": "had", "do": "did", "say": "said", "go": "went",
	"get": "got", "make": "made", "know": "knew", "think": "thought", "take": "took",
	"see": "saw", "come": "came", "find": "found", "give": "gave", "tell": "told",
	"become": "became", "show": "showed", "leave": "left", "feel": "felt",
	"put": "put", "bring": "brought", "begin": "began", "keep": "kept", "hold": "held",
	"write": "wrote", "stand": "stood", "hear": "heard", "let": "let", "mean": "meant",
	"set": "set", "meet": "met", "run": "ran", "pay": "paid", "sit": "sat",
	"speak": "spoke", "lie": "lay", "lead": "led", "read": "read", "grow": "grew",
	"lose": "lost", "fall": "fell", "send": "sent", "build": "built", "understand": "understood",
	"draw": "drew", "break": "broke", "spend": "spent", "cut": "cut", "rise": "rose",
	"drive": "drove", "buy": "bought", "wear": "wore", "choose": "chose", "seek": "sought",
	"strike": "struck", "bear": "bore", "bind": "bound", "cast": "cast", "cling": "clung",
	"deal": "dealt", "dig": "dug", "feed": "fed", "fight": "fought", "flee": "fled",
	"forbid": "forbade", "forget": "forgot", "forgive": "forgave", "freeze": "froze",
	"hang": "hung", "hide": "hid", "hit": "hit", "hurt": "hurt", "lay": "laid",
	"lend": "lent", "light": "lit", "ride": "rode", "ring": "rang", "sell": "sold",
	"shake": "shook", "shed": "shed", "shine": "shone", "shoot": "shot", "shut": "shut",
	"sing": "sang", "sink": "sank", "sleep": "slept", "slide": "slid", "sow": "sowed",
	"split": "split", "spread": "spread", "spring": "sprang", "steal": "stole", "stick": "stuck",
	"swear": "swore", "sweep": "swept", "swell": "swelled", "swim": "swam", "swing": "swung",
	"teach": "taught", "tear": "tore", "throw": "threw", "tread": "trod", "wake": "woke",
	"weave": "wove", "weep": "wept", "win": "won", "wind": "wound", "withdraw": "withdrew",
	"withhold": "withheld", "withstand": "withstood",
}

var builtinIrregPart = map[string]string{
	"be": "been", "do": "done", "go": "gone", "know": "known", "take": "taken",
	"see": "seen", "come": "come", "give": "given", "become": "become", "show": "shown",
	"begin": "begun", "write": "written", "run": "run", "speak": "spoken", "lie": "lain",
	"grow": "grown", "fall": "fallen", "draw": "drawn", "break": "broken", "rise": "risen",
	"drive": "driven", "wear": "worn", "choose": "chosen", "bear": "borne", "forbid": "forbidden",
	"forget": "forgotten", "forgive": "forgiven", "freeze": "frozen", "hide": "hidden",
	"ride": "ridden", "ring": "rung", "shake": "shaken", "sing": "sung", "sink": "sunk",
	"sow": "sown", "spring": "sprung", "steal": "stolen", "swear": "sworn", "swell": "swollen",
	"swim": "swum", "tear": "torn", "throw": "thrown", "tread": "trodden", "wake": "woken",
	"weave": "woven", "withdraw": "withdrawn",
}

var builtinIrregComp = map[string]string{
	"good": "better", "bad": "worse", "far": "farther", "little": "less", "much": "more", "many": "more",
}

var builtinIrregSuper = map[string]string{
	"good": "best", "bad": "worst", "far": "farthest", "little": "least", "much": "most", "many": "most",
}

// IrregularFallback returns copies of the built-in irregular inflection
// tables (lemma -> inflected) so offline lemmatizers can invert them when the
// WordNet 3.1 dictionary has not been unpacked.
func IrregularFallback() (plural, past, part, comp, super map[string]string) {
	return cloneMap(builtinIrregPlural), cloneMap(builtinIrregPast), cloneMap(builtinIrregPart),
		cloneMap(builtinIrregComp), cloneMap(builtinIrregSuper)
}

var builtinThesaurus = map[string][]string{
	"world":    {"earth", "realm", "life", "age"},
	"come":     {"enter", "arrive", "step", "pass"},
	"naked":    {"bare", "unclothed", "unarmed", "stripped"},
	"build":    {"raise", "frame", "erect", "construct", "shape"},
	"order":    {"law", "rule", "peace", "frame", "structure"},
	"truth":    {"fact", "reality", "verity", "witness"},
	"belief":   {"faith", "creed", "conviction", "trust"},
	"man":      {"person", "citizen", "builder", "mortal"},
	"power":    {"force", "strength", "might", "authority", "rule"},
	"strong":   {"firm", "sound", "stout", "steady", "solid"},
	"weak":     {"frail", "fragile", "brittle", "feeble"},
	"clear":    {"plain", "lucid", "sharp", "distinct", "open"},
	"keep":     {"hold", "guard", "preserve", "maintain", "retain"},
	"break":    {"shatter", "fracture", "snap", "sever", "rupture"},
	"make":     {"form", "forge", "craft", "create", "yield"},
	"show":     {"prove", "reveal", "display", "expose", "mark"},
	"fail":     {"collapse", "falter", "give way", "crumble"},
	"require":  {"demand", "need", "call for", "exact"},
	"depend":   {"rest", "rely", "hang", "lean"},
	"system":   {"frame", "order", "scheme", "engine", "arrangement"},
	"society":  {"commonwealth", "community", "polity", "body"},
	"rule":     {"law", "precept", "maxim", "statute", "measure"},
	"good":     {"sound", "just", "worthy", "firm", "fit"},
	"bad":      {"rotten", "corrupt", "unsound", "ruinous", "grave"},
	"change":   {"shift", "turn", "alter", "bend", "transform"},
	"reason":   {"judgment", "mind", "sense", "logic", "thought"},
	"work":     {"labor", "craft", "toil", "task", "service"},
	"bridge":   {"span", "arch", "crossing", "viaduct"},
	"judge":    {"weigh", "measure", "test", "try", "assess"},
	"protect":  {"guard", "shield", "defend", "shelter", "keep"},
	"destroy":  {"ruin", "wreck", "unmake", "shatter", "level"},
	"begin":    {"open", "start", "launch", "commence", "set out"},
	"end":      {"close", "finish", "cease", "halt", "land"},
}
