// Package linear scores sentences against Rule 10 (linearity, early verb,
// clause depth, given-new flow, active voice) and generates tree-safe
// mechanical structural alternatives from a CoreNLP dependency tree.
package linear

import (
	"fmt"
	"strings"
	"unicode"

	"po/civ/sys/corenlp"
)

// SpanLock marks an inclusive 1-based token span that must not be split or
// lexically altered ("set phrase"). Text anchors the locked wording so edits
// earlier in the sentence never shift a lock onto neighboring words.
type SpanLock struct {
	Lo    int    `json:"lo"`
	Hi    int    `json:"hi"`
	Text  string `json:"text,omitempty"`
	Label string `json:"label,omitempty"`
}

// Finding describes one Rule 10 structural issue on a sentence.
type Finding struct {
	Part   int     `json:"part"`
	Kind   string  `json:"kind"`
	Weight float64 `json:"weight"`
	Note   string  `json:"note"`
}

// Metrics holds Rule 10 linearity metrics for a single sentence.
type Metrics struct {
	Words    int       `json:"words"`
	PreVerb  int       `json:"pre_verb"`
	SubjLen  int       `json:"subj_len"`
	MDD      float64   `json:"mdd"`
	Depth    int       `json:"depth"`
	Spine    string    `json:"spine"`
	Voice    string    `json:"voice"`
	Penalty  float64   `json:"penalty"`
	Score    float64   `json:"score"` // 0..100 (higher = cleaner/more linear)
	Findings []Finding `json:"findings,omitempty"`
}

// StructureOption is one candidate mechanical syntactic structure derived from
// a base sentence tree.
type StructureOption struct {
	ID         string        `json:"id"`
	Kind       string        `json:"kind"` // as_is, pull_back, pull_front, split_mark, split_conj, split_subord, split_colon_subord, extract_aside, swap_coord, front_obl, drop_expletive
	Label      string        `json:"label"`
	Note       string        `json:"note"`
	Text       string        `json:"text"`
	Tree       *corenlp.Tree `json:"-"`
	TokenMap   []int         `json:"token_map,omitempty"` // new 1-based idx -> old 1-based idx (0 if synthetic)
	Metrics    Metrics       `json:"metrics"`
}

// Score computes Rule 10 metrics for a single parsed sentence Tree.
func Score(t *corenlp.Tree) Metrics {
	if t == nil || t.N() == 0 {
		return Metrics{Score: 50}
	}
	var m Metrics
	m.Words = t.Words()
	mv := t.MainVerb()
	if mv > 1 {
		m.PreVerb = t.WordsIn(1, mv-1)
	}
	subj := t.Subject()
	if subj > 0 {
		m.SubjLen = t.WordsIn(t.Lo[subj], t.Hi[subj])
	}
	m.MDD = t.MDD()
	m.Depth = t.MaxDepth()
	m.Spine = Spine(t)
	m.Voice = Voice(t)

	add := func(part int, kind string, w float64, note string) {
		m.Findings = append(m.Findings, Finding{Part: part, Kind: kind, Weight: w, Note: note})
		m.Penalty += w
	}

	switch {
	case m.PreVerb > 22:
		add(2, "late verb", 3, fmt.Sprintf("%d words before main verb", m.PreVerb))
	case m.PreVerb > 15:
		add(2, "late verb", 2, fmt.Sprintf("%d words before main verb", m.PreVerb))
	case m.PreVerb > 10:
		add(2, "late verb", 1, fmt.Sprintf("%d words before main verb", m.PreVerb))
	}
	if m.SubjLen > 14 {
		add(2, "long subject", 2, fmt.Sprintf("%d-word subject", m.SubjLen))
	} else if m.SubjLen > 8 {
		add(2, "long subject", 1, fmt.Sprintf("%d-word subject", m.SubjLen))
	}
	if m.Depth >= 3 {
		add(3, "nested clauses", float64(m.Depth-2), fmt.Sprintf("clauses nest %d deep", m.Depth))
	}
	switch {
	case m.Words > 55:
		add(3, "long sentence", 2, fmt.Sprintf("%d words", m.Words))
	case m.Words > 40:
		add(3, "long sentence", 1, fmt.Sprintf("%d words", m.Words))
	}
	if m.Voice == "passive" {
		add(2, "passive voice", 0.5, "passive main clause")
	} else if m.Voice == "expletive" {
		add(2, "expletive", 0.75, "expletive opener (it is / there is)")
	}
	if m.MDD > 3.6 {
		add(2, "high MDD", 0.75, fmt.Sprintf("mean dependency distance %.2f", m.MDD))
	}

	score := 100.0 - m.Penalty*11.0 - maxFloat(0, m.MDD-2.5)*6.0
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	m.Score = score
	return m
}

// Spine summarizes the top-level clause shape (e.g. "S-V-O", "ADV-S-V-O").
func Spine(t *corenlp.Tree) string {
	if t == nil || t.Root == 0 {
		return ""
	}
	var parts []string
	subj := t.Subject()
	mv := t.MainVerb()
	hasFront := false
	for _, k := range t.Kids[t.Root] {
		if k < mv && k != subj && (t.Rel[k] == "advcl" || t.Rel[k] == "obl" || t.Rel[k] == "advmod") {
			if t.WordsIn(t.Lo[k], t.Hi[k]) >= 3 {
				hasFront = true
			}
		}
	}
	if hasFront {
		parts = append(parts, "FRONT")
	}
	if subj > 0 {
		parts = append(parts, "S")
	}
	parts = append(parts, "V")
	for _, k := range t.Kids[t.Root] {
		if k <= t.Root {
			continue
		}
		switch t.Rel[k] {
		case "obj", "iobj":
			parts = append(parts, "O")
		case "ccomp", "xcomp":
			parts = append(parts, "COMP")
		case "advcl":
			parts = append(parts, "ADVCL")
		case "obl":
			parts = append(parts, "OBL")
		}
	}
	return strings.Join(parts, "-")
}

// Voice classifies the main clause as active, passive, expletive, or nominalization.
func Voice(t *corenlp.Tree) string {
	if t == nil || t.Root == 0 {
		return "active"
	}
	r := t.Root
	for _, k := range t.Kids[r] {
		if t.Rel[k] == "expl" {
			return "expletive"
		}
	}
	for _, k := range t.Kids[r] {
		if t.Rel[k] == "nsubj:pass" || t.Rel[k] == "aux:pass" {
			return "passive"
		}
	}
	return "active"
}

// GivenNewOverlap measures content-lemma overlap between the tail of prev and
// the opening (pre-verb/subject) of cur, returning a value in [0, 1].
func GivenNewOverlap(prev, cur *corenlp.Tree) float64 {
	if prev == nil || cur == nil || prev.N() == 0 || cur.N() == 0 {
		return 0
	}
	prevLemmas := contentSet(prev, 1, prev.N())
	if len(prevLemmas) == 0 {
		return 0
	}
	mv := cur.MainVerb()
	if mv < 3 {
		mv = minInt(cur.N(), 6)
	}
	openLemmas := contentSet(cur, 1, mv)
	if len(openLemmas) == 0 {
		return 0
	}
	hits := 0
	for l := range openLemmas {
		if prevLemmas[l] {
			hits++
		}
	}
	// Also count explicit discourse linkers at sentence start.
	first := strings.ToLower(cur.Toks[1].Word)
	bonus := 0.0
	if linkWords[first] {
		bonus = 0.25
	}
	ratio := float64(hits)/float64(len(openLemmas)) + bonus
	if ratio > 1 {
		ratio = 1
	}
	return ratio
}

// TailOverlap measures how strongly the tail of cur (last 6 non-punct tokens)
// sets up the opening (pre-verb) of next. Used by the backward-pass goal!
func TailOverlap(cur, next *corenlp.Tree) float64 {
	if cur == nil || next == nil || cur.N() == 0 || next.N() == 0 {
		return 0
	}
	nextMV := next.MainVerb()
	if nextMV < 3 {
		nextMV = minInt(next.N(), 6)
	}
	target := contentSet(next, 1, nextMV)
	if len(target) == 0 {
		return 0
	}
	// Last 6 words of cur.
	start := cur.N()
	seenW := 0
	for start > 1 && seenW < 6 {
		if !cur.IsPunct(start) {
			seenW++
		}
		start--
	}
	tail := contentSet(cur, start, cur.N())
	if len(tail) == 0 {
		return 0
	}
	hits := 0
	for l := range tail {
		if target[l] {
			hits++
		}
	}
	return float64(hits) / float64(len(target))
}

func contentSet(t *corenlp.Tree, a, b int) map[string]bool {
	out := map[string]bool{}
	if a < 1 {
		a = 1
	}
	if b > t.N() {
		b = t.N()
	}
	for i := a; i <= b; i++ {
		if t.IsPunct(i) {
			continue
		}
		p := t.Toks[i].POS
		if !(strings.HasPrefix(p, "NN") || strings.HasPrefix(p, "VB") || strings.HasPrefix(p, "JJ")) {
			continue
		}
		l := strings.ToLower(t.Toks[i].Lemma)
		if len(l) < 3 || stopLemmas[l] {
			continue
		}
		out[l] = true
	}
	return out
}

// GenerateStructures mechanically produces tree-safe structural variants from
// a parsed sentence while respecting SpanLocks (no lock is split internally).
func GenerateStructures(t *corenlp.Tree, locks []SpanLock) []StructureOption {
	if t == nil || t.N() == 0 {
		return nil
	}
	baseText := t.Text
	if baseText == "" {
		baseText = t.Surface(1, t.N())
	}
	origMap := make([]int, t.N()+1)
	for i := 1; i <= t.N(); i++ {
		origMap[i] = i
	}
	opts := []StructureOption{{
		ID:       "as_is",
		Kind:     "as_is",
		Label:    "Original structure",
		Note:     fmt.Sprintf("Spine %s · %s voice", Spine(t), Voice(t)),
		Text:     baseText,
		Tree:     t,
		TokenMap: origMap,
		Metrics:  Score(t),
	}}
	seenText := map[string]bool{strings.TrimSpace(baseText): true}

	addOpt := func(id, kind, label, note, text string) {
		text = corenlp.Tidy(text)
		if text == "" || seenText[text] {
			return
		}
		seenText[text] = true
		tr := corenlp.ParseFallback(text)
		opts = append(opts, StructureOption{
			ID:      id,
			Kind:    kind,
			Label:   label,
			Note:    note,
			Text:    text,
			Tree:    tr,
			Metrics: Score(tr),
		})
	}

	r := t.Root
	mv := t.MainVerb()
	subj := t.Subject()
	bodyHi, finalPunct := bodyEndTok(t)

	// 1. Pull back: move an opening adverbial/prepositional phrase before subject to end.
	limit := mv
	if subj > 0 {
		limit = t.Lo[subj]
	}
	for _, k := range t.Kids[r] {
		if t.IsPunct(k) || corenlp.SubjRels[t.Rel[k]] || t.Hi[k] >= limit {
			continue
		}
		switch t.Rel[k] {
		case "advcl", "obl", "advmod":
		default:
			continue
		}
		lo, hi := t.Lo[k], t.Hi[k]
		if t.WordsIn(lo, hi) < 3 || lo > 2 || splitsLock(lo, hi, locks) {
			continue
		}
		restLo := hi + 1
		for restLo <= bodyHi && t.Toks[restLo].Word == "," {
			restLo++
		}
		if restLo > bodyHi {
			continue
		}
		moved := lowerFirst(t.Surface(lo, hi), t.Toks[lo].NER)
		rest := capFirst(t.Surface(restLo, bodyHi))
		addOpt(
			fmt.Sprintf("pull_back_%d", k),
			"pull_back",
			"Subject first (pull opener back)",
			fmt.Sprintf("Moves %d-word opener %q to the end", t.WordsIn(lo, hi), trunc(moved, 32)),
			rest+", "+moved+finalPunct,
		)
	}

	// 2. Pull front: move a trailing adverbial clause or oblique phrase to the front.
	for _, k := range t.Kids[r] {
		if k <= r || (t.Rel[k] != "advcl" && t.Rel[k] != "obl") {
			continue
		}
		lo, hi := t.Lo[k], t.Hi[k]
		for lo < hi && t.Toks[lo].Word == "," {
			lo++
		}
		if t.WordsIn(lo, hi) < 3 || hi < bodyHi-1 || lo <= 2 || splitsLock(lo, hi, locks) {
			continue
		}
		leftHi := lo - 1
		for leftHi >= 1 && t.Toks[leftHi].Word == "," {
			leftHi--
		}
		if leftHi < 2 {
			continue
		}
		front := capFirst(t.Surface(lo, hi))
		rest := lowerFirst(t.Surface(1, leftHi), t.Toks[1].NER)
		lbl := "Front trailing modifier"
		if strings.EqualFold(t.Toks[lo].Word, "because") {
			lbl = "Cause before effect"
		}
		addOpt(
			fmt.Sprintf("pull_front_%d", k),
			"pull_front",
			lbl,
			fmt.Sprintf("Opens with %q", trunc(front, 32)),
			front+", "+rest+finalPunct,
		)
	}

	// 3. Split at colon or semicolon.
	for i := 2; i < bodyHi; i++ {
		w := t.Toks[i].Word
		if w != ":" && w != ";" {
			continue
		}
		if insideAnyLock(i, locks) || t.WordsIn(1, i-1) < 3 || t.WordsIn(i+1, bodyHi) < 3 {
			continue
		}
		left := t.Surface(1, i-1)
		right := capFirst(t.Surface(i+1, bodyHi))
		addOpt(
			fmt.Sprintf("split_mark_%d", i),
			"split_mark",
			fmt.Sprintf("Split at %q", w),
			"One claim per sentence",
			left+". "+right+finalPunct,
		)
	}

	// 4. Split coordinated independent clauses at "and" / "but" / "yet" / "so".
	// Guard against breaking serial lists ("A, B, C, and D") or shared-subject
	// VP coordinations ("tells ..., does ..., and fixes ...").
	hasDepStruct := hasAnyRel(t, "conj") || hasAnyRel(t, "advcl")
	for i := 3; i < bodyHi-2; i++ {
		lw := strings.ToLower(t.Toks[i].Word)
		if t.Toks[i].POS != "CC" || (lw != "and" && lw != "but" && lw != "yet" && lw != "so") {
			continue
		}
		if insideAnyLock(i, locks) || t.WordsIn(1, i-1) < 5 || t.WordsIn(i+1, bodyHi) < 5 {
			continue
		}
		if !canSplitAtConj(t, i, bodyHi, hasDepStruct) {
			continue
		}
		leftHi := i - 1
		if t.Toks[leftHi].Word == "," {
			leftHi--
		}
		left := t.Surface(1, leftHi)
		var right string
		if lw == "and" {
			right = capFirst(t.Surface(i+1, bodyHi))
		} else {
			right = capFirst(t.Surface(i, bodyHi))
		}
		addOpt(
			fmt.Sprintf("split_conj_%d", i),
			"split_conj",
			fmt.Sprintf("Split at %q", lw),
			"Breaks compound sentence into two tighter beats",
			left+". "+right+finalPunct,
		)
	}

	// 4b. Split at a trailing subordinate clause boundary (e.g. ", as each person ...",
	// ", wherever a person ...", ", while ...") when the trailing clause has its own
	// subject and sits cleanly at the tail of the sentence.
	emittedSubord := false
	hasColonLeft := func(hi int) bool {
		for j := 1; j <= hi; j++ {
			if t.Toks[j].Word == ":" {
				return true
			}
		}
		return false
	}
	emitSubordSplit := func(markIdx int) {
		mw := strings.ToLower(t.Toks[markIdx].Word)
		mode, ok := splitSubordMarkers[mw]
		if !ok {
			return
		}
		leftHi := markIdx - 1
		for leftHi >= 1 && t.Toks[leftHi].Word == "," {
			leftHi--
		}
		clauseStart := markIdx + 1
		if leftHi < 2 || clauseStart > bodyHi {
			return
		}
		if insideAnyLock(markIdx, locks) || splitsLock(1, leftHi, locks) || splitsLock(clauseStart, bodyHi, locks) {
			return
		}
		if t.WordsIn(1, leftHi) < 4 || t.WordsIn(clauseStart, bodyHi) < 5 {
			return
		}
		left := t.Surface(1, leftHi)
		rightRaw := t.Surface(clauseStart, bodyHi)
		if rightRaw == "" {
			return
		}
		switch mode {
		case "drop":
			addOpt(
				fmt.Sprintf("split_subord_%d", markIdx),
				"split_subord",
				fmt.Sprintf("Split before %q", mw),
				fmt.Sprintf("Starts a clean second sentence at %q", trunc(capFirst(rightRaw), 28)),
				left+". "+capFirst(rightRaw)+finalPunct,
			)
			emittedSubord = true
			if !hasColonLeft(leftHi) && (t.WordsIn(clauseStart, bodyHi) >= 10 || hasSerialListInSpan(t, clauseStart, bodyHi)) {
				addOpt(
					fmt.Sprintf("split_colon_subord_%d", markIdx),
					"split_colon_subord",
					fmt.Sprintf("Colon lead-in at %q", mw),
					"Turns trailing clause into a crisp colon elaboration",
					left+": "+lowerFirst(rightRaw, t.Toks[clauseStart].NER)+finalPunct,
				)
			}
		case "contrast":
			addOpt(
				fmt.Sprintf("split_subord_%d", markIdx),
				"split_subord",
				fmt.Sprintf("Split contrast at %q", mw),
				"Separates the contrast into its own sentence",
				left+". Yet "+lowerFirst(rightRaw, t.Toks[clauseStart].NER)+finalPunct,
			)
			emittedSubord = true
		case "colon":
			if !hasColonLeft(leftHi) && (t.WordsIn(clauseStart, bodyHi) >= 10 || hasSerialListInSpan(t, clauseStart, bodyHi)) {
				addOpt(
					fmt.Sprintf("split_colon_subord_%d", markIdx),
					"split_colon_subord",
					fmt.Sprintf("Colon lead-in at %q", mw),
					"Turns trailing explanation into a colon clause",
					left+": "+lowerFirst(rightRaw, t.Toks[clauseStart].NER)+finalPunct,
				)
				emittedSubord = true
			}
		}
	}
	seenSubordMark := map[int]bool{}
	for k := 1; k <= bodyHi; k++ {
		if t.Rel[k] != "advcl" && t.Rel[k] != "ccomp" {
			continue
		}
		if !hasDirectSubj(t, k) || !trailingTailSafe(t, k, bodyHi) {
			continue
		}
		lo := t.Lo[k]
		markIdx := lo
		if lo >= 2 && lo < bodyHi && t.Toks[lo].Word == "," {
			markIdx = lo + 1
		} else if lo < 3 || t.Toks[lo-1].Word != "," {
			continue
		}
		if seenSubordMark[markIdx] {
			continue
		}
		seenSubordMark[markIdx] = true
		emitSubordSplit(markIdx)
	}
	if !emittedSubord && !hasAnyRel(t, "advcl") {
		for i := maxInt(r+1, 3); i < bodyHi-4; i++ {
			if t.Toks[i-1].Word != "," {
				continue
			}
			if _, ok := splitSubordMarkers[strings.ToLower(t.Toks[i].Word)]; !ok {
				continue
			}
			if !fallbackStartsClause(t, i+1, bodyHi) {
				continue
			}
			emitSubordSplit(i)
		}
	}

	// 5. Extract parenthetical aside.
	for i := 1; i <= bodyHi; i++ {
		if t.Toks[i].POS != "-LRB-" && t.Toks[i].Word != "(" {
			continue
		}
		j := i + 1
		for j <= bodyHi && t.Toks[j].POS != "-RRB-" && t.Toks[j].Word != ")" {
			j++
		}
		if j > bodyHi || t.WordsIn(i+1, j-1) < 3 || splitsLock(i, j, locks) {
			continue
		}
		inner := capFirst(t.Surface(i+1, j-1))
		left := t.Surface(1, i-1)
		right := t.Surface(j+1, bodyHi)
		outside := strings.TrimSpace(left + " " + right)
		addOpt(
			fmt.Sprintf("extract_aside_%d", i),
			"extract_aside",
			"Move aside to own sentence",
			"Removes parenthetical interruption from the main line",
			outside+finalPunct+" "+inner+".",
		)
	}

	// 6. Swap balanced coordinated pair ("A and B" -> "B and A") for given-new / end-stress tuning.
	for i := 1; i <= bodyHi; i++ {
		if t.Rel[i] != "conj" || t.Head[i] <= 0 || t.Head[i] >= i {
			continue
		}
		aHead := t.Head[i]
		bHead := i
		// Require matching major POS class and disallow swapping a verb conjunct that lacks its own subject across the main clause.
		if majorPOS(t.Toks[aHead].POS) != majorPOS(t.Toks[bHead].POS) {
			continue
		}
		if majorPOS(t.Toks[aHead].POS) == "V" && !hasDirectSubj(t, bHead) {
			continue
		}
		// Do not swap endpoints across a 3+ item serial list ("A, B, and C").
		if countConjKids(t, aHead) >= 2 {
			continue
		}
		aLo, aHi := t.Lo[aHead], aHead
		// Exclude the conjunction + second conjunct from a's span.
		for _, k := range t.Kids[aHead] {
			if k < bHead && t.Rel[k] != "cc" && t.Rel[k] != "conj" && t.Rel[k] != "punct" {
				if t.Lo[k] < aLo {
					aLo = t.Lo[k]
				}
				if t.Hi[k] > aHi && t.Hi[k] < bHead {
					aHi = t.Hi[k]
				}
			}
		}
		ccIdx := 0
		for j := aHi + 1; j < bHead; j++ {
			if t.Toks[j].POS == "CC" {
				ccIdx = j
				break
			}
		}
		bLo, bHi := t.Lo[bHead], t.Hi[bHead]
		if ccIdx == 0 || bLo <= ccIdx {
			bLo = ccIdx + 1
		}
		if ccIdx == 0 || aLo < 1 || aHi >= ccIdx || bLo > bHi || bHi > bodyHi {
			continue
		}
		// Keep balanced phrase lengths (don't swap a 2-word head with a 25-word clause).
		wa, wb := t.WordsIn(aLo, aHi), t.WordsIn(bLo, bHi)
		if wa == 0 || wb == 0 || wa > wb*3 || wb > wa*3 {
			continue
		}
		if splitsLock(aLo, aHi, locks) || splitsLock(bLo, bHi, locks) || insideAnyLock(ccIdx, locks) {
			continue
		}
		prefix := t.Surface(1, aLo-1)
		partA := t.Surface(aLo, aHi)
		conjWord := t.Surface(ccIdx, ccIdx)
		partB := t.Surface(bLo, bHi)
		suffix := t.Surface(bHi+1, bodyHi)
		if aLo == 1 {
			partB = capFirst(partB)
			partA = lowerFirst(partA, t.Toks[aLo].NER)
		}
		rebuilt := strings.TrimSpace(strings.Join([]string{prefix, partB, conjWord, partA, suffix}, " ")) + finalPunct
		addOpt(
			fmt.Sprintf("swap_coord_%d", bHead),
			"swap_coord",
			fmt.Sprintf("Swap around %q", conjWord),
			fmt.Sprintf("Places %q at the stronger end position", trunc(partA, 24)),
			rebuilt,
		)
	}

	// 7. Tighten expletive "It is X that Y" -> "Y is X" / direct claim when clean.
	if Voice(t) == "expletive" && strings.EqualFold(t.Toks[1].Word, "it") && bodyHi >= 6 {
		for i := 3; i < bodyHi-2; i++ {
			if strings.EqualFold(t.Toks[i].Word, "that") && !insideAnyLock(i, locks) {
				clause := capFirst(t.Surface(i+1, bodyHi))
				addOpt(
					"drop_expletive",
					"drop_expletive",
					"Drop expletive opener",
					"Leads directly with the core claim",
					clause+finalPunct,
				)
				break
			}
		}
	}

	return opts
}

func bodyEndTok(t *corenlp.Tree) (int, string) {
	for i := t.N(); i >= maxInt(1, t.N()-2); i-- {
		w := t.Toks[i].Word
		if w == "." || w == "?" || w == "!" {
			return i - 1, w
		}
	}
	return t.N(), "."
}

func splitsLock(lo, hi int, locks []SpanLock) bool {
	for _, l := range locks {
		// Split occurs if the moved span partially overlaps a lock without containing it completely.
		overlaps := lo <= l.Hi && hi >= l.Lo
		contains := lo <= l.Lo && hi >= l.Hi
		if overlaps && !contains {
			return true
		}
	}
	return false
}

func insideAnyLock(tok int, locks []SpanLock) bool {
	for _, l := range locks {
		if tok >= l.Lo && tok <= l.Hi {
			return true
		}
	}
	return false
}

func capFirst(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) == 0 {
		return ""
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func lowerFirst(s, ner string) string {
	s = strings.TrimSpace(s)
	if s == "" || (ner != "" && ner != "O") || strings.HasPrefix(s, "I ") {
		return s
	}
	r := []rune(s)
	if len(r) > 1 && unicode.IsUpper(r[1]) {
		return s
	}
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func majorPOS(pos string) string {
	switch {
	case strings.HasPrefix(pos, "NN"):
		return "N"
	case strings.HasPrefix(pos, "VB"):
		return "V"
	case strings.HasPrefix(pos, "JJ"):
		return "J"
	case strings.HasPrefix(pos, "RB"):
		return "R"
	default:
		return pos
	}
}

func hasDirectSubj(t *corenlp.Tree, head int) bool {
	for _, k := range t.Kids[head] {
		if corenlp.SubjRels[t.Rel[k]] {
			return true
		}
	}
	return false
}

func countConjKids(t *corenlp.Tree, head int) int {
	if head <= 0 || head > t.N() {
		return 0
	}
	n := 0
	for _, k := range t.Kids[head] {
		if t.Rel[k] == "conj" {
			n++
		}
	}
	return n
}

func hasAnyRel(t *corenlp.Tree, rel string) bool {
	for i := 1; i <= t.N(); i++ {
		if t.Rel[i] == rel {
			return true
		}
	}
	return false
}

// canSplitAtConj verifies that a coordinating conjunction at ccIdx joins two
// independent subject-bearing clauses rather than closing a serial list or a
// shared-subject VP coordination.
func canSplitAtConj(t *corenlp.Tree, ccIdx, bodyHi int, hasDepStruct bool) bool {
	cHead := t.Head[ccIdx]
	if cHead > 0 && t.Rel[cHead] == "conj" {
		parent := t.Head[cHead]
		if parent != t.Root {
			return false
		}
		if !hasDirectSubj(t, cHead) {
			return false
		}
		// Reject 3+ item serial lists ("A, B, and C") so we never sever just the last item.
		if countConjKids(t, parent) >= 2 {
			return false
		}
		return true
	}
	if hasDepStruct {
		// In a full dependency parse, a CC not heading a root-level conj is
		// inside a phrase or subordinate clause and should not split the sentence.
		return false
	}
	// Heuristic fallback path when no full dependency parse is available.
	if isSerialListTail(t, ccIdx) {
		return false
	}
	return fallbackIndependentRight(t, ccIdx+1, bodyHi)
}

// isSerialListTail detects Oxford-comma serial lists ("..., X, Y, and Z") by
// checking whether at least two short comma-delimited spans precede ", and".
func isSerialListTail(t *corenlp.Tree, ccIdx int) bool {
	if ccIdx < 4 || t.Toks[ccIdx-1].Word != "," {
		return false
	}
	shortSpans := 0
	end := ccIdx - 2
	for j := end; j >= 1 && shortSpans < 2; j-- {
		if t.Toks[j].Word == "," {
			w := t.WordsIn(j+1, end)
			if w >= 1 && w <= 6 {
				shortSpans++
				end = j - 1
				continue
			}
			break
		}
	}
	return shortSpans >= 2
}

// fallbackIndependentRight checks that tokens [start..bodyHi] begin with an
// explicit subject followed by a finite verb (not immediately a verb or a
// bare item whose only verb lies inside a "what/who/which/that" relative clause).
func fallbackIndependentRight(t *corenlp.Tree, start, bodyHi int) bool {
	if start > bodyHi {
		return false
	}
	firstPOS := t.Toks[start].POS
	if strings.HasPrefix(firstPOS, "VB") || firstPOS == "MD" {
		return false
	}
	// Skip a single leading adverb ("and then he went..."), still rejecting "and then went...".
	idx := start
	if idx < bodyHi && strings.HasPrefix(t.Toks[idx].POS, "RB") {
		idx++
		if strings.HasPrefix(t.Toks[idx].POS, "VB") || t.Toks[idx].POS == "MD" {
			return false
		}
	}
	hasSubj := false
	for j := idx; j <= bodyHi; j++ {
		w := t.Toks[j].Word
		lw := strings.ToLower(w)
		p := t.Toks[j].POS
		if w == "," || w == ";" || w == ":" {
			break
		}
		if lw == "what" || lw == "who" || lw == "which" || lw == "whom" || lw == "whose" {
			// A verb appearing only after a wh-relativizer belongs to a relative clause.
			return false
		}
		if strings.HasPrefix(p, "VB") || p == "MD" {
			return hasSubj
		}
		if p == "PRP" || p == "DT" || p == "EX" || strings.HasPrefix(p, "NN") {
			hasSubj = true
		}
	}
	return false
}

// fallbackStartsClause checks whether [start..bodyHi] after a comma-subordinator
// (", as ...", ", while ...") begins a finite clause with its own subject rather
// than a prepositional/appositive phrase (", as a man of honor, ...").
func fallbackStartsClause(t *corenlp.Tree, start, bodyHi int) bool {
	hasVerbAnywhere := false
	for j := start; j <= bodyHi; j++ {
		if strings.HasPrefix(t.Toks[j].POS, "VB") || t.Toks[j].POS == "MD" {
			hasVerbAnywhere = true
			break
		}
	}
	if !hasVerbAnywhere {
		return false
	}
	// Find end of first comma-delimited segment.
	segEnd := bodyHi
	for j := start; j <= bodyHi; j++ {
		w := t.Toks[j].Word
		if w == "," || w == ";" || w == ":" {
			segEnd = j - 1
			break
		}
	}
	if segEnd < start+1 {
		return false
	}
	hasSubj := false
	inPrep := false
	for j := start; j <= segEnd; j++ {
		p := t.Toks[j].POS
		if strings.HasPrefix(p, "VB") || p == "MD" {
			return hasSubj
		}
		if p == "IN" || p == "TO" {
			inPrep = true
			continue
		}
		// In ParseFallback, unknown 3sg verbs like "tells" are tagged NNS;
		// treat a non-PP NNS after a subject (e.g. "each person quietly tells the truth")
		// as a finite verb when followed by a complement rather than a preposition.
		if hasSubj && !inPrep && p == "NNS" && j > start {
			if j == segEnd || (t.Toks[j+1].POS != "IN" && t.Toks[j+1].POS != "TO") {
				return true
			}
		}
		if !inPrep && (p == "PRP" || p == "DT" || p == "EX" || strings.HasPrefix(p, "NN")) {
			hasSubj = true
		}
	}
	return false
}

func trailingTailSafe(t *corenlp.Tree, clauseHead, bodyHi int) bool {
	// If the parser attached the opening verb of a trailing subordinate clause
	// forward to a coordinated verb inside the same trailing span (same left
	// boundary), evaluate the enclosing trailing clause span instead.
	for step := 0; step <= t.N(); step++ {
		h := t.Head[clauseHead]
		if h <= clauseHead || h > bodyHi {
			break
		}
		sameLeft := t.Lo[h] == t.Lo[clauseHead] ||
			(t.Lo[h] == t.Lo[clauseHead]-1 && t.Toks[t.Lo[h]].Word == ",")
		if !sameLeft {
			break
		}
		clauseHead = h
	}
	hi := t.Hi[clauseHead]
	if hi >= bodyHi-1 {
		return true
	}
	// Walk ancestors up to root: allow only trailing coordinations, punctuation,
	// further adverbial clauses, or fallback dep continuations after hi—never a
	// core main-clause argument (nsubj/obj/obl/ccomp) that would be stranded.
	seen := 0
	for anc := t.Head[clauseHead]; anc > 0 && seen <= t.N(); anc = t.Head[anc] {
		seen++
		for _, k := range t.Kids[anc] {
			if k <= hi || k > bodyHi {
				continue
			}
			switch t.Rel[k] {
			case "conj", "punct", "advcl", "cc", "parataxis", "dep":
			default:
				return false
			}
		}
		if anc == t.Root {
			break
		}
	}
	return true
}

func hasSerialListInSpan(t *corenlp.Tree, lo, hi int) bool {
	for i := lo; i <= hi; i++ {
		if countConjKids(t, i) >= 2 {
			return true
		}
	}
	if countConjKids(t, t.Root) >= 2 {
		return true
	}
	// Surface fallback: two commas followed by ", and" / ", or" inside the span.
	commas := 0
	for i := lo; i <= hi; i++ {
		if t.Toks[i].Word == "," {
			commas++
		}
		if commas >= 2 && t.Toks[i].POS == "CC" && i > lo && t.Toks[i-1].Word == "," {
			return true
		}
	}
	return false
}

var splitSubordMarkers = map[string]string{
	"as":       "drop",
	"while":    "drop",
	"where":    "drop",
	"wherever": "drop",
	"when":     "drop",
	"whenever": "drop",
	"although": "contrast",
	"though":   "contrast",
	"whereas":  "contrast",
	"because":  "colon",
	"since":    "colon",
}

var linkWords = map[string]bool{
	"this": true, "these": true, "that": true, "those": true, "such": true,
	"it": true, "they": true, "he": true, "she": true, "his": true, "its": true,
	"their": true, "her": true, "there": true, "then": true, "so": true, "thus": true,
	"therefore": true, "hence": true, "however": true, "but": true, "yet": true,
	"and": true, "also": true, "here": true, "each": true, "both": true, "same": true,
	"instead": true, "again": true, "first": true, "second": true, "third": true, "finally": true,
}

var stopLemmas = map[string]bool{
	"be": true, "have": true, "do": true, "make": true, "take": true, "say": true,
	"get": true, "go": true, "come": true, "give": true, "use": true, "one": true,
	"other": true, "same": true, "such": true, "many": true, "much": true, "more": true,
	"most": true, "own": true, "new": true, "first": true,
}
