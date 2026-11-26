package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ChapterSnapshot is a frozen, immutable local copy of a chapter markdown file
// captured before rebuilding begins. All decision-tree records reference this
// snapshot so edits to civ/chapters/ never invalidate offsets or training data.
type ChapterSnapshot struct {
	Version    int        `json:"version"`
	Chapter    string     `json:"chapter"`     // e.g. "01_ch01"
	SourcePath string     `json:"source_path"` // relative to repo root
	SourceHash string     `json:"source_hash"`
	CreatedAt  time.Time  `json:"created_at"`
	Blocks     []DocBlock `json:"blocks"`
}

// DocBlock is one markdown block in the chapter (heading, prose paragraph, or raw block).
type DocBlock struct {
	ID        string         `json:"id"`   // e.g. "b001"
	Kind      string         `json:"kind"` // "heading", "prose", "raw"
	Heading   string         `json:"heading,omitempty"`
	Raw       string         `json:"raw"`
	Separator string         `json:"separator"` // trailing blank lines/newlines to preserve markdown formatting
	Sentences []SnapSentence `json:"sentences,omitempty"`
}

// SnapSentence is one sentence inside a prose DocBlock.
type SnapSentence struct {
	ID      string `json:"id"` // e.g. "01_ch01:b004:s02"
	BlockID string `json:"block_id"`
	Index   int    `json:"index"` // 0-based within paragraph
	Text    string `json:"text"`
	Hash    string `json:"hash"`
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:12])
}

// ListChapters returns all markdown chapter files under civ/chapters/ in order.
func ListChapters(root string) ([]string, error) {
	paths, err := filepath.Glob(filepath.Join(root, "civ", "chapters", "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, strings.TrimSuffix(filepath.Base(p), ".md"))
	}
	return out, nil
}

// ResolveChapterPath turns "ch01", "01_ch01", or "01_ch01.md" into an absolute path.
func ResolveChapterPath(root, chapter string) (string, string, error) {
	chapDir := filepath.Join(root, "civ", "chapters")
	clean := strings.TrimSuffix(filepath.Base(chapter), ".md")
	exact := filepath.Join(chapDir, clean+".md")
	if _, err := os.Stat(exact); err == nil {
		return exact, clean, nil
	}
	paths, err := filepath.Glob(filepath.Join(chapDir, "*.md"))
	if err != nil {
		return "", "", err
	}
	sort.Strings(paths)
	for _, p := range paths {
		base := strings.TrimSuffix(filepath.Base(p), ".md")
		if base == clean || strings.HasSuffix(base, "_"+clean) {
			return p, base, nil
		}
	}
	return "", "", fmt.Errorf("chapter %q not found in %s", chapter, chapDir)
}

// EnsureSnapshot loads an existing local snapshot or freezes a fresh copy of
// the chapter from civ/chapters/<chapter>.md.
func EnsureSnapshot(root, stateDir, chapter string, force bool) (*ChapterSnapshot, string, error) {
	srcAbs, chapName, err := ResolveChapterPath(root, chapter)
	if err != nil {
		return nil, "", err
	}
	snapPath := filepath.Join(stateDir, "snapshots", chapName+".json")
	if !force {
		if b, err := os.ReadFile(snapPath); err == nil {
			var snap ChapterSnapshot
			if err := json.Unmarshal(b, &snap); err == nil && len(snap.Blocks) > 0 {
				return &snap, snapPath, nil
			}
		}
	}
	raw, err := os.ReadFile(srcAbs)
	if err != nil {
		return nil, "", err
	}
	relPath, _ := filepath.Rel(root, srcAbs)
	snap := BuildSnapshot(chapName, filepath.ToSlash(relPath), string(raw))
	if err := SaveSnapshot(snapPath, snap); err != nil {
		return nil, "", err
	}
	return snap, snapPath, nil
}

// SaveSnapshot atomically writes a ChapterSnapshot JSON file.
func SaveSnapshot(path string, snap *ChapterSnapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// BuildSnapshot splits a chapter Markdown document into blocks and sentences
// while preserving exact inter-block separators so deterministic rebuilds match
// the original Markdown structure byte-for-byte when untouched.
func BuildSnapshot(chapter, sourcePath, markdown string) *ChapterSnapshot {
	snap := &ChapterSnapshot{
		Version:    1,
		Chapter:    chapter,
		SourcePath: sourcePath,
		SourceHash: sha256Hex(markdown),
		CreatedAt:  time.Now().UTC(),
	}
	text := strings.ReplaceAll(markdown, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	curHeading := ""
	blockIdx := 0
	i := 0
	for i < len(lines) {
		// Collect leading blank lines if at start.
		if strings.TrimSpace(lines[i]) == "" {
			if len(snap.Blocks) > 0 {
				snap.Blocks[len(snap.Blocks)-1].Separator += "\n"
			} else {
				blockIdx++
				snap.Blocks = append(snap.Blocks, DocBlock{
					ID:        fmt.Sprintf("b%03d", blockIdx),
					Kind:      "raw",
					Raw:       "",
					Separator: "\n",
				})
			}
			i++
			continue
		}

		start := i
		inFence := strings.HasPrefix(strings.TrimSpace(lines[i]), "```")
		if inFence {
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				i++
			}
			if i < len(lines) {
				i++
			}
		} else if strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			i++
		} else {
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				if i > start && strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
					break
				}
				i++
			}
		}
		rawBlock := strings.Join(lines[start:i], "\n")
		sep := ""
		if i < len(lines) {
			sep = "\n"
		}
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			sep += "\n"
			i++
		}

		blockIdx++
		bid := fmt.Sprintf("b%03d", blockIdx)
		trim := strings.TrimSpace(rawBlock)
		kind := classifyBlock(trim)
		if kind == "heading" {
			curHeading = strings.TrimSpace(strings.TrimLeft(trim, "#"))
		}
		blk := DocBlock{
			ID:        bid,
			Kind:      kind,
			Heading:   curHeading,
			Raw:       rawBlock,
			Separator: sep,
		}
		if kind == "prose" {
			sents := SplitSentences(rawBlock)
			for si, st := range sents {
				blk.Sentences = append(blk.Sentences, SnapSentence{
					ID:      fmt.Sprintf("%s:%s:s%02d", chapter, bid, si),
					BlockID: bid,
					Index:   si,
					Text:    st,
					Hash:    sha256Hex(st),
				})
			}
		}
		snap.Blocks = append(snap.Blocks, blk)
	}
	return snap
}

func classifyBlock(trim string) string {
	switch {
	case trim == "":
		return "raw"
	case strings.HasPrefix(trim, "#"):
		return "heading"
	case strings.HasPrefix(trim, "```"),
		strings.HasPrefix(trim, ">"),
		strings.HasPrefix(trim, "---"),
		strings.HasPrefix(trim, "***"),
		strings.HasPrefix(trim, "|"),
		strings.HasPrefix(trim, "- "),
		strings.HasPrefix(trim, "* "),
		strings.HasPrefix(trim, "!["):
		return "raw"
	case isNumberedList(trim):
		return "raw"
	default:
		return "prose"
	}
}

var reNumList = regexp.MustCompile(`^\d+\.\s+`)

func isNumberedList(s string) bool { return reNumList.MatchString(s) }

var abbrevSet = map[string]bool{
	"mr.": true, "mrs.": true, "ms.": true, "dr.": true, "prof.": true,
	"st.": true, "rev.": true, "gen.": true, "gov.": true, "sen.": true,
	"rep.": true, "u.s.": true, "u.k.": true, "e.g.": true, "i.e.": true,
	"etc.": true, "vs.": true, "cf.": true, "no.": true, "vol.": true,
	"ch.": true, "pp.": true, "p.": true, "a.m.": true, "p.m.": true,
	"a.d.": true, "b.c.": true,
}

// SplitSentences splits a prose paragraph into sentences while protecting
// common abbreviations and decimal numbers.
func SplitSentences(para string) []string {
	para = strings.Join(strings.Fields(para), " ")
	if para == "" {
		return nil
	}
	var out []string
	start := 0
	n := len(para)
	for i := 0; i < n; i++ {
		c := para[i]
		if c != '.' && c != '!' && c != '?' {
			continue
		}
		end := i + 1
		// Consume closing quotes/parens attached to terminal punctuation.
		for end < n && (para[end] == '"' || para[end] == '\'' || para[end] == ')' || para[end] == ']') {
			end++
		}
		// Must be followed by end-of-string or whitespace.
		if end < n && para[end] != ' ' {
			continue
		}
		if c == '.' && isAbbrevEnding(para[start:i+1]) {
			continue
		}
		seg := strings.TrimSpace(para[start:end])
		if seg != "" {
			out = append(out, seg)
		}
		start = end
		for start < n && para[start] == ' ' {
			start++
		}
		i = start - 1
	}
	if start < n {
		if tail := strings.TrimSpace(para[start:]); tail != "" {
			out = append(out, tail)
		}
	}
	return out
}

func isAbbrevEnding(seg string) bool {
	fields := strings.Fields(seg)
	if len(fields) == 0 {
		return false
	}
	last := strings.ToLower(strings.TrimLeft(fields[len(fields)-1], `"'([`))
	if abbrevSet[last] {
		return true
	}
	// Single initial like "J." or "C."
	if len(last) == 2 && last[1] == '.' && last[0] >= 'a' && last[0] <= 'z' {
		return true
	}
	return false
}
