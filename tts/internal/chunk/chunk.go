package chunk

import (
	"strings"
	"unicode/utf8"
)

// RESTMaxChars is the xAI unary TTS hard limit.
const RESTMaxChars = 15000

// DefaultSoftMax is the packing target so speech tags and small drift stay under RESTMaxChars.
const DefaultSoftMax = 12000

// Chunk is one TTS request payload.
type Chunk struct {
	Index int
	Text  string
	Chars int
}

// Split packs annotated text into chunks at softMax characters, preferring
// split points at " [pause] " / " [long-pause] " boundaries, then sentence ends.
// Never exceeds hardMax (RESTMaxChars by default).
func Split(text string, softMax, hardMax int) []Chunk {
	if softMax <= 0 {
		softMax = DefaultSoftMax
	}
	if hardMax <= 0 {
		hardMax = RESTMaxChars
	}
	if hardMax < softMax {
		hardMax = softMax
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if utf8.RuneCountInString(text) <= softMax {
		return []Chunk{{Index: 0, Text: text, Chars: utf8.RuneCountInString(text)}}
	}

	// Work in runes for character limits (API counts characters; for ASCII prose len==runes).
	// xAI docs say "characters"; we use rune count for safety with any non-ASCII.
	var out []Chunk
	rest := text
	idx := 0
	for rest != "" {
		rest = strings.TrimSpace(rest)
		n := utf8.RuneCountInString(rest)
		if n <= softMax {
			out = append(out, Chunk{Index: idx, Text: rest, Chars: n})
			break
		}
		// Prefer a cut at or before softMax; if forced, before hardMax.
		cut := findCut(rest, softMax)
		if cut <= 0 {
			cut = findCut(rest, hardMax)
		}
		if cut <= 0 {
			// hard failsafe: cut at hardMax runes
			cut = runeOffset(rest, hardMax)
		}
		part := strings.TrimSpace(rest[:cut])
		out = append(out, Chunk{Index: idx, Text: part, Chars: utf8.RuneCountInString(part)})
		rest = rest[cut:]
		idx++
	}
	return out
}

// Estimate returns how many chunks a string would need at softMax.
func Estimate(text string, softMax int) int {
	return len(Split(text, softMax, RESTMaxChars))
}

func findCut(s string, maxRunes int) int {
	if utf8.RuneCountInString(s) <= maxRunes {
		return len(s)
	}
	limit := runeOffset(s, maxRunes)
	window := s[:limit]

	// Prefer structural pause boundaries (keep the tag with the left chunk).
	for _, sep := range []string{" [long-pause] ", " [pause] "} {
		if i := strings.LastIndex(window, sep); i > 0 {
			return i + len(sep)
		}
	}
	// Sentence boundary
	for i := len(window) - 1; i > 0; i-- {
		c := window[i]
		if c == '.' || c == '!' || c == '?' {
			// include trailing space if present
			j := i + 1
			if j < len(window) && window[j] == ' ' {
				j++
			}
			if j > maxRunes/4 { // avoid tiny leading chunks
				return j
			}
			break
		}
	}
	// Whitespace
	if i := strings.LastIndexAny(window, " \n\t"); i > maxRunes/4 {
		return i + 1
	}
	return limit
}

func runeOffset(s string, n int) int {
	if n <= 0 {
		return 0
	}
	count := 0
	for i := range s {
		if count == n {
			return i
		}
		count++
	}
	return len(s)
}
