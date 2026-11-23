package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/MichaelAyles/goformer"
)

// Pure-Go semantic embedding engine (Phase 3, Option B):
// Loads BAAI/bge-small-en-v1.5 directly from HuggingFace safetensors via
// github.com/MichaelAyles/goformer (zero CGO, zero ONNX, zero Python), with a
// SHA-256 disk cache in ext-ref/embed/ so repeated runs are near-instant.

const (
	embedModelTag = "bge-small-en-v1.5-goformer-v1"
	embedBaseURL  = "https://huggingface.co/BAAI/bge-small-en-v1.5/resolve/main/"
	embedDims     = 384
)

var embedModelFiles = []string{
	"config.json",
	"tokenizer.json",
	"model.safetensors",
}

func embedModelDir(root string) string {
	return filepath.Join(root, "ext-ref", "embed", "bge-small-en-v1.5")
}

func embedCachePath(root string) string {
	return filepath.Join(root, "ext-ref", "embed", "cache_bge_small.bin")
}

// ensureEmbedModel downloads the canonical BAAI/bge-small-en-v1.5 HuggingFace
// files into ext-ref/embed/bge-small-en-v1.5/ if missing.
func ensureEmbedModel(root string, verbose bool) (string, error) {
	dir := embedModelDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	for _, name := range embedModelFiles {
		dst := filepath.Join(dir, name)
		if st, err := os.Stat(dst); err == nil && st.Size() > 0 {
			if verbose {
				fmt.Printf("embed-model: %s (%d bytes)\n", dst, st.Size())
			}
			continue
		}
		url := embedBaseURL + name
		fmt.Fprintf(os.Stderr, "downloading %s ...\n", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "civ-lint/1.0")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("download %s: %s", url, resp.Status)
		}
		tmp := dst + ".tmp"
		f, err := os.Create(tmp)
		if err != nil {
			resp.Body.Close()
			return "", err
		}
		n, copyErr := io.Copy(f, resp.Body)
		resp.Body.Close()
		closeErr := f.Close()
		if copyErr != nil {
			os.Remove(tmp)
			return "", copyErr
		}
		if closeErr != nil {
			os.Remove(tmp)
			return "", closeErr
		}
		if err := os.Rename(tmp, dst); err != nil {
			os.Remove(tmp)
			return "", err
		}
		fmt.Fprintf(os.Stderr, "embed-model: wrote %s (%.2f MB)\n", dst, float64(n)/(1024*1024))
	}
	return dir, nil
}

// embedder abstracts vector embedding so unit tests can inject a deterministic
// mock without downloading model weights.
type embedder interface {
	Embed(text string) ([]float32, error)
}

type embedCache struct {
	path  string
	mu    sync.RWMutex
	items map[[32]byte][]float32
	dirty bool
}

const cacheMagic = "CIVEMB01"

func loadEmbedCache(path string) (*embedCache, error) {
	c := &embedCache{
		path:  path,
		items: make(map[[32]byte][]float32),
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 256*1024)
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		if err == io.EOF {
			return c, nil
		}
		return nil, err
	}
	if string(magic[:]) != cacheMagic {
		return c, nil
	}
	var dims uint32
	if err := binary.Read(r, binary.LittleEndian, &dims); err != nil {
		return c, nil
	}
	if dims == 0 || dims > 4096 {
		return c, nil
	}

	recBuf := make([]byte, 32+int(dims)*4)
	for {
		_, err := io.ReadFull(r, recBuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var key [32]byte
		copy(key[:], recBuf[:32])
		vec := make([]float32, dims)
		for i := 0; i < int(dims); i++ {
			bits := binary.LittleEndian.Uint32(recBuf[32+i*4 : 36+i*4])
			vec[i] = math.Float32frombits(bits)
		}
		c.items[key] = vec
	}
	return c, nil
}

func (c *embedCache) save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.dirty || c.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	var dims uint32
	for _, v := range c.items {
		dims = uint32(len(v))
		break
	}
	if dims == 0 {
		return nil
	}
	tmp := c.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriterSize(f, 256*1024)
	if _, err := w.WriteString(cacheMagic); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, dims); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	recBuf := make([]byte, 32+int(dims)*4)
	for k, v := range c.items {
		if uint32(len(v)) != dims {
			continue
		}
		copy(recBuf[:32], k[:])
		for i, x := range v {
			binary.LittleEndian.PutUint32(recBuf[32+i*4:36+i*4], math.Float32bits(x))
		}
		if _, err := w.Write(recBuf); err != nil {
			f.Close()
			os.Remove(tmp)
			return err
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, c.path)
}

func cacheKey(text string) [32]byte {
	return sha256.Sum256([]byte(embedModelTag + "\x00" + strings.TrimSpace(text)))
}

type semanticEngine struct {
	root    string
	cache   *embedCache
	model   embedder
	loadErr error
	once    sync.Once
}

func newSemanticEngine(root string) (*semanticEngine, error) {
	cache, err := loadEmbedCache(embedCachePath(root))
	if err != nil {
		return nil, err
	}
	return &semanticEngine{root: root, cache: cache}, nil
}

func (se *semanticEngine) ensureModel() (embedder, error) {
	se.once.Do(func() {
		if se.model != nil {
			return
		}
		dir, err := ensureEmbedModel(se.root, false)
		if err != nil {
			se.loadErr = err
			return
		}
		m, err := goformer.Load(dir)
		if err != nil {
			se.loadErr = err
			return
		}
		se.model = m
	})
	return se.model, se.loadErr
}

// embedAll returns L2-normalized vectors for all unique input strings, loading
// cached vectors from disk first and computing only misses concurrently across
// CPU workers.
func (se *semanticEngine) embedAll(texts []string) (map[string][]float32, error) {
	out := make(map[string][]float32, len(texts))
	var misses []string
	seenMiss := map[string]bool{}

	se.cache.mu.RLock()
	for _, raw := range texts {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if _, ok := out[t]; ok {
			continue
		}
		if vec, ok := se.cache.items[cacheKey(t)]; ok {
			out[t] = vec
			continue
		}
		if !seenMiss[t] {
			seenMiss[t] = true
			misses = append(misses, t)
		}
	}
	se.cache.mu.RUnlock()

	if len(misses) == 0 {
		return out, nil
	}

	model, err := se.ensureModel()
	if err != nil {
		return nil, err
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > len(misses) {
		workers = len(misses)
	}
	if len(misses) >= 50 {
		fmt.Fprintf(os.Stderr, "cadence -semantic: embedding %d uncached segments across %d Go workers...\n", len(misses), workers)
	}

	results := make([][]float32, len(misses))
	var firstErr error
	var errOnce sync.Once
	var nextIdx atomic.Int64
	var doneCnt atomic.Int64

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := int(nextIdx.Add(1) - 1)
				if i >= len(misses) {
					return
				}
				vec, e := model.Embed(misses[i])
				if e != nil {
					errOnce.Do(func() { firstErr = e })
					return
				}
				results[i] = vec
				done := int(doneCnt.Add(1))
				if len(misses) >= 200 && done%250 == 0 {
					fmt.Fprintf(os.Stderr, "  embedded %d / %d\n", done, len(misses))
				}
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	se.cache.mu.Lock()
	for i, t := range misses {
		if results[i] != nil {
			out[t] = results[i]
			se.cache.items[cacheKey(t)] = results[i]
			se.cache.dirty = true
		}
	}
	se.cache.mu.Unlock()

	if err := se.cache.save(); err != nil {
		return nil, err
	}
	return out, nil
}

func cosineSim(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		fa, fb := float64(a[i]), float64(b[i])
		dot += fa * fb
		na += fa * fa
		nb += fb * fb
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func extractParaSegments(text string) (sents []string, clausePairs [][2]string) {
	rawSents := splitCadenceSentences(text)
	for _, s := range rawSents {
		ws := reCadenceWord.FindAllString(s, -1)
		if len(ws) < 2 {
			continue
		}
		s = strings.TrimSpace(s)
		sents = append(sents, s)

		segs := reClausePivot.Split(s, -1)
		if len(segs) >= 2 {
			for i := 0; i < len(segs)-1; i++ {
				if segmentHasWeight(segs[i]) && segmentHasWeight(segs[i+1]) {
					clausePairs = append(clausePairs, [2]string{
						strings.TrimSpace(segs[i]),
						strings.TrimSpace(segs[i+1]),
					})
				}
			}
		}
	}
	return sents, clausePairs
}

// applySemantic enriches one or more paragraphs with semantic-leap and
// semantic-velocity metrics computed from pure-Go BGE-small-en-v1.5 embeddings.
func applySemantic(paras []cadencePara, se *semanticEngine) error {
	type paraSegs struct {
		sents       []string
		clausePairs [][2]string
	}
	segList := make([]paraSegs, len(paras))
	var allTexts []string
	for i, p := range paras {
		ss, cp := extractParaSegments(p.Text)
		segList[i] = paraSegs{sents: ss, clausePairs: cp}
		allTexts = append(allTexts, ss...)
		for _, pair := range cp {
			allTexts = append(allTexts, pair[0], pair[1])
		}
	}

	vecs, err := se.embedAll(allTexts)
	if err != nil {
		return err
	}

	for i := range paras {
		ss := segList[i].sents
		cp := segList[i].clausePairs
		m := &paras[i].Metrics

		var leaps []float64
		sumLeap := 0.0
		maxTurn := 0.0
		for j := 0; j < len(ss)-1; j++ {
			va, okA := vecs[ss[j]]
			vb, okB := vecs[ss[j+1]]
			if !okA || !okB {
				continue
			}
			d := 1.0 - cosineSim(va, vb)
			if d < 0 {
				d = 0
			}
			leaps = append(leaps, d)
			sumLeap += d
			if d > maxTurn {
				maxTurn = d
			}
		}

		var cLeaps []float64
		sumC := 0.0
		for _, pair := range cp {
			va, okA := vecs[pair[0]]
			vb, okB := vecs[pair[1]]
			if !okA || !okB {
				continue
			}
			d := 1.0 - cosineSim(va, vb)
			if d < 0 {
				d = 0
			}
			cLeaps = append(cLeaps, d)
			sumC += d
		}

		m.SemanticOK = true
		if len(leaps) > 0 {
			m.SentLeap = sumLeap / float64(len(leaps))
			m.MaxTurn = maxTurn
			m.EndTurn = leaps[len(leaps)-1]
			m.SemVel = sumLeap / float64(max(1, m.Words)) * 100.0
		}
		if len(cLeaps) > 0 {
			m.ClauseLeap = sumC / float64(len(cLeaps))
		}
		// Flag circular / treading-water prose: >=3 sentences with very low mean
		// sentence-to-sentence movement (calibrated for goformer BGE-small mean pooling).
		if len(ss) >= 3 && m.SentLeap > 0 && m.SentLeap < 0.28 {
			m.TreadWater = true
		}
	}
	return nil
}
