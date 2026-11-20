# 31. Review Queue (Agent Quick Ref)

Build once if `civ/bin/` is missing: `fish civ/tools/build-bin.fish`. Spec: `notes/30_review_viewer_spec.md`. (Legacy calls path: `notes/09_chapter_review_instructions.md`.)

## Commands (run from repo root)

```bash
civ/bin/review [ui]                               # Author: open read-only viewer (writes decisions.json only)
civ/bin/review check [-json]                      # Locate quotes & queue counts
civ/bin/review add -slug <name> -strict [-in F|-] # Agent: validate, pin offsets, write timestamped packet
civ/bin/review export -format json -out -         # Agent: read pending apply, holds, and closed items (or -format md -> batch.md)
civ/bin/review apply -dry-run && civ/bin/review apply # Apply clean verbatim accept/revise edits + write applied packet
```

Common flags: `-root`, `-packets`, `-decisions`. Demo flags: `-packets civ/notes/review/demo-packets -decisions civ/notes/review/demo-decisions.json`.

## Packet Schemas (submit via `civ/bin/review add`)

Question / rewrite (`kind`: `call`, `rewrite`, `sentence`, `lint`):
```json
{"version":1,"items":[{"id":"call:civ/chapters/01_ch01.md:ab12cd34:rule","source":"agent","kind":"call","path":"civ/chapters/01_ch01.md","quote":"exact bytes","before":"...","after":"...","problem":"...","rule":"...","locked":false,"alternates":[{"id":"a","label":"shorter","text":"..."}]}]}
```

Response (hold answers and/or manual applies):
```json
{"version":1,"responds_to":["<id>"],"applied":[{"id":"<id>","path":"civ/chapters/01_ch01.md","ok":true,"decision_at":"<at>"}],"answers":[{"id":"<id>","text":"..."}],"items":[]}
```

## Rules

- **Apply (`accept` / `revise`):** Run `civ/bin/review apply -dry-run`, then `apply` for `auto_apply_eligible` items. Handle skipped items (noted, stale/`reaffirm`, overlapping, locked) per author `note`, then record `applied` (with `decision_at`) via `add`. Pass `-include-noted` only if the replacement already reflects the note; pass `-allow-locked` only if explicitly authorized.
- **Hold (`research` / `explain` / `later`):** Leave chapter untouched; reply via `add` with `answers` (and new `items` if proposing new drafts).
- **Closed (`reject` / `withdraw`):** Do not edit; do not re-raise same quote + rule.
- **IDs & context:** `id` = `<source>:<path>:<8-hex-sha256-of-quote>:<rule>` (reuse `id` to refresh; new quote = new `id`). Provide `before`/`after` when `quote` repeats. Always add via `civ/bin/review add` (writes sortable `YYYYMMDD-HHMMSS-<slug>.json`).
- **Guardrails:** Never edit `decisions.json`. Never run `civ/build.fish`. Locked lines (`00_civ_authoritative_bones.md`, `06_editing_guide.md` §1): set `"locked":true`, leave `alternates` empty.
- **After apply:** Run `civ/bin/lint` (and `civ/bin/linear check -file <path>` on touched chapters). Raise follow-ups as new packet items; never insert unapproved prose.
