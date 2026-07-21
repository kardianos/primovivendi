---
name: book-markup
description: >
  Line-level revision markup for book1-6 (Life Under Axioms): %% comments,
  %+ new paragraphs, %- old paragraphs; extract notes with rg; concat -old vs
  new; human-first rewrite norms for Grok/Kimi agents. Use when editing
  chapters/, reviewing rewrites, listing open edit notes, running concat,
  or when the user mentions %%, %+, %-, book-old, or /book-markup.
---

# Book markup (book1-6)

Work from `book1-6/` (this skill lives under that tree). Manuscript chapters are `chapters/*.md` (skip `*_exp.md`). Paragraphs are **single lines**. Markup is **column 0 only**.

## Marker contract

| Prefix | Meaning | `go run concat.go` (new, default → `book.md`) | `go run concat.go -old` → `book-old.md` |
|--------|---------|-----------------------------------------------|------------------------------------------|
| `%%` | Freeform edit note | drop | drop |
| `%+ ` | New paragraph body | keep body (strip marker + optional space) | drop line |
| `%- ` | Old paragraph body | drop line | keep body |

Rules:

- Match only at **start of line** (`^%%`, `^%+`, `^%-`). Mid-line `%` (e.g. `50%`) is prose.
- Prefer a space after `%+` / `%-`. Concat strips one optional space either way.
- No block/span forms. No mid-sentence add/del.
- Orphan `%+` (insert) or orphan `%-` (delete) is fine. Adjacent `%-` then `%+` is the usual replace pair.
- **HTML comments stay separate:** `<!-- chapter: ... -->`, `<!-- arc: ... | point: ... -->`, `<!-- verify: ... -->` for stable structure. Use `%%` for ephemeral edit chatter only.

Example:

```markdown
%% soften voice; keep the claim
%- Old paragraph that was too stiff.
%+ New paragraph that sounds like a human.
```

## Publish path

- Release always uses **new** mode. `./publish.fish` runs `go run concat.go` with no `-old`.
- Never pass `-old` into the publish script.
- Order in concat: line markers, then HTML comment strip, then templates → `book.md`.

```bash
cd book1-6   # or /home/d/code/po/book1-6
go run concat.go              # book.md (publish view)
go run concat.go -old         # book-old.md (pre-rewrite view)
./publish.fish                # always new; ships lua-book.*
```

Sanity after concat:

```bash
rg -n '<!--|^%%|^%\+|^%-' book.md || true   # expect no matches
```

## Extract notes (prefer rg)

From `book1-6/`:

```bash
# Freeform edit comments
rg -n '^%%' chapters/

# With ±2 lines of context
rg -n -C2 '^%%' chapters/

# Pending insertions / deletions
rg -n '^%+' chapters/
rg -n '^%-' chapters/

# All line markers
rg -n '^%%|^%\+|^%-' chapters/

# One chapter
rg -n '^%%|^%\+|^%-' chapters/1400_personal_virtues.md

# Structured HTML writer notes (not line markers)
rg -n '<!-- (arc|point|verify|chapter):' chapters/
```

No separate notes binary required unless JSON output is needed later.

## Agent norms (human-first)

1. **Humans own the rewrite.** Prefer markers on contested lines over silent overwrite when the prior wording may still matter.
2. If the human said “just fix it,” edit unmarked lines directly; do not force `%+`/`%-` theater.
3. Never invent markers mid-line or wrap multi-line blocks in a single marker.
4. Keep arc/point/verify in HTML comments; put “try softer / check cite later” on `%%` lines.
5. After a batch of rewrites, run `go run concat.go` and optionally `-old`, then skim or diff:

```bash
go run concat.go && go run concat.go -old
diff -u book-old.md book.md | head -200
```

6. Input quarantine for manuscript writers still applies: prose sources are `writer_instructions.md`, vocab files, `plan/`, `sources/` — not earlier books. This skill only adds markup/extract/build behavior.

## Related files

- `concat.go` — strip/select implementation and `-old` flag
- `publish.fish` — release build (new mode only)
- `writer_instructions.md` — W10 HTML comments, W11 line markers
- `OPERATOR_RUNBOOK.md` — orchestrator sweeps
