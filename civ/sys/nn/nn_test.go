package nn

import (
	"math"
	"path/filepath"
	"testing"
)

func TestRankerTrainAndSafetensorsRoundTrip(t *testing.T) {
	r := NewRanker(42)
	// Synthetic preference dataset: author strongly prefers higher linear score,
	// higher cadence snap, and better backward-pass tail handoff.
	var pairs []PairExample
	for i := 0; i < 80; i++ {
		var pos, neg [FeatureDim]float32
		pos[FLinearScore] = 0.85 + float32(i%5)*0.02
		pos[FCadenceSnap] = 0.80
		pos[FTailHandoffNext] = 0.75
		pos[FSemProgressGoal] = 0.80

		neg[FLinearScore] = 0.45 + float32(i%4)*0.02
		neg[FCadenceSnap] = 0.40
		neg[FTailHandoffNext] = 0.15
		neg[FSemProgressGoal] = 0.30

		pairs = append(pairs, PairExample{Pos: pos, Neg: neg, Weight: 1.5})
	}
	stats, err := r.Train(pairs, TrainConfig{Epochs: 35, LR: 5e-3, WeightDecay: 1e-2, BatchSize: 16, Seed: 7})
	if err != nil {
		t.Fatalf("Train failed: %v", err)
	}
	if stats.FinalLoss >= stats.InitialLoss {
		t.Fatalf("expected loss to decrease: init=%.4f final=%.4f", stats.InitialLoss, stats.FinalLoss)
	}
	if stats.PairAcc < 0.90 {
		t.Fatalf("expected pair accuracy >= 0.90, got %.3f", stats.PairAcc)
	}

	path := filepath.Join(t.TempDir(), "ranker.safetensors")
	if err := r.SaveSafetensors(path); err != nil {
		t.Fatalf("SaveSafetensors: %v", err)
	}
	loaded, err := LoadSafetensors(path)
	if err != nil {
		t.Fatalf("LoadSafetensors: %v", err)
	}
	if loaded.Trained != r.Trained || loaded.Step != r.Step {
		t.Fatalf("metadata mismatch: got trained=%d step=%d", loaded.Trained, loaded.Step)
	}
	s1 := r.Score(pairs[0].Pos)
	s2 := loaded.Score(pairs[0].Pos)
	if math.Abs(s1-s2) > 1e-5 {
		t.Fatalf("loaded score mismatch: %.6f vs %.6f", s1, s2)
	}
}
