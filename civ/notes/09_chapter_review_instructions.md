# Chapter review instructions

Point a new session at this file when reviewing `civ/chapters/`. The method is the same one used on `ccp_argon.md`, then on `ccp_eng_argon.md` and `ccp_derive_argon.md`.

The author will not accept unsolicited AI prose as the book. A suggested replacement is a proposal. It enters a chapter only after the author accepts it, or when it restates a decision the author has already accepted.

## What to do

Read one chapter at a time, in filename order, from `00_title.md` through `14_ch14.md`.

For each questionable sentence or phrase:

1. If it conflicts with a change of form already recorded below, or in the Author's decisions of `notes/07_ccp_argon_calls.md` or `notes/08_eng_derive_calls.md`, fix that passage in the chapter so the form matches the decision. Keep the surrounding exposition. Do not restyle the chapter, and do not shorten it to the short sheet.
2. If the issue is new, or if it is unclear whether an old decision already covers it, write it to `notes/10_chapter_calls.md` and leave the passage unchanged.

A doubtful match is a new call. Write the call. Leave the sentence.

After each chapter, report two lists: passages changed to an existing decision, and calls written and left unapplied. Stop there if the author wants to decide the new calls before the next chapter. If the author asks for several chapters in one pass, still leave every new call unapplied.

## What to read first

Read these before editing a chapter:

- `notes/06_editing_guide.md` for the register and the locked lines.
- `notes/00_civ_authoritative_bones.md` for the same locks in short form.
- The Author's decisions at the top of `notes/07_ccp_argon_calls.md` and `notes/08_eng_derive_calls.md`. The sections under those decisions are the original proposals. They are history. Where a decision rejected a proposal, the decision stands.
- `ccp_argon.md`, the short sheet. It is the authority for the locked lines and for the changes of form. It is not a length limit on the chapters.

`notes/07_ccp_argon_calls.md` still says, near the top, that `ccp_book_argon.md` and the chapters say "Generational obligation." That sentence is stale. The chapters now say "Form and listen," and the built book was assembled from those chapters. Leave that header unless the author asks to update the note. `notes/07` also says the book still has "Passing life." That part remains true until chapter 1 is corrected and the book is rebuilt.

`notes/05_rights_intlaw_managerialism_consent.md` still names "Generational obligation." It is a note, not a chapter. Leave it unless the author includes the notes in the review.

## Scope

Edit `civ/chapters/*.md` only.

`civ/build.fish` assembles `ccp_book_argon.md` by concatenating `civ/chapters/*.md` in filename order, then builds the pdf, html, and epub, and may run a release script. Edit the chapters. Leave `ccp_book_argon.md` and `civ/build/` alone. Run `build.fish` only when the author asks.

Leave `ccp_argon.md`, `ccp_derive_argon.md`, and `ccp_eng_argon.md` alone. Those reviews are finished. If a chapter seems to expose an error in the short sheet, write a call. Do not change the sheet in the chapter review.

Leave the historical proposal sections of notes 07 and 08 alone. When the author decides a new call, record that decision in the author's terms at the top of `notes/10_chapter_calls.md`, the way notes 07 and 08 record theirs, and apply only what the author has accepted.

## How to write a call

Create `notes/10_chapter_calls.md` at the first new issue. One section per call, in this shape:

- **Where:** file and section.
- **Wording:** the sentence as it stands.
- **Problem:** what is questionable, in a few sentences.
- **What to do:** a suggested replacement, or the decision the author needs to make.
- **Status:** not applied.

Write proposals in the register of `notes/06_editing_guide.md`: declarative sentences, no em-dash, no exclamation mark, words in their literal dictionary meanings. Do not invent a rewrite of a locked principle sentence or a locked virtue name.

## Locked lines

These stay verbatim. `notes/06_editing_guide.md` and `notes/00_civ_authoritative_bones.md` lock them.

1. Reality is real, knowable, and consequential. Truth is what corresponds to reality.
2. Humans are self-interested, limited, and particular.
3. Goodness must include human life continuing, and that life kept and built.
4. What you keep and build is self-anchored, tiered, and outward. Enemies are real.
5. Justice is scoped procedural symmetry.

Virtues:

1. Align with truth
2. Take ownership, which is responsibility
3. Seek excellence
4. Admire beauty
5. Be humble
6. Have children
7. Form and listen

Virtue 7 was "Generational obligation." The author authorized the shorter name that states both acts, and was unsure of a better title. Do not coin a new virtue title. Ordinary uses of "generational" are not the virtue name. Leave "generational theft," "The Generational Criterion," and "The Generational Horizon" unless a separate decision covers that sentence.

The opening civilization sentence on the short sheet is: "Continuing life from generation to generation, and the ability to build and keep, is civilization."

## The book has room the shorter papers did not

Some answers on the short sheet, the derivation, and the engineering draft were shaped by how much those papers could expound. Where the recorded reason was that the shorter form was sufficient, or that a short paper could not carry the argument, the book may keep the fuller text. Do not delete a chapter exposition to match the shorter file.

A change of form still applies inside that fuller text. A change of form is a correction the author accepted to a name, a phrase, or the relation the sentence states. Apply the correction in the chapter. Leave the exposition longer than the sheet when the book is making the same point at chapter length.

Where a short-file decision both shortened a passage and changed its form, keep the book's exposition and apply the change of form. Where keeping the text would mean a different answer, and the difference is the room to expound, write a call and leave the passage.

Two cuts were explicitly for the shorter form. They are not inline deletions in the chapters:

- Note 07, call 10. The national example was shortened to the Young Turks because that form was sufficient. Chapter 8 may keep the later Turkish measures, including the 1942 wealth tax.
- Note 07, call 20. The neglect argument was dropped because a paper that short could not make it correctly. A chapter may keep that argument and expound it. If the chapter states it without the exposition it needs, write a call. Leave the passage until the author decides.

## Decisions to apply inline

These are changes of form and claims the author has already accepted. Where a chapter still contradicts one, change that passage so the form matches. A longer exposition that argues the same decision may stay. Do not replace a chapter section with the short sheet's paragraph. The full wording is on the short sheet and in the Author's decisions of notes 07 and 08.

**Civilization and technology.** Continuing life from generation to generation, and the ability to build and keep, is civilization. Civilization is will plus method. It is not a collective. A technology is a repeatable method, together with the structure that method keeps in being.

**Knowability.** Sensory contact with reality and reason are linked. Revelation may add a claim. It may not cancel a measurement. The obligation to test it remains. Intuition may suggest a hypothesis and does not confirm a physical claim. Where "fixed" overclaimed, reality is objective. The beam example, and airplanes and telephones as the demonstration, stay. The neglect argument is the call 20 exception above: the book may expound it.

**Particular people, children, and care.** Humanity is the floor of goodness. Humans are self-interested and particular. Responsibility starts with oneself and one's own. There is no generic responsibility to Humanity. Human life is valued both broadly and particularly. Having children continues "us" and the particular parental lineage, which is real. Care order: one's own children first; lacking that, kin broadly defined; lacking that, people who share one's general heredity and are likely to share one's dispositions and tendencies. For those able, having their own children is the normal course. Drop "capable, virtuous," "replaced or conquered," and any barren-spouse nuance that displaces the lineage. The childless path stays open through adoption, teaching, apprenticeship, and durable work.

**Ownership, honesty, tiers.** Ownership is responsibility. Honesty binds inside every tier. It is not a surplus outer tier. Educate the children, not the wife. Lawful taxes for the country's continuance are the fourth tier. A tax levied to confiscate the household, or to destroy a people, is enmity. Cultures may argue the intermediate order. The shape stands: the self and the household come before the stranger.

**Enemies and violence.** An enemy is any person, or any body of persons, who acts to rob, subjugate, or destroy you, your household, or your people. The definition covers a lone burglar and a state program. Refusing to use violence to protect oneself and one's own stands against these principles. A moral code fails when its own commands produce death or ruin. Death at an enemy's hands is Point 4. It does not refute the code.

**Justice.** Justice is scoped procedural symmetry. People differ: men and women, young and old, student and teacher. The justice question is which scope differences are real and relevant. If a Muslim wants a country that differentiates Muslim and non-Muslim as written in the Quran, a non-Muslim will not naturally want to live there, and a person who truly holds that distinction to be morally right is not compatible with a country that refuses it. Those two answers do not make one polity. Do not restore the neutral "both systems illustrate scope" illustration, the sentence that Islam operates as a polity that incorporates a religion, or "all citizens in the same legal position." A secular state does not use religion as a legal position. Age, competence, office, and citizen against non-citizen can remain relevant. Symmetry runs between persons in the same position.

**Property, beauty, commons, self-anchored illness.** Separating "human life" from "mere property" is an illusory distinction. Beauty is vitality, order, and truth showing together. Vitality is living form. Depicting a wound, a ruin, or a deformed body truthfully is a different act from celebrating it. "Deformity" is not part of the cult-of-ugliness fence. A shared work lasts when a named person or a named office is charged with it. The self-anchored illness is being sick from a neglect of the care that was in his power.

**Form and listen.** This rename is already in the chapter lists and in the virtue's reasoning, including chapters 1, 4, 6, 10, 12, and 14. Do not do that pass again. Where a later sentence still treats the virtue as one-directional, bring it to the short sheet. The sheet's definition is the standard:

Elders form those younger than themselves. The duty is broad. It includes kin. It also includes politicians and other leaders, who train the people younger than themselves. Those who are younger seek the experience of the people around them. They listen. They test what they are told. They do not have to agree with all of it. Each person receives an inheritance he did not earn and hands it on in working order. The younger may refuse what fails the test. Listening remains. The fence is elders who will not form, the young who will not listen, and generational theft.

The Burke partnership quotation, the seed granary, and the Washington example stay. They were kept on purpose.

**Lysenko.** "Millions starved" stands. The author read the Grokipedia section on Trofim Lysenko and judged that clause accurate at this level of nuance. If a chapter states it, leave it. Do not add it where the chapter does not already make the claim.

**Sex.** Men and women are not interchangeable. The form to follow is distinct reproductive function, and average strength where the task is carrying a load. A chapter may expound that difference at book length. An unsourced fracture multiple, pack weight, or fertility mechanism is a call. Leave it until the author decides whether the book will argue it.

**Population.** The point is replacing a population. A large cohort taken from one society and culture tends to reproduce that society and culture in a different land. If the people are able, they adopt the culture of the place they enter. If they are unable, they recreate the failures of the culture they carried. The book may expound that claim. The rejected form is immediacy: "the next morning," "instantly," and "any population."

**Household and malleability.** The household feeds and forms its own children at personal cost. The rejected form is "the only institution in history." A design that requires humans to have no sex, no particular attachments, and no private motives fails when actual humans operate it. Holding that design in place has required coercion. The rejected form is "every political regime" and "terror" as the necessary means. The book may expound the corrected claims, including particular cases. It does not restore those rejected phrases.

## Known passages that still contradict a decision

Apply these when the review reaches them. They are not new calls.

**Chapter 1, opening.** `chapters/01_ch01.md` still says "Passing life and the ability to build and keep from generation to generation is civilization." Replace that clause with the short-sheet sentence: "Continuing life from generation to generation, and the ability to build and keep, is civilization." The same sentence is in `ccp_book_argon.md` because the book is assembled from the chapters. Fix the chapter. The book file updates on the next build the author asks for.

**Chapter 8, the national enemy example.** The longer Turkish exposition may stay, including the measures after the Young Turks. Call 10 shortened that example for the short sheet only. Still apply the changes of form in the same chapter. The enemy definition includes one man forcing his way into a house, not only a program against a group. Check the tier list, honesty, the children rather than the wife, and the tax-as-enmity line against the decisions above. Do not add a new history.

## Write a call instead of editing

These are not settled as chapter sentences. Describe the problem and propose wording in `notes/10_chapter_calls.md`. Leave the passage.

**Chapter 10's demographic illustration.** The 0.8 children arithmetic, the 84 percent decline, and the neighboring people who raised three or four children and cross the border are a set piece. Particularity and population replacement are decided in general. The author has not accepted chapter wording for this illustration. Do not rewrite it in silence.

**New universals, new counts, and new mechanisms.** A historical "every" or "only," a statistic with no source, or a causal step the accepted claims do not state is a new call, even when the direction agrees with a principle. Leave it. The exceptions that are already changes of form, and that are inline fixes, are the rejected phrases named under population, household, and malleability. A fuller physical account of the sexes is a call, as that section says, because the book has room to expound it.

**Register.** `notes/06_editing_guide.md` is the prose standard. When a sentence is already being changed for a decided conflict, write the replacement inside that standard. A breach that does not change the claim, found while reading, may be fixed in that sentence: an em-dash, an exclamation mark, or a banned figurative use. Do not open a separate style pass, and do not run the n-gram script in that guide unless the author asks. A literal use stays. "Load-bearing stone wall" in chapter 7 is the allowed engineering sense.

**Taste.** A sentence the reviewer would have written differently, which does not conflict with a decision and does not break the register, is not an edit. If it is genuinely ambiguous, write a call.
