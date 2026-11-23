// Command review is the manuscript review helper, split into subcommands by
// role and action:
//
//	Author (interactive, read-only on chapters):
//	  go run . [ui]         open the review window (writes decisions.json only)
//	  go run . tui          same review session in the terminal
//
//	Inspect (author or agent, read-only):
//	  go run . check        validate packets and locate every quote (-json)
//
//	Mechanical apply (edits chapters deterministically; never runs build.fish):
//	  go run . apply        apply clean verbatim accept/revise decisions (-dry-run)
//
//	Agent handoff (read-only export):
//	  go run . export       write batch.md or JSON for remaining items (-format md|json)
//
//	Agent packet writer:
//	  go run . add          validate, pin offsets, and write a timestamped packet
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bankai-tensei/gowaku"
)

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

type commonFlags struct {
	root      *string
	packets   *string
	decisions *string
}

func bindCommon(fs *flag.FlagSet) commonFlags {
	return commonFlags{
		root:      fs.String("root", "", "repository root (default: walk up to .git)"),
		packets:   fs.String("packets", "", "directory of packet JSON files"),
		decisions: fs.String("decisions", "", "decisions.json path"),
	}
}

func (c commonFlags) resolve() (root, packets, decisions string, err error) {
	root = *c.root
	if root == "" {
		root, err = findRoot(".")
		if err != nil {
			return "", "", "", err
		}
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return "", "", "", err
	}
	packets = *c.packets
	if packets == "" {
		packets = filepath.Join(root, "civ", "notes", "review", "packets")
	}
	decisions = *c.decisions
	if decisions == "" {
		decisions = filepath.Join(root, "civ", "notes", "review", "decisions.json")
	}
	return root, packets, decisions, nil
}

func usageText() string {
	return `usage: civ/bin/review <command> [flags]

Commands (role / action):
  ui      (default) Author viewer (read-only on chapters; writes decisions.json)
  tui     Same viewer in the terminal
  check   Validate packets & locate quotes (-json)
  apply   Apply clean verbatim accept/revise edits (-dry-run, -include-noted, -allow-locked, -id, -json)
  export  Export unapplied items, holds, and closed items (-format md|json, -out PATH|-)
  add     Validate packet/response JSON, pin offsets, write timestamped file (-in PATH|-, -slug NAME, -pin, -strict)
  help    Show this message

Agent quick ref: civ/notes/31_review_queue.md. Run "civ/bin/review <command> -h" for flags.
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
		return runUI(args)
	case "tui":
		return runTUI(args)
	case "check":
		return runCheck(args)
	case "apply":
		return runApply(args)
	case "export":
		return runExport(args)
	case "add":
		return runAdd(args)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", cmd, usageText())
	}
}

func runUI(args []string) error {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	common := bindCommon(fs)
	quitAfter := fs.Duration("quit-after", 0, "close the window after this long; 0 waits")
	titleFlag := fs.String("title", "Review", "window title")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, packets, decisions, err := common.resolve()
	if err != nil {
		return err
	}
	session, err := LoadSession(root, packets, decisions)
	if err != nil {
		return err
	}
	if *quitAfter > 0 {
		time.AfterFunc(*quitAfter, gowaku.Quit)
	}
	return gowaku.Run(gowaku.App{
		Title:  *titleFlag,
		Width:  1320,
		Height: 840,
		AppID:  "po.civ.review",
		Home:   reviewApp{Session: session},
	})
}

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	common := bindCommon(fs)
	asJSON := fs.Bool("json", false, "emit machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, packets, decisions, err := common.resolve()
	if err != nil {
		return err
	}
	session, err := LoadSession(root, packets, decisions)
	if err != nil {
		return err
	}
	if *asJSON {
		b, err := session.CheckJSON()
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		return err
	}
	fmt.Print(session.CheckReport())
	return nil
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	common := bindCommon(fs)
	dryRun := fs.Bool("dry-run", false, "show what would be applied without editing files")
	includeNoted := fs.Bool("include-noted", false, "also apply accept/revise items that carry a note")
	allowLocked := fs.Bool("allow-locked", false, "allow applying edits to items marked locked")
	idFlag := fs.String("id", "", "apply only this item id (or comma-separated ids)")
	asJSON := fs.Bool("json", false, "emit machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, packets, decisions, err := common.resolve()
	if err != nil {
		return err
	}
	session, err := LoadSession(root, packets, decisions)
	if err != nil {
		return err
	}
	opts := ApplyOptions{
		DryRun:       *dryRun,
		IncludeNoted: *includeNoted,
		AllowLocked:  *allowLocked,
	}
	if strings.TrimSpace(*idFlag) != "" {
		opts.OnlyIDs = map[string]bool{}
		for _, part := range strings.Split(*idFlag, ",") {
			if id := strings.TrimSpace(part); id != "" {
				opts.OnlyIDs[id] = true
			}
		}
	}
	res, err := session.RunApply(opts)
	if err != nil {
		return err
	}
	if *asJSON {
		b, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(append(b, '\n'))
		return err
	}
	fmt.Print(FormatApplyResult(res))
	return nil
}

func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	common := bindCommon(fs)
	format := fs.String("format", "md", "output format: md or json")
	outFlag := fs.String("out", "", "output file path, or - for stdout (default: batch.md or batch.json)")
	batchFlag := fs.String("batch", "", "alias for -out")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, packets, decisions, err := common.resolve()
	if err != nil {
		return err
	}
	session, err := LoadSession(root, packets, decisions)
	if err != nil {
		return err
	}
	var payload []byte
	defaultName := "batch.md"
	switch strings.ToLower(*format) {
	case "md", "markdown":
		payload = []byte(session.BatchMarkdown())
	case "json":
		defaultName = "batch.json"
		payload, err = session.ExportJSON()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown format %q (want md or json)", *format)
	}
	dest := *outFlag
	if dest == "" {
		dest = *batchFlag
	}
	if dest == "-" {
		_, err = os.Stdout.Write(payload)
		return err
	}
	if dest == "" {
		dest = filepath.Join(root, "civ", "notes", "review", defaultName)
	}
	if err := writeFileAtomic(dest, payload, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", dest)
	return nil
}

func runAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	common := bindCommon(fs)
	inFlag := fs.String("in", "-", "input packet JSON file, or - for stdin")
	slugFlag := fs.String("slug", "packet", "short name for the timestamped packet file")
	pinFlag := fs.Bool("pin", true, "fill byte_offset and byte_end when a quote matches uniquely")
	strictFlag := fs.Bool("strict", false, "fail if any item is missing or ambiguous")
	asJSON := fs.Bool("json", false, "emit machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, packets, _, err := common.resolve()
	if err != nil {
		return err
	}
	var raw []byte
	if *inFlag == "-" || *inFlag == "" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(*inFlag)
	}
	if err != nil {
		return err
	}
	res, err := AddPacket(root, packets, raw, AddOptions{
		Slug:   *slugFlag,
		Pin:    *pinFlag,
		Strict: *strictFlag,
	})
	if err != nil {
		return err
	}
	if *asJSON {
		b, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(append(b, '\n'))
		return err
	}
	fmt.Print(FormatAddResult(res))
	return nil
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
			return "", fmt.Errorf("no .git above %s", start)
		}
		dir = parent
	}
}
