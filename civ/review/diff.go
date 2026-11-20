package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// DiffOp is one run of a word diff.
type DiffOp int

const (
	DiffEqual DiffOp = iota
	DiffDelete
	DiffInsert
)

// Piece is a run of text that is unchanged, removed, or added.
type Piece struct {
	Op   DiffOp
	Text string
}

// tokenProductCap is the largest len(a)*len(b) the LCS will build.
// Larger inputs fall back to one delete run and one insert run.
const tokenProductCap = 1_000_000

// WordDiff compares a and b by lines, then by words inside each changed run.
// A word is a run of letters, numbers, or apostrophes. A whitespace run is
// its own token. Every other rune, including an em dash, is its own token.
func WordDiff(a, b string) []Piece {
	lineDiff := diffSeq(splitLines(a), splitLines(b))
	var out []Piece
	for i := 0; i < len(lineDiff); {
		if lineDiff[i].Op == DiffEqual {
			out = append(out, lineDiff[i])
			i++
			continue
		}
		var deleted, inserted string
		for i < len(lineDiff) && lineDiff[i].Op != DiffEqual {
			if lineDiff[i].Op == DiffDelete {
				deleted += lineDiff[i].Text
			} else {
				inserted += lineDiff[i].Text
			}
			i++
		}
		out = append(out, diffSeq(tokenize(deleted), tokenize(inserted))...)
	}
	return out
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.SplitAfter(s, "\n")
}

func tokenize(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if size < 1 {
			break
		}
		switch {
		case isWordRune(r):
			j := i + size
			for j < len(s) {
				r2, sz := utf8.DecodeRuneInString(s[j:])
				if !isWordRune(r2) {
					break
				}
				j += sz
			}
			out = append(out, s[i:j])
			i = j
		case unicode.IsSpace(r):
			j := i + size
			for j < len(s) {
				r2, sz := utf8.DecodeRuneInString(s[j:])
				if !unicode.IsSpace(r2) {
					break
				}
				j += sz
			}
			out = append(out, s[i:j])
			i = j
		default:
			out = append(out, s[i:i+size])
			i += size
		}
	}
	return out
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '\'' || r == '’'
}

func diffSeq(a, b []string) []Piece {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	if len(a)*len(b) > tokenProductCap {
		return blockPieces(a, b)
	}
	n, m := len(a), len(b)
	// dp[i][j] is the LCS length of a[i:] and b[j:].
	dp := make([]int, (n+1)*(m+1))
	at := func(i, j int) int { return i*(m+1) + j }
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[at(i, j)] = dp[at(i+1, j+1)] + 1
			} else if dp[at(i+1, j)] >= dp[at(i, j+1)] {
				dp[at(i, j)] = dp[at(i+1, j)]
			} else {
				dp[at(i, j)] = dp[at(i, j+1)]
			}
		}
	}
	raw := make([]Piece, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			raw = append(raw, Piece{DiffEqual, a[i]})
			i++
			j++
		case dp[at(i+1, j)] >= dp[at(i, j+1)]:
			raw = append(raw, Piece{DiffDelete, a[i]})
			i++
		default:
			raw = append(raw, Piece{DiffInsert, b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		raw = append(raw, Piece{DiffDelete, a[i]})
	}
	for ; j < m; j++ {
		raw = append(raw, Piece{DiffInsert, b[j]})
	}
	return coalesce(raw)
}

func blockPieces(a, b []string) []Piece {
	var out []Piece
	if len(a) > 0 {
		out = append(out, Piece{DiffDelete, strings.Join(a, "")})
	}
	if len(b) > 0 {
		out = append(out, Piece{DiffInsert, strings.Join(b, "")})
	}
	return out
}

func coalesce(raw []Piece) []Piece {
	var out []Piece
	for _, p := range raw {
		if p.Text == "" {
			continue
		}
		if len(out) > 0 && out[len(out)-1].Op == p.Op {
			out[len(out)-1].Text += p.Text
			continue
		}
		out = append(out, p)
	}
	return out
}
