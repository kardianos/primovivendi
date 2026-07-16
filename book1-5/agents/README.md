# Agent prompt packs

Each file is a **fixed mask** for attack turns. Copy the pack into an agent session; do not let the agent rewrite the pack to soften itself.

## Usage

1. Author has a positive draft + claim map.  
2. Operator opens one agent with **only** that pack + the chapter text + claim IDs.  
3. Agent outputs YAML attacks (or dictation for `harness.py new-attack`).  
4. Agent never edits `chapters/`.

## Roster

| File | Lane |
|------|------|
| `constructivist.md` | Power/discourse truth |
| `outcome_equity.md` | Equal outcomes as justice |
| `universalist.md` | Anti-tier cosmopolitanism |
| `scripturalist.md` | Christian/theistic anti-correspondence epistemology |
| `muslim_package.md` | Islamic doctrinal-institutional packages |
| `blank_slate.md` | Perfectible humans |
| `near_rival.md` | Overlap friend, anti-sloppiness |
| `rhetorical_predator.md` | Quarantined bad faith |
| `nihilist_anti_thrival.md` | Optional anti-life |

Scripturalist and muslim_package are **different agents** with different sources and different target principles.
