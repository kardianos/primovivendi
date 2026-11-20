package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Part A patterns from civ/notes/06_editing_guide.md, plus the double hyphen
// the typography rules forbid. These scan the line as written, quotations
// included, matching the guide's audit script.
var partA = []struct {
	name string
	re   *regexp.Regexp
}{
	{"Em-dash or exclamation", regexp.MustCompile(`[—!]`)},
	{"Math / symbols in text", regexp.MustCompile(`(\$.*?\$|\\\w+|\s[+=≈×÷&]\s|\b[A-Z]\s*[+=]\s*[A-Z]\b)`)},
	{"exact same / every single", regexp.MustCompile(`(?i)\b(exact same|every single)\b`)},
	{"zero", regexp.MustCompile(`(?i)\bzero\b`)},
	{"Imperative openers", regexp.MustCompile(`(?i)(?:^|[.!?]\s+|\*\*\s+|>\s+)(Look\s+at|Look\s+closely|Look\s+back|Step\s+back|Notice\s+|Watch\s+|Observe\s+|Examine\s+|Consider\s+|Imagine\s+|Picture\s+)`)},
	{"on earth", regexp.MustCompile(`(?i)\bon earth\b`)},
	{"raw / real-world / rot", regexp.MustCompile(`(?i)\b(raw|real-world|rot|rotted|rotting|rots)\b`)},
	{"honest", regexp.MustCompile(`(?i)\bhonest\w*\b`)},
	{"load-bearing", regexp.MustCompile(`(?i)\bload[- ]bearing\b`)},
	{"Why... Because", regexp.MustCompile(`(?i)Why[^.!?]*\?\s*\*?\*?Because\b`)},
	{"Question then fragment", regexp.MustCompile(`\?\s+\*{0,2}(No|Yes|Never|Nothing)\.|What (happened|occurred)[^?]*\?`)},
}

var termRef = regexp.MustCompile(`(?i)\b(principles?|virtues?)\s+\d+`)

var principles = []string{
	"Reality is real, knowable, and consequential. Truth is what corresponds to reality.",
	"Humans are self-interested, limited, and particular.",
	"Goodness must include human life continuing, and that life kept and built.",
	"What you keep and build is self-anchored, tiered, and outward. Enemies are real.",
	"Justice is scoped procedural symmetry.",
}

var virtues = []string{
	"Align with truth",
	"Take ownership, which is responsibility",
	"Seek excellence",
	"Admire beauty",
	"Be humble",
	"Have children",
	"Form and listen",
}

var (
	listItem  = regexp.MustCompile(`^(\s*>\s*)?(\d+)\.\s+(.*)$`)
	prinStmt  = regexp.MustCompile(`(?i)^\s*>\s*\**\s*Principle\s+(\d+)\s+of Reality\**\s*:\s*(.*?)\s*\**\s*$`)
	virtStmt  = regexp.MustCompile(`(?i)^\s*>\s*\**\s*Virtue\s+(\d+)\**\s*:\s*(.*?)\s*\**\s*$`)
	breakLine = regexp.MustCompile(`^\s*>?\s*-{3,}\s*$`)
)

func runProse(root string, files []string) error {
	var grouped [][]hit
	names := []string{}
	for _, c := range partA {
		names = append(names, c.name)
		grouped = append(grouped, nil)
	}
	names = append(names, "Double hyphen")
	grouped = append(grouped, nil)
	var spelling, casing, locked []hit
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		lines := strings.Split(string(b), "\n")
		for i, line := range lines {
			for c, check := range partA {
				if check.re.MatchString(line) {
					grouped[c] = append(grouped[c], hit{f, i + 1, check.name, strings.TrimSpace(line)})
				}
			}
			if doubleHyphen(line) {
				grouped[len(partA)] = append(grouped[len(partA)], hit{f, i + 1, "Double hyphen", strings.TrimSpace(line)})
			}
			spelling = append(spelling, spellingHits(f, i+1, line)...)
			casing = append(casing, casingHits(f, i+1, line)...)
		}
		locked = append(locked, lockedHits(f, lines)...)
	}
	for i, name := range names {
		printHits(root, name, grouped[i])
	}
	printHits(root, "British spelling", spelling)
	printHits(root, "Principle / Virtue casing", casing)
	printHits(root, "Locked Principles and Virtues", locked)
	return nil
}

// doubleHyphen flags "--" in prose. A Markdown rule (a line of three or more
// dashes) is not a double hyphen.
func doubleHyphen(line string) bool {
	if breakLine.MatchString(line) || strings.TrimSpace(line) == "" {
		return false
	}
	trim := strings.Trim(strings.TrimSpace(line), "|:- ")
	if trim == "" && strings.Contains(line, "-") {
		return false // a table rule
	}
	return strings.Contains(line, "--")
}

func casingHits(file string, line int, text string) []hit {
	var out []hit
	for _, m := range termRef.FindAllStringSubmatchIndex(text, -1) {
		word := text[m[2]:m[3]]
		want := canonicalTerm(word)
		if word != want {
			out = append(out, hit{file, line, "casing", fmt.Sprintf("%q should be %q", word, want)})
		}
	}
	return out
}

func canonicalTerm(word string) string {
	switch strings.ToLower(word) {
	case "principle":
		return "Principle"
	case "principles":
		return "Principles"
	case "virtue":
		return "Virtue"
	case "virtues":
		return "Virtues"
	default:
		return word
	}
}

type numbered struct {
	n    int
	text string
	line int
}

func lockedHits(file string, lines []string) []hit {
	var out []hit
	var run []numbered
	flush := func() {
		out = append(out, checkNumbered(file, run)...)
		run = nil
	}
	for i, line := range lines {
		m := listItem.FindStringSubmatch(line)
		if m == nil {
			if strings.TrimSpace(line) != "" {
				flush()
			}
			if pm := prinStmt.FindStringSubmatch(line); pm != nil {
				out = append(out, checkStatement(file, i+1, pm[1], plain(pm[2]), principles, "Principle")...)
			}
			if vm := virtStmt.FindStringSubmatch(line); vm != nil {
				out = append(out, checkStatement(file, i+1, vm[1], plain(vm[2]), virtues, "Virtue")...)
			}
			continue
		}
		n, _ := strconv.Atoi(m[2])
		if len(run) > 0 && n != run[len(run)-1].n+1 {
			flush()
		}
		run = append(run, numbered{n, plain(m[3]), i + 1})
	}
	flush()
	return out
}

func checkNumbered(file string, run []numbered) []hit {
	if len(run) == 0 || run[0].n != 1 {
		return nil
	}
	var out []hit
	if len(run) == len(principles) && resembles(run, principles) {
		for i, it := range run {
			if it.text != principles[i] {
				out = append(out, hit{file, it.line, "principle", fmt.Sprintf("Principle %d is not the locked line: %s", i+1, clip(it.text))})
			}
		}
	}
	if len(run) == len(virtues) && resembles(run, virtues) {
		for i, it := range run {
			if it.text != virtues[i] && it.text != virtues[i]+"." {
				out = append(out, hit{file, it.line, "virtue", fmt.Sprintf("Virtue %d is not the locked line: %s", i+1, clip(it.text))})
			}
		}
	}
	return out
}

func checkStatement(file string, line int, num, body string, canon []string, label string) []hit {
	n, err := strconv.Atoi(num)
	if err != nil || n < 1 || n > len(canon) {
		return nil
	}
	if strings.TrimRight(body, ".") == strings.TrimRight(canon[n-1], ".") {
		return nil
	}
	return []hit{{file, line, label, fmt.Sprintf("%s %s is not the locked line: %s", label, num, clip(body))}}
}

func resembles(run []numbered, canon []string) bool {
	hitN := 0
	for i := range canon {
		if looksSame(run[i].text, canon[i]) {
			hitN++
		}
	}
	return hitN >= 3
}

func looksSame(got, want string) bool {
	g, w := strings.ToLower(plain(got)), strings.ToLower(want)
	cut := min(18, max(8, len(w)/2))
	if strings.Contains(g, w) || (strings.Contains(w, g) && len(g) >= cut) {
		return true
	}
	return sharedPrefix(g, w) >= cut
}

func sharedPrefix(a, b string) int {
	n := 0
	for len(a) > 0 && len(b) > 0 {
		ra, sa := utf8.DecodeRuneInString(a)
		rb, sb := utf8.DecodeRuneInString(b)
		if ra != rb {
			break
		}
		n++
		a, b = a[sa:], b[sb:]
	}
	return n
}

func plain(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "`", "")
	return strings.TrimSpace(s)
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) > 80 {
		return string([]rune(s)[:80]) + "..."
	}
	return s
}

// outsideQuotes blanks direct quotations so British spellings inside them are
// not flagged. Straight and curly double quotes both count.
func outsideQuotes(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch r {
		case '"':
			in = !in
			b.WriteRune(' ')
			continue
		case '“', '«':
			in = true
			b.WriteRune(' ')
			continue
		case '”', '»':
			in = false
			b.WriteRune(' ')
			continue
		}
		if in {
			b.WriteRune(' ')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func wordsOf(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\'' && r != '-'
	})
}
