// Command parser trains, evaluates, and runs the pure-Go NLP pipeline that
// replaces Stanford CoreNLP for civ/compose and civ/linear (see
// civ/notes/47_go_parser_spec.md). Models and treebanks live uncommitted under
// ext-ref/parser/.
//
// Usage, from civ/:
//
//	civ/bin/parser fetch-ud                             # download the gold treebanks
//	civ/bin/parser train-pos [-distill] [-epochs 12]    # train & save pos_tagger.safetensors
//	civ/bin/parser train-dep [-distill] [-epochs 20]    # train & save dep_parser.safetensors
//	civ/bin/parser train-ner                            # train & save ner.safetensors
//	civ/bin/parser train-all [-distill] [-epochs 20]    # run the full training pipeline
//	civ/bin/parser eval [-chapter ch13,ch14]            # UD test + held-out parity + Rule 10
//	civ/bin/parser parse "We come into this world naked."
//	civ/bin/parser bench [-chapter ch13] [-reps 3]      # speed and memory
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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
	case "train-pos":
		err2 = cmdTrainPOS(root, args)
	case "train-dep":
		err2 = cmdTrainDep(root, args)
	case "train-ner":
		err2 = cmdTrainNER(root, args)
	case "train-all":
		err2 = cmdTrainAll(root, args)
	case "eval":
		err2 = cmdEval(root, args)
	case "parse":
		err2 = cmdParse(root, args)
	case "bench":
		err2 = cmdBench(root, args)
	case "pack":
		err2 = cmdPack(root, args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
	if err2 != nil {
		log.Fatal(err2)
	}
}

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

func usage() {
	fmt.Fprint(os.Stderr, `usage: civ/bin/parser <command> [flags]

  fetch-ud                             download UD_English-EWT (+ UD_English-GUM) into ext-ref/parser/
  train-pos [-distill] [-epochs N]     train & save pos_tagger.safetensors
  train-dep [-distill] [-epochs N]     train & save dep_parser.safetensors
  train-ner                            train & save ner.safetensors
  train-all [-distill] [-epochs N]     fetch check, then train POS, dependencies, and NER
  eval [-chapter ch13,ch14]            UD test + held-out CoreNLP parity + Rule 10 parity table
  parse [-n 1] "sentence"              print the dependency tree and latency
  bench [-chapter ch13] [-reps 3]      sentences/sec, latency percentiles, memory
  pack [-f16=true] [-out DIR]          compress ext-ref/parser/*.safetensors into civ/sys/parser/models/*.safetensors.gz

-distill mixes the CoreNLP silver prose (treatises and ch01..ch12) into training.
ch13 and ch14 are held out: they are only ever read by eval.
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

func cmdFetchUD(root string, args []string) error {
	fs := flag.NewFlagSet("fetch-ud", flag.ExitOnError)
	force := fs.Bool("force", false, "re-download files that already exist")
	fs.Parse(args)
	dirOf := map[string]string{
		"UD_English-EWT": filepath.Join(root, "ext-ref", "parser", "ud_ewt"),
		"UD_English-GUM": filepath.Join(root, "ext-ref", "parser", "ud_gum"),
	}
	for _, tb := range treebanks {
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
				return fmt.Errorf("%s: %w (license: %s)", url, err, tb.name)
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
	req.Header.Set("User-Agent", "civ-parser/1.0")
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

func trainFlags(name string, args []string) (distill bool, epochs int) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.BoolVar(&distill, "distill", false, "mix the CoreNLP silver prose into training")
	fs.IntVar(&epochs, "epochs", 0, "training epochs (0 for the default)")
	fs.Parse(args)
	return distill, epochs
}

func cmdTrainPOS(root string, args []string) error {
	distill, epochs := trainFlags("train-pos", args)
	if err := requireUD(root); err != nil {
		return err
	}
	_, err := parser.TrainAndSavePOS(root, distill, epochs, logf)
	return err
}

func cmdTrainDep(root string, args []string) error {
	distill, epochs := trainFlags("train-dep", args)
	if err := requireUD(root); err != nil {
		return err
	}
	if !hasPOS(root) {
		return fmt.Errorf("train-dep needs the POS tagger (run: civ/bin/parser train-pos)")
	}
	_, err := parser.TrainAndSaveDep(root, distill, epochs, logf)
	return err
}

func cmdTrainNER(root string, args []string) error {
	flag.NewFlagSet("train-ner", flag.ExitOnError).Parse(args)
	if !hasPOS(root) {
		return fmt.Errorf("train-ner needs the POS tagger (run: civ/bin/parser train-pos)")
	}
	_, err := parser.TrainAndSaveNER(root, logf)
	return err
}

func cmdTrainAll(root string, args []string) error {
	distill, epochs := trainFlags("train-all", args)
	if err := cmdFetchUD(root, []string{"-force=false"}); err != nil {
		return err
	}
	if _, err := parser.TrainAndSavePOS(root, distill, 0, logf); err != nil {
		return err
	}
	if _, err := parser.TrainAndSaveDep(root, distill, epochs, logf); err != nil {
		return err
	}
	_, err := parser.TrainAndSaveNER(root, logf)
	return err
}

// hasPOS reports whether the POS tagger has been trained.
func hasPOS(root string) bool {
	pos, _, _ := parser.ModelsPresent(root)
	return pos
}

func requireUD(root string) error {
	if !parser.EWTAvailable(root) {
		return fmt.Errorf("UD_English-EWT is not downloaded (run: civ/bin/parser fetch-ud)")
	}
	return nil
}

// ---- eval ---------------------------------------------------------------------

func cmdEval(root string, args []string) error {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	chapter := fs.String("chapter", "ch13,ch14", "held-out CoreNLP chapters (comma separated)")
	linear := fs.String("linear", "", "path to the civ/linear binary (default civ/bin/linear)")
	reps := fs.Int("reps", 5, "timing repetitions per sentence")
	drafts := fs.Int("drafts", 0, "also compare with CoreNLP on drafts, on sentences absent from the chapters and treatises (up to N drafts; 0 skips)")
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
	t0 := time.Now()
	p, err := parser.Load(root)
	if err != nil {
		return fmt.Errorf("%w (run: civ/bin/parser train-all -distill)", err)
	}
	loadTime := time.Since(t0)
	heapMB, rssMB := parser.Memory()
	log.Printf("models loaded in %s (heap %.1f MB, RSS %.1f MB, simd=%v)", loadTime.Round(time.Millisecond), heapMB, rssMB, vec.SIMDEnabled())

	log.Println("evaluating on UD_English-EWT test ...")
	ud, err := parser.EvalUDTest(root, p)
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
		if cs, err := parser.ChapterSentences(root, "ch12"); err == nil {
			sents = cs
		}
	}
	mean, p50, p95, p99 := parser.Latency(p, sents, *reps)
	bookTime, bookSents, err := bookParse(root, p)
	if err != nil {
		return err
	}
	var r10 *parser.Rule10Parity
	if len(chapters) > 0 {
		log.Println("running civ/linear Rule 10 parity (both engines) ...")
		if r10, err = parser.Rule10(root, *linear, chapters); err != nil {
			log.Printf("Rule 10 parity skipped: %v", err)
		}
	}
	var draftsAg *parser.Agreement
	var nDrafts int
	if *drafts != 0 {
		log.Println("comparing with CoreNLP on drafts (sentences absent from the chapters and treatises) ...")
		texts, derr := cachedInputs(root)
		if derr == nil {
			// Trained sources supply the sentences to leave out; the held-out
			// chapters seal the drafts that share a sentence with them.
			var reference, sealed []string
			for _, name := range slices.Sorted(maps.Keys(texts)) {
				if slices.Contains(chapters, name) {
					sealed = append(sealed, texts[name])
				} else {
					reference = append(reference, texts[name])
				}
			}
			limit := *drafts
			if limit < 0 {
				limit = 0 // all
			}
			draftsAg, nDrafts, derr = parser.CompareDrafts(filepath.Join(root, "ext-ref", "linear", "scratch"), p, reference, sealed, limit)
		}
		if derr != nil {
			log.Printf("draft comparison skipped: %v", derr)
			draftsAg = nil
		}
	}
	modelBytes := parser.ModelBytes(root)

	var b strings.Builder
	row := func(name, value string, gate string, ok bool) {
		mark := "PASS"
		if !ok {
			mark = "FAIL"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", name, value, gate, mark)
	}
	pct := func(x float64) string { return fmt.Sprintf("%.2f%%", 100*x) }
	fmt.Fprintf(&b, "\n| Metric | Value | Gate | Result |\n|---|---|---|---|\n")
	row("UD-EWT test POS accuracy", pct(ud.POS), "≥ 96.5%", ud.POS >= 0.965)
	row("UD-EWT test UAS (predicted POS)", pct(ud.UAS), "≥ 90.0%", ud.UAS >= 0.900)
	row("UD-EWT test LAS (predicted POS)", pct(ud.LAS), "≥ 87.5%", ud.LAS >= 0.875)
	if ag != nil {
		name := strings.Join(ag.Chapters, " + ")
		row(name+" POS agreement vs CoreNLP", pct(ag.POS), "≥ 97.0%", ag.POS >= 0.970)
		row(name+" UAS vs CoreNLP", pct(ag.UAS), "≥ 91.0%", ag.UAS >= 0.910)
		row(name+" LAS vs CoreNLP", pct(ag.LAS), "≥ 88.5%", ag.LAS >= 0.885)
	}
	if r10 != nil {
		row(agName(r10)+" civ/linear Rule 10 score correlation (r)", fmt.Sprintf("%.4f", r10.ScoreR), "≥ 0.92", r10.ScoreR >= 0.92)
	}
	row("Mean end-to-end latency (ParseSentence)", fmt.Sprintf("%.3f ms", mean), "< 1.0 ms", mean < 1.0)
	row("Full book (ch01..ch14) parallel parse time", fmt.Sprintf("%.2f s (%d sentences)", bookTime.Seconds(), bookSents), "< 2.5 s", bookTime < 2500*time.Millisecond)
	row("Cold model load time", fmt.Sprintf("%.1f ms", float64(loadTime.Microseconds())/1000), "< 150 ms", loadTime < 150*time.Millisecond)
	row("Total .safetensors size on disk", fmt.Sprintf("%.2f MB", float64(modelBytes)/1e6), "≤ 20 MB", modelBytes <= 20<<20)
	fmt.Println(b.String())

	fmt.Printf("detail: UD-EWT test tokens %d; latency p50 %.3f ms, p95 %.3f ms, p99 %.3f ms over %d sentences x %d; %.0f sentences/sec single-thread\n",
		ud.Tokens, p50, p95, p99, len(sents), *reps, 1000/mean)
	if ag != nil {
		if draftsAg != nil {
			fmt.Printf("detail: drafts on %d sentences absent from the chapters and treatises (%d drafts): POS %.2f%% UAS %.2f%% LAS %.2f%% | on aligned tokens: POS %.2f%% UAS %.2f%% LAS %.2f%% | NER macro F1 %.2f\n",
				draftsAg.Sentences, nDrafts, 100*draftsAg.POS, 100*draftsAg.UAS, 100*draftsAg.LAS, 100*draftsAg.POSAligned, 100*draftsAg.UASAligned, 100*draftsAg.LASAligned, draftsAg.NERMacroF1)
			for _, l := range parser.NERLabels[1:] {
				c := draftsAg.NER[l]
				fmt.Printf("  draft NER %-7s P %.2f R %.2f (tp %d, gold %d, predicted %d)\n", l, float64(c[0])/float64(max(1, c[2])), float64(c[0])/float64(max(1, c[1])), c[0], c[1], c[2])
			}
		}
		fmt.Printf("detail: CoreNLP parity over %d sentences, %d aligned identically (%.2f%%); on aligned tokens: POS %.2f%% UAS %.2f%% LAS %.2f%%\n",
			ag.Sentences, ag.Aligned, 100*float64(ag.Aligned)/float64(max(1, ag.Sentences)), 100*ag.POSAligned, 100*ag.UASAligned, 100*ag.LASAligned)
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
	return nil
}

func agName(r *parser.Rule10Parity) string { return strings.Join(r.Chapters, " + ") }

func f1Of(tp, g, p int) float64 {
	if tp == 0 {
		return 0
	}
	return 2 * float64(tp) / float64(g+p)
}

// bookParse parses ch01..ch14 in parallel and returns the wall time.
func bookParse(root string, p *parser.Pipeline) (time.Duration, int, error) {
	type job struct {
		text string
		n    int
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
		return fmt.Errorf("%w (run: civ/bin/parser train-all -distill)", err)
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
		return fmt.Errorf("%w (run: civ/bin/parser train-all -distill)", err)
	}
	load := time.Since(t0)
	runtime.GC()
	heapMB, rssMB := parser.Memory()
	sents, err := parser.ChapterSentences(root, *chapter)
	if err != nil {
		return err
	}
	mean, p50, p95, p99 := parser.Latency(p, sents, *reps)
	bookTime, bookSents, err := bookParse(root, p)
	if err != nil {
		return err
	}
	size := parser.ModelBytes(root)
	fmt.Printf("model load           %s\n", load.Round(time.Millisecond))
	fmt.Printf("models on disk        %.2f MB\n", float64(size)/1e6)
	fmt.Printf("live heap after load  %.1f MB (RSS %.1f MB)\n", heapMB, rssMB)
	fmt.Printf("SIMD (AVX2+FMA)       %v\n", vec.SIMDEnabled())
	fmt.Printf("%s sentences         %d\n", *chapter, len(sents))
	fmt.Printf("ParseSentence        mean %.3f ms  p50 %.3f ms  p95 %.3f ms  p99 %.3f ms  (%d runs)\n", mean, p50, p95, p99, len(sents)**reps)
	fmt.Printf("single-thread speed   %.0f sentences/sec\n", 1000/mean)
	fmt.Printf("full book (ch01..ch14) %d sentences parsed in parallel: %.2f s (%.0f sentences/sec)\n", bookSents, bookTime.Seconds(), float64(bookSents)/bookTime.Seconds())
	return nil
}

// ---- pack ----------------------------------------------------------------------

func cmdPack(root string, args []string) error {
	fs := flag.NewFlagSet("pack", flag.ExitOnError)
	useF16 := fs.Bool("f16", true, "quantize F32 weight tensors to IEEE-754 F16 before gzip")
	outDir := fs.String("out", "", "destination directory (default civ/sys/parser/models)")
	fs.Parse(args)
	dstDir := *outDir
	if dstDir == "" {
		dstDir = parser.BundledModelDir(root)
	}
	var rawTotal, packedTotal int64
	for _, file := range []string{parser.POSModelFile, parser.DepModelFile, parser.NERModelFile} {
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
	fmt.Printf("%-26s %6.2f MB -> %5.2f MB (%.1f%%)\n",
		"total", float64(rawTotal)/1e6, float64(packedTotal)/1e6, 100*float64(packedTotal)/float64(rawTotal))
	return nil
}
