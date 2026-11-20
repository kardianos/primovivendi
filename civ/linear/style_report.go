package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// writeStyle writes the report sections for close-range echoes, abstract
// subjects, and the share of passive, expletive, and nominalized clauses.
func writeStyle(p func(string, ...any), book, ref []*sentence, chapters []string, byChapter map[string][]*sentence) {
	writeEchoes(p, book, ref)
	writeAbstract(p, book, ref)
	writeVoice(p, book, ref, chapters, byChapter)
}

func countKind(ss []*sentence, kind string) int {
	n := 0
	for _, s := range ss {
		for _, f := range s.Findings {
			if f.Kind == kind {
				n++
			}
		}
	}
	return n
}

func writeEchoes(p func(string, ...any), book, ref []*sentence) {
	p("## 8. Close-range word echoes")
	p("")
	p("A content word that repeats inside one paragraph, where the corpus-wide phrase list does not look: the same lemma twice within 25 words, or in three sentences in a row. Nouns, verbs (other than be, have, do, and the modals), adjectives, and adverbs count. A repeat that stays inside one sentence is kept at half weight and left off the list, because the period sample does the same. A repeat that crosses a sentence, and a word that runs through three sentences, are listed. The count is the number of such repetitions.")
	p("")
	b, r := countKind(book, "close echo"), countKind(ref, "close echo")
	p("| Measure | Book | Period sample |")
	p("|---|---|---|")
	p("| Close-range echoes per 100 sentences | %s | %s |", per100(b, len(book)), per100(r, len(ref)))
	p("")
	type echoItem struct {
		s   *sentence
		f   finding
		run bool
	}
	var items []echoItem
	half := 0
	for _, s := range book {
		for _, f := range s.Findings {
			if f.Kind != "close echo" {
				continue
			}
			if f.Weight < 1 {
				half++
				continue
			}
			items = append(items, echoItem{s: s, f: f, run: strings.Contains(f.Note, "consecutive")})
		}
	}
	p("%d of the book's echoes stay inside one sentence (half weight) and are left off the list below. %d cross a sentence or run through three sentences.", half, len(items))
	p("")
	if len(items) == 0 {
		return
	}
	// Three-sentence runs are the stronger signal. Show them before a
	// two-sentence repeat, and keep chapter order inside each group.
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].run && !items[j].run
	})
	p("Up to 40. A run of three sentences is listed before a two-sentence repeat.")
	p("")
	for i, it := range items {
		if i == 40 {
			break
		}
		p("- **%s.** %s", where(it.s), it.f.Note)
	}
	p("")
}

func writeAbstract(p func(string, ...any), book, ref []*sentence) {
	p("## 9. Abstract subjects")
	p("")
	p("A literal-meaning check, kept out of the Rule 10 score. A common noun rated below %.1f for concreteness, or one of book, chapter, page, account, argument, creed, doctrine, law, rule, system, and text, is the subject of a verb of human will or thought (pretend, seek, ask, want, decide, refuse, assume, claim, forget, choose, ignore, demand, insist, prefer, intend, believe, hope, fear, learn, teach, promise, lie). Names, words for people, and pronouns other than a referential \"it\" are skipped. \"It seeks\" is kept, because that \"it\" stands for a thing.", abstractCut)
	p("")
	b, r := countKind(book, "abstract subject"), countKind(ref, "abstract subject")
	p("| Measure | Book | Period sample |")
	p("|---|---|---|")
	p("| Abstract subjects per 100 sentences | %s | %s |", per100(b, len(book)), per100(r, len(ref)))
	p("")
	shown := 0
	for _, s := range book {
		if shown == 40 {
			break
		}
		for _, f := range s.Findings {
			if f.Kind != "abstract subject" {
				continue
			}
			if shown == 0 {
				p("Up to 40, in chapter order:")
				p("")
			}
			if shown == 40 {
				break
			}
			p("- **%s.** %s", where(s), f.Note)
			p("")
			p("  > %s", s.Text)
			p("")
			shown++
		}
	}
}

type voiceRun struct {
	src *source
	v   []float64 // 1 if the finite clause is stiff
}

func voiceRuns(ss []*sentence) []voiceRun {
	var out []voiceRun
	for _, s := range ss {
		if len(out) == 0 || s.Src != out[len(out)-1].src {
			out = append(out, voiceRun{src: s.Src})
		}
		if !s.Finite {
			continue
		}
		bit := 0.0
		if s.Voice != "active" {
			bit = 1
		}
		out[len(out)-1].v = append(out[len(out)-1].v, bit)
	}
	return out
}

func voiceCounts(ss []*sentence) (n, passive, expl, nom int) {
	for _, s := range ss {
		if !s.Finite {
			continue
		}
		n++
		switch s.Voice {
		case "passive":
			passive++
		case "expletive":
			expl++
		case "nominalization":
			nom++
		}
	}
	return n, passive, expl, nom
}

// stretchMedian is the median stiff share of period stretches with n finite clauses.
func stretchMedian(runs []voiceRun, n int) (float64, int) {
	var shares []float64
	for _, run := range runs {
		v := run.v
		if len(v) < n || n == 0 {
			continue
		}
		pre := make([]float64, len(v)+1)
		for i, x := range v {
			pre[i+1] = pre[i] + x
		}
		for i := 0; i+n <= len(v); i++ {
			shares = append(shares, (pre[i+n]-pre[i])/float64(n))
		}
	}
	if len(shares) == 0 {
		return math.NaN(), 0
	}
	sort.Float64s(shares)
	return shares[len(shares)/2], len(shares)
}

func writeVoice(p func(string, ...any), book, ref []*sentence, chapters []string, byChapter map[string][]*sentence) {
	p("## 10. Passive, expletive, and nominalized clauses")
	p("")
	p("The share of finite main clauses that are passive, expletive (\"there is,\" \"it is clear that\"), or a form of \"be\" whose predicate noun ends in -tion, -sion, -ment, -ity, or -ness. A high share reads as a textbook. Each section is set beside every stretch of the period sample with the same number of finite clauses. The table ranks sections by how far their share sits above that median, so a textbook section leads and a section that is only more active falls to the bottom.")
	p("")
	bn, bp, be, bnom := voiceCounts(book)
	rn, rp, re, rnom := voiceCounts(ref)
	share := func(k, n int) string {
		if n == 0 {
			return "0%"
		}
		return fmt.Sprintf("%.1f%%", 100*float64(k)/float64(n))
	}
	stiff := func(p, e, o, n int) string { return share(p+e+o, n) }
	p("| Measure | Book | Period sample |")
	p("|---|---|---|")
	p("| Finite main clauses | %d | %d |", bn, rn)
	p("| Passive, expletive, or nominalized | %s | %s |", stiff(bp, be, bnom, bn), stiff(rp, re, rnom, rn))
	p("| Passive | %s | %s |", share(bp, bn), share(rp, rn))
	p("| Expletive | %s | %s |", share(be, bn), share(re, rn))
	p("| Copula plus a nominalization | %s | %s |", share(bnom, bn), share(rnom, rn))
	p("")

	runs := voiceRuns(ref)
	type row struct {
		where, title       string
		n                  int
		share, median      float64
		passive, expl, nom int
		windows            int
	}
	var rows []row
	for _, ch := range chapters {
		ss := byChapter[ch]
		for i := 0; i < len(ss); {
			j := i
			for j < len(ss) && ss[j].Sec == ss[i].Sec {
				j++
			}
			n, passive, expl, nom := voiceCounts(ss[i:j])
			if n >= 8 {
				med, windows := stretchMedian(runs, n)
				if windows >= 20 && !math.IsNaN(med) {
					src := ss[i].Src
					title, line := "(opening)", 1
					if k := ss[i].Sec; k >= 0 && k < len(src.SecTitle) {
						title, line = src.SecTitle[k], src.SecLine[k]
					}
					sh := float64(passive+expl+nom) / float64(n)
					label := fmt.Sprintf("%s:%d %s", strings.TrimPrefix(ch, "ch"), line, title)
					rows = append(rows, row{label, title, n, sh, med, passive, expl, nom, windows})
				}
			}
			i = j
		}
	}
	above := 0
	for _, r := range rows {
		if r.share > r.median {
			above++
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		di := rows[i].share - rows[i].median
		dj := rows[j].share - rows[j].median
		if di != dj {
			return di > dj
		}
		return rows[i].share > rows[j].share
	})
	if len(rows) == 0 {
		p("No section has enough finite clauses, and enough period stretches of that length, to compare.")
		p("")
		return
	}
	p("%d of %d compared sections are stiffer than the period median for a stretch of the same length. **Gap** is the section's stiff share minus that median.", above, len(rows))
	p("")
	p("| Section | Finite | Stiff | Passive | Expletive | Nominalized | Period median | Gap |")
	p("|---|---|---|---|---|---|---|---|")
	for i, r := range rows {
		if i == 15 {
			break
		}
		t := r.where
		if rs := []rune(t); len(rs) > 64 {
			t = string(rs[:64]) + "..."
		}
		t = strings.ReplaceAll(t, "|", "/")
		p("| %s | %d | %.0f%% | %d | %d | %d | %.0f%% | %+.0f |", t, r.n, 100*r.share, r.passive, r.expl, r.nom, 100*r.median, 100*(r.share-r.median))
	}
	p("")
}
