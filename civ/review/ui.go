package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bankai-tensei/gowaku"
	"github.com/bankai-tensei/gowaku/geom"
	"github.com/bankai-tensei/gowaku/input"
	"github.com/bankai-tensei/gowaku/layout"
	"github.com/bankai-tensei/gowaku/text"
	"github.com/bankai-tensei/gowaku/theme"
	"github.com/bankai-tensei/gowaku/widgets/basic"
	"github.com/bankai-tensei/gowaku/widgets/collection"
	"github.com/bankai-tensei/gowaku/widgets/control"
	"github.com/bankai-tensei/gowaku/widgets/field"
	"github.com/bankai-tensei/gowaku/widgets/interact"
)

const (
	filterOpen = iota
	filterAll
	filterHold
	filterDecided
	filterStale
)

const (
	textBody  = 14
	textMeta  = 12
	textTitle = 18
)

type itemDraft struct {
	mine string
	note string
}

// reviewApp is the three-pane viewer. Session is shared with the caller.
type reviewApp struct {
	gowaku.StatefulBase
	Session *Session
}

type reviewState struct {
	gowaku.StateBase
	session     *Session
	filter      int
	selected    string
	altIndex    int
	mine        string
	note        string
	drafts      map[string]itemDraft
	diffAltOnly bool
	compare     bool
	mineFocused bool
	noteFocused bool
	status      string
	// beforeLines and afterLines are lines of the chapter shown past the
	// quote. Negative means that side of the file.
	beforeLines int
	afterLines  int
}

func (a reviewApp) CreateState() gowaku.State {
	s := &reviewState{
		session:  a.Session,
		filter:   filterOpen,
		altIndex: -1,
		drafts:   map[string]itemDraft{},
	}
	if s.session == nil {
		s.session = &Session{
			Files:     map[string][]byte{},
			Hits:      map[string]Hit{},
			Answers:   map[string][]Answer{},
			Applied:   map[string]bool{},
			Decisions: DecisionFile{Version: 1, Items: map[string]Decision{}},
		}
	}
	s.resetContext()
	s.selected = s.firstID()
	s.loadDraft()
	return s
}

func (s *reviewState) Build(ctx gowaku.BuildContext) gowaku.Widget {
	th := basic.ThemeOf(ctx)
	body := text.Style{Size: textBody, Color: th.Color.Ink}
	muted := text.Style{Size: textMeta, Color: th.Color.Ink2}
	return interact.Actions{
		Actions: s.actions(),
		Child: basic.Column{Cross: layout.CrossStretch, Children: []gowaku.Widget{
			basic.Expanded{Child: control.SplitPane{
				Pixels: true, Position: 320, MinFirst: 240, MinSecond: 620,
				First: s.listPane(th, body, muted),
				Second: control.SplitPane{
					Pixels: true, Position: 600, MinFirst: 300, MinSecond: 320,
					First:  s.centerPane(th, body, muted),
					Second: s.editPane(th, body, muted),
				},
			}},
			s.statusBar(th, muted),
		}},
	}
}

func (s *reviewState) actions() []interact.Action {
	var out []interact.Action
	add := func(key input.KeyCode, mod input.Modifiers, fn func()) {
		if fn == nil {
			return
		}
		out = append(out, interact.Action{
			Shortcut: interact.Shortcut{Key: key, Modifiers: mod},
			OnInvoke: fn,
		})
	}
	if !s.typing() {
		add(input.KeyJ, 0, s.next)
		add(input.KeyK, 0, s.prev)
		add(input.KeyN, 0, s.next)
		add(input.KeyP, 0, s.prev)
		add(input.KeyA, 0, s.acceptFn())
		add(input.KeyR, 0, s.reject)
		add(input.KeyH, 0, func() { s.hold("later") })
		add(input.KeyW, 0, s.withdraw)
		add(input.KeyC, 0, s.toggleCompare)
		atStart, atEnd, unique := s.contextEdges()
		if unique {
			add(input.KeyBracketLeft, 0, s.moreAbove(atStart))
			add(input.KeyBracketRight, 0, s.moreBelow(atEnd))
		}
	}
	add(input.KeyJ, input.ModControl, s.next)
	add(input.KeyK, input.ModControl, s.prev)
	add(input.KeyR, input.ModControl, s.reload)
	add(input.KeyEnter, input.ModControl, s.reviseFn())
	return out
}

func (s *reviewState) typing() bool { return s.mineFocused || s.noteFocused }

func (s *reviewState) listPane(th *theme.Theme, body, muted text.Style) gowaku.Widget {
	rows := s.visible()
	var list gowaku.Widget = basic.Center{Child: basic.Text{Content: "No items", Style: muted}}
	if len(rows) > 0 {
		list = collection.List{
			ItemCount:           len(rows),
			MeasureItems:        true,
			EstimatedItemExtent: 64,
			ItemBuilder: func(i int) gowaku.Widget {
				if i < 0 || i >= len(rows) {
					return basic.SizedBox{}
				}
				return s.itemRow(th, rows[i], muted)
			},
		}
	}
	return basic.Column{Cross: layout.CrossStretch, Spacing: 8, Children: []gowaku.Widget{
		basic.Padding{Insets: geom.Insets{Left: 8, Right: 8, Top: 8}, Child: basic.Column{
			Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 4,
			Children: []gowaku.Widget{
				basic.Row{Spacing: 8, Cross: layout.CrossCenter, Children: []gowaku.Widget{
					basic.Expanded{Child: basic.Text{Content: "Review", Style: text.Style{Size: textTitle, Weight: 700, Color: th.Color.Ink}}},
					s.button("Reload", control.ButtonQuiet, s.reload),
				}},
				basic.Row{Spacing: 4, Children: []gowaku.Widget{
					s.filterButton("Open", filterOpen),
					s.filterButton("Hold", filterHold),
					s.filterButton("All", filterAll),
				}},
				basic.Row{Spacing: 4, Children: []gowaku.Widget{
					s.filterButton("Decided", filterDecided),
					s.filterButton("Stale", filterStale),
				}},
			},
		}},
		basic.Expanded{Child: list},
	}}
}

func (s *reviewState) filterButton(label string, f int) gowaku.Widget {
	style := control.ButtonQuiet
	if s.filter == f {
		style = control.ButtonPrimary
	}
	return control.Button{
		Label: label, Style: style, Size: control.ButtonSmall,
		OnPress: func() {
			s.SetState(func() {
				s.stashDraft()
				s.filter = f
				if !s.selectedVisible() {
					s.selected = s.firstID()
					s.compare = false
					s.diffAltOnly = false
					s.resetContext()
					s.loadDraft()
				}
			})
		},
	}
}

func (s *reviewState) itemRow(th *theme.Theme, it Item, muted text.Style) gowaku.Widget {
	h := s.session.Hits[it.ID]
	d := s.session.Decisions.Items[it.ID]
	disp := d.Disposition
	if disp == "" {
		disp = "open"
	} else if d.HoldFor != "" {
		disp += "/" + d.HoldFor
	}
	badges := []string{string(h.Status)}
	if it.Kind != "" {
		badges = append(badges, it.Kind)
	}
	badges = append(badges, disp)
	if s.session.Applied[it.ID] {
		badges = append(badges, "applied")
	}
	if s.session.HasAnswer(it.ID) {
		badges = append(badges, "answered")
	}
	if it.Locked {
		badges = append(badges, "locked")
	}
	if it.From != "" {
		badges = append(badges, "from "+it.From)
	}
	for _, a := range it.Alternates {
		if a.Ignore {
			badges = append(badges, "ignore")
			break
		}
	}
	title := rowTitle(it, h)
	preview := singleLinePreview(it.Quote, 46)
	meta := strings.Join(badges, " · ")
	return interact.Pressable{
		Label:   it.ID,
		OnPress: func() { s.choose(it.ID) },
		Build: func(st interact.PressState) gowaku.Widget {
			bg := th.Color.Sidebar
			if it.ID == s.selected {
				bg = th.Color.Accent.WithAlpha(0.22)
			} else if st.Hovered {
				bg = th.Color.AccentTint
			}
			return basic.Container{
				Color:   bg,
				Padding: geom.Insets{Left: 8, Right: 8, Top: 6, Bottom: 6},
				Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 2, Children: []gowaku.Widget{
					basic.Text{Content: title, Style: text.Style{Size: 13, Weight: 700, Color: th.Color.Ink}},
					basic.Text{Content: preview, Style: text.Style{Size: 12, Color: th.Color.Ink}},
					basic.Text{Content: meta, Style: muted},
				}},
			}
		},
	}
}

func rowTitle(it Item, h Hit) string {
	base := filepath.Base(it.Path)
	if base == "" || base == "." {
		return it.ID
	}
	if h.Status == StatusOffset || h.Status == StatusMoved {
		return fmt.Sprintf("%s:%d", base, h.Line)
	}
	return base
}

func singleLinePreview(s string, maxRunes int) string {
	s = strings.Join(strings.Fields(s), " ")
	rs := []rune(s)
	if len(rs) > maxRunes {
		return string(rs[:maxRunes-1]) + "…"
	}
	return s
}

func (s *reviewState) centerPane(th *theme.Theme, body, muted text.Style) gowaku.Widget {
	it, ok := s.current()
	if !ok {
		return basic.Center{Child: basic.Text{Content: "No item selected", Style: muted}}
	}
	h := s.session.Hits[it.ID]
	children := []gowaku.Widget{
		basic.Text{Content: fmt.Sprintf("%s  (%s)", it.Path, it.ID), Style: text.Style{Size: 13, Weight: 700, Color: th.Color.Ink}},
	}
	if it.From != "" {
		children = append(children, basic.Text{Content: "from: " + it.From, Style: text.Style{Size: 13, Weight: 700, Color: th.Color.Ink}})
	}
	children = append(children, basic.Text{Content: s.matchLabel(it, h), Style: muted})
	if it.Locked {
		children = append(children, basic.Container{
			Color:   th.Color.Danger.WithAlpha(0.16),
			Padding: geom.Insets{Left: 8, Right: 8, Top: 6, Bottom: 6},
			Child: basic.Text{
				Content: "LOCKED LINE: Preserve verbatim per civ/notes/00_civ_authoritative_bones.md and Section 1 of civ/notes/06_editing_guide.md.",
				Style:   text.Style{Size: 12, Weight: 700, Color: th.Color.Ink},
			},
		})
	}
	if it.Problem != "" {
		children = append(children, basic.Text{Content: it.Problem, Style: body})
	}
	if answers := s.session.Answers[it.ID]; len(answers) > 0 {
		var ansWidgets []gowaku.Widget
		for i, ans := range answers {
			prefix := "Answer: "
			if len(answers) > 1 {
				prefix = fmt.Sprintf("Answer %d: ", i+1)
			}
			ansWidgets = append(ansWidgets, basic.Text{Content: prefix + ans.Text, Style: body})
		}
		children = append(children, basic.Container{
			Color:   th.Color.AccentTint,
			Padding: geom.Insets{Left: 8, Right: 8, Top: 6, Bottom: 6},
			Child:   basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 4, Children: ansWidgets},
		})
	}
	if bar := s.contextBar(muted); bar != nil {
		children = append(children, bar)
	}
	var main gowaku.Widget
	if s.compare {
		main = s.compareView(th, it, body)
	} else {
		main = s.passageView(th, it, h, body, muted)
	}
	children = append(children, basic.Expanded{Child: main})
	return basic.Padding{
		Insets: geom.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8},
		Child:  basic.Column{Cross: layout.CrossStretch, Spacing: 8, Children: children},
	}
}

func (s *reviewState) matchLabel(it Item, h Hit) string {
	d := s.session.Decisions.Items[it.ID]
	disp := d.Disposition
	if disp == "" {
		disp = "open"
	}
	if d.HoldFor != "" {
		disp += "/" + d.HoldFor
	}
	if s.session.Applied[it.ID] {
		disp += " · applied"
	}
	if s.session.HasAnswer(it.ID) {
		disp += " · answered"
	}
	if d.Reaffirm {
		disp += " · reaffirmed"
	}
	if AlternateDrifted(it, d) {
		disp += " · alternate changed"
	}
	rec := "no recorded offset"
	if it.ByteOffset != nil {
		rec = fmt.Sprintf("recorded %d", *it.ByteOffset)
	}
	switch h.Status {
	case StatusOffset, StatusMoved:
		return fmt.Sprintf("%s · %s · %s · found %d · line %d col %d", it.Kind, disp, rec, h.Offset, h.Line, h.Col)
	case StatusAmbiguous:
		return fmt.Sprintf("%s · %s · %s · ambiguous %s", it.Kind, disp, rec, joinInts(h.Candidates))
	default:
		return fmt.Sprintf("%s · %s · %s · missing", it.Kind, disp, rec)
	}
}

func (s *reviewState) passageView(th *theme.Theme, it Item, h Hit, body, muted text.Style) gowaku.Widget {
	ink := text.Style{Size: body.Size, Color: th.Color.Ink}
	children := []gowaku.Widget{basic.Text{Content: "Passage", Style: muted}}
	if pre, match, post, ok := snippet(s.session.bytes(it.Path), h, s.beforeLines, s.afterLines); ok {
		children = append(children, basic.RichText{
			Style: body,
			Spans: []text.Span{
				{Text: pre, Style: ink},
				{Text: match, Style: ink, Background: th.Color.AccentTint},
				{Text: post, Style: ink},
			},
		})
	} else {
		children = append(children,
			basic.Text{Content: "The quote is not a single match.", Style: muted},
			basic.Text{Content: it.Quote, Style: body},
		)
	}

	alt, hasAlt := s.alt()
	hasMine := strings.TrimSpace(s.mine) != ""
	useMine := hasMine && (!s.diffAltOnly || !hasAlt)

	diffTitle := "Word diff"
	var target string
	hasTarget := false
	if useMine {
		diffTitle = "Word diff (Quote -> Mine)"
		target = s.mine
		hasTarget = true
	} else if hasAlt && alt.Ignore {
		diffTitle = "Ignore"
		if from := SuggestionFrom(it, alt); from != "" {
			diffTitle += " · " + from
		}
	} else if hasAlt {
		lbl := alt.Label
		if lbl == "" {
			lbl = alt.ID
		}
		if from := SuggestionFrom(it, alt); from != "" {
			lbl += " · " + from
		}
		diffTitle = fmt.Sprintf("Word diff (Quote -> %s)", lbl)
		target = alt.Text
		hasTarget = true
	}
	diffHeaderChildren := []gowaku.Widget{basic.Expanded{Child: basic.Text{Content: diffTitle, Style: muted}}}
	if hasMine && hasAlt {
		toggleLabel := "Diff Alt"
		if s.diffAltOnly {
			toggleLabel = "Diff Mine"
		}
		diffHeaderChildren = append(diffHeaderChildren, s.button(toggleLabel, control.ButtonQuiet, func() {
			s.SetState(func() { s.diffAltOnly = !s.diffAltOnly })
		}))
	}
	children = append(children, basic.Row{Spacing: 8, Cross: layout.CrossCenter, Children: diffHeaderChildren})
	if hasAlt && alt.Ignore && !useMine {
		reason := alt.Reason
		if reason == "" {
			reason = "Ignore."
		}
		children = append(children, basic.Text{Content: reason, Style: body})
	} else if hasTarget {
		spans := diffSpans(WordDiff(it.Quote, target), body.Size, th.Color.Ink, th.Color.Danger.WithAlpha(0.28), th.Color.Accent.WithAlpha(0.28))
		if len(spans) == 0 {
			children = append(children, basic.Text{Content: "(empty)", Style: muted})
		} else {
			children = append(children, basic.RichText{Style: body, Spans: spans})
		}
	} else {
		children = append(children, basic.Text{Content: "No alternate selected.", Style: muted})
	}
	return basic.Scroll{Child: basic.Column{
		Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 8, Children: children,
	}}
}

func (s *reviewState) compareView(th *theme.Theme, it Item, body text.Style) gowaku.Widget {
	cols := []gowaku.Widget{s.draftColumn(th, "Current", "", it.Quote, body)}
	for _, a := range it.Alternates {
		label := altLabel(a)
		content := a.Text
		if a.Ignore {
			label = "Ignore"
			content = a.Reason
			if content == "" {
				content = "Ignore."
			}
		}
		cols = append(cols, s.draftColumn(th, label, SuggestionFrom(it, a), content, body))
	}
	if strings.TrimSpace(s.mine) != "" {
		cols = append(cols, s.draftColumn(th, "Mine", "", s.mine, body))
	}
	return basic.Scroll{
		Horizontal: true,
		Child: basic.Row{
			Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 12, Children: cols,
		},
	}
}

func altLabel(a Alternate) string {
	if a.Ignore {
		return "Ignore"
	}
	if a.Label != "" {
		return a.Label
	}
	return a.ID
}

func (s *reviewState) draftColumn(th *theme.Theme, title, from, content string, body text.Style) gowaku.Widget {
	head := []gowaku.Widget{
		basic.Text{Content: title, Style: text.Style{Size: 13, Weight: 700, Color: th.Color.Ink}},
	}
	if from != "" {
		head = append(head, basic.Text{Content: "from: " + from, Style: text.Style{Size: 12, Color: th.Color.Ink2}})
	}
	return basic.SizedBox{Width: 340, Child: basic.Column{Cross: layout.CrossStretch, Children: []gowaku.Widget{
		basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 2, Children: head},
		basic.Expanded{Child: basic.Scroll{Child: basic.Column{
			Size:     layout.MainMin,
			Cross:    layout.CrossStretch,
			Children: []gowaku.Widget{basic.Text{Content: content, Style: body}},
		}}},
	}}}
}

func (s *reviewState) editPane(th *theme.Theme, body, muted text.Style) gowaku.Widget {
	it, ok := s.current()
	if !ok {
		return basic.Center{Child: basic.Text{Content: "No item selected", Style: muted}}
	}
	var altRows []gowaku.Widget
	if len(it.Alternates) == 0 {
		altRows = append(altRows, basic.Text{Content: "No alternates", Style: muted})
	} else {
		const perRow = 3
		for start := 0; start < len(it.Alternates); start += perRow {
			end := min(start+perRow, len(it.Alternates))
			var row []gowaku.Widget
			for i := start; i < end; i++ {
				i := i
				a := it.Alternates[i]
				label := altLabel(a)
				style := control.ButtonQuiet
				if a.Ignore {
					style = control.ButtonDanger
				}
				if i == s.altIndex {
					style = control.ButtonPrimary
				}
				btn := s.button(label, style, func() {
					s.SetState(func() { s.altIndex = i })
				})
				cell := btn
				if from := SuggestionFrom(it, a); from != "" {
					cell = basic.Column{
						Size: layout.MainMin, Cross: layout.CrossCenter, Spacing: 1,
						Children: []gowaku.Widget{
							btn,
							basic.Text{Content: from, Style: muted},
						},
					}
				}
				row = append(row, cell)
			}
			altRows = append(altRows, basic.Row{Spacing: 4, Children: row})
		}
	}
	alt, hasAlt := s.alt()
	var altText string
	if hasAlt {
		altText = alt.Text
	}
	var copyAltFn func()
	if hasAlt && !alt.Ignore {
		copyAltFn = func() {
			s.SetState(func() {
				s.mine = alt.Text
				s.diffAltOnly = false
			})
		}
	}
	acceptLabel := "Accept"
	acceptStyle := control.ButtonPrimary
	if hasAlt && alt.Ignore {
		acceptLabel = "Ignore this"
		acceptStyle = control.ButtonDanger
	}
	topChildren := []gowaku.Widget{
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button("Previous", control.ButtonQuiet, s.prev),
			s.button("Next", control.ButtonQuiet, s.next),
			s.button("Compare", control.ButtonQuiet, s.toggleCompare),
		}},
	}
	topChildren = append(topChildren, altRows...)
	topChildren = append(topChildren,
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button(acceptLabel, acceptStyle, s.acceptFn()),
			s.button("Use my wording", control.ButtonNormal, s.reviseFn()),
			s.button("Reject", control.ButtonDanger, s.reject),
		}},
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button("Hold research", control.ButtonQuiet, func() { s.hold("research") }),
			s.button("Hold explain", control.ButtonQuiet, func() { s.hold("explain") }),
			s.button("Hold later", control.ButtonQuiet, func() { s.hold("later") }),
		}},
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button("Withdraw", control.ButtonQuiet, s.withdraw),
			s.button("Reopen", control.ButtonQuiet, s.reopenFn()),
			s.button("Reaffirm", control.ButtonQuiet, s.reaffirmFn()),
		}},
		basic.Expanded{Child: basic.Scroll{Child: basic.Column{
			Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 8,
			Children: append(altDetail(th, it, alt, hasAlt, altText, muted, body),
				basic.Row{Spacing: 4, Cross: layout.CrossCenter, Children: []gowaku.Widget{
					s.button("Quote -> Mine", control.ButtonQuiet, func() {
						s.SetState(func() {
							s.mine = it.Quote
							s.diffAltOnly = false
						})
					}),
					s.button("Alt -> Mine", control.ButtonQuiet, copyAltFn),
				}},
				field.Input{
					Label: "Mine", Multiline: true, MinLines: 4, MaxLines: 10,
					Value: s.mine,
					OnChange: func(v string) {
						if v != s.mine {
							s.SetState(func() { s.mine = v })
						}
					},
					OnFocus: func(focused bool) {
						if focused != s.mineFocused {
							s.SetState(func() { s.mineFocused = focused })
						}
					},
				},
				field.Input{
					Label: "Note", Value: s.note,
					OnChange: func(v string) {
						if v != s.note {
							s.SetState(func() { s.note = v })
						}
					},
					OnFocus: func(focused bool) {
						if focused != s.noteFocused {
							s.SetState(func() { s.noteFocused = focused })
						}
					},
				},
			),
		}}},
	)
	return basic.Padding{
		Insets: geom.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8},
		Child:  basic.Column{Cross: layout.CrossStretch, Spacing: 8, Children: topChildren},
	}
}

func altDetail(th *theme.Theme, it Item, alt Alternate, hasAlt bool, altText string, muted, body text.Style) []gowaku.Widget {
	heading := "Alternate"
	shown := altText
	if hasAlt {
		name := altLabel(alt)
		if alt.Ignore {
			heading = "Ignore"
			shown = alt.Reason
			if shown == "" {
				shown = "Ignore."
			}
		} else if name != "" {
			heading = "Alternate · " + name
		}
	}
	out := []gowaku.Widget{basic.Text{Content: heading, Style: muted}}
	if hasAlt {
		if from := SuggestionFrom(it, alt); from != "" {
			out = append(out, basic.Text{Content: "from: " + from, Style: text.Style{Size: 13, Weight: 700, Color: th.Color.Ink}})
		}
	}
	out = append(out, basic.Text{Content: shown, Style: body})
	return out
}

func (s *reviewState) button(label string, style control.ButtonStyle, fn func()) gowaku.Widget {
	return control.Button{Label: label, Style: style, Size: control.ButtonSmall, OnPress: fn}
}

func (s *reviewState) statusBar(th *theme.Theme, muted text.Style) gowaku.Widget {
	return basic.Container{
		Color:   th.Color.Sidebar,
		Padding: geom.Insets{Left: 12, Right: 12, Top: 6, Bottom: 6},
		Child:   basic.Text{Content: s.statusLine(), Style: muted},
	}
}

func (s *reviewState) statusLine() string {
	c := s.session.DetailedCounts()
	holdPart := fmt.Sprintf("hold %d", c.Hold)
	if c.Answered > 0 {
		holdPart = fmt.Sprintf("hold %d (%d ans)", c.Hold, c.Answered)
	}
	line := fmt.Sprintf("open %d · %s · pending %d · applied %d · stale %d", c.Open, holdPart, c.PendingApply, c.Applied, c.Stale)
	if s.status != "" {
		line += " · " + s.status
	}
	if len(s.session.ParseErrors) > 0 {
		line += " · err: " + s.session.ParseErrors[0]
	}
	line += " · j/k [ ] a r h w c Ctrl+R Ctrl+Enter"
	return line
}

func diffSpans(pieces []Piece, size float32, ink, del, ins geom.Color) []text.Span {
	spans := make([]text.Span, 0, len(pieces))
	for _, p := range pieces {
		if p.Text == "" {
			continue
		}
		sp := text.Span{Text: p.Text, Style: text.Style{Size: size, Color: ink}}
		switch p.Op {
		case DiffDelete:
			sp.Background = del
		case DiffInsert:
			sp.Background = ins
		}
		spans = append(spans, sp)
	}
	return spans
}

func (s *reviewState) visible() []Item {
	var out []Item
	for _, it := range s.session.Items {
		d := s.session.Decisions.Items[it.ID].Disposition
		switch s.filter {
		case filterOpen:
			if d != "" {
				continue
			}
		case filterHold:
			if d != "hold" {
				continue
			}
		case filterDecided:
			switch d {
			case "accept", "revise", "reject", "withdraw":
			default:
				continue
			}
		case filterStale:
			if !s.session.IsStale(it.ID) {
				continue
			}
		}
		out = append(out, it)
	}
	return out
}

func (s *reviewState) firstID() string {
	if vis := s.visible(); len(vis) > 0 {
		return vis[0].ID
	}
	for _, it := range s.session.Items {
		if s.session.Decisions.Items[it.ID].Disposition == "" {
			return it.ID
		}
	}
	if len(s.session.Items) > 0 {
		return s.session.Items[0].ID
	}
	return ""
}

func (s *reviewState) selectedVisible() bool {
	for _, it := range s.visible() {
		if it.ID == s.selected {
			return true
		}
	}
	return false
}

func (s *reviewState) current() (Item, bool) {
	for _, it := range s.session.Items {
		if it.ID == s.selected {
			return it, true
		}
	}
	return Item{}, false
}

func (s *reviewState) alt() (Alternate, bool) {
	it, ok := s.current()
	if !ok || s.altIndex < 0 || s.altIndex >= len(it.Alternates) {
		return Alternate{}, false
	}
	return it.Alternates[s.altIndex], true
}

func (s *reviewState) altID() string {
	if a, ok := s.alt(); ok {
		return a.ID
	}
	return ""
}

func (s *reviewState) stashDraft() {
	if s.selected == "" {
		return
	}
	saved := s.session.Decisions.Items[s.selected]
	if s.mine != saved.Text || s.note != saved.Note {
		if s.drafts == nil {
			s.drafts = map[string]itemDraft{}
		}
		s.drafts[s.selected] = itemDraft{mine: s.mine, note: s.note}
	} else {
		delete(s.drafts, s.selected)
	}
}

func (s *reviewState) loadDraft() {
	d := s.session.Decisions.Items[s.selected]
	if dr, ok := s.drafts[s.selected]; ok {
		s.mine = dr.mine
		s.note = dr.note
	} else {
		s.mine = d.Text
		s.note = d.Note
	}
	s.altIndex = -1
	it, ok := s.current()
	if !ok {
		return
	}
	if d.Alternate != "" {
		for i, a := range it.Alternates {
			if a.ID == d.Alternate {
				s.altIndex = i
				return
			}
		}
	}
	if len(it.Alternates) > 0 {
		s.altIndex = 0
	}
}

func (s *reviewState) choose(id string) {
	s.SetState(func() {
		if s.selected != id {
			s.stashDraft()
			s.resetContext()
		}
		s.selected = id
		s.compare = false
		s.diffAltOnly = false
		s.loadDraft()
	})
}

func (s *reviewState) next() { s.move(1) }
func (s *reviewState) prev() { s.move(-1) }

// followingID is the next item in the current filter. If the selected item is
// last in the filter and will drop out of it on save (e.g., Open or Hold),
// it falls back to the preceding visible item rather than stranding the view.
func (s *reviewState) followingID() string {
	vis := s.visible()
	for n, it := range vis {
		if it.ID != s.selected {
			continue
		}
		if n+1 < len(vis) {
			return vis[n+1].ID
		}
		if n > 0 && (s.filter == filterOpen || s.filter == filterHold) {
			return vis[n-1].ID
		}
		return ""
	}
	return ""
}

func (s *reviewState) move(delta int) {
	vis := s.visible()
	if len(vis) == 0 {
		return
	}
	i := 0
	for n, it := range vis {
		if it.ID == s.selected {
			i = n
			break
		}
	}
	i += delta
	if i < 0 {
		i = 0
	}
	if i >= len(vis) {
		i = len(vis) - 1
	}
	id := vis[i].ID
	s.SetState(func() {
		if s.selected != id {
			s.stashDraft()
			s.resetContext()
		}
		s.selected = id
		s.compare = false
		s.diffAltOnly = false
		s.loadDraft()
	})
}

func (s *reviewState) reload() {
	if s.session.Root == "" || s.session.PacketsDir == "" {
		return
	}
	fresh, err := LoadSession(s.session.Root, s.session.PacketsDir, s.session.DecisionsPath)
	s.SetState(func() {
		if err != nil {
			s.status = "reload: " + err.Error()
			return
		}
		s.stashDraft()
		s.session = fresh
		if _, ok := s.current(); !ok || !s.selectedVisible() {
			s.selected = s.firstID()
			s.resetContext()
		}
		s.loadDraft()
		s.status = fmt.Sprintf("reloaded (%d items)", len(fresh.Items))
	})
}

func (s *reviewState) resetContext() {
	s.beforeLines = defaultContextLines
	s.afterLines = defaultContextLines
}

// contextEdges reports whether the passage already shows the start and the
// end of the file. unique is false when the quote is not a single match.
func (s *reviewState) contextEdges() (atStart, atEnd, unique bool) {
	it, ok := s.current()
	if !ok {
		return false, false, false
	}
	file := s.session.bytes(it.Path)
	start, stop, ok := snippetWindow(file, s.session.Hits[it.ID], s.beforeLines, s.afterLines)
	if !ok {
		return false, false, false
	}
	return start == 0, stop == len(file), true
}

func (s *reviewState) contextBar(muted text.Style) gowaku.Widget {
	atStart, atEnd, unique := s.contextEdges()
	if !unique {
		return nil
	}
	tight := s.beforeLines == defaultContextLines && s.afterLines == defaultContextLines
	return basic.Row{Spacing: 4, Cross: layout.CrossCenter, Children: []gowaku.Widget{
		s.button("Above", control.ButtonQuiet, s.moreAbove(atStart)),
		s.button("Below", control.ButtonQuiet, s.moreBelow(atEnd)),
		s.button("Chapter", control.ButtonQuiet, s.showChapter(atStart && atEnd)),
		s.button("Tight", control.ButtonQuiet, s.tighten(tight)),
		basic.Text{Content: contextLabel(s.beforeLines, s.afterLines, atStart, atEnd), Style: muted},
	}}
}

func contextLabel(before, after int, atStart, atEnd bool) string {
	if atStart && atEnd {
		return "whole chapter"
	}
	return sideLabel(before, atStart, "above") + " · " + sideLabel(after, atEnd, "below")
}

func sideLabel(n int, edge bool, name string) string {
	if edge || n < 0 {
		return "all " + name
	}
	return fmt.Sprintf("%d %s", n, name)
}

func (s *reviewState) moreAbove(done bool) func() {
	if done {
		return nil
	}
	return func() { s.SetState(func() { s.beforeLines = moreContext(s.beforeLines) }) }
}

func (s *reviewState) moreBelow(done bool) func() {
	if done {
		return nil
	}
	return func() { s.SetState(func() { s.afterLines = moreContext(s.afterLines) }) }
}

func (s *reviewState) showChapter(done bool) func() {
	if done {
		return nil
	}
	return func() {
		s.SetState(func() {
			s.beforeLines = -1
			s.afterLines = -1
		})
	}
}

func (s *reviewState) tighten(done bool) func() {
	if done {
		return nil
	}
	return func() { s.SetState(func() { s.resetContext() }) }
}

func (s *reviewState) toggleCompare() {
	s.SetState(func() { s.compare = !s.compare })
}

func (s *reviewState) canApply() bool {
	it, ok := s.current()
	if !ok {
		return false
	}
	d := s.session.Decisions.Items[it.ID]
	return CanAccept(s.session.Hits[it.ID], d.Reaffirm)
}

func (s *reviewState) acceptFn() func() {
	a, ok := s.alt()
	if !ok {
		return nil
	}
	if !a.Ignore && !s.canApply() {
		return nil
	}
	return s.accept
}

func (s *reviewState) reviseFn() func() {
	if !s.canApply() || strings.TrimSpace(s.mine) == "" {
		return nil
	}
	return s.revise
}

func (s *reviewState) reopenFn() func() {
	it, ok := s.current()
	if !ok {
		return nil
	}
	if s.session.Decisions.Items[it.ID].Disposition == "" {
		return nil
	}
	return s.reopen
}

func (s *reviewState) reaffirmFn() func() {
	it, ok := s.current()
	if !ok {
		return nil
	}
	h := s.session.Hits[it.ID]
	d := s.session.Decisions.Items[it.ID]
	if (h.Status != StatusAmbiguous && h.Status != StatusMissing) || d.Reaffirm {
		return nil
	}
	return s.reaffirm
}

func (s *reviewState) accept() {
	a, ok := s.alt()
	if !ok {
		return
	}
	if a.Ignore {
		if strings.TrimSpace(s.note) == "" {
			s.note = a.Reason
		}
		s.decide("reject", "", a.ID, "")
		return
	}
	s.decide("accept", "", a.ID, a.Text)
}

func (s *reviewState) revise()             { s.decide("revise", "", s.altID(), "") }
func (s *reviewState) reject()             { s.decide("reject", "", s.altID(), "") }
func (s *reviewState) withdraw()           { s.decide("withdraw", "", s.altID(), "") }
func (s *reviewState) hold(forWhat string) { s.decide("hold", forWhat, s.altID(), "") }

func (s *reviewState) reopen() {
	id := s.selected
	if id == "" || s.session.DecisionsPath == "" {
		return
	}
	next := DecisionFile{Version: 1, Items: map[string]Decision{}}
	for k, v := range s.session.Decisions.Items {
		if k != id {
			next.Items[k] = v
		}
	}
	err := SaveDecisions(s.session.DecisionsPath, next)
	s.SetState(func() {
		if err != nil {
			s.status = err.Error()
			return
		}
		delete(s.drafts, id)
		s.session.Decisions = next
		s.loadDraft()
		s.status = "reopened " + id
	})
}

func (s *reviewState) reaffirm() {
	prev := s.session.Decisions.Items[s.selected]
	s.save(Decision{
		Disposition:  prev.Disposition,
		Alternate:    s.altID(),
		AcceptedText: prev.AcceptedText,
		Text:         s.mine,
		Note:         s.note,
		HoldFor:      prev.HoldFor,
		Reaffirm:     true,
	})
}

func (s *reviewState) decide(disposition, holdFor, alternate, acceptedText string) {
	prev := s.session.Decisions.Items[s.selected]
	txt := s.mine
	if disposition != "revise" {
		txt = prev.Text
		if disposition == "accept" {
			txt = ""
		}
	}
	nextID := s.followingID()
	if !s.save(Decision{
		Disposition:  disposition,
		Alternate:    alternate,
		AcceptedText: acceptedText,
		Text:         txt,
		Note:         s.note,
		HoldFor:      holdFor,
		Reaffirm:     prev.Reaffirm,
	}) {
		return
	}
	if nextID != "" {
		s.choose(nextID)
	}
}

func (s *reviewState) save(d Decision) bool {
	id := s.selected
	if id == "" || s.session.DecisionsPath == "" {
		s.SetState(func() { s.status = "no decisions path" })
		return false
	}
	d = Stamp(d)
	next := DecisionFile{Version: 1, Items: map[string]Decision{}}
	for k, v := range s.session.Decisions.Items {
		next.Items[k] = v
	}
	next.Items[id] = d
	err := SaveDecisions(s.session.DecisionsPath, next)
	s.SetState(func() {
		if err != nil {
			s.status = err.Error()
			return
		}
		delete(s.drafts, id)
		s.session.Decisions = next
		s.status = "saved " + id
	})
	return err == nil
}
