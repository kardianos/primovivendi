package annotate

import (
	"strings"

	"po/tts/internal/md"
)

// Options control pause insertion. Structure-only: pauses around headings,
// paragraphs, and list items — no emphasis or delivery wrapping tags.
type Options struct {
	// PauseBetweenParagraphs inserts [pause] between consecutive body blocks.
	// Default true when zero value is used via DefaultOptions.
	PauseBetweenParagraphs bool
}

// DefaultOptions returns the recommended structure-only settings.
func DefaultOptions() Options {
	return Options{PauseBetweenParagraphs: true}
}

// Text converts structural blocks into a single xAI-TTS-ready string using
// only pause tags for structure (no <emphasis>, <slow>, etc.).
//
// Mapping:
//   - heading level 1–2: [long-pause] Title. [long-pause]
//   - heading level 3–6: [long-pause] Title. [pause]
//   - paragraph / list item: body text; [pause] before next body block
func Text(blocks []md.Block, opt Options) string {
	if len(blocks) == 0 {
		return ""
	}

	var b strings.Builder
	prevWasBody := false

	for _, block := range blocks {
		switch block.Kind {
		case md.KindHeading:
			prevWasBody = false
			title := ensureTerminalPunct(block.Text)
			if block.Level <= 2 {
				writeGap(&b)
				b.WriteString("[long-pause] ")
				b.WriteString(title)
				b.WriteString(" [long-pause]")
			} else {
				writeGap(&b)
				b.WriteString("[long-pause] ")
				b.WriteString(title)
				b.WriteString(" [pause]")
			}

		case md.KindParagraph, md.KindListItem:
			text := strings.TrimSpace(block.Text)
			if text == "" {
				continue
			}
			if prevWasBody && opt.PauseBetweenParagraphs {
				b.WriteString(" [pause] ")
			} else {
				writeGap(&b)
			}
			b.WriteString(text)
			prevWasBody = true
		}
	}

	return strings.TrimSpace(b.String())
}

func writeGap(b *strings.Builder) {
	if b.Len() == 0 {
		return
	}
	// ensure whitespace between segments
	s := b.String()
	if s[len(s)-1] != ' ' && s[len(s)-1] != '\n' {
		b.WriteByte(' ')
	}
}

func ensureTerminalPunct(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	last := s[len(s)-1]
	switch last {
	case '.', '!', '?', ':', ';':
		return s
	default:
		return s + "."
	}
}
