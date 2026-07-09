package chapters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"po/tts/internal/annotate"
	"po/tts/internal/chunk"
	"po/tts/internal/md"
)

// Info describes one chapter for listing and planning TTS requests.
type Info struct {
	File       string
	Path       string
	Skip       bool // metadata-only (e.g. title block)
	Blocks     int
	SpokenRaw  int // plain spoken text length (no tags)
	Annotated  int // annotated text length
	Chunks     int
	OverREST   bool // needs more than one chunk or exceeds REST limit
	ChunkSizes []int
}

// Scan loads vars from the chapter dir and builds Info for each markdown file.
func Scan(dir string, softMax int) ([]Info, map[string]string, error) {
	if softMax <= 0 {
		softMax = chunk.DefaultSoftMax
	}
	vars, err := md.LoadVars(dir)
	if err != nil {
		return nil, nil, err
	}
	files, err := md.ListMarkdown(dir)
	if err != nil {
		return nil, nil, err
	}

	var out []Info
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", path, err)
		}
		blocks := md.ParseChapter(data, vars)
		info := Info{
			File:   name,
			Path:   path,
			Blocks: len(blocks),
		}
		if md.IsMetadataOnly(blocks) {
			info.Skip = true
			out = append(out, info)
			continue
		}
		info.SpokenRaw = md.SpokenLen(blocks)
		ann := annotate.Text(blocks, annotate.DefaultOptions())
		info.Annotated = utf8.RuneCountInString(ann)
		chunks := chunk.Split(ann, softMax, chunk.RESTMaxChars)
		info.Chunks = len(chunks)
		for _, c := range chunks {
			info.ChunkSizes = append(info.ChunkSizes, c.Chars)
			if c.Chars > chunk.RESTMaxChars {
				info.OverREST = true
			}
		}
		if info.Chunks > 1 || info.Annotated > chunk.RESTMaxChars {
			info.OverREST = true
		}
		out = append(out, info)
	}
	return out, vars, nil
}

// FormatList returns a human-readable chapter listing with size and chunk plan.
func FormatList(infos []Info, softMax int) string {
	if softMax <= 0 {
		softMax = chunk.DefaultSoftMax
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("REST limit: %d chars  soft max: %d chars\n", chunk.RESTMaxChars, softMax))
	b.WriteString(fmt.Sprintf("%-28s %6s %8s %8s %6s %s\n",
		"FILE", "SKIP", "RAW", "ANNOT", "CHUNKS", "NOTES"))
	totalRaw, totalAnn, totalChunks := 0, 0, 0
	for _, in := range infos {
		skip := ""
		if in.Skip {
			skip = "yes"
		}
		note := ""
		if in.Skip {
			note = "metadata only"
		} else if in.Chunks > 1 {
			note = fmt.Sprintf("split %v", in.ChunkSizes)
		} else if in.Annotated > softMax {
			note = "near soft max"
		}
		b.WriteString(fmt.Sprintf("%-28s %6s %8d %8d %6d %s\n",
			in.File, skip, in.SpokenRaw, in.Annotated, in.Chunks, note))
		if !in.Skip {
			totalRaw += in.SpokenRaw
			totalAnn += in.Annotated
			totalChunks += in.Chunks
		}
	}
	b.WriteString(fmt.Sprintf("%-28s %6s %8d %8d %6d\n",
		"TOTAL (speakable)", "", totalRaw, totalAnn, totalChunks))
	return b.String()
}
