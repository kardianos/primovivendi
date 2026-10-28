# Evaluation of the Latest Edits and Re-Evaluation of the Essays and Book

**Date:** 2026-10-27
**Scope:** `civ/ccp_sheet.md`, `civ/ccp_derive.md`, `civ/ccp_eng.md`, `civ/chapters/01_ch01.md` to `14_ch14.md`. `civ/ccp_book.md` matches the concatenated chapters (only blank separator lines differ), so the chapter line numbers below also locate the book text.
**Edits evaluated:** mainly `cb0a84c` ("additional editing"), which applied most of `notes/11_evaluation_findings.md`. Where an older commit in the 26–27 October series (`cd8cb49` to `a410a0a`) caused a problem that is still present, the commit is named.
**Standard:** `notes/06_editing_guide.md`, read together with the author's recorded decisions in `notes/07_ccp_sheet_calls.md`, `notes/08_eng_derive_calls.md`, and `notes/10_chapter_calls.md`.
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
   - `notes/07_ccp_sheet_calls.md`, decision 24: *"Human life is valued both broadly and in particular. Applied in the self-interest range paragraph, and the floor is restated under 'Have children.'"*
   - `cb0a84c` removed "He values human life broadly and in particular" and "Human life is the floor of goodness" from `ccp_sheet.md:97` (Principle 2, range of views). It also removed "he values human life both broadly and in particular" from `ccp_derive.md:72` and `ccp_eng.md:170`. Advice 1.4.1–1.4.2 in `11_evaluation_findings.md` was to *move* those sentences to Principle 3 or 4 or to Virtue 6. They were deleted instead. A search now finds "broadly and in particular" nowhere in `civ/`. The floor sentence survives only under "Have children".
   - **Result:** an idea the author asked for is gone from the corpus. The new sheet range paragraph is a reasonable statement of the range, but it no longer says what decision 24 required.
   - **Decide:** either restore "Human life is valued both broadly and in particular" (in the range paragraph, as decided, or under Virtue 6) or record that decision 24 has been revised.

2. **The "particular children, in order" sequence was moved in one essay only.**
   - `notes/07`, decision 21: own children first, then kin broadly defined, then general heredity. *"Applied in 'Have children.'"*
   - `cb0a84c` moved those sentences out of Virtue 6 and into the Virtue 7 section in `ccp_derive.md:280`, following advice 1.4.3 of the previous findings. `ccp_sheet.md:283` and `ccp_eng.md:655` still carry the sequence under Virtue 6.
   - **Result:** the three essays now disagree about where the sequence belongs. In derive, the moved sentences also read badly in their new place (see §1.2, item 8).
   - **Decide:** follow decision 21 and restore derive to Virtue 6, so all three essays match. Or, if the author now prefers Virtue 7, move it in the sheet and eng as well.

> **Status: both items resolved (2026-10-27).**
> - **Item 1 (option B).**
>   - The sheet keeps decision 24's placement: the Principle 2 range paragraph now ends "All of them value human life both broadly and in particular."
>   - Derive and eng carry it under Virtue 6, so the derivation order holds: "Human life is the floor of goodness, and it is valued both broadly and in particular."
> - **Item 2 (option A).**
>   - Derive's sequence is back under Virtue 6, per decision 21, and removed from the Virtue 7 paragraph, which again opens "While some individuals are biologically unable…".
>   - All three essays now close the sequence with "He has no generic responsibility for children as such." Derive's Virtue 6 then reads "As Section 3 established, his own children continue him and his particular parental lineage."
>   - This also resolves §1.2 item 9.
> - **Not done:** the optional book additions (a ch08 sentence and a ch10 paragraph).

> **Note on `notes/11_evaluation_findings.md`.** That file contradicted decisions 17, 21, and 24 (it asked for Virtue 7's short sentences to be lengthened, and for the two items above to be moved). Several of its suggested rewrites contain em dashes, which the guide bans. It also cites guide sections that do not exist ("Section 2.1", "2.3", "3.1", "3.3"). Before acting on any of its remaining items, check them against the calls files.

### 1.2 Errors introduced or exposed by the edits

1. **The ch08 enemy conclusion is now historically wrong and repeats itself** (`08_ch08.md:197`, `:202`, `:206–207`). High.
   - Item 1 (`:206`) now says the regime used *"massacre, deportation, confiscatory taxation, and forced labor to destroy Greek, Armenian, and Jewish subjects."* The 1915–1923 massacres and deportations in the first bullet (`:199`) were against Armenians and Greeks (and Assyrians). The Jewish citizens enter the record only through the 1942 tax (`:200`). The new sentence attaches massacre and deportation to them.
   - Items 1 and 2 now state one truth twice: item 1 is "A state that strips a people inside its borders is their enemy," and item 2 is "the government under which you live becomes your enemy if it turns its coercive power toward your destruction." So "two vital truths" is now one truth.
   - The adding of Jewish citizens was applied to the conclusion (`:202`) but not to the premises. The lead-in (`:197`) still says "Greek and Armenian Christian populations," and the motive sentence (`:202`) still says "the Greek and Armenian minorities."
   - `:202` lists "conscription" among the state's means. The Twenty Classes conscription that supported this is no longer in the passage, so the word has no evidence under it.
   - `09_ch09.md:290` still says "the Greek and Armenian shopkeepers in 1915 and 1942."
   - **Suggested fix:** Make item 1 the specific lesson of the 1942 statute: a state can make itself the enemy through a tax law alone, by assessing a people beyond what it owns. Keep item 2 as the general rule. In `:197` and the first sentence of `:202`, use "non-Muslim minorities". Either restore the conscription evidence or drop the word. Align `09_ch09.md:290`.
   - **Status: applied (2026-10-27), as suggested:**
     - `:197`: "the long-established non-Muslim minorities". "Indigenous" was dropped because it fits the Armenians and Greeks less clearly than Istanbul's Jews.
     - `:202`: "the non-Muslim minorities". "Conscription" was replaced by "forced labor", which the 1942 bullet supports.
     - Item 1 is now the specific lesson of the 1942 statute: "A state can make itself the enemy of a people through a tax law alone…". It ties back to the Tier 4 limit on confiscatory taxes.
     - Item 2 stays as the general rule.
     - `09_ch09.md:290`: "Recall the Armenians and Greeks of 1915, the non-Muslim shopkeepers of 1942, and the Moriori…".

2. **Ch07's handicapped race no longer makes sense** **Resolved (2026-10-27) by cutting the item:** the list is now "the two ways" (shielding from failure; engineering failure), followed by the line "A third way, forcing equal outcomes by holding back the fastest, belongs to justice and is taken up in Chapter 9." The author judged that the bullet added little, since ch09 makes the same point where it belongs. (`07_ch07.md:110`). Medium.
   - "The meet cancels the faster heat, **or requires the fastest runners to start well behind the line** … The fast girl spends the race at a pace that does not test her." A runner who starts well behind must run *harder*, so the second sentence fits only the cancelled-heat option. "The fast girl" is new. The vignette has "your daughter" and "the fastest girl in town".
   - **Suggested fix:** Keep one mechanism: "The meet cancels the faster heat and runs every girl in one slow heat, so the recorded results look closer together. The fastest girl in town spends the race at a pace that does not test her."

3. **The Culver City passages in ch07 and ch09 have diverged** **Resolved (2026-10-27):** Culver City now appears only in ch09. (`07_ch07.md:110`, `09_ch09.md:231`). Medium.
   - `notes/10` call 12 put the case in both chapters deliberately. `cb0a84c` revised only ch07 ("under-worked even in the faster room"). Ch09 keeps the earlier wording: "bored even when that class exists", "A student who would do the work had been free to enter", "Remove it, and he spends the year…". A reader now meets the same case twice, with the same citation and slightly different sentences.
   - **Suggested fix:** Either make the two passages match, or let ch09 state the mechanism and refer back ("Culver City High School, Chapter 7").
   - Citation form: "*Los Angeles Times* high-school report" is vague. Per `notes/10` it is a student report in the paper's high-school section (High School Insider), 28 February 2023. Give the article title.

4. **The Virtue 7 question in ch06 now asks the impossible and repeats itself** **Applied (2026-10-27) as suggested,** with "(**Virtue 7: Form and listen**)" kept in the question so the bullet still names the virtue its heading lists. (`06_ch06.md:83`). Medium.
   - "…bind living adults to **form** those younger than themselves, **including great-grandchildren not yet born**." No one can form a person who is not yet born. One can only hand the inheritance on to him. The bullet then restates the virtue twice ("The younger listen and test what they are told. Elders, including kin and those who lead, form the people younger than themselves."). The neighbouring bullets ask "…only if he feels like consenting?". This one dropped that frame, so the parallel is broken.
   - **Suggested fix:** "Do the younger owe the elders who formed them a hearing only if they consent, and do living adults owe the young their formation, and the unborn a working inheritance, only if they consent? None of the unborn can sign a consent contract today."

5. **Ch10 now states the three violations as facts** **Applied (2026-10-27) as suggested.** (`10_ch10.md:105`). Low–Medium.
   - "The grave violations of Virtue 7 are three. Elders refuse to form those younger than themselves. The young refuse to listen…" read as statements that elders *do* refuse.
   - **Suggested fix:** "The grave violations of Virtue 7 are three: elders who refuse to form those younger than themselves; the young who refuse to listen to the experience around them and test it; and **generational theft**, in which the living treat…"

6. **The ch11 appositive is ambiguous** **Applied (2026-10-27) as suggested,** keeping the original bold on the list. (`11_ch11.md:144`). Low–Medium.
   - "The capture operates through institutions, **the teacher-training colleges, the school curricula, the youth media, and the universities**, so that…" reads as a list of five things, the first of which is "institutions". Curricula and media are not institutions in the sense of the sentence. The previous findings proposed an em dash here, which the guide forbids, and the comma was the substitute.
   - **Suggested fix:** "The capture runs through the teacher-training colleges, the school curricula, the youth media, and the universities, so that the builders' children are formed against the household that raised them."

7. **Ch04's "he" has no antecedent** **Applied (2026-10-27) as suggested** ("the heir tests what he was told…"). (`04_ch04.md:131`). Low.
   - "It is **gratitude** to the ancestors who built it, and the practice of **Virtue 7 (Form and listen)**: he tests what he was told…" The subject of the sentence is "The proper moral response". The child appears two sentences earlier.
   - **Suggested fix:** "…: the heir tests what he was told, keeps what holds, and forms his own children in it."

8. **Derive's new virtue-numbering sentence points to a list the essay does not have** **Applied (2026-10-27), with the author's change:** no reference to other documents. `:41` now reads "Each of the seven virtues corresponds to a fixed reference number. They appear in this derivation when the constraint requires them, so the order of appearance is not the order of those numbers." `:272` reads "This gives us the virtue numbered six:" and `:290` reads "This gives us the seventh virtue:". (`ccp_derive.md:41`, `:265`, `:282`). Medium.
   - "The seven virtues are numbered in **the summary list**." The derivation has no summary list. That list is in `ccp_sheet.md`. The first list of virtues in derive is the five-item list at `:249`. Line `:265` says "the sixth virtue in the summary list" and `:282` still says "our seventh and final individual virtue", so the ordinal style is now mixed.
   - **Suggested fix:** At `:41`: "Each virtue carries a fixed number, used in every companion document. The derivation reaches them in the order the constraints require, so the order of appearance is not the order of the numbers." At `:265`: "This gives us the virtue numbered six:". At `:282`: "This gives us the seventh virtue:".

9. **The sentences moved to derive's Virtue 7 do not fit their new place** (`ccp_derive.md:280`). Medium. This also bears on §1.1 item 2.
   - "A man who lacks children of his own **cares for** his kin… Lacking those, he **cares for** the people who share his general heredity…" Then: "**this seventh virtue** remains completely open". The virtue is named only at `:284`, after this. And the sentences are about care (Principle 4 tiers), not about forming the young, so they do not lead into what follows.
   - If §1.1 item 2 is resolved by restoring Virtue 6, this problem goes away. **Resolved (2026-10-27)** by §1.1 item 2, option A.

10. **Eng's Principle 2 paragraph now repeats itself, with a broken "therefore"** **Applied (2026-10-27) as suggested:** the last sentence is deleted. (`ccp_eng.md:170`). Low–Medium.
    - "His care is anchored in his own life, in his spouse, and in his own children… He has no generic responsibility to Humanity. **His self-interest therefore extends to the survival and welfare of his spouse and his own children.**" The last sentence repeats the first new sentence, and "therefore" does not follow from "no generic responsibility".
    - **Suggested fix:** Delete the last sentence.

11. **Eng's enemy passage now denies the definition before it gives it** **Applied (2026-10-27) as suggested:** "Dislike is not the definition of enmity" now follows the definition sentence, and the earlier line reads "…to a state that carries out a program to rob, subjugate, or destroy a peaceful people." (`ccp_eng.md:393`). Low.
    - "Dislike is not the definition of enmity" now comes six lines before the definition (`:399`). "Predatory actors exist, from the solitary violent criminal to **a state program** that seeks…" makes a program an actor.
    - **Suggested fix:** Put the dislike sentence after the definition at `:399`, and write "…to a state that carries out a program to rob, subjugate, or destroy a peaceful people."

12. **The 8.7 million famine citation has the wrong period and the wrong author attribution** (`05_ch05.md:82`, `13_ch13.md:111`). Medium.
    - The 2020 *Journal of Genocide Research* paper is Oleh Wolowyna, "A Demographic Framework for the **1932–1934** Famine in the Soviet Union" (single author; doi:10.1080/14623528.2020.1834741). The 8.7 million all-Union figure (Ukraine 3.92, Russia 3.26, Kazakhstan 1.26) is for **1932–1934**. Ch13 says "in 1932–1933". Ch05 attaches the figure to "the terrible winter and spring of 1932–1933". "Wolowyna and others" should be "Wolowyna" (the estimate comes from his research group, but the article has one author). Sources:
      - The article itself: [doi:10.1080/14623528.2020.1834741](https://doi.org/10.1080/14623528.2020.1834741), with a copy hosted by the [Shevchenko Scientific Society](https://shevchenko.org/research-publications/publications/a-demographic-framework-for-the-1932-1934-famine-in-the-soviet-union/).
      - The all-Union breakdown: the research team's own report through the [Canadian Institute of Ukrainian Studies](https://www.ualberta.ca/canadian-institute-of-ukrainian-studies/news-and-events/news-at-the-cius/2018/march/oleh-wolowyna-measuring-mortality-in-the-holodomor.html).
      - The team's regional loss data for 1932–1934: [Harvard Ukrainian Research Institute, MAPA](https://gis.huri.harvard.edu/population-losses).
    - Ch05 and ch13 now carry the same long sentence and citation, and in ch05 it sits inside the illustrative blockquote. Ch05 could keep "millions starved" and refer forward ("figures in Chapter 13").

    - **Grok's answer (2026-10-27).** Asked in a fork of the session that wrote `notes/10` call 18 (fork `01a17e84-2d7b-7533-979e-38098c2183ac`; original `01a17b4a…` left unchanged).
      - **Origin:** the call-18 citation came from a Wikipedia search snippet, not the article. "1932–1933" was taken from the page title and the chapter's topic. "And others" was added by mixing in the group's multi-author papers.
      - **The article:** Wolowyna, "A Demographic Framework for the 1932–1934 Famine in the Soviet Union", *Journal of Genocide Research* 23, no. 4 (2021): 501–526, published online 23 October 2020, [doi:10.1080/14623528.2020.1834741](https://doi.org/10.1080/14623528.2020.1834741). It has one author, its period is 1932–1934, and its abstract does not give 8.7 million.
      - **The Ukraine and Russia figures** come from Levchuk, Wolowyna, Rudnytskyi, Kovbasiuk, and Kulyk, "Regional 1932–1933 Famine Losses: A Comparative Analysis of Ukraine and Russia", *Nationalities Papers* 48, no. 3 (2020): 492–512, [doi:10.1017/nps.2019.55](https://doi.org/10.1017/nps.2019.55). Grok reports Table 1 as direct losses for 1932–1934 of 3,942.5 thousand (Ukrainian SSR) and 3,322.1 thousand (RSFSR, excluding Kazakhstan). I confirmed the article's existence and bibliographic details; I did not check the table values myself.
      - **The all-Union 8.73 million and Kazakhstan 1.26 million** appear in a 2018 lecture report, not in a table published by the group.
    - **For Kazakhstan**, a citable scholarly source is Sarah Cameron, *The Hungry Steppe: Famine, Violence, and the Making of Soviet Kazakhstan* (Cornell University Press, 2018). Its estimate is about 1.5 million deaths in the famine of 1930–1933, more than a quarter of the republic's population.
    - **`notes/10` call 18 annotated with the correction (2026-10-27).**
    - **Status: applied (2026-10-27), with rounded figures.** `13_ch13.md:111` now reads: "Demographers estimate that for 1932–1934 the famine caused about 3.9 million deaths above the expected number in Soviet Ukraine and about 3.3 million in the Russian republic (Levchuk, Wolowyna, Rudnytskyi, Kovbasiuk, and Kulyk, *Nationalities Papers*, 2020). In Kazakhstan, where the famine began earlier among the nomadic herders, about 1.5 million died, more than a quarter of the republic's population (Sarah Cameron, *The Hungry Steppe*, 2018)." No other file cites the 8.7 million figure.
13. **The Duke citation in ch04 is vague and I could not verify it** (`04_ch04.md:108`). Low–Medium.
    - "(Duke University School of Law, comparison of that season's elite women with boys under eighteen)". A search confirmed Thompson's 10.71 in 2017. It did not find the Duke comparison or the "124 boys" figure. Name the page: Doriane Lambelet Coleman, Center for Sports Law and Policy, "Comparing Athletic Performances: The Best Elite Women to Boys and Men". Also check whether 124 counts *boys* or *performances*. Smaller points: "recruits … interchangeable in a measured race" is evidenced by elite athletes, not recruits, and "They are not. Gestation and nursing are not interchangeable offices." says the same thing twice.

    - **Grok's check (2026-10-27, fork `01a17e84…`).**
      - **Origin:** the figure came from a search snippet of the Duke table in the earlier session.
      - **Primary source:** Doriane Lambelet Coleman and Wickliffe Shreve, *Comparing Athletic Performances: The Best Elite Women to Boys and Men*, Duke Center for Sports Law and Policy, Table 1 ("World's Best Woman v. Under 18 Boys"), [PDF](https://law.duke.edu/sites/default/files/centers/sportslaw/comparingathleticperformances.pdf).
      - **What Table 1 shows:** for 2017, the best senior woman's mark was 10.71, and "# of Boys Outperforming" is 124+, which counts **boys** (not performances), males under 18. The "+" makes it a floor.
      - **Result:** the book's sentence is accurate. Only the citation needs to name the work.
    - **Status: applied (2026-10-27).** The citation now reads "(Doriane Lambelet Coleman and Wickliffe Shreve, *Comparing Athletic Performances: The Best Elite Women to Boys and Men*, Duke Center for Sports Law and Policy, Table 1)".
14. **Ch04's evidence for the transfer claim does not support it** (`04_ch04.md:109`). Low; noted, since the author already made the call.
    - `notes/10` call 4 says the Eastern Air Lines case "does not by themselves prove the transfer claim" and that the claim is the author's. As written, the only cited case is inside a high-trust trade, so a reader can take it as a counterexample. That gives the sentence "A population that never held the method fails it more readily" no evidence of its own. If the author wants it to stand as his claim, consider saying so: "This is a claim about formation, not one proved by the case above."

    - **Grok's check (2026-10-27).**
      - **No direct evidence found:** the earlier session found no sourced case of a transferred population failing a maintenance standard. The fifty-thousand-person transfer remains the author's illustration.
      - **Fisman and Miguel**, "Corruption, Norms, and Legal Enforcement: Evidence from Diplomatic Parking Tickets", *Journal of Political Economy* 115, no. 6 (2007): 1020–1048, [doi:10.1086/527495](https://doi.org/10.1086/527495). United Nations diplomats in New York, immune from parking fines, left more unpaid tickets the more corrupt their home country. After the November 2002 enforcement (plates revoked after three unpaid tickets), unpaid violations fell by about 98 percent. The norm traveled with them, and enforcement overrode it.
      - **Algan and Cahuc**, "Inherited Trust and Growth", *American Economic Review* 100, no. 5 (2010): 2060–2092, [doi:10.1257/aer.100.5.2060](https://doi.org/10.1257/aer.100.5.2060). Trust among US-born descendants of immigrants still varies with the ancestors' country and wave of arrival. This is an inherited attitude, not a measured maintenance practice.
      - **Grok's verdict:** neither paper supports the bullet's last sentence as written ("…fails it more readily, and does not hand the method on"). Both support the narrower claim that a carried norm persists, and Fisman and Miguel also show that enforcement changes behavior.
    - **Status: applied (2026-10-27), option (a).** The Fisman and Miguel sentence and the Algan and Cahuc sentence now sit before the bullet's last sentence. The last sentence is kept as the author's claim.
15. **Smaller points in the edited passages** (Low). **All applied (2026-10-27) as suggested.**
    - `ccp_sheet.md:97`: "continue his parental lineage, which is real". The relative clause is awkward. Try "…continue him and his parental lineage. That lineage is real."
    - `02_ch02.md:25`: "For the shared physical world, access comes through the physical senses and the instruments and reasoning built upon them." The sentence has no agent. Try "we know the shared physical world through our senses and the instruments and reasoning we build on them." The forward reference ("Chapter 3 states that limit") is valid (`03_ch03.md:59–62`).
    - `04_ch04.md:66`: "approximations, not a statute" works. "(the people in daily life-support)" can read as medical. Dunbar's term is "support clique".

### 1.3 Problems from the 26 October commit (`a077b2c`) that are still present

1. **Derive's list item "3." is printed inside a paragraph** (`ccp_derive.md:175–176`). High visibility. **Fixed with §2.1** as described here. The "Honesty in speech…" paragraph is not indented and has no blank line after it. In `build/ccp_derive.html`, item 3 ("Why must effort move beyond the front door?") prints as "…thins as it moves outward. 3. **Why must effort…**" in the middle of the paragraph. Fix: indent the honesty paragraph three spaces so it belongs to item 2, and leave a blank line before "3.".
2. **"Conversely" no longer contrasts with anything** (`ccp_eng.md:89`). **Fixed (2026-10-27):** "Conversely," removed. It used to answer the deleted child-deprivation sentence. Now it follows "Sensory contact with reality and reason are linked." Replace it with "The daily operation of engineered systems shows…".

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
> `ccp_book.md` was regenerated from the chapters with the same assembly code as `build.fish`. Rendering was checked with `pandoc -t html` into a scratch directory: no run-in lists remain. `build.fish` was **not** run, because it ends by calling `../push_release`. Rebuild and publish when ready.

1. **Twenty-five lists in the essays print as run-in text.** The build reads pandoc markdown, where a list that follows a paragraph with no blank line becomes part of that paragraph. Confirmed in `build/ccp_eng.html` ("…analyzed by identifying: 1. The initial…") and `build/ccp_derive.html` ("…attack everything we have built: 1. **Material Degradation:**…"). Affected lines, where the list begins:
   - `ccp_sheet.md`: 224
   - `ccp_derive.md`: 98, 108, 139, 176, 185, 242, 249, 276, 299, 305
   - `ccp_eng.md`: 10, 188, 258, 347, 396, 427, 496, 506, 510, 548, 608, 616, 635, 682
   - **Fix:** insert a blank line before each list. The book chapters are not affected. Lists that follow a heading render correctly.
2. **The two virtues at ch10:23–24 print on one line.** "> **6. Have children**" and "> **7. Form and listen**" are adjacent blockquote lines, so they join into one paragraph ("6. Have children 7. Form and listen" in `build/ccp_book.html:4321`). Fix: put a `>` line between them.
3. **Heading duplicated in ch11** (`11_ch11.md:34–36`). List item 4 is a bare bold title, followed by a blockquote that repeats the same bold title.

### 2.2 Factual accuracy

**Medium (a reader can check these and will find them wrong):**

> **Status: items 1–10 applied (2026-10-27).** Per item:
> 1. **Singapore:** the "instead of prison" contrast is gone. The passage now states the s 394 penalty (five to twenty years and at least twelve strokes) as an example of a penalty fixed by the offense.
> 2. **Typhus:** became **dysentery**, not typhoid. Typhoid usually takes one to two weeks to appear, and the vignette kills the traveler "by the end of the week"; dysentery fits that timing in any century.
> 3. **Hegel:** split from Marx. Hegel is now charged with historicized truth and with setting the state above household and civil society ("the actuality of the ethical Idea", *Philosophy of Right*, section 257), and the bullet says he defended property and the family. Marx keeps the rest. "New Socialist Man" is now attributed to the later regimes, since Marx did not use the phrase.
> 4. **Augustine:** the passage names both works, *On Christian Doctrine* 2.40 and *The Literal Meaning of Genesis* 1.19. "Four centuries later" became "About three and a half centuries later" (Romans, about AD 57, to Augustine's works of about 397–415).
> 5. **Budapest Memorandum:** now "assurances … that they would respect Ukraine's existing borders and refrain from the threat or use of force". It says no signatory undertook to fight for them, so Condition B never existed and Condition A lasted as long as Russia's interest. The question-then-"Nothing." construction is gone.
> 6. **Stolypin and Durnovo:** see item 6 below for the detail.
> 7. **Varlık Vergisi:** now "more than a thousand".
> 8. **Spousal privilege:** "constitutional protections" became "the rules of evidence in many modern legal systems", citing the practice: Police and Criminal Evidence Act 1984, s 80 (England); *Trammel v. United States*, 445 U.S. 40 (1980).
> 9. **UN rights:** the list now quotes the instruments: food, housing, medical care, and social security (UDHR arts. 22, 25); "periodic holidays with pay" (art. 24); higher education made free by "progressive introduction" (ICESCR art. 13(2)(c)). Internet access is described as a 2016 Human Rights Council call (A/HRC/RES/32/13), so the "fiber-optic cables" question that follows still has a referent.
> 10. **Kant:** now "a categorical duty not to lie to him, even if the truth lets him find and murder your friend". Kant's case is a friend; the added "or child" is gone.
>
> `ccp_book.md` was regenerated from the chapters. The guide's Part A audit shows no new hits.

1. **Singapore caning is not an alternative to prison** (`09_ch09.md:243`). The passage contrasts caning with "Western prisons that confine a young offender … for five years" and says it "returns the chastened offender to work … the following week." Under Singapore's Penal Code, robbery with hurt (s 394) carries **5 to 20 years' imprisonment and at least 12 strokes**. For most caning offences, caning is imposed together with a prison term. Primary sources: Penal Code 1871, s 394, on Singapore Statutes Online (sso.agc.gov.sg/Act/PC1871); Singapore Police Force releases that state the statutory penalty when charging under s 394 ([Feb 2024](https://www.police.gov.sg/Media-Hub/News/2024/20240226_two_men_arrested_for_robbery_with_hurt), [Apr 2025](https://www.police.gov.sg/Media-Room/News/20250427_two_men_arrested_for_robbery_with_hurt_within_four_hours)). Keep the point about symmetry (same statute for the banker's son and the dockworker) and drop the "instead of prison" contrast.
2. **Typhus versus typhoid** (`02_ch02.md:55`). Drinking stagnant marsh water causes typhoid fever (or cholera or dysentery). Typhus is carried by lice and fleas. Write "typhoid fever".
3. **Hegel did not propose abolishing private property** (`12_ch12.md:66`). The bullet grouping Hegel with Marx says both promised "that abolishing private property would produce a selfless 'New Socialist Man'" and replaced law "with class liquidation." Hegel's *Philosophy of Right* defends property. Split the bullet: Hegel for historicized truth (Principle 1), Marx for the rest.
4. **The Augustine warning is from a different book** (`03_ch03.md:15`). The "gold of the Egyptians" passage is in *On Christian Doctrine* (II.40). The warning against Christians making false claims about astronomy is in *The Literal Meaning of Genesis* (I.19.39). Name both works.
5. **The Budapest Memorandum gave "assurances," not guarantees** (`09_ch09.md:341`). "written security assurances … **guaranteeing** Ukraine's existing borders". The signatories undertook to *respect* the borders. The gap between assurances and guarantees is the point of the example, so the wording matters. Also, Ukraine's own warheads are not "Condition B" as the chapter defines it (a stronger power ready to punish the violator).
6. **Durnovo's interval and Stolypin's "twenty years"** (`13_ch13.md:87`, `:94`). The "give the state twenty years of internal and external peace" line comes from Stolypin's 1909 newspaper interview, not the 10 May 1907 Duma speech the bullet attaches it to. Durnovo's February 1914 memorandum came about four months before Sarajevo (28 June), not five.
   - **Found while applying the fix:** the "sturdy and strong" line is not from the 10 May 1907 speech either. The text of Stolypin's Duma speech on the land bill of **5 December 1908** reads "оно делало ставку не на убогих и пьяных, а на крепких и на сильных": the government, in enacting the law of 9 November 1906, "wagered not on the wretched and the drunken, but on the sturdy and the strong". So *убогих* is "the wretched", not "the weak". The same speech calls the strong personal owner the barrier to revolution.
   - **Applied:** the bullet now dates the speech 5 December 1908 (Old Style, as elsewhere in the chapter), quotes "the wretched and the drunken", and attributes the twenty-years line to the October 1909 interview.
   - **Sources:** State Duma, Third Convocation, stenographic record of 5 December 1908. The speech is collected in *Избранные выступления П. А. Столыпина в Государственной думе и Государственном совете, 1906–1911* (Izdanie Gosudarstvennoi Dumy, 2012); I checked the wording against a transcription of that text. The interview appeared in the Saratov newspaper *Volga*, 1 (14) October 1909, quoted in the [Great Russian Encyclopedia](https://bigenc.ru/c/daite-gosudarstvu-dvadtsat-let-pokoia-vnutrennego-i-vneshnego-i-vy-ne-uznaete-nyneshnei-ro-4449af).
7. **The Varlık Vergisi deportation count needs a source** (`08_ch08.md:200`). "thousands … shipped … to forced labor battalions". The best-supported count is 1,229 sent to Aşkale, with 21 deaths among those sent to work. That count is given in a written question filed in the Turkish Grand National Assembly ([TBMM document](https://cdn.tbmm.gov.tr/KKBSPublicFile/D27/Y3/T10/WebOnergeMetni/fe5c70dd-6537-4667-8415-c0fdb768c38f.pdf)). The first-hand account is by the official who ran the tax in Istanbul: Faik Ökte, *Varlık Vergisi Faciası* (1951), in English as *The Tragedy of the Turkish Capital Tax*, trans. Geoffrey Cox (Croom Helm, 1987). Check Ökte's own figures before citing. The governing statute is Law No. 4305 (*Resmî Gazete*, 12 November 1942). Write "more than twelve hundred", or cite the count.
8. **Spousal privilege is not constitutional** (`09_ch09.md:16`). The spousal testimonial privilege is common-law and statutory. Write "rules of evidence in modern law".
9. **The UN rights catalog is overstated** (`09_ch09.md:274`). The UDHR names housing, medical care, social security, higher education "on the basis of merit", and "periodic holidays with pay". It does not name a "furnished dwelling", a "guaranteed income", or "high-speed internet". The internet item is a non-binding 2016 resolution on access. In a book that asks for literal accuracy, quote the articles.
10. **Kant's murderer case** (`12_ch12.md:58`). "a categorical duty to tell the murderer the exact truth **so that he may enter the house and murder them**" gives Kant a purpose he does not have. Write "…the truth, even if the truth lets him find them."

**Low:**

> **Status: all applied (2026-10-27).**
> - **Wording fixes:**
>   - "trenches … more than a century ago"
>   - 1790 actions credited to the National Assembly
>   - anvils on their stumps; "heat and temper color"
>   - "some twenty-five thousand days" (both places)
>   - "an infant born ten thousand years ago"
>   - *A Theory of Justice*; "premise of natural goodness" in place of "noble savage"
>   - Chernobyl made consistent in the book and derive: "a test already delayed … would not slip to a later shutdown"; "reactor core"
>   - Rome: "some tens of thousands by the end of the sixth century", with the Forum "later" a pasture
>   - Sparta: "never recovered its former standing … no notable work of science or philosophy"
>   - Ibn Rushd "two generations after"; Faraday, Maxwell, and Mendel moved to "the science that followed"
>   - "31-knot northwest wind"; the mir "approval"
> - **Ch12 Confucianism:** the misattributed sentence about testing what one is told is deleted.
> - **Ch12 taqiyya:** *kitman* removed. *Taqiyya* is now grounded in Qur'an 3:28 and 16:106, with the law of war's permission to deceive the enemy stated separately.
> - **Ch12 Dugin:** the quotation is verified and now cited as Dugin's post on X, 10 July 2026 (@AGDugin, status 2075586373277360520).
> - **Ch12 MacIntyre:**
>   - The name is corrected to **Auron** MacIntyre.
>   - The unverifiable "for losers" quotation is replaced with his verified phrase, calling the friend/enemy distinction "the true essence of the political" (post on X, 4 August 2025, status 1952328296408007142).
>   - The rest of the bullet is now framed as the book's analysis of that view.
> - **Bible quotations:** all now use the **NET Bible** (text from bible.org's NET API). This covers ch03 Romans 1:20; ch09 Daniel 6:8; ch12 Matthew 7:1–5, 7:12, 7:15–16, 5:39, 19:11, 1 Timothy 5:8, Romans 13:4, 1 Corinthians 7:6–7, and Genesis 1:28.
> - **Where NET wording breaks a guide rule,** the quotation is trimmed with an ellipsis or ends early:
>   - the em dashes in Romans 1:20
>   - "You hypocrite!" in Matthew 7:5
>   - "multiply!" in Genesis 1:28

- `10_ch10.md:86`: "the soldiers who died in the trenches eighty years ago". Trench warfare is 1914–1918, more than a century ago. Eighty years ago is the Second World War.
- `10_ch10.md:78`: "In 1790, when the French Jacobins abolished France's laws, churches, courts, and property rights". In 1790 the actors were the National Constituent Assembly (church lands, the Civil Constitution of the Clergy). The Jacobins governed in 1793–94.
- `07_ch07.md:79`: "hammers, tongs, and **anvils** hang … on the shop wall". An anvil weighs 100 pounds or more and sits on a stump. `:81`: straw yellow and peacock blue are temper colours, not "quench colors". Cherry red is a heat colour.
- `04_ch04.md:63` and `06_ch06.md:160`: "some twenty-five thousand adult days". 25,000 days is about 68 years, which is a whole life, not its adult part. Use "some twenty-five thousand days", or "some twenty thousand adult days".
- `01_ch01.md:42`: "the first infant born ten thousand years ago". Humans are far older than that. Write "an infant born ten thousand years ago".
- `04_ch04.md:113`: the title is *A Theory of Justice*. `04_ch04.md:23`: "noble savage" is Dryden's phrase. Rousseau never used it.
- `02_ch02.md:60` and `ccp_derive.md:50`: Chernobyl is "an administrative deadline and … bureaucratic bonuses" in the book and "an administrative quota" in derive. Pick one. "Bonuses" needs a source. The RBMK has no "reactor vessel" in the usual sense, so write "reactor core".
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

1. **Ch02 promises four distinctions and gives three** (`02_ch02.md:82`). "a civilization must master **four** practical distinctions". Subsections 1 to 3 follow, then the chapter ends. This has been the case since the first draft (`eaf9f59`). Either supply the fourth (the obvious candidate is "Precision depends on the task", which the essays carry under correspondence) or change the count to three. The chapter is now at `02_ch02.md:86`. **A proposal and its evaluation are in §2.3a below.** **Status: applied (2026-10-27).** Draft 4 is now the fourth distinction, with the author's final sentence ("Truth can still be known within uncertainty and within a bound."), at the end of `02_ch02.md`.
2. **Ch06 counts five failures but lists four** (`06_ch06.md:50`). "it fails on **five separate foundational points**". Item 5 is "Where Consent Actually Belongs", which places consent rather than describing a failure. Write "four failures, and one proper place". **Status: applied (2026-10-27).** The count now reads "**four separate foundational points**". Former item 5 is ordinary prose after the list: its title line is dropped, its three sub-bullets are joined into one paragraph, and "Keep the hierarchy clear…" is a second paragraph. Apart from punctuation, the wording is unchanged.
3. **The celibacy test is applied to some religions and not to Christianity** (`12_ch12.md:134`, `11_ch11.md:140`). Buddhism, Jainism, and the Shakers fail the test "if every member attains the highest ideal… the society ends". The same test applies to Christian monastic and clerical celibacy (1 Corinthians 7), and the chapter does not apply it there. The chapter opens by promising "the **same inspection checklist** to every … religious framework" (`:9`), so a reader will notice. Either explain why it does not apply (for example, Christianity holds marriage and children as the ordinary vocation and celibacy as a vocation for some, not the summit every member should reach), or apply it.
   - **Author's position:** Christian celibacy is optional, a calling for some. It is not for all, nor should it be, and the blessings to Adam and to Noah counter any reading otherwise. The highest Christian good is obedience to God and salvation, not self-enlightenment, and celibacy is never raised to that level.
   - **Status: applied (2026-10-27).** `12_ch12.md:136` now has a bullet in the Buddhism section, "The same test applied to Christianity". It:
     - places the highest good in obedience to God and salvation, open to married and unmarried alike;
     - quotes Matthew 19:11, 1 Corinthians 7:7, and 1 Corinthians 7:6 (now in the NET Bible, like the chapter's other quotations);
     - cites the blessing of Genesis 1:28, repeated to Noah in Genesis 9:1;
     - concludes that the society continues if every Christian attains the highest good;
     - notes that the Shakers departed from that teaching, pointing to Chapter 11.
   - **Wording choices.** "Even where a church honors consecrated celibacy as a high calling, it is not required of all and is not the condition of salvation." This covers the Council of Trent, which held celibacy "better and more blessed" than marriage (Session 24, canon 10). It also covers Paul's "I would that all men were even as I myself" (1 Corinthians 7:7a), which the quoted second half of the verse answers.
4. **Principle 3 is used as the label for decay** (`ccp_derive.md:120`, `07_ch07.md:63`, `09_ch09.md:232`). "natural decay … (Principle 3)", "Principle 3 (physical things degrade…)", "Recall Principle 3: … failure and death require no effort". Principle 3 is the standard of goodness. Decay is the condition under which "kept" is required. Write "(the decay that Principle 3's *kept* answers)", or cite Principle 1.
   - **Author's direction:** Principle 3 is the opposite of decay, the guard against idolizing it. The rule is that decay is filed under Principle 1, and Principle 3 is the requirement to keep and build against it.
   - **Status: applied (2026-10-27)** at four places: the three above plus `ccp_sheet.md:263` (Excellence, "things decay (Points 1 and 3)"), found by search.
   - **Wording now:** decay sits with Principle 1, and Principle 3 appears as the requirement that what is built be kept. In ch09 the line reads "Recall why Principle 3 must require that life be kept and built: …".
   - **Not added:** the proposed sentence in ch06 stating the guard explicitly, per the author.
5. **A foreign state is put in Tier 4** (`09_ch09.md:99`). "treating Tier 4 and Tier 5 foreign state operations as if they were domestic citizens". Tier 4 is one's own country. Write "foreign state operations". **Status: applied (2026-10-27), option 2 of three offered:** "It is placing the operations of a foreign state in Tier 4, among fellow citizens, when they belong at most in Tier 5 and, once directed against the host, under "Enemies are real.""
6. **The particularity sentence in derive and eng is ambiguous** (`ccp_derive.md:87`, `ccp_eng.md:217`). This is the author's wording (`notes/08` decision 3). "Tends to reproduce that society" is followed by "If the people are able, they adopt the culture of the place they enter." The two sentences agree only if "able" means "able to adopt". Consider "able and willing to adopt the method of the place".
   - **Author's direction:** not all people are willing and able to adopt the standards that make the new culture work.
   - **Status: applied (2026-10-27)** in both files. The text now reads: "…tends to reproduce that society and culture in a different land, because not all people are willing and able to adopt the standards that make the new culture work. Those who are willing and able adopt the culture of the place they enter. Those who are not recreate the failures of the culture they carried." The first sentence now gives its reason, so the two no longer pull against each other.
7. **Ch05's diagnostic question uses the wrong word** (`05_ch05.md:24`). "Is it constructed in form, or is it constructed in substance?" The chapter's thesis is that substance is *discovered*. Write "constructed in form, or discovered in substance?". **Status: applied (2026-10-27).**
8. **Small internal mismatches:**
   - `11_ch11.md:31`: "October-to-April" pantry, but it holds three months of food.
   - `11_ch11.md:13–17`: three "tendencies" are announced, but there are five sections.
   - `11_ch11.md:185`: the London and Singapore bullet is an unfinished conditional with no consequence.
   - `11_ch11.md:108`: bins are "wooden", then "stone".
   - `13_ch13.md:130–134`: the example is called "Two Demolition Crews…" in the text and "Two Crews…" in the blockquote, and the crews "tie their ropes to the same timber ram and pull", but a ram is swung, not pulled.
   - `14_ch14.md:59`: "The ground to the left is quicksand", but the walker stepped straight ahead.
   - `08_ch08.md:191`: "lost at the cost of his flock" is circular.
   - **Status: all applied (2026-10-27):**
     - Ch11 pantry: renamed "The Rotated Three-Month Winter Pantry".
     - Ch11 introduction: now lists four tendencies (Parasitism and Fatalism split), then says a fifth section describes the mechanism that administers them, Managerialism.
     - Ch11 London and Singapore: the bullet is completed: "They did not. London kept its fire watches, factories, and railways working through the Blitz, and Singapore built in one generation what the prevailing theories of the 1960s said it could not (Chapter 13)."
     - Ch11 storehouse: "wooden" dropped from the row of bins.
     - Ch13 example: named "Two Crews Pulling Down One Wall" in both places. The crews now "pull down" the wall by hooking their ropes over the same coping stones and hauling together.
     - Ch14 marsh: the warning is "This ground is quicksand."
     - Ch08 shepherd: "lost because he had given up the means to defend it."

### 2.3a Proposal: the fourth distinction in Chapter 2 (not applied)

**Where it would go:** at the end of `02_ch02.md`, after "…a coin-flipper still cannot be entrusted with a hospital." That is where the chapter now ends, so the new subsection would be its last section.

**What it draws on:**
- The section heading already promises it: "Reason Guides, Reality Vetoes: Models, Practices, and **Precision**". Distinctions 1 to 3 cover practices, models, and claims. None covers precision.
- The essays' range of views for "Truth is what corresponds to reality" says: *"Correspondence may be exact or approximate depending on the task. A machinist … thousandths of an inch; a historian … the surviving records and physical evidence"* (`ccp_sheet.md:79`). Eng says the same in more detail, with a surveyor and Newtonian mechanics (`ccp_eng.md:136`). The book has no statement of this anywhere.

#### Draft 1

> #### 4. Match the Precision to the Task: Required Tolerance Versus False Exactness
>
> No claim about the world is checked to infinite exactness. Correspondence may be exact or approximate depending on the task, and the task sets how close the match must be. A machinist turning an engine piston must hold its diameter within thousandths of an inch, because a fit outside that band seizes in the bore or wears it out early. A surveyor marking a farm boundary must be right within a fraction of a foot. A historian describing a battle must match the surviving orders, letters, and graves, and need not know where each soldier stood. Each statement can be true. None of them is true to the tolerance of the others, and none needs to be.
>
> Two opposite errors come from ignoring the tolerance.
>
> * **Demanding exactness the task does not require.** The radical skeptic described earlier in this chapter treats any imperfection in a measurement as proof that nothing is known. The builder answers with a margin. He does not know the breaking load of each oak beam to the pound; he knows it within a range, sizes the span to carry several times the heaviest wagon expected, and the bridge stands. A claim that holds within the tolerance its task requires is true for that task.
> * **Claiming exactness the evidence does not support.** A harvest forecast given to the last bushel, or a thirty-year projection of a city's population given to the last resident, states more than anyone knows. The extra figures are not additional truth. They misreport how much is known, and they invite decisions that the evidence cannot carry.
>
> Before accepting or rejecting a claim, then, ask two questions: what tolerance does the task require, and what tolerance does the evidence support? The claim is fit for use when the evidence is at least as fine as the task demands. Newton's mechanics passes that test for a bridge and fails it for a satellite clock, which is the incomplete model of the second distinction stated as a number.

#### Evaluation of draft 1

**What it adds:**
1. **It brings the book into line with the essays.** The essays treat "exact or approximate depending on the task" as part of Principle 1's range of views. The book states the rest of that range (one shared world, at `02_ch02.md:19`; revelation's limit, at `:25`) but not this. This is the main reason to include it.
2. **It names a failure the book does not name anywhere else.** False exactness, stating more than the evidence supports, is a breach of Virtue 1: it misreports how much is known. It also gives later chapters a term to lean on, for example ch11's managerial mode that "redefines the metric".
3. **It keeps the promise of the chapter's count ("four") and its heading ("Precision").**

**Weaknesses:**
1. **The first bullet re-answers the skeptic.** Ch02:39–41 already answers him, with the elevator and Reid. Only the bridge-margin example is new.
2. **It seems to conflict with distinction 3.** Distinction 3 praises the narrow wind forecast ("between 28 and 34 knots"). The second bullet condemns figures finer than the evidence. A reader can fairly ask why the narrow forecast is not false exactness, and draft 1 does not say.
3. **The closing sentence leans on distinction 2.** It is the third appearance of Newton in the chapter, and "stated as a number" is loose.
4. **Length.** About 330 words in a chapter of about 3,900.

**Facts checked:**
- A piston outside its clearance band seizes (too tight) or slaps and wears (too loose).
- Boundary surveys to a fraction of a foot match eng's own example.
- Timber spans sized to several times the expected load match historic practice; the factor of safety was about four or more.
- The draft has no em dashes, banned openers, or symbols.

#### Draft 2 (revised for the weaknesses)

> #### 4. Match the Precision to the Task: Required Tolerance Versus False Exactness
>
> Correspondence may be exact or approximate depending on the task, and the task sets how close the match must be. A machinist turning an engine piston must hold its diameter within thousandths of an inch, because a fit outside that band seizes in the bore or wears it out early. A surveyor marking a farm boundary must be right within a fraction of a foot. A historian describing a battle must match the surviving orders, letters, and graves, and need not know where each soldier stood. Each of these statements can be true, though none is true to the tolerance of the others.
>
> Two opposite errors come from ignoring the tolerance.
>
> * **Demanding exactness the task does not require.** This is the radical skeptic's error in practical form. The bridge builder does not know the breaking load of each oak beam to the pound; he knows it within a range, sizes the span to carry several times the heaviest wagon expected, and the bridge stands. A claim that holds within the tolerance its task requires is true for that task.
> * **Claiming exactness the evidence does not support.** A harvest forecast given to the last bushel, or a thirty-year projection of a city's population given to the last resident, states more than anyone knows. The extra figures are not additional truth; they misreport how much is known and invite decisions the evidence cannot carry. The forecaster in the third distinction earned credit for a narrow range because the range could have failed and did not. A figure stated more finely than its method can support has not earned that standing.
>
> Before accepting or rejecting a claim, ask what tolerance the task requires and what tolerance the evidence supports. The claim is fit for use when the evidence is at least as fine as the task demands.

Draft 2 is about 290 words.
- The skeptic is now one linking sentence, not a second rebuttal.
- Two sentences reconcile the bullet with distinction 3.
- The Newton closing line is dropped; distinction 2 already makes that point.

#### Verdict

Draft 2 adds enough to include. The deciding reason is that the book is missing a piece of Principle 1 that the essays carry. False exactness is a smaller gain. If the author would rather keep chapter 2 shorter, the alternative is a one-word change ("four" to "three"). The chapter's argument stands without the new point, but the book would then omit the tolerance of correspondence entirely.

#### Draft 3 (after the author's review, not applied)

The author's notes on draft 2:
- Keep the first example, and close it with a sentence that applies it to truth.
- Replace the second example with an engineering one: the arithmetic is exact, but the measurement errors are carried through to the answer by derivatives.

The opening paragraph and the closing rule are unchanged from draft 2.

> * **Demanding exactness the task does not require.** This is the radical skeptic's error in practical form. The bridge builder does not know the breaking load of each oak beam to the pound; he knows it within a range, sizes the span to carry several times the heaviest wagon expected, and the bridge stands. A claim that holds within the tolerance its task requires is true for that task. The builder's statement that the span will carry the heaviest wagon corresponds to reality, and so it is true, though he cannot name the breaking load of any single beam.
> * **Claiming exactness the evidence does not support.** The same builder computes the bending strength of an oak beam from its measured width and depth by an exact formula, and the arithmetic returns an answer to as many places as he cares to carry. The places are not knowledge. Each measurement carries an error, and the error must be carried through to the answer: for each measurement, the builder works out how fast the answer changes as that measurement changes (the derivative of the answer with respect to that measurement), multiplies that rate by the error in the measurement, and adds the results. Bending strength grows with the square of the depth, so a depth known to within one part in a hundred yields a strength known only to within about two parts in a hundred, from the depth alone, before the errors in the width and in the wood itself are added. A strength reported to the last pound states more than the measurements know, and invites a decision the evidence cannot carry. The forecaster in the third distinction earned credit for a narrow range because the range could have failed and did not; a figure stated more finely than its measurements support has not earned that standing.

**Notes on draft 3:**
- **Shared beam.** Both bullets now use the same oak beam. The first says the builder does not know the breaking load to the pound. The second shows why the measurements cannot give it to him.
- **The physics checks out.**
  - A rectangular beam's bending capacity varies with its width times the square of its depth.
  - So the relative error from the depth is doubled.
  - The material's own strength varies more than any tape measurement, which the clause "and in the wood itself" covers.
- **Error rule.** "Adds the results" is the conservative worst-case rule. The root-sum-square rule gives a smaller combined error, and either is consistent with the text.
- **Length.** The subsection grows to about 380 words.
- **A simpler alternative for the second bullet** (no beam formula): a field measured by chain to the nearest foot, a thousand feet by five hundred. The area is five hundred thousand square feet, and the error carried through is about fifteen hundred square feet each way. Reporting it to the square inch is false exactness.

#### Draft 4 (after the author's second review; applied with one change)

> **Applied to `02_ch02.md` on 2026-10-27.** One change at the author's request: the final sentence "A claim presented as having no error, and a claim dismissed merely because it has one, are both mishandled." was replaced by "Truth can still be known within uncertainty and within a bound." `ccp_book.md` was regenerated. The new section passes the Part A audit and renders correctly.

The author's direction on draft 3:
- The fault is not reporting a figure "to the last pound". On one side it is implying exact understanding with no error; on the other, treating an error or unknown as a lack of knowledge.
- Knowing in part is real knowledge. Error and unknowns coexist with testable knowledge, and reducing them is progress.

> #### 4. Know in Part: Bounded Error Versus Ignorance
>
> Correspondence may be exact or approximate depending on the task. Every measurement carries an error, every calculation built on measurements inherits that error, and every model leaves something out. None of this cancels the knowledge. A machinist who holds a piston within thousandths of an inch, a surveyor who fixes a boundary within a fraction of a foot, and a historian who matches his account to the surviving orders, letters, and graves each know something true about the world, and each knows it in part. Knowing in part is knowing.
>
> Two opposite errors come from mishandling the part that is not known.
>
> * **Treating error as ignorance.** This is the radical skeptic's error in practical form. The bridge builder does not know the breaking load of each oak beam to the pound; he knows it within a range, sizes the span to carry several times the heaviest wagon expected, and the bridge stands. The unknown remainder is real, but it is bounded, and a bounded unknown is not ignorance. The builder's statement that the span will carry the heaviest wagon corresponds to reality, and so it is true, though he cannot name the breaking load of any single beam.
> * **Treating knowledge as exact.** The same builder computes the bending strength of an oak beam from its measured width and depth by an exact formula. The arithmetic is exact; the answer is not. The errors in the measurements must be carried through to the answer: for each measurement, the builder works out how fast the answer changes as that measurement changes (the derivative of the answer with respect to that measurement), multiplies that rate by the error in the measurement, and adds the results. Bending strength grows with the square of the depth, so a depth known to within one part in a hundred yields a strength known to within about two parts in a hundred from the depth alone, before the errors in the width and in the wood itself are added. A result presented as though it had no error claims an understanding no one has, and it invites a margin thinner than the error. Stating the error with the result is part of stating the truth. The forecaster in the third distinction did this: he gave a range, not a single number, and the range could have failed.
>
> The error can often be made smaller. A finer gauge, a repeated measurement, a larger sample, or a test to destruction of beams cut from the same lot each narrows the range. Each narrowing is real progress, and the knowledge held before it was still knowledge. Narrowing has a cost, and the task decides how far it is worth carrying. The unknown is never removed entirely, and it does not need to be.
>
> Before accepting or rejecting a claim, ask how large its error is, whether the error has been stated, and whether it is small enough for the task. A claim with a stated, bounded error that the task can tolerate is sound knowledge. A claim presented as having no error, and a claim dismissed merely because it has one, are both mishandled.

**Notes on draft 4:**
- **Changes from draft 3:**
  - The title moves from precision to partial knowledge.
  - The second error is now the *implication* of no error, not excess digits.
  - A paragraph on reducing error is added.
  - The forecaster now illustrates the right practice (stating a range) instead of being reconciled with a contradiction.
- **Fit with the rest of the book.**
  - It is the measurement-level form of "knowable … with usable confidence" (`02_ch02.md:39`).
  - Distinction 2 makes the same point about models.
  - It prepares Virtue 5 in ch03, "our instruments have tolerances".
- **Alternative title:** "Distinguish Bounded Error from Ignorance", parallel to distinction 2.
- **Length:** about 420 words.

### 2.4 Prose and compliance with the guide

> **Status: items 1–8 applied (2026-10-27); item 9 needed no change.** The changes were applied by a script that required every original passage to match exactly once. Per item:
> 1. **Titles.** Twelve chapters retitled to one colon:
>    - Ch1, *The Human Condition, Technology, and Civilization*
>    - Ch2, *Reality and Truth (Principle 1)*
>    - Ch3, *Scripturalism, Occasionalism, and Constructivism as Rivals to Correspondence*
>    - Ch4, *The Human Builder (Principle 2)*
>    - Ch6, *The Minimum Standard of the Good (Principle 3)*
>    - Ch7, *Ownership, Excellence, Beauty, and Humility in Keeping and Building*
>    - Ch8, *Ordered Obligation and the Reality of Enemies (Principle 4)*
>    - Ch9, *Justice Among Cooperating Households (Principle 5)*
>    - Ch10, *Children, Inherited Debt, and the Seventy-Year Replacement Horizon*
>    - Ch11, *Mechanisms of Civilizational Decline*
>    - Ch13, *Historical Cases from Singapore, Russia, and Iran*
>    - Ch14, *The Daily Practice of Civilization*
>
>    The verbatim principles no longer sit in the titles, so each principle chapter now quotes its principle in the body as "Principle N of Reality: …". Ch2 gained that quote at the end of "Truth Is What Corresponds to Reality". Ch9's "> 5. Justice is…" was brought into the same form. `07_ch07.md:162` is now "Virtue 5: Be Humble".
> 2. **Reader-directing openers.** "Look back at" and "Step back" are gone. So are "Imagine" (ch02), "Drive a car", "Listen to", and "Stand in the barn stall" (ch05), and "Walk into…" and "Now follow…" (ch07). Each is now declarative or uses "Suppose". `ccp_derive.md:132` "Walk into two different…hangars" was the same pattern and was changed too.
> 3. **Question answered by a fragment.** All six listed cases are rewritten as statements. In ch11 the closing rule now says "two structural rules", which is the number it lists. Ch02's Lysenko paragraph lost only "What happened?"; its content is unchanged, per `notes/10` call 2.
> 4. **Repetition.**
>    - Ch10:70 keeps the elders' sentences; the younger's sentences now appear once, at the Burke answer.
>    - Derive's Virtue 6 now begins "As Section 3 established, …".
>    - The sheet's enemy paragraph drops the repeated "you do not keep and build the enemy…".
>    - Ch05 now says "millions starved (Chapter 13 gives the demographic figures)". Chapter 13's citation still says 1932–1933 and "Wolowyna and others"; that is §1.2 item 12, still open.
> 5. **Grammar.** "While … whereas" became "Although …, and …" in derive, eng, and ch06. Also fixed: "its each", "on remaining non-Muslim", "either" with three items, and "on which" (now "on whom"). In ch08 and the sheet, "thereby gain" became "owing a stranger the truth does not give him the child's claim".
> 6. **Colloquial words.** "clean illustration" became "clearest", "same shape" became "same pattern" or "meet the same test", and "listening to what they were given" became "heeding what their elders handed down".
> 7. **Book voice.** "before any census of regimes" is cut. "authoritative statement/specification" is now "statement". "(The Precepts Matrix)" is removed from ch14.
> 8. **Style.** "35,000 feet" is now in words. "Strip #4" is now "his fourth strip". "**Virtue**" is lowercase in ch01. Derive's units are now "nearly five hundred miles per hour".
>
> Checks:
> - `ccp_book.md` was regenerated.
> - The Part A audit shows no new violations.
> - The guide's "Math / symbols" regex flags the indented `>` lines of the ch11 Plank Bridge quote (nested under list item 4 by the §2.1 fix). That is a false positive like the existing `12_ch12.md:179` hit.
> - Pandoc renders all lists and quotes correctly.
> - Not changed, because they were not listed: other rhetorical "…? No." lines (`02_ch02.md:13`, `06_ch06.md:79`, `ccp_derive.md:240`) and "Test …" openers.

1. **Rule 9 (titles).** Twelve of fourteen chapter titles use the banned "Chapter N: Topic: Subtitle" form (all except ch05 and ch12). The five principle chapters also end the heading with a full stop. Section heading `07_ch07.md:162` is "Virtue 5: Be Humble: Paired With…". This was never addressed. The principle chapters could take a single descriptive title and give the principle as the first block quote, as ch10 does.
2. **Rule 5 (directing the reader) and what the regex misses.** `07_ch07.md:63` "**Look back at** Principle 1…" is a direct violation; the audit regex looks only for "Look at" and "Look closely". Close relatives of the banned openers: `01_ch01.md:27` "Step back and ask", `05_ch05.md:33` "Drive a car onto…", `:40` "Listen to a Frenchman…", `:78` "Stand in the barn stall…", `07_ch07.md:17` "Walk into…", `02_ch02.md:53` "Imagine…". Some second-person vignettes are deliberate; at least revise "Look back at" and "Step back".
3. **Rule 8 (question followed by a fragment answer).** `02_ch02.md:59` "What happened? The wheat seeds…", `13_ch13.md:88` "What occurred…? **The policy succeeded.**", `13_ch13.md:140`, `11_ch11.md:114` "…stop the rats from feeding? No.", `11_ch11.md:234–235` "**How does a civilization correct Managerialism…?** By enforcing…", `05_ch05.md:90`.
4. **Repetition:**
   - `10_ch10.md:70` and `:88` repeat the five Virtue 7 sentences ("They seek the experience…They listen. They test… They do not have to agree with all of it.") within the same section. Keep them once. Line 88 is the better home, as the answer to "What do the younger owe the elders?"
   - `ccp_derive.md:72` and `:265` both say "continues his particular parental lineage. That lineage is real." and "He has no generic responsibility to Humanity." In a sequential derivation, the second could say "As Section 3 established, …".
   - `ccp_sheet.md:175` and `:181` both say "You do not keep and build an enemy; you defend against him, deter him, or defeat him." `:181` also restates the two examples just given at `:177–179`.
   - `05_ch05.md:82` and `13_ch13.md:111`: the same famine sentence and citation (see §1.2 item 12).
5. **Grammar:**
   - "**While** a burned building can be rebuilt … **whereas** a murdered man cannot…, physical property is…" uses two subordinators and leaves no main clause for the first: `ccp_derive.md:114`, `ccp_eng.md:283`, `06_ch06.md:161`. Drop "whereas" and write "While a burned building can be rebuilt …, and a murdered man cannot be brought back, physical property is…".
   - `ccp_eng.md:265`: "reduces **its each** successive generation" should be "reduces each successive generation".
   - `09_ch09.md:56`: "the jizya was collected as the tax **on remaining non-Muslim**" should be "as the tax on those who remained non-Muslim".
   - `03_ch03.md:57`: "either … , … , or …" is used with three items.
   - `11_ch11.md:140`: "children **on which** continuance depends" should be "on whom".
   - `08_ch08.md:48` and `ccp_sheet.md:161`: "a stranger does not **thereby** gain the child's claim". Nothing in the sentence supplies the "thereby". Try "Owing a stranger the truth does not give him the child's claim on the household."
6. **Colloquial or figurative words:** `11_ch11.md:140` "The **clean** illustration", `12_ch12.md:134` "Jain mendicants are **the same shape**", `11_ch11.md:140` "The same **shape** appears…", `12_ch12.md:140` "the living **listening to what they were given**".
7. **Book voice slipping into editorial language:**
   - `09_ch09.md:233`: "The consequence follows from the arrangement itself, **before any census of regimes**." This reads as a reply to the review note (`notes/10` call 12: "not as a census"). Write "The consequence follows from the arrangement itself."
   - `07_ch07.md:15` "The **authoritative statement** of this virtue…" and `09_ch09.md:284` "the **authoritative specification** of Virtue 2" use the notes' vocabulary. Write "The statement of this virtue…".
   - `14_ch14.md:8`, `:69`: "(The Precepts Matrix)" is a term from book1-4 and is never defined in `civ/`. Drop it or define it.
8. **Consistency of style:** `11_ch11.md:159` "35,000 feet" against "thirty-five thousand feet" elsewhere. `13_ch13.md:84` "Strip #4". `01_ch01.md:97` "requires **Virtue**" (capitalised mid-sentence). `ccp_derive.md:13` mixes "thirty-five thousand feet" with "eight hundred kilometers per hour".
9. **Virtue 7's short sentences** (`ccp_sheet.md:290`, `ccp_eng.md:675`, `10_ch10.md:70`) are the author's requested form (`notes/07` decision 17). No change recommended, apart from removing the repetition in item 4.

### 2.5 Suggested additions to the audit (guide §5)

> **Status: all approved and applied to `notes/06_editing_guide.md` (2026-10-27).**
> - Rule 5 now lists "Look back at", "Step back and", "Imagine", and "Picture", and Part A's opener pattern matches them, including at the start of a blockquote line.
> - Rule 8 now also prohibits a question answered by a one-word fragment, and stock prompts such as "What happened?". Part A has a matching "Question then fragment" check.
> - New Part C is a structural check for lists that will run into a paragraph, including inside blockquotes.
> - New Part D renders each source with pandoc in memory and counts run-in lists. It does not run `build.fish`, so it does not publish.
> - Section 5's introduction and section 6's workflow now refer to Parts A–D, and section 6 notes that the build publishes.
> - The scripts were extracted from the guide and run. Parts C and D find nothing. Part A's new check flags ten existing lines: `02_ch02.md:13`, `:94`; `06_ch06.md:79`; `07_ch07.md:38`; `09_ch09.md:322–324`; `11_ch11.md:93`; `14_ch14.md:119`; `ccp_derive.md:240`.

- Extend the Rule 5 pattern: `Look\s+back|Step\s+back|Imagine\s|Picture\s`.
- Add a Rule 8 analog: `\?\s+(No|Yes|Never|Nothing)\.` and `What (happened|occurred)[^?]*\?`.
- Add a structural check for a list item (`^\d+\.\s` or `^[*-]\s`) directly after a non-blank, non-heading, non-list line. Apply the same check inside blockquotes, after stripping the leading `> `. This would have caught §2.1 item 1, §1.3 item 1, and the three blockquote lists in ch05 and ch07.
- After each build, grep the HTML for `: 1\. ` inside `<p>` to catch run-in lists.

### 2.6 Found during the review but not recorded until now (2026-10-27)

> **Status: all resolved (2026-10-27).**
> - **Ch05:** trade flows "closely matched the relation between trade and distance…"; livestock "slaughtered or lost between 1929 and 1933".
> - **Ch09 Hammurabi:** twenty shekels of silver (laws 200–201), with the diyya schedules stated separately ("several classical schools…").
> - **Ch12 Rousseau:** "Robespierre and the Jacobins invoked Rousseau's General Will to justify the Terror of 1793–1794."
> - **Ch11 statistics:**
>   - Administrators "have grown faster than" frontline staff (no ratio).
>   - Reading: "more than a third of American fourth-grade pupils read below the basic level (National Assessment of Educational Progress, 2022)".
>   - Environmental reviews: "takes years…", with the CEQ 2020 finding of a 4.5-year average.
> - **Ch13 Singapore:**
>   - Home ownership: Lee's stated reason, a stake in the country tied to National Service (*From Third World to First*).
>   - Quotas "restricted" access.
>   - CPIB: Wee Toon Boon (convicted 1975) and Teh Cheang Wan (died under investigation 1986) named.
> - **Ch13 Stolypin:**
>   - The assassin is named: Bogrov, an anarchist and police informant.
>   - Lenin's 1908 remark is cited (*Collected Works*, vol. 15) in place of the attributed motive.
> - **Ch09 Kellogg-Briand:** "renouncing war as an instrument of national policy … and the Soviet Union adhered soon after". The adherence date in the sources conflicts (1928 or 1929), so no date is given.
> - **Eng:** "tax residency" removed. All twelve "Chapter N: Ordinal Principle/Virtue: Topic" headings now read "Chapter N: Topic (Principle N)" or "(Virtue N)". Part II and Part III lost their second colon.
> - **Book "Why" headings:** the five standalone ones were retitled:
>   - ch01:23 and :55
>   - ch03:3
>   - ch07:150
>   - ch09:130
>
>   The ten where "Why" follows a substantive subject were kept.
> - **Ch04 Rule 1:** "honest trade" became "truthful dealing".
> - **Cross-references:** ch07:9 now says "Chapters 3 and 4"; ch11's table cites "Section 1 Above".

**Factual or sourcing:**
- `05_ch05.md:65`: the trade-gravity result is "with high mathematical precision". Barjamovic and others found distance effects close to modern estimates; "closely matched" is safer.
- `05_ch05.md:84`: livestock "slaughtered in a single season". The losses ran across 1929–1933.
- `09_ch09.md:242`: Hammurabi's penalties are in weighed silver (laws 196–201), not "a few silver coins". "Non-believer" belongs to the *diyya* schedules, not to Hammurabi.
- `12_ch12.md:65`: "The Jacobin terror of 1793 was Rousseau's political doctrine applied in practice" states a contested reading as fact. Attribute it, for example "Robespierre invoked Rousseau…".
- `11_ch11.md:234–235`: three unsourced statistics: administrators outnumbering physicians or faculty "two or three to one", "forty percent of the students cannot read at grade level", and "twelve years to approve the environmental impact statement for a single munitions factory".
- `13_ch13.md:52`: the riot rationale for home ownership is not Lee's stated reason, which was a stake in the country. `:56`: Malaysia's quotas restricted rather than "barred". `:59`: the CPIB prosecutions of "cabinet ministers" need names. Wee Toon Boon was a Minister of State; Teh Cheang Wan died before charge.
- `13_ch13.md:88`: attributing the assassination motive to "Lenin and the socialist revolutionaries" overstates. Bogrov was an anarchist and police informant.
- `09_ch09.md:340`: the Soviet Union adhered to the Kellogg-Briand Pact in 1929; it did not sign in 1928.
- `ccp_eng.md:434`: "tax residency" is listed as an obligation of citizenship, but resident non-citizens also pay tax.

**Guide compliance:**
- Rule 9: `ccp_eng.md` chapter headings use the double-colon form ("Chapter 1: First Principle: The Nature of…"), twelve times, plus the Part II and Part III headings.
- Rule 9: fifteen book section headings use the "Why …" pattern (for example `01_ch01.md:23`, `:55`; `03_ch03.md:3`; `09_ch09.md:130`). This is a judgment call; some are fine.
- Rule 1, borderline: `04_ch04.md:48` "honest trade".

**Cross-references** (all "Chapter N" references checked; two are slightly off):
- `07_ch07.md:9`: "In Chapter 3, we saw how Principle 1 and Principle 2 immediately require Virtue 1 and Virtue 5". Virtue 5 is developed from Principle 2 in Chapter 4, so this should be "In Chapters 3 and 4".
- `11_ch11.md:225`: the table cites "Chapter 11" inside Chapter 11; "this chapter" would be clearer.

---

## Items the author already decided (listed for completeness, not as defects)

- Lysenko and the Great Leap paragraph stay as written (`notes/10` call 2; `notes/08` decision 1).
- "The chain is absolute" (`04_ch04.md:41`) and the absolute family sentence in ch05 stay (calls 6 and 7).
- The choir's 0.6 children is a supposition (call 9).
- `10_ch10.md:14` "A couple with 0.8 children has 0.4 grandchildren": call 13 kept it as "loose arithmetic". **Corrected (2026-10-27) at the author's request to "0.32 grandchildren".** For the record, 0.8 children form 0.4 couples, who have 0.32 children, so a couple has 0.32 grandchildren. The 100 → 40 → 16 line is correct. Deleting the 0.4 clause would remove the inconsistency without changing the illustration.
- "That refusal stands against these principles" after the refusal-of-violence sentence (`ccp_sheet.md:185`, `ccp_eng.md:409`) is the author's wording (`notes/07` decision 22).

---

## Suggested order of work

1. **§1.1:** settle decisions 21 and 24, then make all three essays agree.
2. **§2.1:** add the blank lines before lists and fix derive item 3 and the ch10 blockquote. Mechanical, and visible in every edition.
3. **§1.2 items 1, 2, 4, 8, 12 and §2.2 Medium items 1–5:** errors a reader will catch.
4. The remaining §1.2 items, then §2.3 and §2.4. Edit bottom-up per guide §6, re-run the audit, and rebuild.

---

## Open items (2026-10-27)

1. **The ten lines flagged by the Rule 8 "Question then fragment" check.** Quoted for the author's decision. **Decided:** reworded `02_ch02.md:13`, `06_ch06.md:79`, and `ccp_derive.md:240`. The other seven are kept as deliberate contrasts or series and recorded as permitted under Rule 8 in the guide.
2. **Part A "Math / symbols" false positives:** `11_ch11.md:39–41` and `12_ch12.md:180`. Listed for the author's decision. **Decided: option (b).** `<` and `>` are dropped from the pattern, and the check is now clean.
3. **`notes/10` call 18:** annotated with the correction. Done.
4. **Checks relied on rather than re-run** (the Levchuk and Duke table values via grok; Lee's memoir and Lenin's 1908 remark via secondary summaries): accepted by the author.
5. **TODO:**
   - Read today's edited passages in context.
   - Re-run Part B (n-grams). **Done (2026-10-27).** The corpus grew from 79,614 to 80,580 words.
     - The top lists are unchanged in character: structural labels and the principles' own wording.
     - Phrases that grew since `cb0a84c`:
       - "virtue derivation this virtue" (+7) is an artifact of eng's "(Virtue N)" headings running into "Derivation / This virtue…".
       - "more than a" (+4) comes from four separate factual fixes in ch08, ch10, ch11, and ch13.
       - "willing and able" and "Those who are" (+4 and +3) come from the deliberate parallel pair in derive and eng.
     - Phrases that fell: "test what they are told" (−3), "Greek and Armenian" (−3), and the repeated famine figure.
     - No new rhetorical tic.
   - Build and inspect the PDF and EPUB (the build publishes).
   - Commit.
6. **Declined:** the optional ch08 "broadly and in particular" sentence and the ch10 children-sequence paragraph.
