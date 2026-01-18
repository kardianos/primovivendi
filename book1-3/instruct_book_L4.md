# L4 Chapter Writing Instructions

## 1. Overview
This document outlines the procedure for transforming the Level 4 Outlines (`outline_L4-*.txt`) into full book chapters in `book1-3/chapters_L4/`. The goal is to produce "book quality" prose that expands upon the dense outline logic, integrating definitions and references from the Index (`index_book1-1.md`).

## 2. Input Sources
*   **Outline Files**: `book1-3/outline/outline_L4-*.txt` (Primary structural source).
*   **Index**: `book1-3/index_book1-1.md` (Definitions of Concepts `C`, Principles `P`, etc.).
*   **Syntax Reference**: `book1-3/syntax_reference_L4.md` (Guide to reading the outline annotations).

## 3. Output Location & Naming
All output files must be written to: `book1-3/chapters_L4/`

### File Naming Convention
format: `PNN_Slug.md`
*   `P`: Part number (e.g., `4` for Part 4, `0` for Frontmatter).
*   `NN`: Sequential file number within that Part (01-indexed).
*   `Slug`: 1-3 words describing the content, separated by hyphens. NO "ch01" style generic names.

**Examples:**
*   `001_Meta-Header.md` (Part 0, File 1: YAML Meta header only).
*   `101_Reality-Is-Real.md` (Part 1, File 1: Chapter 1.1).
*   `403_Compatible-Philosophies.md` (Part 4, File 3: Chapter 4.3).

## 4. Work Process
1.  **Read**: Parse the `outline_L4-x.txt` file.
2.  **Lookup**: For every bracketed key (e.g., `{...|idea:C007}`), look up the content in `index_book1-1.md`.
3.  **Draft**: Write the content in the target file.
4.  **Verify**: Ensure tone and formatting match strict requirements.

## 5. Structural Rules

### Headers
*   **Part Header**: Level 1 (`# Part X: Title`).
    *   *Rule*: This appears ONLY in the first file of that Part (e.g., `101_...md`, `401_...md`).
*   **Chapter Header**: Level 2 (`## X.Y Title`).
    *   *Rule*: Every Chapter starts with `## [Part].[Num] [Title]`.
    *   *Example*: `## 1.1 Reality is Real, Knowable, and Consequential`.
*   **Section Header**: Level 3 (`### Title`).
    *   *Rule*: Use for items marked `|section}` in the outline.

### Content Logic
*   **Expansion**: The outline provides the *logic* and *skeleton*. You must expand this into full sentences and paragraphs.
*   **Integration**: When an outline line references an ID (e.g., `C007`), you MUST incorporate the definition and nuances of that concept from the Index into the text. Do not just copy-paste; weave it into the narrative.
*   **One File Per Chapter**: Generally, one outline chapter `{X.Y|chapter}` equals one output file `PNN_Slug.md`.

## 6. Style & Tone Guidelines
**Primary Tone**: Authoritative, serious, direct.
*   **Book Quality**: This is not a blog post. No "Welcome to this chapter" or "In this section we will discuss". Just state the truth.
*   **No Cheap Metaphors**:
    *   **Rule**: Do not use metaphorical titles or descriptions for people or concepts. Use precise, denotative English.
    *   **Bad**: "Thomas Reid is the Jurist of Reality." / "Rawls is the Siren Song." / "Crooked Timber."
    *   **Good**: "Thomas Reid provides the epistemological defense of Real Reality." / "Rawls mimics the language of the Compact." / "Inherent human flaws."
    *   **Reasoning**: English is a language capable of conveying specific meaning. Metaphors blur specificity. Say exactly what the thing *is* or *does*, not what it is "like."
    *   NO: "We are on a journey together."
    *   YES: "Reality is the territory we navigate."
*   **Punctuation Restrictions**:
    *   **NO DASHES** (`-` or `–` or `—`) in the prose.
    *   **NO EM-DASHES**.
    *   **SEMICOLONS**: Use sparingly. Prefer strong periods.

## 7. Special Files
*   **001_Meta.md**: This file contains ONLY the YAML header for the concatenation script.
    ```yaml
---
title: "Tiered Procedural Symmetry"
titleShort: "TPS"
version: "1.2"
subtitle: "Common Principles Toward a Reality-Based Life"
author: "Daniel Theophanes"
date: "2025-01-15"
geometry: margin=3cm
toc: true
toc-depth: 2
---
```
    (Do not add body text to this file. The header will provide the title and will be joined together with the chapters to form the final book.)

## 8. Example Transformation

**Source (Outline)**:
```text
{1.1.4|idea:C010} Reality operates independently of our hopes and feelings.
{1.1.4.1|example:E001} A bridge built on flawed physics will collapse.
```

**Source (Index)**:
`C010`: Reality is consequential. Reality operates independently of our hopes and feelings.
`E001`: A bridge built on flawed physics will collapse.

**Output (Draft)**:
> Reality is consequential because it acts independently of our desires. It does not adjust itself to accommodate our feelings or preferences. A bridge built on flawed physics will collapse regardless of how enthusiastically we celebrate its construction.

**(Note: Strong assertions. No "I think". No dashes.)**
