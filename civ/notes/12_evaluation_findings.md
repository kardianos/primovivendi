# Evaluation of the Latest Edits and Re-Evaluation of the Essays and Book

**Date:** 2026-10-27
**Scope:** `civ/ccp_argon.md`, `civ/ccp_derive_argon.md`, `civ/ccp_eng_argon.md`, `civ/chapters/01_ch01.md` to `14_ch14.md`. `civ/ccp_book_argon.md` matches the concatenated chapters (only blank separator lines differ), so the chapter line numbers below also locate the book text.
**Edits evaluated:** mainly `cb0a84c` ("additional editing"), which applied most of `notes/11_evaluation_findings.md`. Where an older commit in the 26–27 October series (`cd8cb49` to `a410a0a`) caused a problem that is still present, the commit is named.
**Standard:** `notes/06_editing_guide.md`, read together with the author's recorded decisions in `notes/07_ccp_argon_calls.md`, `notes/08_eng_derive_calls.md`, and `notes/10_chapter_calls.md`.
**Method:** I read all three essays and all fourteen chapters in full and reviewed the word-level diff of `cb0a84c`. I ran Parts A and B of the guide's audit and checked the built HTML in `build/` for rendering. I checked four factual claims by web search; sources are cited where they are used.

No source files were edited when this file was written. §2.1 (formatting) has since been fixed; see the status note there.

**Sources policy.** Factual checks cite primary sources first: statutes, the article itself by DOI, official records, and first-hand accounts. Secondary scholarly summaries come next. Wikipedia is not used.

---

## Summary

- **Audit status.** Part A of the guide finds no real violations. The remaining hits are literal uses ("rot" of timber, "honest" for a person, a "load-bearing" stone wall, "below zero" as a temperature) and one regex false positive at `12_ch12.md:178`. The Part B n-gram list is dominated by the essays' structural labels ("What it cannot accept", "Range of views"). That repetition is part of the format and is not a prose tic.
- **Most of `cb0a84c` is correct.** The leaked reviewer remarks are gone (ch11, ch13). The `$` and `%` symbols are now words. "an negligent" is corrected. Eng's definitions of technology and civilization now match. Chapter 8's tiers carry the tax and honesty rules. The duplicate sentences in eng are gone. The Eastern Air Lines date is right (July 1990). The cross-references in ch14 now match their targets. "Form and listen" no longer sits in a noun slot in ch04, ch06, or ch10:101.
- **Two edits in `cb0a84c` reverse recorded author decisions** (§1.1). They followed advice in `11_evaluation_findings.md` that conflicted with `notes/07` decisions 21 and 24, and they left the three essays inconsistent with one another.
- **New errors from the edits** (§1.2). The ch08 enemy passage now attributes massacre and deportation to Jewish subjects, and its two "truths" say the same thing. Ch07's handicapped race no longer makes physical sense. The Culver City passage in ch07 and ch09 has drifted apart. "The summary list" in derive points to a list that the essay does not contain. Several other prose problems are listed there.
- **Rendering defect across the essays** (§2.1, **now fixed**). Twenty-five lists in the three essays, and three lists inside blockquotes in the book, had no blank line before them, so pandoc printed them as run-in text inside the preceding paragraph in the PDF, HTML, and EPUB. One of these, derive item "3.", was created by the 26 October edit.
- **Factual errors in the book** (§2.2). The worst are the Singapore caning passage (caning is imposed *in addition to* prison, not instead of it), typhus versus typhoid, Hegel described as an opponent of private property, the Augustine attribution, the Budapest Memorandum described as "guaranteeing" borders, and the period and author of the 8.7 million famine citation.

---

## Part 1: Evaluation of the Recent Edits

### 1.1 Edits that reverse the author's recorded decisions (High)

1. **"Valued both broadly and in particular" was deleted from all three essays.**
   - `notes/07_ccp_argon_calls.md`, decision 24: *"Human life is valued both broadly and in particular. Applied in the self-interest range paragraph, and the floor is restated under 'Have children.'"*
   - `cb0a84c` removed "He values human life broadly and in particular" and "Human life is the floor of goodness" from `ccp_argon.md:97` (Principle 2, range of views). It also removed "he values human life both broadly and in particular" from `ccp_derive_argon.md:72` and `ccp_eng_argon.md:170`. Advice 1.4.1–1.4.2 in `11_evaluation_findings.md` was to *move* those sentences to Principle 3 or 4 or to Virtue 6. They were deleted instead. A search now finds "broadly and in particular" nowhere in `civ/`. The floor sentence survives only under "Have children".
   - **Result:** an idea the author asked for is gone from the corpus. The new argon range paragraph is a reasonable statement of the range, but it no longer says what decision 24 required.
   - **Decide:** either restore "Human life is valued both broadly and in particular" (in the range paragraph, as decided, or under Virtue 6) or record that decision 24 has been revised.

2. **The "particular children, in order" sequence was moved in one essay only.**
   - `notes/07`, decision 21: own children first, then kin broadly defined, then general heredity. *"Applied in 'Have children.'"*
   - `cb0a84c` moved those sentences out of Virtue 6 and into the Virtue 7 section in `ccp_derive_argon.md:280`, following advice 1.4.3 of the previous findings. `ccp_argon.md:283` and `ccp_eng_argon.md:655` still carry the sequence under Virtue 6.
   - **Result:** the three essays now disagree about where the sequence belongs. In derive, the moved sentences also read badly in their new place (see §1.2, item 8).
   - **Decide:** follow decision 21 and restore derive to Virtue 6, so all three essays match. Or, if the author now prefers Virtue 7, move it in argon and eng as well.

> **Note on `notes/11_evaluation_findings.md`.** That file contradicted decisions 17, 21, and 24 (it asked for Virtue 7's short sentences to be lengthened, and for the two items above to be moved). Several of its suggested rewrites contain em dashes, which the guide bans. It also cites guide sections that do not exist ("Section 2.1", "2.3", "3.1", "3.3"). Before acting on any of its remaining items, check them against the calls files.

### 1.2 Errors introduced or exposed by the edits

1. **The ch08 enemy conclusion is now historically wrong and repeats itself** (`08_ch08.md:197`, `:202`, `:206–207`). High.
   - Item 1 (`:206`) now says the regime used *"massacre, deportation, confiscatory taxation, and forced labor to destroy Greek, Armenian, and Jewish subjects."* The 1915–1923 massacres and deportations in the first bullet (`:199`) were against Armenians and Greeks (and Assyrians). The Jewish citizens enter the record only through the 1942 tax (`:200`). The new sentence attaches massacre and deportation to them.
   - Items 1 and 2 now state one truth twice: item 1 is "A state that strips a people inside its borders is their enemy," and item 2 is "the government under which you live becomes your enemy if it turns its coercive power toward your destruction." So "two vital truths" is now one truth.
   - The adding of Jewish citizens was applied to the conclusion (`:202`) but not to the premises. The lead-in (`:197`) still says "Greek and Armenian Christian populations," and the motive sentence (`:202`) still says "the Greek and Armenian minorities."
   - `:202` lists "conscription" among the state's means. The Twenty Classes conscription that supported this is no longer in the passage, so the word has no evidence under it.
   - `09_ch09.md:290` still says "the Greek and Armenian shopkeepers in 1915 and 1942."
   - **Suggested fix:** Make item 1 the specific lesson of the 1942 statute: a state can make itself the enemy through a tax law alone, by assessing a people beyond what it owns. Keep item 2 as the general rule. In `:197` and the first sentence of `:202`, use "non-Muslim minorities". Either restore the conscription evidence or drop the word. Align `09_ch09.md:290`.

2. **Ch07's handicapped race no longer makes sense** (`07_ch07.md:110`). Medium.
   - "The meet cancels the faster heat, **or requires the fastest runners to start well behind the line** … The fast girl spends the race at a pace that does not test her." A runner who starts well behind must run *harder*, so the second sentence fits only the cancelled-heat option. "The fast girl" is new. The vignette has "your daughter" and "the fastest girl in town".
   - **Suggested fix:** Keep one mechanism: "The meet cancels the faster heat and runs every girl in one slow heat, so the recorded results look closer together. The fastest girl in town spends the race at a pace that does not test her."

3. **The Culver City passages in ch07 and ch09 have diverged** (`07_ch07.md:110`, `09_ch09.md:231`). Medium.
   - `notes/10` call 12 put the case in both chapters deliberately. `cb0a84c` revised only ch07 ("under-worked even in the faster room"). Ch09 keeps the earlier wording: "bored even when that class exists", "A student who would do the work had been free to enter", "Remove it, and he spends the year…". A reader now meets the same case twice, with the same citation and slightly different sentences.
   - **Suggested fix:** Either make the two passages match, or let ch09 state the mechanism and refer back ("Culver City High School, Chapter 7").
   - Citation form: "*Los Angeles Times* high-school report" is vague. Per `notes/10` it is a student report in the paper's high-school section (High School Insider), 28 February 2023. Give the article title.

4. **The Virtue 7 question in ch06 now asks the impossible and repeats itself** (`06_ch06.md:83`). Medium.
   - "…bind living adults to **form** those younger than themselves, **including great-grandchildren not yet born**." No one can form a person who is not yet born. One can only hand the inheritance on to him. The bullet then restates the virtue twice ("The younger listen and test what they are told. Elders, including kin and those who lead, form the people younger than themselves."). The neighbouring bullets ask "…only if he feels like consenting?". This one dropped that frame, so the parallel is broken.
   - **Suggested fix:** "Do the younger owe the elders who formed them a hearing only if they consent, and do living adults owe the young their formation, and the unborn a working inheritance, only if they consent? None of the unborn can sign a consent contract today."

5. **Ch10 now states the three violations as facts** (`10_ch10.md:105`). Low–Medium.
   - "The grave violations of Virtue 7 are three. Elders refuse to form those younger than themselves. The young refuse to listen…" read as statements that elders *do* refuse.
   - **Suggested fix:** "The grave violations of Virtue 7 are three: elders who refuse to form those younger than themselves; the young who refuse to listen to the experience around them and test it; and **generational theft**, in which the living treat…"

6. **The ch11 appositive is ambiguous** (`11_ch11.md:144`). Low–Medium.
   - "The capture operates through institutions, **the teacher-training colleges, the school curricula, the youth media, and the universities**, so that…" reads as a list of five things, the first of which is "institutions". Curricula and media are not institutions in the sense of the sentence. The previous findings proposed an em dash here, which the guide forbids, and the comma was the substitute.
   - **Suggested fix:** "The capture runs through the teacher-training colleges, the school curricula, the youth media, and the universities, so that the builders' children are formed against the household that raised them."

7. **Ch04's "he" has no antecedent** (`04_ch04.md:131`). Low.
   - "It is **gratitude** to the ancestors who built it, and the practice of **Virtue 7 (Form and listen)**: he tests what he was told…" The subject of the sentence is "The proper moral response". The child appears two sentences earlier.
   - **Suggested fix:** "…: the heir tests what he was told, keeps what holds, and forms his own children in it."

8. **Derive's new virtue-numbering sentence points to a list the essay does not have** (`ccp_derive_argon.md:41`, `:265`, `:282`). Medium.
   - "The seven virtues are numbered in **the summary list**." The derivation has no summary list. That list is in `ccp_argon.md`. The first list of virtues in derive is the five-item list at `:249`. Line `:265` says "the sixth virtue in the summary list" and `:282` still says "our seventh and final individual virtue", so the ordinal style is now mixed.
   - **Suggested fix:** At `:41`: "Each virtue carries a fixed number, used in every companion document. The derivation reaches them in the order the constraints require, so the order of appearance is not the order of the numbers." At `:265`: "This gives us the virtue numbered six:". At `:282`: "This gives us the seventh virtue:".

9. **The sentences moved to derive's Virtue 7 do not fit their new place** (`ccp_derive_argon.md:280`). Medium. This also bears on §1.1 item 2.
   - "A man who lacks children of his own **cares for** his kin… Lacking those, he **cares for** the people who share his general heredity…" Then: "**this seventh virtue** remains completely open". The virtue is named only at `:284`, after this. And the sentences are about care (Principle 4 tiers), not about forming the young, so they do not lead into what follows.
   - If §1.1 item 2 is resolved by restoring Virtue 6, this problem goes away.

10. **Eng's Principle 2 paragraph now repeats itself, with a broken "therefore"** (`ccp_eng_argon.md:170`). Low–Medium.
    - "His care is anchored in his own life, in his spouse, and in his own children… He has no generic responsibility to Humanity. **His self-interest therefore extends to the survival and welfare of his spouse and his own children.**" The last sentence repeats the first new sentence, and "therefore" does not follow from "no generic responsibility".
    - **Suggested fix:** Delete the last sentence.

11. **Eng's enemy passage now denies the definition before it gives it** (`ccp_eng_argon.md:393`). Low.
    - "Dislike is not the definition of enmity" now comes six lines before the definition (`:399`). "Predatory actors exist, from the solitary violent criminal to **a state program** that seeks…" makes a program an actor.
    - **Suggested fix:** Put the dislike sentence after the definition at `:399`, and write "…to a state that carries out a program to rob, subjugate, or destroy a peaceful people."

12. **The 8.7 million famine citation has the wrong period and the wrong author attribution** (`05_ch05.md:82`, `13_ch13.md:111`). Medium.
    - The 2020 *Journal of Genocide Research* paper is Oleh Wolowyna, "A Demographic Framework for the **1932–1934** Famine in the Soviet Union" (single author; doi:10.1080/14623528.2020.1834741). The 8.7 million all-Union figure (Ukraine 3.92, Russia 3.26, Kazakhstan 1.26) is for **1932–1934**. Ch13 says "in 1932–1933". Ch05 attaches the figure to "the terrible winter and spring of 1932–1933". "Wolowyna and others" should be "Wolowyna" (the estimate comes from his research group, but the article has one author). Sources:
      - The article itself: [doi:10.1080/14623528.2020.1834741](https://doi.org/10.1080/14623528.2020.1834741), with a copy hosted by the [Shevchenko Scientific Society](https://shevchenko.org/research-publications/publications/a-demographic-framework-for-the-1932-1934-famine-in-the-soviet-union/).
      - The all-Union breakdown: the research team's own report through the [Canadian Institute of Ukrainian Studies](https://www.ualberta.ca/canadian-institute-of-ukrainian-studies/news-and-events/news-at-the-cius/2018/march/oleh-wolowyna-measuring-mortality-in-the-holodomor.html).
      - The team's regional loss data for 1932–1934: [Harvard Ukrainian Research Institute, MAPA](https://gis.huri.harvard.edu/population-losses).
    - Ch05 and ch13 now carry the same long sentence and citation, and in ch05 it sits inside the illustrative blockquote. Ch05 could keep "millions starved" and refer forward ("figures in Chapter 13").

13. **The Duke citation in ch04 is vague and I could not verify it** (`04_ch04.md:108`). Low–Medium.
    - "(Duke University School of Law, comparison of that season's elite women with boys under eighteen)". A search confirmed Thompson's 10.71 in 2017. It did not find the Duke comparison or the "124 boys" figure. Name the page: Doriane Lambelet Coleman, Center for Sports Law and Policy, "Comparing Athletic Performances: The Best Elite Women to Boys and Men". Also check whether 124 counts *boys* or *performances*. Smaller points: "recruits … interchangeable in a measured race" is evidenced by elite athletes, not recruits, and "They are not. Gestation and nursing are not interchangeable offices." says the same thing twice.

14. **Ch04's evidence for the transfer claim does not support it** (`04_ch04.md:109`). Low; noted, since the author already made the call.
    - `notes/10` call 4 says the Eastern Air Lines case "does not by themselves prove the transfer claim" and that the claim is the author's. As written, the only cited case is inside a high-trust trade, so a reader can take it as a counterexample. That gives the sentence "A population that never held the method fails it more readily" no evidence of its own. If the author wants it to stand as his claim, consider saying so: "This is a claim about formation, not one proved by the case above."

15. **Smaller points in the edited passages** (Low).
    - `ccp_argon.md:97`: "continue his parental lineage, which is real". The relative clause is awkward. Try "…continue him and his parental lineage. That lineage is real."
    - `02_ch02.md:25`: "For the shared physical world, access comes through the physical senses and the instruments and reasoning built upon them." The sentence has no agent. Try "we know the shared physical world through our senses and the instruments and reasoning we build on them." The forward reference ("Chapter 3 states that limit") is valid (`03_ch03.md:59–62`).
    - `04_ch04.md:66`: "approximations, not a statute" works. "(the people in daily life-support)" can read as medical. Dunbar's term is "support clique".

### 1.3 Problems from the 26 October commit (`a077b2c`) that are still present

1. **Derive's list item "3." is printed inside a paragraph** (`ccp_derive_argon.md:175–176`). High visibility. **Fixed with §2.1** as described here. The "Honesty in speech…" paragraph is not indented and has no blank line after it. In `build/ccp_derive_argon.html`, item 3 ("Why must effort move beyond the front door?") prints as "…thins as it moves outward. 3. **Why must effort…**" in the middle of the paragraph. Fix: indent the honesty paragraph three spaces so it belongs to item 2, and leave a blank line before "3.".
2. **"Conversely" no longer contrasts with anything** (`ccp_eng_argon.md:89`). It used to answer the deleted child-deprivation sentence. Now it follows "Sensory contact with reality and reason are linked." Replace it with "The daily operation of engineered systems shows…".

### 1.4 Previous findings (`11_evaluation_findings.md`) that are still open

| Item | Status |
| :--- | :--- |
| 1.2.2 choppy cadence in Virtue 7 text | Not applied. **Correctly**: `notes/07` decision 17 says the author wanted it short. Do not lengthen. |
| 1.3.4 ch10 duplicate "Elders form…" | Lines 70 and 74 are fixed. Line 70 and the answer at line 88 still repeat the same five sentences (see §2.4). |
| 1.4.1–1.4.3 | Applied in a way that conflicts with decisions 21 and 24 (§1.1). |
| 2.1.4 rhetorical question openers | Not applied: `02_ch02.md:59` ("What happened?"), `13_ch13.md:88`, `13_ch13.md:140`. |
| 2.2.1 derive ordinals | Applied, but with a reference to a list the essay does not have (§1.2 item 8). |

---

## Part 2: Re-Evaluation of the Essays and Book

### 2.1 Rendering and formatting (High: visible in every built edition)

> **Status: fixed (2026-10-27).** Only whitespace and markup changed, apart from removing the duplicated title in item 3; no wording changed. In detail:
> - A blank line was inserted before each of the 25 essay lists.
> - Derive's "Honesty in speech…" paragraph is indented under tier item 2, with a blank line before item 3, so the list stays one list numbered 1 to 3.
> - A `>` line was added between the two virtue lines in ch10.
> - In ch11, the Plank Bridge vignette is nested under list item 4 and its duplicated bold title removed.
> - A `>` line was added before the lists inside blockquotes at `05_ch05.md` (Kanesh tablets; the peasant's two choices) and `07_ch07.md` (the three institutional failures). My first scan skipped blockquotes; the HTML showed these three lists as run-in text too.
>
> `ccp_book_argon.md` was regenerated from the chapters with the same assembly code as `build.fish`. Rendering was checked with `pandoc -t html` into a scratch directory: no run-in lists remain. `build.fish` was **not** run, because it ends by calling `../push_release`. Rebuild and publish when ready.

1. **Twenty-five lists in the essays print as run-in text.** The build reads pandoc markdown, where a list that follows a paragraph with no blank line becomes part of that paragraph. Confirmed in `build/ccp_eng_argon.html` ("…analyzed by identifying: 1. The initial…") and `build/ccp_derive_argon.html` ("…attack everything we have built: 1. **Material Degradation:**…"). Affected lines, where the list begins:
   - `ccp_argon.md`: 224
   - `ccp_derive_argon.md`: 98, 108, 139, 176, 185, 242, 249, 276, 299, 305
   - `ccp_eng_argon.md`: 10, 188, 258, 347, 396, 427, 496, 506, 510, 548, 608, 616, 635, 682
   - **Fix:** insert a blank line before each list. The book chapters are not affected. Lists that follow a heading render correctly.
2. **The two virtues at ch10:23–24 print on one line.** "> **6. Have children**" and "> **7. Form and listen**" are adjacent blockquote lines, so they join into one paragraph ("6. Have children 7. Form and listen" in `build/ccp_book_argon.html:4321`). Fix: put a `>` line between them.
3. **Heading duplicated in ch11** (`11_ch11.md:34–36`). List item 4 is a bare bold title, followed by a blockquote that repeats the same bold title.

### 2.2 Factual accuracy

**Medium (a reader can check these and will find them wrong):**

1. **Singapore caning is not an alternative to prison** (`09_ch09.md:243`). The passage contrasts caning with "Western prisons that confine a young offender … for five years" and says it "returns the chastened offender to work … the following week." Under Singapore's Penal Code, robbery with hurt (s 394) carries **5 to 20 years' imprisonment and at least 12 strokes**. For most caning offences, caning is imposed together with a prison term. Primary sources: Penal Code 1871, s 394, on Singapore Statutes Online (sso.agc.gov.sg/Act/PC1871); Singapore Police Force releases that state the statutory penalty when charging under s 394 ([Feb 2024](https://www.police.gov.sg/Media-Hub/News/2024/20240226_two_men_arrested_for_robbery_with_hurt), [Apr 2025](https://www.police.gov.sg/Media-Room/News/20250427_two_men_arrested_for_robbery_with_hurt_within_four_hours)). Keep the point about symmetry (same statute for the banker's son and the dockworker) and drop the "instead of prison" contrast.
2. **Typhus versus typhoid** (`02_ch02.md:55`). Drinking stagnant marsh water causes typhoid fever (or cholera or dysentery). Typhus is carried by lice and fleas. Write "typhoid fever".
3. **Hegel did not propose abolishing private property** (`12_ch12.md:66`). The bullet grouping Hegel with Marx says both promised "that abolishing private property would produce a selfless 'New Socialist Man'" and replaced law "with class liquidation." Hegel's *Philosophy of Right* defends property. Split the bullet: Hegel for historicized truth (Principle 1), Marx for the rest.
4. **The Augustine warning is from a different book** (`03_ch03.md:15`). The "gold of the Egyptians" passage is in *On Christian Doctrine* (II.40). The warning against Christians making false claims about astronomy is in *The Literal Meaning of Genesis* (I.19.39). Name both works.
5. **The Budapest Memorandum gave "assurances," not guarantees** (`09_ch09.md:341`). "written security assurances … **guaranteeing** Ukraine's existing borders". The signatories undertook to *respect* the borders. The gap between assurances and guarantees is the point of the example, so the wording matters. Also, Ukraine's own warheads are not "Condition B" as the chapter defines it (a stronger power ready to punish the violator).
6. **Durnovo's interval and Stolypin's "twenty years"** (`13_ch13.md:87`, `:94`). The "give the state twenty years of internal and external peace" line comes from Stolypin's 1909 newspaper interview, not the 10 May 1907 Duma speech the bullet attaches it to. Durnovo's February 1914 memorandum came about four months before Sarajevo (28 June), not five.
7. **The Varlık Vergisi deportation count needs a source** (`08_ch08.md:200`). "thousands … shipped … to forced labor battalions". The best-supported count is 1,229 sent to Aşkale, with 21 deaths among those sent to work. That count is given in a written question filed in the Turkish Grand National Assembly ([TBMM document](https://cdn.tbmm.gov.tr/KKBSPublicFile/D27/Y3/T10/WebOnergeMetni/fe5c70dd-6537-4667-8415-c0fdb768c38f.pdf)). The first-hand account is by the official who ran the tax in Istanbul: Faik Ökte, *Varlık Vergisi Faciası* (1951), in English as *The Tragedy of the Turkish Capital Tax*, trans. Geoffrey Cox (Croom Helm, 1987). Check Ökte's own figures before citing. The governing statute is Law No. 4305 (*Resmî Gazete*, 12 November 1942). Write "more than twelve hundred", or cite the count.
8. **Spousal privilege is not constitutional** (`09_ch09.md:16`). The spousal testimonial privilege is common-law and statutory. Write "rules of evidence in modern law".
9. **The UN rights catalog is overstated** (`09_ch09.md:274`). The UDHR names housing, medical care, social security, higher education "on the basis of merit", and "periodic holidays with pay". It does not name a "furnished dwelling", a "guaranteed income", or "high-speed internet". The internet item is a non-binding 2016 resolution on access. In a book that asks for literal accuracy, quote the articles.
10. **Kant's murderer case** (`12_ch12.md:58`). "a categorical duty to tell the murderer the exact truth **so that he may enter the house and murder them**" gives Kant a purpose he does not have. Write "…the truth, even if the truth lets him find them."

**Low:**

- `10_ch10.md:86`: "the soldiers who died in the trenches eighty years ago". Trench warfare is 1914–1918, more than a century ago. Eighty years ago is the Second World War.
- `10_ch10.md:78`: "In 1790, when the French Jacobins abolished France's laws, churches, courts, and property rights". In 1790 the actors were the National Constituent Assembly (church lands, the Civil Constitution of the Clergy). The Jacobins governed in 1793–94.
- `07_ch07.md:79`: "hammers, tongs, and **anvils** hang … on the shop wall". An anvil weighs 100 pounds or more and sits on a stump. `:81`: straw yellow and peacock blue are temper colours, not "quench colors". Cherry red is a heat colour.
- `04_ch04.md:63` and `06_ch06.md:160`: "some twenty-five thousand adult days". 25,000 days is about 68 years, which is a whole life, not its adult part. Use "some twenty-five thousand days", or "some twenty thousand adult days".
- `01_ch01.md:42`: "the first infant born ten thousand years ago". Humans are far older than that. Write "an infant born ten thousand years ago".
- `04_ch04.md:113`: the title is *A Theory of Justice*. `04_ch04.md:23`: "noble savage" is Dryden's phrase. Rousseau never used it.
- `02_ch02.md:60` and `ccp_derive_argon.md:50`: Chernobyl is "an administrative deadline and … bureaucratic bonuses" in the book and "an administrative quota" in derive. Pick one. "Bonuses" needs a source. The RBMK has no "reactor vessel" in the usual sense, so write "reactor core".
- `07_ch07.md:88`: Rome fell to about 30,000 people over the sixth century, after the Gothic War, not "within two generations" of 476. The Forum's cattle pasture (Campo Vaccino) is medieval.
- `09_ch09.md:128`: after Leuctra, Sparta was a minor power for centuries, not "an insignificant village". "Not a single work of … architecture" is hyperbole the guide's register avoids.
- `12_ch12.md:122`: *kitman* is not a term of classical jurisprudence. It comes from modern polemical writing. Keep *taqiyya* and name the juristic sources.
- `12_ch12.md:50`: "The younger … test what they are told" is not a Confucian teaching. Filial piety stresses deference. The sentence was added when Virtue 7 was renamed.
- `12_ch12.md:164` (MacIntyre, "for losers") and `:178` (Dugin quote): both quote living people. Give a source for each.
- `03_ch03.md:36`: Ibn Rushd wrote about seventy years after al-Ghazali, so "two generations", not "a generation".
- `12_ch12.md:78`: Faraday, Maxwell, and Mendel are nineteenth-century figures and do not belong in the Scientific Revolution. `:106` quotes 1 Timothy 5:8 in a modern translation and Romans 13:4 in the KJV. Use one translation.
- `02_ch02.md:125`: 31 knots is a near gale (Beaufort 7). A gale starts at 34.
- `13_ch13.md:82`: the *mir* assembly usually decided by two-thirds, not unanimity.

### 2.3 Logic and consistency of the argument

1. **Ch02 promises four distinctions and gives three** (`02_ch02.md:82`). "a civilization must master **four** practical distinctions". Subsections 1 to 3 follow, then the chapter ends. This has been the case since the first draft (`eaf9f59`). Either supply the fourth (the obvious candidate is "Precision depends on the task", which the essays carry under correspondence) or change the count to three.
2. **Ch06 counts five failures but lists four** (`06_ch06.md:50`). "it fails on **five separate foundational points**". Item 5 is "Where Consent Actually Belongs", which places consent rather than describing a failure. Write "four failures, and one proper place".
3. **The celibacy test is applied to some religions and not to Christianity** (`12_ch12.md:134`, `11_ch11.md:140`). Buddhism, Jainism, and the Shakers fail the test "if every member attains the highest ideal… the society ends". The same test applies to Christian monastic and clerical celibacy (1 Corinthians 7), and the chapter does not apply it there. The chapter opens by promising "the **same inspection checklist** to every … religious framework" (`:9`), so a reader will notice. Either explain why it does not apply (for example, Christianity holds marriage and children as the ordinary vocation and celibacy as a vocation for some, not the summit every member should reach), or apply it.
4. **Principle 3 is used as the label for decay** (`ccp_derive_argon.md:120`, `07_ch07.md:63`, `09_ch09.md:232`). "natural decay … (Principle 3)", "Principle 3 (physical things degrade…)", "Recall Principle 3: … failure and death require no effort". Principle 3 is the standard of goodness. Decay is the condition under which "kept" is required. Write "(the decay that Principle 3's *kept* answers)", or cite Principle 1.
5. **A foreign state is put in Tier 4** (`09_ch09.md:99`). "treating Tier 4 and Tier 5 foreign state operations as if they were domestic citizens". Tier 4 is one's own country. Write "foreign state operations".
6. **The particularity sentence in derive and eng is ambiguous** (`ccp_derive_argon.md:87`, `ccp_eng_argon.md:217`). This is the author's wording (`notes/08` decision 3). "Tends to reproduce that society" is followed by "If the people are able, they adopt the culture of the place they enter." The two sentences agree only if "able" means "able to adopt". Consider "able and willing to adopt the method of the place".
7. **Ch05's diagnostic question uses the wrong word** (`05_ch05.md:24`). "Is it constructed in form, or is it constructed in substance?" The chapter's thesis is that substance is *discovered*. Write "constructed in form, or discovered in substance?".
8. **Small internal mismatches:**
   - `11_ch11.md:31`: "October-to-April" pantry, but it holds three months of food.
   - `11_ch11.md:13–17`: three "tendencies" are announced, but there are five sections.
   - `11_ch11.md:185`: the London and Singapore bullet is an unfinished conditional with no consequence.
   - `11_ch11.md:108`: bins are "wooden", then "stone".
   - `13_ch13.md:130–134`: the example is called "Two Demolition Crews…" in the text and "Two Crews…" in the blockquote, and the crews "tie their ropes to the same timber ram and pull", but a ram is swung, not pulled.
   - `14_ch14.md:59`: "The ground to the left is quicksand", but the walker stepped straight ahead.
   - `08_ch08.md:191`: "lost at the cost of his flock" is circular.

### 2.4 Prose and compliance with the guide

1. **Rule 9 (titles).** Twelve of fourteen chapter titles use the banned "Chapter N: Topic: Subtitle" form (all except ch05 and ch12). The five principle chapters also end the heading with a full stop. Section heading `07_ch07.md:162` is "Virtue 5: Be Humble: Paired With…". This was never addressed. The principle chapters could take a single descriptive title and give the principle as the first block quote, as ch10 does.
2. **Rule 5 (directing the reader) and what the regex misses.** `07_ch07.md:63` "**Look back at** Principle 1…" is a direct violation; the audit regex looks only for "Look at" and "Look closely". Close relatives of the banned openers: `01_ch01.md:27` "Step back and ask", `05_ch05.md:33` "Drive a car onto…", `:40` "Listen to a Frenchman…", `:78` "Stand in the barn stall…", `07_ch07.md:17` "Walk into…", `02_ch02.md:53` "Imagine…". Some second-person vignettes are deliberate; at least revise "Look back at" and "Step back".
3. **Rule 8 (question followed by a fragment answer).** `02_ch02.md:59` "What happened? The wheat seeds…", `13_ch13.md:88` "What occurred…? **The policy succeeded.**", `13_ch13.md:140`, `11_ch11.md:114` "…stop the rats from feeding? No.", `11_ch11.md:234–235` "**How does a civilization correct Managerialism…?** By enforcing…", `05_ch05.md:90`.
4. **Repetition:**
   - `10_ch10.md:70` and `:88` repeat the five Virtue 7 sentences ("They seek the experience…They listen. They test… They do not have to agree with all of it.") within the same section. Keep them once. Line 88 is the better home, as the answer to "What do the younger owe the elders?"
   - `ccp_derive_argon.md:72` and `:265` both say "continues his particular parental lineage. That lineage is real." and "He has no generic responsibility to Humanity." In a sequential derivation, the second could say "As Section 3 established, …".
   - `ccp_argon.md:175` and `:181` both say "You do not keep and build an enemy; you defend against him, deter him, or defeat him." `:181` also restates the two examples just given at `:177–179`.
   - `05_ch05.md:82` and `13_ch13.md:111`: the same famine sentence and citation (see §1.2 item 12).
5. **Grammar:**
   - "**While** a burned building can be rebuilt … **whereas** a murdered man cannot…, physical property is…" uses two subordinators and leaves no main clause for the first: `ccp_derive_argon.md:114`, `ccp_eng_argon.md:283`, `06_ch06.md:161`. Drop "whereas" and write "While a burned building can be rebuilt …, and a murdered man cannot be brought back, physical property is…".
   - `ccp_eng_argon.md:265`: "reduces **its each** successive generation" should be "reduces each successive generation".
   - `09_ch09.md:56`: "the jizya was collected as the tax **on remaining non-Muslim**" should be "as the tax on those who remained non-Muslim".
   - `03_ch03.md:57`: "either … , … , or …" is used with three items.
   - `11_ch11.md:140`: "children **on which** continuance depends" should be "on whom".
   - `08_ch08.md:48` and `ccp_argon.md:161`: "a stranger does not **thereby** gain the child's claim". Nothing in the sentence supplies the "thereby". Try "Owing a stranger the truth does not give him the child's claim on the household."
6. **Colloquial or figurative words:** `11_ch11.md:140` "The **clean** illustration", `12_ch12.md:134` "Jain mendicants are **the same shape**", `11_ch11.md:140` "The same **shape** appears…", `12_ch12.md:140` "the living **listening to what they were given**".
7. **Book voice slipping into editorial language:**
   - `09_ch09.md:233`: "The consequence follows from the arrangement itself, **before any census of regimes**." This reads as a reply to the review note (`notes/10` call 12: "not as a census"). Write "The consequence follows from the arrangement itself."
   - `07_ch07.md:15` "The **authoritative statement** of this virtue…" and `09_ch09.md:284` "the **authoritative specification** of Virtue 2" use the notes' vocabulary. Write "The statement of this virtue…".
   - `14_ch14.md:8`, `:69`: "(The Precepts Matrix)" is a term from book1-4 and is never defined in `civ/`. Drop it or define it.
8. **Consistency of style:** `11_ch11.md:159` "35,000 feet" against "thirty-five thousand feet" elsewhere. `13_ch13.md:84` "Strip #4". `01_ch01.md:97` "requires **Virtue**" (capitalised mid-sentence). `ccp_derive_argon.md:13` mixes "thirty-five thousand feet" with "eight hundred kilometers per hour".
9. **Virtue 7's short sentences** (`ccp_argon.md:290`, `ccp_eng_argon.md:675`, `10_ch10.md:70`) are the author's requested form (`notes/07` decision 17). No change recommended, apart from removing the repetition in item 4.

### 2.5 Suggested additions to the audit (guide §5)

- Extend the Rule 5 pattern: `Look\s+back|Step\s+back|Imagine\s|Picture\s`.
- Add a Rule 8 analog: `\?\s+(No|Yes|Never|Nothing)\.` and `What (happened|occurred)[^?]*\?`.
- Add a structural check for a list item (`^\d+\.\s` or `^[*-]\s`) directly after a non-blank, non-heading, non-list line. Apply the same check inside blockquotes, after stripping the leading `> `. This would have caught §2.1 item 1, §1.3 item 1, and the three blockquote lists in ch05 and ch07.
- After each build, grep the HTML for `: 1\. ` inside `<p>` to catch run-in lists.

---

## Items the author already decided (listed for completeness, not as defects)

- Lysenko and the Great Leap paragraph stay as written (`notes/10` call 2; `notes/08` decision 1).
- "The chain is absolute" (`04_ch04.md:41`) and the absolute family sentence in ch05 stay (calls 6 and 7).
- The choir's 0.6 children is a supposition (call 9).
- `10_ch10.md:14` "A couple with 0.8 children has 0.4 grandchildren": call 13 kept it as "loose arithmetic". For the record, 0.8 children form 0.4 couples, who have 0.32 children, so a couple has 0.32 grandchildren. The 100 → 40 → 16 line is correct. Deleting the 0.4 clause would remove the inconsistency without changing the illustration.
- "That refusal stands against these principles" after the refusal-of-violence sentence (`ccp_argon.md:185`, `ccp_eng_argon.md:409`) is the author's wording (`notes/07` decision 22).

---

## Suggested order of work

1. **§1.1:** settle decisions 21 and 24, then make all three essays agree.
2. **§2.1:** add the blank lines before lists and fix derive item 3 and the ch10 blockquote. Mechanical, and visible in every edition.
3. **§1.2 items 1, 2, 4, 8, 12 and §2.2 Medium items 1–5:** errors a reader will catch.
4. The remaining §1.2 items, then §2.3 and §2.4. Edit bottom-up per guide §6, re-run the audit, and rebuild.
