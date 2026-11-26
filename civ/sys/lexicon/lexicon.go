// Package lexicon provides local lexical resources for prose composition:
// full Princeton WordNet 3.1 synsets/relations/glosses, Penn-Treebank English
// inflection, Brysbaert concreteness, Kuperman Age of Acquisition, and CMUdict
// pronunciation/stress data.
package lexicon

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const wordnet31URL = "https://wordnetcode.princeton.edu/wn3.1.dict.tar.gz"

// Pointer is a typed WordNet semantic relation between synsets.
type Pointer struct {
	Symbol    string // "@", "~", "&", "^", "+", "!", etc.
	TargetOff int
	TargetPOS byte // 'n', 'v', 'a', 'r'
}

// Synset is one Princeton WordNet 3.1 synset.
type Synset struct {
	Offset   int
	POS      byte
	Words    []string
	Pointers []Pointer
	Gloss    string
}

// Alternative is a single lexical substitute for a token, inflected to match
// the target Penn Treebank POS tag and enriched with lexicon metadata.
type Alternative struct {
	Lemma        string  `json:"lemma"`
	Surface      string  `json:"surface"`
	Relation     string  `json:"relation"` // synonym, similar, hyponym, hypernym, also, derivation, fallback
	Gloss        string  `json:"gloss,omitempty"`
	SenseRank    int     `json:"sense_rank"`
	Concreteness float64 `json:"concreteness,omitempty"`
	AoA          float64 `json:"aoa,omitempty"`
	Polysemy     int     `json:"polysemy,omitempty"`
	Syllables    int     `json:"syllables,omitempty"`
	EndStress    bool    `json:"end_stress,omitempty"`
}

// AoAEntry holds Kuperman Age of Acquisition metadata for a word.
type AoAEntry struct {
	AoA   float64
	POS   string
	Lemma string
}

// Bundle holds all loaded lexical databases in memory.
type Bundle struct {
	// WordNet synsets keyed by (POS, Offset).
	Synsets map[byte]map[int]*Synset
	// Lemma index keyed by (POS, lower_lemma) -> ordered synset offsets.
	Index map[byte]map[string][]int
	// Irregular inflections extracted from WordNet .exc files:
	// plural["child"]="children", past["strike"]="struck", etc.
	IrregPlural map[string]string
	IrregPast   map[string]string
	IrregPart   map[string]string
	IrregComp   map[string]string
	IrregSuper  map[string]string

	// Auxiliary lexicons already present in ext-ref/.
	Concreteness map[string]float64
	AoA          map[string]AoAEntry
	SenseCount   map[string]int
	CMU          map[string][]string // word -> phoneme stress sequence ("0","1","2")
	EtymMap      map[string]Etymology
	WordNetReady bool
}

// Load reads all available local lexicons under <root>/ext-ref/. Missing
// optional files are tolerated gracefully (and WordNet 3.1 can be unpacked via
// EnsureWordNet31).
func Load(root string) (*Bundle, error) {
	b := &Bundle{
		Synsets:      map[byte]map[int]*Synset{'n': {}, 'v': {}, 'a': {}, 'r': {}},
		Index:        map[byte]map[string][]int{'n': {}, 'v': {}, 'a': {}, 'r': {}},
		IrregPlural:  cloneMap(builtinIrregPlural),
		IrregPast:    cloneMap(builtinIrregPast),
		IrregPart:    cloneMap(builtinIrregPart),
		IrregComp:    cloneMap(builtinIrregComp),
		IrregSuper:   cloneMap(builtinIrregSuper),
		Concreteness: map[string]float64{},
		AoA:          map[string]AoAEntry{},
		SenseCount:   map[string]int{},
		CMU:          map[string][]string{},
		EtymMap:      map[string]Etymology{},
	}
	b.loadConcreteness(filepath.Join(root, "ext-ref", "concreteness", "brysbaert_2014.tsv"))
	b.loadAoA(filepath.Join(root, "ext-ref", "lexicon", "kuperman_aoa.tsv"))
	b.loadSenseCounts(filepath.Join(root, "ext-ref", "lexicon", "wordnet_senses.tsv"))
	b.loadCMU(filepath.Join(root, "ext-ref", "phonetics", "cmudict.dict"))
	b.loadEtymologyTSV(etymologyTSVPath(root))

	wnDir := filepath.Join(root, "ext-ref", "lexicon", "wordnet31")
	if err := b.loadWordNetDir(wnDir); err == nil && len(b.Index['n']) > 0 {
		b.WordNetReady = true
	}
	return b, nil
}

// EnsureWordNet31 downloads and unpacks Princeton WordNet 3.1 dict files into
// <root>/ext-ref/lexicon/wordnet31/ if not already present.
func EnsureWordNet31(root string) (string, error) {
	dir := filepath.Join(root, "ext-ref", "lexicon", "wordnet31")
	if st, err := os.Stat(filepath.Join(dir, "data.noun")); err == nil && st.Size() > 0 {
		return dir, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", wordnet31URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "civ-compose/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download wordnet 3.1: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download wordnet 3.1: %s", resp.Status)
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	want := map[string]bool{
		"data.noun": true, "data.verb": true, "data.adj": true, "data.adv": true,
		"index.noun": true, "index.verb": true, "index.adj": true, "index.adv": true,
		"noun.exc": true, "verb.exc": true, "adj.exc": true, "adv.exc": true,
	}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		base := filepath.Base(hdr.Name)
		if !want[base] {
			continue
		}
		dst := filepath.Join(dir, base)
		tmp := dst + ".tmp"
		f, err := os.Create(tmp)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			os.Remove(tmp)
			return "", err
		}
		if err := f.Close(); err != nil {
			os.Remove(tmp)
			return "", err
		}
		if err := os.Rename(tmp, dst); err != nil {
			os.Remove(tmp)
			return "", err
		}
	}
	return dir, nil
}

func (b *Bundle) loadWordNetDir(dir string) error {
	files := []struct {
		pos  byte
		data string
		idx  string
	}{
		{'n', "data.noun", "index.noun"},
		{'v', "data.verb", "index.verb"},
		{'a', "data.adj", "index.adj"},
		{'r', "data.adv", "index.adv"},
	}
	for _, f := range files {
		if err := b.parseWNData(filepath.Join(dir, f.data), f.pos); err != nil {
			return err
		}
		if err := b.parseWNIndex(filepath.Join(dir, f.idx), f.pos); err != nil {
			return err
		}
	}
	b.loadWNExc(filepath.Join(dir, "noun.exc"), 'n')
	b.loadWNExc(filepath.Join(dir, "verb.exc"), 'v')
	b.loadWNExc(filepath.Join(dir, "adj.exc"), 'a')
	return nil
}

func (b *Bundle) parseWNData(path string, pos byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if len(line) == 0 || line[0] == ' ' {
			continue // copyright header
		}
		meta, gloss, _ := strings.Cut(line, "|")
		fields := strings.Fields(meta)
		if len(fields) < 5 {
			continue
		}
		off, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		wCnt, err := strconv.ParseInt(fields[3], 16, 32)
		if err != nil || wCnt < 0 {
			continue
		}
		idx := 4
		words := make([]string, 0, wCnt)
		for i := 0; i < int(wCnt) && idx+1 < len(fields); i++ {
			w := strings.ReplaceAll(fields[idx], "_", " ")
			if cut := strings.IndexByte(w, '('); cut > 0 {
				w = w[:cut]
			}
			words = append(words, strings.ToLower(w))
			idx += 2
		}
		var ptrs []Pointer
		if idx < len(fields) {
			pCnt, err := strconv.Atoi(fields[idx])
			idx++
			if err == nil && pCnt > 0 {
				ptrs = make([]Pointer, 0, pCnt)
				for i := 0; i < pCnt && idx+3 < len(fields); i++ {
					sym := fields[idx]
					tOff, _ := strconv.Atoi(fields[idx+1])
					tPOS := fields[idx+2]
					var tp byte
					if len(tPOS) > 0 {
						tp = tPOS[0]
						if tp == 's' {
							tp = 'a'
						}
					}
					ptrs = append(ptrs, Pointer{Symbol: sym, TargetOff: tOff, TargetPOS: tp})
					idx += 4
				}
			}
		}
		b.Synsets[pos][off] = &Synset{
			Offset:   off,
			POS:      pos,
			Words:    words,
			Pointers: ptrs,
			Gloss:    strings.TrimSpace(gloss),
		}
	}
	return sc.Err()
}

func (b *Bundle) parseWNIndex(path string, pos byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<16), 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if len(line) == 0 || line[0] == ' ' {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		lemma := strings.ReplaceAll(fields[0], "_", " ")
		synCnt, err := strconv.Atoi(fields[2])
		if err != nil || synCnt <= 0 {
			continue
		}
		pCnt, err := strconv.Atoi(fields[3])
		if err != nil || pCnt < 0 {
			continue
		}
		offStart := 4 + pCnt + 2
		if offStart+synCnt > len(fields) {
			continue
		}
		offs := make([]int, 0, synCnt)
		for i := 0; i < synCnt; i++ {
			if off, err := strconv.Atoi(fields[offStart+i]); err == nil {
				offs = append(offs, off)
			}
		}
		b.Index[pos][lemma] = offs
	}
	return sc.Err()
}

func (b *Bundle) loadWNExc(path string, pos byte) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		inflected, base := fields[0], fields[1]
		switch pos {
		case 'n':
			if _, exists := b.IrregPlural[base]; !exists {
				b.IrregPlural[base] = inflected
			}
		case 'a':
			if strings.HasSuffix(inflected, "er") {
				if _, exists := b.IrregComp[base]; !exists {
					b.IrregComp[base] = inflected
				}
			} else if strings.HasSuffix(inflected, "est") {
				if _, exists := b.IrregSuper[base]; !exists {
					b.IrregSuper[base] = inflected
				}
			}
		}
	}
}

// PTBToWN maps a Penn Treebank POS tag to a WordNet POS byte ('n','v','a','r').
func PTBToWN(ptb string) byte {
	switch {
	case strings.HasPrefix(ptb, "NN") && !strings.HasPrefix(ptb, "NNP"):
		return 'n'
	case strings.HasPrefix(ptb, "VB"):
		return 'v'
	case strings.HasPrefix(ptb, "JJ"):
		return 'a'
	case strings.HasPrefix(ptb, "RB"):
		return 'r'
	default:
		return 0
	}
}

// Alternatives returns ranked, grammatically inflected alternatives for a word
// given its lemma, original surface text, and Penn Treebank POS tag.
func (b *Bundle) Alternatives(lemma, orig, ptbPOS string, maxResults int) []Alternative {
	if b == nil {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 12
	}
	wnPOS := PTBToWN(ptbPOS)
	if wnPOS == 0 {
		return nil
	}
	ll := strings.ToLower(strings.TrimSpace(lemma))
	if ll == "" {
		ll = strings.ToLower(strings.TrimSpace(orig))
	}
	if len(ll) < 2 || skipSwapWord[ll] {
		return nil
	}

	type rawCand struct {
		lemma     string
		rel       string
		relWeight int
		senseRank int
		gloss     string
	}
	seen := map[string]rawCand{ll: {lemma: ll}}
	consider := func(cand, rel string, relWeight, senseRank int, gloss string) {
		cand = strings.ToLower(strings.TrimSpace(cand))
		if cand == "" || cand == ll || strings.ContainsAny(cand, "0123456789") {
			return
		}
		// Keep single words and tight two-word idioms only.
		if strings.Count(cand, " ") > 1 {
			return
		}
		if prev, ok := seen[cand]; ok {
			if relWeight > prev.relWeight || (relWeight == prev.relWeight && senseRank < prev.senseRank) {
				seen[cand] = rawCand{lemma: cand, rel: rel, relWeight: relWeight, senseRank: senseRank, gloss: gloss}
			}
			return
		}
		seen[cand] = rawCand{lemma: cand, rel: rel, relWeight: relWeight, senseRank: senseRank, gloss: gloss}
	}

	if b.WordNetReady {
		offs := b.Index[wnPOS][ll]
		// Cap to top 4 most common senses to avoid bizarre obscure senses.
		limit := len(offs)
		if limit > 4 {
			limit = 4
		}
		for sIdx := 0; sIdx < limit; sIdx++ {
			syn := b.Synsets[wnPOS][offs[sIdx]]
			if syn == nil {
				continue
			}
			for _, w := range syn.Words {
				consider(w, "synonym", 100, sIdx, syn.Gloss)
			}
			for _, p := range syn.Pointers {
				if p.TargetPOS != wnPOS {
					continue
				}
				target := b.Synsets[p.TargetPOS][p.TargetOff]
				if target == nil {
					continue
				}
				var rel string
				var rw int
				switch p.Symbol {
				case "&":
					rel, rw = "similar", 85
				case "~":
					rel, rw = "hyponym", 72
				case "@":
					rel, rw = "hypernym", 62
				case "^":
					rel, rw = "also", 55
				default:
					continue
				}
				for wi, w := range target.Words {
					if wi >= 3 {
						break
					}
					consider(w, rel, rw, sIdx, target.Gloss)
				}
			}
		}
	}

	// Built-in fallback thesaurus ensures offline / test operation even before
	// WordNet 3.1 dict files are downloaded.
	for _, w := range builtinThesaurus[ll] {
		consider(w, "fallback", 70, 0, "")
	}

	delete(seen, ll)
	raws := make([]rawCand, 0, len(seen))
	for _, r := range seen {
		raws = append(raws, r)
	}
	sort.Slice(raws, func(i, j int) bool {
		if raws[i].relWeight != raws[j].relWeight {
			return raws[i].relWeight > raws[j].relWeight
		}
		if raws[i].senseRank != raws[j].senseRank {
			return raws[i].senseRank < raws[j].senseRank
		}
		// Prefer single-word, more concrete, simpler alternatives.
		si := strings.Contains(raws[i].lemma, " ")
		sj := strings.Contains(raws[j].lemma, " ")
		if si != sj {
			return !si
		}
		ci := b.Concreteness[raws[i].lemma]
		cj := b.Concreteness[raws[j].lemma]
		if ci != cj {
			return ci > cj
		}
		return raws[i].lemma < raws[j].lemma
	})

	if len(raws) > maxResults {
		raws = raws[:maxResults]
	}
	out := make([]Alternative, 0, len(raws))
	for _, r := range raws {
		surf := PreserveCase(orig, b.Inflect(r.lemma, ptbPOS))
		syl, endStr := b.Prosody(surf)
		out = append(out, Alternative{
			Lemma:        r.lemma,
			Surface:      surf,
			Relation:     r.rel,
			Gloss:        r.gloss,
			SenseRank:    r.senseRank,
			Concreteness: b.Concreteness[r.lemma],
			AoA:          b.AoA[r.lemma].AoA,
			Polysemy:     b.SenseCount[r.lemma],
			Syllables:    syl,
			EndStress:    endStr,
		})
	}
	return out
}

// Prosody returns syllable count and whether the word ends on a stressed syllable.
func (b *Bundle) Prosody(word string) (int, bool) {
	lw := strings.ToLower(strings.Trim(word, ".,;:!?\"'()"))
	if b != nil {
		if st, ok := b.CMU[lw]; ok && len(st) > 0 {
			last := st[len(st)-1]
			return len(st), last == "1" || last == "2"
		}
	}
	syl := approxSyllables(lw)
	return syl, syl == 1
}

func approxSyllables(w string) int {
	if w == "" {
		return 0
	}
	vowels := "aeiouy"
	count := 0
	inV := false
	for _, r := range w {
		isV := strings.ContainsRune(vowels, unicode.ToLower(r))
		if isV && !inV {
			count++
		}
		inV = isV
	}
	if strings.HasSuffix(w, "e") && count > 1 {
		count--
	}
	if count < 1 {
		count = 1
	}
	return count
}

func (b *Bundle) loadConcreteness(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		w, v, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		if x, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			b.Concreteness[strings.ToLower(w)] = x
		}
	}
}

func (b *Bundle) loadAoA(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		x, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			continue
		}
		e := AoAEntry{AoA: x}
		if len(parts) >= 3 {
			e.POS = parts[2]
		}
		if len(parts) >= 4 {
			e.Lemma = parts[3]
		}
		b.AoA[strings.ToLower(parts[0])] = e
	}
}

func (b *Bundle) loadSenseCounts(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		w, v, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			b.SenseCount[strings.ToLower(w)] = n
		}
	}
}

func (b *Bundle) loadCMU(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, ";;;") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		w := strings.ToLower(fields[0])
		if idx := strings.IndexByte(w, '('); idx > 0 {
			continue // keep primary pronunciation
		}
		var stresses []string
		for _, ph := range fields[1:] {
			if len(ph) > 0 {
				last := ph[len(ph)-1]
				if last == '0' || last == '1' || last == '2' {
					stresses = append(stresses, string(last))
				}
			}
		}
		if len(stresses) > 0 {
			b.CMU[w] = stresses
		}
	}
}

var skipSwapWord = map[string]bool{
	"be": true, "have": true, "do": true, "not": true, "no": true,
	"will": true, "would": true, "can": true, "could": true,
	"shall": true, "should": true, "may": true, "might": true, "must": true,
}

func cloneMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
