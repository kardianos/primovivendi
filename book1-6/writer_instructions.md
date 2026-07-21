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

### W12. Writing Style

Write in the formal English prose style of the 1820s following the rhetorical habits taught in Hugh Blair’s Lectures on Rhetoric. Use balanced periodic sentences, moderately elevated but not archaic diction, and the measured, slightly oratorical cadence found in serious non-fiction of the period. Prefer abstract nouns and careful qualification over vivid metaphor. Avoid contractions, modern idioms, and short choppy sentences. Model the tone on the philosophical and critical prose of the era rather than novels.

For example:

> MY design in the four preceding lectures, was not merely to appreciate the merit of Mr. Addison's style, by pointing out the faults and the beanties that are mingled in the writings of that great author. They were not composed with any view to gain the reputation of a critic: but intended for the assistance of such as are desirous of studying the most proper and elegant construction of sentences in the English language. To such, it is hoped, that they may be of advantage; as the proper application of rules respecting style, will always be best learned by the means of the illustration which examples afford. I conceived that examples, taken from the writings of an author so justly esteemed, would on that account, not only be more attended to, but would also produce this good effect, of familiarising those who study composition with the style of a writer, from whom they may, upon the whole, derive great benefit. With the same view, I shall, in this lecture, give one critical exercise more of the same kind, upon the style of an author, of a different character, Dean Swift; repeating the intimation I gave formerly, that such as stand in need of no assistance of this kind, and who, therefore, will naturally consider such minute discussions concerning the propriety of words, and structure of sentences, as beneath their attention, had best pass over what will seem to them a tedious part of the work.
> I formerly gave the general character of Dean Swift's style. He is esteemed one of our most correct writers. His style is of the plain and simple kind; free from all affectation, and all superfluity; perspicuous, manly, and pure. These are its advantages. But we are not to look for much ornament and grace in it.* On the contrary, Dean Swift seems to have slighted and despised the ornaments of language, rather than to have studied them. His arrangement is often loose and negligent. In elegant, musical, and figurative language, he is much inferior to Mr. Addison. His manner of writing carries in it the character of one who rests altogether upon his sense, and aims at no more than giving his meaning in a clear and concise manner.


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

---

## Files

- `vocab_whitelist.txt` — approved terms with fixed definitions.
- `vocab_blacklist.txt` — banned terms with reason and replacement.
- `concat.go` — strips writer comments at build.
- `build/comment_test.*` — renderer evidence for the comment convention.
