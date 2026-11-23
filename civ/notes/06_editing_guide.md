# Civilization (`civ`) Prose & Editing Guide

This guide consolidates the stylistic standards, lexical rules, punctuation constraints, and automated N-gram verification methods used across all documents in `civ/` (`civ/chapters/*.md` and `civ/ccp_sheet.md`).

---

## 1. Authoritative Formulations (Preserve Verbatim)

The following two lists are the foundational axioms and derived virtues of the work. Their exact wording must never be altered when editing surrounding prose:

### The 5 Common Principles of Reality
1. **Reality is real, knowable, and consequential. Truth is what corresponds to reality.**
2. **Humans are self-interested, limited, and particular.**
3. **Goodness must include human life continuing, and that life kept and built.**
4. **What you keep and build is self-anchored, tiered, and outward. Enemies are real.**
5. **Justice is scoped procedural symmetry.**

### The 7 Virtues of Civilization
1. **Align with truth**
2. **Take ownership, which is responsibility**
3. **Seek excellence**
4. **Admire beauty**
5. **Be humble**
6. **Have children**
7. **Form and listen**

---

## 2. Target Prose Register & Literal Meaning

1. **1860s / 1950s Engineering-Analysis Register:**
   - Write in clear, calm, declarative analytical prose modeled on classical nineteenth-century and mid-twentieth-century engineering and constitutional treatises.
   - Avoid modern journalistic slang, debate-stage rhetoric, conversational fillers, and artificial urgency.
2. **Literal Dictionary Definitions Only:**
   - Every word must be used in its strict, literal dictionary meaning.
   - Do not use "cute" or non-literal metaphors drawn from anatomy/digestion (`"chewed up"`, `"stomach for"`, `"blind spot"` when `"oversight"` is meant), animals (`"dog-eat-dog"`, `"canary in the coal mine"`), theater (`"performative"`, `"stagecraft"`), computing (`"hard-coded"`, `"operating system"`, `"bandwidth"`), or colloquial accounting/gambling (`"skin in the game"`, `"cook the books"`, `"write a blank check"`).
   - When an engineering or physical illustration is used (such as a suspension bridge, a steam boiler, or a ship's hull), present it as an explicit, literal physical case or analogy rather than mixing figurative slang into abstract sentences.

---

## 3. Typography, Punctuation, and Notation Rules

1. **No Em-Dashes (`—`):**
   - Never use the em-dash character (`—`) or double-hyphen dash (`--`) in prose.
   - Replace with a colon (`:`) when introducing an explanation or list, a semicolon (`;`) when joining independent clauses, commas or parentheses when setting off a parenthetical qualification, or a full stop (`.`).
2. **No Exclamation Marks (`!`):**
   - Never use exclamation marks in prose, headings, or dialogue. Calm declarative statements carry greater weight.
3. **No Mathematical Equations or Symbolic Shorthand in Titles or Prose:**
   - Do not use LaTeX math mode (`$c$`, `$\approx 2.1$`, `$29^\circ\text{F}$`, `$X$`) in headings, tables, or prose. Write all quantities and units out in English words (`the speed of light`, `approximately 2.1`, `twenty-nine degrees Fahrenheit`).
   - Do not use mathematical operators (`+`, `=`, `>`, `<`) as shorthand in headings, bullet titles, or prose (for example, write `Order and Truth Without Vitality Produce Sterility` rather than `Order + Truth without Vitality = Sterility`).
   - Do not use ampersands (`&`) in headings, bullet titles, or tables; always write `and`.
   - Do not use algebraic placeholder letters (`Procedure $X$`, `Doctrine $X$`, `Rich Citizen A`, `"Not Real X"`) in titles or prose; use concrete descriptive nouns (`the established legal procedure`, `authentic doctrine`, `a wealthy citizen`).

---

## 4. Lexical & Syntactic Rules (From Audit & N-Gram Analysis)

### Rule 1: Literal "Honest" Only
- **Standard:** Reserve `honest`, `honesty`, and `honestly` strictly for a human moral agent's deliberate truthfulness, candor, or refusal to lie, cheat, or steal (`an honest witness`, `personal honesty`, `earning a living honestly`).
- **Prohibited:**
  - *Inanimate objects or materials:* Never write `honest steel`, `honest iron`, `honest stone`, `honest bridge`, or `honest weight`. Use `sound steel`, `genuine stone`, `durable bridge`, `accurate weight`.
  - *Analytical processes or documents:* Never write `honest audit`, `honest accounting`, `honest ledger`, `honest test`, or `honest comparison`. Use `accurate audit`, `complete accounting`, `symmetrical comparison`, `verifiable ledger`.
  - *Conversational intensifiers:* Never write `honest mistake`, `honest error`, `honest question`, or `honest disagreement`. Use `unintentional error`, `good-faith question`, `sincere disagreement`.

### Rule 2: Literal "Load-Bearing" Only
- **Standard:** Reserve `load-bearing` strictly for literal structural engineering and architecture where a physical member supports physical weight (`load-bearing stone wall`, `load-bearing columns`).
- **Prohibited:** Never apply `load-bearing` figuratively to words, concepts, virtues, or social institutions (`load-bearing word`, `load-bearing premise`, `load-bearing wall of the republic`). Use `foundational`, `essential`, `indispensable`, `structural`, or `governing`.

### Rule 3: No Colloquial Pleonasms ("Exact Same", "Every Single")
- **Prohibited:** `"exact same"` and `"every single"`.
- **Replacement:** Replace `"the exact same"` with `"the identical"` or `"the same"`. Replace `"every single"` with `"every"` or `"each"`.

### Rule 4: Numerical "Zero" Only
- **Standard:** Reserve `"zero"` for literal numerical values, temperatures, or mathematical limits (`fifty degrees below zero Celsius`, `reduced to zero`).
- **Prohibited:** Never use `"zero"` as a rhetorical intensifier for `"no"` or `"none"` (`"zero tolerance"`, `"zero accountability"`, `"zero evidence"`, `"zero dollars"`, `"Year Zero"`).
- **Replacement:** Use `"no tolerance"`, `"no accountability"`, `"no evidence"`, `"nothing"`, `"from the beginning"`.

### Rule 5: No Imperative Reader-Directing Openers
- **Prohibited:** Do not begin sentences or clauses by pointing at the reader with imperative verbs:
  - `"Look at..."` / `"Look closely at..."`
  - `"Notice how..."` / `"Notice that..."` / `"Notice what..."`
  - `"Watch what happens..."`
  - `"Observe how..."` / `"Observe that..."`
  - `"Examine..."` / `"Consider..."`
  - `"Look back at..."` / `"Step back and..."`
  - `"Imagine..."` / `"Picture..."`
- **Replacement:** State the subject and the analytical observation directly in declarative form (for example, replace `"Look at what happens when a bridge corrodes"` with `"When a bridge corrodes..."`; replace `"Notice the double standard here"` with `"This argument applies a double standard"`).

### Rule 6: No Hyperbolic Scope Tags ("On Earth")
- **Prohibited:** Never use `"on earth"` or `"on Earth"` as a rhetorical intensifier (`"every court on earth"`, `"no power on earth"`, `"the wealthiest nation on earth"`).
- **Replacement:** Omit the phrase (`"every court"`, `"no power"`) or use `"in the world"` (`"the wealthiest nation in the world"`). Use `"on the Earth's surface"` only when contrasting terrestrial physics with celestial mechanics.

### Rule 7: Literal "Raw", "Real-World", and "Rot" Only
- **"Raw":** Reserve `"raw"` for unprocessed physical materials (`raw iron ore`, `raw timber`, `raw sewage`, `raw clay`). Replace figurative uses (`raw power`, `raw assertion`, `raw numbers`, `raw coercion`) with `"brute force"`, `"unvarnished assertion"`, `"unadjusted figures"`, `"naked coercion"`.
- **"Real-World":** Avoid the hyphenated compound `"real-world"` (`real-world consequences`, `real-world engineering`). Use `"physical"`, `"practical"`, `"actual"`, or `"historical"`.
- **"Rot":** Reserve `"rot"`, `"rotted"`, and `"rotting"` for literal biological decay of organic matter (`rotted timber`, `rotting grain`). Replace institutional, moral, or civic `"rot"` with `"decay"`, `"deterioration"`, `"degradation"`, or `"corruption"`.

### Rule 8: No Rhetorical Question-Fragment Transitions ("Why? Because...")
- **Prohibited:** Never use standalone `"Why? Because..."` transitions or inline `"Why does...? Because..."` question-and-fragment pairs.
- **Also prohibited:** A question answered by a one-word fragment (`"...? No."`, `"...? Never."`) or a stock prompt such as `"What happened?"` / `"What occurred...?"` followed by the answer.
- **Permitted (author's decision, 2026-10-27):** A question answered by a word may stay when it forms a deliberate contrast (two questions with opposite answers) or completes a series of parallel questions. The audit still flags these for review. Kept cases: `02_ch02.md` (practice versus explanation), `07_ch07.md` (fault versus responsibility), `09_ch09.md` (the international-law series), and `11_ch11.md` (Step 5).
- **Replacement:** Combine the cause and effect into a single declarative sentence (`"This occurs because..."`, `"A bridge stands only because..."`) or answer a section's framing question with a complete independent sentence rather than a `"Because..."` fragment.

### Rule 9: Substantive Titles and Headings
- Avoid repetitive colon formulas in chapter and section titles (`Chapter N: Topic: Subtitle`) and avoid abstract conversational pairings (`Why X Needs Y`). Give every chapter and section a distinct, substantive title that states its specific subject matter.

### Rule 10: Linear Sentences
- Each sentence reads forward. It starts from something the reader already has and ends on the new point, and the next sentence starts from that point.
- Put the subject and main verb early. Do not hold the verb behind a long opening clause or a list.
- One claim per sentence. Split a sentence that carries a parenthetical aside, a nested clause, or a second claim after a colon.
- A pronoun ("he," "it," "this") refers to something in the same or the previous sentence. Otherwise, name it. Do not chain "it" across consecutive sentences, and re-name the subject when several nouns sit between it and the pronoun.
- Follow the order of the events or the argument: cause before effect, rule before case.
- Open directly on the claim or contrast rather than with a generic wind-up sentence, and state an opposing view concretely rather than in abstract paraphrase.

*Before:* "From conception onward he belongs to the first of the two horizons named at the start of this section: he is one of the living human beings who must stay alive today, and not only a member of the next generation."

*After:* "From conception the child belongs to the present, the first of the two horizons named at the start of this section, and not only to the next generation."

---

## 5. Automated N-Gram & Pattern Verification Method

Before finalizing any document in `civ/`, run the checks below from `/home/d/code/po`. Build all helper binaries once with `fish civ/tools/build-bin.fish` (writes gitignored binaries to `civ/bin/{lint,linear,ngram,review,readback}`). Parts A, C, and D are one command. Part B checks stock phrases. Part F scores sentence shape. A finding is a place to read again.

### Part A: Deterministic Rule, Spelling, and Locked-Line Audit

`civ/bin/lint` scans `civ/chapters/*.md` and `civ/ccp_sheet.md`. It applies the pattern list below, flags a double hyphen (a line that is only dashes, or a table rule, is left alone), flags British spellings outside quotation marks, and checks the capitalization and the verbatim wording of the Principles and Virtues in Section 1. American spelling is the standard of the chapters: neighbor, honor, labor, favor, defense, civilization. The British forms it flags are `-our` and its inflections, `-ise` / `-isation` where American uses `-ize` (promise, precise, exercise, and the other shared `-ise` words are kept), `-re` (centre, metre, theatre, and the rest of that short list), programme, judgement, and the verb practise. `Principle N` and `Virtue N` keep those capitals. A numbered list of the five Principles or the seven Virtues, and a block quotation that states one of them (`> **Principle N of Reality:** ...`, `> **Virtue N:** ...`), must match Section 1. A virtue line may end with a period. Pass `-notes` to include `civ/notes/*.md`; the notes quote the forbidden patterns in order to describe them, so they are not the default. Extra arguments are file paths.

The patterns are: an em dash or an exclamation mark; math or symbols (`$...$`, a LaTeX command, a spaced `+ = ≈ × ÷ &`, or `A = B`); "exact same" and "every single"; "zero"; an imperative opener (Look at, Look closely, Look back, Step back, Notice, Watch, Observe, Examine, Consider, Imagine, Picture) at the start of a line or a sentence; "on earth"; "raw", "real-world", "rot", "rotted", "rotting", "rots"; "honest" and words built on it; "load-bearing"; "Why ...? Because"; and a question followed by No, Yes, Never, or Nothing, or "What happened / occurred ...?".

```bash
civ/bin/lint            # this part, plus Parts C and D
civ/bin/lint prose      # this part only
civ/bin/lint -notes     # also civ/notes/*.md
```

### Part B: N-Gram & Sentence-Opener Frequency Scanner

Run the N-gram scanner (1-grams to 5-grams and 3-word sentence openers) across the corpus to detect any newly introduced repetitive stock phrases or rhetorical tics:

```bash
python3 -c '
import glob, re, collections

files = sorted(glob.glob("civ/chapters/*.md")) + [
    "civ/ccp_sheet.md",
]

text = "\n".join(open(p).read() for p in files)
clean = re.sub(r"[#*>`_\[\]\(\)]", " ", text)
words = re.findall(r"\b[a-zA-Z\x27-]+\b", clean.lower())

for n in (3, 4, 5):
    counts = collections.Counter(tuple(words[i:i+n]) for i in range(len(words)-n+1))
    print(f"=== TOP {n}-GRAMS ===")
    for ng, c in counts.most_common(25):
        if c >= 8:
            phrase = " ".join(ng)
            print(f"  {c:4d}: {phrase}")

sents = re.split(r"(?<=[.!?])\s+|\n+", clean)
openers = collections.Counter()
for s in sents:
    w = re.findall(r"\b[a-zA-Z\x27-]+\b", s.strip())
    if len(w) >= 3:
        openers[" ".join(w[:3])] += 1

print("=== TOP 3-WORD SENTENCE OPENERS ===")
for op, c in openers.most_common(25):
    if c >= 6:
        print(f"  {c:4d}: {op}")
'
```

### Part C: Structural Markdown Check (Lists That Will Run Into a Paragraph)

The build uses pandoc's default Markdown reader. A list that follows a paragraph line with no blank line between them is printed as run-in text inside that paragraph, in the PDF, HTML, and EPUB alike. The same holds inside a blockquote, where the separating line must be a bare `>`. `civ/lint` finds every such list in `civ/chapters/*.md` and `civ/ccp_sheet.md`. `go run .` runs this with Parts A and D.

```bash
civ/bin/lint structure
```

Fix each hit by inserting a blank line (or a bare `>` line inside a blockquote) before the list.

### Part D: Rendered-Output Check

Render each source with pandoc and look for a list that ended up inside a paragraph. `civ/bin/lint` renders into memory only, so it does not run `build.fish` and does not publish. After a build, the same pattern (`:\s+(1\.|\*) `) can be grepped in `civ/build/*.html`.

```bash
civ/bin/lint render
```

### Part E: Period Comparison (Google Books and a Reference Corpus)

`civ/bin/ngram` compares the chapters with English prose of 1850 to 1950 and writes `civ/notes/17_ngram_comparison.md`. The report lists words and phrases that were rare or absent in Google Books for those years (a word used once is measured against the words the period writers used once), the keyness of words and phrases against a corpus of 36 public-domain treatises (`civ/ngram/reference.tsv`), Burrows's Delta and cosine Delta, sentence statistics, and a table by section: distance from the nearest period writer, judged against stretches of the reference works of the same length, and the shares of bold type, lists, and "(Principle N)" cross-references. Block quotations are counted with the prose, because the book sets its own examples in them; quotations of other writers are left out. Downloads are kept in `ext-ref/ngram/`, which is not committed.

```bash
civ/bin/ngram fetch-ref         # once: the reference works
civ/bin/ngram google-unigrams   # once: Google's unigram counts (about 10 GB streamed, about 20 minutes)
civ/bin/ngram google-phrases    # after edits: looks up new recurring phrases only; earlier lookups are cached
civ/bin/ngram report
civ/bin/ngram lookup "cannot be" honest   # one phrase or word: the cached rate, looked up if new
civ/bin/ngram lookup -contains bridge     # every cached phrase that holds a word
```

The results are signals for a reader, not rules. A phrase absent from the period is a reason to read the phrase again, not proof that it is wrong. The book's subject dominates the keyness lists, so the function-word tables and the sentence tables say the most about register.

### Part F: Rule 10 Aid (Linear Sentences)

`civ/bin/linear` parses the chapters with Stanford CoreNLP (unpacked in `ext-ref/corenlp/`; Java is required) and scores each sentence against the five parts of Rule 10: words before the main verb, long subjects and opening phrases, a second clause after a colon or semicolon, parenthetical asides, nested clauses, pronouns with no noun of their kind nearby, bare "this," new material before old, and a cause stated after its effect. Part 4 also flags two consecutive sentences in one paragraph that each open on, or take as their subject, a real "it" (a dummy "it", as in "it is clear that", is skipped), and an "it" with two or more singular inanimate nouns between it and the noun it points back to. It compares the book with a sample of period prose parsed the same way, and it drafts a rewrite where the move is mechanical: pull back a long opening phrase, pull forward a short cause, split at a colon or semicolon, move out an aside, or name a pronoun. It also measures paragraph texture (paragraphs whose sentences are all of one length, and runs of sentences built on the same frame) and the concreteness of the nouns (Brysbaert concreteness ratings, by section, and runs of sentences with no concrete noun), each set against stretches of the period sample of the same size.

Three further counts are printed in the same report and are not added to the Rule 10 score. A close-range echo is a content word used twice inside a 25-word window of one paragraph, or in three consecutive sentences of that paragraph. A repeat that stays inside one sentence is kept at half weight and left off the printed list. An abstract subject is an inanimate or abstract noun (or "it" standing for one) as the subject of a verb of will or thought, such as seek, decide, or pretend; a person, and a concrete noun that is not on the short list of textual nouns (book, chapter, account, creed, and the like), is left alone. A stiff clause is a finite main clause that is passive, an expletive ("there is", "it is ... that"), or a copula whose predicate is a nominalization (`-tion`, `-sion`, `-ment`, `-ity`, `-ness`). Each rate is set beside the period sample. The report writes `civ/notes/18_linearity_review.md`. Do not edit that file by hand; `civ/bin/linear report` rewrites it.

`check` runs the same scoring on one Markdown file and does not add it to the book. The parse is cached under `ext-ref/linear/scratch/` by a hash of the plain text, and the default heap for that one file is 6g.

```bash
civ/bin/linear fetch-concreteness   # once: the concreteness ratings, to ext-ref/concreteness/
civ/bin/linear parse                # parses new or changed chapters only (a few minutes for the whole book)
civ/bin/linear report
civ/bin/linear check -file civ/notes/28_intro_draft1_rewrite.md
civ/bin/linear check -file civ/chapters/01_ch01.md -from 17 -to 19
```

The drafts are offered for the editor to judge. The parser misreads some sentences, and a moved phrase can need a new connective.

### Part G: Suffix and Rhyme Clutter

`civ/bin/lint` can also listen for a pile of heavy suffixes and for an accidental rhyme. On first use it downloads the CMU pronouncing dictionary to `ext-ref/phonetics/`, which is not committed. A sentence is listed when three or more of its words share one of the unstressed suffixes `-tion` / `-sion`, `-ment`, `-ity`, `-ly`, or `-ness`, and when two content words within ten words of each other, and not the same lemma or a prefixed contrast pair, share the dictionary's stressed-vowel-to-end tail.

```bash
civ/bin/lint phonetics
civ/bin/lint all          # Parts A, C, D, then the phonetic pass
```

### Part H: Neutral Read-Back

`civ/bin/readback` (symlink to `civ/tools/readback.fish`) speaks a line range with Piper, using the local `en_US-ryan-high` voice kept in `ext-ref/tts/voices/` by default (override with `-voice NAME`, e.g. `en_US-lessac-medium`). It strips headings, emphasis, links, and tables the way `civ/linear` prepares a chapter, writes a `.wav` under `ext-ref/tts/out/`, and plays it when `--play` is given (`aplay`, `pw-play`, or `mpv`).

```bash
civ/bin/readback -chapter ch01 -from 17 -to 19
civ/bin/readback -file civ/notes/28_intro_draft1_rewrite.md --play
```

### Part I: Two-Voice Diagnostic (Plain Style vs. Textbook Voice)

`civ/bin/lint voice` scores each chapter section (`>= 150` words by default) on six pure-prose dimensions anchored to fixed Plain-Style baselines (`0.0` = Plain baseline; `>= +1.0` = Textbook drift; adjacent within-chapter jump `>= 1.2` = sharp voice seam): long words (`>= 8` letters), Latinate suffixes, lack of short sentences (`<= 12` words), didactic signpost openers, parallel triplets (`X, Y, and Z`), and coaching markers (`you/your`, rhetorical `?`, scare quotes, intensifier `-ly` adverbs). It deliberately excludes primary-source citation density and Markdown formatting so it measures prose voice itself.

```bash
civ/bin/lint voice                         # all chapters, sections >= +1.0 and seams >= 1.2
civ/bin/lint voice -threshold 1.35         # focus on the sharpest textbook sections
civ/bin/lint voice -fail-score 1.5 -fail-seam 1.5   # optional ratchet gate
```

### Part J: List-Block Audit and Review-Packet Pre-Flight

`civ/bin/lint lists` scans bulleted and numbered blocks across chapters and flags blocks that read like prose wearing bullets (multi-paragraph indented lists, 2-item pairs averaging `>= 35` words, or blocks averaging `>= 55` words per item). `civ/bin/lint sim` applies a review packet JSON in memory (without touching chapter files) to verify unique quote matches, zero overlapping byte spans, zero Part A / spelling / locked-line regressions, and before/after `voice` scores.

```bash
civ/bin/lint lists                         # prose-like list blocks across chapters
civ/bin/lint lists -all                    # full inventory of every list block
civ/bin/lint sim civ/notes/review/packets/<packet>.json
```

### Part K: Cadence and Voice Presence (Rhythm, Word Weight, and Punch)

`civ/bin/lint cadence` measures positive voice presence at the paragraph level in pure Go (~0.4 seconds across all 14 chapters) using four local lexicons kept in `ext-ref/` (not committed): the CMU Pronouncing Dictionary (`ext-ref/phonetics/cmudict.dict`), Brysbaert concreteness (`ext-ref/concreteness/brysbaert_2014.tsv`), Kuperman Age of Acquisition with SUBTLEX dominant POS and lemmas (`ext-ref/lexicon/kuperman_aoa.tsv`), and Princeton WordNet 3.1 sense counts (`ext-ref/lexicon/wordnet_senses.tsv`). Each paragraph receives an anchored `Snap` score (`0–100`, higher = stronger voice) composed of three sub-scores: **Rhythm** (primary stress-beat density, 4+ weak-syllable troughs, masculine stressed landings, sentence-length variation and setup-to-punch contrast), **Weight** (syllables per content word, concreteness, early Age of Acquisition, WordNet polysemy, kinetic verb ratio), and **Punch** (balanced clause pivots, closing snap, and absence of abstract/Latinate prepositional mud chains).

```bash
civ/bin/lint fetch-lexicons                # once: downloads cmudict, Kuperman AoA, and WordNet 3.1 senses
civ/bin/lint cadence                       # chapter summary + 20 muddiest and 10 snappiest paragraphs (0.4 s)
civ/bin/lint cadence -text "..."           # score a single paragraph or draft rewrite
civ/bin/lint cadence -json                 # emit per-paragraph metrics as JSON for downstream tools
```
---

## 6. Bottom-Up Editing Workflow

1. Run `civ/bin/lint` (Parts A, C, and D) to obtain exact file paths and line numbers for the pattern, spelling, locked-line, and list checks.
2. Read the surrounding context of each flagged line.
3. **Always edit files from the bottom up** (from the highest line number to the lowest line number within each file) so that earlier line numbers do not shift during multi-region edits.
4. Re-run Parts A and C to verify that no new violations remain, and run Part D to confirm the rendering. Then execute `fish civ/build.fish` to rebuild all targets in `civ/build/`. The script also runs `../push_release`, so building publishes.
