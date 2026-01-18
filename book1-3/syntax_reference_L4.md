# Syntax Reference for Book1-3 (Level 4: Detailed Outline)

This document defines the annotation syntax used in Level 4 outline files (`outline_L4-x.txt`). This syntax is designed for high-density information structure, supporting granular referencing and type tagging.

---

## 1. Item Syntax

The core format for every outline item is:

`{ID|Type:Key} Content`

or, if no Key is required:

`{ID|Type} Content`

### Components:

*   **ID**: The hierarchical identifier (e.g., `1.1`, `4.3.3.1`).
*   **Type**: The semantic category of the item (e.g., `idea`, `example`).
*   **Key** (Optional): A cross-reference ID or descriptive slug (e.g., `C007`, `system_integrity`).
*   **Content**: The actual text of the outline item.

---

## 2. Item Types

| Type | Description | Usage |
|------|-------------|-------|
| `part` | Top-level division | `{1.0|part} Part 1: Foundation` |
| `chapter` | Major section header | `{1.1|chapter} Reality is Real` |
| `section` | Sub-section header | `{4.3.2|section} The Compatible Core` |
| `name` | Name of a specific entity | `{4.3.2.1|name} John Locke` |
| `idea` | A core concept or principle | `{1.1.1|idea:C007} Reality is real.` |
| `tie` | Connects ideas/principles | `{1.3.1|tie:C003} Responsibility begins...` |
| `explanation` | Clarifies the preceding item | `{1.6.3.1|explanation} If you cling to...` |
| `example` | Illustrative instance | `{1.1.4.1|example:E001} A bridge built on...` |
| `application` | Practical application | `{1.6.2|application} Reality check: Does...` |
| `ref` | External reference/Source | `{1.2.7.1|ref:Rousseau} This view opposes...` |
| `quote` | Direct quotation | `{4.3.3.1b|quote} "Scripturalism denies..."` |
| `distinction` | Contrasting concept | `{3.2.3|distinction:D006} The choice is...` |
| `insight` | Key takeaway or synthesis | `{4.1.7|insight} Reason untethered...` |
| `analysis` | Analytical breakdown | `{1.6.3|analysis} Angle 1: Goodness...` |
| `consequence` | Logic flow consequence | `{1.6.3.2|consequence} This leads to...` |
| `result` | Final outcome | `{1.6.3.3|result} Result: Disappointment.` |
| `assertion` | Strong positive claim | `{4.4.5.3|assertion} Reality is REAL.` |

---

## 3. ID Structure used in L4

*   **Numeric Hierarchical**: `Part.Chapter.Section.Item` (e.g., `1.6.3.1`).
*   **Alphabetic Suffix**: Used for sub-points within a dense logic block where full numbering differs (e.g., `4.3.3.1a`, `4.3.3.1b`).

---

## 4. Keys and References

Keys provide cross-linking to the Categorical Index (L3) or serve as local handles.

*   **Categorical IDs**: `C` (Concept), `P` (Principle), `E` (Example), `D` (Distinction), `Q` (Quote).
    *   Example: `{1.1.3|idea:C008}`
*   **Descriptive Slugs**: Short text strings for un-cataloged concepts.
    *   Example: `{1.6.3|analysis:Utopian_Trap}`
*   **External Refs**: Source names.
    *   Example: `{4.4.7|ref:scripturalism.xhtml}`

---

## 5. Usage Guidelines

1.  **One Thought Per Line**: Break complex paragraphs into individual, numbered logic steps.
2.  **Explicit Typing**: Every line must have a type. Do not leave lines un-annotated.
3.  **Hierarchy**: Use the ID to strictly enforce logical nesting.
4.  **Consistency**: Use standard keys (C001, etc.) whenever possible to maintain linkage with L3 artifacts.

---

## 6. Example Snippet

```text
{4.4.5|idea:constructivism_def} The Core Error of "Constructivism": Reality as Mind-Dependent.
{4.4.5.1|explanation} "Constructivism" is the belief that because a thing differs by culture or origin, it is merely a product of Mind.
{4.4.6|idea:C099} Variant A: Social Constructivism (The Collective Mind).
{4.4.6.1|tie:C099} Claims that categories like gender or hierarchy are purely products of social language.
{4.4.7|ref:scripturalism.xhtml} Variant B: Divine Constructivism (The Divine Mind).
{4.4.7.1|ref:scripturalism.xhtml} Claims that Truth is exclusively "The Mind of God".
```
