# 29. Spec: Mechanical Writing-Style Assistants for `civ`

Recorded 2026-11-19. Brief implementation specification for extending the existing mechanical editing tools in `civ/` (`civ/linear`, offline linting/consistency checks, and local TTS readback). Hand this file to an implementation agent.

## Guardrails

- All project code lives under `civ/`. Downloaded datasets, models, and caches go in `ext-ref/` (uncommitted), matching `civ/ngram` and `civ/linear`.
- Do not edit `civ/chapters/*.md` or `civ/ccp_sheet.md`, and do not run `civ/build.fish` (it publishes).
- Treat every new flag as an editing signal, not an absolute rule. Where `civ/linear` reports section or book rates, compare them against the existing 1859 to 1911 period sample already cached in `ext-ref/linear/`.
- Update `civ/notes/06_editing_guide.md` (Section 5) with usage instructions once implemented.

---

## Track 1: `civ/linear` Additions (Go, Reusing Cached CoreNLP Parses & Brysbaert Ratings)

Existing code to build on:
- `civ/linear/analyze.go` (`analyzeDoc`, `sentInfo`, `finding`, `personLike`, `dummyIt`, `contentLemma`)
- `civ/linear/concrete.go` (`loadConcreteness`, reads `ext-ref/concreteness/brysbaert_2014.tsv`)
- `civ/linear/texture.go` (paragraph-level texture and period-sample benchmarking)
- `civ/linear/prepare.go` (`chapterSource`)
- `civ/linear/corenlp.go` (`cmdParse`, `loadParsed`)
- `civ/linear/lines.go` (`cmdLines`) and `civ/linear/main.go`

### 1.1 Close-Range Word-Echo Detector
- **Goal:** Catch unintentional local repetition of content words that `civ/ngram` (corpus-wide) misses (for example: `"Chapter 9 sets out..."` followed by `"does not set out..."`; `"follow this argument"` followed by `"what follows"`; `"creed"` in three consecutive sentences).
- **Mechanics:**
  - Scan content lemmas (`NN`, `NNS`, `VB*` excluding auxiliaries/copulas `be`/`have`/`do`/`can`/`may`/`must`/`will`/`shall`, `JJ*`, `RB*`) inside each paragraph.
  - Flag when the same content lemma appears:
    1. twice within a 25-word window inside the same paragraph (excluding identical repeated terms inside an explicit contrast/parallel list if practical, or simply flagging for human judgment with low weight, e.g. `0.5`), or
    2. in 3 or more consecutive sentences of the same paragraph.
  - Maintain a tiny configurable ignore list for unavoidable structural words if noise is high, and report the echo rate per 100 sentences beside the period sample in `civ/notes/18_linearity_review.md`.

### 1.2 `"it"`-Chain & Competing-Antecedent Check (Rule 10, Part 4)
- **Goal:** Enforce the updated Rule 10 pronoun rule in `civ/notes/06_editing_guide.md`: *"Do not chain 'it' across consecutive sentences, and re-name the subject when several nouns sit between it and the pronoun."*
- **Mechanics (in `civ/linear/analyze.go`, Part 4 block):**
  - Skip `dummyIt(s, i)` and quoted tokens, just as existing Part 4 does.
  - **Consecutive `"it"` chain:** Flag when two consecutive sentences in the same paragraph both use non-dummy `"it"` as their opening token or main subject (`nsubj` / `nsubj:pass` of `s.root`). Example that must fire: `"Yet it is not a theology. It does not seek to diminish any faith."`
  - **Competing singular nouns:** When a non-dummy `"it"` refers back to a noun in the previous sentence (via CoreNLP coref chain or nearest singular non-person noun), count how many distinct singular inanimate nouns (`NN`, not `personLike`) lie between that antecedent and `"it"`. Flag when 2 or more intervening singular nouns sit between them (`Kind: "ambiguous it"`, `Move: "name it"`).

### 1.3 Abstract-Subject / Personification Check (Rule 2: Literal Meaning)
- **Goal:** Catch inanimate or abstract nouns paired with human volitional/cognitive verbs (for example: `"That account pretends"`, `"This book seeks"`, `"A creed keeps a civilization"`).
- **Mechanics:**
  - Pass the Brysbaert concreteness map (`loadConcreteness(data)`) into `analyzeDoc` (or a post-pass in `analyze.go` / `texture.go`).
  - For each clause subject (`nsubj`) whose head is a finite verb:
    - Skip pronouns (`PRP`, `DT`, `WP`), proper nouns (`NNP*`), and `personLike` nouns.
    - Look up the subject head lemma's Brysbaert concreteness rating.
    - Flag when concreteness is `< 2.5` (or subject lemma is in a small explicit inanimate-artifact list like `book`, `chapter`, `page`, `account`, `argument`, `creed`, `doctrine`, `law`, `rule`, `system`, `text`) AND the governing verb lemma is in a curated list of human volitional/cognitive verbs (`pretend`, `seek`, `ask`, `want`, `decide`, `refuse`, `assume`, `claim`, `forget`, `choose`, `ignore`, `demand`, `insist`, `prefer`, `intend`, `believe`, `hope`, `fear`, `learn`, `teach`, `promise`, `lie`).
  - Report as a separate Rule 2 / literal-meaning finding (or a dedicated section of `18_linearity_review.md`) so it does not distort the existing Rule 10 score baseline, and benchmark its rate against the period sample.

### 1.4 Active-Verb vs. Copula/Passive/Expletive Share (Per Section)
- **Goal:** Surface stiff, textbook-voiced sections mechanically.
- **Mechanics:**
  - For each section (and matched stretches of the period sample), compute:
    - Share of finite main clauses that are passive (`nsubj:pass` / `aux:pass`), expletive (`expl`, e.g., `"There is..."`, `"It is ... that"`), or copula-plus-nominalization (copula `be` whose predicate noun ends in `-tion`, `-sion`, `-ment`, `-ity`, `-ness`).
  - Add a compact table to `civ/notes/18_linearity_review.md` ranking sections furthest from the period sample.

### 1.5 Single-File Draft Checker (`go run . check`)
- **Goal:** Run all `civ/linear` checks on an arbitrary Markdown scratch file (such as `civ/notes/28_intro_draft1_rewrite.md`) without adding it to `civ/chapters/` or reparsing the whole book.
- **CLI:**
  ```bash
  cd civ/linear
  go run . check -file ../notes/28_intro_draft1_rewrite.md [-from 1] [-to 0]
  ```
- **Mechanics:**
  - Load the file via `chapterSource(path)`.
  - Hash `src.Text` (SHA-256 prefix) and cache CoreNLP input/output under `ext-ref/linear/scratch/<hash>.txt` and `<hash>.txt.json` so repeated runs on unchanged text are instant.
  - Run CoreNLP on that single file only if uncached (reuse `corenlpArgs`, smaller heap flag `-mem 6g` default for single-file runs).
  - Output sentence-by-sentence findings in the same format as `cmdLines`, plus summary lines for close-range echoes, `"it"`-chains, abstract subjects, and concreteness.

---

## Track 2: Lightweight Prose Linter, Spelling Consistency, & Phonetic Clutter

### 2.1 Vale Style Pack (or Standalone Fast Linter)
- **Goal:** Replace manual copy-pasting of the Python regex audit in `civ/notes/06_editing_guide.md` Part A with a single fast command that works on `civ/chapters/*.md`, `civ/ccp_sheet.md`, and `civ/notes/*.md`.
- **Implementation:**
  - If `vale` is available (or installable as a local static binary), create `civ/.vale.ini` and `civ/.vale/styles/Civ/` with YAML rules covering:
    1. All Part A rules from `06_editing_guide.md` (em-dashes, double hyphens, `!`, math/symbols, `"exact same"`/`"every single"`, `"zero"`, imperative openers, `"on earth"`, `"raw"`/`"real-world"`/`"rot"`, `"honest"`, `"load-bearing"`, `"Why? Because..."`, question-then-fragment).
    2. **US vs. UK spelling consistency:** Standardize on American English (matching `civ/chapters/01_ch01.md`: `neighbor`, `honor`, `labor`, `favor`, `defense`, `civilization`) and flag British variants (`-our`, `-ise`, `-re`, `programme`, `judgement`, `practise` verb) outside direct quotations.
    3. **Defined-term capitalization & locked-line check:** Flag inconsistent casing of `Principle N` / `Virtue N` and verify the verbatim 5 Principles and 7 Virtues wherever listed.
  - If keeping zero external binaries is preferred, put an equivalent fast Go or Python CLI under `civ/lint/` that runs Parts A, C, D, spelling consistency, and locked-line verification in one command.

### 2.2 Suffix-Jingle & Phonetic-Clutter Check
- **Goal:** Flag heavy suffix clustering and unintentional rhymes inside a sentence.
- **Mechanics:**
  - Download the public-domain CMU Pronouncing Dictionary (`cmudict.dict`) to `ext-ref/phonetics/` (uncommitted).
  - Add a small script (or subcommand in `civ/lint/`) that flags:
    - 3+ words in a single sentence sharing a heavy unstressed suffix (`-tion`/`-sion`, `-ment`, `-ity`, `-ly`, `-ness`);
    - 2+ content words in a sentence (within 10 words of each other, not identical lemmas) sharing the same stressed-vowel-to-end rhyme tail in CMUdict.

---

## Track 3: Local Neutral TTS Readback Helper

- **Goal:** Let the author listen to a draft paragraph or chapter range in a neutral voice so ear-reading catches pronoun drift, flat rhythm, and buried subjects.
- **Implementation:**
  - Add a small helper script (e.g., `civ/tools/readback.sh` or `civ/tools/readback.fish`) wrapping local **Piper TTS** (binary/venv + one clear, flat English voice model stored in `ext-ref/tts/`, uncommitted).
  - Input: either `-file <path> [-from N -to M]` or `-chapter ch01 -from N -to M`.
  - Pre-process Markdown to plain prose (strip headings, `**`/`*`, blockquote `>`, links, and tables, matching `civ/linear/prepare.go`) so Piper reads only the prose cleanly.
  - Output: write a `.wav` to `/tmp/` (or `ext-ref/tts/out/`) and optionally play it with `aplay`/`mpv`/`pw-play` if `--play` is passed.

---

## Verification & Acceptance Criteria

1. **`civ/linear` unit tests (`go test ./...` in `civ/linear/`):**
   - Synthetic test cases confirming:
     - `"Yet it is not a theology. It does not seek to diminish any faith."` triggers both the `"it"`-chain check and the abstract-subject check (`seek` with inanimate subject).
     - `"Chapter 9 sets out the rule. This book does not set out to diminish any faith."` triggers the close-range echo check (`set`).
     - `"It is clear that rain falls. It is true that wood rots."` does **not** trigger `"it"`-chain (dummy `"it"` skipped).
     - Literal physical subjects (`"A mason seeks sound stone."`, `"The father decides."`) do **not** trigger the abstract-subject check.
2. **Smoke tests:**
   - Run `go run . check -file ../notes/28_intro_draft1_rewrite.md` and confirm clean, readable output.
   - Run `go run . report` in `civ/linear/` (reusing cached parses; no re-parse needed unless `01_ch01.md` changed, in which case run `go run . parse` first) and inspect the new sections of `civ/notes/18_linearity_review.md` to tune thresholds so false positives stay low.
   - Run the Track 2 linter across `civ/chapters/*.md` and `civ/ccp_sheet.md` and verify zero crashes and sensible findings.
   - Run Track 3 readback on `civ/chapters/01_ch01.md` lines 17 to 19 and confirm a valid `.wav` file is generated.
