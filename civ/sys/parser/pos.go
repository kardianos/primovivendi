// pos.go — Stage 2: averaged-perceptron POS tagger (Collins 2002), greedy
// left-to-right over the canonical Penn Treebank tags. The same sparse linear
// model and trainer back the Stage 4 NER tagger.
//
// A trained model is a sparse table: each feature (identified by a 64-bit hash
// of its template key) holds only its non-zero class weights, so the whole
// tagger is a few megabytes on disk and in memory while scoring stays a
// handful of adds per token.
package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"unicode"
	"unicode/utf8"

	"po/civ/sys/vec"
)

const posModelFormat = "civ-parser-pos-v1"

// ---- sparse multi-class linear model ------------------------------------------

type apModel struct {
	Classes  []string
	nClass   int
	keys     []uint64          // feature hash of each row
	off      []int32           // CSR row offsets (len = rows+1)
	col      []uint8           // class of each non-zero weight
	val      []float32         // non-zero weights
	tab      []int32           // open addressing: slot -> row+1 (0 = empty)
	shift    uint              // slot = hash >> shift
	dict     map[string]int    // unambiguous word (lowercase) -> class
	dictHash map[uint64]int    // the same dictionary keyed by the case-folded word hash
	tagsets  map[uint64]uint64 // case-folded word hash -> bitmask of the classes it took in training
	Meta     map[string]string // "__metadata__" of the .safetensors file
	wn       *lemmatizer       // WordNet lexical priors (shared, read-only; nil disables them)
	shortcut bool              // tag dictionary words directly, skipping feature scoring
}

func (m *apModel) lookup(h uint64) (int32, bool) {
	mask := uint64(len(m.tab) - 1)
	slot := (h >> m.shift) & mask
	for {
		r := m.tab[slot]
		if r == 0 {
			return 0, false
		}
		if m.keys[r-1] == h {
			return r - 1, true
		}
		slot = (slot + 1) & mask
	}
}

// buildTable indexes the rows in an open-addressed hash table.
func (m *apModel) buildTable() {
	n := 1
	for n < 2*len(m.keys)+1 {
		n <<= 1
	}
	m.tab = make([]int32, n)
	bits := 0
	for 1<<bits < n {
		bits++
	}
	m.shift = uint(64 - bits)
	mask := uint64(n - 1)
	for r, h := range m.keys {
		slot := (h >> m.shift) & mask
		for m.tab[slot] != 0 {
			slot = (slot + 1) & mask
		}
		m.tab[slot] = int32(r + 1)
	}
}

func (m *apModel) buildDictHash() { m.dictHash = dictByHash(m.dict) }

// score accumulates the feature rows of feats into scores (len = nClass).
//
// Rows are resolved in passes (table probes, key checks, row bounds) over a
// chunk of features instead of one feature at a time: every pass is a set of
// independent loads whose cache misses overlap, whereas resolving a single
// feature serializes three dependent misses. The accumulation order (rows in
// feature order, weights in stored order) is unchanged.
func (m *apModel) score(feats []uint64, scores []float32) {
	// acc is 256 long so that acc[col] needs no bounds check for the uint8
	// class ids; loadAPModel rejects ids >= nClass.
	var acc [256]float32
	mask := uint64(len(m.tab) - 1)
	var rows, lo, hi [scoreChunk]int32
	for c0 := 0; c0 < len(feats); c0 += scoreChunk {
		f := feats[c0:min(c0+scoreChunk, len(feats))]
		for i, h := range f {
			rows[i] = m.probe(h, mask)
		}
		nr := 0
		for i, h := range f {
			if r := rows[i]; r != 0 && m.keys[r-1] == h {
				rows[nr] = r - 1
				nr++
			}
		}
		for i := 0; i < nr; i++ {
			lo[i], hi[i] = m.off[rows[i]], m.off[rows[i]+1]
		}
		for i := 0; i < nr; i++ {
			cols, vals := m.col[lo[i]:hi[i]], m.val[lo[i]:hi[i]]
			if len(cols) == 0 {
				continue
			}
			_ = vals[len(cols)-1] // len(vals) == len(cols): keeps vals[k] unchecked below
			for k := range cols {
				acc[cols[k]] += vals[k]
			}
		}
	}
	copy(scores, acc[:m.nClass])
}

// scoreChunk is the number of features score resolves per pass; it covers
// the richest feature set of the POS tagger in one chunk.
const scoreChunk = 64

// probe returns the table entry of h (row+1, 0 if absent) by linear probing.
func (m *apModel) probe(h uint64, mask uint64) int32 {
	slot := (h >> m.shift) & mask
	for {
		r := m.tab[slot]
		if r == 0 || m.keys[r-1] == h {
			return r
		}
		slot = (slot + 1) & mask
	}
}

func argmaxFloat(scores []float32) int {
	best, bi := float32(-math.MaxFloat32), 0
	for c, s := range scores {
		if s > best {
			best, bi = s, c
		}
	}
	return bi
}

func (m *apModel) save(path, format string, extra map[string]string) error {
	w := newStWriter()
	w.Meta("format", format)
	for k, v := range extra {
		w.Meta(k, v)
	}
	classes, _ := json.Marshal(m.Classes)
	w.U8("classes", []int{len(classes)}, classes)
	keys := make([]int64, len(m.keys))
	for i, k := range m.keys {
		keys[i] = int64(k)
	}
	w.I64("keys", []int{len(keys)}, keys)
	w.I32("off", []int{len(m.off)}, m.off)
	w.U8("col", []int{len(m.col)}, m.col)
	w.F32("val", []int{len(m.val)}, m.val)
	dw, dt := make([]string, 0, len(m.dict)), make([]int, 0, len(m.dict))
	for word, c := range m.dict {
		dw = append(dw, word)
		dt = append(dt, c)
	}
	dict, _ := json.Marshal(map[string]any{"w": dw, "t": dt})
	w.U8("dict", []int{len(dict)}, dict)
	tk, tm := make([]int64, 0, len(m.tagsets)), make([]int64, 0, len(m.tagsets))
	for k, v := range m.tagsets {
		tk = append(tk, int64(k))
		tm = append(tm, int64(v))
	}
	w.I64("tagset.keys", []int{len(tk)}, tk)
	w.I64("tagset.mask", []int{len(tm)}, tm)
	return w.Save(path)
}

func loadAPModel(path, format string) (*apModel, error) {
	f, err := stOpen(path)
	if err != nil {
		return nil, err
	}
	if got := f.Meta("format"); got != format {
		return nil, fmt.Errorf("%s: format %q, want %q", path, got, format)
	}
	m := &apModel{}
	if err := f.JSON("classes", &m.Classes); err != nil {
		return nil, err
	}
	m.nClass = len(m.Classes)
	keys, _, err := f.I64("keys")
	if err != nil {
		return nil, err
	}
	m.keys = make([]uint64, len(keys))
	for i, k := range keys {
		m.keys[i] = uint64(k)
	}
	if m.off, _, err = f.I32("off"); err != nil {
		return nil, err
	}
	if m.col, _, err = f.U8("col"); err != nil {
		return nil, err
	}
	if m.val, _, err = f.F32("val"); err != nil {
		return nil, err
	}
	if len(m.off) != len(m.keys)+1 || len(m.col) != len(m.val) {
		return nil, fmt.Errorf("%s: inconsistent sparse table", path)
	}
	// score indexes its class accumulator with col unchecked, so the table
	// is validated once here.
	if m.nClass > 256 {
		return nil, fmt.Errorf("%s: %d classes, want at most 256", path, m.nClass)
	}
	for i, c := range m.col {
		if int(c) >= m.nClass {
			return nil, fmt.Errorf("%s: row weight %d has class %d out of range", path, i, c)
		}
	}
	for i := 1; i < len(m.off); i++ {
		if m.off[i] < m.off[i-1] || m.off[i] > int32(len(m.col)) {
			return nil, fmt.Errorf("%s: row offsets are not monotone", path)
		}
	}
	if len(m.off) > 0 && m.off[0] != 0 {
		return nil, fmt.Errorf("%s: row offsets do not start at 0", path)
	}
	var d struct {
		W []string `json:"w"`
		T []int    `json:"t"`
	}
	if err := f.JSON("dict", &d); err == nil {
		m.dict = make(map[string]int, len(d.W))
		for i, word := range d.W {
			if i < len(d.T) {
				m.dict[word] = d.T[i]
			}
		}
	}
	if f.Has("tagset.keys") {
		tk, _, err := f.I64("tagset.keys")
		if err != nil {
			return nil, err
		}
		tm, _, err := f.I64("tagset.mask")
		if err != nil {
			return nil, err
		}
		m.tagsets = make(map[uint64]uint64, len(tk))
		for i := range tk {
			if i < len(tm) {
				m.tagsets[uint64(tk[i])] = uint64(tm[i])
			}
		}
	}
	m.Meta = f.meta
	m.buildTable()
	m.buildDictHash()
	return m, nil
}

// ---- feature hashing ---------------------------------------------------------

func hashStr(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

func mix(kind, a uint64) uint64 {
	x := kind*0x9E3779B97F4A7C15 ^ (a + 0x165667B19E3779F9)
	x ^= x >> 30
	x *= 0xBF58476D1CE4E5B9
	x ^= x >> 27
	x *= 0x94D049BB133111EB
	x ^= x >> 31
	return x
}

func mix2(kind, a, b uint64) uint64 {
	return mix(kind, a*0x9E3779B97F4A7C15^(b+0xD6E8FEB86659FD93))
}

func bit(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

// wordShape maps letters to case marks and digits to "d": "McDonald's" -> "XxXxxxx'x".
// shapeHash hashes that shape without allocating it (shapes longer than 63
// bytes are truncated).
func shapeHash(w string) uint64 {
	var buf [64]byte
	n := 0
	for _, r := range w {
		c := byte(0)
		switch {
		case r >= 'A' && r <= 'Z':
			c = 'X'
		case r >= 'a' && r <= 'z':
			c = 'x'
		case unicode.IsUpper(r):
			c = 'X'
		case unicode.IsLower(r):
			c = 'x'
		case unicode.IsDigit(r):
			c = 'd'
		default:
			if r < utf8.RuneSelf {
				c = byte(r)
			} else {
				c = '?' // non-ASCII punctuation and symbols collapse together
			}
		}
		if n < len(buf) {
			buf[n] = c
			n++
		}
	}
	h := uint64(14695981039346656037)
	for i := 0; i < n; i++ {
		h ^= uint64(buf[i])
		h *= 1099511628211
	}
	return h
}

func suffix(w string, n int) string {
	if len(w) <= n {
		return w
	}
	s := w[len(w)-n:]
	for len(s) > 1 && s[0]&0xC0 == 0x80 {
		s = s[1:]
	}
	return s
}

func prefix(w string, n int) string {
	if len(w) <= n {
		return w
	}
	s := w[:n]
	for len(s) > 1 && !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s
}

// tokHashes holds the per-token values the POS feature template needs, so a
// token's strings are hashed once per sentence rather than once per feature.
type tokHashes struct {
	raw, low, shape uint64
	suf             [6]uint64
	pre             [4]uint64
	length          int
	caps, allCaps   bool
	hasDigit, punct bool
	tagset          uint64 // bitmask of the tags a lowercase word took in training (0 if unseen)
	wn              uint64 // WordNet part-of-speech prior: bits 0-3 for the form, 4-7 for its base form
}

func hashTokens(words []string, dst []tokHashes, tagsets map[uint64]uint64, wn *lemmatizer) []tokHashes {
	dst = dst[:0]
	for _, w := range words {
		t := tokHashes{raw: hashStr(w), low: vec.HashWord(w), shape: shapeHash(w), length: len(w)}
		if tagsets != nil {
			t.tagset = tagsets[t.low]
		}
		if wn != nil {
			t.wn = wn.posPrior(w)
		}
		for k := 1; k <= 6; k++ {
			t.suf[k-1] = hashStr(suffix(w, k))
		}
		for k := 1; k <= 4; k++ {
			t.pre[k-1] = hashStr(prefix(w, k))
		}
		letters, uppers, digits := 0, 0, 0
		for _, r := range w {
			switch {
			case unicode.IsUpper(r):
				uppers++
				letters++
			case unicode.IsLetter(r):
				letters++
			case unicode.IsDigit(r):
				digits++
			}
		}
		first, _ := utf8.DecodeRuneInString(w)
		t.caps = unicode.IsUpper(first)
		t.allCaps = letters > 1 && uppers == letters
		t.hasDigit = digits > 0
		t.punct = letters == 0 && digits == 0
		dst = append(dst, t)
	}
	return dst
}

// posFeatures writes the feature hashes of token i. p1 and p2 are the class ids
// of the two previous predicted tags (-1 at the start of a sentence); dictTag
// is the unambiguous tag of the lowercase word (-1 when absent).
func posFeatures(h []tokHashes, i, p1, p2, dictTag int, dst []uint64) []uint64 {
	t := &h[i]
	n := len(h)
	dst = append(dst[:0],
		mix(1, 0), mix(2, t.raw), mix(3, t.low), mix(11, t.shape),
		mix(4, t.suf[0]), mix(5, t.suf[1]), mix(6, t.suf[2]), mix(7, t.suf[3]),
		mix(8, t.pre[0]), mix(9, t.pre[1]), mix(10, t.pre[2]),
		mix(33, uint64(t.length/3)), mix(36, bit(t.caps && i > 0)), mix(37, bit(t.hasDigit)),
		mix(38, bit(t.allCaps)), mix(39, bit(t.punct)))
	if i == 0 {
		dst = append(dst, mix(31, 0))
	} else {
		p := &h[i-1]
		dst = append(dst, mix(13, p.raw), mix(14, p.low), mix(19, p.suf[2]), mix(21, p.shape), mix2(30, p.low, t.low))
		if i > 1 {
			dst = append(dst, mix(17, h[i-2].low))
		}
	}
	if i == n-1 {
		dst = append(dst, mix(32, 0))
	} else {
		x := &h[i+1]
		dst = append(dst, mix(15, x.raw), mix(16, x.low), mix(20, x.suf[2]), mix(22, x.shape), mix2(29, t.low, x.low))
		if i+2 < n {
			dst = append(dst, mix(18, h[i+2].low))
		}
	}
	if p1 < 0 {
		dst = append(dst, mix(23, 0xFFFF))
	} else {
		dst = append(dst, mix(23, uint64(p1)), mix2(26, uint64(p1), t.low), mix2(27, uint64(p1), t.suf[2]),
			mix2(28, uint64(p1), t.shape), mix2(35, uint64(p1), bit(t.caps && i > 0)))
		if p2 < 0 {
			dst = append(dst, mix(24, 0xFFFF))
		} else {
			dst = append(dst, mix(24, uint64(p2)), mix2(25, uint64(p2), uint64(p1)))
		}
	}
	if dictTag >= 0 {
		dst = append(dst, mix(34, uint64(dictTag)))
	}
	if t.tagset != 0 {
		dst = append(dst, mix(40, t.tagset))
	}
	if t.wn != 0 { // WordNet lexical prior: which parts of speech the form and its base can have
		dst = append(dst, mix(50, t.wn&15), mix(51, t.wn>>4))
		if p1 >= 0 {
			dst = append(dst, mix2(52, uint64(p1), t.wn))
		}
	}
	if posExtra&1 != 0 { // longer affixes
		dst = append(dst, mix(41, t.suf[4]), mix(42, t.suf[5]), mix(43, t.pre[3]))
	}
	if posExtra&2 != 0 { // second-order context
		if i > 0 && i < n-1 {
			dst = append(dst, mix2(44, h[i-1].low, h[i+1].low))
		}
		if i+2 < n {
			dst = append(dst, mix2(45, t.low, h[i+2].low))
		}
		if p1 >= 0 {
			if i+1 < n {
				dst = append(dst, mix2(46, uint64(p1), h[i+1].low))
			}
			if p2 >= 0 {
				dst = append(dst, mix3(47, uint64(p2), uint64(p1), t.low))
			}
		}
		if i > 1 {
			dst = append(dst, mix(48, h[i-2].suf[2]))
		}
		if i+2 < n {
			dst = append(dst, mix(49, h[i+2].suf[2]))
		}
	}
	return dst
}

// posExtra is an experiment switch for optional POS feature groups (0 = none).
var posExtra int

func mix3(kind, a, b, c uint64) uint64 {
	return mix(kind, (a*0x9E3779B97F4A7C15^b)+(c*0xD6E8FEB86659FD93))
}

// ---- averaged-perceptron trainer ---------------------------------------------

// apTrainer is a single-threaded averaged perceptron trainer with lazy weight
// averaging (weights, totals, timestamps) over dense class rows.
type apTrainer struct {
	nClass int
	rows   map[uint64]int32
	w      []float32
	total  []float64
	stamp  []int32
	t      int32
	scores []float32
	idRows []int32

	raw bool // debug: export the final weights instead of the running average
}

func newAPTrainer(nClass int) *apTrainer {
	return &apTrainer{nClass: nClass, rows: map[uint64]int32{}, scores: make([]float32, nClass)}
}

// rowsFor maps feature hashes to rows; with create, unseen features get a new
// zero row. It returns a slice reused across calls.
func (tr *apTrainer) rowsFor(feats []uint64, create bool) []int32 {
	tr.idRows = tr.idRows[:0]
	for _, h := range feats {
		r, ok := tr.rows[h]
		if !ok {
			if !create {
				continue
			}
			r = int32(len(tr.w) / tr.nClass)
			tr.rows[h] = r
			tr.w = append(tr.w, make([]float32, tr.nClass)...)
			tr.total = append(tr.total, make([]float64, tr.nClass)...)
			tr.stamp = append(tr.stamp, make([]int32, tr.nClass)...)
		}
		tr.idRows = append(tr.idRows, r)
	}
	return tr.idRows
}

func (tr *apTrainer) score(rows []int32, scores []float32) {
	clear(scores)
	for _, r := range rows {
		vec.Axpy(scores, tr.w[int(r)*tr.nClass:(int(r)+1)*tr.nClass], 1)
	}
}

// update moves one unit of weight from class bad to class good on every row.
func (tr *apTrainer) update(rows []int32, good, bad int) {
	tr.t++
	for _, r := range rows {
		for c, v := range map[int]float32{good: 1, bad: -1} {
			i := int(r)*tr.nClass + c
			tr.total[i] += float64(tr.t-tr.stamp[i]) * float64(tr.w[i])
			tr.w[i] += v
			tr.stamp[i] = tr.t
		}
	}
}

// model materializes the averaged weights as a sparse model; class weights
// smaller than prune are dropped. Training can continue afterwards.
func (tr *apTrainer) model(classes []string, dict map[string]int, shortcut bool, prune float32) *apModel {
	m := &apModel{Classes: classes, nClass: tr.nClass, dict: dict, shortcut: shortcut}
	m.keys = make([]uint64, 0, len(tr.rows))
	m.off = make([]int32, 1, len(tr.rows)+1)
	tt := float64(tr.t + 1)
	averaged := func(i int) float64 {
		if tr.raw {
			return float64(tr.w[i])
		}
		return (tr.total[i] + float64(tr.t+1-tr.stamp[i])*float64(tr.w[i])) / tt
	}
	for h, r := range tr.rows {
		base := int(r) * tr.nClass
		// First pass: is any class weight above the pruning threshold?
		keep := false
		for c := 0; c < tr.nClass && !keep; c++ {
			keep = float32(math.Abs(averaged(base+c))) >= prune
		}
		if !keep {
			continue
		}
		m.keys = append(m.keys, h)
		for c := 0; c < tr.nClass; c++ {
			avg := averaged(base + c)
			if float32(math.Abs(avg)) >= prune {
				m.col = append(m.col, uint8(c))
				m.val = append(m.val, float32(avg))
			}
		}
		m.off = append(m.off, int32(len(m.col)))
	}
	m.buildTable()
	m.buildDictHash()
	return m
}

// ---- POS tagger ----------------------------------------------------------------

// POSTrainConfig controls POS tagger training.
type POSTrainConfig struct {
	Epochs     int     // shuffled passes over the training data (default 10)
	Seed       uint64  // shuffle seed (default 20261125)
	DictMin    int     // a word needs this many occurrences to enter the tag dictionary (10)
	DictPurity float64 // ...and this share of a single tag (0.97)
	Shortcut   bool    // tag dictionary words in O(1), skipping feature scoring
	Prune      float32 // drop averaged weights below this magnitude (1e-4)
}

func (c *POSTrainConfig) defaults() {
	if c.Epochs <= 0 {
		c.Epochs = 10
	}
	if c.Seed == 0 {
		c.Seed = 20261125
	}
	if c.DictMin <= 0 {
		c.DictMin = 10
	}
	if c.DictPurity <= 0 {
		c.DictPurity = 0.97
	}
	if c.Prune <= 0 {
		c.Prune = 1e-4
	}
}

// POSTrainResult summarizes a training run.
type POSTrainResult struct {
	Model     *apModel
	BestEpoch int
	BestAcc   float64 // dev accuracy of the kept model
	History   []float64
}

func ptbClass(tag string) int { return ptbIndex[tag] }

// goldClasses maps a sentence's gold tags to class ids (ok is false when a tag
// is outside the canonical inventory).
func goldClasses(s *Sentence, classOf func(string) int) ([]int, bool) {
	n := s.Len()
	out := make([]int, n)
	for i := 0; i < n; i++ {
		c := classOf(s.POS[i])
		if c < 0 {
			return nil, false
		}
		out[i] = c
	}
	return out, true
}

// buildTagDict returns the words whose tag is unambiguous in the corpus.
func buildTagDict(train []*Sentence, classOf func(string) int, nClass, min int, purity float64) map[string]int {
	type cnt struct {
		n    int
		byCl []int32
	}
	m := map[string]*cnt{}
	for _, s := range train {
		if s == nil {
			continue
		}
		for i, w := range s.Words {
			c := classOf(s.POS[i])
			if c < 0 {
				continue
			}
			lw := strings.ToLower(w)
			e := m[lw]
			if e == nil {
				e = &cnt{byCl: make([]int32, nClass)}
				m[lw] = e
			}
			e.n++
			e.byCl[c]++
		}
	}
	out := map[string]int{}
	for w, e := range m {
		if e.n < min {
			continue
		}
		best, bi := int32(0), 0
		for c, n := range e.byCl {
			if n > best {
				best, bi = n, c
			}
		}
		if float64(best) >= purity*float64(e.n) {
			out[w] = bi
		}
	}
	return out
}

// tagTokens tags a sentence of pre-hashed tokens with class ids.
func (m *apModel) tagTokens(h []tokHashes, dst []int, scores []float32, feats []uint64) []int {
	dst = dst[:0]
	if len(scores) < m.nClass {
		scores = make([]float32, m.nClass)
	}
	scores = scores[:m.nClass]
	p1, p2 := -1, -1
	for i := range h {
		dt := -1
		if c, ok := m.dictHash[h[i].low]; ok {
			dt = c
		}
		pred := dt
		if !(m.shortcut && dt >= 0) {
			m.score(posFeatures(h, i, p1, p2, dt, feats[:0]), scores)
			pred = argmaxFloat(scores)
		}
		dst = append(dst, pred)
		p1, p2 = pred, p1
	}
	return dst
}

// POSAccuracy scores a model on gold-tagged sentences (sentences with
// non-canonical tags are skipped).
func POSAccuracy(m *apModel, sents []*Sentence) (correct, total int) {
	var (
		h     []tokHashes
		feats [96]uint64
		preds []int
		sc    = make([]float32, m.nClass)
	)
	for _, s := range sents {
		if s == nil || s.Len() == 0 {
			continue
		}
		gold, ok := goldClasses(s, ptbClass)
		if !ok {
			continue
		}
		h = hashTokens(s.Words, h, m.tagsets, m.wn)
		preds = m.tagTokens(h, preds, sc, feats[:])
		for i, g := range gold {
			if preds[i] == g {
				correct++
			}
			total++
		}
	}
	return
}

// wnTrain is the lemmatizer whose WordNet vocabulary provides lexical prior
// features during training (set by TrainPOSWithLexicon).
var wnTrain *lemmatizer

// TrainPOS trains the tagger and keeps the epoch with the best dev accuracy.
func TrainPOS(train, dev []*Sentence, cfg POSTrainConfig) *POSTrainResult {
	return TrainPOSWithLexicon(train, dev, cfg, nil)
}

// TrainPOSWithLexicon is TrainPOS with WordNet part-of-speech prior features
// taken from lex's vocabulary (lex may be nil).
func TrainPOSWithLexicon(train, dev []*Sentence, cfg POSTrainConfig, lex *lemmatizer) *POSTrainResult {
	wnTrain = lex
	cfg.defaults()
	dict := buildTagDict(train, ptbClass, len(PTBTags), cfg.DictMin, cfg.DictPurity)
	dictH := dictByHash(dict)
	tagsets := buildTagsets(train, ptbClass, len(PTBTags))
	tr := newAPTrainer(len(PTBTags))
	rng := rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9e3779b97f4a7c15))
	var (
		h     []tokHashes
		feats [96]uint64
	)
	order := make([]int, 0, len(train))
	for i := range train {
		if train[i] == nil {
			continue
		}
		for k := 0; k < max(1, train[i].Weight); k++ {
			order = append(order, i)
		}
	}
	// Pre-compute gold classes once.
	goldAll := make([][]int, len(train))
	for i, s := range train {
		if s == nil || s.Len() == 0 {
			continue
		}
		goldAll[i], _ = goldClasses(s, ptbClass)
	}
	res := &POSTrainResult{BestAcc: -1}
	for ep := 1; ep <= cfg.Epochs; ep++ {
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for _, idx := range order {
			s := train[idx]
			gold := goldAll[idx]
			if s == nil || len(gold) != s.Len() || len(gold) == 0 {
				continue
			}
			h = hashTokens(s.Words, h, tagsets, wnTrain)
			p1, p2 := -1, -1
			for i := 0; i < s.Len(); i++ {
				dt := -1
				if c, ok := dictH[h[i].low]; ok {
					dt = c
				}
				if cfg.Shortcut && dt >= 0 {
					p1, p2 = dt, p1
					continue
				}
				f := posFeatures(h, i, p1, p2, dt, feats[:0])
				tr.idRows = tr.idRows[:0]
				for _, hh := range f {
					if r, ok := tr.rows[hh]; ok {
						tr.idRows = append(tr.idRows, r)
					}
				}
				tr.score(tr.idRows, tr.scores)
				pred := argmaxFloat(tr.scores)
				if pred != gold[i] {
					tr.rowsFor(f, true)
					tr.update(tr.idRows, gold[i], pred)
				}
				p1, p2 = pred, p1
			}
		}
		model := tr.model(PTBTags, dict, cfg.Shortcut, cfg.Prune)
		model.tagsets = tagsets
		model.wn = lex
		acc := 0.0
		if len(dev) > 0 {
			c, t := POSAccuracy(model, dev)
			acc = float64(c) / float64(max(1, t))
		}
		res.History = append(res.History, acc)
		if acc > res.BestAcc || len(dev) == 0 {
			res.BestAcc, res.BestEpoch, res.Model = acc, ep, model
		}
	}
	if res.Model == nil {
		res.Model = tr.model(PTBTags, dict, cfg.Shortcut, cfg.Prune)
		res.Model.tagsets = tagsets
		res.Model.wn = lex
	}
	return res
}

// buildTagsets maps each lowercase word seen at least twice to the bitmask of
// the classes it received (classes below a fifth of its occurrences are noise
// and are dropped).
func buildTagsets(train []*Sentence, classOf func(string) int, nClass int) map[uint64]uint64 {
	type wc struct {
		n    int
		byCl []int32
	}
	m := map[uint64]*wc{}
	for _, s := range train {
		if s == nil {
			continue
		}
		for i, w := range s.Words {
			c := classOf(s.POS[i])
			if c < 0 {
				continue
			}
			k := vec.HashWord(w)
			e := m[k]
			if e == nil {
				e = &wc{byCl: make([]int32, nClass)}
				m[k] = e
			}
			e.n++
			e.byCl[c]++
		}
	}
	out := make(map[uint64]uint64, len(m))
	for k, e := range m {
		if e.n < 2 {
			continue
		}
		var mask uint64
		for c, n := range e.byCl {
			if 5*int(n) >= e.n {
				mask |= 1 << uint(c)
			}
		}
		if mask != 0 {
			out[k] = mask
		}
	}
	return out
}

// dictByHash re-keys a word -> class dictionary by case-folded word hash.
func dictByHash(dict map[string]int) map[uint64]int {
	h := make(map[uint64]int, len(dict))
	for w, c := range dict {
		h[vec.HashWord(w)] = c
	}
	return h
}
