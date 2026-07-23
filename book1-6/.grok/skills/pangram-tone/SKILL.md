# Pangram tone check (book1-6)

Use when: final editing, FIX_PLAN Phase F, user says "pangram", "AI tone", "detector", or after W12/E reflow.

## What it is

Soft **style / AI-texture** signal via [Pangram](https://www.pangram.com) bulk/task API.  
AI-sounding text often means a W12 / W2 / empty-diction problem. **Not** a hard fail; not proof of bad doctrine. Elevated Campbell can false-positive.

## Key (never commit)

1. `PANGRAM_API_KEY` env, or  
2. `book1-6/.pangram_key` (or parent `.pangram_key`) — one line, gitignored.  
See `.pangram_key.example`.

Do not print the key. Do not invent a key.

## Run

From `book1-6/`:

```bash
# one or more files (2+ uses bulk API)
go run pangram.go chapters/1100_axioms.md

# all chapter md files (bulk)
go run pangram.go -all chapters/

# JSON for agents
go run pangram.go -json build/pangram/scan.json -all chapters/
```

Plain prose only: tool drops `%%` and `%-`, keeps `%+` bodies (and unmarked lines), strips HTML comments, YAML front matter, and markdown chrome (headings, bold, links, list markers) so Pangram scores wording, not markup.

## Agent rules

1. Read the tool summary and **flagged windows**, not the whole raw JSON unless needed.  
2. Map flags to house rules: W2 (pictures), W3 (arcs), W12/A8 (every word meaningful), empty abstract glue.  
3. **Do not** auto-humanize with another LLM paste.  
4. **Do not** change doctrine to chase a Human score.  
5. Prefer FIX_PLAN loop: narrow D8 on flagged chapters → E-flow-1 → pangram again.  
6. Quotes in history chapters: inspect before rewriting quoted material (E4).  
7. Soft thresholds (tool defaults): warn if `prediction_short != Human` or `fraction_human < 0.85`. Author may accept residual flags.

## FIX_PLAN placement

Phase F after first-pass E: F1 scan → recursive D8 → E → F1 until satisfied (§6.1).
