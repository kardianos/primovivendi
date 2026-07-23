# Fix plan (post-audit) — content first, then style and arcs

**Status:** Plan only. Do not execute until the author approves Phase 0 decisions and the phase order.  
**Inputs:** Four audit runs (annotated `%%` in `chapters/`), `writer_instructions.md` (W + E rules, especially W12), `AUDIT_INSTRUCTIONS.md`, plan dictionaries, book-markup skill.  
**Goal:** Repair doctrine and plan gaps without breaking chapter arcs; then raise prose toward W12 and mechanical E-passes; leave protected load-bearing blocks intact.

---

## 0. Strategy (why this order)

### 0.1 Content in isolation, then style and flow

**Yes: fix substance first; style and arc polish second.**

| Phase | Job | Why separate |
|-------|-----|----------------|
| **A. Author decisions** | Lock house formulas (one page) | Stops agents rewriting the same seam three ways |
| **B. Doctrine / plan repair** | Small, claim-level edits | Substance changes rearrange later style work if mixed |
| **C. Consistency stitching** | Align floors, scopes, world-wins language | Cross-chapter; needs B finished first |
| **D. Style (W12 + E1–E7)** | One rule per pass, narrow file sets | Matches `writer_instructions.md`: editing is one rule per pass; W12 is secondary but can reshape presentation |
| **E. Arc and flow** | Paragraph transitions, chapter rhythm | Last: every substance edit can open joints; re-flow after the joints are set |

Mixing W12 cadence into CRIT doctrine fixes is a common failure mode: the agent “improves” the sentence and loses the refuse, or lengthens a paragraph until the arc splits (W3). Keep CRIT fixes short and claim-true; polish cadence later.

### 0.2 Arc preservation (throughout)

Editing introduces **disjoint flow**. Countermeasures:

1. **Edit only publish text** (`%+` or unmarked). Do not silently delete author `%-` history unless the author says clean.
2. **One paragraph, one job.** If a fix needs a second claim, add a second paragraph (W3), do not stuff.
3. **Before/after arc check (mandatory on every content pass):**
   - What was the paragraph’s first-sentence claim?
   - What does the next paragraph expect?
   - Does the chapter’s `<!-- chapter: ... -->` still hold?
4. **No length cuts of protected blocks** (list in §2) without author sign-off.
5. **After any multi-paragraph rewrite in a chapter, re-read the section aloud (or one full pass)** before marking done—not the whole book.
6. **Prefer `%+`/`%-` pairs** for contested doctrine lines so the author can compare (W11). Use bare overwrite only for mechanical renames or author-approved “just fix it.”

### 0.3 Agent model: narrow, fresh context

| Rule | Detail |
|------|--------|
| **One agent ≈ one pass ≈ one tight job** | Not “fix Part 1.” Prefer “restore two refuses in 1100 justice + charge.” |
| **Fresh context per agent** | Do not resume a long chat that already rewrote three files; doctrine drift accumulates. |
| **Read list is a hard ceiling** | Packet names every file the agent may open. No “browse the repo.” |
| **Orchestrator (human or lead)** | Decides order, reviews diffs, runs greps, launches next pass. Does not draft manuscript (same spirit as `OPERATOR_RUNBOOK.md`). |
| **Parallelism** | Independent chapters in Phase B only if jobs do not share a house formula still under decision. After Phase A, mechanical renames can parallelize. Style passes: one rule, many files OK if the rule is mechanical (E1, E5). Judgment style (W12) = one chapter per agent. |

### 0.4 Markup during fixes

- Resolve or supersede audit `%%` that the pass addresses: either delete the note after fix, or leave `%% done: …` then strip in a cleanup pass.
- Keep HTML `<!-- arc: / point: / verify: -->` for structure (W10).
- New contested wording: `%+` / `%-` (book-markup skill).
- Do not invent new doctrine in `%%` chatter.

---

## 1. Phase A — Author decisions (gate; no agent prose yet)

**Owner:** author (or author + orchestrator notes).  
**Output:** a short locked note, e.g. `e0_working_notes.md` appendix or a new `HOUSE_FORMULAS.md` (writer-facing). Agents in later phases **must** read it.

Decide once:

| ID | Decision | Options (audit-informed) | Default if silent |
|----|----------|--------------------------|-------------------|
| **A1** | Axiom floor formula | (i) Merged: “reality is real, knowable, and consequential; truth is correspondence to it” as one public answer to Q1; (ii) Split B1 + B2 as plan | Prefer **(i)** if 1100 publish is house; then update 1200, E12, 2700, charges to match |
| **A2** | “World as judge” scope | Always pair with testable: “where the model can be tried / the world can decide” | **Yes** — author already ruled on 1100 |
| **A3** | Positive-first on axioms | State our answers first; name other packages later (H8 / P2.F), not beside every bold head | **Yes** — author ruling |
| **A4** | “Honest scope” diction | Replace book-wide with plain phrase | Recommended: **“same rules for every person in the same position”** / **“inside a defined membership or role”** (keep “scope” only when defined in-place) |
| **A5** | Q1–Q5 wording vs dictionary | (i) Restore dictionary full text; (ii) revise dictionary Q3 to match author `%+` | Must pick one; do not leave plan and opener disagreeing |
| **A6** | Twin / genetic claims | Soften to observable difference; or add source under `sources/` | Soften until sourced |
| **A7** | Dead-earth / consent-only B3 | Restore short form in 1100 or leave to 1800 anti-natal dossier only | Prefer **short restore in 1100** (one sentence chain) |
| **A8** | Style intensity | Full W12 (Blair periodic) vs plain-correct first | **Phase D = full W12**, but content phases use **clear formal prose without forced archaism** so substance stays legible |

**Phase A exit:** formulas written down; then launch Phase B.

---

## 2. Protected blocks (do not thin, reorder, or “summarize”)

Agents in every phase get this list in the packet. Thinning these undoes the audit’s good findings.

| Location | Why |
|----------|-----|
| `1100` reality paragraph (merged B1+B2) correspondence core | House truth |
| `1400` D15 allow failure / race picture | Anti-marble / envy-equality |
| `1500` E0 three units; justice triad block; “trade is discovered, not constructed” | Anti-marble; justice refuse; form/substance |
| `1600` lottery / generic-person paragraph | F1–B8 link |
| `1800` social constructivism dossier; blank slate dossier; outcome equity dossier; “truth as instrument” as **hostile** | External packages |
| `2100` discovered vs constructed definition + reinvention test | Part 2 spine tool |
| `2300` full justice triad | Primary justice definitional home |
| `2400` four false foundations section | P2.D3 |
| `2500` Dugin false-friend grading | Jersey ≠ friend |
| `2600` constructivism dossier; red–green “truth as instrument” | Same |
| `2700` filters 2–3 (wrong/incomplete; form/substance) | Practice lock on anti-totalism |

**Protect rule:** if a style pass would cut these for length, stop and flag the orchestrator.

---

## 3. Phase B — Doctrine and plan repair (content isolation)

**Style constraint for Phase B:** Write clear, formal, complete sentences. **Do not** chase full W12 cadence. Prefer the claim surviving a deletion test (E3) over ornament. Avoid new metaphors; keep existing painted pictures.

**Authority:** author notes in file > Phase A formulas > plan dictionary > sources > manuscript.

### Pass B0 — Dictionary / plan alignment (writer-facing only)

**Agent:** 1, plan-only.  
**Files:** `plan/dictionary_p1.md` (Q, B sections), optional note in `chapter_map_exp.md`.  
**Do not edit** `chapters/` in this pass.

**Jobs:**

1. Record A1–A5 decisions as short “author ruling” bullets under Q and B if the manuscript will diverge from old dictionary text (e.g. merged B1+B2; Q paraphrase; Q8 non-confession close).
2. Do not rewrite the whole dictionary—only the seams that would make future writers reintroduce bugs.

**Read:** Phase A note; dictionary Q/B; skim 1000/1100 publish text.  
**Exit:** dictionary no longer fights A1–A5.

---

### Pass B1 — `1100_axioms.md` CRIT fixes only

**Agent:** 1, single chapter, high attention.  
**Files to edit:** `chapters/1100_axioms.md` only.  
**Read (in order):**

1. `writer_instructions.md` (W11 markup; skip deep W12)
2. Phase A note (A1–A3, A7)
3. `plan/dictionary_p1.md` B1–B8, Q5 map
4. `sources/logic/logical_sequences_full_book.md` SEQ-3, SEQ-4
5. `sources/philosophy/life_axiom_long_argument.md` §3.2 (if restoring dead-earth)
6. Full `1100_axioms.md` including existing `%%` and `%+`/`%-`

**Jobs (do these only):**

| # | Job | Acceptance |
|---|-----|------------|
| B1.1 | **CRIT anthropology:** After positive human nature, one plain refuse: not blank slate; not naturally good / perfectible by remaking the whole order. Positive-first; no rival parade. | Sentence present; blank slate named |
| B1.2 | **CRIT justice:** After positive procedural symmetry, one plain refuse: not equal end-states; not equal return in kind. No need for jargon names. | Q5 fully answered |
| B1.3 | **Opener:** Remove “beside chief antagonists” / “not sole or necessary conclusion” if still present; positive list of answers; other answers later | Matches A3 |
| B1.4 | **Charge:** Mirror B1.1–B1.2 refuses in the charge list | Charge and body agree |
| B1.5 | Optional A7: one short dead-earth / empty-world ranking motivation | If A7 yes |
| B1.6 | A6: soften or `<!-- verify -->` twins claim | No unsourced hard empirics |
| B1.7 | Light merge note: limited + particular as one human premise (cross-ref bold heads; no full rewrite) | B8 plant still clear |

**Explicit non-goals:** W12 polish; rewriting bridge/harvest/knife picture; renaming “honest scope” book-wide; editing other chapters.

**Arc check:** Open → each bold answer → lock together → charge still one rising sequence. Re-read full chapter once after edits.

**Exit report:** quote the new refuse sentences; list open `%%` left for later.

---

### Pass B2 — `1000_questions.md` plan/author seam

**Agent:** 1.  
**Edit:** `1000_questions.md` only.  
**Read:** Phase A (A5); dictionary Q1–Q8; full 1000; author `%%` already on file.

**Jobs:**

1. Implement A5 (restore plan Q text **or** leave author text and confirm B0 updated dictionary).
2. Keep Q2 three-way fork (blank slate / perfectible) intact.
3. Keep Q3 free of thrival smuggling.
4. Keep non-confession charge (author ruling).
5. Optional: one clause that policy fights fail while premises stay hidden (MED audit).

**Non-goals:** Blair style; inventing new questions.

**Exit:** Q list and charge stable; no plan/dictionary fight.

---

### Pass B3 — Empirical / verify hygiene (narrow)

**Agent:** 1.  
**Scope:** only tagged or HIGH source issues, not a full E4 yet.

**Jobs:**

1. `1100` twins: soften or verify per A6.
2. Resolve or re-tag any `<!-- verify: -->` the agent can close from `sources/` without new research (e.g. 2600 Qur’an edition note stays until publish).
3. Do **not** invent quotes or death tolls.

**Read:** `sources/README.md` verification standards; only the files with tags.  
**Exit:** no new bare strong empirics; open verifies listed for human.

---

### Pass B4 — Protect-list integrity scan (read-only agent)

**Agent:** 1, **read-only**.  
**Job:** Confirm protected blocks (§2) still present after B1–B3; report any accidental damage.  
**Exit:** green light for Phase C, or fix ticket back to B1.

---

## 4. Phase C — Cross-chapter consistency (still content, not W12)

House formulas only. Short string/claim alignment. One job family per agent.

### Pass C1 — World-as-judge scope language

**Edit candidates:** `1200`, `1400` (D1), any bare “world wins / things win” without testable scope.  
**Formula (A2):** align to 1100: testable claims; world decides where trial is possible; M1 still sections off divine.  
**Agent:** 1 file-set in one pass if short; else one chapter per agent.  
**Read:** 1100 reality paragraph; M1 dictionary; target chapter only.  
**Non-goal:** rewriting whole method section.

### Pass C2 — Axiom floor list (E12, 2700, charges)

**Edit candidates:** `1500` E12 floor paragraph; `2700` matrix intro if it restates floors; any “first five axioms” lists.  
**Formula (A1):** merged or split consistently; human limits in floor only if A1 says so (plan E12 is B1–B5; author may add knowable/consequential without adding B7).  
**Agent:** 1.  
**Read:** Phase A; 1100 charge; E12 + 2700 only.

### Pass C3 — “Honest scope” diction (mechanical)

**Grep:** `honest scope|honest scopes` across `chapters/`.  
**Replace** with A4 phrase(s). Where “scope” is load-bearing (borders, membership), keep “scope” but define once.  
**Agent:** 1, whole tree, mechanical + light sentence rebalance only.  
**Read:** A4; `vocab_whitelist.txt` if “scope” is defined.  
**Exit:** `rg 'honest scope' chapters/` empty (or only in `%%` notes to strip).

### Pass C4 — Optional thrival / blacklist mechanical

**Grep** blacklist (`thrival`, `frame`/`mirror` as product, etc.).  
**Agent:** 1, mechanical E5 early.  
**Non-goal:** style.

---

## 5. Phase D — Style and mechanical editing (after content is stable)

Follow `writer_instructions.md`: **one rule per pass.** W12 is the cadence goal; apply it **after** E1–E7 (or interleave E1, E5, E7 first because they are mechanical).

### Recommended order

| Pass | Rule | Scope | Agent grain |
|------|------|-------|-------------|
| **D1** | E1 em-dash | all `chapters/` | 1 agent, grep-driven |
| **D2** | E5 vocabulary | all | 1 agent |
| **D3** | E7 actors | worst chapters first (1100, 1500, 1800, 2400, 2500) | 1 chapter per agent |
| **D4** | E2 metaphor | substitution only; do not gut painted pictures | 1–2 chapters per agent |
| **D5** | E4 quote provenance | history + H8 packages (`1800`, `2500`, `2600`) | 1 agent; open `sources/` only |
| **D6** | E6 verify tags | all remaining `<!-- verify -->` | 1 agent |
| **D7** | E3 deletion test | one chapter per agent; **skip protected blocks unless redundant inside them** | tight |
| **D8** | **W12 style** | one chapter per agent | see §5.1 |

### 5.1 W12 style pass (narrow grain)

W12 asks for 1820s Blair-style: balanced periodic sentences, elevated but not archaic, abstract nouns and careful qualification, no contractions/idioms, not choppy modern blog prose.

**Risks:**

- Over-ornamenting doctrine (refuses get soft).
- Killing vitality of **allowed** pictures (W2) by abstracting them into dead labels.
- Splitting or merging paragraphs until W3 arcs break.

**W12 packet rules:**

1. **Preserve every doctrinal refuse and every protected block claim.** Style may rebalance clauses; may not delete the refuse.
2. **Pictures:** keep order / logic / vitality; do not replace a painted bridge with “the mechanism of mutual assistance.”
3. **One chapter per agent.** Read only that chapter + W12 example in `writer_instructions.md` + Phase A formulas (for term consistency).
4. **Do not** open earlier books for voice.
5. Prefer **`%+`/`%-`** for any paragraph that changes more than clause order, so author can reject W12 drift.
6. **Suggested chapter order for W12** (foundations first, long bulk later):
   - D8a: `1000`, `1100` (author calibration files)
   - D8b: `1200`, `1300`
   - D8c: `1400` (may split into two agents: D1–D8 and D9–one day)
   - D8d: `1500` (may split E0 + rest)
   - D8e: `1600`, `1700`
   - D8f: `1800`
   - D8g: Part 2 opens `2000`–`2300` (one or two agents)
   - D8h: `2400` alone (longest)
   - D8i: `2500` alone (quote-sensitive: E4 before or with W12; do not restyle inside quotation marks)
   - D8j: `2600`, `2700`

**W12 acceptance:** chapter still matches its `<!-- chapter -->` job; no new blacklist terms; no new doctrine; Blair-ish cadence on non-picture exposition.

---

## 6. Phase E — Arc and flow (last)

Substance and local style exist. Now fix **joints** between paragraphs and sections.

### Pass E-flow-1 — Chapter-internal arcs

**One chapter per agent.**  
**Job:**

1. Read chapter end-to-end (publish view: mentally drop `%%`/`%-`).
2. Flag or fix: throat-clearing openers, missing transitions, multi-arc paragraphs (W3), first-sentence claim failures (W6) without inventing new claims.
3. Improve memory cadence where author asked (e.g. 1000: reality / self in reality / self among others) **without** dropping Q4–Q5 distinctness.
4. Ensure charges still close the chapter.

**Read:** target chapter only; `writer_instructions` W2–W6, W12 lightly; chapter `<!-- chapter -->` and `<!-- arc -->` comments.  
**Non-goal:** new examples, new packages, length expansion.

### Pass E-flow-2 — Book spine transitions

**Agent:** 1, limited file set of **last/first paragraphs only**:

- `1000` charge ↔ `1100` open  
- `1100` charge ↔ `1200` open  
- `1200` charge ↔ `1300` open  
- `1300` close ↔ `1400` open  
- `1400` close ↔ `1500` E0  
- `1800` close ↔ `2000` open  
- `2600` charge ↔ `2700` open  

**Job:** one-sentence bridges if disjoint; no section rewrites.

### Pass E-flow-3 — Protect + concat sanity

1. Re-run protect-list scan (§2).  
2. `go run concat.go` and skim `book.md` for leftover `%%` / `%+` markers if any should have been cleaned.  
3. `rg` blacklist; `rg` bare world-wins without scope; `rg` honest scope.  
4. Author reads 1000–1100 and one long chapter (1400 or 2400) for voice.

---

## 7. Agent packet templates

### 7.1 Content packet (Phase B/C)

```text
You are fixing doctrine/plan gaps in book1-6. Content only.
Working directory: book1-6/ (or project root with book1-6/ prefixes).

READ ONLY these files, in order:
1. book1-6/FIX_PLAN.md — this pass’s section only [PASS ID]
2. book1-6/HOUSE_FORMULAS.md (or Phase A note) if present
3. book1-6/writer_instructions.md — W11 markup; NOT a full W12 rewrite
4. book1-6/.grok/skills/book-markup/SKILL.md
5. [PLAN / SOURCE FILES NAMED FOR THIS PASS]
6. [TARGET CHAPTER FILES ONLY]

Rules:
- Do not invent new philosophy. Plan + Phase A + author %% win.
- Do not thin protected blocks (FIX_PLAN §2).
- Prefer short claim-true sentences; no Blair ornament hunt.
- Use %+/%- for contested doctrine rewrites.
- Address listed jobs only. Then re-read the edited section for arc continuity.
- Return: what changed, quotes of new refuse lines, remaining %%.

Jobs for this pass:
[LIST]
```

### 7.2 Style packet (Phase D8 W12)

```text
You are applying W12 (Blair / 1820s formal prose) to ONE chapter.
READ: writer_instructions.md W12 + example; book-markup skill;
HOUSE_FORMULAS.md; the single target chapter; FIX_PLAN §2 protect list.

Rules:
- Preserve every doctrinal refuse and protected claim.
- Keep painted pictures (order, logic, vitality); do not abstract them away.
- No new doctrine, packages, or examples.
- No contractions; no modern idioms; prefer periodic balance.
- Contested full-paragraph rewrites: %+/%- pairs.
- After edit, re-read whole chapter once for arc continuity (W3).

Target: chapters/[FILE]
```

### 7.3 Flow packet (Phase E)

```text
You are repairing paragraph and section flow only.
No new claims. No package rewrites. No length cuts of protected blocks.
Read: FIX_PLAN §0.2 and §6; writer_instructions W3 W6 W12 lightly; target chapter.
Job: transitions, one arc per paragraph, charge still closes, first sentences carry claims where possible.
```

---

## 8. Orchestrator checklist

| Step | Action |
|------|--------|
| 1 | Author completes Phase A → write `HOUSE_FORMULAS.md` |
| 2 | Pass B0 dictionary alignment |
| 3 | Pass B1 (1100 CRITs) → **author reviews before any other prose** |
| 4 | Pass B2 (1000), B3 (verify), B4 protect scan |
| 5 | Passes C1–C4 |
| 6 | Mechanical D1, D2, D5, D6 |
| 7 | D3, D4, D7 by chapter as needed |
| 8 | D8 W12 chapter-by-chapter with author spot-checks after 1000/1100 |
| 9 | E-flow-1 on edited chapters; E-flow-2 spine; E-flow-3 concat |
| 10 | Optional: strip resolved audit `%%` in a dedicated cleanup agent (drop-only) |

**Stop conditions:** agent rewrites a protected block; agent reintroduces constructivism as house method; agent “simplifies” justice by dropping outcome/value refuse; W12 pass deletes a refuse for euphony.

---

## 9. Priority map (what actually matters)

### Must fix (blocks honest axioms chapter)

1. `1100` blank-slate / perfectibility refuse  
2. `1100` justice not-outcomes / not-value refuse + charge  
3. `1100` opener vs positive-first body  

### Should fix (consistency / author)

4. Phase A formulas locked and propagated (C1–C3)  
5. `1000` Q wording vs dictionary (A5)  
6. Twins claim (A6)  
7. Optional dead-earth in B3 (A7)  

### Protect (already good)

8. All of §2 — never “improve” by shortening  

### Style later

9. W12 chapter-by-chapter  
10. E1–E7 mechanical  
11. Arc/flow joints  

### Explicit non-goals for the first execution week

- Rewriting Part 2 history  
- Expanding packages  
- Full book voice homogenization before B1 author review  
- Parallel W12 on all chapters before content freeze  

---

## 10. Context budget guidance

| Pass type | Max target size | Why |
|-----------|-----------------|-----|
| B1 1100 CRITs | one chapter ~70 lines + plan slices | Highest stakes; full attention |
| C3 honest scope | whole tree, mechanical | Grep job |
| D8 W12 | one chapter; split 1400/1500/2400/2500 | W12 needs whole-chapter re-read |
| E-flow-2 | only boundary paragraphs | Avoid re-editing bodies |

If an agent’s context is full of audit notes and three dictionaries, **cut the read list**, do not cut the re-read of the edited section.

---

## 11. Relation to existing docs

| Doc | Role after this plan |
|-----|----------------------|
| `AUDIT_INSTRUCTIONS.md` | Historical: how issues were found; still useful to interpret `%%` tags |
| `FIX_PLAN.md` (this file) | Execution order for repair |
| `OPERATOR_RUNBOOK.md` | Launch mechanics; batches were for first draft—use this plan’s passes instead for post-audit |
| `writer_instructions.md` | W/E law; Phase B respects W without full W12; Phase D is E + W12 |
| Chapter `%%` | Work tickets until resolved |

---

## 12. Changelog

| Date | Note |
|------|------|
| 2026-07-21 | Initial fix plan after audit Runs 1–4: content isolation first, W12/flow later, narrow agent passes, protect list, Phase A house formulas. |
