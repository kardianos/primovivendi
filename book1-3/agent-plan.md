# Book1-3 Agent Plan

## Overview

This plan outlines a multi-phase effort to annotate, index, and expand the book outline from L1 to L3. The goal is to create a structured, cross-referenced system that tracks all ideas, concepts, examples, and references across the book1-1 chapters and integrates them into progressively detailed outlines.

---

## Phase 1: Syntax Reference Document

**File:** `book1-3/syntax_reference.md`

### Objectives
1. Define consistent annotation syntax for all book1-3 documents
2. Establish conventions for linking, hierarchy marking, and transition notation

### Syntax Definitions

#### Idea Links
- Format: `[idea:ID]` where ID is the categorical index number
- Example: `[idea:C001]` links to concept C001

#### Reference Material Links
- Format: `[ref:SOURCE]` where SOURCE is a short identifier
- Example: `[ref:Lysenko]` links to the Lysenko historical reference
- Sources should be indexed in the reference section

#### Idea Hierarchy Levels
- Format: `[L1]`, `[L2]`, `[L3]` for primary, secondary, and tertiary ideas
- These tags are placed at the beginning of outline items
- Example: `[L1] Reality is Real, knowable, and consequential`

#### Transition Placement
- Format: Section-level metadata at the start of each section
- Syntax:
  ```yaml
  ---
  section: "Section Title"
  transitions:
    - to: "Target Section"
      type: "builds_on|leads_to|contrasts_with|clarifies|exemplifies"
      note: "Brief explanation of the transition"
  ---
  ```

- Transition types:
  - `builds_on`: The current section builds upon the foundation laid by the target
  - `leads_to`: The current section logically leads into the target
  - `contrasts_with`: The current section provides a contrast to the target
  - `clarifies`: The current section clarifies or expands upon the target
  - `exemplifies`: The current section provides examples of the target

#### Idea Ties and Connections
- Format: `[tie:ID1|ID2|ID3]` for indicating related ideas
- Example: `[tie:C001|C002]` indicates these concepts are related

### Categorical Index Codes

All ideas and references in the index are categorized with the following codes:

| Code | Category | Description | Examples |
|------|----------|-------------|----------|
| C | Concepts | Abstract ideas, definitions, philosophical terms | Reality, Truth, Responsibility |
| E | Examples | Illustrative cases, stories, case studies | Lysenko, Moriori, Good Samaritan |
| R | References | Historical figures, authors, texts, citations | Newton, Rousseau, Bible verses |
| P | Principles | Actionable guidelines, rules, procedures | Procedural Symmetry, Tiered Responsibility |
| D | Distinctions | Contrasting concepts, comparisons | Dispositional vs Action-Based Love |
| Q | Quotes | Direct quotations or citations | Various inline quotes |

### Numbering Format
- Format: `CATEGORY###` where `###` is sequential (001, 002, 003...)
- Example: `C001`, `E001`, `R001`, `P001`, `D001`, `Q001`

### Section Structure Template

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

## Phase 2: Book1-1 Index Creation

**File:** `book1-3/index_book1-1.md` (TSV format)

### Objectives
1. Read all chapters in `book1-1/chapters/*.md`
2. Extract and categorize every idea, tie, concept, and example
3. Assign categorical index numbers
4. Create TSV format for easy parsing

### Categorization Scheme

| Category Code | Description | Examples |
|--------------|-------------|----------|
| C | Concepts | Abstract ideas, definitions, philosophical terms | Reality, Truth, Responsibility |
| E | Examples | Illustrative cases, stories, case studies | Lysenko, Moriori, Good Samaritan |
| R | References | Historical figures, authors, texts, citations | Newton, Rousseau, Bible verses |
| P | Principles | Actionable guidelines, rules, procedures | Procedural Symmetry, Tiered Responsibility |
| D | Distinctions | Contrasting concepts, comparisons | Dispositional vs Action-Based Love |
| Q | Quotes | Direct citations | Various inline quotes |

### Index Numbering Format
- Format: `CATEGORY###` where `###` is sequential (001, 002, 003...)
- Numbering is global across all chapters
- Example: `C001`, `C002`, `E001`, `R001`, `P001`...

### TSV Structure
```
ID	Chapter	Title	Full Text	Location	Usage
C001	01_introduction	Reality is real	Reality is real, consequential, and knowable.	ch01:1	[used:L2:Part1:1]
C002	01_introduction	Humans are limited	Humans are limited and selfish.	ch01:2	[used:L2:Part1:2]
...
```

### Fields
- `ID`: Categorical index (C001, E001, etc.)
- `Chapter`: Source chapter filename
- `Title`: Brief title for the item
- `Full Text`: Complete text of the idea/example/reference
- `Location`: Chapter identifier and line number (ch01:1 = Chapter 1, line 1)
- `Usage`: Where this index is used (added during Phase 3)

### Process
1. Read all 22 chapters in `book1-1/chapters/`
2. Parse each line for:
   - Concept definitions (sentences starting with definitions, bold terms)
   - Examples (names, case studies, illustrative stories)
   - References (historical figures, philosophers, authors, Bible verses)
   - Principles (clear actionable guidelines)
   - Distinctions (comparisons between concepts)
   - Quotes (direct citations)
3. Extract full text context
4. Assign categorical IDs
5. Compile into TSV format
6. Save as `book1-3/index_book1-1.md`

---

## Phase 3: L2 Outline Generation

**Files:**
- `book1-3/outline/outline_L2.md` (main outline)
- `book1-3/outline/unused_L2.md` (unused indices)

### Objectives
1. Generate `outline_L2.md` from `outline_L1.md`
2. Pull in all numbered indices from book1-1 index
3. Place indices in the correct outline location
4. Track which indices are used and which are not
5. Annotate the index with usage locations

### Process

#### Step 1: Read `outline_L1.md`
- Parse the existing structure
- Identify all sections and subsections

#### Step 2: Map Indices to Outline Locations
For each index entry:
- Determine the most appropriate location in the outline
- Match based on:
  - Part/Section thematic alignment
  - Concept category
  - Chapter source context
- If no clear match, mark as "uncategorized"

#### Step 3: Generate `outline_L2.md`
1. Start with `outline_L1.md` content
2. For each outline item:
   - Add section-level transition metadata (YAML frontmatter)
   - Add `[L1]`, `[L2]`, `[L3]` hierarchy tags
   - Insert relevant `[idea:ID]` links where appropriate
   - Add `[ref:SOURCE]` links where appropriate

Example structure:
```markdown
---
section: "Part 1: Foundation"
transitions:
  - to: "Introduction"
    type: "builds_on"
    note: "Establishes foundational axioms"
---

# Part 1: Foundation

[L1] Reality is Real, knowable, and consequential. Truth is what corresponds to reality. [idea:C001]

[L1] Humans are limited, selfish, and not naturally good. [idea:C002]
```

#### Step 4: Track Index Usage
For each index entry:
- Record in the index where it's used: `[used:L2:Part1:1]`
- If an index is not used in L2, mark as `[unused]`

#### Step 5: Generate `unused_L2.md`
- List all indices marked as `[unused]`
- Organize by category
- Provide context why it might not fit in current outline

### Index Annotation Format
In `index_book1-1.md`, add column for usage:
```
ID	Usage
C001	[used:L2:Part1:1, L2:Part3:2]
C002	[unused]
```

---

## Phase 4: L2 Adjustments

**File:** `book1-3/adjustments_L2.md`

### Objectives
1. Create a document for additions/modifications to the L2 outline
2. Structure allows for iterative improvements
3. Maintain traceability of changes

### Format
```markdown
# Adjustments to L2 Outline

## Additions

### New Section: [Section Name]

Add to [location]:

```markdown
[L1] Section title
  [L2] Subsection
    [L3] Detail
```

## Modifications

### Modify: [Section Name]

[Description of modification]

**Original:**
```markdown
[Original content]
```

**Modified to:**
```markdown
[Modified content]
```
```

---

## Phase 5: L3 Outline Generation

**File:** `book1-3/outline/outline_L3.md`

### Objectives
1. Combine `outline_L2.md` with `adjustments_L2.md`
2. Generate final detailed outline
3. Ensure all transitions are documented
4. Verify all indices are properly linked

### Process

#### Step 1: Merge L2 and Adjustments
1. Start with `outline_L2.md` content
2. Apply all additions from `adjustments_L2.md`
3. Apply all modifications
4. Apply all reorganizations

#### Step 2: Update Index Usage
- Mark any newly used indices
- Update `index_book1-1.md` usage column

#### Step 3: Verify Completeness
- Check all section transitions
- Verify all `[idea:ID]` links are valid
- Ensure hierarchy tags `[L1]`/`[L2]`/`[L3]` are consistent

#### Step 4: Final Format
- Ensure YAML frontmatter for each section
- Verify all syntax is correct

---

## Phase 6: Interactive Prompts and Iteration

### Objectives
1. Prompt user for each part and section
2. Gather suggestions, missing pieces, concept clarifications
3. Iterate based on feedback

### Prompt Structure

#### For Each Part:
```
=== PART N: [Part Title] ===

Current outline items:
- [L1] Item 1
- [L2] Item 1.1
- [L2] Item 1.2

Questions:
1. Are there any missing concepts or ideas?
2. Should any item be moved, split, or combined?
3. Are the transitions clear? Any additional connections needed?
4. Are there examples from book1-1 that should be referenced?
5. Any clarifications needed on existing items?

Suggestions and notes:
[User input]
```

#### For Section Transitions:
```
=== TRANSITION: [Previous Section] → [Current Section] ===

Current transition notes:
[type: builds_on, note: Establishes foundational axioms]

Questions:
1. Is this transition accurate?
2. Are there additional connections to document?
3. Should the transition type be different?

Refined transition:
[User input]
```

### Iteration Process
1. Present section with current outline
2. Ask targeted questions
3. Collect user input
4. Update outline/adjustments
5. Move to next section
6. After full review, generate final L3 outline

---

## Phase 7: Final Documentation

### Files to Create
1. `book1-3/agent-plan.md` (this document)
2. `book1-3/syntax_reference.md` (Phase 1)
3. `book1-3/index_book1-1.md` (Phase 2)
4. `book1-3/outline/outline_L2.md` (Phase 3)
5. `book1-3/outline/unused_L2.md` (Phase 3)
6. `book1-3/adjustments_L2.md` (Phase 4)
7. `book1-3/outline/outline_L3.md` (Phase 5)

### Validation Checklist
- [ ] All syntax in `syntax_reference.md` is consistent
- [ ] All book1-1 chapters are indexed in `index_book1-1.md`
- [ ] TSV format is correctly structured
- [ ] All indices have categorical IDs
- [ ] `outline_L2.md` includes all section transitions
- [ ] All indices from book1-1 are either in L2 or unused_L2
- [ ] `adjustments_L2.md` clearly documents all changes
- [ ] `outline_L3.md` is a clean merge of L2 and adjustments
- [ ] All `[idea:ID]` links are valid
- [ ] All section transitions are documented in YAML frontmatter

---

## Implementation Order

1. Create `book1-3/syntax_reference.md`
2. Parse all book1-1 chapters and create `book1-3/index_book1-1.md`
3. Generate `book1-3/outline/outline_L2.md` and `book1-3/outline/unused_L2.md`
4. Interactive review prompts for each section (Part 1-5)
5. Create `book1-3/adjustments_L2.md` during interactive review
6. Generate final `book1-3/outline/outline_L3.md`
7. Final validation and documentation

---

## Notes

- All work is contained within `book1-3/` directory
- No modifications to `book1-1/` or other directories
- Use Go template syntax compatibility where existing ({{.title}}, etc.)
- Maintain markdown formatting consistency
- Ensure all generated files are human-readable and version-control friendly

---

## Completion Summary

### Files Created:
1. ✓ `book1-3/syntax_reference.md` - Complete annotation syntax (YAML frontmatter style)
2. ✓ `book1-3/index_book1-1.md` - 180+ indexed items (C, E, R, P, D, Q categories) in TSV format
3. ✓ `book1-3/outline/outline_L2.md` - Fully annotated outline with transitions and hierarchy
4. ✓ `book1-3/outline/unused_L2.md` - Tracking indices not used in L2
5. ✓ `book1-3/adjustments_L2.md` - Comprehensive additions and modifications
6. ✓ `book1-3/outline/outline_L3.md` - Final merged outline (1,762 lines, 7 sections, 33 L1 items)

### Statistics:
- **Total indexed items:** ~180
- **Categories:** Concepts (C), Examples (E), References (R), Principles (P), Distinctions (D), Quotes (Q)
- **Outline structure:** 7 main sections with YAML transitions
- **Hierarchy levels:** L1 (33 primary), L2 (secondary), L3 (tertiary)
- **Syntax tags:** 766 total links (idea/ref/tie/example/quote/distinction)

### Key Features:
- Pre-registered medical study comparison
- Historical examples (US Founding, Soviet Union, Singapore, Nordic Model, California, Rome, etc.)
- Child activist examples (Greta, Parkland, state standards, Hitler Youth)
- Gender studies and career preference examples
- Faith-based generational obligation
- Divine constructivism framework
- Social outliers with tiered judgment
- Rulership concept
- Comprehensive thought experiments
- Five concrete actions for living procedurally
- Complete integration of all individual and societal principles

### Interactive Review Process:
✓ All 5 parts reviewed for suggestions, missing pieces, and clarifications
✓ Adjustments documented in adjustments_L2.md
✓ L3 outline generated from merged content
✓ All syntax validated (766 links, 33 L1 items, 7 YAML transitions)
