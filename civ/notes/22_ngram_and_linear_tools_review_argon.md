# 22. Technical Review of `civ/ngram` and `civ/linear` Tools, Enhancements, and Datasets

**Reviewer:** Argon  
**Date:** 2026-11-16  
**Scope:** `civ/ngram/*.go`, `civ/ngram/reference.tsv`, `civ/linear/*.go`, and `civ/notes/06_editing_guide.md`

---

## 1. Bugs and Blind Spots in the Current Implementation

### A. Both Tools Silently Drop All Blockquoted Parables (`>` lines)
* **Where:** `civ/ngram/text.go:271` and `civ/linear/prepare.go:74`.
* **The Bug:** Both tools strip every line starting with `>` on the assumption that blockquotes are external quotations:
  ```go
  if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, ">") || ...
  ```
* **Impact:** In `civ/chapters/`, nearly all of the author's own narrative set pieces—*The Map Drawn to Please the Patron* (`02_ch02.md:59–61`), *The 1900 BCE Old Assyrian Trade Network* (`05_ch05.md:59–68`), *The Peasant's Cow Stall* (`05_ch05.md:80–87`), *The Lifted Shingle* (`06_ch06.md:142–146`), *The Cold Forge* (`07_ch07.md:75–79`), *The Daughter and the Chalked Footrace* (`07_ch07.md:96–109`), *The Family Creek-Bridge* (`08_ch08.md:54–62`), *The Shepherd's Paddock* (`08_ch08.md:171–177`), *The Moriori* (`08_ch08.md:202–212`), *The Winter Storehouse* (`09_ch09.md:98–102`), *The Seed Granary* (`10_ch10.md:126–130`), *The Plank Bridge* (`11_ch11.md:40–42`), *The Hill Town Grain Storehouse and the Rats* (`11_ch11.md:106–114`), *Two Crews Pulling Down One Wall* (`13_ch13.md:137–143`), and *The Walker in the Marsh* (`14_ch14.md:58–64`)—are formatted as `>` blockquotes. Roughly **3,500+ words of the book's own prose are completely unparsed and unaudited** by both `civ/ngram` and `civ/linear`.
* **Fix:** Only skip `>` blocks that are pure italicized block quotations (`> *"..."*`), or strip the leading `> ` on case-study blockquotes (`> **The ...:**`) so they are analyzed alongside the rest of the book.

### B. Inconsistent `"cannot"` Unigram Handling in `civ/ngram`
* **Where:** `civ/ngram/report.go:171–174` vs. `civ/ngram/section.go:86–91`.
* **The Bug:** Google Books Ngram v3 splits `"cannot"` into `"can"` + `"not"`. In `section.go:88`, `gpm(w)` substitutes `uni["can"]` for `"cannot"`, whereas `report.go:171` looks up `uni["cannot"]` directly (which is near-zero in Google v3) and works around it with a hardcoded `w != "cannot"` check at line 281.
* **Fix:** Unify `gpm(w)` across `report.go` and `section.go`.

### C. Dead Weight-Assignment Branch in `civ/linear`
* **Where:** `civ/linear/analyze.go:359–366`.
* **The Bug:**
  ```go
  fw := s.wordsIn(s.lo[k], s.hi[k])
  if fw < 8 {
      continue
  }
  w := 1.0
  if fw < 8 {
      w = 0.5
  }
  ```
  Because `fw < 8` continues immediately, the `w = 0.5` branch is unreachable dead code (and `if fw >= 8` at line 392 is redundant).

### D. Semicolon/Colon Rule Penalizes Balanced Antithesis (`Part 3: second claim`)
* **Where:** `civ/linear/analyze.go:427–461`.
* **The Bug:** The analyzer flags *every* colon or semicolon followed by a 4+ word clause with a subject as a defect to be split (`Weight: 1, Move: "split"`). In classical 1850–1950 prose—and in the author's best plain-style passages—a semicolon joining two short, parallel, antithetical clauses (*"The rotation was justified by the grain; the spirit account was not"*, *"The arithmetic is exact; the answer is not"*) is a signature rhetorical strength (isocolon/antithesis), not a run-on sentence.
* **Fix:** Exempt `;` clauses from `second claim` when both clauses are short (e.g., `< 12` words each), single-clause, and share parallel dependency/POS structure (`nsubj + verb/cop`).

---

## 2. Mechanical Enhancements to `civ/ngram` and `civ/linear`

### 1. Section-Level & Sliding-Window "Two-Voice Seam" Detector
Right now, `civ/ngram` reports Burrows's Delta only at the whole-chapter level (`civ/ngram/report.go:513–517`). However, the stylistic seams in the book occur **between sections inside chapters** (e.g., Chapter 3's primary-source Al-Ghazali section vs. its textbook Foucault section; Chapter 12's Islam section vs. its bulleted philosopher survey).
* **Enhancement:** Compute stylistic distance at the **`###` section level** (or a 500-word sliding window), and measure each section against **two baselines**:
  1. The 1850–1950 external reference corpus (`reference.tsv`), and
  2. An **"Authorial Gold Standard" internal baseline** composed of the author's hand-reworked sections (`01_ch01` opening, `03_ch03` Ash'arite section, `09_ch09` Value Symmetry, `10_ch10` "Not Ready" section, `12_ch12` Islam section).
* Ranking all ~75 sections of the book by their distance from the Authorial Gold Standard will mechanically pinpoint every remaining "textbook/AI-scaffolded" passage.

### 2. Upgrade Classic Manhattan Delta to Cosine Delta (Würzburg Delta)
* **Where:** `civ/ngram/stats.go:146–152`.
* **Enhancement:** `stats.go` currently uses Burrows's 2002 Classic Manhattan Delta ($\frac{1}{N}\sum |z_1 - z_2|$). In computational stylometry (Eder et al.; Jannidis, Pielström, Schöch, and Vitt 2015), **Cosine Delta** ($1 - \frac{z_1 \cdot z_2}{\|z_1\| \|z_2\|}$ on the top 150–300 MFW z-scores) is markedly more reliable because it prevents a handful of topic/structural z-score spikes (such as `"every"` at $+4.7\sigma$ or `"because"` at $+3.2\sigma$) from swamping the grammatical profile. Adding Cosine Delta alongside Classic Delta takes ~15 lines of Go.

### 3. Paragraph Cadence & Syntactic Template Monotony in `civ/linear`
Currently, `civ/linear` scores each sentence in isolation. Yet a paragraph of five 18-word sentences can score `0.0` on every sentence while reading mechanically because every sentence has the same length and dependency spine.
* **Enhancement:** Add a paragraph-level cadence check in `civ/linear`:
  * **Syntactic Spine Repetition:** Extract the top-level dependency skeleton of each sentence (e.g., `[advcl:when] -> [nsubj] -> [ROOT] -> [obj]`) and flag any paragraph or section where the same top-level clause skeleton repeats 3+ times in a row.
  * **Sentence-Length Flatline:** Flag paragraphs of 5+ sentences where sentence length has abnormally low variance (coefficient of variation $\sigma/\mu < 0.20$), lacking the alternation of medium analytical sentences and short declarative verdicts that characterizes both period prose and the author's plain style.
  * **Didactic Scaffolding Counter:** Flag section-opening enumerative announcements (`"Three concrete..."`, `"Four essential..."`, `"Two opposite..."`) and inline parenthetical principle tags (`(Principle N)` / `(Virtue N)`), which occur dozens of times in the textbook-voiced sections and zero times in the 1850–1950 corpus.

### 4. Typographical & Morphological Register Profile per Section
Add a table to `civ/ngram` (and `section`) measuring four mechanical indicators that immediately separate 1850–1950 treatise prose from modern study-guide prose:
1. **Boldface Density:** Percentage of words wrapped in `**...**` per section (0% in 1850–1950 treatises; >10% in several chapters of `civ`).
2. **List/Bullet Share:** Percentage of words residing inside bulleted or numbered lists per section.
3. **Nominalization-to-Active-Verb Ratio:** Ratio of abstract Latinate nominalizations (`-tion`, `-sion`, `-ment`, `-ism`, `-ity`, `-ness`) to finite active verbs (`VB`, `VBD`, `VBP`, `VBZ` excluding copulas/auxiliaries).
4. **Single-Occurrence Anachronism Check:** Currently `civ/ngram/report.go:228` ignores words used only once (`if n < 2 { continue }`). Lowering the unigram floor to `n >= 1` when `gpm(w) < 0.01` will catch one-off modernisms (`power-warfare`, `risk-insulation`, `error-checkers`, `front-companies`, `hyperinflates`) across the entire book, not just in single-section checks.

### 5. Cross-Chapter Prop & Illustration Deduplicator
`civ/ngram` catches exact n-gram repeats (`thirty-five thousand feet`), but misses when the **same physical illustration** is reused with slightly varied wording across eight chapters.
* **Enhancement:** Group concrete noun lemmas (`NN`/`NNS`) across chapters using TF-IDF or co-occurrence clustering to report **Recurring Concrete Props Across Chapters** (e.g., `{airliner, aircraft, wing, spar, pilot, bolt, hangar}`, `{bridge, beam, oak, timber, span, wagon}`, `{examination, student, honors, mathematics, score}`). This gives the editor a mechanical map of overused metaphors and examples.

### 6. Move `06_editing_guide.md` Part A Regex Checks into `civ/linear` (POS/Dependency-Aware)
The Python regex script in `06_editing_guide.md` flags every literal occurrence of `honest`, `rot`, `raw`, and `load-bearing` because regular expressions cannot tell a person from an inanimate noun.
* **Enhancement:** Check those rules inside `civ/linear/analyze.go`, where POS tags, NER, `personLike(s, i)`, and dependency arcs (`amod`, `nsubj`) are already parsed:
  * Flag `honest` **only** when its head noun is not `personLike` (automatically allowing `an honest merchant` while flagging `honest steel` or `honest accounting`).
  * Flag `rot`/`rotted`/`rotting` **only** when its head/subject is not a biological material (`timber`, `wood`, `grain`, `beam`, `oak`, `rafter`, `tooth`, `flesh`).
  * Flag `load-bearing` **only** when its head is not a physical structural member (`wall`, `column`, `beam`, `pier`, `timber`, `arch`).

---

## 3. Alternative or Upgraded Underlying Tools

### A. Replace CoreNLP's 2014 `englishSR` Shift-Reduce Parser with Stanza / spaCy Transformer UD (`en_core_web_trf`) + `fastcoref`
* **The Current Bottleneck:** `civ/linear/corenlp.go:22` uses `englishSR.ser.gz` (a 2014 shift-reduce constituency parser) and converts constituency trees to basic dependencies. As noted in `18_linearity_review.md` (`01:21`, `02:13`), `englishSR` regularly misparses long fronted adverbial clauses (`If we want...`, `Suppose the timbers...`) as part of a 25-word nominal subject, creating false-positive `long subject` and `late verb` scores. Its 2016 neural coreference model also requires a 16 GB Java heap and still mislinks ~20% of pronouns.
* **Recommended Upgrade:**
  1. **Within CoreNLP (Zero new dependencies):** If staying on Java CoreNLP, run the neural dependency parser (`depparse` with `english_UD.gz`) in a separate fast pass or use the transition-based constituency parser (`englishPCFG.ser.gz` / updated neural models) so `basicDependencies` are predicted directly rather than converted from misattached shift-reduce constituency trees.
  2. **Modern Transformer Pipeline (Highest accuracy):** Replace the Java CoreNLP invocation with a small Python/ONNX CLI script wrapping **Stanford Stanza** (or **spaCy `en_core_web_trf`** RoBERTa-large) + **`fastcoref`** (F-Coref / LingMess) that writes the exact same JSON structure (`cnDoc` in `corenlp.go:31–70`). Transformer-based UD parsing cuts subject/clause attachment errors by more than 60%, runs in ~15 seconds on the whole book, and uses a fraction of the 16 GB RAM required by Java CoreNLP.

### B. Local DuckDB / SQLite Google Books 2-gram & 3-gram Index (Eliminating Web Scraping & Rate Limits)
* **The Current Bottleneck:** `civ/ngram/google.go:398–465` queries `https://books.google.com/ngrams/json` in batches of 10 with a 1.5-second sleep (`-delay 1500ms`), hitting HTTP 429 rate limits and taking ~7–10 minutes on cold runs. Because it uses the web viewer, it also cannot query possessives (`'s`), punctuation-adjacent tokens, or POS-tagged n-grams.
* **Recommended Upgrade:** Just as `google-unigrams` streams the 24 unigram files and filters them to 1850–1950, add a one-time streamer (or DuckDB script) over the Google Books v3 2-gram and 3-gram shards that **filters on the fly to pairs/triples where all words belong to the union of the book's vocabulary and the reference corpus's vocabulary (~28,000 words)**.
* The resulting filtered 1850–1950 2-gram and 3-gram table fits in a single local **SQLite or DuckDB file (~300 MB)** in `ext-ref/ngram/`. Every phrase lookup in `report` and `section` then runs **locally in <50 milliseconds** with zero network calls and no rate limits, and possessives (`man's`, `father's`, `citizen's`) no longer have to be stripped.

---

## 4. Datasets That Could Assist to a Greater Mechanical Degree

### 1. Expand and Stratify `civ/ngram/reference.tsv` (Adding 1750–1850 Foundational Realists & 1925–1955 Engineers)
Currently, `civ/ngram/reference.tsv` stops at 1852 on the early end and 1928 on the late end, and omits the very authors the book cites as its intellectual bedrock. Adding the following public-domain texts from Project Gutenberg / Internet Archive and tagging each work by **sub-register** (`philosophy`, `law-constitution`, `engineering`, `history`) will make Delta and Keyness much sharper:

* **The Book's Explicit Philosophical & Constitutional Models (1759–1848):**
  * **Thomas Reid**, *An Inquiry into the Human Mind on the Principles of Common Sense* (1764) and *Essays on the Active Powers of Man* (1788)
  * **Edmund Burke**, *Reflections on the Revolution in France* (1790)
  * **Adam Smith**, *The Theory of Moral Sentiments* (1759) and *The Wealth of Nations* (1776)
  * **Alexander Hamilton, James Madison, and John Jay**, *The Federalist* (1788, PG `#1404`)
  * **Alexis de Tocqueville**, *Democracy in America* (Henry Reeve translation, 1835/1840, PG `#815` / `#816`)
  * **Thomas Babington Macaulay**, *Critical and Historical Essays* (1843, PG `#2332` / `#2333`) and *The History of England* (1848, PG `#1468`)
* **High-Precision Victorian & Early 20th-Century Legal/Political Prose:**
  * **James Fitzjames Stephen**, *Liberty, Equality, Fraternity* (1873) — arguably the closest stylistic and philosophical ancestor to `civ` in the nineteenth century
  * **A. V. Dicey**, *Introduction to the Study of the Law of the Constitution* (1885)
  * **William James**, *Pragmatism* (1907, PG `#5116`)
  * **G. K. Chesterton**, *What's Wrong with the World* (1910, PG `#1717`)

### 2. Brysbaert Concreteness Ratings (`brysbaert_2014_concreteness.tsv`, 40,000 English Lemmas)
* **Dataset:** Brysbaert, Warriner, & Kuperman (2014), *Behavior Research Methods* — an open 40,000-word TSV rating English lemmas from `1.0` (purely abstract: *epistemology*, *relativism*, *normative*) to `5.0` (purely concrete/tactile: *shingle*, *anvil*, *barley*, *hammer*, *river*).
* **Mechanical Use:** Joining this 1.5 MB TSV into `civ/ngram` and `civ/linear` allows the tools to compute a **Concreteness Score for every sentence, paragraph, and section**, plus the **Abstract-Run Length** (how many consecutive sentences pass without a single concrete noun rated $\ge 3.8$). Whenever a section drops below a mean concreteness threshold or runs 6+ sentences without a physical noun, the tool flags it as drifting into abstract academic prose.

### 3. Historical Semantic Drift & Post-1950 Sense Detection (HistWords / COHA + Selectional Restrictions)
* **The Problem:** Google Ngrams only tells you whether a *string of letters* existed between 1850 and 1950, not whether it was used in its modern sociological sense. Words like *narrative*, *framework*, *construct*, *privilege*, *actor*, *space*, *issue*, *impact*, *dynamic*, and *model* existed in 1880, so Google Unigrams lets them pass—even when they are used with a post-1960 academic meaning.
* **Datasets & Mechanical Use:**
  1. **Stanford HistWords Pre-computed Decade Embeddings (1850–1990):** HistWords provides downloadable historical word-embedding matrices by decade (`sgns` 1850s through 1990s). Words whose nearest neighbors shifted dramatically between 1900 and 1990 (high semantic displacement: e.g., *narrative*, *construct*, *privilege*, *equity*, *identity*, *gender*, *agency*, *toxic*, *trauma*) can be pre-compiled into a **Semantic-Shift Watchlist** so `civ/ngram` flags their modern collocations even though the unigram existed in 1890.
  2. **Open English WordNet + VerbNet Selectional Restrictions:** A compact local TSV of verb selectional restrictions (`[+animate]` subject required vs. `[-animate]` subject) paired with `civ/linear`'s dependency parse can mechanically flag **dead metaphors and unintended personifications** (Rule 2 of `civ/notes/06_editing_guide.md:34–37`), such as when an abstract noun is made the agent of a human cognitive or volitional verb.

---

## Decisions (2026-11-16)

Each item was checked against the code and the data before it was acted on. Reports 17 and 18 were regenerated afterwards.

**Done**

* **1A, block quotations.** Both tools now keep `>` blocks with the prose (3,897 words, almost all of them the book's own examples). A block that is wholly a quotation of another writer still drops out: in `civ/ngram` through the existing quotation filters, and in `civ/linear` when nothing but quotation is left in the block.
* **1B, "cannot."** One shared Google lookup counts "cannot" under "can" in both `report` and `section`. (`section` never loaded "can", so a section that used "cannot" without "can" listed it as rare.)
* **1C, dead branch.** Removed. The behavior is unchanged and matches the table in note 18.
* **1D, semicolons.** A second clause after a semicolon now counts half (0.5); after a colon it still counts 1. Rule 10 names only the colon, and the typography rules join independent clauses with a semicolon. The proposed parallel-structure detector was not built.
* **2.1 and 2.2, Delta by section, and cosine Delta.** Cosine Delta is measured as the distance to the nearest reference work. The mean distance to all works is about 1.00 for every text, so it says nothing. Each section is set beside 500 stretches of the reference works of the same length (note 17, section 7). At section length, Burrows's Delta does not separate the book from the period, so only cosine Delta is listed by section. No "gold standard" baseline was built.
* **2.3, paragraph texture.** Built in `civ/linear` and measured against the period sample (note 18, section 6). The book's paragraphs vary in sentence length *more* than the period's, not less, and frame repetition is about the period's rate (12.6% against 9.8%). Neither result supports the review's premise; both are kept as measures. The "(Principle N)" tags are counted per section in note 17. The enumerative-announcement counter was not built.
* **2.4, register.** Bold and list shares per section are in note 17, section 7. Words used once are in note 17, section 2b, measured against the words the period writers used once rather than against a fixed floor. The five examples in the review were already reported; the new table adds about 90 more. The nominalization ratio was not built.
* **3A, misparses.** Rather than replacing the parser, `civ/linear` makes no long-subject finding when the subject begins with "if," "when," "because," or a like word.
* **3B, phrase index.** No local 2-gram or 3-gram index: about 370 GB and 4.8 TB of Google files. The phrase cache stays, and `go run . lookup` searches it (`-contains` lists cached phrases that hold a word).
* **4.2, concreteness.** Brysbaert ratings, fetched by `go run . fetch-concreteness` to `ext-ref/concreteness/` (note 18, section 7). The book's nouns average 3.41 against the period sample's 3.32. Runs without a concrete noun are shorter in the book than in the period.

**Not done:** 2.4.3 (nominalization), 2.5 (prop clustering; keyness already shows chapter spread), 2.6 (only about 20 lines are hit, and all of them comply), 3A as written (a new parser stack), 4.1 (pre-1850 works fall outside the period compared), 4.3 (HistWords and VerbNet).

**Found along the way in `civ/linear`:** for the period sample, the first-sentence-of-paragraph test was inverted, so Part 1 ("new before old") was checked on the wrong sentences. Separately, CoreNLP returns wrong token indices for about one sentence in eight, and the tool had been dropping those sentences: 301 of the book's and 132 of the period sample's. Both are fixed, and the figures in note 18 now cover every sentence.
