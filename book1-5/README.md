# Book 1-5 — The Conservative Frame (Harness Edition)

This directory is a **positive-first manuscript** plus an **adversarial writing harness**.

The book is written in `chapters/`.  
The forge lives in `harness/`.  
Agents attack; the author clarifies; the harness tracks whether every serious objection is disposed.

## Doctrine (non-negotiable)

1. **Positive draft first.** No attack cycle on a blank page.
2. **Attackers never write reader-facing chapters.** They only file attacks.
3. **Author wins by positive clarity** (scene, definition, office, exemplum, hard case), not by matching bile.
4. **One side per agent, forever.**
5. **Insult/defamation is quarantined** (optional Rhetorical Predator pass).
6. **Hard cap:** two full attack cycles per unit unless a `critical` item remains open.
7. **Human author is final aesthetic authority.**

Success metric for any chapter:

> A sympathetic non-specialist can explain the chapter’s positive teaching in 60 seconds **without mentioning the enemy.**

## Layout

```text
book1-5/
  README.md
  AGENT_PROTOCOLS.md          # fixed masks, turns, stop rules
  CICERO_TEMPLATES.md         # exterior-system templates + chapter form
  instructions.md             # chapter file conventions
  outline/                    # high-level plan (evolves)
  chapters/                   # reader-facing prose (source of truth)
  harness/
    claims/                   # per-chapter claim maps (YAML)
    attacks/                  # immutable attack records once filed
    rulings/                  # dispositions for attacks
    sources/                  # bibliographic keys + quote store
    cycles/                   # cycle logs
    state.yaml                # active unit, cycle number, freeze flags
    gates.md                  # freeze rules (human-readable)
  scripts/                    # harness CLI
  agents/                     # prompt packs for each fixed mask
  pilot/                      # first unit workflow notes
```

## Publish

```fish
cd /home/d/code/po/book1-5
fish publish.fish
```

Concatenates `chapters/` → `book.md`, then builds into `../release/`:

| Output | Notes |
|--------|--------|
| `cfc-book.pdf` | print PDF (does not overwrite `cf_book.pdf` from book1-4) |
| `cfc-book.epub` | EPUB3 |
| `cfc-book.html` | standalone HTML |

Requires: `go`, `pandoc`, `xelatex`; optional `gs` for print grayscale pass.

## Quick start (harness)

```fish
cd /home/d/code/po/book1-5

# One-time: local venv (PEP 668 friendly)
python3 -m venv .venv
.venv/bin/pip install -r scripts/requirements.txt

# Prefer the venv interpreter
set h .venv/bin/python scripts/harness.py

# Status of open attacks / freeze readiness
$h status

# List claims for the pilot chapter
$h claims --chapter 0101_introduction

# Unanswered / partial attacks
$h unanswered

# Validate links (attacks → claims → rulings → loci)
$h check

# Start a new cycle log shell for the active unit
$h new-cycle --note "pilot round 1"
```

Agent masks are listed with `$h agents`. Note **scripturalist** and **muslim_package** are separate lanes.

## Pilot unit

- Chapter: `chapters/0101_introduction.md`
- Claims: `harness/claims/0101_introduction.yaml`
- See `pilot/README.md` for the first two-agent drill (near-rival + one exterior).

## Relation to book1-4

`book1-4` remains the prior manuscript and research corpus.  
`book1-5` is a **new writing system and rewrite surface**, not an automatic copy. Seed material may be ported deliberately; do not bulk-overwrite from r4 without a positive draft pass.
