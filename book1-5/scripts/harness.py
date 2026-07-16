#!/usr/bin/env python3
"""Adversarial writing harness for book1-5.

State lives under book1-5/harness/. Reader prose lives under book1-5/chapters/.
Attackers never write chapters through this tool.
"""

from __future__ import annotations

import argparse
import re
import sys
from datetime import date
from pathlib import Path
from typing import Any

try:
    import yaml
except ImportError:
    print("PyYAML is required: pip install pyyaml", file=sys.stderr)
    sys.exit(1)


ROOT = Path(__file__).resolve().parents[1]
HARNESS = ROOT / "harness"
CHAPTERS = ROOT / "chapters"
CLAIMS_DIR = HARNESS / "claims"
ATTACKS_DIR = HARNESS / "attacks"
RULINGS_DIR = HARNESS / "rulings"
CYCLES_DIR = HARNESS / "cycles"
STATE_PATH = HARNESS / "state.yaml"

AGENTS = {
    "constructivist",
    "outcome_equity",
    "universalist",
    "scripturalist",
    "muslim_package",
    "blank_slate",
    "near_rival",
    "rhetorical_predator",
    "nihilist_anti_thrival",
    "fate_sharing_anti_success",  # Dugin-package / sacral anti-success exterior
}

SEVERITIES = {"critical", "major", "minor", "rhetorical"}
MODES = {
    "primary_source",
    "popular_book",
    "x_discourse",
    "logic",
    "motive_attack",
    "ambiguity",
}
STATUSES = {"answered", "partial", "wontfix", "deferred"}
METHODS = {
    "positive_clarify",
    "fence",
    "new_case",
    "glossary",
    "ledger_only",
    "structural",
}

ATTACK_ID_RE = re.compile(r"^A-\d{4}$")
CLAIM_ID_RE = re.compile(r"^[PC]-\d{4}-\d{2}$")


def load_yaml(path: Path) -> Any:
    with path.open(encoding="utf-8") as f:
        return yaml.safe_load(f)


def dump_yaml(path: Path, data: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        yaml.safe_dump(
            data,
            f,
            sort_keys=False,
            allow_unicode=True,
            default_flow_style=False,
            width=88,
        )


def load_state() -> dict[str, Any]:
    if not STATE_PATH.exists():
        return {}
    return load_yaml(STATE_PATH) or {}


def save_state(state: dict[str, Any]) -> None:
    dump_yaml(STATE_PATH, state)


def iter_yaml_files(directory: Path) -> list[Path]:
    if not directory.exists():
        return []
    return sorted(
        p for p in directory.glob("*.yaml") if p.is_file() and not p.name.startswith(".")
    )


def load_all_claims() -> dict[str, dict[str, Any]]:
    """claim_id -> claim dict merged with chapter."""
    out: dict[str, dict[str, Any]] = {}
    for path in iter_yaml_files(CLAIMS_DIR):
        doc = load_yaml(path) or {}
        chapter = doc.get("chapter") or path.stem
        for claim in doc.get("claims") or []:
            cid = claim.get("id")
            if not cid:
                continue
            merged = dict(claim)
            merged["_chapter"] = chapter
            merged["_file"] = str(path.relative_to(ROOT))
            out[cid] = merged
    return out


def load_all_attacks() -> dict[str, dict[str, Any]]:
    out: dict[str, dict[str, Any]] = {}
    for path in iter_yaml_files(ATTACKS_DIR):
        doc = load_yaml(path) or {}
        aid = doc.get("id") or path.stem
        doc = dict(doc)
        doc["_file"] = str(path.relative_to(ROOT))
        out[aid] = doc
    return out


def load_all_rulings() -> dict[str, dict[str, Any]]:
    out: dict[str, dict[str, Any]] = {}
    for path in iter_yaml_files(RULINGS_DIR):
        doc = load_yaml(path) or {}
        attack = doc.get("attack") or path.stem.replace("R-", "A-")
        doc = dict(doc)
        doc["_file"] = str(path.relative_to(ROOT))
        out[attack] = doc
    return out


def next_attack_id(attacks: dict[str, dict[str, Any]]) -> str:
    nums = []
    for aid in attacks:
        m = re.match(r"A-(\d+)$", aid)
        if m:
            nums.append(int(m.group(1)))
    n = max(nums) + 1 if nums else 1
    return f"A-{n:04d}"


def chapter_path(chapter: str) -> Path:
    slug = chapter if chapter.endswith(".md") else f"{chapter}.md"
    # allow chapter id without path
    if slug.startswith("chapters/"):
        return ROOT / slug
    return CHAPTERS / Path(slug).name


def locus_exists(locus: str) -> tuple[bool, str]:
    """Return (ok, message). Locus forms: file.md#anchor | file.md | #anchor (active chapter)."""
    locus = locus.strip()
    if not locus:
        return False, "empty locus"
    if locus.startswith("#"):
        return True, "anchor-only locus (not mechanically verified)"
    if "#" in locus:
        file_part, anchor = locus.split("#", 1)
    else:
        file_part, anchor = locus, ""
    path = chapter_path(file_part)
    if not path.exists():
        # try relative from root
        alt = ROOT / file_part
        if alt.exists():
            path = alt
        else:
            return False, f"missing chapter file for locus: {locus}"
    if anchor:
        text = path.read_text(encoding="utf-8")
        # {#anchor} or raw anchor id in heading attributes
        if f"{{#{anchor}}}" not in text and f'id="{anchor}"' not in text:
            # soft warn: heading text match not required
            return True, f"file ok; anchor '{anchor}' not found as {{#{anchor}}} (soft)"
    return True, "ok"


def cmd_status(_: argparse.Namespace) -> int:
    state = load_state()
    attacks = load_all_attacks()
    rulings = load_all_rulings()
    claims = load_all_claims()

    active = (state.get("active_unit") or {})
    print(f"Book root: {ROOT}")
    print(f"Active unit: {active.get('chapter')}  cycle={active.get('cycle')}  phase={active.get('phase')}")
    print(f"Claims: {len(claims)}  Attacks: {len(attacks)}  Rulings: {len(rulings)}")
    print()

    open_crit = []
    open_major = []
    open_other = []
    for aid, atk in sorted(attacks.items()):
        st = (rulings.get(aid) or {}).get("status")
        if st in ("answered", "wontfix"):
            continue
        bucket = (aid, atk, st)
        sev = atk.get("severity")
        if sev == "critical":
            open_crit.append(bucket)
        elif sev == "major":
            open_major.append(bucket)
        else:
            open_other.append(bucket)

    def show(title: str, items: list) -> None:
        print(f"## {title} ({len(items)})")
        if not items:
            print("  (none)")
            return
        for aid, atk, st in items:
            print(
                f"  {aid}  [{atk.get('severity')}]  {atk.get('agent')}  -> {atk.get('against')}  "
                f"ruling={st or 'NONE'}"
            )
            summary = (atk.get("summary") or "").strip().replace("\n", " ")
            if len(summary) > 100:
                summary = summary[:97] + "..."
            print(f"      {summary}")
        print()

    show("Open critical", open_crit)
    show("Open major", open_major)
    show("Open minor/rhetorical/partial/deferred", open_other)

    frozen = state.get("frozen_units") or []
    if frozen:
        print("## Frozen units")
        for u in frozen:
            print(f"  {u}")
    return 0


def cmd_claims(args: argparse.Namespace) -> int:
    chapter = args.chapter
    path = CLAIMS_DIR / f"{chapter}.yaml"
    if not path.exists():
        print(f"No claims file: {path}", file=sys.stderr)
        return 1
    doc = load_yaml(path) or {}
    print(f"Chapter: {doc.get('chapter', chapter)}  status={doc.get('status')}")
    for c in doc.get("claims") or []:
        deps = ",".join(c.get("depends_on") or []) or "-"
        anchors = ",".join(c.get("anchors") or []) or "-"
        print(f"{c.get('id')}  ({c.get('type')})  deps=[{deps}]  anchors=[{anchors}]")
        print(f"  {c.get('text')}")
    return 0


def cmd_unanswered(_: argparse.Namespace) -> int:
    attacks = load_all_attacks()
    rulings = load_all_rulings()
    n = 0
    for aid, atk in sorted(attacks.items()):
        st = (rulings.get(aid) or {}).get("status")
        if st in ("answered", "wontfix"):
            continue
        n += 1
        print(f"{aid}\t{atk.get('severity')}\t{atk.get('agent')}\t{atk.get('against')}\t{st or 'NONE'}")
    if n == 0:
        print("All attacks answered or wontfix.")
    return 0


def cmd_check(_: argparse.Namespace) -> int:
    errors: list[str] = []
    warnings: list[str] = []
    claims = load_all_claims()
    attacks = load_all_attacks()
    rulings = load_all_rulings()
    state = load_state()
    active_chapter = (state.get("active_unit") or {}).get("chapter")

    for aid, atk in attacks.items():
        if not ATTACK_ID_RE.match(str(aid)):
            warnings.append(f"{aid}: id should match A-0001 form")
        agent = atk.get("agent")
        if agent not in AGENTS:
            errors.append(f"{aid}: unknown agent '{agent}'")
        sev = atk.get("severity")
        if sev not in SEVERITIES:
            errors.append(f"{aid}: bad severity '{sev}'")
        mode = atk.get("mode")
        if mode not in MODES:
            errors.append(f"{aid}: bad mode '{mode}'")
        if not (atk.get("summary") or "").strip():
            errors.append(f"{aid}: missing summary")
        against = atk.get("against")
        if against == "NEW":
            warnings.append(f"{aid}: against NEW (blocks freeze until retargeted or wontfix)")
        elif against not in claims:
            errors.append(f"{aid}: against unknown claim '{against}'")
        # separate-lane reminder
        if agent == "scripturalist" and "islam" in (atk.get("summary") or "").lower():
            warnings.append(
                f"{aid}: scripturalist attack mentions Islam — consider muslim_package lane"
            )
        if agent == "muslim_package" and "crampton" in (atk.get("summary") or "").lower():
            warnings.append(
                f"{aid}: muslim_package cites Christian scripturalism — consider scripturalist lane"
            )

    for attack_id, ruling in rulings.items():
        if attack_id not in attacks:
            errors.append(f"ruling for missing attack {attack_id}")
            continue
        st = ruling.get("status")
        if st not in STATUSES:
            errors.append(f"{attack_id}: bad ruling status '{st}'")
        method = ruling.get("method")
        if method not in METHODS:
            errors.append(f"{attack_id}: bad method '{method}'")
        if st in ("answered", "partial"):
            loci = ruling.get("response_loci") or []
            if not loci:
                errors.append(f"{attack_id}: {st} requires response_loci")
            for loc in loci:
                ok, msg = locus_exists(str(loc))
                if not ok:
                    errors.append(f"{attack_id}: locus {msg}")
                elif msg.startswith("file ok; anchor"):
                    warnings.append(f"{attack_id}: {msg}")

    # critical gate report (not always hard-error unless --strict-freeze)
    open_critical = []
    for aid, atk in attacks.items():
        if atk.get("severity") != "critical":
            continue
        st = (rulings.get(aid) or {}).get("status")
        if st not in ("answered", "wontfix"):
            open_critical.append(aid)

    print(f"check: {len(claims)} claims, {len(attacks)} attacks, {len(rulings)} rulings")
    if active_chapter:
        print(f"active chapter: {active_chapter}")
    for w in warnings:
        print(f"WARNING: {w}")
    for e in errors:
        print(f"ERROR: {e}")
    if open_critical:
        print(f"OPEN CRITICAL (blocks freeze): {', '.join(open_critical)}")
    if errors:
        print("FAIL")
        return 1
    print("OK (mechanical). Human energy/positive-ratio gates still required for freeze.")
    if open_critical:
        print("Note: open critical attacks remain.")
        return 2
    return 0


def cmd_new_attack(args: argparse.Namespace) -> int:
    attacks = load_all_attacks()
    aid = args.id or next_attack_id(attacks)
    if not ATTACK_ID_RE.match(aid):
        print(f"Bad id {aid}; use A-0001 form", file=sys.stderr)
        return 1
    if aid in attacks:
        print(f"Attack {aid} already exists", file=sys.stderr)
        return 1
    agent = args.agent
    if agent not in AGENTS:
        print(f"Unknown agent {agent}. Choose from: {', '.join(sorted(AGENTS))}", file=sys.stderr)
        return 1
    against = args.against
    state = load_state()
    chapter = args.chapter or (state.get("active_unit") or {}).get("chapter")
    cycle = (state.get("active_unit") or {}).get("cycle", 0)

    doc = {
        "id": aid,
        "chapter": chapter,
        "cycle": cycle,
        "agent": agent,
        "against": against,
        "severity": args.severity,
        "mode": args.mode,
        "summary": args.summary or "TODO: write attacker summary in true-believer voice.",
        "quote": args.quote or "",
        "source_ref": args.source_ref or "",
        "subversion": args.subversion or "TODO: how the author's wording can be twisted.",
        "created": date.today().isoformat(),
    }
    path = ATTACKS_DIR / f"{aid}.yaml"
    dump_yaml(path, doc)
    print(f"Wrote {path.relative_to(ROOT)}")
    return 0


def cmd_rule(args: argparse.Namespace) -> int:
    aid = args.attack_id
    if not aid.startswith("A-"):
        aid = f"A-{aid}" if aid.isdigit() else aid
    # normalize A-1 -> A-0001 if possible
    m = re.match(r"A-(\d+)$", aid)
    if m:
        aid = f"A-{int(m.group(1)):04d}"

    attacks = load_all_attacks()
    if aid not in attacks:
        print(f"Unknown attack {aid}", file=sys.stderr)
        return 1
    if args.status not in STATUSES:
        print(f"Bad status; use {STATUSES}", file=sys.stderr)
        return 1
    if args.method not in METHODS:
        print(f"Bad method; use {METHODS}", file=sys.stderr)
        return 1

    loci = list(args.loci or [])
    if args.status in ("answered", "partial") and not loci:
        print("answered/partial require --loci", file=sys.stderr)
        return 1

    state = load_state()
    cycle = (state.get("active_unit") or {}).get("cycle", 0)
    doc = {
        "attack": aid,
        "status": args.status,
        "method": args.method,
        "response_loci": loci,
        "note": args.note or "",
        "cycle": cycle,
        "updated": date.today().isoformat(),
    }
    path = RULINGS_DIR / f"R-{aid[2:]}.yaml"
    dump_yaml(path, doc)
    print(f"Wrote {path.relative_to(ROOT)}")
    return 0


def cmd_new_cycle(args: argparse.Namespace) -> int:
    state = load_state()
    active = state.setdefault("active_unit", {})
    chapter = args.chapter or active.get("chapter")
    if not chapter:
        print("No chapter; pass --chapter", file=sys.stderr)
        return 1
    cycle = int(active.get("cycle") or 0) + 1
    max_cycles = int(active.get("max_cycles") or 2)
    if cycle > max_cycles and not args.force:
        print(
            f"Cycle {cycle} exceeds max_cycles={max_cycles}. Use --force if a critical remains.",
            file=sys.stderr,
        )
        return 1
    active["chapter"] = chapter
    active["cycle"] = cycle
    active["phase"] = "attack"
    save_state(state)

    path = CYCLES_DIR / f"{chapter}_c{cycle:02d}.md"
    body = f"""# Cycle {cycle} — {chapter}

Date: {date.today().isoformat()}
Note: {args.note or ""}

## Attackers this cycle

- (list agents)

## Dedupe / ranking

-

## Author revision summary

- Positive clarity changes:
- Fences added:
- Cases added:

## Open after rulings

-

## Energy check (human)

- [ ] 60-second retell without naming the enemy
- [ ] Positive ratio acceptable
"""
    path.write_text(body, encoding="utf-8")
    print(f"Cycle {cycle} started for {chapter}")
    print(f"Wrote {path.relative_to(ROOT)}")
    return 0


def cmd_freeze(args: argparse.Namespace) -> int:
    # run check; disallow open critical
    rc = cmd_check(argparse.Namespace())
    if rc == 1:
        print("Cannot freeze: mechanical check failed", file=sys.stderr)
        return 1
    attacks = load_all_attacks()
    rulings = load_all_rulings()
    chapter = args.chapter or (load_state().get("active_unit") or {}).get("chapter")
    for aid, atk in attacks.items():
        if atk.get("chapter") and atk.get("chapter") != chapter:
            continue
        if chapter and not atk.get("chapter"):
            pass
        if atk.get("severity") == "critical":
            st = (rulings.get(aid) or {}).get("status")
            if st not in ("answered", "wontfix"):
                print(f"Cannot freeze: open critical {aid}", file=sys.stderr)
                return 1
        if atk.get("against") == "NEW":
            st = (rulings.get(aid) or {}).get("status")
            if st not in ("wontfix", "answered"):
                print(f"Cannot freeze: unretargeted NEW attack {aid}", file=sys.stderr)
                return 1

    state = load_state()
    active = state.setdefault("active_unit", {})
    active["chapter"] = chapter
    active["phase"] = "frozen"
    frozen = state.setdefault("frozen_units", [])
    entry = {
        "chapter": chapter,
        "date": date.today().isoformat(),
        "note": args.note or "",
    }
    frozen = [u for u in frozen if (u.get("chapter") if isinstance(u, dict) else u) != chapter]
    frozen.append(entry)
    state["frozen_units"] = frozen
    save_state(state)

    claims_path = CLAIMS_DIR / f"{chapter}.yaml"
    if claims_path.exists():
        doc = load_yaml(claims_path) or {}
        doc["status"] = "frozen"
        dump_yaml(claims_path, doc)

    print(f"Frozen unit: {chapter}")
    print("Remember human gates: energy check + positive ratio.")
    return 0


def cmd_agents(_: argparse.Namespace) -> int:
    print("Fixed agents (one side each):")
    for a in sorted(AGENTS):
        marker = ""
        if a == "scripturalist":
            marker = "  # Christian/theistic anti-correspondence epistemology"
        elif a == "muslim_package":
            marker = "  # Islamic doctrinal-institutional packages (not all persons)"
        print(f"  {a}{marker}")
    return 0


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="book1-5 adversarial harness")
    sub = p.add_subparsers(dest="cmd", required=True)

    sub.add_parser("status", help="Show open attacks and active unit").set_defaults(
        func=cmd_status
    )
    sub.add_parser("unanswered", help="List attacks not answered/wontfix").set_defaults(
        func=cmd_unanswered
    )
    sub.add_parser("check", help="Validate claims/attacks/rulings links").set_defaults(
        func=cmd_check
    )
    sub.add_parser("agents", help="List fixed agent masks").set_defaults(func=cmd_agents)

    pc = sub.add_parser("claims", help="List claims for a chapter")
    pc.add_argument("--chapter", required=True)
    pc.set_defaults(func=cmd_claims)

    pa = sub.add_parser("new-attack", help="Create a new attack YAML stub")
    pa.add_argument("--agent", required=True)
    pa.add_argument("--against", required=True, help="Claim id or NEW")
    pa.add_argument("--severity", default="major", choices=sorted(SEVERITIES))
    pa.add_argument("--mode", default="logic", choices=sorted(MODES))
    pa.add_argument("--chapter")
    pa.add_argument("--id")
    pa.add_argument("--summary")
    pa.add_argument("--quote")
    pa.add_argument("--source_ref")
    pa.add_argument("--subversion")
    pa.set_defaults(func=cmd_new_attack)

    pr = sub.add_parser("rule", help="Write a ruling for an attack")
    pr.add_argument("attack_id")
    pr.add_argument("--status", required=True, choices=sorted(STATUSES))
    pr.add_argument("--method", required=True, choices=sorted(METHODS))
    pr.add_argument("--loci", action="append", default=[], help="Repeatable locus")
    pr.add_argument("--note", default="")
    pr.set_defaults(func=cmd_rule)

    pn = sub.add_parser("new-cycle", help="Increment cycle and write cycle log")
    pn.add_argument("--chapter")
    pn.add_argument("--note", default="")
    pn.add_argument("--force", action="store_true")
    pn.set_defaults(func=cmd_new_cycle)

    pf = sub.add_parser("freeze", help="Freeze a unit if gates pass")
    pf.add_argument("--chapter")
    pf.add_argument("--note", default="")
    pf.set_defaults(func=cmd_freeze)

    return p


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return int(args.func(args))


if __name__ == "__main__":
    sys.exit(main())
