# Freeze Gates

A unit may be marked `frozen` in `state.yaml` only when all of the following hold.

## Mechanical (scripts enforce)

1. `python scripts/harness.py check` exits 0.
2. Every attack with `severity: critical` has a ruling status in `{answered, wontfix}`.
3. Every `answered` or `partial` ruling has at least one `response_loci` entry.
4. Every locus that looks like `file.md#anchor` or `file.md` refers to an existing chapter file.
5. No attack references a claim ID that is missing from the unit’s claim map (unless `against: NEW` still open—those block freeze).
6. No `against: NEW` attacks remain without either a new claim + retarget or `wontfix`.

## Human (scripts cannot fully enforce)

7. **Energy check:** a smart teenager can retell the unit’s positive teaching in 60 seconds without naming the enemy.
8. **Positive ratio:** the latest cycle’s chapter diff is not dominated by rebuttal scaffolding (see `CICERO_TEMPLATES.md`).
9. **Cycle cap:** no more than two full attack loops unless a remaining `critical` forced a third—and that third is then final.
10. **Tone check:** no in-body reply to pure defamation; rhetorical attacks are `wontfix` or `ledger_only` unless they exposed real ambiguity.

## Recording a freeze

```fish
python scripts/harness.py freeze --chapter 0101_introduction --note "Pilot complete"
```
