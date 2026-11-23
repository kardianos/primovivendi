package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Voice diagnostic: measures the "two-voice" seam between the book's Plain Style
// (terse, concrete, punchy declarative prose) and its Drafted/Textbook Voice
// (polysyllabic Latinate abstraction, uniform long sentences, outline signposts,
// parallel triplets, and instructional coaching posture).
//
// Deliberately excludes primary-source citation density (conceptual chapters
// legitimately cite fewer historical sources) and Markdown formatting (bold/lists)
// so it measures prose voice itself.
//
// Each of the 6 dimensions is anchored to fixed Plain-Style baseline constants
// (mean, scale) so scores do not drift as chapters are edited:
//   0.0  = Plain Style baseline
//   <0   = plainer / terser than baseline
//   >=1  = noticeably Textbook Voice
//   |Δ| >= 1.2 between adjacent sections in one chapter = sharp voice seam.

type voiceOptions struct {
	minWords  int
	threshold float64
	seamCut   float64
	showAll   bool
	failScore float64
	failSeam  float64
}

type voiceMetrics struct {
	Words        int
	Sentences    int
	LongWordPct  float64 // % words with >= 8 letters
	LatinatePct  float64 // % words matching Latinate abstract suffixes
	ShortSentPct float64 // % sentences with <= 12 words
	SignpostPct  float64 // % sentences opening with outline/didactic signposts
	Triplets100  float64 // "X, Y, and Z" triplets per 100 words
	Coaching100  float64 // you/your + 3*? + 2*scareQuotes + 2*intensifierLy per 100 words
	Score        float64 // unweighted mean of the 6 anchored z-scores
	Z            [6]float64
}

type voiceSection struct {
	File    string
	Line    int
	Heading string
	Metrics voiceMetrics
}

type voiceSeam struct {
	File string
	Jump float64
	Prev voiceSection
	Next voiceSection
}

// Fixed Plain-Style calibration constants (mean, scale) for the 6 dimensions.
var voiceBaselines = [6]struct {
	name  string
	mean  float64
	scale float64
}{
	{"lw", 15.0, 5.0},    // LongWordPct (%)
	{"lat", 3.0, 2.0},    // LatinatePct (%)
	{"shS", 33.0, 10.0},  // ShortSentPct (%, inverted: lower short_s => higher score)
	{"sig", 2.0, 10.0},   // SignpostPct (%)
	{"tri", 0.35, 0.45},  // Triplets per 100 words
	{"coach", 0.50, 1.5}, // Coaching markers per 100 words
}

var (
	voiceWordRe       = regexp.MustCompile(`[A-Za-z']+`)
	voiceLatinateRe   = regexp.MustCompile(`^[a-z]{5,}(?:tion|sion|ment|ity|ism|ist|ive|ize|ise|ance|ence|ical|ology|atory)$`)
	voiceSignpostRe   = regexp.MustCompile(`^(?:First|Second|Third|Fourth|Fifth|Finally|Therefore|Consequently|Conversely|However|Moreover|Furthermore|In practice|In reality|By contrast|For example|Suppose|Consider|Notice|Remember|Tested against)\b`)
	voiceScareQuoteRe = regexp.MustCompile(`"[A-Za-z][A-Za-z '-]{1,28}"`)
	voiceTripletRe    = regexp.MustCompile(`(?i)\b[a-z]{3,},\s+(?:[a-z]{3,}\s+){0,2}[a-z]{3,},\s+and\s+(?:[a-z]{3,}\s+){0,2}[a-z]{3,}\b`)
	voiceHeadingRe    = regexp.MustCompile(`^#{1,6}\s+(.*)$`)
	voiceListLeadRe   = regexp.MustCompile(`^\s*(?:\* |- |\d+\.\s+)`)
	voiceQuoteLeadRe  = regexp.MustCompile(`^\s*>\s*`)
	voiceSpaceRe      = regexp.MustCompile(`\s+`)
	voiceSentSplitRe  = regexp.MustCompile(`[.!?]+(?:\s+|$)`)
)

var voiceCoachingYou = map[string]bool{
	"you": true, "your": true, "yourself": true, "yours": true,
}

var voiceIntensifierLy = map[string]bool{
	"explicitly": true, "systematically": true, "fundamentally": true,
	"deliberately": true, "strictly": true, "completely": true,
	"entirely": true, "immediately": true, "inevitably": true,
	"inherently": true, "simultaneously": true, "predictably": true,
	"progressively": true, "unconditionally": true, "indiscriminately": true,
	"permanently": true, "precipitously": true, "exclusively": true,
	"consequently": true, "conversely": true, "crucially": true,
	"simply": true, "merely": true, "constantly": true,
	"regularly": true, "actively": true, "directly": true,
	"genuinely": true, "visibly": true, "literally": true,
	"physically": true, "biologically": true, "historically": true,
	"morally": true, "legally": true, "empirically": true,
}

func measureVoice(raw string) voiceMetrics {
	var cleanLines []string
	for _, line := range strings.Split(raw, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || breakLine.MatchString(t) {
			cleanLines = append(cleanLines, "")
			continue
		}
		line = voiceQuoteLeadRe.ReplaceAllString(line, "")
		line = voiceListLeadRe.ReplaceAllString(line, "")
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "*", "")
		cleanLines = append(cleanLines, line)
	}
	plain := strings.Join(cleanLines, "\n")

	qmarks := strings.Count(plain, "?")
	scareQuotes := len(voiceScareQuoteRe.FindAllString(plain, -1))
	triplets := len(voiceTripletRe.FindAllString(plain, -1))

	rawWords := voiceWordRe.FindAllString(plain, -1)
	var words []string
	longW, latinate, youCnt, intensCnt := 0, 0, 0, 0
	for _, rw := range rawWords {
		w := strings.ToLower(strings.Trim(rw, "'"))
		if w == "" {
			continue
		}
		words = append(words, w)
		if len(w) >= 8 {
			longW++
		}
		if voiceLatinateRe.MatchString(w) {
			latinate++
		}
		if voiceCoachingYou[w] {
			youCnt++
		}
		if voiceIntensifierLy[w] {
			intensCnt++
		}
	}

	n := len(words)
	if n == 0 {
		return voiceMetrics{}
	}

	// Sentence splitting on normalized paragraphs so line wraps don't split sentences.
	flat := strings.TrimSpace(voiceSpaceRe.ReplaceAllString(plain, " "))
	parts := voiceSentSplitRe.Split(flat, -1)
	sentCount, shortSents, signposts := 0, 0, 0
	for _, p := range parts {
		p = strings.TrimSpace(strings.TrimLeft(p, `"'(`))
		if p == "" {
			continue
		}
		sw := voiceWordRe.FindAllString(p, -1)
		if len(sw) < 2 {
			continue
		}
		sentCount++
		if len(sw) <= 12 {
			shortSents++
		}
		if voiceSignpostRe.MatchString(p) {
			signposts++
		}
	}
	if sentCount == 0 {
		sentCount = 1
	}

	fn := float64(n)
	fs := float64(sentCount)

	m := voiceMetrics{
		Words:        n,
		Sentences:    sentCount,
		LongWordPct:  float64(longW) / fn * 100.0,
		LatinatePct:  float64(latinate) / fn * 100.0,
		ShortSentPct: float64(shortSents) / fs * 100.0,
		SignpostPct:  float64(signposts) / fs * 100.0,
		Triplets100:  float64(triplets) / fn * 100.0,
		Coaching100: (float64(youCnt) +
			3.0*float64(qmarks) +
			2.0*float64(scareQuotes) +
			2.0*float64(intensCnt)) / fn * 100.0,
	}

	m.Z[0] = (m.LongWordPct - voiceBaselines[0].mean) / voiceBaselines[0].scale
	m.Z[1] = (m.LatinatePct - voiceBaselines[1].mean) / voiceBaselines[1].scale
	m.Z[2] = (voiceBaselines[2].mean - m.ShortSentPct) / voiceBaselines[2].scale
	m.Z[3] = (m.SignpostPct - voiceBaselines[3].mean) / voiceBaselines[3].scale
	m.Z[4] = (m.Triplets100 - voiceBaselines[4].mean) / voiceBaselines[4].scale
	m.Z[5] = (m.Coaching100 - voiceBaselines[5].mean) / voiceBaselines[5].scale

	sum := 0.0
	for _, z := range m.Z {
		sum += z
	}
	m.Score = sum / float64(len(m.Z))
	return m
}

func parseVoiceSections(file string, content string, minWords int) []voiceSection {
	lines := strings.Split(content, "\n")
	var out []voiceSection
	curLine := 1
	curHeading := filepath.Base(file)
	var buf []string

	flush := func() {
		if len(buf) == 0 {
			return
		}
		m := measureVoice(strings.Join(buf, "\n"))
		if m.Words >= minWords {
			out = append(out, voiceSection{
				File:    file,
				Line:    curLine,
				Heading: curHeading,
				Metrics: m,
			})
		}
	}

	for i, line := range lines {
		if hm := voiceHeadingRe.FindStringSubmatch(line); hm != nil {
			flush()
			curLine = i + 1
			curHeading = strings.TrimSpace(hm[1])
			buf = nil
		} else {
			buf = append(buf, line)
		}
	}
	flush()
	return out
}

func findVoiceSeams(secs []voiceSection, seamCut float64) []voiceSeam {
	var seams []voiceSeam
	for i := 1; i < len(secs); i++ {
		if secs[i].File != secs[i-1].File {
			continue
		}
		jump := math.Abs(secs[i].Metrics.Score - secs[i-1].Metrics.Score)
		if jump >= seamCut {
			seams = append(seams, voiceSeam{
				File: secs[i].File,
				Jump: jump,
				Prev: secs[i-1],
				Next: secs[i],
			})
		}
	}
	return seams
}

func voiceTopDrivers(m voiceMetrics) string {
	type kv struct {
		label string
		z     float64
	}
	items := []kv{
		{fmt.Sprintf("lw=%.0f%%", m.LongWordPct), m.Z[0]},
		{fmt.Sprintf("lat=%.1f%%", m.LatinatePct), m.Z[1]},
		{fmt.Sprintf("shS=%.0f%%", m.ShortSentPct), m.Z[2]},
		{fmt.Sprintf("sig=%.0f%%", m.SignpostPct), m.Z[3]},
		{fmt.Sprintf("tri=%.2f", m.Triplets100), m.Z[4]},
		{fmt.Sprintf("coach=%.1f", m.Coaching100), m.Z[5]},
	}
	sort.Slice(items, func(i, j int) bool { return items[i].z > items[j].z })
	var parts []string
	for i := 0; i < 3 && i < len(items); i++ {
		if items[i].z > 0.5 || len(parts) < 2 {
			parts = append(parts, items[i].label)
		}
	}
	return strings.Join(parts, " ")
}

func voiceDefaultChapters(root string) ([]string, error) {
	chapters, err := filepath.Glob(filepath.Join(root, "civ", "chapters", "[0-9][0-9]_ch*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(chapters)
	return chapters, nil
}

func runVoice(root string, files []string, opts voiceOptions) error {
	var allSecs []voiceSection
	var allSeams []voiceSeam

	fmt.Println("=== Two-Voice Diagnostic (0.0 = Plain baseline; >= +1.0 = Textbook; pure prose, no citations/formatting) ===")
	fmt.Printf("%-14s %4s %6s %6s %5s %6s %7s | %5s %5s %5s %5s %5s %5s\n",
		"Chapter", "Secs", "Words", "Mean", "SD", "Max", "MaxSeam", "lw%", "lat%", "shS%", "sig%", "tri", "coach")
	fmt.Println(strings.Repeat("-", 96))

	var allScores []float64
	maxScoreOverall := math.Inf(-1)
	maxSeamOverall := 0.0

	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		secs := parseVoiceSections(f, string(b), opts.minWords)
		if len(secs) == 0 {
			continue
		}
		allSecs = append(allSecs, secs...)

		sumScore, maxScore, maxSeam := 0.0, math.Inf(-1), 0.0
		words := 0
		var lw, lat, shS, sig, tri, coach float64
		for i, s := range secs {
			sc := s.Metrics.Score
			allScores = append(allScores, sc)
			sumScore += sc
			if sc > maxScore {
				maxScore = sc
			}
			if sc > maxScoreOverall {
				maxScoreOverall = sc
			}
			if i > 0 {
				j := math.Abs(sc - secs[i-1].Metrics.Score)
				if j > maxSeam {
					maxSeam = j
				}
				if j > maxSeamOverall {
					maxSeamOverall = j
				}
			}
			words += s.Metrics.Words
			lw += s.Metrics.LongWordPct
			lat += s.Metrics.LatinatePct
			shS += s.Metrics.ShortSentPct
			sig += s.Metrics.SignpostPct
			tri += s.Metrics.Triplets100
			coach += s.Metrics.Coaching100
		}
		n := float64(len(secs))
		mean := sumScore / n
		varSq := 0.0
		for _, s := range secs {
			d := s.Metrics.Score - mean
			varSq += d * d
		}
		sd := 0.0
		if len(secs) > 1 {
			sd = math.Sqrt(varSq / (n - 1))
		}
		fmt.Printf("%-14s %4d %6d %+6.2f %5.2f %+6.2f %7.2f | %5.1f %5.1f %5.1f %5.1f %5.2f %5.2f\n",
			filepath.Base(f), len(secs), words, mean, sd, maxScore, maxSeam,
			lw/n, lat/n, shS/n, sig/n, tri/n, coach/n)

		allSeams = append(allSeams, findVoiceSeams(secs, opts.seamCut)...)
	}

	if len(allScores) > 0 {
		mean := 0.0
		for _, s := range allScores {
			mean += s
		}
		mean /= float64(len(allScores))
		varSq := 0.0
		for _, s := range allScores {
			d := s - mean
			varSq += d * d
		}
		sd := 0.0
		if len(allScores) > 1 {
			sd = math.Sqrt(varSq / float64(len(allScores)-1))
		}
		fmt.Printf("\nBook summary: %d sections (>= %d words), mean=%+.2f, SD=%.2f, max=%+.2f, max_seam=%.2f\n",
			len(allScores), opts.minWords, mean, sd, maxScoreOverall, maxSeamOverall)
	}

	var flagged []voiceSection
	for _, s := range allSecs {
		if opts.showAll || s.Metrics.Score >= opts.threshold {
			flagged = append(flagged, s)
		}
	}
	sort.Slice(flagged, func(i, j int) bool {
		return flagged[i].Metrics.Score > flagged[j].Metrics.Score
	})

	label := fmt.Sprintf("Sections at or above %+.2f (Textbook Voice)", opts.threshold)
	if opts.showAll {
		label = "All sections (sorted highest Textbook score first)"
	}
	fmt.Printf("\n=== %s (%d) ===\n", label, len(flagged))
	if len(flagged) == 0 {
		fmt.Println("  (none)")
	}
	for _, s := range flagged {
		h := s.Heading
		if len([]rune(h)) > 48 {
			h = string([]rune(h)[:48]) + "..."
		}
		fmt.Printf("  %+5.2f  %-28s (%4dw)  [%-28s]  %s\n",
			s.Metrics.Score,
			fmt.Sprintf("%s:%d", rel(root, s.File), s.Line),
			s.Metrics.Words,
			voiceTopDrivers(s.Metrics),
			h)
	}

	sort.Slice(allSeams, func(i, j int) bool { return allSeams[i].Jump > allSeams[j].Jump })
	fmt.Printf("\n=== Adjacent Voice Seams >= %.2f (%d) ===\n", opts.seamCut, len(allSeams))
	if len(allSeams) == 0 {
		fmt.Println("  (none)")
	}
	for _, sm := range allSeams {
		h1, h2 := sm.Prev.Heading, sm.Next.Heading
		if len([]rune(h1)) > 30 {
			h1 = string([]rune(h1)[:30]) + "..."
		}
		if len([]rune(h2)) > 30 {
			h2 = string([]rune(h2)[:30]) + "..."
		}
		fmt.Printf("  Δ=%4.2f  %-14s  L%-3d (%+5.2f) %-33s  ->  L%-3d (%+5.2f) %s\n",
			sm.Jump, filepath.Base(sm.File),
			sm.Prev.Line, sm.Prev.Metrics.Score, h1,
			sm.Next.Line, sm.Next.Metrics.Score, h2)
	}

	if opts.failScore > 0 && maxScoreOverall > opts.failScore {
		return fmt.Errorf("voice: max section score %+.2f exceeds -fail-score %.2f", maxScoreOverall, opts.failScore)
	}
	if opts.failSeam > 0 && maxSeamOverall > opts.failSeam {
		return fmt.Errorf("voice: max adjacent seam %.2f exceeds -fail-seam %.2f", maxSeamOverall, opts.failSeam)
	}
	return nil
}
