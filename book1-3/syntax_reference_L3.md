# Syntax Reference for Book1-3

This document defines the consistent annotation syntax used throughout all book1-3 documents for linking ideas, references, marking hierarchy, and documenting transitions.

---

## Overview

The annotation syntax is designed to be:
- **Readable** in markdown without breaking rendering
- **Consistent** across all documents
- **Parseable** for automation
- **YAML-compatible** for metadata

---

## Annotation Types

### 1. Idea Links

Format: `[idea:ID]`

Links to a specific idea, concept, or principle in the index.

- `ID`: The categorical index number (e.g., C001, E001, P015)

**Examples:**
```markdown
[L1] Reality is Real, knowable, and consequential. [idea:C001]

Procedural Symmetry requires applying the same rules to all. [idea:P001]
```

---

### 2. Reference Material Links

Format: `[ref:SOURCE]`

Links to a reference source such as a historical figure, book, or document.

- `SOURCE`: A short, descriptive identifier for the reference

**Examples:**
```markdown
Consider Trofim Lysenko, who rejected genetics. [ref:Lysenko]

The principle is similar to Newton's approach to gravity. [ref:Newton]

This aligns with the biblical teaching. [ref:Matthew:7:12]
```

---

### 3. Idea Hierarchy Levels

Format: `[L1]`, `[L2]`, `[L3]`

Marks the hierarchical level of an idea within the outline structure.

- `[L1]`: Primary (core) ideas - main sections or fundamental principles
- `[L2]`: Secondary ideas - subsections or elaborations of L1 concepts
- `[L3]`: Tertiary ideas - detailed points or specific examples within L2

**Usage:**
- Place at the beginning of an outline item
- Can be used multiple times in nested structures

**Examples:**
```markdown
[L1] Reality is Real, knowable, and consequential.
  [L2] Existence has primacy over consciousness.
  [L2] Reality operates independently of our hopes.
    [L3] A dropped ball falls regardless of your thoughts.

[L1] Tiered Self-Anchored Responsibility.
  [L2] Responsibility begins with self.
  [L2] Responsibility radiates outward in concentric circles.
```

---

### 4. Transition Placement

Format: YAML frontmatter at the start of each section

Documents how sections connect to each other and the nature of the relationship.

**Syntax:**
```yaml
---
section: "Section Title"
transitions:
  - to: "Target Section"
    type: "builds_on|leads_to|contrasts_with|clarifies|exemplifies"
    note: "Brief explanation of the transition"
---
```

**Transition Types:**
- `builds_on`: The current section builds upon the foundation laid by the target
- `leads_to`: The current section logically leads into the target
- `contrasts_with`: The current section provides a contrast to the target
- `clarifies`: The current section clarifies or expands upon the target
- `exemplifies`: The current section provides examples of the target

**Examples:**
```markdown
---
section: "Part 1: Foundation"
transitions:
  - to: "Introduction"
    type: "builds_on"
    note: "Establishes the foundational axioms introduced"
  - to: "Part 2: Individual"
    type: "leads_to"
    note: "Foundation enables the development of individual virtues"
---

# Part 1: Foundation
...
```

---

### 5. Idea Ties and Connections

Format: `[tie:ID1|ID2|ID3]`

Indicates that multiple ideas are related or should be considered together.

**Usage:**
- Place inline where the relationship is relevant
- Can connect any number of related ideas

**Examples:**
```markdown
This concept of tiered responsibility [tie:C010|C011|P005] connects to the broader principle of generational obligation.

The balance between love and forgiveness [tie:C020|C021|C025] requires careful judgment.
```

---

## Categorical Index Codes

All ideas and references in the index are categorized with the following codes:

| Code | Category | Description |
|------|----------|-------------|
| C | Concepts | Abstract ideas, definitions, philosophical terms |
| E | Examples | Illustrative cases, stories, case studies |
| R | References | Historical figures, authors, texts, citations |
| P | Principles | Actionable guidelines, rules, procedures |
| D | Distinctions | Contrasting concepts, comparisons |
| Q | Quotes | Direct quotations or citations |

### Numbering Format
- Format: `CATEGORY###` where `###` is sequential (001, 002, 003...)
- Example: `C001`, `E015`, `R002`, `P001`

---

## Section Structure Template

When creating outline sections, use this template:

```markdown
---
section: "Section Title"
transitions:
  - to: "Previous/Next Section"
    type: "relationship_type"
    note: "Explanation"
---

# Section Title

[L1] Primary idea with main concept. [idea:ID]

  [L2] Secondary idea elaboration. [idea:ID]

    [L3] Tertiary detail or example. [idea:ID] [ref:SOURCE]

  [L2] Another secondary point. [tie:ID1|ID2] [idea:ID]
```

---

## Complete Example

```markdown
---
section: "Part 1: Foundation"
transitions:
  - to: "Introduction"
    type: "builds_on"
    note: "Establishes the foundational axioms of reality and truth"
  - to: "Part 2: Individual"
    type: "leads_to"
    note: "Foundation provides the basis for individual virtues"
---

# Part 1: Foundation

[L1] Reality is Real, knowable, and consequential. [idea:C001]

  [L2] Existence has primacy over consciousness. [idea:C002]

    [L3] Consciousness develops through interaction with pre-existing reality. [tie:C001|C002]

  [L2] Truth is what corresponds to reality. [idea:C003]

    [L3] Truth is knowable but our knowledge is provisional. [tie:C003|C004]

[L1] Humans are limited, selfish, and not naturally good. [idea:C005]

  [L2] Recognizing human limits is the "constrained vision." [ref:Sowell]

  [L2] Goodness must be taught. [tie:C005|C006]
```

---

## Usage Guidelines

1. **Be Consistent**: Always use the same format for the same type of annotation
2. **Be Specific**: When referencing ideas, use the most specific ID available
3. **Be Clear**: Use descriptive identifiers in `[ref:SOURCE]` links (e.g., `ref:Newton` not `ref:N`)
4. **Be Complete**: Document all transitions between sections in the YAML frontmatter
5. **Be Minimal**: Only annotate when necessary; don't over-use links

---

## Notes for Automation

When parsing these annotations:

1. Idea links follow the pattern `\[idea:(C|E|R|P|D|Q)\d{3}\]`
2. Reference links follow the pattern `\[ref:[^\]]+\]`
3. Hierarchy tags are `\[L[123]\]`
4. Ties follow the pattern `\[tie:(C|E|R|P|D|Q)\d{3}(?:\|(C|E|R|P|D|Q)\d{3})*\]`
5. YAML frontmatter is enclosed in `---` delimiters at the start of sections

---

## Version History

- v1.0: Initial syntax definition for Book1-3 project
