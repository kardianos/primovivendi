// Command compose is the interactive sentence/paragraph/chapter rebuilder for
// civ. It freezes a local snapshot of each chapter, diagrams sentences with
// CoreNLP, explores mechanical syntactic structures and WordNet lexical
// alternatives in a live hover-preview GUI (Lattice / Dependency Diagram /
// Hybrid), runs backward -> forward -> looping seam passes, logs the full
// decision tree, trains a pure-Go inkline-style local NN ranker, and rebuilds
// chapters deterministically on demand.
//
// Sentences that are not in the chapter's cached CoreNLP parses (manual
// rewrites, custom phrases) are parsed in memory by the pure-Go Round-2
// BiGRU + Biaffine pipeline in civ/sys/parser (pre-trained F16+gzip models
// are bundled in civ/sys/parser/models/; optional local retrains live in
// ext-ref/parser/).
//
// Usage:
//
//	civ/bin/compose [ui] [-chapter 01_ch01] [-view lattice|diagram|hybrid]
//	civ/bin/compose fetch-wordnet           # unpack full WordNet 3.1 dict into ext-ref/lexicon/wordnet31/
//	civ/bin/compose snapshot -chapter ch01  # freeze local immutable snapshot
//	civ/bin/compose passes -chapter ch01    # inspect backward goals & paragraph seams
//	civ/bin/compose train                   # train local .safetensors NN ranker from decision trees
//	civ/bin/compose rebuild -chapter ch01   # deterministically rebuild chapter from snapshot + decision tree
//	civ/bin/compose diff -chapter ch01      # show diff between frozen snapshot and current rebuild
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bankai-tensei/gowaku"
	"po/civ/sys/lexicon"
	"po/civ/sys/prof"
)

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func usageText() string {
	return `usage: civ/bin/compose <command> [flags]

Interactive sentence, phrase, and paragraph rebuilder for civ/chapters.
Freezes a local snapshot first, logs a full decision tree (hovers, options,
locks, rewrites, commits), trains a local pure-Go inkline-style NN ranker
(.safetensors), and rebuilds chapters deterministically on demand without
mutating civ/chapters/*.md.

Commands:
  ui             (default) Open interactive gowaku GUI
                   -chapter NAME          chapter to open (e.g. ch01 or 01_ch01)
                   -view MODE             initial graph view: lattice | diagram | hybrid
                   -corenlp-server        use a local StanfordCoreNLPServer for rewrite re-parsing instead
                                          of the pure-Go parser / fallback (opt-in)
                   -profile PATH          periodically write span timing & slow-call profiler log to PATH
                   -profile-interval DUR  how often to flush profiler snapshots (default 5s)
                   -state DIR             override state directory (default civ/notes/compose)
  fetch-wordnet  Download & unpack full Princeton WordNet 3.1 dict into ext-ref/lexicon/wordnet31/
  snapshot       Freeze a local immutable copy of a chapter (-chapter, -force)
  passes         Run backward-pass goals & seam diagnostics on a chapter (-chapter, -para)
  train          Train local pure-Go inkline-style NN ranker (.safetensors) from decision trees
  rebuild        Deterministically rebuild a chapter from its snapshot + decision tree (-chapter, -out)
  diff           Show diff between a chapter's frozen snapshot and its current decision-tree rebuild
  help           Show this message

GUI Keyboard Shortcuts (when not typing in a text field):
  1 / 2 / 3      Switch graph view (1=Lattice, 2=Dependency Diagram, 3=Hybrid)
  B / F / L      Switch pass mode (Backward anchor, Forward solve, Loop seams)
  J / K          Next / Previous sentence (follows active pass direction)
  N / P          Next / Previous paragraph
  Left / Right   Scroll Word Lattice or Dependency Diagram horizontally
  Esc            Clear active multi-word phrase selection
  A / S / R      Accept / Keep As Is / Reject (reset to base)
  Shift+A        Accept entire paragraph (mark settled/completed) & advance to next paragraph
  Ctrl+Enter     Adopt Manual Rewrite text as the new diagrammed base

Quick Start:
  civ/bin/compose fetch-wordnet
  civ/bin/compose ui -chapter ch01 -corenlp-server
  civ/bin/compose train
  civ/bin/compose rebuild -chapter ch01

Full documentation: civ/compose/README.md
`
}

func run(args []string) error {
	cmd := "ui"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	} else if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "-help") {
		fmt.Print(usageText())
		return nil
	}
	switch cmd {
	case "help":
		fmt.Print(usageText())
		return nil
	case "ui":
		return cmdUI(args)
	case "fetch-wordnet":
		return cmdFetchWordNet(args)
	case "snapshot":
		return cmdSnapshot(args)
	case "passes":
		return cmdPasses(args)
	case "train":
		return cmdTrain(args)
	case "rebuild":
		return cmdRebuild(args)
	case "diff":
		return cmdDiff(args)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", cmd, usageText())
	}
}

func findRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .git directory above %s", start)
		}
		dir = parent
	}
}

func cmdUI(args []string) error {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	rootFlag := fs.String("root", "", "repository root")
	stateFlag := fs.String("state", "", "compose state directory (default civ/notes/compose)")
	chapFlag := fs.String("chapter", "", "chapter to open (e.g. ch01 or 01_ch01)")
	viewFlag := fs.String("view", "lattice", "initial graph view: lattice | diagram | hybrid")
	corenlpFlag := fs.Bool("corenlp-server", false, "opt in to the local StanfordCoreNLPServer for re-parsing instead of the pure-Go parser")
	profileFlag := fs.String("profile", "", "write periodic span profiling report to this file path")
	profileInt := fs.Duration("profile-interval", 5*time.Second, "interval between periodic profiler flushes")
	quitAfter := fs.Duration("quit-after", 0, "close the window after this duration (0 waits)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *profileFlag != "" {
		if err := prof.Enable(*profileFlag, *profileInt); err != nil {
			return fmt.Errorf("enable profiler: %w", err)
		}
		defer prof.Stop()
	}
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot(".")
		if err != nil {
			return err
		}
	}
	// Load reference corpus & Google unigrams asynchronously so the GUI window opens immediately.
	eng, err := NewEngine(root, *stateFlag, false)
	if err != nil {
		return err
	}
	go eng.Ngram.LoadRefCorpus(root)
	go eng.Ngram.LoadUnigrams(root)
	if *corenlpFlag {
		// Explicit opt-in: without the flag the engine never contacts a server,
		// even one that happens to be running.
		eng.UseCoreNLPServer = true
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			_ = eng.CoreNLP.Start(ctx)
		}()
		defer eng.CoreNLP.Stop()
	}

	sess, err := NewSession(eng, *chapFlag)
	if err != nil {
		return err
	}
	switch strings.ToLower(*viewFlag) {
	case "diagram":
		sess.View = ViewDiagram
	case "hybrid":
		sess.View = ViewHybrid
	default:
		sess.View = ViewLattice
	}
	if *quitAfter > 0 {
		time.AfterFunc(*quitAfter, gowaku.Quit)
	}
	defer eng.Embed.Save()
	return gowaku.Run(gowaku.App{
		Title:  "Civ Compose",
		Width:  1440,
		Height: 900,
		AppID:  "po.civ.compose",
		Home:   composeApp{Session: sess},
	})
}

func cmdFetchWordNet(args []string) error {
	fs := flag.NewFlagSet("fetch-wordnet", flag.ContinueOnError)
	rootFlag := fs.String("root", "", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot(".")
		if err != nil {
			return err
		}
	}
	dir, err := lexicon.EnsureWordNet31(root)
	if err != nil {
		return err
	}
	fmt.Printf("WordNet 3.1 ready at %s\n", dir)
	return nil
}

func cmdSnapshot(args []string) error {
	fs := flag.NewFlagSet("snapshot", flag.ContinueOnError)
	rootFlag := fs.String("root", "", "repository root")
	stateFlag := fs.String("state", "", "compose state directory")
	chapFlag := fs.String("chapter", "ch01", "chapter to snapshot")
	forceFlag := fs.Bool("force", false, "overwrite existing snapshot")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot(".")
		if err != nil {
			return err
		}
	}
	stateDir := *stateFlag
	if stateDir == "" {
		stateDir = filepath.Join(root, "civ", "notes", "compose")
	}
	snap, path, err := EnsureSnapshot(root, stateDir, *chapFlag, *forceFlag)
	if err != nil {
		return err
	}
	prose, sents := 0, 0
	for _, b := range snap.Blocks {
		if b.Kind == "prose" {
			prose++
			sents += len(b.Sentences)
		}
	}
	fmt.Printf("snapshot %s -> %s (%d prose paragraphs, %d sentences)\n", snap.Chapter, path, prose, sents)
	return nil
}

func cmdPasses(args []string) error {
	fs := flag.NewFlagSet("passes", flag.ContinueOnError)
	rootFlag := fs.String("root", "", "repository root")
	stateFlag := fs.String("state", "", "compose state directory")
	chapFlag := fs.String("chapter", "ch01", "chapter to inspect")
	paraFlag := fs.Int("para", 1, "1-based prose paragraph index")
	profileFlag := fs.String("profile", "", "write periodic span profiling report to this file path")
	profileInt := fs.Duration("profile-interval", 5*time.Second, "interval between periodic profiler flushes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *profileFlag != "" {
		if err := prof.Enable(*profileFlag, *profileInt); err != nil {
			return fmt.Errorf("enable profiler: %w", err)
		}
		defer prof.Stop()
	}
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot(".")
		if err != nil {
			return err
		}
	}
	eng, err := NewEngine(root, *stateFlag, false)
	if err != nil {
		return err
	}
	sess, err := NewSession(eng, *chapFlag)
	if err != nil {
		return err
	}
	sess.JumpParagraph(*paraFlag - 1)
	blk := sess.ActiveBlock()
	if blk == nil {
		return fmt.Errorf("paragraph %d out of range", *paraFlag)
	}
	fmt.Printf("=== %s · Paragraph %d (%s) ===\n", sess.Chapter, *paraFlag, blk.Heading)
	for i, g := range sess.Goals {
		fmt.Printf("  S%d goal: %s\n", i+1, g.Summary)
	}
	fmt.Printf("Seams (%d):\n", len(sess.Seams))
	for _, sm := range sess.Seams {
		fmt.Printf("  - [%s] %s\n", sm.Kind, sm.Message)
	}
	if sess.Plan != nil {
		fmt.Printf("Active S%d structures (%d):\n", sess.SentIdx+1, len(sess.Plan.Structures))
		for _, sc := range sess.Plan.Structures {
			fmt.Printf("  * %-30s score=%.1f lin=%.0f snap=%.0f  %s\n", sc.Label, sc.Combined, sc.Metrics.Score, sc.Cadence.Snap, truncUI(sc.Text, 72))
		}
	}
	return nil
}

func cmdTrain(args []string) error {
	fs := flag.NewFlagSet("train", flag.ContinueOnError)
	rootFlag := fs.String("root", "", "repository root")
	stateFlag := fs.String("state", "", "compose state directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot(".")
		if err != nil {
			return err
		}
	}
	eng, err := NewEngine(root, *stateFlag, false)
	if err != nil {
		return err
	}
	stats, err := eng.TrainFromDecisions()
	if err != nil {
		return err
	}
	fmt.Printf("trained local ranker -> %s (%d pairs, %d epochs, loss %.4f -> %.4f, pair acc %.1f%%)\n",
		eng.ModelPath(), stats.Pairs, stats.Epochs, stats.InitialLoss, stats.FinalLoss, stats.PairAcc*100)
	return nil
}

func cmdRebuild(args []string) error {
	fs := flag.NewFlagSet("rebuild", flag.ContinueOnError)
	rootFlag := fs.String("root", "", "repository root")
	stateFlag := fs.String("state", "", "compose state directory")
	chapFlag := fs.String("chapter", "ch01", "chapter to rebuild")
	outFlag := fs.String("out", "", "output markdown path (default <state>/rebuilt/<chapter>.md)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot(".")
		if err != nil {
			return err
		}
	}
	stateDir := *stateFlag
	if stateDir == "" {
		stateDir = filepath.Join(root, "civ", "notes", "compose")
	}
	snap, _, err := EnsureSnapshot(root, stateDir, *chapFlag, false)
	if err != nil {
		return err
	}
	ds, err := OpenDecisionStore(stateDir, snap.Chapter, snap)
	if err != nil {
		return err
	}
	outPath := *outFlag
	if outPath == "" {
		outPath = filepath.Join(stateDir, "rebuilt", snap.Chapter+".md")
	}
	res, err := RebuildChapter(snap, ds, outPath)
	if err != nil {
		return err
	}
	fmt.Printf("rebuilt %s -> %s (%d/%d changed, %d accepted)\n", res.Chapter, res.OutPath, res.ChangedSents, res.TotalSents, res.AcceptedSents)
	return nil
}

func cmdDiff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	rootFlag := fs.String("root", "", "repository root")
	stateFlag := fs.String("state", "", "compose state directory")
	chapFlag := fs.String("chapter", "ch01", "chapter to diff")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot(".")
		if err != nil {
			return err
		}
	}
	stateDir := *stateFlag
	if stateDir == "" {
		stateDir = filepath.Join(root, "civ", "notes", "compose")
	}
	snap, _, err := EnsureSnapshot(root, stateDir, *chapFlag, false)
	if err != nil {
		return err
	}
	ds, err := OpenDecisionStore(stateDir, snap.Chapter, snap)
	if err != nil {
		return err
	}
	fmt.Print(DiffSummary(snap, ds))
	return nil
}
