package md

import (
	"regexp"
	"strings"
)

var (
	reHeading  = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	reListItem = regexp.MustCompile(`^(\d+\.|[-*+])\s+(.*)$`)
	reLink     = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)
	reImage    = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	reBold     = regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	reItalic   = regexp.MustCompile(`\*([^*]+)\*|_([^_]+)_`)
	reCode     = regexp.MustCompile("`([^`]+)`")
	reHTML     = regexp.MustCompile(`<[^>]+>`)
)

// Parse converts markdown chapter body into structural blocks suitable for speech.
// It does not expand template variables; call ExpandVars first if needed.
func Parse(markdown string) []Block {
	// Drop fenced code blocks entirely (not spoken in v1).
	markdown = stripFencedCode(markdown)

	lines := strings.Split(markdown, "\n")
	var blocks []Block
	var para []string
	var listBuf []string

	flushPara := func() {
		if len(para) == 0 {
			return
		}
		text := cleanInline(strings.Join(para, " "))
		text = collapseSpace(text)
		if text != "" {
			blocks = append(blocks, Block{Kind: KindParagraph, Text: text})
		}
		para = nil
	}
	flushList := func() {
		for _, item := range listBuf {
			text := cleanInline(item)
			text = collapseSpace(text)
			if text != "" {
				blocks = append(blocks, Block{Kind: KindListItem, Text: text})
			}
		}
		listBuf = nil
	}

	inList := false
	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t\r")
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			flushPara()
			flushList()
			inList = false
			continue
		}

		// Horizontal rules — treat as structural break
		if isHR(trimmed) {
			flushPara()
			flushList()
			inList = false
			continue
		}

		if m := reHeading.FindStringSubmatch(trimmed); m != nil {
			flushPara()
			flushList()
			inList = false
			level := len(m[1])
			text := cleanInline(m[2])
			text = collapseSpace(text)
			if text != "" {
				blocks = append(blocks, Block{Kind: KindHeading, Level: level, Text: text})
			}
			continue
		}

		if m := reListItem.FindStringSubmatch(trimmed); m != nil {
			flushPara()
			inList = true
			listBuf = append(listBuf, m[2])
			continue
		}

		if inList {
			// continuation line of a list item
			if len(listBuf) > 0 && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
				listBuf[len(listBuf)-1] += " " + trimmed
				continue
			}
			flushList()
			inList = false
		}

		para = append(para, trimmed)
	}
	flushPara()
	flushList()
	return blocks
}

// ParseChapter strips front matter, expands vars, and parses blocks.
func ParseChapter(data []byte, vars map[string]string) []Block {
	body := string(StripFrontMatter(data))
	body = ExpandVars(body, vars)
	return Parse(body)
}

func stripFencedCode(s string) string {
	var b strings.Builder
	lines := strings.Split(s, "\n")
	inFence := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func isHR(s string) bool {
	if len(s) < 3 {
		return false
	}
	for _, r := range s {
		if r != '-' && r != '*' && r != '_' && r != ' ' {
			return false
		}
	}
	compact := strings.ReplaceAll(s, " ", "")
	return len(compact) >= 3 && (strings.Trim(compact, "-") == "" ||
		strings.Trim(compact, "*") == "" || strings.Trim(compact, "_") == "")
}

func cleanInline(s string) string {
	s = reImage.ReplaceAllString(s, "")
	s = reLink.ReplaceAllString(s, "$1")
	s = reBold.ReplaceAllString(s, "$1$2")
	s = reItalic.ReplaceAllString(s, "$1$2")
	s = reCode.ReplaceAllString(s, "$1")
	s = reHTML.ReplaceAllString(s, "")
	// leftover emphasis markers
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	return s
}

func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// SpokenLen estimates characters that will be sent to TTS (approx; tags add a little more).
func SpokenLen(blocks []Block) int {
	n := 0
	for _, b := range blocks {
		n += len(b.Text)
	}
	return n
}

// IsMetadataOnly reports chapters that are only front matter / comments (e.g. title block).
func IsMetadataOnly(blocks []Block) bool {
	return len(blocks) == 0
}
