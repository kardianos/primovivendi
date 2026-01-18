# Instructions for creating Outline L4

## Goal
Create a new set of outline files `outline_L4-0.txt` to `outline_L4-5.txt` in the `book1-3/outline/` directory.
Each file corresponds to a "Part" from `book1-3/outline/outline_L3.md`.

## Files Mapping
- `outline_L4-0.txt` corresponds to **Introduction** (Part 0)
- `outline_L4-1.txt` corresponds to **Part 1: Foundation**
- `outline_L4-2.txt` corresponds to **Part 2: Individual**
- `outline_L4-3.txt` corresponds to **Part 3: Society**
- `outline_L4-4.txt` corresponds to **Part 4: Deeper Grounding**
- `outline_L4-5.txt` corresponds to **Part 5: Distinctions**

## Input Sources
1.  **Structure**: Use `book1-3/outline/outline_L3.md` as the primary guide.
2.  **Content**: Use `book1-3/index_book1-1.md` to look up the text for references (IDs start with `C`, `E`, `Q`, `P`, `D`, `R`).

## Output Format
- **Text on every line**: No empty lines.
- **Identifier Format**: `{identifier|label} Text` or `{identifier} Text`
    - Example: `{2.4-0|name} Take Responsibility`
    - Ref Example: `{2.4-1.1|idea:C015} Truth is what corresponds to reality.`
- **Transitions**: `{identifier|transition} Transitional text...`
- **Correction**: Correct spelling/grammar.
- **Classification**: Classify non-reference points yourself (e.g., `{...|explanation}`).

## Instructions
1.  Read `outline_L3.md` for the specific Part.
2.  For each point, construct an identifier (e.g., `Part.L1.L2.L3`).
3.  If it references an ID in `index_book1-1.md`, fetch the text.
4.  If it's standalone, use the text from `outline_L3.md`.
5.  Insert transitions where needed for flow.
6.  Output strictly detailed text with the `{...}` prefix.
