// Command tts converts book chapter markdown into xAI TTS speech-tag text,
// plans REST chunking, and synthesizes audio.
//
//	tts list      — chapter sizes and chunk plan
//	tts annotate  — write annotated *.tts.txt (no API)
//	tts join      — concatenate MP3 chunks without re-encoding (frame-level)
//	tts speak     — synthesize chapters via xAI TTS (key: XAI_API_KEY or .ttskey)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"po/tts/internal/annotate"
	"po/tts/internal/chapters"
	"po/tts/internal/chunk"
	"po/tts/internal/md"
	"po/tts/internal/mp3join"
	"po/tts/internal/xai"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "list":
		err = runList(args)
	case "annotate":
		err = runAnnotate(args)
	case "join":
		err = runJoin(args)
	case "speak":
		err = runSpeak(args)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: tts <command> [flags]

Commands:
  list       List chapters with char counts and REST chunk plan
  annotate   Write structure-pause annotated text (*.tts.txt); no API calls
  join       Concatenate MP3 files without re-encoding (strip tags, append frames)
  speak      Synthesize speech for chapters (xAI TTS)

Auth: XAI_API_KEY env, or a .ttskey file (gitignored) in the repo root / cwd.

Default chapter dir: book1-4/chapters
Default voice: %s  language: %s

`, xai.DefaultVoice, xai.DefaultLanguage)
}

func defaultChapterDir() string {
	candidates := []string{
		"book1-4/chapters",
		"../book1-4/chapters",
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return "book1-4/chapters"
}

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	dir := fs.String("dir", defaultChapterDir(), "chapter markdown directory")
	soft := fs.Int("max-chars", chunk.DefaultSoftMax, "soft max characters per REST chunk")
	if err := fs.Parse(args); err != nil {
		return err
	}
	infos, _, err := chapters.Scan(*dir, *soft)
	if err != nil {
		return err
	}
	fmt.Print(chapters.FormatList(infos, *soft))
	return nil
}

func runAnnotate(args []string) error {
	fs := flag.NewFlagSet("annotate", flag.ContinueOnError)
	dir := fs.String("dir", defaultChapterDir(), "chapter markdown directory")
	outDir := fs.String("o", "out/annotated", "output directory for *.tts.txt")
	soft := fs.Int("max-chars", chunk.DefaultSoftMax, "soft max characters per REST chunk")
	only := fs.String("only", "", "annotate only this filename (e.g. 0201_foundations.md)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	vars, err := md.LoadVars(*dir)
	if err != nil {
		return err
	}
	files, err := md.ListMarkdown(*dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}

	n := 0
	for _, name := range files {
		if *only != "" && name != *only {
			continue
		}
		path := filepath.Join(*dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		blocks := md.ParseChapter(data, vars)
		if md.IsMetadataOnly(blocks) {
			fmt.Printf("skip %s (metadata only)\n", name)
			continue
		}
		ann := annotate.Text(blocks, annotate.DefaultOptions())
		chunks := chunk.Split(ann, *soft, chunk.RESTMaxChars)

		base := strings.TrimSuffix(name, filepath.Ext(name))
		outPath := filepath.Join(*outDir, base+".tts.txt")
		if err := os.WriteFile(outPath, []byte(ann+"\n"), 0o644); err != nil {
			return err
		}

		if len(chunks) > 1 {
			for _, c := range chunks {
				cp := filepath.Join(*outDir, fmt.Sprintf("%s.part%02d.tts.txt", base, c.Index))
				if err := os.WriteFile(cp, []byte(c.Text+"\n"), 0o644); err != nil {
					return err
				}
			}
		}

		fmt.Printf("wrote %s (%d chars, %d chunk(s))\n",
			outPath, utf8.RuneCountInString(ann), len(chunks))
		n++
	}
	if *only != "" && n == 0 {
		return fmt.Errorf("no chapter matched -only %q", *only)
	}
	fmt.Printf("annotated %d chapter(s) → %s\n", n, *outDir)
	return nil
}

func runJoin(args []string) error {
	fs := flag.NewFlagSet("join", flag.ContinueOnError)
	out := fs.String("o", "joined.mp3", "output MP3 path")
	loose := fs.Bool("loose", false, "allow mixed sample-rate/channel profiles")
	if err := fs.Parse(args); err != nil {
		return err
	}
	inputs := fs.Args()
	if len(inputs) < 1 {
		return fmt.Errorf("join: provide one or more input.mp3 files")
	}
	opt := mp3join.Options{RequireSameProfile: !*loose}
	res, err := mp3join.JoinFiles(*out, inputs, opt)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d frames, %d audio bytes, %d inputs, vbr=%v, %s)\n",
		*out, res.Frames, res.Bytes, res.InputFiles, res.VBR, res.Profile)
	return nil
}

func runSpeak(args []string) error {
	fs := flag.NewFlagSet("speak", flag.ContinueOnError)
	dir := fs.String("dir", defaultChapterDir(), "chapter markdown directory")
	outDir := fs.String("o", "out/audio", "output directory for MP3")
	soft := fs.Int("max-chars", chunk.DefaultSoftMax, "soft max characters per REST chunk")
	voice := fs.String("voice", xai.DefaultVoice, "voice_id")
	only := fs.String("only", "", "comma-separated filenames, or empty for all")
	first := fs.Int("first", 0, "only the first N speakable chapters (after skips)")
	keyFile := fs.String("key-file", "", "path to .ttskey (optional; also searches cwd/parents)")
	dryRun := fs.Bool("dry-run", false, "plan only; do not call the API")
	normalize := fs.Bool("normalize", true, "enable text_normalization for numbers/abbrevs")
	if err := fs.Parse(args); err != nil {
		return err
	}

	onlySet := map[string]bool{}
	if *only != "" {
		for _, p := range strings.Split(*only, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				onlySet[p] = true
			}
		}
	}

	vars, err := md.LoadVars(*dir)
	if err != nil {
		return err
	}
	files, err := md.ListMarkdown(*dir)
	if err != nil {
		return err
	}

	type job struct {
		name   string
		path   string
		ann    string
		chunks []chunk.Chunk
	}
	var jobs []job
	speakable := 0
	for _, name := range files {
		path := filepath.Join(*dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		blocks := md.ParseChapter(data, vars)
		if md.IsMetadataOnly(blocks) {
			fmt.Printf("skip %s (metadata only)\n", name)
			continue
		}
		speakable++
		if len(onlySet) > 0 && !onlySet[name] {
			continue
		}
		if *first > 0 && speakable > *first {
			continue
		}
		// When -first is set without -only, take first N by speakable order
		if *first > 0 && len(onlySet) == 0 {
			// speakable already counted; we want only when speakable <= first
			// already handled by continue above when speakable > first
		}
		ann := annotate.Text(blocks, annotate.DefaultOptions())
		chunks := chunk.Split(ann, *soft, chunk.RESTMaxChars)
		jobs = append(jobs, job{name: name, path: path, ann: ann, chunks: chunks})
	}

	// Fix -first logic: when onlySet empty, we included every speakable with speakable<=first
	// Actually when first=2: intro speakable=1 included, foundations speakable=2 included, next speakable=3 continue. Good.

	if len(jobs) == 0 {
		return fmt.Errorf("no chapters selected")
	}

	totalChunks := 0
	for _, j := range jobs {
		totalChunks += len(j.chunks)
		fmt.Printf("plan %s → %d chunk(s), %d annotated chars\n",
			j.name, len(j.chunks), utf8.RuneCountInString(j.ann))
	}
	fmt.Printf("voice=%s language=%s total REST requests=%d out=%s\n",
		*voice, xai.DefaultLanguage, totalChunks, *outDir)

	if *dryRun {
		fmt.Println("dry-run: no API calls")
		return nil
	}

	var keyPaths []string
	if *keyFile != "" {
		keyPaths = append(keyPaths, *keyFile)
	}
	client, err := xai.NewClient(keyPaths...)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}

	ctx := context.Background()
	for _, j := range jobs {
		base := strings.TrimSuffix(j.name, filepath.Ext(j.name))
		var partPaths []string
		var partBytes [][]byte

		for _, ch := range j.chunks {
			fmt.Printf("  TTS %s part %d/%d (%d chars)…\n",
				j.name, ch.Index+1, len(j.chunks), ch.Chars)
			start := time.Now()
			audio, err := client.Synthesize(ctx, xai.Request{
				Text:              ch.Text,
				VoiceID:           *voice,
				Language:          xai.DefaultLanguage,
				TextNormalization: *normalize,
				OutputFormat:      xai.DefaultMP3,
			})
			if err != nil {
				return fmt.Errorf("%s chunk %d: %w", j.name, ch.Index, err)
			}
			partPath := filepath.Join(*outDir, fmt.Sprintf("%s.part%02d.mp3", base, ch.Index))
			if err := os.WriteFile(partPath, audio, 0o644); err != nil {
				return err
			}
			partPaths = append(partPaths, partPath)
			partBytes = append(partBytes, audio)
			fmt.Printf("    wrote %s (%d bytes, %s)\n", partPath, len(audio), time.Since(start).Round(time.Millisecond))
		}

		outMP3 := filepath.Join(*outDir, base+".mp3")
		if len(partBytes) == 1 {
			if err := os.WriteFile(outMP3, partBytes[0], 0o644); err != nil {
				return err
			}
			fmt.Printf("  wrote %s (single chunk)\n", outMP3)
		} else {
			joined, res, err := mp3join.JoinBytes(partBytes, mp3join.Options{RequireSameProfile: true})
			if err != nil {
				// fall back to file join
				res2, err2 := mp3join.JoinFiles(outMP3, partPaths, mp3join.Options{RequireSameProfile: true})
				if err2 != nil {
					return fmt.Errorf("join %s: %v (also %v)", j.name, err, err2)
				}
				fmt.Printf("  wrote %s (%d frames via files)\n", outMP3, res2.Frames)
			} else {
				if err := os.WriteFile(outMP3, joined, 0o644); err != nil {
					return err
				}
				fmt.Printf("  wrote %s (%d frames, %d bytes, vbr=%v)\n",
					outMP3, res.Frames, res.Bytes, res.VBR)
			}
		}
		// Drop intermediate parts once the chapter file is complete.
		for _, p := range partPaths {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "warning: remove %s: %v\n", p, err)
			}
		}
	}
	fmt.Printf("done: %d chapter(s) → %s\n", len(jobs), *outDir)
	return nil
}
