# Pilot workflow

**Unit:** `chapters/0101_introduction.md`  
**Claims:** `harness/claims/0101_introduction.yaml`  
**Swarm (first loop):** `near_rival` + `constructivist`  
**Seeded also (for lane demo):** `scripturalist` + `muslim_package` sample attacks — do not merge those lanes.

## Why this pilot

Prove the forge before scaling:

1. Positive draft exists.
2. Claims are explicit.
3. Separate agents file separate attacks (especially scripturalist ≠ muslim_package).
4. Author answers with **positive clarity**, not counter-sermons.
5. Harness `check` / `status` stay green enough to freeze.

## Seeded attacks (cycle 0 examples)

| ID | Agent | Against | Point |
|----|--------|---------|--------|
| A-0001 | near_rival | P-0101-10 | Holism / purity pressure |
| A-0002 | constructivist | P-0101-05 | Correspondence as power |
| A-0003 | scripturalist | P-0101-05 | Truth as divine think (Christian epistemology lane) |
| A-0004 | muslim_package | P-0101-09 | Revealed-law sovereignty vs civic procedure |

Notice A-0003 and A-0004 hit **different** principles and use **different** source keys. That separation is intentional.

## Suggested first author pass (before freeze)

For each open attack, prefer:

1. One clearer sentence in the positive draft, or  
2. One short fence, or  
3. One concrete case —  

then file a ruling with loci.

Example:

```fish
cd /home/d/code/po/book1-5
python scripts/harness.py status
python scripts/harness.py check

# After revising the chapter:
python scripts/harness.py rule A-0001 \
  --status partial \
  --method fence \
  --loci "0101_introduction.md#five-principles" \
  --note "Interlock kept; add that scope refinements inside a principle are not heresy"

python scripts/harness.py rule A-0002 \
  --status answered \
  --method positive_clarify \
  --loci "0101_introduction.md#five-principles" \
  --note "Bridge/theft consequences: world answers back; power does not hold bridges up"
```

## Starting cycle 1

```fish
python scripts/harness.py new-cycle --note "pilot loop 1"
```

Then open agent prompt packs in `agents/` and file new attacks only against open or weakened claims.

## Freeze

```fish
python scripts/harness.py freeze --chapter 0101_introduction --note "pilot"
```

Blocked while any `critical` is open (A-0002 until ruled).
