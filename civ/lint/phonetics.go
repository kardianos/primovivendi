package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const cmuURL = "https://raw.githubusercontent.com/cmusphinx/cmudict/master/cmudict.dict"

// notAdverbLy are words that end in the letters "ly" without the unstressed
// adverb suffix.
var notAdverbLy = map[string]bool{
	"family": true, "only": true, "early": true, "holy": true, "ugly": true,
	"apply": true, "reply": true, "supply": true, "rely": true, "fly": true,
	"italy": true, "assembly": true, "monopoly": true, "anomaly": true,
	"comply": true, "imply": true, "multiply": true, "july": true,
}

var functionWord = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`the a an of to and or but in on for with at by from as that this these those it its be is are was were been being have has had do does did not no nor so if then than too very just also only into over after before about between through under such other own same more most many much can may will shall must would could should their there here when where which who whom what how all each every both few some any our your his her she him they them we you he one out up down`) {
		functionWord[w] = true
	}
}

var (
	reMDLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	reMDFoot = regexp.MustCompile(`\[\^[^\]]*\]`)
	reMDHTML = regexp.MustCompile(`<[^>]+>`)
	reMDSent = regexp.MustCompile(`([.!?])\s+`)
)

func cmuPath(root string) string {
	return filepath.Join(root, "ext-ref", "phonetics", "cmudict.dict")
}

func ensureCMU(root string) (string, error) {
	path := cmuPath(root)
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		fmt.Printf("cmudict: %s\n", path)
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	fmt.Printf("downloading %s\n", cmuURL)
	resp, err := http.Get(cmuURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cmudict: %s", resp.Status)
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}
	fmt.Printf("cmudict: %s\n", path)
	return path, nil
}

// dict maps an uppercase word to its phonemes. The first pronunciation wins.
type dict map[string][]string

func loadCMU(r io.Reader) (dict, error) {
	d := dict{}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, ";;;") {
			continue
		}
		word, phones, ok := strings.Cut(line, "  ")
		if !ok {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			word, phones = fields[0], strings.Join(fields[1:], " ")
		}
		word, _, _ = strings.Cut(word, "(")
		word = strings.ToUpper(word)
		if _, exists := d[word]; exists {
			continue
		}
		d[word] = strings.Fields(phones)
	}
	return d, sc.Err()
}

func loadCMUFile(path string) (dict, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadCMU(f)
}

// rhymeTail is the phonemes from the last primary-stressed vowel through the
// end. A tail of a single vowel is dropped; it matches too many words.
func rhymeTail(phones []string) string {
	last := -1
	for i, p := range phones {
		if strings.HasSuffix(p, "1") {
			last = i
		}
	}
	if last < 0 {
		for i, p := range phones {
			if strings.HasSuffix(p, "2") {
				last = i
			}
		}
	}
	if last < 0 || len(phones)-last < 2 {
		return ""
	}
	return strings.Join(phones[last:], " ")
}

func heavySuffix(word string, phones []string) string {
	w := strings.ToLower(word)
	switch {
	case strings.HasSuffix(w, "tion") || strings.HasSuffix(w, "sion"):
		if len(w) > 5 && (phones == nil || tailIs(phones, "SH AH0 N") || tailIs(phones, "ZH AH0 N") || tailIs(phones, "CH AH0 N")) {
			return "-tion"
		}
	case strings.HasSuffix(w, "tions") || strings.HasSuffix(w, "sions"):
		if len(w) > 6 && (phones == nil || tailIs(phones, "SH AH0 N Z") || tailIs(phones, "ZH AH0 N Z") || tailIs(phones, "CH AH0 N Z")) {
			return "-tion"
		}
	case strings.HasSuffix(w, "ment"):
		if len(w) > 6 && (phones == nil || tailIs(phones, "M AH0 N T")) {
			return "-ment"
		}
	case strings.HasSuffix(w, "ments"):
		if len(w) > 7 && (phones == nil || tailIs(phones, "M AH0 N T S")) {
			return "-ment"
		}
	case strings.HasSuffix(w, "ity"):
		if len(w) > 5 && (phones == nil || tailIs(phones, "AH0 T IY0")) {
			return "-ity"
		}
	case strings.HasSuffix(w, "ities"):
		if len(w) > 7 && (phones == nil || tailIs(phones, "AH0 T IY0 Z")) {
			return "-ity"
		}
	case strings.HasSuffix(w, "ness"):
		if len(w) > 5 && (phones == nil || tailIs(phones, "N AH0 S")) {
			return "-ness"
		}
	case strings.HasSuffix(w, "nesses"):
		if len(w) > 7 && (phones == nil || tailIs(phones, "N AH0 S AH0 Z") || tailIs(phones, "N AH0 S IH0 Z")) {
			return "-ness"
		}
	case strings.HasSuffix(w, "ly"):
		if len(w) > 4 && !notAdverbLy[w] && (phones == nil || tailIs(phones, "L IY0")) {
			return "-ly"
		}
	}
	return ""
}

func tailIs(phones []string, tail string) bool {
	return strings.HasSuffix(strings.Join(phones, " "), tail)
}

func roughLemma(w string) string {
	w = strings.ToLower(w)
	for _, suf := range []string{"ing", "es", "ed", "s"} {
		if strings.HasSuffix(w, suf) && len(w)-len(suf) >= 3 {
			return strings.TrimSuffix(w, suf)
		}
	}
	return w
}

type prosePara struct {
	line int
	text string
}

// proseParas drops headings, tables, rules, and code, strips the markup a
// reader does not hear, and keeps paragraph boundaries so sentences never
// merge across paragraphs or list items.
func proseParas(text string) []prosePara {
	var out []prosePara
	var cur []string
	curLine := 0
	fence, quote := false, false
	flush := func() {
		if len(cur) == 0 {
			return
		}
		t := strings.Join(cur, " ")
		cur = nil
		t = reMDLink.ReplaceAllString(t, "$1")
		t = reMDFoot.ReplaceAllString(t, "")
		t = reMDHTML.ReplaceAllString(t, " ")
		t = strings.NewReplacer("**", "", "*", "", "`", "", "\\", "").Replace(t)
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, prosePara{curLine, t})
		}
	}
	for i, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "```") {
			fence = !fence
			flush()
			continue
		}
		if fence {
			continue
		}
		q := false
		for strings.HasPrefix(t, ">") {
			t = strings.TrimSpace(strings.TrimPrefix(t, ">"))
			q = true
		}
		if q != quote {
			flush()
			quote = q
		}
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "|") ||
			strings.HasPrefix(t, "---") || strings.HasPrefix(t, "<!--") || strings.HasPrefix(t, ":::") {
			flush()
			continue
		}
		if listMark.MatchString(t) {
			flush()
			t = strings.TrimSpace(listMark.ReplaceAllString(t, ""))
		}
		if t == "" {
			continue
		}
		if len(cur) == 0 {
			curLine = i + 1
		}
		cur = append(cur, t)
	}
	flush()
	return out
}

type sentence struct {
	line int
	text string
}

func sentencesOf(text string) []sentence {
	var out []sentence
	for _, pp := range proseParas(text) {
		var cur strings.Builder
		flush := func() {
			t := strings.TrimSpace(cur.String())
			cur.Reset()
			if t != "" {
				out = append(out, sentence{pp.line, t})
			}
		}
		rest := pp.text
		for {
			loc := reMDSent.FindStringIndex(rest)
			if loc == nil {
				cur.WriteString(rest)
				break
			}
			cur.WriteString(rest[:loc[1]])
			// Keep the terminator; start a new sentence when the word before
			// the dot is not an abbreviation such as "Mr."
			end := loc[0]
			word := lastWord(rest[:end])
			if !abbrev(word) {
				flush()
			}
			rest = rest[loc[1]:]
		}
		flush()
	}
	return out
}

func lastWord(s string) string {
	s = strings.TrimRight(s, ".!?")
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[len(f)-1]
}

func abbrev(w string) bool {
	w = strings.Trim(w, "*_")
	switch strings.ToLower(w) {
	case "mr", "mrs", "ms", "dr", "st", "jr", "sr", "prof", "rev", "gen", "col", "lt", "sgt", "no", "fig", "eq", "vs", "etc", "approx":
		return true
	}
	return len(w) == 1 && unicode.IsLetter([]rune(w)[0])
}

type tok struct {
	word, lemma, tail, suffix string
	idx                       int
}

func analyzeSentence(s string, d dict) (suffixes []string, rhymes []string) {
	raw := wordsOf(s)
	var toks []tok
	for i, w := range raw {
		w = strings.Trim(w, "'")
		low := strings.ToLower(w)
		if low == "" || functionWord[low] {
			continue
		}
		phones := d[strings.ToUpper(low)]
		// idx is the word's place in the sentence, so "within 10 words"
		// counts the function words that sit between two content words.
		t := tok{word: w, lemma: roughLemma(low), idx: i}
		if phones != nil {
			t.tail = rhymeTail(phones)
		}
		t.suffix = heavySuffix(low, phones)
		toks = append(toks, t)
	}
	bySuffix := map[string][]string{}
	suffixOrder := []string{"-tion", "-ment", "-ity", "-ness", "-ly"}
	for _, t := range toks {
		if t.suffix != "" {
			bySuffix[t.suffix] = append(bySuffix[t.suffix], t.word)
		}
	}
	for _, suf := range suffixOrder {
		if ws := bySuffix[suf]; len(ws) >= 3 {
			suffixes = append(suffixes, fmt.Sprintf("%s: %s", suf, strings.Join(ws, ", ")))
		}
	}
	var emitted []rhymeGroup
	for i := 0; i < len(toks); i++ {
		if toks[i].tail == "" {
			continue
		}
		group := []string{toks[i].word}
		lemmas := map[string]bool{toks[i].lemma: true}
		lemmaList := []string{toks[i].lemma}
		for j := i + 1; j < len(toks) && toks[j].idx-toks[i].idx <= 10; j++ {
			if toks[j].tail != toks[i].tail || lemmas[toks[j].lemma] {
				continue
			}
			if prefixRelatedAny(toks[j].lemma, lemmaList) {
				continue
			}
			lemmas[toks[j].lemma] = true
			lemmaList = append(lemmaList, toks[j].lemma)
			group = append(group, toks[j].word)
		}
		if len(group) < 2 {
			continue
		}
		if isSubsetRhyme(emitted, toks[i].tail, lemmas) {
			continue
		}
		emitted = append(emitted, rhymeGroup{toks[i].tail, lemmas})
		rhymes = append(rhymes, fmt.Sprintf("%s (%s)", strings.Join(group, " / "), toks[i].tail))
	}
	return suffixes, rhymes
}

var contrastPrefixes = []string{
	"counter", "super", "under", "inter", "anti", "meta", "over", "post", "self",
	"non", "dis", "mis", "pre", "sub", "out", "con", "un", "in", "im", "il", "ir", "re", "de", "co",
}

// prefixRelated reports deliberate morphological contrast pairs such as
// moral/immoral, order/disorder, physical/metaphysical, or believers/non-believers.
func prefixRelated(a, b string) bool {
	if len(a) == len(b) {
		return false
	}
	long, short := a, b
	if len(long) < len(short) {
		long, short = short, long
	}
	if len(short) < 4 || !strings.HasSuffix(long, short) {
		return false
	}
	pre := strings.TrimSuffix(long, short)
	if strings.HasSuffix(pre, "-") {
		return true
	}
	for _, p := range contrastPrefixes {
		if pre == p {
			return true
		}
	}
	return false
}

func prefixRelatedAny(lemma string, others []string) bool {
	for _, o := range others {
		if prefixRelated(lemma, o) {
			return true
		}
	}
	return false
}

type rhymeGroup struct {
	tail   string
	lemmas map[string]bool
}

func isSubsetRhyme(emitted []rhymeGroup, tail string, lemmas map[string]bool) bool {
	for _, e := range emitted {
		if e.tail != tail {
			continue
		}
		sub := true
		for l := range lemmas {
			if !e.lemmas[l] {
				sub = false
				break
			}
		}
		if sub {
			return true
		}
	}
	return false
}

func runPhonetics(root string, files []string) error {
	path, err := ensureCMU(root)
	if err != nil {
		return err
	}
	d, err := loadCMUFile(path)
	if err != nil {
		return err
	}
	fmt.Printf("cmudict entries: %d\n", len(d))
	fmt.Println("=== Heavy suffixes (3 or more in one sentence) ===")
	sufN, rhymeN := 0, 0
	var rhymeHits []string
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		for _, s := range sentencesOf(string(b)) {
			suf, rhy := analyzeSentence(s.text, d)
			for _, item := range suf {
				fmt.Printf("  %s:%d: %s\n", rel(root, f), s.line, item)
				sufN++
			}
			for _, item := range rhy {
				if rhymeN < 40 {
					rhymeHits = append(rhymeHits, fmt.Sprintf("  %s:%d: %s", rel(root, f), s.line, item))
				}
				rhymeN++
			}
		}
	}
	if sufN == 0 {
		fmt.Println("  (none)")
	}
	fmt.Printf("=== Rhyme tails (%d; up to 40 shown) ===\n", rhymeN)
	if rhymeN == 0 {
		fmt.Println("  (none)")
	}
	for _, line := range rhymeHits {
		fmt.Println(line)
	}
	return nil
}
