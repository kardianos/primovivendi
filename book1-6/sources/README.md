# book1-6 / sources — standalone writer source pack

**Purpose:** Give an isolated, careful writer enough **logic, primary excerpts, and adversary notes** to draft the book without the rest of the monorepo — and without inventing quotes.

**Does not replace:** Full libraries for publication-grade citation of long works (Mao *Selected Works*, full Ghazālī/Ibn Rushd, complete Federalist HTML cleanup, etc.). Where a full text is missing, this pack says so and points to what to fetch.

---

## How to use (read order)

1. `../plan/dictionary_p1.md` + `../plan/outline_p1.md` — Part 1 architecture  
2. `../plan/dictionary_p2.md` + `../plan/outline_p2.md` — Part 2 architecture  
3. `logic/logical_sequences_full_book.md` — numbered chains tying the whole argument  
4. Open the **source folder** matching the section you write (below)  
5. `notes/` for red–green and purity working notes (copied from book1-5)

---

## Map: section → sources

| Book section | Start here |
|--------------|------------|
| Axioms, life good | `logic/logical_sequences_full_book.md` SEQ-1–5; `philosophy/life_axiom_long_argument.md` |
| Testable truth / models | `philosophy/models_astronomy_chain.md`; SEQ-2 |
| Anti-marble / Rawls misuse | `philosophy/anti_marble_self_and_rawls.md` |
| Justice triad | `philosophy/justice_three_kinds.md` |
| Bible / two-books | `scripture/bible_excerpts.md` |
| Qur’an / packages | `scripture/quran_excerpts.md`; `philosophy/islam_secondary_causes_brief.md` |
| Form vs substance | `philosophy/discovered_vs_constructed.md` |
| Misused compassion | `philosophy/compassion_misuse_and_false_foundations.md` |
| Hermetic argument | `philosophy/hermetic_argument.md` |
| Making standards (G mechanisms, not art proof) | `philosophy/beauty_standards_mechanisms.md` |
| Singapore | **`singapore/DISTILL.md`** (full LKY txts only if expanding) |
| Russia | **`russia/DISTILL.md`** (full Lenin/Durnovo/Cheremukhin if expanding) |
| Iran | **`iran/DISTILL.md`** (full Khomeini + gaps if expanding) |
| China / Mao | `china/mao_and_great_leap_sources.md` |
| USA founding | `usa/federalist_*_excerpts.txt` |
| Dugin | `dugin/` (tweet primary + interview + excerpts) |
| Red–green alliances | `notes/red_green_alliance.md` |
| Purity vs reality procedure | `notes/purity_virtue_reality.md` |

---

## Directory layout

```text
sources/
  README.md                          (this file)
  logic/
    logical_sequences_full_book.md
  philosophy/
    models_astronomy_chain.md
    life_axiom_long_argument.md
    anti_marble_self_and_rawls.md
    justice_three_kinds.md
    discovered_vs_constructed.md
    compassion_misuse_and_false_foundations.md
    islam_secondary_causes_brief.md
  scripture/
    bible_excerpts.md
    quran_excerpts.md
  notes/
    red_green_alliance.md            (copy from book1-5)
    purity_virtue_reality.md         (copy from book1-5)
  singapore/DISTILL.md               (writer-facing distill; full LKY txts alongside)
  russia/DISTILL.md                  (writer-facing distill; full primaries alongside)
  iran/DISTILL.md                    (writer-facing distill; full Khomeini + gaps alongside)
  china/                             (Mao / Great Leap pack)
  usa/                               (Federalist excerpts)
  dugin/                             (tweet, interview, excerpts)
```

---

## Verification standards

| Material | Standard in this pack |
|----------|------------------------|
| Bible KJV | Public domain; check a printed KJV if publishing |
| Qur’an English | Sahih International-*style* wording; **name translation in print** and verify |
| LKY / Lenin / Durnovo / Stolypin / Khomeini | From book1-4/ref transcripts; OCR/HTML quality varies — spot-check |
| Federalist | Extracted from LOC HTML scrape; verify against LOC or a standard print edition |
| Dugin tweet | Full text + URL in file |
| Death tolls (China famine) | **Do not invent precision**; cite one scholarly secondary |
| al-Ghazālī / Ibn Rushd | Pointers only — fetch editions for block quotes |

---

## What an isolated writer can do with only this folder + dictionaries

**Can draft:** full logical skeleton; virtues; models section; life axiom; justice; compassion misuse; form/substance; package fairness paragraphs; alliance mechanisms; practice matrix design.

**Must still fetch for polished historical chapters:**  
cleaner Federalist text if quoting long; full Mao volume; one famine monograph; optional full Dugin books (excerpts only here); Arabic/English Qur’an edition you will cite by name.

---

## Changelog

| Date | Note |
|------|------|
| 2026-07-17 | Initial source pack: logic, philosophy, scripture, notes copies, history primaries from book1-4/ref |
| 2026-07-17 | Renamed folder `ref/` → `sources/` (root .gitignore ignores `ref/`) |
