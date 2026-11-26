// train.go — the training recipes behind `civ/bin/parser train-pos`,
// train-dep, train-ner, and train-all: which corpora are mixed with which
// weights, and where the models are written.
package parser

import (
	"fmt"
	"path/filepath"
	"time"
)

// Training mixture weights: gold treebanks count once each; the CoreNLP silver
// treatises and book chapters are weighted so in-domain punctuation and
// subordination are represented while the gold labels stay authoritative.
const (
	wEWT       = 1 // UD_English-EWT
	wGUM       = 1 // UD_English-GUM
	wTreatise  = 2 // pg*.txt.json (CoreNLP silver)
	wChapter   = 4 // ch01..ch12 (CoreNLP silver)
	wEWTForDep = 2
)

// POSEpochs is the default number of POS training epochs.
const POSEpochs = 12

// DepEpochs is the default number of dependency-parser training epochs.
const DepEpochs = 20

// NEREpochs is the number of NER training epochs.
const NEREpochs = 8

func modelPath(root, file string) string { return filepath.Join(ModelDir(root), file) }

// POSModelPath returns where the POS tagger is stored.
func POSModelPath(root string) string { return modelPath(root, POSModelFile) }

// DepModelPath returns where the dependency parser is stored.
func DepModelPath(root string) string { return modelPath(root, DepModelFile) }

// NERModelPath returns where the NER model is stored.
func NERModelPath(root string) string { return modelPath(root, NERModelFile) }

// loadGold loads the gold treebanks that have been fetched: UD_English-EWT
// (required) and UD_English-GUM (optional).
func loadGold(root, split string, logf func(string, ...any)) ([]*Sentence, error) {
	ewt, err := LoadEWT(root, split)
	if err != nil {
		return nil, err
	}
	for _, s := range ewt {
		s.Weight = 1
	}
	if gum, err := LoadGUM(root, split); err == nil {
		for _, s := range gum {
			s.Src = "gum-" + s.Src
		}
		logf("  gold: UD_English-EWT %d sentences + UD_English-GUM %d sentences (%s)", len(ewt), len(gum), split)
		return append(ewt, gum...), nil
	}
	logf("  gold: UD_English-EWT %d sentences (%s)", len(ewt), split)
	return ewt, nil
}

// loadSilverForTraining returns the distillation sentences with their weights:
// treatises at wTreatise, chapters ch01..ch12 at wChapter. The held-out
// chapters are never read.
func loadSilverForTraining(root string, logf func(string, ...any)) ([]*Sentence, error) {
	srcs := TrainingSources()
	silver, err := LoadSilver(root, srcs)
	if err != nil {
		return nil, err
	}
	n, bad := 0, 0
	for _, s := range silver {
		if s == nil {
			bad++
			continue
		}
		s.Weight = wTreatise
		if len(s.Src) > 2 && s.Src[:2] == "ch" {
			s.Weight = wChapter
		}
		n++
	}
	logf("  silver: %d CoreNLP sentences from %d sources (%d treatises, ch01..ch12; %d malformed)",
		n, len(srcs), len(TreatiseIDs), bad)
	return silver, nil
}

// TrainAndSavePOS trains the POS tagger with the recipe of `train-pos` and
// writes pos_tagger.safetensors. distill adds the CoreNLP silver data.
func TrainAndSavePOS(root string, distill bool, epochs int, logf func(string, ...any)) (*POSTrainResult, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	if epochs <= 0 {
		epochs = POSEpochs
	}
	train, err := loadGold(root, "train", logf)
	if err != nil {
		return nil, err
	}
	dev, err := LoadEWT(root, "dev")
	if err != nil {
		return nil, err
	}
	if distill {
		silver, err := loadSilverForTraining(root, logf)
		if err != nil {
			return nil, err
		}
		train = append(train, silver...)
	}
	lex := newLemmatizer(root)
	t0 := time.Now()
	logf("training POS tagger: %d sentences, %d epochs", len(train), epochs)
	res := TrainPOSWithLexicon(train, dev, POSTrainConfig{Epochs: epochs}, lex)
	res.Model.wn = lex
	path := POSModelPath(root)
	if err := res.Model.save(path, posModelFormat, map[string]string{
		"epochs": fmt.Sprint(epochs), "dev_acc": fmt.Sprintf("%.4f", res.BestAcc), "distill": fmt.Sprint(distill),
	}); err != nil {
		return nil, err
	}
	logf("saved %s (best epoch %d, UD_English-EWT dev accuracy %.2f%%, %s)",
		path, res.BestEpoch, 100*res.BestAcc, time.Since(t0).Round(time.Second))
	return res, nil
}

// TrainAndSaveNER trains the NER tagger on the CoreNLP silver data and writes
// ner.safetensors. The silver POS tags are replaced by the local tagger's
// predictions so the model sees the same tags at inference.
func TrainAndSaveNER(root string, logf func(string, ...any)) (*apModel, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	posPath, ok := ResolveModelPath(root, POSModelFile)
	if !ok {
		return nil, fmt.Errorf("train-ner needs the POS tagger (run train-pos first): %w", ErrNotTrained)
	}
	pos, err := loadAPModel(posPath, posModelFormat)
	if err != nil {
		return nil, fmt.Errorf("train-ner needs the POS tagger (run train-pos first): %w", err)
	}
	silver, err := loadSilverForTraining(root, logf)
	if err != nil {
		return nil, err
	}
	// Predicted POS tags in place of CoreNLP's.
	var (
		h     []tokHashes
		preds []int
		sc    = make([]float32, pos.nClass)
		feats [96]uint64
	)
	pos.wn = newLemmatizer(root)
	for _, s := range silver {
		if s == nil {
			continue
		}
		h = hashTokens(s.Words, h, pos.tagsets, pos.wn)
		preds = pos.tagTokens(h, preds, sc, feats[:])
		for i := range preds {
			s.POS[i] = PTBTags[preds[i]]
		}
	}
	t0 := time.Now()
	logf("training NER tagger: %d silver sentences, %d epochs", len(silver), NEREpochs)
	m := TrainNER(silver, NEREpochs, 20261125)
	path := NERModelPath(root)
	if err := m.save(path, nerModelFormat, map[string]string{"epochs": fmt.Sprint(NEREpochs)}); err != nil {
		return nil, err
	}
	logf("saved %s (%s)", path, time.Since(t0).Round(time.Second))
	return m, nil
}

// TrainAndSaveDep trains the dependency parser and writes
// dep_parser.safetensors. The dev set is UD_English-EWT dev with *predicted*
// POS tags, matching end-to-end inference. distill adds the CoreNLP silver
// data. An existing model at finetuneFrom (optional) is continued at a lower
// learning rate (Phase B of the two-phase recipe).
func TrainAndSaveDep(root string, distill bool, epochs int, logf func(string, ...any)) (*DepResult, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	if epochs <= 0 {
		epochs = DepEpochs
	}
	train, err := loadGold(root, "train", logf)
	if err != nil {
		return nil, err
	}
	for _, s := range train {
		s.Weight = wEWTForDep
	}
	if distill {
		silver, err := loadSilverForTraining(root, logf)
		if err != nil {
			return nil, err
		}
		train = append(train, silver...)
	}
	devGold, err := LoadEWT(root, "dev")
	if err != nil {
		return nil, err
	}
	dev, err := predictPOS(root, devGold)
	if err != nil {
		return nil, err
	}
	logf("training dependency parser: %d sentences, %d epochs (dev: UD_English-EWT dev with predicted POS)", len(train), epochs)
	cfg := DepTrainConfig{
		Epochs:      epochs,
		BatchSize:   512,
		LR:          3e-3,
		WeightDecay: 5e-4,
		Dropout:     0.30,
		WordDropout: 0.20,
		RareCount:   5,
		Dev:         dev,
		Log:         logf,
	}
	res, err := TrainDep(train, cfg)
	if err != nil {
		return nil, err
	}
	path := DepModelPath(root)
	if err := res.Model.save(path, map[string]string{
		"epochs": fmt.Sprint(epochs), "best_epoch": fmt.Sprint(res.BestEpoch), "distill": fmt.Sprint(distill),
		"dev_las": fmt.Sprintf("%.4f", res.BestLAS),
	}); err != nil {
		return nil, err
	}
	logf("saved %s (best epoch %d: dev UAS %.2f%% LAS %.2f%%, %s)",
		path, res.BestEpoch, 100*res.BestUAS, 100*res.BestLAS, res.TrainTime.Round(time.Second))
	return res, nil
}

// predictPOS returns copies of sents whose POS tags are the local tagger's
// predictions.
func predictPOS(root string, sents []*Sentence) ([]*Sentence, error) {
	posPath, ok := ResolveModelPath(root, POSModelFile)
	if !ok {
		return nil, fmt.Errorf("parser: need %s (run train-pos first): %w", POSModelPath(root), ErrNotTrained)
	}
	pos, err := loadAPModel(posPath, posModelFormat)
	if err != nil {
		return nil, fmt.Errorf("parser: need %s (run train-pos first): %w", posPath, err)
	}
	pos.wn = newLemmatizer(root)
	var (
		h     []tokHashes
		preds []int
		sc    = make([]float32, pos.nClass)
	)
	out := make([]*Sentence, 0, len(sents))
	for _, s := range sents {
		c := &Sentence{Words: s.Words, POS: make([]string, s.Len()), Head: s.Head, Rel: s.Rel, Src: s.Src}
		h = hashTokens(c.Words, h, pos.tagsets, pos.wn)
		preds = pos.tagTokens(h, preds, sc, nil)
		for i := range preds {
			c.POS[i] = PTBTags[preds[i]]
		}
		out = append(out, c)
	}
	return out, nil
}

// ModelsPresent reports which model files exist (locally or bundled).
func ModelsPresent(root string) (pos, dep, ner bool) {
	for _, m := range []struct {
		file string
		p    *bool
	}{{POSModelFile, &pos}, {DepModelFile, &dep}, {NERModelFile, &ner}} {
		if _, ok := ResolveModelPath(root, m.file); ok {
			*m.p = true
		}
	}
	return
}
