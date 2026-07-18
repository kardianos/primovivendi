# Writer instructions (book1-6)

**Scope:** Manuscript prose (chapters/). Plan files, dictionaries, and notes are writer-facing and exempt.
**Input quarantine:** Writers read only this file, `vocab_whitelist.txt`, `vocab_blacklist.txt`, `plan/`, and `sources/`. Earlier books in this repository (`book/`, `book1-1/`, `book1-4/`, `book1-5/`, `outline/`, `release/`) are not phrasing sources. Do not read them for voice, structure, or pictures. Content comes from the plan and the sources; every sentence is new.
**Phases:** Writing and editing are separate. Rules marked "editing" run as passes after drafting, one rule per pass. Do not interrupt drafting flow to satisfy them.
**Applies to:** Human writers and LLM writers alike. LLM writers: these rules are restated in the prompt for every session; compliance is checked by the editing passes, not assumed.

---

## Writing phase

### W1. No em dashes
Use parentheses for grouping. Restructure the sentence if parentheses make it heavy. Enforcement is a character sweep in editing (E1), so do not slow down drafting over it.

### W2. Direct words; no metaphoric substitution
Use words that directly mean the thing.

- **Substitution (banned):** the named thing stands in for a meaning never stated ("candyland" for goods-without-burdens, "tapestry" for a set of norms, "joints" for connections).
- **Picture (allowed and required by method):** the named thing does causal work in the argument (bridge, harvest, knife, wall). One claim, one chain, one picture.

A picture is not decoration. If you use one, stop and paint it fully:

- **Order:** the picture has structure the reader can follow (parts, sequence, direction).
- **Logic:** each step of the picture maps to a step of the claim. No orphan details.
- **Vitality:** concrete life (frost, weight, the mason's hands), not a dead label.

Every picture does all three. If you will not paint it, state the claim directly instead.

### W3. Proper sentences; one arc per paragraph
Each paragraph carries one arc of thought. If a second arc appears, split the paragraph.

### W4. Every sentence is load-bearing (editing E3)
Each sentence, if removed, would change or remove significant meaning in its paragraph. Each paragraph, if removed, would change or remove significant meaning in its chapter. Run this as an editing pass, not while drafting. Run other rules as editing passes whenever flow matters more than first-pass compliance.

### W5. Specific actors
Name who does what. Avoid "it" and "them" unless the referent is extremely clear. Passive constructions that hide the actor are banned where an actor exists ("standards decline" becomes "grant bodies stop rewarding finish").

### W6. First sentence carries the claim (default, not absolute)
Reasonable starting point for most paragraphs. Where it does not fit, a point comment (W10) is mandatory so the writer and the LLM know the paragraph's job.

### W7. Closed vocabulary
Book terms live in `vocab_whitelist.txt` (approved term: fixed definition). Banned terms live in `vocab_blacklist.txt` (term: reason -> replacement). Format is one `term: entry` per line for grep (`rg '^term:'`).

- New term: add it to the whitelist first, or do not use it.
- Blacklisted terms never appear in manuscript prose.
- One word per concept. No coinages where a defined phrase exists.

### W8. Steelman before conflict (ideal, scope permitting)
A package criticized gets at least one sentence its holders would recognize as their own view, before the conflict is marked. Apply where it makes sense in the scope of the book; not every passing mention needs one.

### W9. Empirical provenance (ideal, not absolute)
Strong empirical claims: cite a source in `sources/`, soften to what is commonly observable, or tag `<!-- verify: what needs checking -->` for the editing phase. Do not interrupt drafting to look things up; tag and move on.

### W10. Writer comments (arc, point, verify)
Use HTML comments for notes to writers and LLMs. They are stripped at concat (see "Comment convention" below) and never reach book.md or any renderer.

- Chapter level: `<!-- chapter: point of this chapter -->` at the top.
- Paragraph level: `<!-- arc: D7 | point: gratitude is not lottery guilt -->` by a paragraph whose point is not obvious from its first sentence.
- Verification: `<!-- verify: source needed for the Assur figure -->`.

---

## Editing phase (after drafting; one rule per pass)

- **E1. Em-dash sweep.** Grep the character. Restructure each hit (parentheses or two sentences).
- **E2. Metaphor sweep.** Find substitution metaphors. Convert each to a direct statement or a fully painted picture (W2).
- **E3. Deletion test.** For each sentence: remove it mentally; if nothing significant is lost, delete it. Then the same test for each paragraph.
- **E4. Quote provenance (absolute).** Every quotation has a source in `sources/`, exact wording. Otherwise it becomes paraphrase with no quote marks. No "close enough." LLMs fabricate plausible quotes; this pass is mandatory.
- **E5. Vocabulary check.** Grep the blacklist. Grep suspect new terms against the whitelist.
- **E6. Verify tags.** Every `<!-- verify: ... -->` is resolved: cited, softened, or cut.
- **E7. Actor pass.** Every unclear "it" or "them" gets a named referent. Every actorless mechanism gets its actor.

---

## Comment convention

Tested 2026-07-17 (evidence in `build/comment_test.*`): pandoc passes HTML comments through into EPUB and HTML output (invisible when displayed, present in the shipped source). The PDF path (xelatex) drops them. Conclusion: comments are not safe at render time, so `concat.go` (book1-6 copy) strips all `<!-- ... -->` before template execution and rendering. Do not place a literal `<!-- -->` inside fenced code blocks; the stripper does not parse code fences.

---

## Paired arcs (negative -> positive)

Format: a composite before-passage showing several violations at once, then the after-passage with all fixed, then notes. `(A, B, C) -> (A', B', C')`.

### Arc 1: substitution metaphor, em dash, stuffed sentence

**Before:**
> Justice is redescribed as what that generic person would pick without knowing their place—and often, in practice, as if many would simply pick candyland: goods without burden, claims without the household and line that make goods possible.

**After:**
> Justice is then redefined as whatever that generic someone would choose while knowing nothing of their place. In practice the choosing tends to come out the same way: goods without burdens, claims on what is produced, with no duty to the households and lines that produce it.

**Notes:** "candyland" substitutes for a meaning that was never stated; the after states it (goods without burdens; the specific get, the missing responsibility). The em dash becomes a period. One arc per sentence.

### Arc 2: decoration metaphor vs painted picture; throat-clearing; actorless prose

**Before:**
> Society is a tapestry of norms. When one thread frays, the whole thing unravels. It is important to note that standards must be maintained.

**After:**
> A dry stone wall stands because each stone carries the weight above it and sheds water to the course below. Pull three stones from the bottom course. Nothing falls that week. Frost comes, water sits where it should not, and in a year the courses above are on the ground. Standards in a craft work the same way. The stones are habits: measure twice, finish the surface, charge the price you named. The mason who skips one does not see the cost that week either. The guild sees it in the year.

**Notes:** "tapestry" and "unravels" do no causal work; they substitute for a mechanism. The wall picture does causal work and is painted: order (courses, water path), logic (each step maps to the claim), vitality (frost, weight, the year). "It is important to note" is throat-clearing; cut. Actors named: the mason, the guild.

### Arc 3: coinage vs defined phrase

**Before:**
> Thrival requires discipline.

**After:**
> A life that continues well requires discipline.

**Notes:** "thrival" is a coinage with no definition. Use the defined phrase from the vocabulary file.

---

## Files

- `vocab_whitelist.txt` — approved terms with fixed definitions.
- `vocab_blacklist.txt` — banned terms with reason and replacement.
- `concat.go` — strips writer comments at build.
- `build/comment_test.*` — renderer evidence for the comment convention.
