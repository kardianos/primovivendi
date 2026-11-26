// doc.go — the public API of the pure-Go NLP parser: Load, ParseSentence,
// and ParseDoc. A Pipeline is immutable after Load and safe for concurrent
// use; every call keeps its working memory in a pooled parseState.
package parser

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"po/civ/sys/corenlp"
	"po/civ/sys/prof"
)

// Model file names under ext-ref/parser/.
const (
	POSModelFile = "pos_tagger.safetensors"
	DepModelFile = "dep_parser.safetensors"
	NERModelFile = "ner.safetensors"
)

// ErrNotTrained is returned by Load when the model files have not been built.
var ErrNotTrained = errors.New("parser: models not trained (run: civ/bin/parser train-all)")

// ModelDir returns the local (uncommitted) model directory of root (ext-ref/parser).
func ModelDir(root string) string { return filepath.Join(root, "ext-ref", "parser") }

// BundledModelDir returns the committed compressed model directory under civ/.
func BundledModelDir(root string) string { return filepath.Join(root, "civ", "sys", "parser", "models") }

// ResolveModelPath finds a model file, preferring freshly trained local weights
// under ext-ref/parser/ and falling back to compressed models committed under
// civ/sys/parser/models/.
func ResolveModelPath(root, file string) (string, bool) {
	candidates := []string{
		filepath.Join(ModelDir(root), file),
		filepath.Join(ModelDir(root), file+".gz"),
		filepath.Join(BundledModelDir(root), file+".gz"),
		filepath.Join(BundledModelDir(root), file),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, true
		}
	}
	return "", false
}

// Pipeline holds the trained stages: tokenizer/lemmatizer (Stage 1), POS
// tagger (Stage 2), dependency parser (Stage 3), and the optional NER tagger
// and coreference resolver (Stage 4).
type Pipeline struct {
	Root string
	POS  string // "go" once loaded
	pos  *apModel
	dep  *depEngine
	ner  *apModel
	lem  *lemmatizer
	pool sync.Pool
	// sem bounds concurrent sentence parsing across all callers.
	sem chan struct{}
}

// HasNER reports whether the Stage 4 NER model is loaded.
func (p *Pipeline) HasNER() bool { return p != nil && p.ner != nil }

// parseState is the per-call working memory.
type parseState struct {
	sc     *depScratch
	h      []tokHashes
	preds  []int
	scores []float32
	feats  [96]uint64
	words  []string
	tags   []int
	wids   []uint16
	pids   []uint8
	nerH   []tokHashes
	nerOut []string
	nerScr []float32
}

func newParseState(H int) *parseState {
	sc := newDepScratch()
	sc.ensureHidden(H)
	return &parseState{sc: sc}
}

// Load reads pos_tagger, dep_parser, and optional ner models from
// ext-ref/parser/ (local overrides) or civ/sys/parser/models/ (committed
// compressed bundles). It returns (nil, ErrNotTrained) when no models exist.
func Load(root string) (*Pipeline, error) {
	posPath, okPos := ResolveModelPath(root, POSModelFile)
	depPath, okDep := ResolveModelPath(root, DepModelFile)
	if !okPos || !okDep {
		return nil, ErrNotTrained
	}
	done := prof.Span("parser.load")
	defer done()
	lem := newLemmatizer(root)
	pos, err := loadAPModel(posPath, posModelFormat)
	if err != nil {
		return nil, err
	}
	pos.wn = lem
	pos.shortcut = false
	dm, err := loadDepModel(depPath)
	if err != nil {
		return nil, err
	}
	p := &Pipeline{Root: root, POS: "go", pos: pos, dep: newDepEngine(dm), lem: lem}
	if nerPath, ok := ResolveModelPath(root, NERModelFile); ok {
		if m, err := loadAPModel(nerPath, nerModelFormat); err == nil {
			p.ner = m
		}
	}
	p.pool.New = func() any { return newParseState(dm.dims.H) }
	// An empty buffered channel is the counting semaphore: a send blocks while
	// GOMAXPROCS sentences are already being parsed.
	p.sem = make(chan struct{}, runtime.GOMAXPROCS(0))
	return p, nil
}

// tagPOS fills POS (and the PTB word form of quotes) in place.
func (p *Pipeline) tagPOS(toks []corenlp.Tok, st *parseState) {
	st.words = st.words[:0]
	for i := range toks {
		st.words = append(st.words, toks[i].Orig)
	}
	st.h = hashTokens(st.words, st.h[:0], p.pos.tagsets, p.pos.wn)
	preds := p.pos.tagTokens(st.h, st.preds[:0], st.scores, st.feats[:])
	st.preds = preds
	for i, c := range preds {
		toks[i].POS = PTBTags[c]
		switch toks[i].Orig {
		case `"`, "'", "`", "“", "”", "‘", "’":
			if PTBTags[c] == "``" || PTBTags[c] == "''" {
				toks[i].Word = PTBTags[c]
			}
		}
	}
}

// runStages tags, lemmatizes, and parses one sentence in place, returning its
// 1-based head and DepLabels arrays (index 0 unused).
func (p *Pipeline) runStages(toks []corenlp.Tok, st *parseState) ([]int32, []int32) {
	if len(toks) == 0 {
		return nil, nil
	}
	done := prof.Span("parser.pos")
	p.tagPOS(toks, st)
	done()
	done = prof.Span("parser.lemma")
	p.lem.lemmatize(toks)
	done()
	if p.ner != nil {
		done = prof.Span("parser.ner")
		st.tags = st.tags[:0]
		for i := range toks {
			st.tags = append(st.tags, ptbIndex[toks[i].POS])
		}
		st.nerOut = p.ner.NERLabelsFor(st.words, st.tags, st.nerH[:0], st.nerOut[:0], st.nerScr, st.feats[:])
		for i := range toks {
			toks[i].NER = st.nerOut[i]
		}
		done()
	}
	done = prof.Span("parser.dep")
	st.wids = st.wids[:0]
	st.pids = st.pids[:0]
	for i := range toks {
		st.wids = append(st.wids, uint16(p.dep.m.WordOf(toks[i].Orig)))
		st.pids = append(st.pids, uint8(depPosID(toks[i].POS)))
	}
	p.dep.parse(st.wids, st.pids, st.sc)
	// Copy the result out: the scratch state goes back to the pool.
	heads := append([]int32(nil), st.sc.c.head...)
	rels := append([]int32(nil), st.sc.c.rel...)
	done()
	return heads, rels
}

// buildSent assembles a CoreNLP Sent from one sentence's tokens and arcs. The
// head and label arrays are 1-based; token indexes and arc endpoints are
// 1-based within the sentence.
func buildSent(toks []corenlp.Tok, heads, rels []int32, index int) corenlp.Sent {
	n := len(toks)
	deps := make([]corenlp.Dep, 0, n)
	for i := 1; i <= n; i++ {
		deps = append(deps, corenlp.Dep{Dep: DepLabels[rels[i]], Governor: int(heads[i]), Dependent: i, DepGloss: toks[i-1].Word})
	}
	for i := range toks {
		toks[i].Index = i + 1
	}
	return corenlp.Sent{Index: index, Toks: toks, Deps: deps}
}

// ParseSentence runs Stages 1-3 (plus NER when its model is trained) on one
// sentence and returns a valid 1-based dependency tree. Text with several
// sentences yields one tree whose later sentences attach to the first root
// by parataxis, so the tree always has exactly one ROOT. Empty or whitespace
// text yields an empty tree.
func (p *Pipeline) ParseSentence(text string) *corenlp.Tree {
	done := prof.Span("parser.tok")
	toks := Tokenize(text)
	done()
	if len(toks) == 0 {
		return &corenlp.Tree{Text: text}
	}
	st := p.pool.Get().(*parseState)
	defer p.pool.Put(st)
	ranges := SplitSentences(text, toks)
	allH, allR := []int32{0}, []int32{0} // 1-based
	offset := 0
	firstRoot := int32(0)
	for _, r := range ranges {
		part := toks[r.lo:r.hi]
		heads, rels := p.runStages(part, st)
		for i := 1; i <= len(part); i++ {
			h, rel := heads[i], rels[i]
			if h == 0 {
				if firstRoot == 0 {
					firstRoot = int32(offset + i)
				} else {
					h, rel = firstRoot, parataxisIdx()
				}
			} else {
				h += int32(offset)
			}
			allH = append(allH, h)
			allR = append(allR, rel)
		}
		offset += len(part)
	}
	if firstRoot == 0 {
		firstRoot = 1
		allH[1], allR[1] = 0, 0
	}
	sent := buildSent(toks, allH, allR, 0)
	tree, _ := corenlp.BuildTree(sent, text)
	if tree == nil {
		tree = &corenlp.Tree{Text: text}
	}
	return tree
}

// parataxisIdx is the DepLabels index of parataxis (resolved after the label
// inventory's init).
func parataxisIdx() int32 { return int32(depIndex["parataxis"]) }

// ParseDoc splits text into paragraphs and sentences, runs all stages, and
// returns a CoreNLP-compatible Doc (sentences with UTF-8 byte offsets and the
// coreference chains). Sentences are parsed in parallel across CPU cores.
func (p *Pipeline) ParseDoc(text string) *corenlp.Doc {
	done := prof.Span("parser.tok")
	toks := Tokenize(text)
	ranges := SplitSentences(text, toks)
	done()
	doc := &corenlp.Doc{Corefs: map[string][]corenlp.Mention{}}
	if len(ranges) == 0 {
		return doc
	}
	heads := make([][]int32, len(ranges))
	rels := make([][]int32, len(ranges))
	var wg sync.WaitGroup
	for ri, r := range ranges {
		wg.Add(1)
		go func(ri int, r sentRange) {
			defer wg.Done()
			p.sem <- struct{}{}
			defer func() { <-p.sem }()
			st := p.pool.Get().(*parseState)
			defer p.pool.Put(st)
			heads[ri], rels[ri] = p.runStages(toks[r.lo:r.hi], st)
		}(ri, r)
	}
	wg.Wait()
	trees := make([]*corenlp.Tree, len(ranges))
	doc.Sentences = make([]corenlp.Sent, len(ranges))
	for ri, r := range ranges {
		doc.Sentences[ri] = buildSent(toks[r.lo:r.hi], heads[ri], rels[ri], ri)
		trees[ri], _ = corenlp.BuildTree(doc.Sentences[ri], text)
	}
	done = prof.Span("parser.coref")
	if len(doc.Sentences) > 0 {
		doc.Corefs = Corefs(trees)
	}
	done()
	return doc
}
