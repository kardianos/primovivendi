# Operator runbook (book1-6)

How to write this book with an orchestrator session and writer/editing agents. Self-contained: a new operator (human or AI) can run the whole project from this file plus the files it names. Read sections in order.

**Prime rules:**
1. All writing and editing of manuscript files is done by agents. The orchestrator assigns, reviews, runs sweeps, and updates status. The orchestrator never drafts or hand-edits files in `chapters/`.
2. Agents run from the project root (`/home/d/code/po`), so the relative paths in the packets resolve (`book1-6/...`).
3. Input quarantine: agents read only `writer_instructions.md`, the vocab files, `plan/`, and `sources/`. Earlier books in this repository do not exist for them.

---

## 1. Project state (as of 2026-07-17)

- Plan (fixed): `plan/outline_p1.md`, `plan/outline_p2.md`, `plan/dictionary_p1.md`, `plan/dictionary_p2.md`.
- Style rules (fixed): `writer_instructions.md`, `vocab_whitelist.txt`, `vocab_blacklist.txt`.
- Chapters: `chapters/0000_title_block.md` (YAML config only) and `chapter_map_exp.md` (task map, lives at the book root; concat skips `_exp` files). No manuscript chapters exist yet; all are agent-written, starting with Batch 0.
- Build: from `book1-6/`, `go run concat.go` produces `book.md` (new/publish mode: strip `%%`, keep `%+`, drop `%-`, strip HTML comments). `go run concat.go -old` produces `book-old.md` for the pre-rewrite view. `./publish.fish` always uses new mode; full publication build (adapted from book1-5): `../release/lua-book.{pdf,epub,html}`, intermediates in `../.out/`; it ends by invoking `../push_release`, which uploads to the author's WebDAV mount when that mount is present — run the script only when an upload is intended.
- Line markup skill (agents): `.grok/skills/book-markup/SKILL.md` — extract `%%` / `%+` / `%-`, revision norms for Grok/Kimi.
- Working notes: `e0_working_notes.md` (E0 variation material), `plan_review_commentary.md` (external review, open seams), `build/comment_test.*` (renderer evidence for the comment convention).

## 2. Startup context for a new operator

Read in this order:

1. `writer_instructions.md` (the rules you enforce).
2. `chapter_map_exp.md` (task list and per-chapter watch items).
3. `plan/outline_p1.md`, `plan/outline_p2.md` (structure).
4. `plan/dictionary_p1.md`, `plan/dictionary_p2.md` (content source of truth).
5. `plan_review_commentary.md` (open seams; do not paper over them).
6. `vocab_whitelist.txt`, `vocab_blacklist.txt`.

The operator may read anything (quarantine binds writers, not reviewers). Voice judgment on finished chapters belongs to the author, not to the operator's memory of earlier books.

## 3. Architecture and agent mechanics

- **Orchestrator (you):** fill and dispatch packets, review returns, run sweeps, update `chapter_map_exp.md` status. No manuscript writing or editing.
- **Writer agents:** one per chapter, fresh context each, `coder` subagent type, launched from the project root with the packet from section 5.
- **Editing agents:** one per judgment pass (E2 metaphor, E3 deletion, E7 actor), fresh context, never the writer of the chapter they edit.
- **Mechanical sweeps:** scripts, not agents (section 7).
- **Launch mechanics:** launch 3 to 5 writer agents in parallel per batch (one tool call each, same message). Each returns a short report; the chapter is on disk. For larger batches, a swarm call with the packet template and one item per chapter works the same way.
- **Batch discipline:** review the whole batch before launching the next. Voice drift in one chapter means fixing the packet before later chapters inherit it.

## 4. Batches

- **Batch 0 (calibration):** `1000_questions.md` (Q), `1100_axioms.md` (B). Ready-made packets in section 6. The author reviews voice before any further batch launches.
- **Batch 1:** `1200_testable_truth.md` (M), `1300_survival_thriving.md` (C), `1400_personal_virtues.md` (D).
- **Batch 2:** `1500_social_virtues.md` (E incl. E0), `1600_fall_harm.md` (F), `1700_fall_standards.md` (G).
- **Batch 3:** `1800_ending_arc.md` (H).
- **Batch 4 (Part 2, only after Part 1 voice is proven):** `2000` through `2700` per the map.

## 5. Writer agent packet (template)

Fill the bracketed slots. One agent per chapter.

```
You are drafting one chapter of a book. Working directory is the project
root. Read ONLY these files, in order:
- book1-6/writer_instructions.md (rules you must follow; read first)
- book1-6/vocab_whitelist.txt and book1-6/vocab_blacklist.txt
- book1-6/plan/outline_p1.md — the span for [ARC ID]
- book1-6/plan/dictionary_p1.md — entries [ENTRY IDS]
- book1-6/sources/ — [SPECIFIC FILES FOR QUOTES / HISTORY, if any]
Do not read anything else in the repository. Earlier books (book/,
book1-1/, book1-4/, book1-5/, outline/, release/) do not exist for you.
The plan and the sources are the only content base. Every sentence is new.

Task: write book1-6/chapters/[FILENAME] covering the [ARC ID] arc.
- Cover every outline bullet in your span, in outline order, or note a
  conscious skip in a writer comment.
- Start the file with <!-- chapter: point of this chapter -->, then a blank
  line, then the chapter's H2 title (## ...), then a blank line. ### section
  headings are allowed in long chapters. Headings use sentence case
  (capitalize only the first word and proper nouns). Only the two genuine
  parts get an H1 (# Foundations at the top of 1000; # The world is
  consequential and has natural constraints at the top of 2000); nothing
  else uses #. (House convention, set in Batch 1, revised
  2026-07-18: with pandoc --top-level-division=part, # maps to \part, ## to
  \chapter, ### to \section, so chapters must be ## for the PDF hierarchy
  to come out as two parts, seventeen chapters.)
- Add <!-- arc: ID | point: ... --> by any paragraph whose point is not in
  its first sentence.
- Watch items: [PASTE FROM MAP ROW]
- Length: [from outline length instructions / dictionary I3]
- End with the arc's charge if the dictionary entry defines one.

Before returning, self-run: no em dash anywhere (use parentheses); no
blacklisted term; every picture painted with order, logic, vitality or
replaced with direct statement; proper sentences; actors named.

Return: word count, sections written, any deviations or open questions.
Do not paste the chapter into your reply; it is on disk.
```

## 6. Batch 0 packets (ready to paste)

**Agent 1 (1000):**
```
You are drafting one chapter of a book. Working directory is the project
root. Read ONLY these files, in order:
- book1-6/writer_instructions.md (rules you must follow; read first)
- book1-6/vocab_whitelist.txt and book1-6/vocab_blacklist.txt
- book1-6/plan/outline_p1.md — the "Opening: five questions that divide
  societies" block
- book1-6/plan/dictionary_p1.md — section Q (entries Q1–Q8)
Do not read anything else in the repository. Earlier books do not exist
for you. The plan and the sources are the only content base. Every
sentence is new.

Task: write book1-6/chapters/1000_questions.md covering the Q arc.
- Cover every outline bullet in the Q span, in outline order.
- The five questions appear verbatim as written in dictionary Q3; do not
  paraphrase them. Question 3 stays open: do not load it with this book's
  answer.
- Start the file with <!-- chapter: ... -->; add <!-- arc: Qn | point: ... -->
  where a paragraph's point is not in its first sentence.
- Length: tight (this is the opener; the axioms chapter follows).
- End with the Q8 charge.

Before returning, self-run: no em dash anywhere (use parentheses); no
blacklisted term; every picture painted with order, logic, vitality or
replaced with direct statement; proper sentences; actors named.

Return: word count, sections written, any deviations or open questions.
Do not paste the chapter into your reply; it is on disk.
```

**Agent 2 (1100):**
```
You are drafting one chapter of a book. Working directory is the project
root. Read ONLY these files, in order:
- book1-6/writer_instructions.md (rules you must follow; read first)
- book1-6/vocab_whitelist.txt and book1-6/vocab_blacklist.txt
- book1-6/plan/outline_p1.md — the "Opening: the axioms (tight)" block
- book1-6/plan/dictionary_p1.md — section B (entries B1–B8)
Do not read anything else in the repository. Earlier books do not exist
for you. The plan and the sources are the only content base. Every
sentence is new.

Task: write book1-6/chapters/1100_axioms.md covering the B arc.
- State the axioms as this book's answers to the five questions, affirmed
  among rivals, not as if no rival answers exist.
- B3: keep the three layers distinct (axiom, corollaries, participation).
  Affirmed, not deduced. Motivations, not fake proof.
- B8: plant one line only; the full primer comes later at E0.
- Start the file with <!-- chapter: ... -->; add <!-- arc: Bn | point: ... -->
  where a paragraph's point is not in its first sentence.
- Length: tight, clear, early (dictionary I3).
- End with the opening charge: these are the answers held here; next is
  how to look at what can be tested.

Before returning, self-run: no em dash anywhere (use parentheses); no
blacklisted term; every picture painted with order, logic, vitality or
replaced with direct statement; proper sentences; actors named.

Return: word count, sections written, any deviations or open questions.
Do not paste the chapter into your reply; it is on disk.
```

## 7. Mechanical sweeps (run from `book1-6/`)

```
# em dash (E1)
grep -rn '—' chapters/

# blacklist (E5); keep in sync with vocab_blacklist.txt
grep -rniE 'candyland|thrival|tapestry|delve|unravel|important to note' chapters/

# writer comments stripped / no leaks in book.md
go run concat.go && grep -n '<!--' book.md
# line markers must not leak either
go run concat.go && rg -n '^%%|^%\+|^%-' book.md || true

# open verify tags (E6)
grep -rn 'verify:' chapters/

# freeform edit notes and pending rewrites (see book-markup skill)
rg -n '^%%' chapters/
rg -n '^%\+|^%-' chapters/

# blacklist words that slipped into book.md
grep -niE 'candyland|thrival|tapestry' book.md
```

## 8. Orchestrator review checklist (per returned chapter)

1. Read the chapter against `writer_instructions.md` (not against memory of earlier books).
2. Run the sweeps above.
3. Check the chapter covers its outline span (every bullet addressed or a skip noted in a comment).
4. Check the watch items from the map row.
5. Check writer comments exist and say something real.
6. If accepted: set status to `drafted` in `chapter_map_exp.md`. If not: dispatch a fresh agent with the specific fixes listed.

## 9. Editing passes (after a batch is assembled)

- E2 (metaphor sweep), E3 (deletion test), E7 (actor pass): one fresh agent per pass per batch. Packet: the chapter paths, `writer_instructions.md`, and the single rule to enforce. The editor edits files in place and reports a change list.
- E4 (quote provenance): any quotation must have a source in `sources/` or become paraphrase. Absolute.
- Final: orchestrator reads assembled `book.md` start to finish for cross-chapter term consistency and the fixed ending order (H8, H9, H10, H2/H3, H7). The author makes the voice call.

## 10. Known open seams (do not paper over; flag in writer comments)

- **E0 variation slot (blocking decision for 1500):** the author has not yet picked bullets from `e0_working_notes.md`. If still undecided when 1500 is assigned, instruct the writer to render the minimal variation ("People also differ, in talent, in formation, in foresight; some of the difference is inborn and no school makes potentials equal") and tag `<!-- verify: author to confirm variation bullets -->`.
- **F / H10 hard case:** the slide thesis lacks a named hard counterexample and a stated falsifier (see `plan_review_commentary.md` item 1). The F writer must not overclaim; G5 shows the honesty pattern to imitate.
- **B4 tier seam:** instrumental ("practical reach") vs constitutive (chain grounds duties) is unresolved. Writers stay inside B4's current wording; do not improvise a resolution.
- **Part 2 matrix:** seven filters with an optional seventh; watch tool weight.

## 11. Failure recovery

Everything important is on disk. If a session dies: new operator reads section 2, checks `chapter_map_exp.md` status column, resumes from the first row not marked `written` or `drafted`. Update the status column as chapters complete; it is the single source of progress truth.
