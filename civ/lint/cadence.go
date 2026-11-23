package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Cadence & Voice Presence diagnostic (Phases 1 + 2, pure Go):
// Measures positive "voice" presence — rhythm/snap, broad connotative word
// weight ("greater meaning per word" vs. narrow incremental precision), and
// meaning-connection-meaning clause structure — at the paragraph and chapter
// level using four local deterministic lexicons:
//
//   1. CMU Pronouncing Dictionary  (ext-ref/phonetics/cmudict.dict)
//   2. Brysbaert Concreteness 2014 (ext-ref/concreteness/brysbaert_2014.tsv)
//   3. Kuperman AoA + SUBTLEX POS  (ext-ref/lexicon/kuperman_aoa.tsv)
//   4. Princeton WordNet 3.1       (ext-ref/lexicon/wordnet_senses.tsv)
//
// Scores are anchored to fixed baselines on a 0..100 scale (higher = stronger
// rhythm, connotative density, and punch; lower = flatter, muddier prose).

const (
	kupermanAoAURL = "https://huggingface.co/datasets/StephanAkkerman/English-Age-of-Acquisition/resolve/main/en.aoa.csv"
	wordnet31URL   = "https://wordnetcode.princeton.edu/wn3.1.dict.tar.gz"
)

type cadenceOptions struct {
	minWords int
	topN     int
	bestN    int
	asJSON   bool
	text     string
	showAll  bool
	semantic bool
}

type aoaEntry struct {
	aoa   float64
	pos   string
	lemma string
}

type cadenceLexicons struct {
	cmu  dict
	conc map[string]float64
	aoa  map[string]aoaEntry
	poly map[string]int
}

type cadenceMetrics struct {
	Words          int     `json:"words"`
	ContentWords   int     `json:"content_words"`
	Sentences      int     `json:"sentences"`
	StrPct         float64 `json:"str_pct"`          // primary-stressed syllables / total syllables (%)
	MaxWeak        int     `json:"max_weak"`         // longest consecutive run of unstressed (0) syllables
	Trough100      float64 `json:"trough_100"`       // runs of >=4 weak syllables per 100 words
	EndStrPct      float64 `json:"end_str_pct"`      // % of sentences ending on a stressed final syllable
	ParaEndStr     bool    `json:"para_end_str"`     // whether paragraph-final sentence lands on stress
	ParaEndWords   int     `json:"para_end_words"`   // word count of paragraph-final sentence
	LenCV          float64 `json:"len_cv"`           // SD/mean of sentence lengths
	ShortPunch     bool    `json:"short_punch"`      // has a <=12w stressed masculine sentence after a >=15w setup (or standalone short stressed punch)
	SylPerCW       float64 `json:"syl_per_cw"`       // syllables per content word
	ConcMean       float64 `json:"conc_mean"`        // mean Brysbaert concreteness (1..5)
	AoAMean        float64 `json:"aoa_mean"`         // mean Kuperman Age of Acquisition (years)
	PolyMean       float64 `json:"poly_mean"`        // mean log2(1 + WordNet sense count)
	KineticVerbPct float64 `json:"kinetic_verb_pct"` // % of matched non-aux verbs that are concrete/early kinetic verbs
	Pivot100       float64 `json:"pivot_100"`        // balanced clause pivots + setup->punch turns per 100 words
	Mud100         float64 `json:"mud_100"`          // abstract/Latinate prepositional chains per 100 words
	Rhythm         float64 `json:"rhythm"`           // 0..100 sub-score
	Weight         float64 `json:"weight"`           // 0..100 sub-score
	Punch          float64 `json:"punch"`            // 0..100 sub-score
	Snap           float64 `json:"snap"`             // 0..100 composite voice presence score
	SemanticOK     bool    `json:"semantic_ok,omitempty"`
	SentLeap       float64 `json:"sent_leap,omitempty"`   // mean 1 - cos(s_i, s_{i+1})
	MaxTurn        float64 `json:"max_turn,omitempty"`    // max 1 - cos(s_i, s_{i+1})
	EndTurn        float64 `json:"end_turn,omitempty"`    // 1 - cos(s_{n-2}, s_{n-1})
	ClauseLeap     float64 `json:"clause_leap,omitempty"` // mean 1 - cos across balanced pivots
	SemVel         float64 `json:"sem_vel,omitempty"`     // sum(sent leaps) / words * 100
	TreadWater     bool    `json:"tread_water,omitempty"` // >=3 sentences with very low semantic leap
}

type cadencePara struct {
	File    string         `json:"file"`
	Line    int            `json:"line"`
	Heading string         `json:"heading"`
	Text    string         `json:"text"`
	Metrics cadenceMetrics `json:"metrics"`
}

// Prosodically unstressed mid-sentence function words. In CMUdict citation form
// many monosyllabic function words carry primary stress ("1"), which artificially
// inflates stress density in connected prose. At sentence end, stressed particles
// and negators keep their stress.
var prosodyWeak = map[string]bool{}

// Sentence-final stressed particles/negators that keep primary stress when they
// land at the end of a sentence (e.g. "Science marches on.", "...did not.").
var finalStressed = map[string]bool{}

// Weak static/abstract linkage verbs excluded from kinetic verb credit.
var weakLinkVerbs = map[string]bool{}

// Prepositions that can introduce incremental abstract/nominalization mud chains.
var mudPreps = map[string]bool{}

var mudDeterminers = map[string]bool{
	"the": true, "a": true, "an": true, "its": true, "their": true,
	"this": true, "that": true, "these": true, "those": true, "such": true,
	"our": true, "your": true, "his": true, "her": true, "any": true, "every": true,
}

var (
	reCadenceWord  = regexp.MustCompile(`[A-Za-z]+(?:[-'][A-Za-z]+)*`)
	reVowelGroup   = regexp.MustCompile(`[aeiouy]+`)
	reClausePivot  = regexp.MustCompile(`(?i)(?:[;:—]|,\s+(?:but|yet|while|whereas|and|or|so|nor)\b)`)
	reCadenceSpace = regexp.MustCompile(`\s+`)
)

func init() {
	for _, w := range strings.Fields(`the a an of to and or in for with at by from as that than if so is are was were be been being have has had do does did it its their our your his her they them we you he she who which will would can could may might shall should must into onto upon about toward towards within without through under over between among during before after since until while although though because`) {
		prosodyWeak[w] = true
	}
	for _, w := range strings.Fields(`not no nor never none on off out up down back over through away in here there now too all one more less own self`) {
		finalStressed[w] = true
	}
	for _, w := range strings.Fields(`be is are was were been being have has had do does did seem appear remain become constitute demonstrate exhibit require enable involve represent indicate provide ensure maintain establish determine reflect suggest tend exist occur serve operate function allow permit cause result depend consist include contain relate correspond apply refer define describe explain mean`) {
		weakLinkVerbs[w] = true
	}
	for _, w := range strings.Fields(`of for in with through by under within across toward towards regarding concerning upon into`) {
		mudPreps[w] = true
	}
}

func lexiconDir(root string) string {
	return filepath.Join(root, "ext-ref", "lexicon")
}

func aoaPath(root string) string {
	return filepath.Join(lexiconDir(root), "kuperman_aoa.tsv")
}

func wordnetPath(root string) string {
	return filepath.Join(lexiconDir(root), "wordnet_senses.tsv")
}

func concretenessTSVPath(root string) string {
	return filepath.Join(root, "ext-ref", "concreteness", "brysbaert_2014.tsv")
}

func ensureLexicons(root string, verbose bool) error {
	if _, err := ensureCMUQuiet(root, verbose); err != nil {
		return err
	}
	if err := os.MkdirAll(lexiconDir(root), 0o755); err != nil {
		return err
	}
	if err := ensureKupermanAoA(root, verbose); err != nil {
		return err
	}
	if err := ensureWordNetSenses(root, verbose); err != nil {
		return err
	}
	return nil
}

func ensureCMUQuiet(root string, verbose bool) (string, error) {
	path := cmuPath(root)
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		if verbose {
			fmt.Printf("cmudict: %s\n", path)
		}
		return path, nil
	}
	return ensureCMU(root)
}

func ensureKupermanAoA(root string, verbose bool) error {
	out := aoaPath(root)
	if st, err := os.Stat(out); err == nil && st.Size() > 0 {
		if verbose {
			fmt.Printf("kuperman-aoa: %s\n", out)
		}
		return nil
	}
	fmt.Printf("downloading %s\n", kupermanAoAURL)
	req, _ := http.NewRequest("GET", kupermanAoAURL, nil)
	req.Header.Set("User-Agent", "civ-lint/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kuperman-aoa: %s", resp.Status)
	}
	r := csv.NewReader(resp.Body)
	r.FieldsPerRecord = -1
	hdr, err := r.Read()
	if err != nil {
		return err
	}
	idx := map[string]int{}
	for i, h := range hdr {
		idx[strings.TrimSpace(h)] = i
	}
	iWord, okW := idx["Word"]
	iPos, okP := idx["Dom_PoS_SUBTLEX"]
	iLem, okL := idx["Lemma_highest_PoS"]
	iAoA, okA := idx["AoA_Kup"]
	iAoALem, okAL := idx["AoA_Kup_lem"]
	if !okW || !okP || !okL || !okA || !okAL {
		return fmt.Errorf("kuperman-aoa: unexpected header %v", hdr)
	}
	var sb strings.Builder
	n := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(rec) <= iAoALem {
			continue
		}
		w := strings.ToLower(strings.TrimSpace(rec[iWord]))
		if w == "" || strings.ContainsAny(w, " \t") {
			continue
		}
		aoaStr := strings.TrimSpace(rec[iAoA])
		if aoaStr == "" || aoaStr == "NA" {
			aoaStr = strings.TrimSpace(rec[iAoALem])
		}
		v, err := strconv.ParseFloat(aoaStr, 64)
		if err != nil || v <= 0 {
			continue
		}
		pos := strings.TrimSpace(rec[iPos])
		lem := strings.ToLower(strings.TrimSpace(rec[iLem]))
		if lem == "" {
			lem = w
		}
		fmt.Fprintf(&sb, "%s\t%.2f\t%s\t%s\n", w, v, pos, lem)
		n++
	}
	if err := os.WriteFile(out, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("kuperman-aoa: wrote %s (%d entries)\n", out, n)
	return nil
}

func ensureWordNetSenses(root string, verbose bool) error {
	out := wordnetPath(root)
	if st, err := os.Stat(out); err == nil && st.Size() > 0 {
		if verbose {
			fmt.Printf("wordnet-senses: %s\n", out)
		}
		return nil
	}
	fmt.Printf("downloading %s\n", wordnet31URL)
	req, _ := http.NewRequest("GET", wordnet31URL, nil)
	req.Header.Set("User-Agent", "civ-lint/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wordnet3.1: %s", resp.Status)
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	counts := map[string]int{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		base := filepath.Base(hdr.Name)
		if base != "index.noun" && base != "index.verb" && base != "index.adj" && base != "index.adv" {
			continue
		}
		sc := bufio.NewScanner(tr)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			if line == "" || strings.HasPrefix(line, "  ") {
				continue
			}
			f := strings.Fields(line)
			if len(f) < 3 {
				continue
			}
			lemma := strings.ToLower(f[0])
			if strings.Contains(lemma, "_") {
				continue // skip multi-word phrases
			}
			cnt, err := strconv.Atoi(f[2])
			if err != nil || cnt <= 0 {
				continue
			}
			counts[lemma] += cnt
		}
		if err := sc.Err(); err != nil {
			return err
		}
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&sb, "%s\t%d\n", k, counts[k])
	}
	if err := os.WriteFile(out, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("wordnet-senses: wrote %s (%d single-word lemmas)\n", out, len(keys))
	return nil
}

func loadCadenceLexicons(root string) (*cadenceLexicons, error) {
	if err := ensureLexicons(root, false); err != nil {
		return nil, err
	}
	cmu, err := loadCMUFile(cmuPath(root))
	if err != nil {
		return nil, err
	}
	conc, err := loadSimpleFloatTSV(concretenessTSVPath(root))
	if err != nil {
		// Concreteness TSV is optional if not fetched yet, though present in ext-ref.
		conc = map[string]float64{}
	}
	aoa, err := loadKupermanTSV(aoaPath(root))
	if err != nil {
		return nil, err
	}
	poly, err := loadSimpleIntTSV(wordnetPath(root))
	if err != nil {
		return nil, err
	}
	return &cadenceLexicons{cmu: cmu, conc: conc, aoa: aoa, poly: poly}, nil
}

func loadSimpleFloatTSV(path string) (map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := map[string]float64{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w, v, ok := strings.Cut(sc.Text(), "\t")
		if !ok {
			continue
		}
		if x, err := strconv.ParseFloat(v, 64); err == nil {
			m[w] = x
		}
	}
	return m, sc.Err()
}

func loadSimpleIntTSV(path string) (map[string]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := map[string]int{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w, v, ok := strings.Cut(sc.Text(), "\t")
		if !ok {
			continue
		}
		if x, err := strconv.Atoi(v); err == nil {
			m[w] = x
		}
	}
	return m, sc.Err()
}

func loadKupermanTSV(path string) (map[string]aoaEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := map[string]aoaEntry{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), "\t")
		if len(parts) < 4 {
			continue
		}
		v, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		m[parts[0]] = aoaEntry{aoa: v, pos: parts[2], lemma: parts[3]}
	}
	return m, sc.Err()
}

// wordStresses returns the syllable stress sequence (each 0, 1, or 2) for a word.
// Hyphenated compounds are split and concatenated; OOV words use a vowel-group
// fallback so syllable counts never drop to zero.
func (lx *cadenceLexicons) wordStresses(word string, isSentFinal bool) []int {
	low := strings.ToLower(strings.Trim(word, "'"))
	if low == "" {
		return nil
	}
	if strings.Contains(low, "-") {
		var out []int
		parts := strings.Split(low, "-")
		for i, p := range parts {
			if p == "" {
				continue
			}
			out = append(out, lx.wordStresses(p, isSentFinal && i == len(parts)-1)...)
		}
		if len(out) > 0 {
			return out
		}
	}

	var st []int
	if phones, ok := lx.cmu[strings.ToUpper(low)]; ok {
		st = phonesToStresses(phones)
	} else if strings.HasSuffix(low, "'s") {
		if phones, ok := lx.cmu[strings.ToUpper(strings.TrimSuffix(low, "'s"))]; ok {
			st = phonesToStresses(phones)
		}
	} else if strings.HasSuffix(low, "s") && len(low) > 3 {
		if phones, ok := lx.cmu[strings.ToUpper(strings.TrimSuffix(low, "s"))]; ok {
			st = phonesToStresses(phones)
		}
	}
	if len(st) == 0 {
		st = fallbackStresses(low)
	}

	// Connected-speech prosody adjustment: unstressed mid-sentence function
	// words drop to 0, while sentence-final stressed particles/negators keep 1.
	if len(st) == 1 {
		if isSentFinal && finalStressed[low] {
			st[0] = 1
		} else if prosodyWeak[low] && !finalStressed[low] {
			st[0] = 0
		} else if prosodyWeak[low] && !isSentFinal {
			st[0] = 0
		}
	} else if prosodyWeak[low] && !isSentFinal {
		for i := range st {
			st[i] = 0
		}
	}
	return st
}

func phonesToStresses(phones []string) []int {
	var st []int
	for _, p := range phones {
		if len(p) == 0 {
			continue
		}
		last := p[len(p)-1]
		if last >= '0' && last <= '2' {
			st = append(st, int(last-'0'))
		}
	}
	return st
}

func fallbackStresses(low string) []int {
	groups := reVowelGroup.FindAllString(low, -1)
	n := len(groups)
	if n > 1 && strings.HasSuffix(low, "e") && !strings.HasSuffix(low, "le") {
		n--
	}
	if n < 1 {
		n = 1
	}
	st := make([]int, n)
	st[0] = 1
	return st
}

func (lx *cadenceLexicons) lookupLemma(low string) (lemma string, aoa float64, pos string, okAoA bool) {
	if e, ok := lx.aoa[low]; ok {
		return e.lemma, e.aoa, e.pos, true
	}
	rl := roughLemma(low)
	if e, ok := lx.aoa[rl]; ok {
		return e.lemma, e.aoa, e.pos, true
	}
	return rl, 0, "", false
}

func (lx *cadenceLexicons) lookupConc(low, lemma string) (float64, bool) {
	if v, ok := lx.conc[low]; ok {
		return v, true
	}
	if lemma != "" {
		if v, ok := lx.conc[lemma]; ok {
			return v, true
		}
	}
	return 0, false
}

func (lx *cadenceLexicons) lookupPoly(low, lemma string) (int, bool) {
	if v, ok := lx.poly[lemma]; ok && v > 0 {
		return v, true
	}
	if v, ok := lx.poly[low]; ok && v > 0 {
		return v, true
	}
	return 0, false
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func scale100(val, lo, hi float64) float64 {
	if hi == lo {
		return 50
	}
	return 100.0 * clamp01((val-lo)/(hi-lo))
}

func splitCadenceSentences(para string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		t := strings.TrimSpace(cur.String())
		cur.Reset()
		if t != "" {
			out = append(out, t)
		}
	}
	rest := strings.TrimSpace(reCadenceSpace.ReplaceAllString(para, " "))
	for {
		loc := reMDSent.FindStringIndex(rest)
		if loc == nil {
			cur.WriteString(rest)
			break
		}
		cur.WriteString(rest[:loc[1]])
		end := loc[0]
		word := lastWord(rest[:end])
		if !abbrev(word) {
			flush()
		}
		rest = rest[loc[1]:]
	}
	flush()
	return out
}

func measureCadence(para string, lx *cadenceLexicons) cadenceMetrics {
	sents := splitCadenceSentences(para)
	if len(sents) == 0 {
		return cadenceMetrics{}
	}

	var (
		totalWords   int
		totalSyl     int
		totalPriStr  int
		cwCount      int
		cwSyl        int
		maxWeak      int
		troughs      int
		endStrCnt    int
		validSents   int
		sentLens     []float64
		sentStrPcts  []float64
		sentEndStrs  []bool
		concSum      float64
		concN        int
		aoaSum       float64
		aoaN         int
		polySum      float64
		polyN        int
		verbTotal    int
		verbKinetic  int
		pivotCount   int
		mudCount     int
	)

	for _, s := range sents {
		rawWords := reCadenceWord.FindAllString(s, -1)
		if len(rawWords) == 0 {
			continue
		}
		validSents++
		sentLens = append(sentLens, float64(len(rawWords)))
		totalWords += len(rawWords)

		var sentStresses []int
		lows := make([]string, len(rawWords))
		for i, rw := range rawWords {
			low := strings.ToLower(strings.Trim(rw, "'"))
			lows[i] = low
			isFinal := i == len(rawWords)-1
			st := lx.wordStresses(low, isFinal)
			sentStresses = append(sentStresses, st...)

			if !functionWord[low] && len(low) >= 2 {
				cwCount++
				cwSyl += len(st)
				lemma, aoa, pos, okAoA := lx.lookupLemma(low)
				if okAoA {
					aoaSum += aoa
					aoaN++
				}
				conc, okConc := lx.lookupConc(low, lemma)
				if okConc {
					concSum += conc
					concN++
				}
				if senses, okPoly := lx.lookupPoly(low, lemma); okPoly {
					polySum += math.Log2(1.0 + float64(senses))
					polyN++
				}
				// Kinetic vs weak/abstract verbs.
				if strings.EqualFold(pos, "Verb") {
					verbTotal++
					if !weakLinkVerbs[low] && !weakLinkVerbs[lemma] {
						if (okConc && conc >= 2.65) || (okAoA && aoa <= 8.0) {
							verbKinetic++
						}
					}
				}
			} else if weakLinkVerbs[low] && (low == "is" || low == "are" || low == "was" || low == "were" || low == "be" || low == "been" || low == "being") {
				// Count finite/copula be-verbs in denominator so copula-heavy
				// prose does not show 100% kinetic verbs from a single verb.
				verbTotal++
			}
		}

		// Rhythm statistics on the sentence's stress stream.
		sPri := 0
		weakRun := 0
		for _, v := range sentStresses {
			if v == 1 {
				sPri++
				if weakRun >= 4 {
					troughs++
				}
				weakRun = 0
			} else if v == 2 {
				if weakRun >= 4 {
					troughs++
				}
				weakRun = 0
			} else {
				weakRun++
				if weakRun > maxWeak {
					maxWeak = weakRun
				}
			}
		}
		if weakRun >= 4 {
			troughs++
		}
		totalSyl += len(sentStresses)
		totalPriStr += sPri

		sStrPct := 0.0
		if len(sentStresses) > 0 {
			sStrPct = float64(sPri) / float64(len(sentStresses)) * 100.0
		}
		sentStrPcts = append(sentStrPcts, sStrPct)

		endStr := len(sentStresses) > 0 && sentStresses[len(sentStresses)-1] == 1
		sentEndStrs = append(sentEndStrs, endStr)
		if endStr {
			endStrCnt++
		}

		// Balanced clause pivots ("meaning-connection-meaning").
		segs := reClausePivot.Split(s, -1)
		if len(segs) >= 2 {
			for i := 0; i < len(segs)-1; i++ {
				if segmentHasWeight(segs[i]) && segmentHasWeight(segs[i+1]) {
					pivotCount++
				}
			}
		}

		// Abstract/Latinate prepositional mud chains inside the sentence.
		sMudLinks := 0
		for i := 0; i < len(lows)-1; i++ {
			if !mudPreps[lows[i]] {
				continue
			}
			j := i + 1
			if j < len(lows) && mudDeterminers[lows[j]] {
				j++
			}
			if j >= len(lows) {
				continue
			}
			target := lows[j]
			// Check target (and next word if target is an adjective modifying a nominalization).
			if lx.isAbstractHead(target) || (j+1 < len(lows) && !functionWord[target] && lx.isAbstractHead(lows[j+1])) {
				sMudLinks++
			}
		}
		if sMudLinks >= 2 {
			mudCount += sMudLinks - 1
		}
	}

	if totalWords == 0 || validSents == 0 {
		return cadenceMetrics{}
	}

	// Setup -> short stressed masculine punch across adjacent sentences.
	shortPunch := false
	for i := 1; i < len(sentLens); i++ {
		if sentLens[i-1] >= 15 && sentLens[i] <= 12 && sentEndStrs[i] && sentStrPcts[i] >= 46.0 {
			shortPunch = true
			pivotCount++ // inter-sentence setup -> punch turn
		}
	}
	if len(sentLens) == 1 && sentLens[0] <= 10 && sentEndStrs[0] && sentStrPcts[0] >= 52.0 {
		shortPunch = true
	}

	// Sentence-length CV.
	lenCV := 0.0
	if len(sentLens) >= 2 {
		mean := 0.0
		for _, l := range sentLens {
			mean += l
		}
		mean /= float64(len(sentLens))
		if mean > 0 {
			v := 0.0
			for _, l := range sentLens {
				d := l - mean
				v += d * d
			}
			lenCV = math.Sqrt(v/float64(len(sentLens))) / mean
		}
	}

	fw := float64(totalWords)
	m := cadenceMetrics{
		Words:        totalWords,
		ContentWords: cwCount,
		Sentences:    validSents,
		StrPct:       float64(totalPriStr) / float64(max(1, totalSyl)) * 100.0,
		MaxWeak:      maxWeak,
		Trough100:    float64(troughs) / fw * 100.0,
		EndStrPct:    float64(endStrCnt) / float64(validSents) * 100.0,
		ParaEndStr:   sentEndStrs[len(sentEndStrs)-1],
		ParaEndWords: int(sentLens[len(sentLens)-1]),
		LenCV:        lenCV,
		ShortPunch:   shortPunch,
		SylPerCW:     float64(cwSyl) / float64(max(1, cwCount)),
		Pivot100:     float64(pivotCount) / fw * 100.0,
		Mud100:       float64(mudCount) / fw * 100.0,
	}
	if concN > 0 {
		m.ConcMean = concSum / float64(concN)
	} else {
		m.ConcMean = 2.85 // neutral fallback
	}
	if aoaN > 0 {
		m.AoAMean = aoaSum / float64(aoaN)
	} else {
		m.AoAMean = 8.5 // neutral fallback
	}
	if polyN > 0 {
		m.PolyMean = polySum / float64(polyN)
	} else {
		m.PolyMean = 2.4 // neutral fallback
	}
	if verbTotal > 0 {
		m.KineticVerbPct = float64(verbKinetic) / float64(verbTotal) * 100.0
	} else {
		m.KineticVerbPct = 45.0
	}

	// --- Anchored Sub-scores (0..100, higher = stronger voice presence) ---
	strScore := scale100(m.StrPct, 24.0, 56.0)
	troughScore := scale100(m.Trough100, 6.0, 0.0)
	paraEndVal := 0.0
	if m.ParaEndStr {
		paraEndVal = 100.0
	}
	endScore := 0.55*paraEndVal + 0.45*m.EndStrPct
	contrastScore := scale100(m.LenCV, 0.12, 0.62)
	if m.ShortPunch {
		contrastScore = math.Min(100.0, contrastScore+35.0)
	}
	m.Rhythm = 0.35*strScore + 0.25*troughScore + 0.25*endScore + 0.15*contrastScore

	sylScore := scale100(m.SylPerCW, 2.95, 1.45)
	concScore := scale100(m.ConcMean, 2.15, 3.60)
	aoaScore := scale100(m.AoAMean, 10.8, 6.0)
	polyScore := scale100(m.PolyMean, 1.50, 3.30)
	verbScore := scale100(m.KineticVerbPct, 10.0, 75.0)
	m.Weight = 0.30*sylScore + 0.25*concScore + 0.20*aoaScore + 0.15*polyScore + 0.10*verbScore

	mudScore := scale100(m.Mud100, 6.0, 0.0)
	pivotScore := scale100(m.Pivot100, 0.0, 3.5)
	lastStr := sentStrPcts[len(sentStrPcts)-1]
	closeScore := 30.0
	if m.ParaEndStr && m.ParaEndWords <= 12 && lastStr >= 45.0 {
		closeScore = 100.0
	} else if m.ParaEndStr {
		closeScore = 70.0
	} else if m.ParaEndWords >= 28 {
		closeScore = 5.0
	}
	m.Punch = 0.45*mudScore + 0.30*pivotScore + 0.25*closeScore

	m.Snap = 0.38*m.Rhythm + 0.42*m.Weight + 0.20*m.Punch
	return m
}

func (lx *cadenceLexicons) isAbstractHead(w string) bool {
	if functionWord[w] || len(w) < 4 {
		return false
	}
	if voiceLatinateRe.MatchString(w) {
		return true
	}
	lemma, aoa, _, okAoA := lx.lookupLemma(w)
	if conc, okConc := lx.lookupConc(w, lemma); okConc {
		if conc < 2.55 {
			return true
		}
		if conc >= 3.2 {
			return false
		}
	}
	if okAoA && aoa >= 10.5 {
		return true
	}
	return false
}

func segmentHasWeight(seg string) bool {
	ws := reCadenceWord.FindAllString(seg, -1)
	if len(ws) < 3 {
		return false
	}
	cw := 0
	for _, rw := range ws {
		low := strings.ToLower(strings.Trim(rw, "'"))
		if !functionWord[low] && len(low) >= 2 {
			cw++
		}
	}
	return cw >= 2
}

// extractCadenceParas returns prose paragraphs along with their section heading.
func extractCadenceParas(file, content string) []cadencePara {
	lines := strings.Split(content, "\n")
	var out []cadencePara
	curHeading := filepath.Base(file)
	var cur []string
	curLine := 0
	fence, quote := false, false

	flush := func() {
		if len(cur) == 0 {
			return
		}
		t := strings.Join(cur, " ")
		cur = nil
		t = reMDLink.ReplaceAllString(t, "$1")
		t = reMDFoot.ReplaceAllString(t, "")
		t = reMDHTML.ReplaceAllString(t, " ")
		t = strings.NewReplacer("**", "", "*", "", "`", "", "\\", "").Replace(t)
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, cadencePara{
				File:    file,
				Line:    curLine,
				Heading: curHeading,
				Text:    t,
			})
		}
	}

	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "```") {
			fence = !fence
			flush()
			continue
		}
		if fence {
			continue
		}
		if hm := voiceHeadingRe.FindStringSubmatch(t); hm != nil {
			flush()
			curHeading = strings.TrimSpace(hm[1])
			continue
		}
		q := false
		for strings.HasPrefix(t, ">") {
			t = strings.TrimSpace(strings.TrimPrefix(t, ">"))
			q = true
		}
		if q != quote {
			flush()
			quote = q
		}
		if t == "" || strings.HasPrefix(t, "|") ||
			strings.HasPrefix(t, "---") || strings.HasPrefix(t, "<!--") || strings.HasPrefix(t, ":::") {
			flush()
			continue
		}
		if listMark.MatchString(t) {
			flush()
			t = strings.TrimSpace(listMark.ReplaceAllString(t, ""))
		}
		if t == "" {
			continue
		}
		if len(cur) == 0 {
			curLine = i + 1
		}
		cur = append(cur, t)
	}
	flush()
	return out
}

func runCadence(root string, files []string, opts cadenceOptions) error {
	lx, err := loadCadenceLexicons(root)
	if err != nil {
		return err
	}

	if opts.text != "" {
		p := cadencePara{Text: opts.text, Metrics: measureCadence(opts.text, lx)}
		if opts.semantic {
			se, err := newSemanticEngine(root)
			if err != nil {
				return err
			}
			ps := []cadencePara{p}
			if err := applySemantic(ps, se); err != nil {
				return err
			}
			p = ps[0]
		}
		if opts.asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(p.Metrics)
		}
		printSingleCadence(opts.text, p.Metrics)
		return nil
	}

	var all []cadencePara
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		ps := extractCadenceParas(f, string(b))
		for _, p := range ps {
			p.Metrics = measureCadence(p.Text, lx)
			if p.Metrics.Words < opts.minWords {
				continue
			}
			p.File = rel(root, f)
			all = append(all, p)
		}
	}

	if opts.semantic && len(all) > 0 {
		se, err := newSemanticEngine(root)
		if err != nil {
			return err
		}
		if err := applySemantic(all, se); err != nil {
			return err
		}
	}

	if opts.asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(all)
	}

	type chRow struct {
		file    string
		paras   int
		words   int
		snap    float64
		rhy     float64
		wgt     float64
		pun     float64
		str     float64
		end     float64
		syl     float64
		conc    float64
		aoa     float64
		poly    float64
		mud     float64
		leap    float64
		leapN   int
		endT    float64
		clLeap  float64
		clLeapN int
		vel     float64
	}
	rowByFile := map[string]*chRow{}
	var rowOrder []string
	for _, p := range all {
		base := filepath.Base(p.File)
		r, ok := rowByFile[base]
		if !ok {
			r = &chRow{file: base}
			rowByFile[base] = r
			rowOrder = append(rowOrder, base)
		}
		r.paras++
		r.words += p.Metrics.Words
		r.snap += p.Metrics.Snap
		r.rhy += p.Metrics.Rhythm
		r.wgt += p.Metrics.Weight
		r.pun += p.Metrics.Punch
		r.str += p.Metrics.StrPct
		if p.Metrics.ParaEndStr {
			r.end += 100.0
		}
		r.syl += p.Metrics.SylPerCW
		r.conc += p.Metrics.ConcMean
		r.aoa += p.Metrics.AoAMean
		r.poly += p.Metrics.PolyMean
		r.mud += p.Metrics.Mud100
		if p.Metrics.SemanticOK && p.Metrics.SentLeap > 0 {
			r.leap += p.Metrics.SentLeap
			r.endT += p.Metrics.EndTurn
			r.vel += p.Metrics.SemVel
			r.leapN++
		}
		if p.Metrics.SemanticOK && p.Metrics.ClauseLeap > 0 {
			r.clLeap += p.Metrics.ClauseLeap
			r.clLeapN++
		}
	}

	var rows []chRow
	for _, base := range rowOrder {
		r := *rowByFile[base]
		n := float64(r.paras)
		r.snap /= n
		r.rhy /= n
		r.wgt /= n
		r.pun /= n
		r.str /= n
		r.end /= n
		r.syl /= n
		r.conc /= n
		r.aoa /= n
		r.poly /= n
		r.mud /= n
		if r.leapN > 0 {
			ln := float64(r.leapN)
			r.leap /= ln
			r.endT /= ln
			r.vel /= ln
		}
		if r.clLeapN > 0 {
			r.clLeap /= float64(r.clLeapN)
		}
		rows = append(rows, r)
	}

	if opts.semantic {
		fmt.Println("=== Cadence & Voice Presence + Semantic Leap (Phases 1–3, pure Go + BGE-small-en-v1.5) ===")
		fmt.Printf("%-14s %4s %6s | %5s %5s %5s %5s | %5s %5s %5s %4s %4s %4s %4s | %5s %5s %5s %5s\n",
			"Chapter", "Para", "Words", "Snap", "Rhy", "Wgt", "Pun", "Str%", "End%", "SylCW", "Conc", "AoA", "Poly", "Mud", "Leap", "EndT", "ClLp", "Vel")
		fmt.Println(strings.Repeat("-", 124))
	} else {
		fmt.Println("=== Cadence & Voice Presence (Phases 1+2, pure Go; 0..100 higher = stronger rhythm, word weight, and punch) ===")
		fmt.Printf("%-14s %4s %6s | %5s %5s %5s %5s | %5s %5s %5s %4s %4s %4s %4s\n",
			"Chapter", "Para", "Words", "Snap", "Rhy", "Wgt", "Pun", "Str%", "End%", "SylCW", "Conc", "AoA", "Poly", "Mud")
		fmt.Println(strings.Repeat("-", 98))
	}

	var bSnap, bRhy, bWgt, bPun, bStr, bEnd, bSyl, bConc, bAoA, bPoly, bMud float64
	var bLeap, bEndT, bClLeap, bVel float64
	bLeapN, bClN := 0, 0
	bWords := 0
	for _, r := range rows {
		if opts.semantic {
			fmt.Printf("%-14s %4d %6d | %5.1f %5.1f %5.1f %5.1f | %5.1f %5.1f %5.2f %4.2f %4.1f %4.2f %4.2f | %5.3f %5.3f %5.3f %5.2f\n",
				r.file, r.paras, r.words, r.snap, r.rhy, r.wgt, r.pun,
				r.str, r.end, r.syl, r.conc, r.aoa, r.poly, r.mud,
				r.leap, r.endT, r.clLeap, r.vel)
		} else {
			fmt.Printf("%-14s %4d %6d | %5.1f %5.1f %5.1f %5.1f | %5.1f %5.1f %5.2f %4.2f %4.1f %4.2f %4.2f\n",
				r.file, r.paras, r.words, r.snap, r.rhy, r.wgt, r.pun,
				r.str, r.end, r.syl, r.conc, r.aoa, r.poly, r.mud)
		}
	}
	for _, p := range all {
		bWords += p.Metrics.Words
		bSnap += p.Metrics.Snap
		bRhy += p.Metrics.Rhythm
		bWgt += p.Metrics.Weight
		bPun += p.Metrics.Punch
		bStr += p.Metrics.StrPct
		if p.Metrics.ParaEndStr {
			bEnd += 100.0
		}
		bSyl += p.Metrics.SylPerCW
		bConc += p.Metrics.ConcMean
		bAoA += p.Metrics.AoAMean
		bPoly += p.Metrics.PolyMean
		bMud += p.Metrics.Mud100
		if p.Metrics.SemanticOK && p.Metrics.SentLeap > 0 {
			bLeap += p.Metrics.SentLeap
			bEndT += p.Metrics.EndTurn
			bVel += p.Metrics.SemVel
			bLeapN++
		}
		if p.Metrics.SemanticOK && p.Metrics.ClauseLeap > 0 {
			bClLeap += p.Metrics.ClauseLeap
			bClN++
		}
	}
	if len(all) > 0 {
		n := float64(len(all))
		if opts.semantic {
			fmt.Println(strings.Repeat("-", 124))
			avgLeap, avgEndT, avgVel, avgCl := 0.0, 0.0, 0.0, 0.0
			if bLeapN > 0 {
				avgLeap = bLeap / float64(bLeapN)
				avgEndT = bEndT / float64(bLeapN)
				avgVel = bVel / float64(bLeapN)
			}
			if bClN > 0 {
				avgCl = bClLeap / float64(bClN)
			}
			fmt.Printf("%-14s %4d %6d | %5.1f %5.1f %5.1f %5.1f | %5.1f %5.1f %5.2f %4.2f %4.1f %4.2f %4.2f | %5.3f %5.3f %5.3f %5.2f\n",
				"BOOK MEAN", len(all), bWords,
				bSnap/n, bRhy/n, bWgt/n, bPun/n,
				bStr/n, bEnd/n, bSyl/n, bConc/n, bAoA/n, bPoly/n, bMud/n,
				avgLeap, avgEndT, avgCl, avgVel)
		} else {
			fmt.Println(strings.Repeat("-", 98))
			fmt.Printf("%-14s %4d %6d | %5.1f %5.1f %5.1f %5.1f | %5.1f %5.1f %5.2f %4.2f %4.1f %4.2f %4.2f\n",
				"BOOK MEAN", len(all), bWords,
				bSnap/n, bRhy/n, bWgt/n, bPun/n,
				bStr/n, bEnd/n, bSyl/n, bConc/n, bAoA/n, bPoly/n, bMud/n)
		}
	}

	sorted := append([]cadencePara(nil), all...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Metrics.Snap < sorted[j].Metrics.Snap })

	lowN := opts.topN
	if opts.showAll || lowN > len(sorted) {
		lowN = len(sorted)
	}
	fmt.Printf("\n=== Lowest-Snap / Muddiest Paragraphs (%d shown, >= %dw) ===\n", lowN, opts.minWords)
	for i := 0; i < lowN; i++ {
		p := sorted[i]
		printParaRow(p)
	}

	if !opts.showAll && opts.bestN > 0 && len(sorted) > 0 {
		hiN := opts.bestN
		if hiN > len(sorted) {
			hiN = len(sorted)
		}
		fmt.Printf("\n=== Highest-Snap / Strongest Voice Paragraphs (%d shown, >= %dw) ===\n", hiN, opts.minWords)
		for i := len(sorted) - 1; i >= len(sorted)-hiN; i-- {
			printParaRow(sorted[i])
		}
	}

	if opts.semantic && !opts.showAll {
		var multi []cadencePara
		for _, p := range all {
			if p.Metrics.Sentences >= 3 && p.Metrics.SentLeap > 0 {
				multi = append(multi, p)
			}
		}
		if len(multi) > 0 {
			sort.Slice(multi, func(i, j int) bool { return multi[i].Metrics.SentLeap < multi[j].Metrics.SentLeap })
			lim := min(10, len(multi))
			fmt.Printf("\n=== Lowest Semantic Movement / Potential Treading-Water (%d shown, >= 3 sents) ===\n", lim)
			for i := 0; i < lim; i++ {
				printParaRow(multi[i])
			}

			sort.Slice(multi, func(i, j int) bool { return multi[i].Metrics.EndTurn > multi[j].Metrics.EndTurn })
			fmt.Printf("\n=== Strongest Closing Semantic Turns (%d shown, >= 3 sents) ===\n", lim)
			for i := 0; i < lim; i++ {
				printParaRow(multi[i])
			}
		}
	}
	return nil
}

func printParaRow(p cadencePara) {
	snip := p.Text
	if len([]rune(snip)) > 110 {
		snip = string([]rune(snip)[:110]) + "..."
	}
	endMark := "weak"
	if p.Metrics.ParaEndStr {
		endMark = "STR"
	}
	sem := ""
	if p.Metrics.SemanticOK {
		sem = fmt.Sprintf(" | Leap=%.3f EndT=%.3f ClLp=%.3f Vel=%.2f",
			p.Metrics.SentLeap, p.Metrics.EndTurn, p.Metrics.ClauseLeap, p.Metrics.SemVel)
		if p.Metrics.TreadWater {
			sem += " [TREAD]"
		}
	}
	fmt.Printf("  %5.1f (R:%4.0f W:%4.0f P:%4.0f)  %-28s (%3dw, Str=%4.1f%% Syl=%.2f Conc=%.2f AoA=%.1f Poly=%.2f Mud=%.1f End=%s%s)\n        %s\n",
		p.Metrics.Snap, p.Metrics.Rhythm, p.Metrics.Weight, p.Metrics.Punch,
		fmt.Sprintf("%s:%d", p.File, p.Line),
		p.Metrics.Words, p.Metrics.StrPct, p.Metrics.SylPerCW, p.Metrics.ConcMean,
		p.Metrics.AoAMean, p.Metrics.PolyMean, p.Metrics.Mud100, endMark, sem, snip)
}

func printSingleCadence(text string, m cadenceMetrics) {
	endMark := "weak (feminine)"
	if m.ParaEndStr {
		endMark = "STRESSED (masculine)"
	}
	fmt.Printf("Snap:   %5.1f / 100  (Rhythm: %4.1f | Weight: %4.1f | Punch: %4.1f)\n",
		m.Snap, m.Rhythm, m.Weight, m.Punch)
	fmt.Printf("Rhythm: Str=%.1f%%  MaxWeak=%d  Trough/100w=%.2f  SentEndStr=%.0f%%  ParaEnd=%s (%dw)  LenCV=%.2f  ShortPunch=%v\n",
		m.StrPct, m.MaxWeak, m.Trough100, m.EndStrPct, endMark, m.ParaEndWords, m.LenCV, m.ShortPunch)
	fmt.Printf("Weight: Syl/CW=%.2f  Conc=%.2f  AoA=%.2f  Poly=%.2f  KineticVerb=%.0f%%\n",
		m.SylPerCW, m.ConcMean, m.AoAMean, m.PolyMean, m.KineticVerbPct)
	fmt.Printf("Punch:  Pivot/100w=%.2f  Mud/100w=%.2f  Words=%d  Sents=%d\n",
		m.Pivot100, m.Mud100, m.Words, m.Sentences)
	if m.SemanticOK {
		fmt.Printf("Sem:    SentLeap=%.3f  MaxTurn=%.3f  EndTurn=%.3f  ClauseLeap=%.3f  SemVel=%.2f  TreadWater=%v\n",
			m.SentLeap, m.MaxTurn, m.EndTurn, m.ClauseLeap, m.SemVel, m.TreadWater)
	}
}
