# Writer instructions (book1-6)

**Scope:** Manuscript prose (chapters/). Plan files, dictionaries, and notes are writer-facing and exempt.
**Input quarantine:** Writers read only this file, `vocab_whitelist.txt`, `vocab_blacklist.txt`, `plan/`, `sources/`, `HOUSE_FORMULAS.md` when present (author-locked claim formulas), and (for style only) `style/campbell_agent_rules.md`. Earlier books in this repository (`book/`, `book1-1/`, `book1-4/`, `book1-5/`, `outline/`, `release/`) are not phrasing sources. Do not read them for voice, structure, or pictures. Content comes from the plan, the sources, and house formulas; every sentence is new. Style comes from W12 and `style/campbell_agent_rules.md`.
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

**After the picture is set, finish in direct speech.** One picture per claim-chain is enough. Once the picture has done its causal work, state the remaining steps in words that mean those steps. Do not:

- **Restack figures** for the same relation (a second metaphor for what the first already showed).
- **Inventory exemplars** where a single class-term carries the logic (a list of instances that adds no new step).
- **Pile near-synonyms** for one social or logical role (three names for one pressure, three labels for one test).

Abstract form: *establish one picture → name the next logical step directly → use one term per role → end with the discrimination the paragraph exists to make.*  
If a draft line only multiplies images or examples without advancing the chain, cut it to the class or the claim (editing may treat this under E2 and E3 together).

See Arc 4 below for a worked before/after.

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

### W11. Line revision markers (edit / rewrite)
Column-0 only; one paragraph per line. Stripped or selected by `concat.go` (see Comment convention).

| Prefix | Meaning | Default (`go run concat.go` → `book.md`) | `-old` (`book-old.md`) |
|--------|---------|------------------------------------------|-------------------------|
| `%%` | freeform edit note | drop | drop |
| `%+` | new paragraph body | keep body | drop |
| `%-` | old paragraph body | drop | keep body |

Example replace pair:

```markdown
%% soften voice; keep the claim
%- Old paragraph that was too stiff.
%+ New paragraph that sounds like a human.
```

Use HTML comments for stable arc/point/verify structure. Use `%%` for ephemeral edit chatter. Use `%+`/`%-` for contested rewrites you may want to compare. Agents: see `.grok/skills/book-markup/SKILL.md`.

### W12. Writing Style (Campbell)

**Required before drafting or style-editing manuscript prose:** read `style/campbell_agent_rules.md` in full (compact rules and both cleaned examples). Do not rely on this W12 summary alone.

Write in the formal English prose style of George Campbell, *The Philosophy of Rhetoric* (1776), especially Book I (*The Nature and Foundations of Eloquence*). Use balanced periodic sentences, moderately elevated but not archaic diction, and the measured, slightly oratorical cadence of serious didactic non-fiction of the period. Prefer abstract nouns and careful qualification; when a picture is used (W2), make it do causal work in the argument, as Campbell’s arch and tower do for two kinds of evidence. Avoid contractions, modern idioms, and short choppy sentences. One principal end per stretch of discourse; secondary effects only as means; perspicuity first.

**Author priority (see also `HOUSE_FORMULAS.md` A8):** The chief end of this style is not to *sound like* Campbell as costume. It is that **every word carry meaning**, and that every concept be dictated clearly. Sentence structure serves that end. A metaphor that can be deleted without loss is a failure (W2, W4). Careful Campbell habits (define, divide, discriminate; few words with purity and perspicuity) are required because they **increase clarity**. Content-repair passes may write plain formal prose first; full cadence passes come after substance is stable (`FIX_PLAN` Phase D).

**Canonical style file (must read):** `style/campbell_agent_rules.md`

That file is the agent-facing digest of Campbell: ends of eloquence, persuasion, evidence, hearer/speaker, use and purity, perspicuity and vivacity, sentence habits, checklists, and cleaned specimens of Campbell’s own prose.

**Emulate Campbell’s cleaned samples in that file, not other rhetoricians:**

- **Example A** — four ends of eloquence (Book I, Ch. I): definition, partition, progression of faculties, discrimination of terms.
- **Example B** — arch and tower (Book I, Ch. V): scientific vs moral evidence; the arch (one chain) vs the tower (stacked independent proofs).

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

Tested 2026-07-17 (evidence in `build/comment_test.*`): pandoc passes HTML comments through into EPUB and HTML output (invisible when displayed, present in the shipped source). The PDF path (xelatex) drops them. Conclusion: comments are not safe at render time, so `concat.go` strips markup before template execution and rendering.

**Order in `concat.go`:** (1) line markers at column 0 (`%%`, `%+`, `%-`); (2) HTML comments `<!-- ... -->`.

**Line markers:** only at the start of a line. Optional single space after `%+` / `%-` is stripped with the marker. Mid-line `%` (e.g. `50%`) is ordinary text. Default mode is **new** (publish): keep `%+` bodies, drop `%-` and `%%`. Flag `-old` writes `book-old.md` with the reverse for `%+`/`%-`. `./publish.fish` always uses new mode.

**HTML comments:** do not place a literal `<!-- -->` inside fenced code blocks; the stripper does not parse code fences.

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

### Arc 4: picture set, then direct finish (no restack, no inventory)

**Abstract fault:** A sound picture opens the relation, then the prose piles more metaphors, instance-lists, and synonym stacks that do not add a new step in the chain. Vivacity falls; the spine of the claim is harder to hear (Campbell: fewer words with purity and perspicuity; one principal end).

**Before (second half only; first half already good):**

> The world does not take its seat in the room where speeches are scored. It answers when a claim is made to stand for what will happen to bodies, harvests, children, and the works of men's hands. There the prior, the fashion, and the permission of the company may still win the hour; they do not rewrite the consequence. Truth is not the applause of the room, but the name of a model of reality rather than of falsehood or of hope.

**After:**

> The world does not take its seat in the room where speeches are scored. When a claim is offered as knowledge of what is, or of what will follow from what is, the world answers by what in fact comes to pass. What the company accepts may still prevail among speakers for a time; it does not alter what comes to pass. Truth is the name of a model of reality, rather than of falsehood or of hope.

**Notes:**

| Fault in before | Fix in after |
|-----------------|--------------|
| Instance inventory (bodies, harvests, children, hands) | One class: *what is* / *what will follow from what is* |
| Synonym pile (prior, fashion, permission) | One role: *what the company accepts* |
| Restacked figures (win the hour, rewrite, applause) after the room picture | Direct verbs: *prevail*, *alter*, *comes to pass* |
| Discrimination (truth vs acceptance) | Kept, in plain terms, as the end of the arc |

The room picture stays once and still does causal work. Everything after it advances the chain without a second painted scene.

---

## Files

- `vocab_whitelist.txt` — approved terms with fixed definitions.
- `vocab_blacklist.txt` — banned terms with reason and replacement.
- `HOUSE_FORMULAS.md` — author-locked claim formulas (Phase A); when present, agents must follow them.
- `style/campbell_agent_rules.md` — Campbell style rules and cleaned examples (required by W12).
- `concat.go` — strips writer comments at build.
- `build/comment_test.*` — renderer evidence for the comment convention.
