// Package embed wraps pure-Go BAAI/bge-small-en-v1.5 sentence embeddings via
// goformer, sharing the disk cache at ext-ref/embed/cache_bge_small.bin,
// accelerated with SIMD cosine/SimHash kernels (civ/sys/vec) and a non-blocking
// parallel worker pool for background prefetching.
package embed

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/MichaelAyles/goformer"
	"po/civ/sys/prof"
	"po/civ/sys/vec"
)

const (
	modelTag   = "bge-small-en-v1.5-goformer-v1"
	cacheMagic = "CIVEMB01"
)

// Engine provides cached local 384-dim sentence embeddings with non-blocking
// multi-core background prefetching and 64-bit bitwise SimHash signatures.
type Engine struct {
	root      string
	cachePath string

	mu       sync.RWMutex
	cache    map[[32]byte][]float32
	sigs     map[[32]byte]uint64
	inflight map[[32]byte]bool
	dirty    bool

	once    sync.Once
	model   *goformer.Model
	loadErr error

	poolOnce sync.Once
	jobs     chan string
	closed   atomic.Bool
}

// Open initializes an Engine using the shared ext-ref/embed/ cache and model.
// Model weights are loaded lazily on the first cache miss.
func Open(root string) *Engine {
	e := &Engine{
		root:      root,
		cachePath: filepath.Join(root, "ext-ref", "embed", "cache_bge_small.bin"),
		cache:     make(map[[32]byte][]float32),
		sigs:      make(map[[32]byte]uint64),
		inflight:  make(map[[32]byte]bool),
	}
	e.loadCache()
	return e
}

func cacheKey(text string) [32]byte {
	return sha256.Sum256([]byte(modelTag + "\x00" + strings.TrimSpace(text)))
}

// Ready reports whether the local BGE-small model directory exists on disk.
func (e *Engine) Ready() bool {
	if e == nil {
		return false
	}
	dir := filepath.Join(e.root, "ext-ref", "embed", "bge-small-en-v1.5")
	_, err := os.Stat(filepath.Join(dir, "model.safetensors"))
	return err == nil
}

func (e *Engine) ensureModel() (*goformer.Model, error) {
	e.once.Do(func() {
		defer prof.Span("embed.load_model")()
		dir := filepath.Join(e.root, "ext-ref", "embed", "bge-small-en-v1.5")
		m, err := goformer.Load(dir)
		if err != nil {
			e.loadErr = err
			return
		}
		e.model = m
	})
	return e.model, e.loadErr
}

// EmbedCached returns a cached embedding immediately in O(1) without blocking
// or running transformer inference.
func (e *Engine) EmbedCached(text string) ([]float32, bool) {
	if e == nil {
		return nil, false
	}
	t := strings.TrimSpace(text)
	if t == "" {
		return nil, false
	}
	k := cacheKey(t)
	e.mu.RLock()
	v, ok := e.cache[k]
	e.mu.RUnlock()
	if ok {
		prof.Count("embed.cache_hit")
		return v, true
	}
	return nil, false
}

// SimHashCached returns the cached 64-bit bitwise signature for text if available.
func (e *Engine) SimHashCached(text string) (uint64, bool) {
	if e == nil {
		return 0, false
	}
	t := strings.TrimSpace(text)
	if t == "" {
		return 0, false
	}
	k := cacheKey(t)
	e.mu.RLock()
	sig, ok := e.sigs[k]
	e.mu.RUnlock()
	return sig, ok
}

// Prefetch enqueues uncached texts onto a background multi-core worker pool so
// the UI thread never blocks on transformer inference. Duplicate or already
// cached texts are skipped in O(1).
func (e *Engine) Prefetch(texts []string) {
	if e == nil || len(texts) == 0 || e.closed.Load() || !e.Ready() {
		return
	}
	e.startWorkers()
	for _, raw := range texts {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		k := cacheKey(t)
		e.mu.Lock()
		if _, ok := e.cache[k]; ok || e.inflight[k] {
			e.mu.Unlock()
			continue
		}
		e.inflight[k] = true
		e.mu.Unlock()

		select {
		case e.jobs <- t:
			prof.Count("embed.prefetch_enqueued")
		default:
			// Queue full: drop quietly without stalling UI; will re-enqueue if revisited.
			e.mu.Lock()
			delete(e.inflight, k)
			e.mu.Unlock()
		}
	}
}

func (e *Engine) startWorkers() {
	e.poolOnce.Do(func() {
		workers := runtime.NumCPU() / 2
		if workers < 2 {
			workers = 2
		}
		if workers > 8 {
			workers = 8
		}
		e.jobs = make(chan string, 256)
		for w := 0; w < workers; w++ {
			go func() {
				for t := range e.jobs {
					if e.closed.Load() {
						continue
					}
					_, _ = e.Embed(t)
				}
			}()
		}
	})
}

// Embed returns the L2-normalized embedding for text, using the disk cache first.
// Returns (nil, err) if the model is unavailable.
func (e *Engine) Embed(text string) ([]float32, error) {
	if e == nil {
		return nil, nil
	}
	t := strings.TrimSpace(text)
	if t == "" {
		return nil, nil
	}
	k := cacheKey(t)
	e.mu.RLock()
	if v, ok := e.cache[k]; ok {
		e.mu.RUnlock()
		prof.Count("embed.cache_hit")
		return v, nil
	}
	e.mu.RUnlock()

	m, err := e.ensureModel()
	if err != nil {
		e.mu.Lock()
		delete(e.inflight, k)
		e.mu.Unlock()
		return nil, err
	}
	detail := t
	if len(detail) > 48 {
		detail = detail[:48] + "..."
	}
	done := prof.SpanDetail("embed.infer", detail)
	v, err := m.Embed(t)
	done()
	e.mu.Lock()
	delete(e.inflight, k)
	if err == nil && len(v) > 0 {
		e.cache[k] = v
		e.sigs[k] = vec.SimHash64(v)
		e.dirty = true
	}
	e.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return v, nil
}

// EmbedBatch computes embeddings for all uncached texts in parallel across CPU
// workers (useful for CLI passes or offline batch scoring).
func (e *Engine) EmbedBatch(texts []string) map[string][]float32 {
	out := make(map[string][]float32, len(texts))
	if e == nil || len(texts) == 0 {
		return out
	}
	var misses []string
	seenMiss := map[string]bool{}
	for _, raw := range texts {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if v, ok := e.EmbedCached(t); ok {
			out[t] = v
			continue
		}
		if !seenMiss[t] {
			seenMiss[t] = true
			misses = append(misses, t)
		}
	}
	if len(misses) == 0 || !e.Ready() {
		return out
	}
	workers := runtime.NumCPU() / 2
	if workers < 2 {
		workers = 2
	}
	if workers > 8 {
		workers = 8
	}
	if workers > len(misses) {
		workers = len(misses)
	}
	var idx atomic.Int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := int(idx.Add(1) - 1)
				if i >= len(misses) {
					return
				}
				if v, err := e.Embed(misses[i]); err == nil && len(v) > 0 {
					mu.Lock()
					out[misses[i]] = v
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	return out
}

// Save flushes any newly computed embeddings back to the shared cache file.
func (e *Engine) Save() error {
	if e == nil {
		return nil
	}
	e.closed.Store(true)
	e.mu.RLock()
	if !e.dirty || e.cachePath == "" {
		e.mu.RUnlock()
		return nil
	}
	snapshot := make(map[[32]byte][]float32, len(e.cache))
	var dims uint32
	for k, v := range e.cache {
		snapshot[k] = v
		if dims == 0 {
			dims = uint32(len(v))
		}
	}
	e.mu.RUnlock()
	if dims == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(e.cachePath), 0o755); err != nil {
		return err
	}
	tmp := e.cachePath + ".tmp"
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
	buf := make([]byte, 32+int(dims)*4)
	for k, v := range snapshot {
		if uint32(len(v)) != dims {
			continue
		}
		copy(buf[:32], k[:])
		for i, x := range v {
			binary.LittleEndian.PutUint32(buf[32+i*4:36+i*4], math.Float32bits(x))
		}
		if _, err := w.Write(buf); err != nil {
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
	if err := os.Rename(tmp, e.cachePath); err != nil {
		os.Remove(tmp)
		return err
	}
	e.mu.Lock()
	e.dirty = false
	e.mu.Unlock()
	return nil
}

func (e *Engine) loadCache() {
	f, err := os.Open(e.cachePath)
	if err != nil {
		return
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 256*1024)
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil || string(magic[:]) != cacheMagic {
		return
	}
	var dims uint32
	if err := binary.Read(r, binary.LittleEndian, &dims); err != nil || dims == 0 || dims > 4096 {
		return
	}
	rec := make([]byte, 32+int(dims)*4)
	for {
		if _, err := io.ReadFull(r, rec); err != nil {
			break
		}
		var k [32]byte
		copy(k[:], rec[:32])
		v := make([]float32, dims)
		for i := 0; i < int(dims); i++ {
			v[i] = math.Float32frombits(binary.LittleEndian.Uint32(rec[32+i*4 : 36+i*4]))
		}
		e.cache[k] = v
		e.sigs[k] = vec.SimHash64(v)
	}
}

// Cosine returns the SIMD-accelerated cosine similarity in [-1, 1] between two embeddings.
func Cosine(a, b []float32) float64 {
	return vec.Cosine(a, b)
}
