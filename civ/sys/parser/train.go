// train.go — the training recipes behind `civ/bin/parser train-pos`,
// train-dep, train-ner, and train-all: which corpora are mixed with which
// weights, and where the models are written.
package parser

import (
	"fmt"
	"math"
	"path/filepath"
	"time"

	"po/civ/sys/corenlp"
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
// ner.safetensors. The silver POS tags are replaced by the pipeline's own
// predictions (the unified model's layer 1, or the Round-1 tagger) so the
// model sees the same tags at inference. The silver set is the eight
// treatises, ch01..ch11, and the sealed drafts; ch12 stays a clean dev set
// for Stage 4 and ch13/ch14 are never read.
func TrainAndSaveNER(root string, logf func(string, ...any)) (*apModel, *DraftLoad, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	tag, err := loadPOSTagger(root)
	if err != nil {
		return nil, nil, fmt.Errorf("train-ner needs the tagger (run train-parser first): %w", err)
	}
	treatises, chapters, drafts, info, err := LoadSilverCorpus(root)
	if err != nil {
		return nil, nil, err
	}
	silver := append(append(treatises, chapters...), drafts...)
	n, bad := 0, 0
	for _, s := range silver {
		if s == nil {
			bad++
		} else {
			n++
		}
	}
	logf("  silver: %d CoreNLP sentences (%d treatises, ch01..ch11, %d unsealed draft sentences; %d malformed)", n, len(TreatiseIDs), info.Sentences, bad)
	// Predicted POS tags in place of CoreNLP's.
	for _, s := range silver {
		if s == nil {
			continue
		}
		preds := tag(s.Words)
		for i := range preds {
			s.POS[i] = PTBTags[preds[i]]
		}
	}
	t0 := time.Now()
	logf("training NER tagger: %d silver sentences, %d epochs", n, NEREpochs)
	m := TrainNER(silver, NEREpochs, 20261125)
	path := NERModelPath(root)
	if err := m.save(path, nerModelFormat, map[string]string{"epochs": fmt.Sprint(NEREpochs), "drafts": fmt.Sprint(info.Sentences)}); err != nil {
		return nil, nil, err
	}
	logf("saved %s (%s)", path, time.Since(t0).Round(time.Second))
	return m, &info, nil
}

// loadPOSTagger returns a tagger of surface forms (PTB class ids): the
// unified model's layer 1 when parser.safetensors exists, else the Round-1
// perceptron. The returned function is for single-threaded training use.
func loadPOSTagger(root string) (func(words []string) []int, error) {
	if path, ok := ResolveModelPath(root, ParserModelFile); ok {
		m, err := loadBiaffineModel(path)
		if err == nil {
			err = m.checkInventories()
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		e := newBiEngine(m, newLemmatizer(root))
		st := &biState{}
		return func(words []string) []int {
			pos, _, _ := e.parse(words, st)
			out := make([]int, len(words))
			for i := range out {
				out[i] = int(pos[i])
			}
			return out
		}, nil
	}
	posPath, ok := ResolveModelPath(root, POSModelFile)
	if !ok {
		return nil, fmt.Errorf("parser: need %s (run train-parser first): %w", POSModelPath(root), ErrNotTrained)
	}
	pos, err := loadAPModel(posPath, posModelFormat)
	if err != nil {
		return nil, fmt.Errorf("parser: need %s (run train-parser first): %w", posPath, err)
	}
	pos.wn = newLemmatizer(root)
	var (
		h     []tokHashes
		preds []int
		sc    = make([]float32, pos.nClass)
		feats [96]uint64
	)
	return func(words []string) []int {
		h = hashTokens(words, h, pos.tagsets, pos.wn)
		preds = pos.tagTokens(h, preds[:0], sc, feats[:])
		return append([]int(nil), preds...)
	}, nil
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

// ModelsPresent reports which model files exist (locally or bundled). The
// Round-2 unified model provides both the tagger and the parser.
func ModelsPresent(root string) (pos, dep, ner bool) {
	if _, ok := ResolveModelPath(root, ParserModelFile); ok {
		pos, dep = true, true
	} else {
		for _, m := range []struct {
			file string
			p    *bool
		}{{POSModelFile, &pos}, {DepModelFile, &dep}} {
			if _, ok := ResolveModelPath(root, m.file); ok {
				*m.p = true
			}
		}
	}
	_, ner = ResolveModelPath(root, NERModelFile)
	return
}

// ---- Round 2: the unified parser ---------------------------------------------

// ParserModelPath returns where the Round-2 unified model is stored.
func ParserModelPath(root string) string { return modelPath(root, ParserModelFile) }

// ParserTrainOptions are the flags of `civ/bin/parser train-parser`.
type ParserTrainOptions struct {
	EpochsA  int    // Phase A epochs (0: 20)
	PhaseB   bool   // run Phase B when dev_book lags dev_ud by more than 1.5 LAS points
	EpochsB  int    // Phase B epochs (0: 5)
	LinES    bool   // add UD_English-LinES to the gold mix (CC BY-NC-SA 4.0)
	Workers  int    // (0: min(NumCPU, 8))
	Seed     uint64 // (0: the default)
	InitPath string // optional: start from a saved model of the same vocabulary
	// Silver is the weight of ch01..ch11 in Phase A (0: 2, the
	// specification's). The treatises and drafts always have weight 1.
	Silver int
	// DevEvery selects a checkpoint every epoch (default); 0 keeps the last.
	DevEvery int
}

// ParserTrainResult summarizes the recipe.
type ParserTrainResult struct {
	Model      *biaffineModel
	Phase      string // "A" or "B"
	BestEpoch  int
	DevUD      DevScores
	DevBook    Agreement
	History    []string
	Drafts     DraftLoad
	GoldLines  int
	TrainTime  time.Duration
	PhaseBSkip string // why Phase B was not run ("" when it was)
}

// parserGoldMix loads the gold treebanks: UD-EWT (required), UD-GUM and
// UD-LinES when they are downloaded (LinES only on request: its license is
// non-commercial).
func parserGoldMix(root string, withLinES bool, logf func(string, ...any)) ([]*Sentence, int, error) {
	ewt, err := LoadEWT(root, "train")
	if err != nil {
		return nil, 0, err
	}
	gold := ewt
	logf("  gold: UD_English-EWT train %d sentences", len(ewt))
	if GUMAvailable(root) {
		gum, err := LoadGUM(root, "train")
		if err != nil {
			return nil, 0, err
		}
		for _, s := range gum {
			s.Src = "gum-" + s.Src
		}
		gold = append(gold, gum...)
		logf("  gold: UD_English-GUM train %d sentences", len(gum))
	}
	if withLinES {
		if !LinESAvailable(root) {
			return nil, 0, fmt.Errorf("UD_English-LinES is not downloaded (run: civ/bin/parser fetch-ud -lines)")
		}
		lines, err := LoadLinES(root, "train")
		if err != nil {
			return nil, 0, err
		}
		gold = append(gold, lines...)
		logf("  gold: UD_English-LinES train %d sentences (CC BY-NC-SA 4.0)", len(lines))
		return gold, len(lines), nil
	}
	return gold, 0, nil
}

// parserCorpus is the Round-2 training corpus: the gold treebanks (weight
// 1), the treatises (1), the sealed drafts (1), and ch01..ch11 (2).
func parserCorpus(root string, withLinES bool, logf func(string, ...any)) (gold, treatises, drafts, chapters []*Sentence, info DraftLoad, err error) {
	gold, _, err = parserGoldMix(root, withLinES, logf)
	if err != nil {
		return nil, nil, nil, nil, info, err
	}
	treatises, chapters, drafts, info, err = LoadSilverCorpus(root)
	if err != nil {
		return nil, nil, nil, nil, info, err
	}
	return gold, treatises, drafts, chapters, info, nil
}

// newDevPipeline builds a minimal pipeline around a model for dev_book.
func newDevPipeline(m *biaffineModel, lem *lemmatizer) *Pipeline {
	p := &Pipeline{POS: "go", uni: newBiEngine(m, lem), lem: lem, sem: make(chan struct{}, 1)}
	p.initPools(64)
	return p
}

// DevBookScores parses dev_book (ch12) with the model and scores it against
// CoreNLP's parse.
func DevBookScores(m *biaffineModel, lem *lemmatizer, cn *corenlp.Doc, text string) *Agreement {
	p := newDevPipeline(m, lem)
	ag := newAgreement()
	ag.add(p.ParseDoc(text), cn)
	ag.finish()
	return ag
}

// TrainAndSaveParser runs the Round-2 recipe and writes parser.safetensors.
func TrainAndSaveParser(root string, opt ParserTrainOptions, logf func(string, ...any)) (*ParserTrainResult, error) {
	if logf == nil {
		logf = func(string, ...any) {}
	}
	if opt.EpochsA <= 0 {
		opt.EpochsA = 30
	}
	if opt.EpochsB <= 0 {
		opt.EpochsB = 5
	}
	if opt.Seed == 0 {
		opt.Seed = 20261126
	}
	if opt.DevEvery == 0 {
		opt.DevEvery = 1
	}
	if opt.Silver <= 0 {
		opt.Silver = 2
	}
	gold, treatises, drafts, chapters, info, err := parserCorpus(root, opt.LinES, logf)
	if err != nil {
		return nil, err
	}
	logf("  silver: %d treatise, %d chapter (ch01..ch11) sentences; drafts: %d files, %d sealed (ch13/ch14 overlap), %d sealed (ch12 overlap), %d unreadable, %d loaded, %d sentences after dropping %d verbatim",
		len(treatises), len(chapters), info.Files, info.Sealed, info.SealedDev, info.Unreadable, info.Loaded, info.Sentences, info.Deduped)
	lem := newLemmatizer(root)
	devUDGold, err := LoadEWT(root, "dev")
	if err != nil {
		return nil, err
	}
	devBookDoc, devBookText, err := LoadBookDev(root)
	if err != nil {
		return nil, err
	}
	// The model and its vocabulary: raw counts over the unique sentences.
	unique := make([]*Sentence, 0, len(gold)+len(treatises)+len(drafts)+len(chapters))
	unique = append(append(append(unique, gold...), treatises...), drafts...)
	unique = append(unique, chapters...)
	var m *biaffineModel
	cfg := DefaultBiaffineConfig()
	cfg.Seed = opt.Seed
	if opt.Workers > 0 {
		cfg.Workers = opt.Workers
	}
	if opt.InitPath != "" {
		if m, err = loadBiaffineModel(opt.InitPath); err == nil {
			err = m.checkInventories()
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", opt.InitPath, err)
		}
		// The saved model keeps its own dimensions and vocabulary.
		cfg.Dims = m.dims
	} else {
		m = newBiaffineModel(biaffineWordVocab(unique, cfg.MaxVocab), cfg.Dims, len(PTBTags), len(DepLabels), newRand(opt.Seed))
	}
	m.cacheForms = biaffineCacheForms(unique, charCacheForms)
	counts := m.biaffineWordCounts(unique)
	goldEnc, _ := m.encode(gold, lem, false)
	treEnc, _ := m.encode(treatises, lem, true)
	drfEnc, _ := m.encode(drafts, lem, true)
	chaEnc, _ := m.encode(chapters, lem, true)
	devUD, _ := m.encode(devUDGold, lem, false)
	corpus := make([]*biSent, 0, len(goldEnc)+len(treEnc)+len(drfEnc)+len(chaEnc))
	weights := make([]int, 0, len(corpus))
	add := func(ss []*biSent, w int) {
		for _, s := range ss {
			corpus, weights = append(corpus, s), append(weights, w)
		}
	}
	add(goldEnc, 1)
	add(treEnc, 1)
	add(drfEnc, 1)
	add(chaEnc, opt.Silver)
	logf("  mix: gold 1x, treatises 1x, drafts 1x, ch01..ch11 %dx", opt.Silver)
	res := &ParserTrainResult{Drafts: info, Phase: "A"}
	t0 := time.Now()
	// selector rates a checkpoint by 0.5·dev_ud LAS + 0.5·dev_book non-punct
	// LAS (both logged every epoch); below the floor for dev_ud LAS it is
	// never kept.
	selector := func(minUDLAS float64) func(*biaffineModel, int) (float64, string) {
		return func(m *biaffineModel, ep int) (float64, string) {
			ud := BiaffineEvaluate(m, devUD)
			book := DevBookScores(m, lem, devBookDoc, devBookText)
			report := fmt.Sprintf("dev_ud %s | dev_book non-punct UAS %.2f%% LAS %.2f%% (all-token UAS %.2f%% / LAS %.2f%%), root %.2f%%",
				ud, 100*book.UASNonPunct, 100*book.LASNonPunct, 100*book.UASAligned, 100*book.LASAligned, 100*book.Root)
			if ud.LAS < minUDLAS {
				return math.Inf(-1), report + " | rejected: dev_ud LAS below the floor"
			}
			return 0.5*ud.LAS + 0.5*book.LASNonPunct, report
		}
	}
	devOf := func(m *biaffineModel) (DevScores, Agreement) {
		ud := BiaffineEvaluate(m, devUD)
		book := DevBookScores(m, lem, devBookDoc, devBookText)
		return ud, *book
	}
	cfg.Epochs, cfg.Log, cfg.Select = opt.EpochsA, logf, selector(0)
	ra, err := TrainBiaffine(m, corpus, weights, counts, cfg)
	if err != nil {
		return nil, err
	}
	res.Model, res.BestEpoch, res.History = ra.Model, ra.BestEpoch, ra.History
	res.DevUD, res.DevBook = devOf(res.Model)
	logf("phase A: best epoch %d (selection score %.4f), dev_ud %s, dev_book non-punct LAS %.2f%%, %s",
		ra.BestEpoch, ra.Score, res.DevUD, 100*res.DevBook.LASNonPunct, time.Since(t0).Round(time.Second))
	if opt.PhaseB {
		lag := 100 * (res.DevUD.LAS - res.DevBook.LASNonPunct)
		if lag <= 1.5 {
			res.PhaseBSkip = fmt.Sprintf("dev_book non-punct LAS lags dev_ud LAS by only %.2f points (threshold 1.5)", lag)
			logf("phase B skipped: %s", res.PhaseBSkip)
		} else {
			// A balanced mix, closer to the book: UD-EWT 1, GUM 1 (the gold
			// list), treatises 2, sealed drafts 2, ch01..ch11 3.
			var corpusB []*biSent
			var weightsB []int
			addB := func(ss []*biSent, w int) {
				for _, s := range ss {
					corpusB, weightsB = append(corpusB, s), append(weightsB, w)
				}
			}
			addB(goldEnc, 1)
			addB(treEnc, 2)
			addB(drfEnc, 2)
			addB(chaEnc, 3)
			cfgB := cfg
			cfgB.Epochs, cfgB.LR, cfgB.FinalLR = opt.EpochsB, 3e-4, 1e-5
			// Kept only while dev_ud LAS stays within half a point of Phase A.
			cfgB.Select = selector(res.DevUD.LAS - 0.005)
			rb, err := TrainBiaffine(ra.Model, corpusB, weightsB, counts, cfgB)
			if err != nil {
				return nil, err
			}
			res.History = append(res.History, rb.History...)
			if rb.Score > math.Inf(-1) {
				res.Model, res.BestEpoch, res.Phase = rb.Model, rb.BestEpoch, "B"
				res.DevUD, res.DevBook = devOf(res.Model)
				logf("phase B: best epoch %d (selection score %.4f), dev_ud %s, dev_book non-punct LAS %.2f%%", rb.BestEpoch, rb.Score, res.DevUD, 100*res.DevBook.LASNonPunct)
			} else {
				res.PhaseBSkip = "no Phase B epoch kept dev_ud LAS within 0.5 points of Phase A"
				logf("phase B rejected: %s", res.PhaseBSkip)
			}
		}
	}
	path := ParserModelPath(root)
	meta := map[string]string{
		"phase": res.Phase, "best_epoch": fmt.Sprint(res.BestEpoch), "epochs_a": fmt.Sprint(opt.EpochsA),
		"dev_ud_las": fmt.Sprintf("%.4f", res.DevUD.LAS), "dev_book_las_np": fmt.Sprintf("%.4f", res.DevBook.LASNonPunct),
		"lines": fmt.Sprint(opt.LinES), "drafts": fmt.Sprint(info.Sentences), "seed": fmt.Sprint(opt.Seed),
		"silver": fmt.Sprint(opt.Silver),
	}
	if err := res.Model.save(path, meta); err != nil {
		return nil, err
	}
	res.TrainTime = time.Since(t0)
	logf("saved %s (phase %s, epoch %d, %d parameters, %s)", path, res.Phase, res.BestEpoch, res.Model.numParams(), res.TrainTime.Round(time.Second))
	return res, nil
}
