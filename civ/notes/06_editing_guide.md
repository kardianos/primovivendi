# Civilization (`civ`) Prose & Editing Guide

This guide consolidates the stylistic standards, lexical rules, punctuation constraints, and automated N-gram verification methods used across all documents in `civ/` (`civ/chapters/*.md`, `civ/ccp_argon.md`, `civ/ccp_derive_argon.md`, and `civ/ccp_eng_argon.md`).

---

## 1. Authoritative Formulations (Preserve Verbatim)

The following two lists are the foundational axioms and derived virtues of the work. Their exact wording must never be altered when editing surrounding prose:

### The 5 Common Principles of Reality
1. **Reality is real, knowable, and consequential. Truth is what corresponds to reality.**
2. **Humans are self-interested, limited, and particular.**
3. **Goodness must include human life continuing, and that life kept and built.**
4. **What you keep and build is self-anchored, tiered, and outward. Enemies are real.**
5. **Justice is scoped procedural symmetry.**

### The 7 Virtues of Civilization
1. **Align with truth**
2. **Take ownership, which is responsibility**
3. **Seek excellence**
4. **Admire beauty**
5. **Be humble**
6. **Have children**
7. **Form and listen**

---

## 2. Target Prose Register & Literal Meaning

1. **1860s / 1950s Engineering-Analysis Register:**
   - Write in clear, calm, declarative analytical prose modeled on classical nineteenth-century and mid-twentieth-century engineering and constitutional treatises.
   - Avoid modern journalistic slang, debate-stage rhetoric, conversational fillers, and artificial urgency.
2. **Literal Dictionary Definitions Only:**
   - Every word must be used in its strict, literal dictionary meaning.
   - Do not use "cute" or non-literal metaphors drawn from anatomy/digestion (`"chewed up"`, `"stomach for"`, `"blind spot"` when `"oversight"` is meant), animals (`"dog-eat-dog"`, `"canary in the coal mine"`), theater (`"performative"`, `"stagecraft"`), computing (`"hard-coded"`, `"operating system"`, `"bandwidth"`), or colloquial accounting/gambling (`"skin in the game"`, `"cook the books"`, `"write a blank check"`).
   - When an engineering or physical illustration is used (such as a suspension bridge, a steam boiler, or a ship's hull), present it as an explicit, literal physical case or analogy rather than mixing figurative slang into abstract sentences.

---

## 3. Typography, Punctuation, and Notation Rules

1. **No Em-Dashes (`—`):**
   - Never use the em-dash character (`—`) or double-hyphen dash (`--`) in prose.
   - Replace with a colon (`:`) when introducing an explanation or list, a semicolon (`;`) when joining independent clauses, commas or parentheses when setting off a parenthetical qualification, or a full stop (`.`).
2. **No Exclamation Marks (`!`):**
   - Never use exclamation marks in prose, headings, or dialogue. Calm declarative statements carry greater weight.
3. **No Mathematical Equations or Symbolic Shorthand in Titles or Prose:**
   - Do not use LaTeX math mode (`$c$`, `$\approx 2.1$`, `$29^\circ\text{F}$`, `$X$`) in headings, tables, or prose. Write all quantities and units out in English words (`the speed of light`, `approximately 2.1`, `twenty-nine degrees Fahrenheit`).
   - Do not use mathematical operators (`+`, `=`, `>`, `<`) as shorthand in headings, bullet titles, or prose (for example, write `Order and Truth Without Vitality Produce Sterility` rather than `Order + Truth without Vitality = Sterility`).
   - Do not use ampersands (`&`) in headings, bullet titles, or tables; always write `and`.
   - Do not use algebraic placeholder letters (`Procedure $X$`, `Doctrine $X$`, `Rich Citizen A`, `"Not Real X"`) in titles or prose; use concrete descriptive nouns (`the established legal procedure`, `authentic doctrine`, `a wealthy citizen`).

---

## 4. Lexical & Syntactic Rules (From Audit & N-Gram Analysis)

### Rule 1: Literal "Honest" Only
- **Standard:** Reserve `honest`, `honesty`, and `honestly` strictly for a human moral agent's deliberate truthfulness, candor, or refusal to lie, cheat, or steal (`an honest witness`, `personal honesty`, `earning a living honestly`).
- **Prohibited:**
  - *Inanimate objects or materials:* Never write `honest steel`, `honest iron`, `honest stone`, `honest bridge`, or `honest weight`. Use `sound steel`, `genuine stone`, `durable bridge`, `accurate weight`.
  - *Analytical processes or documents:* Never write `honest audit`, `honest accounting`, `honest ledger`, `honest test`, or `honest comparison`. Use `accurate audit`, `complete accounting`, `symmetrical comparison`, `verifiable ledger`.
  - *Conversational intensifiers:* Never write `honest mistake`, `honest error`, `honest question`, or `honest disagreement`. Use `unintentional error`, `good-faith question`, `sincere disagreement`.

### Rule 2: Literal "Load-Bearing" Only
- **Standard:** Reserve `load-bearing` strictly for literal structural engineering and architecture where a physical member supports physical weight (`load-bearing stone wall`, `load-bearing columns`).
- **Prohibited:** Never apply `load-bearing` figuratively to words, concepts, virtues, or social institutions (`load-bearing word`, `load-bearing premise`, `load-bearing wall of the republic`). Use `foundational`, `essential`, `indispensable`, `structural`, or `governing`.

### Rule 3: No Colloquial Pleonasms ("Exact Same", "Every Single")
- **Prohibited:** `"exact same"` and `"every single"`.
- **Replacement:** Replace `"the exact same"` with `"the identical"` or `"the same"`. Replace `"every single"` with `"every"` or `"each"`.

### Rule 4: Numerical "Zero" Only
- **Standard:** Reserve `"zero"` for literal numerical values, temperatures, or mathematical limits (`fifty degrees below zero Celsius`, `reduced to zero`).
- **Prohibited:** Never use `"zero"` as a rhetorical intensifier for `"no"` or `"none"` (`"zero tolerance"`, `"zero accountability"`, `"zero evidence"`, `"zero dollars"`, `"Year Zero"`).
- **Replacement:** Use `"no tolerance"`, `"no accountability"`, `"no evidence"`, `"nothing"`, `"from the beginning"`.

### Rule 5: No Imperative Reader-Directing Openers
- **Prohibited:** Do not begin sentences or clauses by pointing at the reader with imperative verbs:
  - `"Look at..."` / `"Look closely at..."`
  - `"Notice how..."` / `"Notice that..."` / `"Notice what..."`
  - `"Watch what happens..."`
  - `"Observe how..."` / `"Observe that..."`
  - `"Examine..."` / `"Consider..."`
  - `"Look back at..."` / `"Step back and..."`
  - `"Imagine..."` / `"Picture..."`
- **Replacement:** State the subject and the analytical observation directly in declarative form (for example, replace `"Look at what happens when a bridge corrodes"` with `"When a bridge corrodes..."`; replace `"Notice the double standard here"` with `"This argument applies a double standard"`).

### Rule 6: No Hyperbolic Scope Tags ("On Earth")
- **Prohibited:** Never use `"on earth"` or `"on Earth"` as a rhetorical intensifier (`"every court on earth"`, `"no power on earth"`, `"the wealthiest nation on earth"`).
- **Replacement:** Omit the phrase (`"every court"`, `"no power"`) or use `"in the world"` (`"the wealthiest nation in the world"`). Use `"on the Earth's surface"` only when contrasting terrestrial physics with celestial mechanics.

### Rule 7: Literal "Raw", "Real-World", and "Rot" Only
- **"Raw":** Reserve `"raw"` for unprocessed physical materials (`raw iron ore`, `raw timber`, `raw sewage`, `raw clay`). Replace figurative uses (`raw power`, `raw assertion`, `raw numbers`, `raw coercion`) with `"brute force"`, `"unvarnished assertion"`, `"unadjusted figures"`, `"naked coercion"`.
- **"Real-World":** Avoid the hyphenated compound `"real-world"` (`real-world consequences`, `real-world engineering`). Use `"physical"`, `"practical"`, `"actual"`, or `"historical"`.
- **"Rot":** Reserve `"rot"`, `"rotted"`, and `"rotting"` for literal biological decay of organic matter (`rotted timber`, `rotting grain`). Replace institutional, moral, or civic `"rot"` with `"decay"`, `"deterioration"`, `"degradation"`, or `"corruption"`.

### Rule 8: No Rhetorical Question-Fragment Transitions ("Why? Because...")
- **Prohibited:** Never use standalone `"Why? Because..."` transitions or inline `"Why does...? Because..."` question-and-fragment pairs.
- **Also prohibited:** A question answered by a one-word fragment (`"...? No."`, `"...? Never."`) or a stock prompt such as `"What happened?"` / `"What occurred...?"` followed by the answer.
- **Permitted (author's decision, 2026-10-27):** A question answered by a word may stay when it forms a deliberate contrast (two questions with opposite answers) or completes a series of parallel questions. The audit still flags these for review. Kept cases: `02_ch02.md` (practice versus explanation), `07_ch07.md` (fault versus responsibility), `09_ch09.md` (the international-law series), `11_ch11.md` (Step 5), and `14_ch14.md` (the stolen-land professor).
- **Replacement:** Combine the cause and effect into a single declarative sentence (`"This occurs because..."`, `"A bridge stands only because..."`) or answer a section's framing question with a complete independent sentence rather than a `"Because..."` fragment.

### Rule 9: Substantive Titles and Headings
- Avoid repetitive colon formulas in chapter and section titles (`Chapter N: Topic: Subtitle`) and avoid abstract conversational pairings (`Why X Needs Y`). Give every chapter and section a distinct, substantive title that states its specific subject matter.

---

## 5. Automated N-Gram & Pattern Verification Method

Before finalizing any document in `civ/`, run the following four-part verification suite from `/home/d/code/po`. Parts A and B check wording; Parts C and D check that the Markdown renders as intended.

### Part A: Deterministic Rule & Pattern Audit Script

```bash
python3 -c '
import glob, re

files = sorted(glob.glob("civ/chapters/*.md")) + [
    "civ/ccp_argon.md",
    "civ/ccp_derive_argon.md",
    "civ/ccp_eng_argon.md",
]

checks = [
    ("Em-dash or exclamation", re.compile(r"[—!]")),
    ("Math / symbols in text", re.compile(r"(\$.*?\$|\\\w+|\s[\+=≈×÷&]\s|\b[A-Z]\s*[\+=]\s*[A-Z]\b)")),
    ("exact same / every single", re.compile(r"\b(exact same|every single)\b", re.I)),
    ("zero", re.compile(r"\bzero\b", re.I)),
    ("Imperative openers", re.compile(r"(?:^|[.!?]\s+|\*\*\s+|>\s+)(Look\s+at|Look\s+closely|Look\s+back|Step\s+back|Notice\s+|Watch\s+|Observe\s+|Examine\s+|Consider\s+|Imagine\s+|Picture\s+)", re.I)),
    ("on earth", re.compile(r"\bon earth\b", re.I)),
    ("raw / real-world / rot", re.compile(r"\b(raw|real-world|rot|rotted|rotting|rots)\b", re.I)),
    ("honest", re.compile(r"\bhonest\w*\b", re.I)),
    ("load-bearing", re.compile(r"\bload[- ]bearing\b", re.I)),
    ("Why... Because", re.compile(r"Why[^.!?]*\?\s*\*?\*?Because\b", re.I)),
    ("Question then fragment", re.compile(r"\?\s+\*{0,2}(No|Yes|Never|Nothing)\.|What (happened|occurred)[^?]*\?")),
]

for name, pat in checks:
    print(f"=== {name} ===")
    for p in files:
        with open(p) as f:
            for idx, line in enumerate(f, 1):
                if pat.search(line):
                    print(f"  {p}:{idx}: {line.strip()[:140]}")
'
```

### Part B: N-Gram & Sentence-Opener Frequency Scanner

Run the N-gram scanner (1-grams to 5-grams and 3-word sentence openers) across the corpus to detect any newly introduced repetitive stock phrases or rhetorical tics:

```bash
python3 -c '
import glob, re, collections

files = sorted(glob.glob("civ/chapters/*.md")) + [
    "civ/ccp_argon.md",
    "civ/ccp_derive_argon.md",
    "civ/ccp_eng_argon.md",
]

text = "\n".join(open(p).read() for p in files)
clean = re.sub(r"[#*>`_\[\]\(\)]", " ", text)
words = re.findall(r"\b[a-zA-Z\x27-]+\b", clean.lower())

for n in (3, 4, 5):
    counts = collections.Counter(tuple(words[i:i+n]) for i in range(len(words)-n+1))
    print(f"=== TOP {n}-GRAMS ===")
    for ng, c in counts.most_common(25):
        if c >= 8:
            phrase = " ".join(ng)
            print(f"  {c:4d}: {phrase}")

sents = re.split(r"(?<=[.!?])\s+|\n+", clean)
openers = collections.Counter()
for s in sents:
    w = re.findall(r"\b[a-zA-Z\x27-]+\b", s.strip())
    if len(w) >= 3:
        openers[" ".join(w[:3])] += 1

print("=== TOP 3-WORD SENTENCE OPENERS ===")
for op, c in openers.most_common(25):
    if c >= 6:
        print(f"  {c:4d}: {op}")
'
```

### Part C: Structural Markdown Check (Lists That Will Run Into a Paragraph)

The build uses pandoc's default Markdown reader. A list that follows a paragraph line with no blank line between them is printed as run-in text inside that paragraph, in the PDF, HTML, and EPUB alike. The same holds inside a blockquote, where the separating line must be a bare `>`. This check finds every such list:

```bash
python3 -c '
import glob, re

files = sorted(glob.glob("civ/chapters/*.md")) + [
    "civ/ccp_argon.md",
    "civ/ccp_derive_argon.md",
    "civ/ccp_eng_argon.md",
]

item = re.compile(r"^(\d+\.|[*-])\s")
print("=== Lists that will run into the preceding paragraph ===")
for p in files:
    lines = open(p).read().split("\n")
    for i in range(1, len(lines)):
        cur, prev = lines[i], lines[i - 1]
        if cur.startswith("> ") and prev.startswith("> "):
            cur, prev = cur[2:], prev[2:]
        elif prev.startswith((" ", "#", ">")):
            continue
        if item.match(cur) and prev.strip() and not item.match(prev.lstrip()) and not prev.startswith(" "):
            print(f"  {p}:{i + 1}: {cur.strip()[:100]}")
'
```

Fix each hit by inserting a blank line (or a bare `>` line inside a blockquote) before the list.

### Part D: Rendered-Output Check

Render each source with pandoc and look for a list that ended up inside a paragraph. The check renders into memory only, so it does not run `build.fish` and does not publish:

```bash
for f in civ/chapters/*.md civ/ccp_argon.md civ/ccp_derive_argon.md civ/ccp_eng_argon.md; do
  n=$(pandoc "$f" -t html | tr '\n' ' ' | grep -o -E ':\s+(1\.|\*) ' | wc -l)
  [ "$n" -gt 0 ] && echo "$f: $n run-in list(s)"
done; echo "Part D done"
```

After a build, the same `grep` can be run on `civ/build/*.html`.

---

## 6. Bottom-Up Editing Workflow

1. Run the audit scripts (Parts A and C) to obtain exact file paths and line numbers for all violations.
2. Read the surrounding context of each flagged line.
3. **Always edit files from the bottom up** (from the highest line number to the lowest line number within each file) so that earlier line numbers do not shift during multi-region edits.
4. Re-run Parts A and C to verify that no new violations remain, and run Part D to confirm the rendering. Then execute `fish civ/build.fish` to rebuild all targets in `civ/build/`. The script also runs `../push_release`, so building publishes.
