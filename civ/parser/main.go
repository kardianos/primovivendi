// Command parser trains, evaluates, and runs the pure-Go NLP pipeline that
// replaces Stanford CoreNLP for civ/compose and civ/linear (see
// civ/notes/47_go_parser_spec.md, civ/notes/48_go_parser_r2_spec.md, and
// civ/notes/48a_go_parser_r2_handoff.md).
//
// Pre-trained F16+gzip models are committed under civ/sys/parser/models/
// (parser.safetensors.gz + ner.safetensors.gz, ~3.79 MB total) and loaded
// automatically when no local retrain exists under ext-ref/parser/.
//
// Usage, from civ/:
//
//	./bin/parser fetch-ud [-lines]                    # download the gold treebanks
//	./bin/parser train-parser [-epochs 30] [-phase-b] # train & save parser.safetensors (BiGRU + biaffine)
//	./bin/parser train-ner                            # train & save ner.safetensors
//	./bin/parser train-all [-epochs 30] [-phase-b]    # fetch check, then train parser and NER
//	./bin/parser pack [-f16=true] [-out DIR]          # compress ext-ref/parser/*.safetensors to civ/sys/parser/models/
//	./bin/parser eval [-chapter ch13,ch14] [-models auto|bundled|local]
//	./bin/parser parse "We come into this world naked."
//	./bin/parser bench [-chapter ch13] [-reps 3]      # speed and memory
//	./bin/parser part4 [-chapter ch01,...,ch12]       # Stage 4 diagnostics against CoreNLP
//
// ch13 and ch14 are held out: only eval reads them (the training loaders read
// their raw text for nothing but sealing drafts that share a sentence with
// them). ch12 is the Round-2 in-domain dev set and is kept out of training.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"po/civ/sys/parser"
	"po/civ/sys/vec"
)

func main() {
	log.SetFlags(log.Ltime)
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	root, err := repoRoot()
	if err != nil {
		log.Fatal(err)
	}
	var err2 error
	switch cmd {
	case "help", "-h", "--help":
		usage()
	case "fetch-ud":
		err2 = cmdFetchUD(root, args)
	case "train-parser":
		err2 = cmdTrainParser(root, args)
	case "train-ner":
		err2 = cmdTrainNER(root, args)
	case "train-all":
		err2 = cmdTrainAll(root, args)
	case "train-pos":
		err2 = cmdTrainPOS(root, args)
	case "train-dep":
		err2 = cmdTrainDep(root, args)
	case "eval":
		err2 = cmdEval(root, args)
	case "parse":
		err2 = cmdParse(root, args)
	case "bench":
		err2 = cmdBench(root, args)
	case "pack":
		err2 = cmdPack(root, args)
	case "part4":
		err2 = cmdPart4(root, args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
	if err2 != nil {
		log.Fatal(err2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: civ/bin/parser <command> [flags]

  fetch-ud [-lines]                    download UD_English-EWT and UD_English-GUM (and, on request,
                                      UD_English-LinES, CC BY-NC-SA 4.0) into ext-ref/parser/
  train-parser [-epochs 30] [-phase-b] [-epochs-b 5] [-lines] [-seed S] [-workers W] [-silver 2]
                                      train & save parser.safetensors: the Round-2 BiGRU + biaffine
                                      tagger and parser (Phase A, default 30 epochs, optional Phase B);
                                      -silver weighs ch01..ch11 in the mix (default 2)
  train-ner                            train & save ner.safetensors (treatises, ch01..ch11, sealed drafts)
  train-all [-epochs 30] [-phase-b]    fetch check, then train parser and NER
  train-pos [-distill] [-epochs N]     (Round 1) train & save pos_tagger.safetensors
  train-dep [-distill] [-epochs N]     (Round 1) train & save dep_parser.safetensors
  eval [-chapter ch13,ch14] [-ud-split test|dev] [-models auto|bundled|local] [-drafts N] [-no-book]
                                      the gate table: UD, held-out CoreNLP parity, Rule 10, speed, size
  parse [-n 1] "sentence"              print the dependency tree and latency
  bench [-chapter ch13] [-reps 3]      sentences/sec, latency percentiles, memory
  pack [-f16=true] [-out DIR]          compress ext-ref/parser/*.safetensors into civ/sys/parser/models/*.safetensors.gz
  part4 [-chapter ch01,...,ch12]       Stage 4 diagnostics: every Part 4 finding where Go and CoreNLP differ

Models resolve from ext-ref/parser/ first, falling back to bundled civ/sys/parser/models/*.safetensors.gz.
ch13 and ch14 are held out: only eval reads them. ch12 is the Round-2 dev set.
`)
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .git directory above the working directory")
		}
		dir = parent
	}
}

func logf(format string, args ...any) { log.Printf(format, args...) }

// cachedInputs returns the cached CoreNLP input texts by source name (the
// chapters and treatises under ext-ref/linear/in).
func cachedInputs(root string) (map[string]string, error) {
	paths, err := filepath.Glob(filepath.Join(root, "ext-ref", "linear", "in", "*.txt"))
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		out[strings.TrimSuffix(filepath.Base(p), ".txt")] = string(b)
	}
	return out, nil
}

func containsString(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// ---- fetch-ud ----------------------------------------------------------------

type treebank struct {
	name  string
	files []string
}

// udRelease is the pinned Universal Dependencies release tag.
const udRelease = "r2.15"

var treebanks = []treebank{
	{"UD_English-EWT", []string{"en_ewt-ud-train.conllu", "en_ewt-ud-dev.conllu", "en_ewt-ud-test.conllu"}},
	{"UD_English-GUM", []string{"en_gum-ud-train.conllu", "en_gum-ud-dev.conllu", "en_gum-ud-test.conllu"}},
}

// linesTreebank is optional: its text is licensed CC BY-NC-SA 4.0.
var linesTreebank = treebank{"UD_English-LinES", []string{"en_lines-ud-train.conllu", "en_lines-ud-dev.conllu", "en_lines-ud-test.conllu"}}

func cmdFetchUD(root string, args []string) error {
	fs := flag.NewFlagSet("fetch-ud", flag.ExitOnError)
	force := fs.Bool("force", false, "re-download files that already exist")
	withLines := fs.Bool("lines", false, "also fetch UD_English-LinES (CC BY-NC-SA 4.0)")
	fs.Parse(args)
	dirOf := map[string]string{
		"UD_English-EWT":   filepath.Join(root, "ext-ref", "parser", "ud_ewt"),
		"UD_English-GUM":   filepath.Join(root, "ext-ref", "parser", "ud_gum"),
		"UD_English-LinES": filepath.Join(root, "ext-ref", "parser", "ud_lines"),
	}
	list := append([]treebank{}, treebanks...)
	if *withLines {
		list = append(list, linesTreebank)
	}
	for _, tb := range list {
		dir := dirOf[tb.name]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for _, f := range tb.files {
			dst := filepath.Join(dir, f)
			if !*force {
				if st, err := os.Stat(dst); err == nil && st.Size() > 1000 {
					log.Printf("%s/%s: present", tb.name, f)
					continue
				}
			}
			url := fmt.Sprintf("https://raw.githubusercontent.com/UniversalDependencies/%s/%s/%s", tb.name, udRelease, f)
			if err := download(url, dst); err != nil {
				return fmt.Errorf("%s: %w", url, err)
			}
			st, _ := os.Stat(dst)
			log.Printf("%s/%s: %d bytes", tb.name, f, st.Size())
		}
	}
	return nil
}

func download(url, dst string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "civ-parser/2.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", resp.Status)
	}
	tmp := dst + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// ---- training ----------------------------------------------------------------

func cmdTrainParser(root string, args []string) error {
	fs := flag.NewFlagSet("train-parser", flag.ExitOnError)
	opt := parser.ParserTrainOptions{}
	fs.IntVar(&opt.EpochsA, "epochs", 0, "Phase A epochs (0 for 30)")
	fs.BoolVar(&opt.PhaseB, "phase-b", false, "run Phase B when dev_book lags dev_ud by more than 1.5 LAS points")
	fs.IntVar(&opt.EpochsB, "epochs-b", 0, "Phase B epochs (0 for 5)")
	fs.BoolVar(&opt.LinES, "lines", false, "add UD_English-LinES (CC BY-NC-SA 4.0) to the gold mix")
	fs.Uint64Var(&opt.Seed, "seed", 0, "random seed (0 for the default)")
	fs.IntVar(&opt.Workers, "workers", 0, "training goroutines (0 for min(NumCPU, 8))")
	fs.StringVar(&opt.InitPath, "init", "", "start from a saved parser.safetensors of the same vocabulary")
	fs.IntVar(&opt.DevEvery, "dev-every", 1, "select a checkpoint every N epochs")
	fs.IntVar(&opt.Silver, "silver", 0, "weight of ch01..ch11 in the mix (0 for 2)")
	fs.Parse(args)
	if err := requireUD(root); err != nil {
		return err
	}
	_, err := parser.TrainAndSaveParser(root, opt, logf)
	return err
}

func cmdTrainNER(root string, args []string) error {
	flag.NewFlagSet("train-ner", flag.ExitOnError).Parse(args)
	if pos, _, _ := parser.ModelsPresent(root); !pos {
		return fmt.Errorf("train-ner needs the tagger (run: civ/bin/parser train-parser)")
	}
	_, _, err := parser.TrainAndSaveNER(root, logf)
	return err
}

func cmdTrainAll(root string, args []string) error {
	fs := flag.NewFlagSet("train-all", flag.ExitOnError)
	epochs := fs.Int("epochs", 0, "Phase A epochs (0 for 30)")
	phaseB := fs.Bool("phase-b", false, "run Phase B when dev_book lags dev_ud by more than 1.5 LAS points")
	lines := fs.Bool("lines", false, "add UD_English-LinES to the gold mix")
	seed := fs.Uint64("seed", 0, "random seed (0 for the default)")
	workers := fs.Int("workers", 0, "training goroutines (0 for min(NumCPU, 8))")
	fs.Parse(args)
	opt := parser.ParserTrainOptions{EpochsA: *epochs, PhaseB: *phaseB, LinES: *lines, Seed: *seed, Workers: *workers}
	if err := cmdFetchUD(root, []string{"-force=false"}); err != nil {
		return err
	}
	if _, err := parser.TrainAndSaveParser(root, opt, logf); err != nil {
		return err
	}
	_, _, err := parser.TrainAndSaveNER(root, logf)
	return err
}

func cmdTrainPOS(root string, args []string) error {
	fs := flag.NewFlagSet("train-pos", flag.ExitOnError)
	distill := fs.Bool("distill", false, "mix the CoreNLP silver prose (ch01..ch12) into training")
	epochs := fs.Int("epochs", 0, "epochs (0 for the default)")
	fs.Parse(args)
	if err := requireUD(root); err != nil {
		return err
	}
	_, err := parser.TrainAndSavePOS(root, *distill, *epochs, logf)
	return err
}

func cmdTrainDep(root string, args []string) error {
	fs := flag.NewFlagSet("train-dep", flag.ExitOnError)
	distill := fs.Bool("distill", false, "mix the CoreNLP silver prose (ch01..ch12) into training")
	epochs := fs.Int("epochs", 0, "epochs (0 for the default)")
	fs.Parse(args)
	if err := requireUD(root); err != nil {
		return err
	}
	if pos, _, _ := parser.ModelsPresent(root); !pos {
		return fmt.Errorf("train-dep needs the POS tagger (run: civ/bin/parser train-pos)")
	}
	_, err := parser.TrainAndSaveDep(root, *distill, *epochs, logf)
	return err
}

func requireUD(root string) error {
	if !parser.EWTAvailable(root) {
		return fmt.Errorf("UD_English-EWT is not downloaded (run: civ/bin/parser fetch-ud)")
	}
	return nil
}

// ---- eval ---------------------------------------------------------------------

// round1 holds the Round-1 figures of the gate table (47a_go_parser_handoff.md).
type gate struct {
	name, round1, gate string
	// pass reports whether value meets the gate.
	pass func(v float64) bool
}

func cmdEval(root string, args []string) error {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	chapter := fs.String("chapter", "ch13,ch14", "held-out CoreNLP chapters (comma separated)")
	linear := fs.String("linear", "", "path to the civ/linear binary (default civ/bin/linear)")
	reps := fs.Int("reps", 5, "timing repetitions per sentence")
	drafts := fs.Int("drafts", 0, "also compare with CoreNLP on drafts, on sentences absent from the chapters and treatises (up to N drafts; 0 skips)")
	udSplit := fs.String("ud-split", "test", "UD_English-EWT split for the UD rows (test is the official one; dev for development)")
	models := fs.String("models", "auto", "auto: local overrides then bundled; bundled: civ/sys/parser/models only; local: ext-ref/parser only")
	noBook := fs.Bool("no-book", false, "skip the full-book timing (development runs that keep the holdout unread)")
	fs.Parse(args)
	chapters := []string{}
	for _, c := range strings.Split(*chapter, ",") {
		if c = strings.TrimSpace(c); c != "" {
			chapters = append(chapters, c)
		}
	}
	if err := requireUD(root); err != nil {
		return err
	}
	official := strings.Join(chapters, ",") == "ch13,ch14" && *udSplit == "test"
	t0 := time.Now()
	var p *parser.Pipeline
	var err error
	switch *models {
	case "bundled":
		p, err = parser.LoadBundled(root)
	case "local":
		p, err = parser.LoadLocal(root)
	default:
		p, err = parser.Load(root)
	}
	if err != nil {
		return fmt.Errorf("%w (run: civ/bin/parser train-all)", err)
	}
	loadTime := time.Since(t0)
	heapMB, rssMB := parser.Memory()
	mode := *models
	if mode == "auto" {
		mode = "local+bundled"
	}
	log.Printf("models loaded in %s from %s (heap %.1f MB, RSS %.1f MB, simd=%v, unified=%v)", loadTime.Round(time.Millisecond), mode, heapMB, rssMB, vec.SIMDEnabled(), p.Unified())

	log.Printf("evaluating on UD_English-EWT %s ...", *udSplit)
	ud, err := evalUD(root, p, *udSplit)
	if err != nil {
		return err
	}
	var ag *parser.Agreement
	if len(chapters) > 0 {
		log.Printf("comparing with CoreNLP on %s ...", strings.Join(chapters, ", "))
		if ag, err = parser.CompareWithCoreNLP(root, p, chapters); err != nil {
			return err
		}
	}
	log.Println("timing ...")
	var sents []string
	for _, c := range chapters {
		if cs, err := parser.ChapterSentences(root, c); err == nil {
			sents = append(sents, cs...)
		}
	}
	if len(sents) == 0 {
		if cs, err := parser.ChapterSentences(root, parser.DevBookChapter); err == nil {
			sents = append(sents, cs...)
		}
	}
	mean, p50, p95, p99 := parser.Latency(p, sents, *reps)
	var bookTime time.Duration
	bookSents := 0
	if !*noBook {
		bookTime, bookSents, err = bookParse(root, p)
		if err != nil {
			return err
		}
	}
	var r10 *parser.Rule10Parity
	if len(chapters) > 0 {
		log.Println("running civ/linear Rule 10 parity (both engines) ...")
		if r10, err = parser.Rule10(root, *linear, chapters); err != nil {
			log.Printf("Rule 10 parity skipped: %v", err)
		}
	}
	var draftsAg *parser.Agreement
	nDrafts := 0
	if *drafts != 0 {
		log.Println("comparing with CoreNLP on drafts (sentences absent from the chapters and treatises) ...")
		texts, derr := cachedInputs(root)
		if derr == nil {
			// Sentences that occur in any cached chapter or treatise are left
			// out; a draft that shares a sentence with a held-out chapter
			// (always ch13 and ch14, plus the -chapter list) is sealed whole.
			var reference, sealed []string
			names := make([]string, 0, len(texts))
			for name := range texts {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				reference = append(reference, texts[name])
				if parser.IsHeldout(name) || containsString(chapters, name) {
					sealed = append(sealed, texts[name])
				}
			}
			draftsAg, nDrafts, derr = parser.CompareDrafts(parser.DraftDir(root), p, reference, sealed, *drafts)
		}
		if derr != nil {
			log.Printf("draft comparison skipped: %v", derr)
			draftsAg = nil
		}
	}
	modelBytes := parser.ModelBytes(root)
	bundledBytes := parser.BundledModelBytes(root)

	holdout := strings.Join(chapters, " + ")
	known := official // the Round-1 column is only meaningful for the official run
	r1 := func(v string) string {
		if known {
			return v
		}
		return "—"
	}
	var b strings.Builder
	row := func(name, round1, gateStr, value string, ok bool) {
		mark := "PASS"
		if !ok {
			mark = "FAIL"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", name, round1, gateStr, value, mark)
	}
	pct := func(x float64) string { return fmt.Sprintf("%.2f%%", 100*x) }
	// report prints a figure that has no gate.
	report := func(name, value string) {
		fmt.Fprintf(&b, "| %s | — | report | %s | — |\n", name, value)
	}
	fmt.Fprintf(&b, "\n| Metric | Round 1 | Round-2 Gate | Value | Result |\n|---|---|---|---|---|\n")
	row("UD-EWT "+*udSplit+" POS accuracy", "94.92%", "≥ 96.5%", pct(ud.POS), ud.POS >= 0.965)
	row("UD-EWT "+*udSplit+" UAS (predicted POS)", "83.49%", "≥ 90.0%", pct(ud.UAS), ud.UAS >= 0.900)
	row("UD-EWT "+*udSplit+" LAS (predicted POS)", "80.27%", "≥ 87.5%", pct(ud.LAS), ud.LAS >= 0.875)
	if ag != nil {
		row(holdout+" POS agreement vs CoreNLP", r1("97.19%"), "≥ 97.5%", pct(ag.POS), ag.POS >= 0.975)
		row(holdout+" non-punct UAS vs CoreNLP", r1("~81%"), "≥ 91.0%", pct(ag.UASNonPunct)+" (all-token "+pct(ag.UAS)+")", ag.UASNonPunct >= 0.910)
		row(holdout+" non-punct LAS vs CoreNLP", r1("~77%"), "≥ 88.5%", pct(ag.LASNonPunct)+" (all-token "+pct(ag.LAS)+")", ag.LASNonPunct >= 0.885)
		report(holdout+" non-punct UAS vs CoreNLP, CoreNLP dep left out", pct(ag.UASNonPunctNoDep)+" over "+fmt.Sprint(ag.NonPunctNoDepTokens())+" tokens")
		report(holdout+" non-punct LAS vs CoreNLP, CoreNLP dep left out", pct(ag.LASNonPunctNoDep))
		row(holdout+" Root agreement vs CoreNLP", r1("79.7%"), "≥ 92.0%", pct(ag.Root), ag.Root >= 0.920)
		report(holdout+" Root agreement, copula or auxiliary as head allowed", pct(ag.RootContentHead))
	}
	if r10 != nil {
		row(holdout+" civ/linear Rule 10 correlation (r)", r1("0.6898"), "≥ 0.92", fmt.Sprintf("%.4f", r10.ScoreR), r10.ScoreR >= 0.92)
		row(holdout+" civ/linear Part 4 findings F1", r1("0.462"), "≥ 0.85", fmt.Sprintf("%.3f", r10.Part4F1), r10.Part4F1 >= 0.85)
	}
	row("Mean end-to-end latency (ParseSentence)", "0.620 ms", "< 1.5 ms (stretch < 1.0 ms)", fmt.Sprintf("%.3f ms", mean), mean < 1.5)
	if !*noBook {
		row("Full book (ch01..ch14) parallel parse time", "0.27 s", "< 1.5 s", fmt.Sprintf("%.2f s (%d sentences)", bookTime.Seconds(), bookSents), bookTime < 1500*time.Millisecond)
	}
	row("Cold model load time", "104 ms", "< 150 ms", fmt.Sprintf("%.1f ms", float64(loadTime.Microseconds())/1000), loadTime < 150*time.Millisecond)
	row("Live heap after load", "36.6 MB", "< 45 MB", fmt.Sprintf("%.1f MB", heapMB), heapMB < 45)
	row("Total packed .safetensors.gz size on disk", "7.34 MB", "≤ 15 MB", fmt.Sprintf("%.2f MB", float64(bundledBytes)/1e6), bundledBytes <= 15<<20)
	fmt.Println(b.String())

	fmt.Printf("detail: UD-EWT %s tokens %d; latency p50 %.3f ms, p95 %.3f ms, p99 %.3f ms over %d sentences x %d; %.0f sentences/sec single-thread\n",
		*udSplit, ud.Tokens, p50, p95, p99, len(sents), *reps, 1000/mean)
	if ag != nil {
		if draftsAg != nil {
			fmt.Printf("detail: drafts on %d sentences absent from every cached chapter and treatise (%d drafts): POS %.2f%% UAS %.2f%% LAS %.2f%% | on aligned tokens: POS %.2f%% UAS %.2f%% LAS %.2f%% | NER macro F1 %.2f\n",
				draftsAg.Sentences, nDrafts, 100*draftsAg.POS, 100*draftsAg.UAS, 100*draftsAg.LAS, 100*draftsAg.POSAligned, 100*draftsAg.UASAligned, 100*draftsAg.LASAligned, draftsAg.NERMacroF1)
			for _, l := range parser.NERLabels[1:] {
				c := draftsAg.NER[l]
				fmt.Printf("  draft NER %-7s P %.2f R %.2f (tp %d, gold %d, predicted %d)\n", l, float64(c[0])/float64(max(1, c[2])), float64(c[0])/float64(max(1, c[1])), c[0], c[1], c[2])
			}
		}
		fmt.Printf("detail: CoreNLP parity over %d sentences, %d aligned identically (%.2f%%); on aligned tokens: POS %.2f%% UAS %.2f%% LAS %.2f%%\n",
			ag.Sentences, ag.Aligned, 100*float64(ag.Aligned)/float64(max(1, ag.Sentences)), 100*ag.POSAligned, 100*ag.UASAligned, 100*ag.LASAligned)
		fmt.Printf("detail: non-punct tokens %d: UAS %.2f%% LAS %.2f%%; without CoreNLP dep (%d tokens): UAS %.2f%% LAS %.2f%%; root agreement %.2f%% (%d sentences; %.2f%% with the copula or auxiliary as head)\n", nonPunctTokens(ag), 100*ag.UASNonPunct, 100*ag.LASNonPunct, ag.NonPunctNoDepTokens(), 100*ag.UASNonPunctNoDep, 100*ag.LASNonPunctNoDep, 100*ag.Root, ag.SentenceCount(), 100*ag.RootContentHead)
		labels := make([]string, 0, len(ag.NER))
		for l := range ag.NER {
			labels = append(labels, l)
		}
		sort.Strings(labels)
		var parts []string
		for _, l := range labels {
			c := ag.NER[l]
			parts = append(parts, fmt.Sprintf("%s F1 %.2f (%d gold, %d predicted)", l, f1Of(c[0], c[1], c[2]), c[1], c[2]))
		}
		fmt.Printf("detail: NER vs CoreNLP (aligned sentences): macro F1 %.2f; %s\n", ag.NERMacroF1, strings.Join(parts, "; "))
	}
	if r10 != nil {
		fmt.Printf("detail: Rule 10 parity: %d CoreNLP sentences, %d Go sentences, %d aligned; mean |score difference| %.3f; Part 4 findings F1 %.3f\n",
			r10.CoreNLPSent, r10.GoSent, r10.Aligned, r10.MeanAbsD, r10.Part4F1)
		for _, k := range []string{"distant pronoun", "it chain", "ambiguous it"} {
			c := r10.Kinds[k]
			fmt.Printf("detail:   %-17s CoreNLP %3d  Go %3d  matched %3d  F1 %.3f\n", k+":", c[1], c[2], c[0], f1Of(c[0], c[1], c[2]))
		}
	}
	fmt.Printf("detail: models %.2f MB resolved; %.2f MB packed under civ/sys/parser/models\n", float64(modelBytes)/1e6, float64(bundledBytes)/1e6)
	return nil
}

// evalUD scores the pipeline on a UD_English-EWT split.
func evalUD(root string, p *parser.Pipeline, split string) (parser.UDScores, error) {
	if split == "test" {
		return parser.EvalUDTest(root, p)
	}
	return parser.EvalUDSplit(root, p, split)
}

func nonPunctTokens(ag *parser.Agreement) int { return ag.NonPunctTokens() }

func f1Of(tp, g, p int) float64 {
	if tp == 0 {
		return 0
	}
	return 2 * float64(tp) / float64(g+p)
}

// bookParse parses ch01..ch14 in parallel and returns the wall time.
func bookParse(root string, p *parser.Pipeline) (time.Duration, int, error) {
	type job struct {
		n int
	}
	jobs := make([]job, 14)
	var wg sync.WaitGroup
	errs := make([]error, 14)
	start := time.Now()
	for i := range jobs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b, err := os.ReadFile(filepath.Join(root, "ext-ref", "linear", "in", fmt.Sprintf("ch%02d.txt", i+1)))
			if err != nil {
				errs[i] = err
				return
			}
			jobs[i].n = len(p.ParseDoc(string(b)).Sentences)
		}(i)
	}
	wg.Wait()
	d := time.Since(start)
	total := 0
	for i, err := range errs {
		if err != nil {
			return 0, 0, err
		}
		total += jobs[i].n
	}
	return d, total, nil
}

// ---- parse ---------------------------------------------------------------------

func cmdParse(root string, args []string) error {
	fs := flag.NewFlagSet("parse", flag.ExitOnError)
	n := fs.Int("n", 1, "parse this many times and report the latency")
	fs.Parse(args)
	text := strings.Join(fs.Args(), " ")
	if text == "" {
		return fmt.Errorf("usage: civ/bin/parser parse \"sentence\"")
	}
	t0 := time.Now()
	p, err := parser.Load(root)
	if err != nil {
		return fmt.Errorf("%w (run: civ/bin/parser train-all)", err)
	}
	load := time.Since(t0)
	tree := p.ParseSentence(text)
	var total time.Duration
	for i := 1; i < *n; i++ {
		s := time.Now()
		p.ParseSentence(text)
		total += time.Since(s)
	}
	if tree == nil || tree.N() == 0 {
		fmt.Printf("(no tokens)\n")
		return nil
	}
	fmt.Printf("%s\n\n", tree.Text)
	fmt.Printf("%-4s %-20s %-6s %-18s %-6s %-12s %-8s\n", "idx", "word", "pos", "lemma", "head", "rel", "ner")
	for i := 1; i <= tree.N(); i++ {
		t := tree.Toks[i]
		fmt.Printf("%-4d %-20s %-6s %-18s %-6d %-12s %-8s\n", i, t.Word, t.POS, t.Lemma, tree.Head[i], tree.Rel[i], t.NER)
	}
	if *n > 1 {
		fmt.Printf("\n%.3f ms per sentence over %d runs (model load %s)\n", float64(total.Microseconds())/1000/float64(*n-1), *n-1, load.Round(time.Millisecond))
	} else {
		fmt.Printf("\nmodel load %s\n", load.Round(time.Millisecond))
	}
	return nil
}

// ---- bench ---------------------------------------------------------------------

func cmdBench(root string, args []string) error {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	chapter := fs.String("chapter", "ch13", "chapter whose sentences are timed")
	reps := fs.Int("reps", 3, "repetitions")
	fs.Parse(args)
	t0 := time.Now()
	p, err := parser.Load(root)
	if err != nil {
		return fmt.Errorf("%w (run: civ/bin/parser train-all)", err)
	}
	load := time.Since(t0)
	runtime.KeepAlive(p)
	runtime.GC()
	heapMB, rssMB := parser.Memory()
	sents, err := parser.ChapterSentences(root, *chapter)
	if err != nil {
		return err
	}
	mean, p50, p95, p99 := parser.Latency(p, sents, *reps)
	size := parser.ModelBytes(root)
	fmt.Printf("model load           %s\n", load.Round(time.Millisecond))
	fmt.Printf("models on disk       %.2f MB (resolved), %.2f MB packed\n", float64(size)/1e6, float64(parser.BundledModelBytes(root))/1e6)
	fmt.Printf("live heap after load  %.1f MB (RSS %.1f MB)\n", heapMB, rssMB)
	fmt.Printf("SIMD (AVX2+FMA)       %v; unified model %v\n", vec.SIMDEnabled(), p.Unified())
	fmt.Printf("%s sentences         %d\n", *chapter, len(sents))
	fmt.Printf("ParseSentence        mean %.3f ms  p50 %.3f ms  p95 %.3f ms  p99 %.3f ms  (%d runs)\n", mean, p50, p95, p99, len(sents)**reps)
	fmt.Printf("single-thread speed   %.0f sentences/sec\n", 1000/mean)
	return nil
}

// ---- pack ----------------------------------------------------------------------

// cmdPack compresses the raw weights under ext-ref/parser/ into
// civ/sys/parser/models/: the Round-2 parser and NER when parser.safetensors
// exists (the superseded Round-1 bundles are removed then), else the Round-1
// trio.
func cmdPack(root string, args []string) error {
	fs := flag.NewFlagSet("pack", flag.ExitOnError)
	useF16 := fs.Bool("f16", true, "quantize F32 weight tensors to IEEE-754 F16 before gzip")
	outDir := fs.String("out", "", "destination directory (default civ/sys/parser/models)")
	fs.Parse(args)
	dstDir := *outDir
	if dstDir == "" {
		dstDir = parser.BundledModelDir(root)
	}
	files := []string{parser.POSModelFile, parser.DepModelFile, parser.NERModelFile}
	if _, err := os.Stat(parser.ParserModelPath(root)); err == nil {
		files = []string{parser.ParserModelFile, parser.NERModelFile}
	}
	var rawTotal, packedTotal int64
	for _, file := range files {
		src := filepath.Join(parser.ModelDir(root), file)
		st, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("pack: missing source model %s: %w", src, err)
		}
		dst := filepath.Join(dstDir, file+".gz")
		if err := parser.PackSafetensors(src, dst, *useF16); err != nil {
			return fmt.Errorf("pack %s: %w", file, err)
		}
		dstSt, err := os.Stat(dst)
		if err != nil {
			return err
		}
		rawTotal += st.Size()
		packedTotal += dstSt.Size()
		fmt.Printf("%-26s %6.2f MB -> %5.2f MB (%.1f%%)\n",
			file+".gz", float64(st.Size())/1e6, float64(dstSt.Size())/1e6, 100*float64(dstSt.Size())/float64(st.Size()))
	}
	if len(files) == 2 {
		// The Round-1 tagger and greedy parser are superseded by the
		// unified model; leaving their bundles would count against the size
		// gate and could be picked up by a loader.
		for _, stale := range []string{parser.POSModelFile, parser.DepModelFile} {
			old := filepath.Join(dstDir, stale+".gz")
			if _, err := os.Stat(old); err == nil {
				if err := os.Remove(old); err != nil {
					return err
				}
				log.Printf("removed the superseded Round-1 bundle %s", old)
			}
		}
	}
	fmt.Printf("%-26s %6.2f MB -> %5.2f MB (%.1f%%)\n",
		"total", float64(rawTotal)/1e6, float64(packedTotal)/1e6, 100*float64(packedTotal)/float64(rawTotal))
	return nil
}
