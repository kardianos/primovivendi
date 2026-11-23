package main

import (
	"math"
	"path/filepath"
	"strings"
	"testing"
)

type mockEmbedder struct {
	vecs map[string][]float32
}

func (m *mockEmbedder) Embed(text string) ([]float32, error) {
	low := strings.ToLower(text)
	for k, v := range m.vecs {
		if strings.Contains(low, k) {
			return v, nil
		}
	}
	// Deterministic fallback vector along axis 0.
	return []float32{1, 0, 0, 0}, nil
}

func TestSemanticCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.bin")

	c, err := loadEmbedCache(path)
	if err != nil {
		t.Fatalf("load empty cache: %v", err)
	}
	k := cacheKey("Science marches on.")
	c.items[k] = []float32{0.5, -0.5, 0.5, 0.5}
	c.dirty = true
	if err := c.save(); err != nil {
		t.Fatalf("save cache: %v", err)
	}

	c2, err := loadEmbedCache(path)
	if err != nil {
		t.Fatalf("reload cache: %v", err)
	}
	got, ok := c2.items[k]
	if !ok || len(got) != 4 {
		t.Fatalf("expected 4-dim vector from reloaded cache, got ok=%v len=%d", ok, len(got))
	}
	if math.Abs(float64(got[1])+0.5) > 1e-6 {
		t.Fatalf("unexpected value got[1]=%v", got[1])
	}
}

func TestSemanticLeapAndTreadWater(t *testing.T) {
	lx := mockCadenceLexicons()
	dir := t.TempDir()
	cache, _ := loadEmbedCache(filepath.Join(dir, "cache.bin"))

	se := &semanticEngine{
		cache: cache,
		model: &mockEmbedder{vecs: map[string][]float32{
			"numerous studies": {1, 0, 0, 0},
			"science marches":  {0, 1, 0, 0},
			"stone wall":       {0.99, 0.14, 0, 0},
			"stacked":          {0.98, 0.19, 0, 0},
			"held its shape":   {0.99, 0.12, 0, 0},
		}},
	}

	quipText := "Numerous studies have demonstrated that men are responsive to visual stimuli. Science marches on."
	treadText := "The mason built the stone wall. He stacked stones to make it. That stone wall held its shape."

	paras := []cadencePara{
		{Text: quipText, Metrics: measureCadence(quipText, lx)},
		{Text: treadText, Metrics: measureCadence(treadText, lx)},
	}
	if err := applySemantic(paras, se); err != nil {
		t.Fatalf("applySemantic: %v", err)
	}

	if !paras[0].Metrics.SemanticOK || paras[0].Metrics.EndTurn < 0.9 {
		t.Fatalf("expected high EndTurn on orthogonal quip, got %+v", paras[0].Metrics)
	}
	if !paras[1].Metrics.TreadWater || paras[1].Metrics.SentLeap >= 0.05 {
		t.Fatalf("expected TreadWater=true and near-zero SentLeap on circular paragraph, got %+v", paras[1].Metrics)
	}
}
