package lexicon

import (
	"os"
	"path/filepath"
	"strings"
)

// Etymology identifies the primary historical lexical stratum of an English word.
type Etymology uint8

const (
	EtymNeutral Etymology = iota
	EtymAngloSaxon
	EtymGreek
	EtymLatinate
)

// Short returns a compact UI badge ("OE", "Grk", "Lat", or "").
func (e Etymology) Short() string {
	switch e {
	case EtymAngloSaxon:
		return "OE"
	case EtymGreek:
		return "Grk"
	case EtymLatinate:
		return "Lat"
	default:
		return ""
	}
}

func (b *Bundle) loadEtymologyTSV(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		w, tag, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		w = strings.ToLower(strings.TrimSpace(w))
		switch strings.ToLower(strings.TrimSpace(tag)) {
		case "anglo", "oe", "germanic", "old_english":
			b.EtymMap[w] = EtymAngloSaxon
		case "greek", "grk":
			b.EtymMap[w] = EtymGreek
		case "latin", "lat", "latinate", "french", "norman":
			b.EtymMap[w] = EtymLatinate
		}
	}
}

// ClassifyEtymology classifies a word as Anglo-Saxon (Old English/Germanic),
// Greek, Latinate (Latin/Norman-French), or Neutral using:
//  1. Optional ext-ref/lexicon/etymology.tsv override table
//  2. High-precision Greek morpheme/orthography patterns
//  3. Latinate prefix/suffix morphology + polysyllabic/late-AoA signals
//  4. Anglo-Saxon strong verbs, Germanic orthography, and early-AoA short stems
func (b *Bundle) ClassifyEtymology(word string) Etymology {
	lw := strings.ToLower(strings.Trim(word, ".,;:!?\"'()[]—-"))
	if len(lw) < 2 {
		return EtymNeutral
	}
	if b != nil && len(b.EtymMap) > 0 {
		if e, ok := b.EtymMap[lw]; ok {
			return e
		}
	}
	if greekExact[lw] {
		return EtymGreek
	}
	if latinateExact[lw] {
		return EtymLatinate
	}
	if angloExact[lw] {
		return EtymAngloSaxon
	}
	if hasGreekMorpheme(lw) {
		return EtymGreek
	}
	if b != nil {
		if _, ok := b.IrregPast[lw]; ok {
			return EtymAngloSaxon
		}
	}

	// Latinate suffix / prefix morphology.
	if hasLatinateMorphology(lw) {
		return EtymLatinate
	}

	// Germanic orthographic & suffix markers.
	if hasAngloMorphology(lw) {
		return EtymAngloSaxon
	}

	// Lexical statistical fallback using local AoA, concreteness, and syllable count.
	if b != nil {
		syl, _ := b.Prosody(lw)
		aoa := b.AoA[lw].AoA
		conc := b.Concreteness[lw]
		if syl >= 3 && (aoa >= 8.5 || (aoa == 0 && len(lw) >= 9)) {
			return EtymLatinate
		}
		if syl <= 2 && aoa > 0 && aoa <= 6.2 && (conc >= 2.8 || syl == 1) {
			return EtymAngloSaxon
		}
	}
	return EtymNeutral
}

// PhraseEtymologyScore computes the mean etymology preference score over content
// words in text along with the dominant stratum badge ("OE", "Grk", "Lat", "").
func (b *Bundle) PhraseEtymologyScore(text string, angloBoost, greekBoost, latinPenalty float64) (float64, Etymology) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return 0, EtymNeutral
	}
	var sum float64
	var cnt, nOE, nGrk, nLat int
	for _, f := range fields {
		w := strings.ToLower(strings.Trim(f, ".,;:!?\"'()[]—-"))
		if len(w) < 3 || skipSwapWord[w] || etymFuncWord[w] {
			continue
		}
		e := b.ClassifyEtymology(w)
		cnt++
		switch e {
		case EtymAngloSaxon:
			sum += angloBoost
			nOE++
		case EtymGreek:
			sum += greekBoost
			nGrk++
		case EtymLatinate:
			sum -= latinPenalty
			nLat++
		}
	}
	if cnt == 0 {
		return 0, EtymNeutral
	}
	dom := EtymNeutral
	switch {
	case nGrk > 0 && nGrk >= nOE && nGrk >= nLat:
		dom = EtymGreek
	case nLat > 0 && nLat >= nOE && nLat > nGrk:
		dom = EtymLatinate
	case nOE > 0:
		dom = EtymAngloSaxon
	}
	return sum / float64(cnt), dom
}

func hasGreekMorpheme(w string) bool {
	if len(w) < 4 {
		return false
	}
	// Classic Greek spelling signatures: ph, rh, ps, pn, mn, pt (initial), ch+vowel/r on Greek stems.
	if strings.Contains(w, "ph") && !strings.HasPrefix(w, "uphold") && !strings.HasPrefix(w, "sheph") {
		return true
	}
	for _, p := range greekPrefixes {
		if strings.HasPrefix(w, p) && len(w) > len(p)+2 {
			return true
		}
	}
	for _, s := range greekSuffixes {
		if strings.HasSuffix(w, s) && len(w) > len(s)+2 {
			return true
		}
	}
	for _, sub := range greekRoots {
		if strings.Contains(w, sub) {
			return true
		}
	}
	return false
}

func hasLatinateMorphology(w string) bool {
	if len(w) < 6 {
		return false
	}
	for _, suf := range latinateSuffixes {
		if strings.HasSuffix(w, suf) && len(w) > len(suf)+2 {
			return true
		}
	}
	if len(w) >= 7 {
		for _, pre := range latinatePrefixes {
			if strings.HasPrefix(w, pre) && len(w) > len(pre)+3 {
				return true
			}
		}
	}
	return false
}

func hasAngloMorphology(w string) bool {
	if len(w) < 3 {
		return false
	}
	for _, pre := range []string{"kn", "wr", "wh", "sh", "th", "tw", "sw", "dw"} {
		if strings.HasPrefix(w, pre) {
			return true
		}
	}
	if strings.Contains(w, "gh") || strings.Contains(w, "ck") || strings.Contains(w, "ow") || strings.Contains(w, "ee") || strings.Contains(w, "oo") {
		if len(w) <= 8 {
			return true
		}
	}
	for _, suf := range []string{"hood", "ship", "dom", "ward", "wise", "ful", "less", "stead", "worth", "craft"} {
		if strings.HasSuffix(w, suf) && len(w) > len(suf)+2 {
			return true
		}
	}
	return false
}

func etymologyTSVPath(root string) string {
	return filepath.Join(root, "ext-ref", "lexicon", "etymology.tsv")
}

var etymFuncWord = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "from": true, "into": true,
	"that": true, "this": true, "these": true, "those": true, "they": true, "them": true,
	"their": true, "our": true, "your": true, "his": true, "her": true, "its": true,
	"but": true, "nor": true, "yet": true, "than": true, "then": true, "when": true,
	"where": true, "what": true, "which": true, "who": true, "whom": true, "whose": true,
}

var greekPrefixes = []string{
	"anti", "arch", "auto", "bio", "chron", "cosm", "crypt", "cycl", "democra", "demagog", "demograph",
	"dyn", "eccles", "epi", "eth", "eu", "geo", "graph", "hemi", "hetero",
	"hier", "homo", "hydr", "hyper", "hypo", "icon", "idio", "iso",
	"macro", "mech", "mega", "meta", "micro", "mono", "morph", "myth", "neo",
	"olig", "ortho", "paleo", "pan", "para", "path", "peri",
	"phil", "phon", "phot", "phys", "pneum", "poly", "pragm", "prot",
	"pseud", "psych", "rhet", "rhythm", "schem", "schol", "scop", "soph",
	"sym", "syn", "syst", "tech", "tele", "theo", "therm", "tyran",
}

var greekSuffixes = []string{
	"logy", "logical", "graphy", "graphic", "cracy", "crat", "archy", "arch",
	"metry", "meter", "pathy", "pathic", "scope", "scopy", "gamy", "genesis",
	"gram", "phile", "phobia", "phone", "polis", "polity", "thesis", "sis",
	"dox", "doxy", "nomy", "nomic", "mancy", "morphism", "plasm", "stasis",
}

var greekRoots = []string{
	"chron", "cosmos", "dogma", "drama", "echo", "energ", "epoch", "ethos",
	"hero", "idea", "logic", "method", "music", "myth", "period", "poet",
	"politic", "practic", "problem", "program", "scene", "school", "sphere",
	"stigma", "story", "strateg", "symbol", "system", "telos", "theor",
	"topic", "traged", "tyrant", "zeal",
}

var greekExact = map[string]bool{
	"chaos": true, "character": true, "chorus": true, "civic": false, "climate": true,
	"crisis": true, "critic": true, "criterion": true, "criteria": true, "cycle": true,
	"demon": true, "dogma": true, "drama": true, "echo": true, "energy": true,
	"epoch": true, "ethics": true, "ethos": true, "harmony": true, "heresy": true,
	"hero": true, "history": true, "hubris": true, "hymn": true, "idea": true,
	"idiom": true, "idol": true, "irony": true, "logic": true, "logos": true,
	"lyric": true, "martyr": true, "melody": true, "method": true, "music": true,
	"mystery": true, "myth": true, "nemesis": true, "oikos": true, "oracle": true,
	"organ": true, "panic": true, "paradigm": true, "paradox": true, "pathos": true,
	"period": true, "phase": true, "pneuma": true, "poem": true, "poet": true,
	"policy": true, "polis": true, "polity": true, "practice": true, "praxis": true,
	"problem": true, "prophet": true, "prose": false, "psyche": true, "rhetoric": true,
	"rhythm": true, "scene": true, "scheme": true, "school": true, "skeptic": true,
	"sophist": true, "sphere": true, "stigma": true, "stoic": true, "strategy": true,
	"style": true, "symbol": true, "sympathy": true, "synod": true, "syntax": true,
	"system": true, "tactic": true, "talent": true, "techne": true, "technique": true,
	"telos": true, "theme": true, "theory": true, "thesis": true, "throne": true,
	"tone": true, "topic": true, "tragedy": true, "type": true, "tyrant": true,
	"zeal": true, "zone": true,
}

var latinateSuffixes = []string{
	"tion", "sion", "ation", "ition", "ment", "ity", "ety", "ance", "ence",
	"ancy", "ency", "ious", "eous", "uous", "ous", "ative", "itive", "ive",
	"able", "ible", "ial", "ual", "ary", "ory", "ate", "ify", "itude",
	"ure", "ent", "ant", "ism", "ist",
}

var latinatePrefixes = []string{
	"circum", "contra", "counter", "extra", "infra", "inter", "intra",
	"multi", "omni", "post", "pre", "retro", "semi", "super", "supra",
	"trans", "ultra", "con", "com", "col", "cor", "dis", "sub", "per",
}

var latinateExact = map[string]bool{
	"acquire": true, "acquisition": true, "adequate": true, "administer": true,
	"animation": true, "appropriate": true, "attain": true, "authority": true,
	"benefit": true, "capacity": true, "commence": true, "community": true,
	"conceive": true, "concept": true, "condition": true, "conduct": true,
	"consist": true, "constitute": true, "construct": true, "contain": true,
	"continue": true, "create": true, "culture": true, "decide": true,
	"define": true, "demonstrate": true, "depend": true, "describe": true,
	"destroy": true, "determine": true, "diminish": true, "doctrine": true,
	"eliminate": true, "enable": true, "endeavor": true, "ensure": true,
	"establish": true, "evaluate": true, "evidence": true, "exhibit": true,
	"exist": true, "existence": true, "experience": true, "explain": true,
	"facilitate": true, "factor": true, "firm": true, "form": true,
	"function": true, "generate": true, "generation": true, "govern": true,
	"implement": true, "indicate": true, "individual": true, "induce": true,
	"institution": true, "involve": true, "justify": true, "maintain": true,
	"manner": true, "modify": true, "nature": true, "necessary": true,
	"obtain": true, "occur": true, "operate": true, "order": true,
	"ordinary": true, "perceive": true, "permanent": true, "permit": true,
	"person": true, "possess": true, "practice": false, "preserve": true,
	"principle": true, "procedure": true, "process": true, "produce": true,
	"protect": true, "provide": true, "public": true, "real": true,
	"reality": true, "reason": true, "reduce": true, "refer": true,
	"reflect": true, "relate": true, "remain": true, "represent": true,
	"require": true, "result": true, "retain": true, "satisfy": true,
	"secure": true, "serve": true, "society": true, "solid": true,
	"stable": true, "state": true, "structure": true, "substance": true,
	"sufficient": true, "suggest": true, "support": true, "sustain": true,
	"terminate": true, "utilize": true, "value": true, "verify": true,
}

var angloExact = map[string]bool{
	"age": false, "ask": true, "bare": true, "bear": true, "begin": true,
	"bind": true, "birth": true, "blood": true, "body": true, "bold": true,
	"bone": true, "book": true, "born": true, "bread": true, "break": true,
	"breath": true, "bridge": true, "bright": true, "bring": true, "broad": true,
	"brother": true, "build": true, "builder": true, "burn": true, "call": true,
	"care": true, "child": true, "clean": true, "cold": true, "come": true,
	"craft": true, "cut": true, "daily": true, "dark": true, "daughter": true,
	"day": true, "dead": true, "deal": true, "death": true, "deed": true,
	"deep": true, "draw": true, "drink": true, "drive": true, "dwell": true,
	"earth": true, "eat": true, "end": true, "enough": true, "even": true,
	"everyday": true, "eye": true, "fall": true, "fast": true, "father": true,
	"fear": true, "feed": true, "feel": true, "field": true, "fight": true,
	"fill": true, "find": true, "fire": true, "first": true, "fit": true,
	"flesh": true, "fold": true, "folk": true, "follow": true, "food": true,
	"foot": true, "forbid": true, "forget": true, "forgive": true, "free": true,
	"freedom": true, "friend": true, "full": true, "gather": true, "get": true,
	"give": true, "glad": true, "go": true, "god": true, "good": true,
	"great": true, "ground": true, "grow": true, "growth": true, "hand": true,
	"hang": true, "happen": true, "hard": true, "harm": true, "harvest": true,
	"head": true, "heal": true, "health": true, "hear": true, "heart": true,
	"hearth": true, "heat": true, "heavy": true, "height": true, "help": true,
	"hide": true, "high": true, "hold": true, "home": true, "hope": true,
	"house": true, "household": true, "hunger": true, "hurt": true, "iron": true,
	"keep": true, "kin": true, "kind": true, "king": true, "know": true,
	"land": true, "last": true, "lasting": true, "law": true, "lay": true,
	"lead": true, "learn": true, "learning": true, "leave": true, "length": true,
	"life": true, "lifetime": true, "lift": true, "light": true, "like": true,
	"live": true, "livelihood": true, "living": true, "load": true, "long": true,
	"look": true, "lord": true, "lose": true, "loss": true, "love": true,
	"low": true, "make": true, "man": true, "mark": true, "mean": true,
	"meet": true, "might": true, "mind": true, "mother": true, "mouth": true,
	"naked": true, "name": true, "narrow": true, "near": true, "need": true,
	"neighbor": true, "new": true, "night": true, "oath": true, "old": true,
	"open": true, "own": true, "path": true, "plow": true, "rain": true,
	"raise": true, "reach": true, "read": true, "ready": true, "rest": true,
	"rich": true, "ride": true, "right": true, "ring": true, "rise": true,
	"road": true, "roof": true, "room": true, "root": true, "rough": true,
	"run": true, "sake": true, "salt": true, "same": true, "say": true,
	"sea": true, "see": true, "seed": true, "seek": true, "seem": true,
	"self": true, "sell": true, "send": true, "set": true, "settle": true,
	"settled": true, "shake": true, "shape": true, "share": true, "sharp": true,
	"shed": true, "shelter": true, "shield": true, "shift": true, "shine": true,
	"ship": true, "short": true, "show": true, "shut": true, "sick": true,
	"side": true, "sight": true, "sing": true, "sink": true, "sister": true,
	"sit": true, "skill": true, "sleep": true, "slow": true, "small": true,
	"smith": true, "smooth": true, "son": true, "song": true, "soon": true,
	"sow": true, "span": true, "speak": true, "speech": true, "speed": true,
	"spell": true, "spend": true, "split": true, "spread": true, "spring": true,
	"stand": true, "star": true, "start": true, "stead": true, "steady": true,
	"steal": true, "steel": true, "step": true, "stick": true, "still": true,
	"stone": true, "stop": true, "storm": true, "stout": true, "strength": true,
	"strike": true, "strong": true, "sun": true, "swear": true, "sweat": true,
	"sweep": true, "sweet": true, "swell": true, "swift": true, "sword": true,
	"take": true, "tale": true, "talk": true, "teach": true, "tear": true,
	"tell": true, "thick": true, "thief": true, "thin": true, "thing": true,
	"think": true, "thirst": true, "thought": true, "thread": true, "threat": true,
	"threshold": true, "thrive": true, "throw": true, "tide": true, "tie": true,
	"tight": true, "till": true, "time": true, "token": true, "tongue": true,
	"tool": true, "tooth": true, "town": true, "trade": true, "tread": true,
	"tree": true, "true": true, "trust": true, "truth": true, "turn": true,
	"understand": true, "wake": true, "walk": true, "wall": true, "want": true,
	"war": true, "ward": true, "warm": true, "warn": true, "wash": true,
	"watch": true, "water": true, "way": true, "weak": true, "wealth": true,
	"wear": true, "weather": true, "weave": true, "wed": true, "week": true,
	"weep": true, "weigh": true, "weight": true, "well": true, "wheel": true,
	"whole": true, "wide": true, "widow": true, "width": true, "wife": true,
	"wild": true, "will": true, "win": true, "wind": true, "winter": true,
	"wisdom": true, "wise": true, "wish": true, "witness": true, "wood": true,
	"word": true, "work": true, "world": true, "worth": true, "worthy": true,
	"wound": true, "write": true, "wrong": true, "year": true, "yield": true,
	"yoke": true, "young": true, "youth": true,
}
