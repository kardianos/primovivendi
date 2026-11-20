package main

import (
	"bytes"
	"unicode/utf8"
)

// HitStatus is how a quote was found in its file.
type HitStatus string

const (
	// StatusOffset means the recorded byte range still equals the quote.
	StatusOffset HitStatus = "offset"
	// StatusMoved means the quote occurs once after the offset missed,
	// or the surrounding bytes picked one hit out of several.
	StatusMoved HitStatus = "moved"
	// StatusAmbiguous means more than one hit remains. The viewer does not pick.
	StatusAmbiguous HitStatus = "ambiguous"
	// StatusMissing means the quote does not occur.
	StatusMissing HitStatus = "missing"
)

// Hit is where an item sits in its file now. Line and Col are 1-based and
// derived from Offset. Col counts runes.
type Hit struct {
	Status     HitStatus
	Offset     int
	End        int
	Line       int
	Col        int
	Candidates []int
}

// Locate finds quote in file.
//
// A recorded byte_offset that still matches is StatusOffset. Otherwise every
// occurrence is collected. before and after, when set, keep the hits they
// border. One remaining hit is StatusMoved. Several, or none while raw hits
// exist, is StatusAmbiguous. No raw hit is StatusMissing.
func Locate(file []byte, it Item) Hit {
	if it.Quote == "" {
		return Hit{Status: StatusMissing}
	}
	if off, end, ok := offsetMatch(file, it); ok {
		line, col := lineCol(file, off)
		return Hit{Status: StatusOffset, Offset: off, End: end, Line: line, Col: col}
	}
	raw := findAll(file, it.Quote)
	if len(raw) == 0 {
		return Hit{Status: StatusMissing}
	}
	endOf := func(off int) int { return off + len(it.Quote) }
	var fit []int
	if it.Before != "" || it.After != "" {
		for _, off := range raw {
			if contextOK(file, off, endOf(off), it.Before, it.After) {
				fit = append(fit, off)
			}
		}
	} else {
		fit = raw
	}
	switch {
	case len(fit) == 1:
		off := fit[0]
		end := endOf(off)
		line, col := lineCol(file, off)
		return Hit{Status: StatusMoved, Offset: off, End: end, Line: line, Col: col, Candidates: fit}
	case len(fit) == 0:
		return Hit{Status: StatusAmbiguous, Candidates: append([]int(nil), raw...)}
	default:
		return Hit{Status: StatusAmbiguous, Candidates: append([]int(nil), fit...)}
	}
}

func offsetMatch(file []byte, it Item) (off, end int, ok bool) {
	if it.ByteOffset == nil {
		return 0, 0, false
	}
	off = *it.ByteOffset
	end = off + len(it.Quote)
	if it.ByteEnd != nil {
		end = *it.ByteEnd
	}
	if off < 0 || end < off || end > len(file) {
		return 0, 0, false
	}
	if string(file[off:end]) != it.Quote {
		return 0, 0, false
	}
	return off, end, true
}

func findAll(file []byte, quote string) []int {
	q := []byte(quote)
	var hits []int
	for from := 0; from <= len(file); {
		i := bytes.Index(file[from:], q)
		if i < 0 {
			break
		}
		hits = append(hits, from+i)
		from = from + i + 1
	}
	return hits
}

func contextOK(file []byte, off, end int, before, after string) bool {
	if before != "" {
		if off < len(before) || string(file[off-len(before):off]) != before {
			return false
		}
	}
	if after != "" {
		if end > len(file) || end+len(after) > len(file) || string(file[end:end+len(after)]) != after {
			return false
		}
	}
	return true
}

// lineCol returns a 1-based line and a 1-based rune column for a byte offset.
func lineCol(file []byte, off int) (line, col int) {
	if off < 0 {
		off = 0
	}
	if off > len(file) {
		off = len(file)
	}
	line, col = 1, 1
	for i := 0; i < off; {
		if file[i] == '\n' {
			line++
			col = 1
			i++
			continue
		}
		_, size := utf8.DecodeRune(file[i:])
		if size < 1 {
			size = 1
		}
		col++
		i += size
	}
	return line, col
}

// defaultContextLines is how many lines the passage shows on each side of a
// quote until the reader asks for more.
const defaultContextLines = 2

// snippet returns the bytes around a hit, split so the quote can be marked.
// beforeLines and afterLines count lines beyond the match line. A negative
// count means that side through the end of the file. ok is false when the
// hit has no single offset.
func snippet(file []byte, h Hit, beforeLines, afterLines int) (pre, match, post string, ok bool) {
	start, stop, ok := snippetWindow(file, h, beforeLines, afterLines)
	if !ok {
		return "", "", "", false
	}
	return string(file[start:h.Offset]), string(file[h.Offset:h.End]), string(file[h.End:stop]), true
}

// snippetWindow is the byte range snippet shows. The match line is always
// included. beforeLines and afterLines add whole lines beyond it, and a
// negative count runs to that end of the file.
func snippetWindow(file []byte, h Hit, beforeLines, afterLines int) (start, stop int, ok bool) {
	if h.Status != StatusOffset && h.Status != StatusMoved {
		return 0, 0, false
	}
	if h.Offset < 0 || h.End < h.Offset || h.End > len(file) {
		return 0, 0, false
	}
	start = lineStart(file, h.Offset)
	if beforeLines < 0 {
		start = 0
	} else {
		for i := 0; i < beforeLines && start > 0; i++ {
			start = lineStart(file, start-1)
		}
	}
	stop = h.End
	if afterLines < 0 {
		stop = len(file)
	} else {
		// Finish the match's own line before counting lines below it,
		// symmetric with lineStart(file, h.Offset) above.
		if stop < len(file) && (stop == 0 || file[stop-1] != '\n') {
			if next := bytes.IndexByte(file[stop:], '\n'); next < 0 {
				stop = len(file)
			} else {
				stop += next + 1
			}
		}
		for i := 0; i < afterLines && stop < len(file); i++ {
			next := bytes.IndexByte(file[stop:], '\n')
			if next < 0 {
				stop = len(file)
				break
			}
			stop += next + 1
		}
	}
	return start, stop, true
}

// moreContext is the next wider line count: 2, 8, 24, 80, then the rest of
// the file.
func moreContext(n int) int {
	switch {
	case n < 0:
		return -1
	case n < 8:
		return 8
	case n < 24:
		return 24
	case n < 80:
		return 80
	default:
		return -1
	}
}

func lineStart(file []byte, off int) int {
	if off > len(file) {
		off = len(file)
	}
	for off > 0 && file[off-1] != '\n' {
		off--
	}
	return off
}
