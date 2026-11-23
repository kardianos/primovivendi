package main

import (
	"fmt"
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
	textBody  = 14
	textMeta  = 12
	textTitle = 18
)

// reviewApp is the three-pane viewer. Session is shared with the caller.
type reviewApp struct {
	gowaku.StatefulBase
	Session *Session
}

// reviewState is the gowaku shell around reviewView. Mutations go through ui
// so the window rebuilds. The terminal calls reviewView directly.
type reviewState struct {
	gowaku.StateBase
	reviewView
}

func (a reviewApp) CreateState() gowaku.State {
	return &reviewState{reviewView: newReviewView(a.Session)}
}

// ui runs fn and rebuilds the window. fn may be nil when a control is disabled.
func (s *reviewState) ui(fn func()) func() {
	if fn == nil {
		return nil
	}
	return func() { s.SetState(fn) }
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
		add(input.KeyJ, 0, s.ui(s.next))
		add(input.KeyK, 0, s.ui(s.prev))
		add(input.KeyN, 0, s.ui(s.next))
		add(input.KeyP, 0, s.ui(s.prev))
		add(input.KeyA, 0, s.acceptFn())
		add(input.KeyR, 0, s.ui(s.reject))
		add(input.KeyH, 0, s.ui(func() { s.hold("later") }))
		add(input.KeyW, 0, s.ui(s.withdraw))
		add(input.KeyC, 0, s.ui(s.toggleCompare))
		atStart, atEnd, unique := s.contextEdges()
		if unique {
			add(input.KeyBracketLeft, 0, s.moreAbove(atStart))
			add(input.KeyBracketRight, 0, s.moreBelow(atEnd))
		}
	}
	add(input.KeyJ, input.ModControl, s.ui(s.next))
	add(input.KeyK, input.ModControl, s.ui(s.prev))
	add(input.KeyR, input.ModControl, s.ui(s.reload))
	add(input.KeyEnter, input.ModControl, s.reviseFn())
	return out
}

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
					s.button("Reload", control.ButtonQuiet, s.ui(s.reload)),
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
		OnPress: s.ui(func() { s.setFilter(f) }),
	}
}

func (s *reviewState) itemRow(th *theme.Theme, it Item, muted text.Style) gowaku.Widget {
	h := s.session.Hits[it.ID]
	title := rowTitle(it, h)
	preview := singleLinePreview(it.Quote, 46)
	meta := itemBadges(s.session, it)
	return interact.Pressable{
		Label:   it.ID,
		OnPress: s.ui(func() { s.choose(it.ID) }),
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
				Content: lockedBanner,
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

	_, hasAlt := s.alt()
	hasMine := strings.TrimSpace(s.mine) != ""
	diffTitle, target, hasTarget, showIgnore, reason := s.passageDiff(it)
	diffHeaderChildren := []gowaku.Widget{basic.Expanded{Child: basic.Text{Content: diffTitle, Style: muted}}}
	if hasMine && hasAlt {
		toggleLabel := "Diff Alt"
		if s.diffAltOnly {
			toggleLabel = "Diff Mine"
		}
		diffHeaderChildren = append(diffHeaderChildren, s.button(toggleLabel, control.ButtonQuiet, s.ui(s.toggleDiffTarget)))
	}
	children = append(children, basic.Row{Spacing: 8, Cross: layout.CrossCenter, Children: diffHeaderChildren})
	if showIgnore {
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
				btn := s.button(label, style, s.ui(func() { s.selectAlt(i) }))
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
		copyAltFn = s.ui(s.copyAltToMine)
	}
	acceptLabel := "Accept"
	acceptStyle := control.ButtonPrimary
	if hasAlt && alt.Ignore {
		acceptLabel = "Ignore this"
		acceptStyle = control.ButtonDanger
	}
	topChildren := []gowaku.Widget{
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button("Previous", control.ButtonQuiet, s.ui(s.prev)),
			s.button("Next", control.ButtonQuiet, s.ui(s.next)),
			s.button("Compare", control.ButtonQuiet, s.ui(s.toggleCompare)),
		}},
	}
	topChildren = append(topChildren, altRows...)
	topChildren = append(topChildren,
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button(acceptLabel, acceptStyle, s.acceptFn()),
			s.button("Use my wording", control.ButtonNormal, s.reviseFn()),
			s.button("Reject", control.ButtonDanger, s.ui(s.reject)),
		}},
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button("Hold research", control.ButtonQuiet, s.ui(func() { s.hold("research") })),
			s.button("Hold explain", control.ButtonQuiet, s.ui(func() { s.hold("explain") })),
			s.button("Hold later", control.ButtonQuiet, s.ui(func() { s.hold("later") })),
		}},
		basic.Row{Spacing: 4, Children: []gowaku.Widget{
			s.button("Withdraw", control.ButtonQuiet, s.ui(s.withdraw)),
			s.button("Reopen", control.ButtonQuiet, s.reopenFn()),
			s.button("Reaffirm", control.ButtonQuiet, s.reaffirmFn()),
		}},
		basic.Expanded{Child: basic.Scroll{Child: basic.Column{
			Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 8,
			Children: append(altDetail(th, it, alt, hasAlt, altText, muted, body),
				basic.Row{Spacing: 4, Cross: layout.CrossCenter, Children: []gowaku.Widget{
					s.button("Quote -> Mine", control.ButtonQuiet, s.ui(s.copyQuoteToMine)),
					s.button("Alt -> Mine", control.ButtonQuiet, copyAltFn),
				}},
				field.Input{
					Label: "Mine", Multiline: true, MinLines: 4, MaxLines: 10,
					Value: s.mine,
					OnChange: func(v string) {
						if v != s.mine {
							s.ui(func() { s.setMine(v) })()
						}
					},
					OnFocus: func(focused bool) {
						if focused != s.mineFocused {
							s.ui(func() { s.setMineFocused(focused) })()
						}
					},
				},
				field.Input{
					Label: "Note", Value: s.note,
					OnChange: func(v string) {
						if v != s.note {
							s.ui(func() { s.setNote(v) })()
						}
					},
					OnFocus: func(focused bool) {
						if focused != s.noteFocused {
							s.ui(func() { s.setNoteFocused(focused) })()
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
	heading, from, shown := altDetailText(it, alt, hasAlt)
	if !hasAlt {
		shown = altText
	}
	out := []gowaku.Widget{basic.Text{Content: heading, Style: muted}}
	if from != "" {
		out = append(out, basic.Text{Content: "from: " + from, Style: text.Style{Size: 13, Weight: 700, Color: th.Color.Ink}})
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

func (s *reviewState) moreAbove(done bool) func() {
	if done {
		return nil
	}
	return s.ui(s.expandAbove)
}

func (s *reviewState) moreBelow(done bool) func() {
	if done {
		return nil
	}
	return s.ui(s.expandBelow)
}

func (s *reviewState) showChapter(done bool) func() {
	if done {
		return nil
	}
	return s.ui(s.showWholeChapter)
}

func (s *reviewState) tighten(done bool) func() {
	if done {
		return nil
	}
	return s.ui(s.tightenContext)
}

func (s *reviewState) acceptFn() func() {
	if !s.acceptReady() {
		return nil
	}
	return s.ui(s.accept)
}

func (s *reviewState) reviseFn() func() {
	if !s.reviseReady() {
		return nil
	}
	return s.ui(s.revise)
}

func (s *reviewState) reopenFn() func() {
	if !s.reopenReady() {
		return nil
	}
	return s.ui(s.reopen)
}

func (s *reviewState) reaffirmFn() func() {
	if !s.reaffirmReady() {
		return nil
	}
	return s.ui(s.reaffirm)
}
