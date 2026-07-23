# Fix plan (post-audit) — content first, then style and arcs

**Status:** Phase 0 **approved** (2026-07-23). Phase A locks live in `HOUSE_FORMULAS.md`. Execution is live; state and waits live in **`EXECUTION_LOG.md`**.  
**Inputs:** Four audit runs (annotated `%%` in `chapters/`), `writer_instructions.md` (W + E rules, especially W12), `AUDIT_INSTRUCTIONS.md`, plan dictionaries, book-markup skill, `HOUSE_FORMULAS.md`.  
**Goal:** Repair doctrine and plan gaps without breaking chapter arcs; then raise prose toward W12 and mechanical E-passes; Pangram tone check and recursive polish last; leave protected load-bearing blocks intact.

**Orchestrator entry:** Read §0.6 (roles), §0.8 (execution log), §0.9 (human gates), then `EXECUTION_LOG.md`. Do not invent a parallel plan.

---

## 0. Strategy (why this order)

### 0.1 Content in isolation, then style and flow

**Yes: fix substance first; style and arc polish second; tone check and re-loop last.**

| Phase | Job | Why separate |
|-------|-----|----------------|
| **A. Author decisions** | Lock house formulas | Stops agents rewriting the same seam three ways — **done** → `HOUSE_FORMULAS.md` |
| **B. Doctrine / plan repair** | Small, claim-level edits | Substance changes rearrange later style work if mixed |
| **C. Consistency stitching** | Align floors, scopes, world-wins language | Cross-chapter; needs B finished first |
| **D. Style (W12 + E1–E7)** | One rule per pass, narrow file sets | Matches `writer_instructions.md`: one rule per pass; W12 secondary but can reshape presentation |
| **E. Arc and flow** | Paragraph transitions, chapter rhythm | Every substance/style edit can open joints; re-flow after joints are set |
| **F. Pangram + loop** | Tone scan; recursive **D8 → E → F1** until satisfied | Soft style signal after first full pass; see §6.1 |

Mixing W12 cadence into CRIT doctrine fixes is a common failure mode: the agent “improves” the sentence and loses the refuse, or lengthens a paragraph until the arc splits (W3). Keep CRIT fixes short and claim-true; polish cadence later.

### 0.2 Arc preservation (throughout)

Editing introduces **disjoint flow**. Countermeasures:

1. **Edit publish text** (unmarked body, or accept `%+` into unmarked when the pass lands). Prefer not to leave dead `%-` history in the file.
2. **One paragraph, one job.** If a fix needs a second claim, add a second paragraph (W3), do not stuff.
3. **Before/after arc check (mandatory on every content pass):**
   - What was the paragraph’s first-sentence claim?
   - What does the next paragraph expect?
   - Does the chapter’s `<!-- chapter: ... -->` still hold?
4. **No length cuts of protected blocks** (list in §2) without author sign-off.
5. **After any multi-paragraph rewrite in a chapter, re-read the section aloud (or one full pass)** before marking done—not the whole book.
6. **Markup:** see §0.4. Default is **accept into publish text** (no lingering `%-`). Use `%+`/`%-` only when both versions are acceptable and simply contested for author comparison.

### 0.3 Agent model: narrow, fresh context

| Rule | Detail |
|------|--------|
| **One agent ≈ one pass ≈ one tight job** | Not “fix Part 1.” Prefer “restore two refuses in 1100 justice + charge.” Pass ID is the agent’s whole life (e.g. `B1`, `D8-1100`, `E-flow-1-1100`). |
| **Fresh context per agent** | **Do not resume** a chat that already rewrote doctrine or style. Doctrine drift and old-verbiage poison accumulate. Next pass = **new** agent even in the same human session. |
| **Read list is a hard ceiling** | Packet names every file the agent may open. No “browse the repo.” No earlier books (`book1-4/`, etc.). |
| **Roles** | Orchestrator / Writer / Editor — see §0.6. Orchestrator does not draft chapters. |
| **Parallelism** | Independent chapters in Phase B only if jobs do not share a house formula still under decision. After Phase A, mechanical renames can parallelize. Style passes: one rule, many files OK if the rule is mechanical (E1, E5). Judgment style (W12) = one chapter per agent. |
| **Git commits authorized** | Agents **may and should** commit completed pass work (see §0.5). Scope: **`book1-6/` only**. Message prefix: **`book1-6: `**. |
| **Execution log** | Single source of resume state: `EXECUTION_LOG.md` (§0.8). |
| **Human gates** | Hard stops in §0.9 — orchestrator waits; does not skip. |

#### 0.3.1 Agent lifecycle (every spawn)

**Lifespan = one FIX_PLAN pass ID.** When the exit report is filed (and commit made if files changed), the agent is **done**. It does not start the next pass.

**Boot sequence (mandatory order; fail the pass if skipped):**

1. `FIX_PLAN.md` — **this pass’s section only** (named pass ID).  
2. `writer_instructions.md` — every new agent; full W overview. Deep W12 + `style/campbell_agent_rules.md` **only** if the pass is D8, F2 style work, or the packet says so.  
3. `HOUSE_FORMULAS.md` — if the pass touches axioms, floors, diction, anthropology, or world-as-judge.  
4. Book-markup skill — if editing `chapters/`.  
5. **Named** plan/source files for this pass only.  
6. **Target chapter file(s) only.**

**Authority when sources disagree:**

| Higher | Lower |
|--------|--------|
| Author direction in-session / HOUSE_FORMULAS locks | Plan dictionary on A1–A8 seams |
| HOUSE_FORMULAS + this pass’s jobs | Stale `%%` audit notes |
| Publish body (unmarked / accepted text) | Dead `%-` history; do not reintroduce it |

**Poison sources (do not treat as house voice):**

- `%-` lines and any unaccepted old pair left in other files  
- `%%` audit notes (work tickets; may be **wrong or superseded** by HOUSE_FORMULAS)  
- `work_files/`, earlier books, long prior agent transcripts  
- Plan dictionary wording that fights HOUSE_FORMULAS on locked seams  

If `%%` and HOUSE_FORMULAS conflict: **HOUSE wins**; fix the prose; delete or mark `%% done:` in the exit report.

**Exit report (mandatory, short):**

1. Pass ID  
2. Files touched  
3. Quotes of new refuse / formula lines (if any)  
4. Open `%%` left (ids or one-line topics)  
5. Explicit “Did not open: …” (prove read ceiling)  
6. Git: commit hash or “no commit (no file changes)”  

**Lifespan by phase:**

| Phase | Typical life | Resume? |
|-------|----------------|---------|
| B content | One chapter or one job family | **No** |
| C mechanical | One grep family (e.g. C3) | **No** |
| D1–D2, D5–D6 | One rule; tree OK | **No** |
| D8 W12 | **One chapter** | **No** |
| E-flow-1 | One chapter | **No** |
| F1 Pangram | Tool + report; orchestrator interprets | N/A |
| F2 loop | New D8 agent per flagged chapter; new E; new F1 | **No** resume across loop iterations |

### 0.4 Markup during fixes

- **Default:** write the final paragraph as **unmarked publish text**. Do **not** leave `%-` in the file after a pass lands.
- **`%+` / `%-` only when both versions are acceptable and simply contested** (author may still choose). If the old line is wrong or superseded, **delete it**; do not keep it as `%-` “for history” in normal repair passes. History lives in git.
- When a contested pair is resolved: keep the winner as unmarked text; drop the loser and the markers.
- Resolve or supersede audit `%%` that the pass addresses: delete after fix, or `%% done: …` then strip in cleanup.
- Keep HTML `<!-- arc: / point: / verify: -->` for structure (W10).
- Do not invent new doctrine in `%%` chatter.
- Agents: book-markup skill; prefer not to re-poison later agents with dead old wording.

### 0.5 Git: agents commit progress

**Authorized.** After a pass’s jobs are done and the exit report is ready, the agent should commit if the working tree has intended changes for that pass.

| Do | Do not |
|----|--------|
| `git status` / `git diff` before commit | Commit secrets (`.pangram_key`, keys, credentials) |
| Stage **only paths under `book1-6/`** for this pass | Stage `book1-4/`, `book1-5/`, or other monorepo dirt |
| Message: **`book1-6: `** then complete sentences (e.g. `book1-6: pass B1 — 1100 anthropology refuses`) | Messages without the prefix; force-push; rewrite shared history |
| One logical commit per pass when practical | Amend commits already on remote unless author asks |
| Include HOUSE/plan edits if this pass owns them | Commit `book.md` / `../release/` / `.out` unless the pass is concat or human-review publish |

Orchestrator may re-commit or split if review requires; agents should still leave a clean commit trail so progress is not lost and later agents are not fed a soup of uncommitted mixed passes.

### 0.6 Roles

Use the **smallest role** that fits the pass. One spawn = one role + one pass ID.

| Role | When to use | Does | Does not |
|------|-------------|------|----------|
| **Orchestrator** | Default control process; resume after interrupt; between passes | Read FIX_PLAN + `EXECUTION_LOG.md`; instantiate packets; spawn **Writer** or **Editor** (new agent each time); run or request tools (§0.7); review exit reports and diffs; update the execution log; stop at human gates (§0.9); ask human for review with build via `./publish.fish` when gate requires reading prose | Edit `chapters/` prose; invent doctrine; skip gates; resume a Writer/Editor chat; “just fix this paragraph” itself |
| **Writer** | Phase B, C (content), and any pass that **changes claims** or plan/dictionary seams | Claim-true unmarked prose; HOUSE_FORMULAS; protect list; exit report; `book1-6:` commit | Full W12 ornament hunt; Pangram auto-rewrite; open other chapters than listed; force-push |
| **Editor** | Phase D (E1–E7, D8 W12), Phase E (flow), Phase F2 style re-entry | One rule or one flow job per pass; preserve refuses and protected blocks; Campbell only when packet is D8/F style; exit report; commit | New doctrine; package expansion; thin protected blocks; treat Pangram as hard fail |

**Role selection by phase:**

| Phase | Role |
|-------|------|
| A (done) | Author (human) |
| B, C | **Writer** |
| D1–D7 mechanical / judgment | **Editor** |
| D8 W12 | **Editor** (one chapter) |
| E-flow | **Editor** |
| F1 Pangram | **Orchestrator** runs tool (or thin tool agent); does not rewrite prose |
| F2 loop body | **Editor** (D8, E); **Orchestrator** runs F1 and log |
| Log / gates / spawn | **Orchestrator** only |

**Orchestrator interrupt resume:** Open `EXECUTION_LOG.md` → row **Status** / **Waiting on** → continue from **Next action**. Do not re-run completed passes unless the log says rework.

### 0.7 Tools (orchestrator and packets)

| Tool | Command (cwd `book1-6/`) | Who | When |
|------|--------------------------|-----|------|
| Concat (publish view) | `go run concat.go` → `book.md` | Orchestrator; Editor on E-flow-3 | After multi-chapter edits; before human prose review |
| Concat (old view) | `go run concat.go -old` → `book-old.md` | Rare; author compare | Not default repair |
| Publish suite | `./publish.fish` | **Orchestrator** when requesting **human review** of readable output | At human gates that need PDF/EPUB/HTML (see §0.9); needs go, pandoc, texlive-xetex |
| Pangram tone | `go run pangram.go [files]` or `-all chapters/` | Orchestrator (F1) | After first-pass E; each F loop. Key: `.pangram_key` or `PANGRAM_API_KEY` |
| Grep hygiene | `rg` blacklist, `honest scope`, bare world-wins | Orchestrator or Editor | C3, E-flow-3, post-pass checks |
| Git | status/diff/add/commit under `book1-6/` only | Writer, Editor, Orchestrator (log/plan) | End of each pass (§0.5) |

Do not commit `.pangram_key`. Do not require Pangram for B/C.

### 0.8 Execution log

**File:** `EXECUTION_LOG.md` (writer-facing; orchestrator maintains it).

**Purpose:** Always show whether we are waiting on an agent, a human, or ready for the next spawn; allow orchestrator resume after interruption.

**Orchestrator must update the log when:**

1. Starting a pass (status → in progress; agent/role noted)  
2. Pass exit report received (status → done or blocked; commit hash)  
3. Hitting a human gate (status → **waiting on human**; what to review; how to unstick)  
4. Human decision recorded (resume next action)  
5. F loop round number changes  

**Do not** keep parallel “memory” only in chat. The log wins.

### 0.9 Human gates (hard stops)

Orchestrator **stops** and sets `EXECUTION_LOG.md` to **waiting on human**. No next Writer/Editor until the log records the human decision.

| Gate ID | After | Human reviews | Unstick when human says | Build for review |
|---------|--------|---------------|-------------------------|------------------|
| **G-A** | Phase A | HOUSE_FORMULAS complete | *Phase A accepted* (already true 2026-07-23) | — |
| **G-B1** | Pass B1 (`1100` CRITs) | Diff + refuses + charge; positive-first | *B1 accepted* or list rework → new B1 agent | **Required:** `./publish.fish` before asking human to review (works in this tree today). Also `go run concat.go` if useful for markdown skim. |
| **G-B** | B2+B3+B4 complete | 1000/1100 still coherent; protect scan green | *Phase B accepted* | **Required:** `./publish.fish` |
| **G-C** | C1–C4 complete | Floors, world-as-judge, no `honest scope` in publish | *Phase C accepted* | **Required:** `./publish.fish` |
| **G-D8-cal** | First D8 on `1000` and `1100` | Voice calibration before bulk D8 | *1000/1100 W12 accepted* | **Required:** `./publish.fish` |
| **G-E** | First full E-flow (1–3) | Arcs and spine | *Phase E accepted* (enter F) | **Required:** `./publish.fish` |
| **G-F** | Each F loop after F1 (or author-chosen cadence) | Pangram flags vs acceptable formality | *F round N accepted* or *F done / ship tone* | `go run pangram.go …`; **Required:** `./publish.fish` when human reads book form |
| **G-ship** | Cleanup / ready to call repair done | Author sign-off | *Repair complete* | **Required:** `./publish.fish` |

**Hard rules:** Orchestrator never marks a gate passed without a human line in the log (date + short decision text). Orchestrator **never requests human prose review without having run `./publish.fish` successfully** for that gate (log the command and exit status). Ignore monorepo paths outside `book1-6/`.

---

## 1. Phase A — Author decisions (**complete**)

**Owner:** author.  
**Output:** `HOUSE_FORMULAS.md` (writer-facing). Agents in later phases **must** read it.  
**Status:** Locked 2026-07-23 (A1–A8). See HOUSE file for formulas (incl. A4 anthropology A4 publish prose, A6/A7).

Decide once:

| ID | Decision | Options (audit-informed) | Default if silent |
|----|----------|--------------------------|-------------------|
| **A1** | Axiom floor formula | (i) Merged: “reality is real, knowable, and consequential; truth is correspondence to it” as one public answer to Q1; (ii) Split B1 + B2 as plan | Prefer **(i)** if 1100 publish is house; then update 1200, E12, 2700, charges to match |
| **A2** | “World as judge” scope | Always pair with testable: “where the model can be tried / the world can decide” | **Yes** — author already ruled on 1100 |
| **A3** | Positive-first on axioms | State our answers first; name other packages later (H8 / P2.F), not beside every bold head | **Yes** — author ruling |
| **A4** | “Honest scope” diction | Replace book-wide with plain phrase | Recommended: **“same rules for every person in the same position”** / **“inside a defined membership or role”** (keep “scope” only when defined in-place) |
| **A5** | Q1–Q5 wording vs dictionary | (i) Restore dictionary full text; (ii) revise dictionary Q3 to match author `%+` | Must pick one; do not leave plan and opener disagreeing |
| **A6** | Twin / genetic claims | Soften to observable difference; or add source under `sources/` | Soften until sourced |
| **A7** | Dead-earth / consent-only B3 | Restore short form in 1100 or leave to 1800 anti-natal dossier only | Author lock: **not in 1100** (assertion only); reason/rhetoric later (e.g. 1800) — see `HOUSE_FORMULAS.md` |
| **A8** | Style intensity | Full W12 (Campbell) vs plain-correct first | **Phase D = full W12**, but content phases use **clear formal prose without forced archaism** so substance stays legible |

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
| B1.5 | A7: **skip** dead-earth / empty-world in 1100 (author: assert only; reason later) | No dead-earth chain in 1100 |
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

**Non-goals:** Full W12 ornament; inventing new questions.

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

### Pass C3 — “Honest scope” diction (flag, then revise)

**Grep:** `honest scope|honest scopes` across `chapters/` (and whitelist/plan if they re-teach the phrase).  
**Do not blind-replace** with one canned string. Flag every hit; rewrite each so the sentence states the real claim (A4: same procedure in the same position; or defined membership/role/jurisdiction; *scope* alone only when defined and load-bearing). Do not blacklist *honest* or *scope* alone.  
**Agent:** 1, whole tree; judgment per hit, not a global substitute table.  
**Read:** `HOUSE_FORMULAS.md` A4; `vocab_whitelist.txt` (fix procedural symmetry definition if it still says “honest scope”).  
**Exit:** `rg 'honest scope' chapters/` empty in publish text (or only in `%%` notes to strip).

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

There is no standalone D9 in the first D pass. **Pangram tone check** and the **D8 ↔ E ↔ check loop** are Phase F (§6.1), after first-pass flow.

### 5.1 W12 style pass (narrow grain)

W12 asks for George Campbell formal didactic prose (see `style/campbell_agent_rules.md`): balanced periodic sentences, elevated but not archaic, abstract nouns and careful qualification, no contractions/idioms, not choppy modern blog prose.

**Risks:**

- Over-ornamenting doctrine (refuses get soft).
- Killing vitality of **allowed** pictures (W2) by abstracting them into dead labels.
- Splitting or merging paragraphs until W3 arcs break.

**W12 packet rules:**

1. **Preserve every doctrinal refuse and every protected block claim.** Style may rebalance clauses; may not delete the refuse.
2. **Pictures:** keep order / logic / vitality; do not replace a painted bridge with “the mechanism of mutual assistance.”
3. **One chapter per agent.** Read only that chapter + W12 example in `writer_instructions.md` + Phase A formulas (for term consistency).
4. **Do not** open earlier books for voice.
5. Prefer **unmarked publish text** when the revision is the house line. Use `%+`/`%-` only if both versions are acceptable and contested (author compare). Do not leave dead `%-` after the pass.
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

**W12 acceptance:** chapter still matches its `<!-- chapter -->` job; no new blacklist terms; no new doctrine; Campbell cadence on non-picture exposition (meaning per word; not costume).

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

## 6.1 Phase F — Pangram tone check + recursive polish (goal loop)

**When:** After first-pass Phase E (reflow done once). Not during B/C content isolation.

**Tool:** `go run .` / `go run pangram.go` (see program header). Key: gitignored `.pangram_key` or env `PANGRAM_API_KEY`. Prefer **bulk API** when scanning more than one chapter.

**Soft rule only:** AI-sounding tone usually signals a style violation (W12 / W2 / empty diction). It is **not** a hard fail and not proof of false doctrine. Elevated Campbell can false-positive; inspect windows, do not thrash claims.

### Pass F1 — Pangram scan

1. Run on publish text of target chapter(s) (tool strips `%%` / `%+` / `%-` / HTML comments like concat new-mode).  
2. Report `headline`, `prediction_short`, fractions, and flagged windows.  
3. Author or orchestrator decides which chapters re-enter the loop.

### Pass F2 — Recursive D8 → E → check (until satisfied)

**Goal loop** (last three steps of the book polish):

```text
loop until author satisfied (or max N rounds, default 3):
  1. D8  — W12 style only on chapters/windows still flagged or still ugly
  2. E   — E-flow-1 on those chapters; E-flow-2 if spine joints moved
  3. F1  — Pangram check again (bulk if many files)
```

**Exit when:** author accepts tone and arcs, or remaining flags are accepted false positives (formal prose, long periods, scripture/history quotes).  
**Do not:** auto-rewrite from the detector; do not humanize with a second LLM paste; fix under W/A8/HOUSE_FORMULAS only.

**Skill:** `.grok/skills/pangram-tone/SKILL.md` when present.

---

## 7. Agent packet templates

### 7.0 Lifecycle block (prepend to every packet)

```text
AGENT LIFECYCLE — pass [PASS_ID] only. Do not start adjacent passes.
Boot read order (mandatory):
1. FIX_PLAN.md — this pass’s section only
2. writer_instructions.md (every agent; deep W12 only if this pass is D8/F style)
3. HOUSE_FORMULAS.md if listed for this pass
4. book-markup skill if editing chapters/
5. Named plan/source files only
6. Target file(s) only

Publish body is authority. %% = tickets (may be stale; HOUSE wins).
Do not leave %- after you land a fix unless both versions are acceptable and contested.
Do not invent doctrine. Do not resume a prior agent chat.
When done: exit report (§0.3.1) + git commit for this pass (§0.5) if files changed.
Then STOP.
```

### 7.1 Content packet (Phase B/C)

```text
[7.0 lifecycle with PASS_ID]
You are fixing doctrine/plan gaps in book1-6. Content only.
Working directory: book1-6/ (or project root with book1-6/ prefixes).

READ ONLY these files, in order:
1. book1-6/FIX_PLAN.md — this pass’s section only [PASS ID]
2. book1-6/writer_instructions.md — W overview; W11 markup; NOT a full W12 rewrite
3. book1-6/HOUSE_FORMULAS.md
4. book1-6/.grok/skills/book-markup/SKILL.md
5. [PLAN / SOURCE FILES NAMED FOR THIS PASS]
6. [TARGET CHAPTER FILES ONLY]

Rules:
- Do not invent new philosophy. HOUSE_FORMULAS + this pass’s jobs + author direction win over stale %%.
- Do not thin protected blocks (FIX_PLAN §2).
- Prefer short claim-true sentences; no Campbell ornament hunt (claim-true first).
- Default: unmarked publish text. %+/%- only if both versions acceptable and contested.
- Address listed jobs only. Then re-read the edited section for arc continuity.
- Commit this pass (§0.5). Exit report (§0.3.1).

Jobs for this pass:
[LIST]
```

### 7.2 Style packet (Phase D8 W12)

```text
[7.0 lifecycle with PASS_ID e.g. D8-1100]
You are applying W12 (Campbell formal prose) to ONE chapter.
READ: FIX_PLAN this pass; writer_instructions.md W12 + style/campbell_agent_rules.md;
book-markup skill; HOUSE_FORMULAS.md; the single target chapter; FIX_PLAN §2 protect list.

Rules:
- Preserve every doctrinal refuse and protected claim.
- Keep painted pictures (order, logic, vitality); do not abstract them away.
- No new doctrine, packages, or examples.
- No contractions; no modern idioms; prefer periodic balance; every word meaningful (A8).
- Default: unmarked publish text. %+/%- only if both versions acceptable and contested.
- After edit, re-read whole chapter once for arc continuity (W3).
- Commit this pass. Exit report. STOP.

Target: chapters/[FILE]
```

### 7.3 Flow packet (Phase E)

```text
[7.0 lifecycle with PASS_ID]
You are repairing paragraph and section flow only.
No new claims. No package rewrites. No length cuts of protected blocks.
Read: FIX_PLAN §0.2, §0.3.1, §6; writer_instructions W3 W6 W12 lightly; target chapter.
Job: transitions, one arc per paragraph, charge still closes, first sentences carry claims where possible.
Default: unmarked publish text. Commit this pass. Exit report. STOP.
```

---

## 8. Orchestrator checklist

| Step | Action |
|------|--------|
| 0 | Read `EXECUTION_LOG.md`; act as **Orchestrator** only (§0.6). Update log every transition. |
| 1 | Every child launch: role + pass ID + §7.0 lifecycle; **do not resume** Writer/Editor agents |
| 2 | Phase A done → HOUSE_FORMULAS present (**G-A** passed) |
| 3 | Pass B0 (Writer) → `book1-6:` commit → log |
| 4 | Pass B1 (Writer) → commit → **hard stop G-B1** (human; optional `./publish.fish`) |
| 5 | After G-B1: B2, B3, B4 → commits → **hard stop G-B** |
| 6 | After G-B: C1–C4 → commits → **hard stop G-C** |
| 7 | After G-C: D1, D2, D5, D6 (Editor) → commits |
| 8 | D3, D4, D7 as needed → commits |
| 9 | D8 on `1000` then `1100` only → **hard stop G-D8-cal** (`./publish.fish` recommended) |
| 10 | After G-D8-cal: remaining D8 chapters (one Editor per chapter) → commits |
| 11 | E-flow-1 / E-flow-2 / E-flow-3 → **hard stop G-E** |
| 12 | **Phase F:** F1 Pangram → log flags → **G-F** as needed → F2 loop (new Editors; new F1 each round) until human *F done* |
| 13 | Optional `%%` cleanup agent → commit → **G-ship** |

**Stop conditions:** agent rewrites a protected block; agent reintroduces constructivism as house method; agent “simplifies” justice by dropping outcome/value refuse; W12 pass deletes a refuse for euphony; Pangram used as hard gate or auto-humanizer; agent resumes a prior pass chat; agent leaves `%-` that re-poisons; orchestrator drafts chapter prose; orchestrator skips a §0.9 gate; commit outside `book1-6/` or without `book1-6:` prefix.

---

## 9. Priority map (what actually matters)

### Must fix (blocks honest axioms chapter)

1. `1100` blank-slate / perfectibility refuse  
2. `1100` justice not-outcomes / not-value refuse + charge  
3. `1100` opener vs positive-first body  

### Should fix (consistency / author)

4. Phase A formulas locked and propagated (C1–C3)  
5. `1000` Q wording vs dictionary (A5)  
6. Twins / study language (A6: soften; assert limited–particular–self-interested)  
7. Dead-earth: **not** in 1100 (A7); later dossiers only  

### Protect (already good)

8. All of §2 — never “improve” by shortening  

### Style later

9. W12 chapter-by-chapter (Campbell)  
10. E1–E7 mechanical  
11. Arc/flow joints  
12. Phase F: Pangram + D8/E loop until satisfied  

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
| `FIX_PLAN.md` (this file) | Execution order, roles, gates, tools |
| `EXECUTION_LOG.md` | Live run state; waits; resume after interrupt |
| `HOUSE_FORMULAS.md` | Phase A locks (authority on A1–A8) |
| `OPERATOR_RUNBOOK.md` | Legacy draft batches—use this plan’s passes for post-audit |
| `writer_instructions.md` | W/E law; Phase B respects W without full W12; Phase D is E + W12 |
| `pangram.go` / `.grok/skills/pangram-tone/` | Phase F tone tool + skill |
| `publish.fish` | Human-review render (PDF/EPUB/HTML) |
| `concat.go` | Publish markdown concat |
| Chapter `%%` | Work tickets until resolved |

### 11.1 Orchestrator launch (generic)

```text
You are the Orchestrator for book1-6 post-audit repair (FIX_PLAN §0.6).
You do not write chapter prose. You read FIX_PLAN.md and EXECUTION_LOG.md,
spawn one Writer or Editor per pass, enforce human gates, run tools in §0.7,
and keep the execution log current so work can resume after interruption.
Start from EXECUTION_LOG.md "Next action." Commits: book1-6/ only, message
prefix "book1-6: ".
```

---

## 12. Changelog

| Date | Note |
|------|------|
| 2026-07-21 | Initial fix plan after audit Runs 1–4: content isolation first, W12/flow later, narrow agent passes, protect list, Phase A house formulas. |
| 2026-07-23 | W12 wording: Campbell (not Blair). Phase F: Pangram bulk tone check after E; recursive D8→E→check until satisfied. Tool: `pangram.go`, key `.pangram_key`. |
| 2026-07-23 | §0.3.1 agent lifecycle (boot, poison list, exit, no resume); §0.4 default unmarked publish (avoid `%-` unless contested); §0.5 agents commit per pass; packets §7.0. |
| 2026-07-23 | Phase 0 approved. Roles §0.6; tools §0.7; EXECUTION_LOG §0.8; human gates §0.9; F in phase table; git prefix `book1-6:`; publish.fish at review gates. |
