package main

import (
	"fmt"
	"strings"
	"unicode"
)

// Words spelled -ise in American English as well as British. An -ise form not
// on this list is flagged (organise, civilisation). "practise" is handled
// with the other whole-word British forms, not as an -ize word.
var iseOK = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`advertise advise appraise apprise arise braise bruise chaise chastise chemise circumcise comprise compromise concise cruise demise despise devise disguise enfranchise enterprise excise exercise expertise franchise guise improvise incise malaise merchandise mortise noise otherwise paradise poise porpoise praise precise premise promise raise reprise revise rise sunrise supervise surmise surprise televise tortoise treatise turquoise uprise valise wise`) {
		iseOK[w] = true
	}
}

// ourBase is a British -our noun. Inflections are generated.
var ourBase = []string{
	"colour", "favour", "honour", "labour", "neighbour", "behaviour",
	"flavour", "humour", "rumour", "vigour", "splendour", "odour",
	"armour", "endeavour", "harbour", "savour", "clamour", "glamour",
	"rancour", "valour", "vapour", "candour",
}

// wholeWord maps a British spelling to the American one, for forms that are
// not a regular -our or -ise inflection.
var wholeWord = map[string]string{
	"programme": "program", "programmes": "programs",
	"grey": "gray", "greys": "grays",
	"judgement": "judgment", "judgements": "judgments",
	"practise": "practice", "practised": "practiced", "practising": "practicing",
	"analyse": "analyze", "analysed": "analyzed", "analysing": "analyzing", "analyses": "analyzes",
	"defence": "defense", "defences": "defenses",
	"offence": "offense", "offences": "offenses",
	"pretence": "pretense", "pretences": "pretenses",
	"licence": "license", "licences": "licenses",
	"centre": "center", "centres": "centers", "centred": "centered", "centring": "centering",
	"metre": "meter", "metres": "meters",
	"litre": "liter", "litres": "liters",
	"fibre": "fiber", "fibres": "fibers",
	"theatre": "theater", "theatres": "theaters",
	"calibre": "caliber", "calibres": "calibers",
	"spectre": "specter", "spectres": "specters",
	"sombre": "somber",
	"lustre": "luster",
	"meagre": "meager",
	"sabre":  "saber", "sabres": "sabers",
	"mitre":      "miter",
	"centimetre": "centimeter", "centimetres": "centimeters",
	"kilometre": "kilometer", "kilometres": "kilometers",
	"millimetre": "millimeter", "millimetres": "millimeters",
}

func spellingHits(file string, line int, text string) []hit {
	var out []hit
	for _, w := range wordsOf(outsideQuotes(text)) {
		w = strings.Trim(w, "'-")
		// An acronym or a name with interior capitals is left alone.
		// A capitalized sentence opening ("Colour") is still checked.
		if w == "" || (w != strings.ToLower(w) && hasInteriorUpper(w)) {
			continue
		}
		low := strings.ToLower(w)
		if am, ok := british(low); ok {
			out = append(out, hit{file, line, "spelling", fmt.Sprintf("%s (American: %s)", w, matchCase(w, am))})
		}
	}
	return out
}

func hasInteriorUpper(w string) bool {
	first := true
	for _, r := range w {
		if unicode.IsUpper(r) && !first {
			return true
		}
		first = false
	}
	return false
}

func british(low string) (string, bool) {
	if am, ok := wholeWord[low]; ok {
		return am, true
	}
	if am, ok := ourForm(low); ok {
		return am, true
	}
	if am, ok := iseForm(low); ok {
		return am, true
	}
	return "", false
}

func ourForm(w string) (string, bool) {
	if am, ok := ourStemForm(w); ok {
		return am, true
	}
	for _, p := range []string{"un", "dis", "mis"} {
		if rest, ok := strings.CutPrefix(w, p); ok {
			if am, ok := ourStemForm(rest); ok {
				return p + am, true
			}
		}
	}
	return "", false
}

func ourStemForm(w string) (string, bool) {
	for _, base := range ourBase {
		stem := strings.TrimSuffix(base, "our")
		switch {
		case w == base:
			return stem + "or", true
		case w == base+"s":
			return stem + "ors", true
		case w == base+"ed":
			return stem + "ored", true
		case w == base+"ing":
			return stem + "oring", true
		case w == base+"er":
			return stem + "orer", true
		case w == base+"ers":
			return stem + "orers", true
		case w == base+"ful":
			return stem + "orful", true
		case w == base+"able":
			return stem + "orable", true
		case w == base+"ably":
			return stem + "orably", true
		case w == base+"ite":
			return stem + "orite", true
		case w == base+"ites":
			return stem + "orites", true
		case w == base+"hood":
			return stem + "orhood", true
		case w == base+"ly":
			return stem + "orly", true
		case w == base+"al":
			return stem + "oral", true
		}
	}
	return "", false
}

func iseForm(w string) (string, bool) {
	lemma, suf, ok := iseLemma(w)
	if !ok || iseAllowed(lemma) || len(lemma) < 6 {
		return "", false
	}
	am := strings.TrimSuffix(lemma, "ise") + "ize"
	switch suf {
	case "":
		return am, true
	case "d":
		return am + "d", true
	case "s":
		return am + "s", true
	case "ing":
		return strings.TrimSuffix(am, "e") + "ing", true
	case "ation":
		return strings.TrimSuffix(am, "e") + "ation", true
	case "ations":
		return strings.TrimSuffix(am, "e") + "ations", true
	default:
		return "", false
	}
}

// iseAllowed reports an -ise word that American spelling keeps: the shared
// list, a -wise compound (likewise), a hyphenated compound whose final element
// is allowed (barn-raising), or that word behind a prefix (unsupervised).
func iseAllowed(lemma string) bool {
	if iseOK[lemma] || strings.HasSuffix(lemma, "wise") {
		return true
	}
	if i := strings.LastIndexByte(lemma, '-'); i >= 0 {
		return iseAllowed(lemma[i+1:])
	}
	for _, p := range []string{"under", "over", "ultra", "semi", "non", "dis", "mis", "pre", "out", "un", "re", "en"} {
		if rest, ok := strings.CutPrefix(lemma, p); ok && len(rest) >= 4 && iseAllowed(rest) {
			return true
		}
	}
	return false
}

func iseLemma(w string) (lemma, suf string, ok bool) {
	switch {
	case strings.HasSuffix(w, "isations"):
		return strings.TrimSuffix(w, "ations") + "e", "ations", true
	case strings.HasSuffix(w, "isation"):
		return strings.TrimSuffix(w, "ation") + "e", "ation", true
	case strings.HasSuffix(w, "ising"):
		return strings.TrimSuffix(w, "ing") + "e", "ing", true
	case strings.HasSuffix(w, "ised"):
		return strings.TrimSuffix(w, "d"), "d", true
	case strings.HasSuffix(w, "ises"):
		return strings.TrimSuffix(w, "s"), "s", true
	case strings.HasSuffix(w, "ise"):
		return w, "", true
	default:
		return "", "", false
	}
}

func matchCase(got, american string) string {
	if got == strings.ToUpper(got) {
		return strings.ToUpper(american)
	}
	if len(got) > 0 && unicode.IsUpper([]rune(got)[0]) {
		r := []rune(american)
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
	return american
}
