# Chapter Instructions (book1-5)

## File naming

`chapters/<00part><00chapter>_<slug>.md`

- Exception: `0000_title_block.md`
- Lexical order of filenames = book order

## Prose rules

1. **Literal declarative prose.** Prefer scene and definition over structural meta-commentary about “rings” or the harness.
2. **Direct, not meta (about-about).** Name the thing itself, not the writing job or the essay’s length and machinery.
   - **Titles:** prefer a plain map of the content (*The Five Principles in Full*, *Defensive Distinctions*, *The Precepts Matrix*) over process or craft labels (*Unpacking the Five*, *…Revisited*, *Engaged Directly*, *The Tool Made Explicit*, *What the Scenarios Teach*, *Results, Briefly*).
   - **Openings:** state the claims or the ground covered. Avoid stage directions about the chapter or the book (*This chapter unpacks…*, *Three movements will structure what follows…*, *Along the way, one full exemplum will show…*, *the map is drawn*, *carry the skeleton*, *the method Cicero would recognize*).
   - **Length and form:** do not talk about paragraphs, pages, or “essay scale.” Describe the substance (*each developed fully*, *six decisions from kitchen to nation*), not the medium.
   - **Craft jargon:** keep *exemplum*, partitio talk, and similar terms out of reader-facing headings and bridges; say *case*, *example*, or just name the case.
   - **Correction rule:** when cutting meta language, do not replace it with a cleverer dual-noun or metaphor that is still about framing (*Filter, Not Blueprint*, *Full Life, Not Headcount*). If the old title was already a direct description of scope (*What This Vision Is and Is Not*), keep or restore it. Prefer plain English the reader already holds (*Life, Not Nihilism*; *The Aim Is Full Life*) over opaque contrast-pairs.
   - **Allowed:** short content maps of substance (not of writing), Charge sections as imperatives, and instructional heads only where the section is a tool (*How to Use the Matrix*).
   - **Partitio must still be prose.** Do not dump bare telegrams (*First: … Second: … Third: …*) that force the reader to reverse-engineer a map. Lead with a plain sentence that orients (*The ground runs in three stretches*), then list the substance in ordinary rhythm. Scan openings for the same abrupt stack after edits that cut meta scaffolding.
   - **Parallel packets ((A,B,C)→(A′,B′,C′), not A→A′, B→B′).** When listing several items that each carry a job, stake, or relative clause, do not stack mini-pairs (*the trade that would feed the city, the skills the schools would form, the order that would hold…*). That reads as (A→A′, B→B′, C→C′) and forces a restart on every item. Prefer: name the set, then cash the stakes or jobs in a second movement (*He had to name trade, talent, order, and enemies. The port would eat or go hungry; the schools would form real skills or empty credentials…*). Same idea for *who*-stacks (*the plumber who…, the cook who…, the clerk who…* → *Plumber, cook, clerk: seal the joint; keep the kitchen; keep the records*). Keep tight two-item contrasts and deliberate *X over Y* mnemonics. Inventory of current hits and suggested rewrites: `PARALLEL_PACKET_SCAN.md`.
3. **Do not mention the harness, agents, or attack IDs** in reader-facing text.
4. **Positive first.** Fences are short; ledgers hold the rest.
5. **Premises cut labels.** When discussing Christians, atheists, Muslims, or progressives, sort by answers to the five questions / five principles, not by jersey color alone.
6. **Persons ≠ packages.** Institutional and doctrinal packages may be incompatible with the frame; persons are not automatically the package.
7. Keep part architecture compatible with book1-4 unless outline explicitly changes:
   - Part 1: The Five Principles  
   - Part 2: Foundations  
   - Part 3: Reality, Self, and Others  
   - Part 4: Ramifications, Contrasts, and Context  
   - Part 5: Application and Practice  

## Punctuation

Authorial prose in `chapters/` must **not** use the em dash (—). Prefer ordinary punctuation so logical relations stay explicit.

### Allowed (primary tools)

| Mark | Use for |
|------|---------|
| `.` | Full stop; preferred for punch and energy |
| `,` | Light join, lists, short non-restrictive asides |
| `:` | Definition, unpacking, what follows specifies what precedes |
| `;` | Two balanced independent clauses |
| `()` | True aside (lower emphasis than a dash) |
| `/` | Sparse alternatives only; prefer “or” in body prose |

### Em dash (—)

- **Forbidden** in authorial prose.
- Do not use it for asides, apposition, breaks, or summaries.
- Inside a **quotation from a primary source**, keep the source’s punctuation if the source uses a dash.

### En dash (–) and hyphen (-)

- Prefer **hyphen** for compounds and year spans in body prose: `1965-67`, `is-ought`, `self-interested`.
- En dash is optional for typographic year ranges in citations; if used, do not confuse it with the em dash. When in doubt, use hyphen.

### Replacement guide (when editing)

| Temptation | Prefer |
|------------|--------|
| Em dash before a definition or unpacking | Colon |
| Em dash linking two strong clauses | Semicolon or period |
| Em dash for a true aside | Parentheses, or a new sentence |
| Em dash for punch | Period + short sentence |
| Em dash in labels (`Habits—sleep`) | Colon (`Habits: sleep`) |

### Slash

Use sparingly. Prefer “or” / “and” in continuous prose. `/` is acceptable in tight labels only when clarity requires it.

## Anchors

When a ruling points at a locus, prefer an explicit HTML anchor or a stable `##` / `###` heading:

```markdown
## Compatible Premises {#compatible-premises}
```

## Variables

If using pandoc substitutions later, reserve `{{.title}}` and `{{.bundle}}` as in book1-4. Do not invent reader-facing placeholders without documenting them in `outline/`.
