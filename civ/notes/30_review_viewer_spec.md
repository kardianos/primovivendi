# 30. Spec: A review viewer for lints, calls, and rewrites

Recorded 2026-11-20. Implementation specification for a local viewer over editorial items. Hand this file to an implementation agent. The product is the packet format. The window is a shell over that format.

## Recommendation

Build it. No existing tool covers the loop this manuscript already uses.

What exists, and what it covers:

- The calls notes (`civ/notes/10_chapter_calls.md`, `civ/notes/15_sheet_calls.md`, `civ/notes/16_followup_calls.md`) are the current queue. They already carry the real dispositions: accepted, accepted in a shorter form, the author's own sentence, withdrawn, open, research, explain the objection, leave the passage. They are a poor viewer. Line numbers go stale, options sit inside long prose, two rewrites cannot be set side by side, and the agent that wrote the file is the agent that has to parse the author's notes back out of it.
- [SARIF Explorer](https://github.com/trailofbits/vscode-sarif-explorer) (VS Code) triages findings as bug, false positive, or todo, with a comment, and stores that in a sidecar. It does not show several rewrites of one passage, and it does not hand the author's wording to a later local agent.
- [reviewdog](https://github.com/reviewdog/reviewdog) rdjson allows a list of suggestions on one diagnostic. Only the GitHub pull-request reporter and the GitLab merge-request reporter render them. The local reporter does not. A suggestion there is one hunk to apply on a pull request. This manuscript is not reviewed that way.
- Manuscript (getmanuscript.app) can diff a chapter snapshot and accept or reject parts of it. It is a novelist application with its own store. It does not read `civ/lint` or `civ/linear`, and it does not return a decision batch to local agents.
- Grammarly, LanguageTool, ProWritingAid, and Word track changes offer one suggestion stream and an accept or reject. They do not take an agent packet, and they do not set several section rewrites side by side.

An editor diagnostic (LSP code actions, the Problems panel) can list lints and offer a short fix. A paragraph-length alternate, a section rewrite, and an author sentence that replaces both do not fit in a code-action menu, and the decision is not a file a later agent can read.

## Guardrails

- The program lives in `civ/review/`, its own Go module, beside `civ/lint` and `civ/linear`.
- Packets and decisions live in `civ/notes/review/`, so the decision record sits with the other notes and is committed.
- The viewer window (`ui`) reads chapter and sheet files. It never writes them. It never runs `civ/build.fish`.
- A decision is data. Mechanical verbatim edits (`accept` and verbatim `revise`, uniquely located, non-overlapping, unnoted, unlocked) are applied deterministically by `go run . apply` (never from the GUI, never running `civ/build.fish`). Everything requiring judgment (`hold`, stale/reaffirmed items, overlapping spans, locked lines, noted items, and post-apply coherence checks) is handled by an agent reading `export`.
- Agents write packets via `go run . add` (timestamped filename `YYYYMMDD-HHMMSS-<slug>.json`). Agents read `decisions.json` and `export` (`batch.md` or `-format json`). Agents never write `decisions.json`.
- A new packet may add or refresh an item. It must not clear a disposition the author already saved.
- Do not import the historical calls notes in the first version. They remain the record they are.

## Format

JSON, not csvc. A packet is nested: one item has several alternates, and a quote is a passage that may contain newlines. csvc, the soliddb composite CSV (`#csvc`, `#set`, `#data`), fits flat tables. It can dump items, alternates, and decisions later if a soliddb database should hold them. It is not the format tools should emit, and it does not locate a quote.

Matching uses bytes. `byte_offset` and `byte_end` are optional. Zero is a real offset, so a missing field means unknown and a present `0` means the start of the file. `quote` is the exact bytes. `before` and `after` are the bytes immediately around that quote. They are used only to tell two hits apart.

On load:

1. If `byte_offset` is present and the file bytes from that offset through `byte_end` (or through the end of `quote`) equal `quote`, the hit is `offset`.
2. Otherwise the quote is searched. No hit is `missing`.
3. `before` and `after`, when set, keep only the hits they border. One hit left is `moved`. Several left, or none of the context fits while raw hits exist, is `ambiguous`. The viewer lists those offsets and does not pick one.
4. Line and column are derived from the byte offset for display. They are not stored, and they are not trusted.

Accept and revise stay disabled unless the hit is `offset` or `moved`, or the author sets `reaffirm`. Reject, withdraw, and hold stay available.

The id is the one the tool emits. A useful shape is `source`, `path`, the first 8 hex chars of the quote's SHA-256, and `rule`. The offset is not part of the id. A later packet with the same id refreshes `problem`, `alternates`, `quote`, `from`, and the byte fields in place. A saved decision stays. On a refresh, repeat `from`. A blank `from` in the later file clears it.

## Toolkit

Native, with [gowaku](https://github.com/bankai-tensei/gowaku) v0.2.0. It matches the rest of `civ/`: pure Go, no cgo, one `go build`. This machine has Go 1.27.4; gowaku requires 1.27.3. Keep that requirement in `civ/review/go.mod` only. Do not raise `civ/lint` or `civ/linear` (those modules still say `go 1.26.0`).

The prototype window uses `control.SplitPane`, `collection.List`, `basic.RichText`, `field.Input`, and `control.Button`. gowaku has no diff widget. A word diff is painted as spans of `basic.RichText`: a word is a run of letters, numbers, or apostrophes, a whitespace run is its own token, and every other rune, including an em dash, is its own token. Changed lines are diffed by those tokens. The markdown module and `editor.Editor` are not used, so the quote highlight and the word diff stay in one text widget.

`gowakutest` covers a tap of Accept. Packet merge, locate, and export are ordinary `go test` and need no window. The JSON files are the work product and must stay valid if the window layer is replaced. There is no web shell.

## What an item is

An item is one decision the author can make. Four kinds arrive in the same shape:

- `lint`. A span from `civ/lint` (em dash, spelling, locked line, list-in-paragraph). Zero or more short replacements.
- `sentence`. A sentence from `civ/linear`, with the mechanical draft already stored on `finding.Suggest` (`pull back`, `pull forward`, `split`, `move out`, `name it`).
- `call`. An agent raised a passage, stated the problem, and offered one or more alternates. This is a calls-note entry, in data.
- `rewrite`. Two or more drafts of the same section, to be read against each other and against the current text.

```json
{
  "version": 1,
  "id": "linear:civ/chapters/09_ch09.md:9f3a:split",
  "source": "civ/linear",
  "from": "grok-4.7",
  "kind": "sentence",
  "path": "civ/chapters/09_ch09.md",
  "byte_offset": 1840,
  "byte_end": 2012,
  "quote": "the exact current text of that range",
  "before": "\n\n",
  "after": "\n",
  "problem": "The verb sits behind a long opening.",
  "rule": "rule-10/split",
  "alternates": [
    {"id": "a", "label": "split", "text": "First sentence. Second sentence."},
    {"id": "b", "label": "pull forward", "from": "opus-5.5", "text": "The verb leads, and the opening follows."},
    {"id": "no", "from": "cabon", "ignore": true, "reason": "this was actually good"}
  ]
}
```

`path` is relative to the repository root and uses forward slashes. `quote` is the exact bytes, including a newline the tool included. `byte_offset` is the start of those bytes when the tool knew it, and `byte_end` is the byte after the quote. Omit both when the tool does not know an offset. `before` and `after` are optional context. The locate rules are in Format, above. A hit of `ambiguous` or `missing` can be read, rejected, withdrawn, or held. Accept and revise stay disabled until the quote matches once, or the author sets `reaffirm`.

`source` is the id prefix and the tool that emitted the item (`call`, `civ/linear`, `civ/lint`). `from` is the model or agent that wrote the suggestion: `argon`, `cabon`, `grok-4.7`, `opus-5.5`, and the same kind of name. It is not the book's author. The item `from` covers the problem and every alternate that agent wrote. An alternate has its own `from` slot. Leave that slot blank when the same agent wrote the alternate. A blank alternate `from` displays as the item `from`. Set the alternate `from` only when a different agent adds that alternate. `review add` clears an alternate `from` that equals the item `from`. A tool with no model name may omit `from`.

An alternate with `"ignore": true` is a vote against the whole suggestion, not a replacement. `reason` says why (`this was actually good`). Leave `text` empty. One such alternate is enough. The viewer labels it Ignore and shows the reason. Agreeing with it saves `reject`, copies the reason into the note when the note is empty, and does not change the passage. `apply` will not splice it.

A later run of the same tool refreshes the item in place when the id is unchanged. A changed quote is a new id. The old item stays until the author withdraws it or its quote matches again.

A `rewrite` uses the same object. `quote` is the current section. Each alternate is a full draft. `label` is the draft name (`current`, `draft A`, `shorter`). The compare view is the point of this kind.

A locked line from `civ/notes/00_civ_authoritative_bones.md` or Section 1 of `civ/notes/06_editing_guide.md` gets `"locked": true` when the importer knows. The viewer shows a banner naming those two files. It does not block the author.

## Packets and decisions

```
civ/notes/review/
  packets/          # one JSON file per tool run or agent pass; append-only
  decisions.json    # the viewer's file; agents read it and do not write it
  batch.md          # generated by `go run . export`; hand this to the next agent
```

Agents that add a question or read a decision are pointed at `notes/31_review_queue.md`. This file remains the implementation spec.

A packet file is either one item or `{"version":1,"items":[...]}`. A response packet may also carry `applied` and `answers`. The viewer loads every `packets/*.json` at startup. Files that fail to parse are listed in the status bar and skipped. Last file wins for a given `id` on the fields agents own (`problem`, `alternates`, `rule`, `quote`, `byte_offset`, `byte_end`, `before`, `after`, `from`). `decisions.json` wins on the disposition. Reloading when the window gains focus is not in the prototype yet.

`decisions.json`:

```json
{
  "version": 1,
  "items": {
    "linear:civ/chapters/09_ch09.md:9f3a:split": {
      "disposition": "revise",
      "alternate": "a",
      "text": "the author's sentence, copied as typed",
      "note": "shorter, and name the subject",
      "hold_for": "",
      "reaffirm": false,
      "at": "2026-11-20T15:04:00Z"
    }
  }
}
```

Write it by writing a temporary file in the same directory and renaming it over the destination.

Dispositions, in the words the calls notes already use:

| Disposition | Meaning | Chapter text |
| --- | --- | --- |
| `accept` | Use `alternates[alternate].text` as the replacement. | Later agent may apply it. |
| `revise` | Use `text`, the author's wording. `alternate` records which draft it started from, when it started from one. | Later agent may apply `text`. |
| `reject` | The proposal is wrong. `note` is optional. | Unchanged. Do not raise the same quote and rule again. |
| `hold` | The passage stays for now. `hold_for` is `research`, `explain`, or `later`. | Unchanged. The agent answers in a new packet. |
| `withdraw` | The item should not have been raised. | Unchanged. Do not raise it again against this quote. |

`hold` / `explain` is "explain the objection" in note 10. `hold` / `research` is "Research. Passage unchanged." `hold` / `later` is "suggestions are allowed, no change until the queue is done." `revise` is "accepted in reduced form" and the `DT:` sentences in note 15. `withdraw` is "Withdrawn."

Empty disposition means the item is still open. The viewer opens on the first open item. An empty open list selects nothing.

## The batch an agent answers

`go run . export` writes `civ/notes/review/batch.md` from `decisions.json` and the packets. The file starts with these rules, then three sections:

- **Apply.** Each `accept` and `revise`. Path, the quote, the replacement text verbatim, the note. The agent applies an item only where `quote` is still in the file exactly once. A miss is reported in the response packet and the file is left alone.
- **Answer.** Each `hold`. The agent's reply is a new packet: an `answers` entry, and any new items. The chapter stays as it is.
- **Closed.** Each `reject` and `withdraw`, with the quote and the note, so the next pass does not open the same item.

The response packet, written by the agent into `packets/`, names the ids it touched:

```json
{
  "version": 1,
  "responds_to": ["linear:civ/chapters/09_ch09.md:9f3a:split"],
  "applied": [{"id": "linear:civ/chapters/09_ch09.md:9f3a:split", "path": "civ/chapters/09_ch09.md", "ok": true}],
  "answers": [{"id": "call:civ/chapters/04_ch04.md:b21c:load", "text": "the research, in the agent's words"}],
  "items": []
}
```

The viewer marks an applied item only when `ok` is true and the file now contains the replacement. It does not apply the replacement itself. An answer shows under the item. New `items` enter the queue as open.

## Window

One window, three regions. The prototype opens on the first open item. An empty open list leaves the passage and the actions blank.

- **Left.** Filters: open, all, hold, decided, stale. Then the items. A row shows the id, the match (`offset`, `moved`, `ambiguous`, `missing`), the kind, the disposition, and the item `from` when it is set.
- **Center.** The path, the item `from` when it is set, the recorded offset against the offset that was found, and the problem. Under that, the source around the quote, with the quote marked, and a word diff of the quote against the focused alternate. The diff title names that alternate's `from` (its own, or the item `from` when the alternate's slot is blank). The passage starts at two lines above the quote and two below. Above, Below, `[`, and `]` open that side further (8 lines, then 24, then 80, then the rest of the file). Chapter shows the whole file. Tight returns to two lines. Moving to another item returns to two lines. Compare replaces the passage with columns: the current quote, each alternate, and Mine when it has text. Each alternate column shows its `from`.
- **Right.** One button per alternate, with that alternate's `from` under the button. An `ignore` alternate is the button Ignore. Then Accept, Use my wording, Reject, the three holds, Withdraw, Reaffirm, Compare, Next, and Previous. When Ignore is focused, Accept is labeled Ignore this and records `reject` with the reason. Under those, the full alternate, its `from`, the Mine field, and the note. An `ignore` alternate shows its reason instead of a replacement.

Keys, and the same actions as buttons. A focused Mine or Note field takes the plain letters, because those keys type. Ctrl+J and Ctrl+K still move, and Ctrl+Enter saves Mine.

- `j` / `k` next and previous item, and `n` / `p` the same, when neither field is focused.
- `[` / `]` open more of the chapter above and below the quote. The same actions are the Above and Below buttons. Chapter and Tight are buttons.
- `a` accepts the focused alternate. `r` rejects. `h` holds for later. `w` withdraws. `c` toggles compare.

Accept, Use my wording, Reject, the three holds, and Withdraw save the decision and show the next item in the current filter. When that filter has nothing left, the passage and the actions clear. A last item that remains in the filter stays selected. Reaffirm stays on the same item.
- The three hold buttons set `hold_for` to `research`, `explain`, or `later`. Reaffirm sets `reaffirm` and leaves the disposition as it is.

## Terminal

`go run . tui` is the same session in the terminal. It reads and writes the same packets and `decisions.json`. It does not write chapters. The shared rules (filter, selection, context window, drafts, decisions) live in `reviewView`. The window is a gowaku shell over that type. The terminal is a Charmbracelet shell over it.

Filters are `1` Open, `2` Hold, `3` All, `4` Decided, `5` Stale. `j` / `k` and `n` / `p` move. Tab moves the focused alternate. `a` accepts, or records Ignore this. `u` is Use my wording. `r` rejects. `w` withdraws. `h` holds for later, `e` for explain, `s` for research. `o` reopens. `f` reaffirms. `c` compares. `d` switches the diff between Mine and the alternate. `[` and `]` widen the passage, `C` shows the chapter, `t` tightens it. `q` copies the quote into Mine and `y` copies the focused alternate. Enter edits Mine, `N` edits Note, and Esc leaves the field. Ctrl+J and Ctrl+K move while a field is focused. Ctrl+S, while editing, is Use my wording. Ctrl+R reloads. `?` opens the key list. Ctrl+C quits.

The window binds Use my wording to Ctrl+Enter. Terminals usually do not send that chord, so the terminal screen uses `u` and Ctrl+S for the same action.

The detail pane is split in half. Each half has tabs for Current, Proposed, and Actions, and the two halves are chosen separately, including the same view on both. The top starts on Current and the bottom on Proposed. `,` cycles the top half. `.` cycles the bottom half. Current holds the passage and the context controls: the above/below label, Above, Below, Chapter, and Tight. Proposed lists the alternates, then a line, then Quote -> Mine, Alt -> Mine, the diff target, and the diff or the compare. When both halves show Proposed, each half keeps its own alternate and its own diff target, and the wheel scrolls only the half under the pointer. Actions holds the decision chips, a large Mine field, and Note. The alternate text is not repeated on Actions. Enter and `N` show Actions on the bottom when neither half is showing it. The other keys still run when Actions is hidden. Tab and `d` follow the Proposed half that was clicked last, or the only Proposed half.

A left click runs the chip, filter, tab, or item under the pointer. The wheel over the list scrolls the list. The wheel over a Current body scrolls the passage, and the wheel over a Proposed body scrolls the diff or the compare. Prose in the passage, the diff, and the problem wraps inside the pane. Each pane keeps a one-cell inset with a blank column between them, and chips keep a blank cell between them so a gap or a dim chip does not fire.

The author's text is stored as typed, including a rough sentence. The viewer does not correct it. Accept and Use my wording write nothing when the quote is ambiguous or missing, unless `reaffirm` is set. Reject, withdraw, and hold write anyway.

The status bar shows open, hold, decided, and stale counts, the last save, and the name of any packet that failed to parse.

Not in the prototype: a per-file pass, a command palette, a locked-line banner, and clearing a disposition back to open. A locked quote is still only a banner when it is added. The viewer does not block the author.

## Importers

Default output of `civ/lint` and `civ/linear` stays text. The first version parses that text from inside `civ/review`. A `-json` flag on those tools can wait until the packet shape has been used.

```bash
fish civ/tools/build-bin.fish                            # once (builds gitignored civ/bin/*)
civ/bin/review [ui]                                      # author: open the read-only review window
civ/bin/review tui                                       # author: same session in the terminal
civ/bin/review check [-json]                             # inspect: validate packets and locate every quote
civ/bin/review apply [-dry-run]                          # applier: apply clean verbatim accept/revise decisions
civ/bin/review export [-format md|json]                  # handoff: export remaining unapplied items, holds, and closed items
civ/bin/review add -slug NAME [-in FILE|-] [-strict]     # agent: validate, pin offsets, write timestamped packet
```

The subcommand comes before its flags. Common flags: `-root`, `-packets`, `-decisions`. `apply` also accepts `-dry-run`, `-include-noted`, `-allow-locked`, `-id`, `-json`. `export` also accepts `-format md|json`, `-out PATH|-`. `add` also accepts `-in`, `-slug`, `-pin`, `-strict`, `-json`.

The sample test fixture lives in `civ/review/testdata/packets/sample.json` (over `civ/review/testdata/passage.md`), keeping `civ/notes/review/packets/` clean for the live manuscript queue.

`import-lint` and `import-linear` are not in the prototype. Default output of those tools stays text. When they grow a JSON flag, it should write this packet. Until then an agent writes `packets/<name>.json` directly.

`import-lint`, once added, maps each printed hit (`path:line: text`) to a `lint` item. `quote` is the full line. `alternates` is empty unless the hit text contains a replacement, in which case that replacement is alternate `a`. The line number is not stored. The quote, and `before` / `after` when the tool can spare them, are.

`import-linear` runs `go run . lines -chapter chNN` in `civ/linear` and reads `formatFinding`. A finding with `Suggest` becomes a `sentence` item whose one alternate is labeled with `Move`. Findings with no draft still enter, as items with no alternate, so the author can reject, hold, or write Mine.

An agent that already knows the schema writes `packets/<name>.json` directly and skips the importers. That is the path for a call with several hand-written alternates, and for a pair of section rewrites.

## Tests

Plain `go test` in `civ/review`:

- Two packets with the same `id`: the later `problem` shows, and an existing `revise` disposition remains, including `text`.
- A recorded offset that still matches is `offset`, including offset 0. A quote that occurs once elsewhere is `moved`. `before` can pick one of two hits. Two hits with no context, or context that fits none of them, is `ambiguous`, and accept refuses it without `reaffirm`. No hit is `missing`.
- An em dash is its own word-diff token.
- `export` places an `accept` and a `revise` under Apply, a `hold` under Answer, and a `reject` under Closed. The replacement text is the alternate for `accept` and the author's `text` for `revise`. An open item is left out.
- A response packet with `ok: true` marks applied only when the file contains the replacement. `ok: true` while the replacement is absent stays unapplied. `ok: false` stays unapplied.
- `gowakutest` opens the sample packet, taps Accept on the offset hit, and reads `accept` back from a temporary `decisions.json`.
- A blank alternate `from` displays as the item `from`. An alternate `from` that equals the item `from` is cleared by `add`. A different alternate `from` is kept and shown on that alternate.
- An `ignore` alternate shows its reason, and agreeing with it writes `reject` without a replacement. `apply` skips an accept that names an `ignore` alternate.

`import-lint` is not covered yet, because that importer is not written. The sample packet is `civ/notes/review/packets/sample.json`, over `civ/review/testdata/passage.md`. It has an exact offset, a moved hit, an ambiguous hit, a missing quote, a two-draft rewrite, and an em-dash lint.

## Out of scope

- Writing chapters, the sheet, the bones file, or the editing guide.
- Running `civ/build.fish` or `push_release`.
- A chat pane. The batch file is how the next agent hears the decision.
- Importing notes 07, 08, 10, 15, or 16 into packets.
- Editing `decisions.json` by hand is possible and must keep working. It is not the interface.
