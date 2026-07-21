# Chapter audit instructions (plan / sources adherence)

**Purpose:** Independent audit of manuscript chapters in `chapters/*.md` after Kimi-heavy drafting. Primary worry: constructivist (or adjacent) bends, omitted plan edges, rearranged structure, and concepts that fight the axioms themselves. Secondary: style and clarity only where they hide a content error.

**Author signal (Batch 0 already under human edit):** On `1000_questions.md` and `1100_axioms.md`, human review found (1) new concepts opposed to the principles, (2) odd re-arrangement (e.g. human limits pushed to an afterword instead of living inside the five-answer sequence), (3) plan/source under-adherence in favor of a writer-default constructivist or “rival-first” posture. Treat those two files as the calibration sample of failure modes. Do not re-litigate every human `%%` already present; build on them and finish the rest of the book.

**Output of this process:** Annotated chapters. Not a rewrite pass. Not a publish.

---

## 1. What success looks like

For each assigned chapter, the agent:

1. Reads the **plan and sources first** (quarantine-breaking is allowed for auditors; plan still wins on content).
2. Reads the chapter end to end against that map.
3. Adds column-0 `%%` notes on real issues (and optional `%+` / `%-` only when a short concrete alternative is worth parking).
4. Returns a short **chapter scorecard** (see §7) in the agent report. Does **not** invent a parallel review essay on disk unless the orchestrator asks.

Orchestrator (human or lead agent) then skims scorecards, may launch a fix pass later. This document is audit-only.

---

## 2. Authority order (content)

When sources conflict, use this order:

1. **Author-in-file notes already present** (`%%`, HTML `<!-- ... -->` operator notes in the chapter). Author voice wins over plan if they explicitly re-order (e.g. merge B1+B2; move B7 into the sequence).
2. **`plan/outline_p*.md` + `plan/dictionary_p*.md`** for that arc. Dictionary is the definitional source of truth; outline is order and job.
3. **`sources/logic/logical_sequences_full_book.md`** for argument chains (SEQ-*).
4. **`sources/` philosophy / history / notes** named in `sources/README.md` for that section.
5. **`writer_instructions.md`** + vocab files for prose rules.
6. **Existing manuscript** is on trial. It is not a source of doctrine.

**Explicit ban for auditors:** Do not “improve” the philosophy toward social constructivism, pure consensus truth, blank-slate anthropology, outcome equity as justice, marble-self lottery guilt, anti-natal ranking, or hermetic protection of favored models. Those are named **external packages** (H8 / P2.F), not house doctrine.

**Constructivism is especially load-bearing as an adversary, not a method:**

| Allowed (plan) | Not allowed (drift) |
|----------------|---------------------|
| Form vs substance: labels/ceremonies can change; constraints reappear (`discovered_vs_constructed.md`, P2.A) | “Reality is (mostly) social agreement / power / language” as the book’s own view |
| Critique of social constructivism as a package that erases constraints | Softening B1/B2 into “shared maps we co-create” without world-as-judge |
| Steelman one sentence of a rival before conflict (W8), in package chapters | Leading every axiom with rivals so the positive never stands alone (author: positive vision first on axioms) |
| “Useful prediction ≠ full true ontology” (M2) | “All models equal” or truth as pure usefulness |

---

## 3. Input set (every agent reads)

### Required for every run

| File | Why |
|------|-----|
| `AUDIT_INSTRUCTIONS.md` (this file) | Job, severity, markup, scorecard |
| `writer_instructions.md` | W/E rules, comment convention |
| `.grok/skills/book-markup/SKILL.md` | `%%` / `%+` / `%-` contract |
| `vocab_whitelist.txt`, `vocab_blacklist.txt` | Term discipline |
| `sources/logic/logical_sequences_full_book.md` | End-to-end chains |
| `sources/README.md` | Section → source map |

### Plus the plan span for the run (see §5)

- Part 1 runs: `plan/outline_p1.md`, `plan/dictionary_p1.md`
- Part 2 runs: `plan/outline_p2.md`, `plan/dictionary_p2.md`
- Cross-book packages: both dictionaries as needed for H8 / P2.F names

### Calibration sample (every run, brief)

Skim human markup already on:

- `chapters/1000_questions.md`
- `chapters/1100_axioms.md`

Learn the author’s failure modes (paraphrased Q wording, rival-first axioms, human nature deferred, metaphoric openers, forcing the reader to “admit,” etc.). Apply the same nose elsewhere.

### Do not use as doctrine

Earlier books (`book/`, `book1-1/`, `book1-4/`, `book1-5/`, `outline/`, `release/`) are not phrasing or structure sources for rewriting. Auditors may open them only if chasing a provenance mystery; prefer plan + sources.

---

## 4. Audit checklist (apply to every chapter)

Work top to bottom. Prefer **fewer sharp notes** over a fog of nitpicks. One `%%` per distinct issue, placed **immediately above** the offending paragraph (column 0).

### A. Structure and plan coverage

- [ ] Chapter job matches `<!-- chapter: ... -->` and the outline arc for that file (`chapter_map_exp.md` row).
- [ ] Outline bullets for the arc appear **in order**, or a conscious skip is noted. Missing load-bearing bullets = high severity.
- [ ] Section order matches plan order rules (examples: F0 cushion before F1 slide; H8 → H9 → H10 → alliances → H7; M after axioms before virtues bulk).
- [ ] Dictionary **definitions** (not just titles) survive in prose. Watch for diluted B1–B5, B7, B8, E0 load-bearing line, justice triad, form/substance test.
- [ ] Author rulings already in plan (e.g. B8 plant inside B7; E0 three units; no theorist names in E0) are kept.

### B. Philosophical fidelity (highest priority)

Flag any prose that:

1. **Treats reality or truth as primarily constructed** by language, power, group will, or pure dialectic — unless clearly naming an **external** package being refused.
2. **Erases human limits** (B7): blank slate, natural goodness, perfectibility-by-total-remake presented as the book’s view or left unopposed where the plan requires the limited-human premise.
3. **Marble-self / generic person / lottery guilt** as moral baseline (B8, E0, anti_marble source). The book refuses this.
4. **Outcome equity or feud (value-for-value) as justice** where the plan requires procedural symmetry (B5, P2.C).
5. **Smuggles B3 (life/thriving) into Q3** in the questions chapter, or smuggles answers into open diagnostic questions.
6. **Hermetic logic**: success credited to a package, failure blamed on “not real X,” without the narrow text/practice test (P2.B, hermetic source).
7. **Misuses form/substance**: either “nothing may change” (reactionary freeze) or “everything is costume” (constructivist totalism). Plan requires both errors named.
8. **Inverts priority**: unscoped action-love, tiers-as-bigotry, or outer fashion virtue over near duty (B4).
9. **Confuses useful model with true ontology** (M2) or wrong with incomplete (M3).
10. **Persons = packages** (guilt by jersey) or the reverse dodge that refuses to name packages at all (H8, P2.F).

When in doubt, quote the dictionary line in the `%%` note and say how the paragraph fights it.

### C. Source and empirical provenance

- [ ] Quotations: exact words in `sources/` (E4). Fabricated or “close enough” quotes = high.
- [ ] Strong empirical claims: source, soften, or `<!-- verify: ... -->` (W9/E6).
- [ ] History chapters: grade by thrival under constraints; depth over name-drop; Singapore / Russia / Iran / Madison / Dugin jobs per outline.
- [ ] Do not invent death-toll precision (sources README).

### D. Arrangement and rhetoric (content-affecting only)

- [ ] **Positive first** where the author has ruled it (axioms block: state our answers; do not open every point with a rival parade). Package chapters *do* name externals by design.
- [ ] Human nature (B7) and particular person (B8) not demoted to orphan afterthought if plan/author put them in the sequence.
- [ ] One arc per paragraph (W3). Stuffed multi-arc paragraphs that scramble the plan chain = medium.
- [ ] Pictures: full paint or direct claim (W2). Substitution metaphors that smuggle soft constructivism (“tapestry of meanings,” etc.) = flag.
- [ ] Blacklist hits (`frame`/`mirror` product names, `thrival` coinage, etc.) = low unless they carry false doctrine.

### E. What not to flag (unless extreme)

- Pure Blair/1820s cadence mismatches (W12) without content error — note once per chapter max if the whole file is off.
- Micro-wording polish with no plan stake.
- Re-arguing settled author rewrites already marked with `%+` / `%-` unless the **new** text still fails the plan.

---

## 5. Division into four agent runs

Sizes are approximate line counts (including existing markup). Run order is fixed: foundations first (sets the nose), then Part 1 bulk, then Part 1 close, then Part 2.

| Run | Files | ~Lines | Plan arcs | Why this bucket |
|-----|-------|--------|-----------|-----------------|
| **1** | `1000_questions.md`, `1100_axioms.md`, `1200_testable_truth.md` | ~150 | Q, B, M | Highest doctrinal density; author already mid-edit; method section must not re-smuggle constructivism |
| **2** | `1300_survival_thriving.md`, `1400_personal_virtues.md`, `1500_social_virtues.md` | ~320 | C, D, E0–E12 | Longest Part 1 bulk; E0 primer is anti-marble load-bearing; virtues must stay under B1–B7 |
| **3** | `1600_fall_harm.md`, `1700_fall_standards.md`, `1800_ending_arc.md` | ~290 | F, G, H | Falls + external packages; H8 must name constructivism as **external**, not house method; order H8→H9→H10→alliances→H7 |
| **4** | `2000_part2_open.md` … `2700_practice_close.md` (8 files) | ~500 | P2.0–P2.G | Form/substance, rigged tests, justice triad, compassion misuse, history, packages, practice matrix |

**Skip:** `0000_title_block.md` (YAML only).

**Parallelism:** Runs 2 and 3 may run in parallel **after** Run 1 scorecards are read (so agents inherit calibrated severity). Run 4 after Part 1 runs, or in parallel with Run 3 if the lead has already internalised Run 1 findings. Prefer sequential 1 → 2 → 3 → 4 when one lead agent is doing all four.

**If a single agent is too small for a whole run:** split Run 2 into `1300+1400` and `1500`; split Run 4 into `2000–2300` and `2400–2700`. Keep the same checklist.

---

## 6. Markup rules (audit mode)

From `writer_instructions.md` W11 and `.grok/skills/book-markup/SKILL.md`:

| Marker | Use in this audit |
|--------|-------------------|
| `%%` | **Default.** Freeform issue note, column 0, immediately above the paragraph. Start with severity + tag. |
| `%-` / `%+` | **Optional, not priority.** Only when a short replacement paragraph is clearly better and you want the author to compare. Do not rewrite whole chapters. |
| `<!-- arc: / point: / verify: -->` | Do not invent new arc structure unless an existing comment is wrong relative to plan; prefer `%%` for audit chatter. |

### `%%` line format (required pattern)

```text
%% [severity] [tags] short diagnosis. Plan/source anchor. Fix direction (one line).
```

**Severity:**

| Tag | Meaning |
|-----|---------|
| `CRIT` | Fights axioms or plan spine; would mis-teach the book if published |
| `HIGH` | Missing plan bullet, wrong order with doctrinal effect, fake quote, package misnamed as ours |
| `MED` | Softening, rival-first where positive-first was ruled, muddy chain, weak steelman |
| `LOW` | Style/blacklist/cadence; only if quick |

**Tags (pick 1–3):** `plan-gap` `plan-order` `constructivist` `anthropology` `marble` `justice` `truth` `life-axiom` `tiers` `procedure` `source` `quote` `verify` `rival-first` `metaphor` `vocab` `voice` `author-ruling`

**Examples:**

```text
%% [CRIT] [constructivist] [truth] States truth as group agreement in the book's own voice. Dictionary B2: correspondence to reality; group consensus is a named error. Rewrite as external package or cut.
%% [HIGH] [plan-order] [anthropology] B7 human limits appears only as afterword; outline/author put limits inside the answer sequence (and B8 plant inside B7). Move into sequence; do not leave as coda.
%% [MED] [rival-first] [author-ruling] Opens the axiom with a rival parade; author wants positive statement first, others later. Lead with our claim.
%% [LOW] [metaphor] "on the table" is dead metaphor; say the questions have been stated.
```

If the paragraph is already under active human `%+`/`%-`, add `%%` **above the pair** and address the **publish (`%+`) text** unless the old text is what still confuses.

---

## 7. Per-chapter scorecard (agent report only)

After editing the file(s), the agent’s return message includes one block per chapter:

```text
### chapters/NNNN_name.md
- Arc: [Q/B/M/...]
- Plan coverage: complete | mostly | major gaps (list missing IDs)
- Doctrine risk: none | low | medium | high (one sentence)
- Constructivist / package-bleed notes: n (severities)
- Source/quote issues: n
- Markup added: n × %% ; n × %± pairs
- Top 3 fixes for a later rewrite pass:
  1. ...
  2. ...
  3. ...
```

Optional run rollup: worst chapter, whether Run N is safe to rewrite yet, any cross-chapter inconsistency (e.g. axioms charge list ≠ later E12 floor list).

---

## 8. Agent packet template

Copy for each run. Working directory: `book1-6/` (or project root with `book1-6/` prefixes).

```text
You are AUDITING manuscript chapters for Life Under Axioms (book1-6).
You do not draft a new book. You annotate issues.

Read first, in order:
1. book1-6/AUDIT_INSTRUCTIONS.md  (full; follow severity, markup, scorecard)
2. book1-6/writer_instructions.md
3. book1-6/.grok/skills/book-markup/SKILL.md
4. book1-6/vocab_whitelist.txt and vocab_blacklist.txt
5. book1-6/sources/logic/logical_sequences_full_book.md
6. book1-6/sources/README.md
7. [PLAN FILES FOR THIS RUN]
8. [SOURCE FILES FOR THIS RUN]
9. Skim existing human %% on chapters/1000_questions.md and chapters/1100_axioms.md
10. Then read and annotate only: [CHAPTER LIST]

Authority: plan dictionary + outline + logical sequences over manuscript prose.
Author %% already in a file outranks plan where they conflict.
Do NOT bend the text toward social constructivism, consensus-truth,
blank slate, outcome equity as justice, or marble-self ethics.
Those are external packages to be named and refused, not house method.

Task for each chapter:
- Check structure, doctrine, sources per AUDIT_INSTRUCTIONS §4.
- Insert %% notes (format §6). Optional short %+/%- only if high value.
- Do not delete author prose or existing author %%. You may add notes above them.
- Do not run concat or publish.

Return: scorecards §7 for each file + a 5–10 line run summary.
```

### Run 1 fill-ins

- **PLAN:** `plan/outline_p1.md` (Q, B, M); `plan/dictionary_p1.md` (Q1–Q8, B1–B8, M1–M7)
- **SOURCES:** `philosophy/life_axiom_long_argument.md`, `philosophy/models_astronomy_chain.md`, SEQ-0–3
- **CHAPTERS:** `1000_questions.md`, `1100_axioms.md`, `1200_testable_truth.md`
- **Extra watch:** Q questions must stay full (dictionary Q3 warns against paraphrase — note tension with any author paraphrase already in `%+`); axioms positive-first; B7 in sequence; B8 plant; M sections off divine; wrong vs incomplete; no constructivist method.

### Run 2 fill-ins

- **PLAN:** outline/dictionary C, D1–D15, E0–E12
- **SOURCES:** SEQ-5–6; `philosophy/anti_marble_self_and_rawls.md` (for E0 / marble); life axiom as needed
- **CHAPTERS:** `1300_survival_thriving.md`, `1400_personal_virtues.md`, `1500_social_virtues.md`
- **Extra watch:** E0 load-bearing line and three units; no theorist names in E0; social forms as tools for life; D7 particular person; D15 allow failure; floors in E12 match B1–B5 (or note intentional author merge of B1+B2).

### Run 3 fill-ins

- **PLAN:** outline/dictionary F, G, H1–H10
- **SOURCES:** SEQ-7–9; `philosophy/beauty_standards_mechanisms.md`; `notes/red_green_alliance.md`; package names from H8
- **CHAPTERS:** `1600_fall_harm.md`, `1700_fall_standards.md`, `1800_ending_arc.md`
- **Extra watch:** F0 before F1; reactionary false fix refused; G lower certainty + falsifier; **H order fixed**; social constructivism named as external; two-books / non-settling divine; alliances only after packages; close on construction not villain tour.

### Run 4 fill-ins

- **PLAN:** `outline_p2.md`, `dictionary_p2.md` full
- **SOURCES:** `philosophy/discovered_vs_constructed.md`, `hermetic_argument.md`, `justice_three_kinds.md`, `compassion_misuse_and_false_foundations.md`, `notes/purity_virtue_reality.md`, `notes/red_green_alliance.md`; history `singapore/DISTILL.md`, `russia/DISTILL.md`, `iran/DISTILL.md`, `usa/federalist_*`, `dugin/` as needed; SEQ-10+
- **CHAPTERS:** `2000_part2_open.md`, `2100_constraints.md`, `2200_rigged_tests.md`, `2300_justice_counterfeits.md`, `2400_misused_compassion.md`, `2500_history.md`, `2600_packages_purity.md`, `2700_practice_close.md`
- **Extra watch:** form/substance positive tool (not freeze, not total constructivism); hermetic + selective logic; justice triad; false foundations; history graded by thrival; persons ≠ packages; matrix filters match Part 1 axioms.

---

## 9. Orchestrator procedure

1. Confirm working tree clean enough to review (or commit WIP) so audit diffs are readable.
2. Launch **Run 1** (one agent). Review scorecards and a sample of new `%%` lines.
3. Adjust severity if the agent is too chatty or too timid (one note in this file or in the next packet).
4. Launch **Run 2** and **Run 3** (parallel OK).
5. Launch **Run 4**.
6. Optional: `rg -n '^%%' chapters/` and triage CRIT/HIGH only for a human rewrite queue.
7. Do **not** auto-rewrite from audit notes without a separate author-approved fix pass.

### Extract commands

```bash
cd book1-6   # or /home/d/code/po/book1-6

rg -n '^%%' chapters/
rg -n '\[CRIT\]|\[HIGH\]' chapters/
rg -n '^%%|^%\+|^%-' chapters/1*.md    # Part 1
rg -n '^%%|^%\+|^%-' chapters/2*.md    # Part 2
```

---

## 10. Known calibration issues (from human edit of 1000 / 1100)

Use as a prior; re-verify on the current file state rather than re-stating blindly.

| Pattern | Where seen | Plan / author stake |
|---------|------------|---------------------|
| Human limits as afterword, not in axiom sequence | 1100 | B7 answers Q2; author moved limits into sequence; particular person under human premise |
| Reality / truth split vs merged “real, knowable, consequential” | 1100 | Plan lists B1 and B2 separately; author may merge — auditor notes consistency with later chapters (1200, E12, charges) |
| Rival-first axiom statements | 1100 | Author: positive vision first; name others later |
| Paraphrase of the five questions | 1000 | Dictionary Q3: full text in opener — flag plan tension even if author softened wording |
| Forced “admit your answers” close | 1000 | Author wants examined priors + consider book’s start, not confession |
| Metaphoric throat-clearing | both | W2 direct words |
| Twin-study style claims | 1100 | Need source or `<!-- verify -->` / soften |
| “Every culture must continue” over-read of life axiom | 1100 | Author: cultures are judged; claim is human life/thriving, not sacralizing every culture |

---

## 11. Out of scope for this audit

- Full E1–E7 editing sweeps (em dash, deletion test, etc.) except where they carry doctrine.
- Rebuilding `book.md` / publish.
- Expanding `sources/` or rewriting the plan.
- Voice homogenization to Blair samples without a content bug.

---

## 12. Changelog

| Date | Note |
|------|------|
| 2026-07-21 | Initial audit instructions: four runs, constructivist/plan adherence focus, `%%`-first markup, scorecards. |
