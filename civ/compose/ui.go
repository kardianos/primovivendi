package main

import (
	"fmt"
	"strings"

	"github.com/bankai-tensei/gowaku"
	"github.com/bankai-tensei/gowaku/geom"
	"github.com/bankai-tensei/gowaku/input"
	"github.com/bankai-tensei/gowaku/layout"
	"github.com/bankai-tensei/gowaku/paint"
	"github.com/bankai-tensei/gowaku/text"
	"github.com/bankai-tensei/gowaku/theme"
	"github.com/bankai-tensei/gowaku/widgets/basic"
	"github.com/bankai-tensei/gowaku/widgets/collection"
	"github.com/bankai-tensei/gowaku/widgets/control"
	"github.com/bankai-tensei/gowaku/widgets/field"
	"github.com/bankai-tensei/gowaku/widgets/interact"
	"po/civ/sys/prof"
)

const (
	szTitle = 17
	szBody  = 14
	szSmall = 12
	szTiny  = 11
)

type composeApp struct {
	gowaku.StatefulBase
	Session *Session
}

type composeState struct {
	gowaku.StateBase
	s              *Session
	latticeHScroll basic.ScrollController
	latticeHPos    basic.ScrollPosition
	diagramHScroll basic.ScrollController
	diagramHPos    basic.ScrollPosition
	lastSentID     string
	suppressScroll bool
}

func (a composeApp) CreateState() gowaku.State {
	lastID := ""
	if a.Session != nil && a.Session.Plan != nil {
		lastID = a.Session.Plan.SentID
	}
	return &composeState{s: a.Session, lastSentID: lastID}
}

func (c *composeState) ui(fn func()) func() {
	if fn == nil {
		return nil
	}
	return func() {
		c.SetState(fn)
		if c.s.Plan != nil && c.s.Plan.SentID != c.lastSentID {
			c.lastSentID = c.s.Plan.SentID
			c.jumpHoriz(&c.latticeHScroll, &c.latticeHPos, 0)
			c.jumpHoriz(&c.diagramHScroll, &c.diagramHPos, 0)
		}
	}
}

func (c *composeState) jumpHoriz(ctrl *basic.ScrollController, pos *basic.ScrollPosition, offset float64) {
	if offset < 0 {
		offset = 0
	}
	if maxOff := ctrl.MaxOffset(); maxOff > 0 && offset > maxOff {
		offset = maxOff
	}
	c.suppressScroll = true
	ctrl.JumpTo(offset)
	c.suppressScroll = false
	pos.Offset = offset
}

func (c *composeState) typing() bool {
	return c.s.RewriteFocus || c.s.LockFocus || c.s.CustomPhraseFocus || c.s.PhraseSearchFocus || c.s.WeightFocus
}

func (c *composeState) scrollHoriz(delta float64) {
	if c.s.View == ViewDiagram {
		c.jumpHoriz(&c.diagramHScroll, &c.diagramHPos, c.diagramHScroll.Offset()+delta)
		return
	}
	c.jumpHoriz(&c.latticeHScroll, &c.latticeHPos, c.latticeHScroll.Offset()+delta)
}

func (c *composeState) Build(ctx gowaku.BuildContext) gowaku.Widget {
	defer prof.Span("ui.build")()
	th := basic.ThemeOf(ctx)
	body := text.Style{Size: szBody, Color: th.Color.Ink}
	muted := text.Style{Size: szSmall, Color: th.Color.Ink2}

	return interact.Actions{
		Actions: c.shortcuts(),
		Child: basic.Column{Cross: layout.CrossStretch, Children: []gowaku.Widget{
			c.topBar(th, muted),
			basic.Expanded{Child: control.SplitPane{
				Pixels: true, Position: 260, MinFirst: 210, MinSecond: 720,
				First: c.navPane(th, muted),
				Second: control.SplitPane{
					Pixels: true, Position: 760, MinFirst: 460, MinSecond: 290,
					First:  c.workspacePane(th, body, muted),
					Second: c.inspectorPane(th, body, muted),
				},
			}},
			c.statusBar(th, muted),
		}},
	}
}

func (c *composeState) shortcuts() []interact.Action {
	var out []interact.Action
	add := func(key input.KeyCode, mod input.Modifiers, fn func()) {
		out = append(out, interact.Action{
			Shortcut: interact.Shortcut{Key: key, Modifiers: mod},
			OnInvoke: fn,
		})
	}
	if !c.typing() {
		add(input.Key1, 0, c.ui(func() { c.s.View = ViewLattice }))
		add(input.Key2, 0, c.ui(func() { c.s.View = ViewDiagram }))
		add(input.Key3, 0, c.ui(func() { c.s.View = ViewHybrid }))
		add(input.KeyJ, 0, c.ui(c.s.NextSentence))
		add(input.KeyK, 0, c.ui(c.s.PrevSentence))
		add(input.KeyN, 0, c.ui(c.s.NextParagraph))
		add(input.KeyP, 0, c.ui(c.s.PrevParagraph))
		add(input.KeyA, 0, c.ui(c.s.Accept))
		add(input.KeyA, input.ModShift, c.ui(c.s.AcceptParagraph))
		add(input.KeyS, 0, c.ui(c.s.KeepAsIs))
		add(input.KeyR, 0, c.ui(c.s.Reject))
		add(input.KeyB, 0, c.ui(func() { c.s.SetPass(PassBackward) }))
		add(input.KeyF, 0, c.ui(func() { c.s.SetPass(PassForward) }))
		add(input.KeyL, 0, c.ui(func() { c.s.SetPass(PassLoop) }))
		add(input.KeyLeft, 0, c.ui(func() { c.scrollHoriz(-220) }))
		add(input.KeyRight, 0, c.ui(func() { c.scrollHoriz(220) }))
		add(input.KeyEscape, 0, c.ui(c.s.ClearSelection))
	}
	add(input.KeyEnter, input.ModControl, c.ui(func() { c.s.ApplyManualRewrite(c.s.RewriteDraft) }))
	return out
}

func (c *composeState) btn(label string, style control.ButtonStyle, fn func()) gowaku.Widget {
	return control.Button{Label: label, Style: style, Size: control.ButtonSmall, OnPress: fn}
}

func (c *composeState) topBar(th *theme.Theme, muted text.Style) gowaku.Widget {
	passBtn := func(label string, dir PassDirection) gowaku.Widget {
		st := control.ButtonQuiet
		if c.s.Dir == dir {
			st = control.ButtonPrimary
		}
		return c.btn(label, st, c.ui(func() { c.s.SetPass(dir) }))
	}
	viewBtn := func(label string, vm ViewMode) gowaku.Widget {
		st := control.ButtonQuiet
		if c.s.View == vm {
			st = control.ButtonPrimary
		}
		return c.btn(label, st, c.ui(func() { c.s.View = vm }))
	}
	return basic.Container{
		Color:   th.Color.Sidebar,
		Padding: geom.Insets{Left: 10, Right: 10, Top: 6, Bottom: 6},
		Border:  paint.Stroke{Width: 1, Color: th.Color.Hair},
		Child: basic.Row{Spacing: 8, Cross: layout.CrossCenter, Children: []gowaku.Widget{
			basic.Text{Content: "Compose", Style: text.Style{Size: szTitle, Weight: 700, Color: th.Color.Ink}},
			basic.Text{Content: c.s.Chapter, Style: text.Style{Size: szBody, Weight: 700, Color: th.Color.Accent}},
			basic.Text{Content: "| Pass:", Style: muted},
			passBtn("Backward (B)", PassBackward),
			passBtn("Forward (F)", PassForward),
			passBtn(fmt.Sprintf("Loop %d (L)", max(2, c.s.PassNum)), PassLoop),
			basic.Text{Content: "| View:", Style: muted},
			viewBtn("1. Lattice", ViewLattice),
			viewBtn("2. Diagram", ViewDiagram),
			viewBtn("3. Hybrid", ViewHybrid),
			basic.Expanded{Child: basic.SizedBox{}},
			c.btn("Train NN", control.ButtonNormal, c.ui(c.s.TrainNN)),
			c.btn(c.s.NNStatus(), control.ButtonQuiet, c.ui(func() {
				c.s.Engine.UseNN = !c.s.Engine.UseNN
				c.s.refreshParagraph()
			})),
			c.btn("Rebuild Chapter", control.ButtonPrimary, c.ui(c.s.Rebuild)),
		}},
	}
}

func (c *composeState) navPane(th *theme.Theme, muted text.Style) gowaku.Widget {
	var chapChips []gowaku.Widget
	for _, ch := range c.s.Chapters {
		ch := ch
		st := control.ButtonQuiet
		if ch == c.s.Chapter {
			st = control.ButtonPrimary
		}
		short := ch
		if idx := strings.IndexByte(ch, '_'); idx >= 0 {
			short = ch[idx+1:]
		}
		chapChips = append(chapChips, c.btn(short, st, c.ui(func() {
			_ = c.s.LoadChapter(ch)
		})))
	}

	paraList := collection.List{
		ItemCount:           len(c.s.ProseBlocks),
		MeasureItems:        true,
		EstimatedItemExtent: 58,
		ItemBuilder: func(i int) gowaku.Widget {
			if i < 0 || i >= len(c.s.ProseBlocks) {
				return basic.SizedBox{}
			}
			blk := c.s.Snapshot.Blocks[c.s.ProseBlocks[i]]
			done := 0
			for _, ss := range blk.Sentences {
				if st, ok := c.s.Store.States[ss.ID]; ok && st.Status != "open" {
					done++
				}
			}
			preview := ""
			if len(blk.Sentences) > 0 {
				preview = truncUI(blk.Sentences[0].Text, 44)
			}
			return interact.Pressable{
				Label:   blk.ID,
				OnPress: c.ui(func() { c.s.JumpParagraph(i) }),
				Build: func(ps interact.PressState) gowaku.Widget {
					bg := th.Color.Sidebar
					if i == c.s.ParaPos {
						bg = th.Color.Accent.WithAlpha(0.22)
					} else if ps.Hovered {
						bg = th.Color.AccentTint
					}
					paraDone := c.s.IsParagraphDone(blk.ID)
					prefix := ""
					if paraDone {
						prefix = "✓ "
					}
					return basic.Container{
						Color:   bg,
						Padding: geom.Insets{Left: 8, Right: 8, Top: 5, Bottom: 5},
						Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 2, Children: []gowaku.Widget{
							basic.Text{
								Content: fmt.Sprintf("%sP%d · %s (%d/%d)", prefix, i+1, truncUI(blk.Heading, 17), done, len(blk.Sentences)),
								Style:   text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink},
							},
							basic.Text{Content: preview, Style: text.Style{Size: szTiny, Color: th.Color.Ink2}},
						}},
					}
				},
			}
		},
	}

	return basic.Column{Cross: layout.CrossStretch, Spacing: 6, Children: []gowaku.Widget{
		basic.Padding{Insets: geom.Insets{Left: 8, Right: 8, Top: 8}, Child: basic.Column{
			Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 6,
			Children: []gowaku.Widget{
				basic.Text{Content: "Chapters", Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
				basic.Wrap{Spacing: 3, RunSpacing: 3, Children: chapChips},
				basic.Text{Content: fmt.Sprintf("Paragraphs (%d)", len(c.s.ProseBlocks)), Style: muted},
			},
		}},
		basic.Expanded{Child: paraList},
	}}
}

func (c *composeState) workspacePane(th *theme.Theme, body, muted text.Style) gowaku.Widget {
	if c.s.Plan == nil {
		return basic.Center{Child: basic.Text{Content: "No prose sentence selected.", Style: muted}}
	}
	var graphWidget gowaku.Widget
	switch c.s.View {
	case ViewDiagram:
		graphWidget = c.diagramView(th, muted, false)
	case ViewHybrid:
		graphWidget = basic.Column{Cross: layout.CrossStretch, Spacing: 6, Children: []gowaku.Widget{
			basic.SizedBox{Height: 210, Child: c.diagramView(th, muted, true)},
			basic.Expanded{Child: c.latticeView(th, muted)},
		}}
	default:
		graphWidget = c.latticeView(th, muted)
	}

	paneChildren := []gowaku.Widget{
		c.paragraphPreviewCard(th, body, muted),
		c.structureStrip(th, muted),
	}
	if c.s.HasSelection() {
		paneChildren = append(paneChildren, c.phraseBar(th, muted))
	}
	paneChildren = append(paneChildren, basic.Expanded{Child: graphWidget})

	return basic.Padding{
		Insets: geom.Insets{Left: 10, Right: 10, Top: 8, Bottom: 8},
		Child:  basic.Column{Cross: layout.CrossStretch, Spacing: 8, Children: paneChildren},
	}
}

func (c *composeState) phraseBar(th *theme.Theme, muted text.Style) gowaku.Widget {
	ph := c.s.SelectedPhraseText()
	cards := make([]gowaku.Widget, 0, len(c.s.PhraseOptions))
	for _, opt := range c.s.PhraseOptions {
		opt := opt
		hovered := c.s.HoverKind == "phrase" && c.s.HoverOptID == opt.ID
		bg := th.Color.Sidebar
		bCol := th.Color.Hair
		if hovered {
			bg = th.Color.AccentTint
			bCol = th.Color.Accent
		}
		cards = append(cards, interact.MouseRegion{
			OnEnter: c.ui(func() { c.s.HoverPhrase(opt.ID) }),
			OnExit:  c.ui(c.s.ClearHover),
			Child: interact.GestureDetector{
				OnTap: c.ui(func() { c.s.ApplyPhraseOption(opt.ID) }),
				Child: basic.Container{
					Color:   bg,
					Border:  paint.Stroke{Width: 1, Color: bCol},
					Radius:  5,
					Padding: geom.Insets{Left: 8, Right: 8, Top: 4, Bottom: 4},
					Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStart, Spacing: 1, Children: []gowaku.Widget{
						basic.Text{Content: opt.Phrase, Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
						basic.Text{
							Content: fmt.Sprintf("%s%s · Sim %.2f · %.1f", etymPrefix(opt.Etym), shortRel(opt.Relation), opt.SemSim, opt.Combined),
							Style:   text.Style{Size: szTiny, Color: th.Color.Ink2},
						},
					}},
				},
			},
		})
	}
	var optionsRow gowaku.Widget
	if len(cards) == 0 {
		optionsRow = basic.Text{Content: "Extend selection, click Expand 2-Hop, search (e.g. steady *, * living), or type custom phrasing.", Style: muted}
	} else {
		optionsRow = basic.SizedBox{
			Height: 48,
			Child: basic.Scroll{
				Horizontal: true,
				Child:      basic.Row{Size: layout.MainMin, Cross: layout.CrossCenter, Spacing: 6, Children: cards},
			},
		}
	}

	// WordNet sense toggle chips (click to disable/re-enable an off-sense like "living: animation").
	var senseChips []gowaku.Widget
	for _, sn := range c.s.ActiveSenses {
		sn := sn
		disabled := c.s.DisabledSenses[sn.Offset]
		prefix := "✓ "
		bStyle := control.ButtonQuiet
		if disabled {
			prefix = "× "
			bStyle = control.ButtonDanger
		}
		lbl := fmt.Sprintf("%s%s: %s", prefix, sn.Word, truncUI(sn.Gloss, 34))
		senseChips = append(senseChips, c.btn(lbl, bStyle, c.ui(func() { c.s.ToggleSense(sn.Offset) })))
	}

	expandStyle := control.ButtonQuiet
	if c.s.Expand2Hop {
		expandStyle = control.ButtonPrimary
	}
	weightCount := len(c.s.LocalBoosts) + len(c.s.LocalDeweights) + len(c.s.Store.Weights.Boost) + len(c.s.Store.Weights.Deweight)
	weightStyle := control.ButtonQuiet
	if c.s.ShowWeightChooser {
		weightStyle = control.ButtonPrimary
	}
	weightLbl := "Weights +/−"
	if weightCount > 0 {
		weightLbl = fmt.Sprintf("Weights +/− (%d)", weightCount)
	}

	barRows := []gowaku.Widget{
		basic.Row{Spacing: 5, Cross: layout.CrossCenter, Children: []gowaku.Widget{
			basic.Text{Content: fmt.Sprintf("Phrase (%d words):", c.s.SelHi-c.s.SelLo+1), Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Accent}},
			basic.Text{Content: fmt.Sprintf("%q", ph), Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
			c.btn("◀ Extend", control.ButtonQuiet, c.ui(c.s.ExtendSelectionLeft)),
			c.btn("Extend ▶", control.ButtonQuiet, c.ui(c.s.ExtendSelectionRight)),
			c.btn("Lock Phrase", control.ButtonQuiet, c.ui(c.s.LockSelectedSpan)),
			c.btn("Lock Global", control.ButtonQuiet, c.ui(c.s.LockSelectedSpanGlobally)),
			basic.Expanded{Child: field.Input{
				Label: "Custom phrasing",
				Value: c.s.CustomPhraseDraft,
				OnChange: func(v string) {
					if v != c.s.CustomPhraseDraft {
						c.SetState(func() { c.s.CustomPhraseDraft = v })
					}
				},
				OnFocus: func(focused bool) {
					if focused != c.s.CustomPhraseFocus {
						c.SetState(func() { c.s.CustomPhraseFocus = focused })
					}
				},
			}},
			c.btn("Apply", control.ButtonPrimary, c.ui(func() { c.s.ApplyCustomPhrase(c.s.CustomPhraseDraft) })),
			c.btn("Clear", control.ButtonQuiet, c.ui(c.s.ClearSelection)),
		}},
		basic.Row{Spacing: 5, Cross: layout.CrossCenter, Children: []gowaku.Widget{
			basic.Expanded{Child: field.Input{
				Label: "Search phrases (steady *, * living, concept, +boost -avoid)",
				Value: c.s.PhraseSearchDraft,
				OnChange: func(v string) {
					if v != c.s.PhraseSearchDraft {
						c.SetState(func() { c.s.PhraseSearchDraft = v })
					}
				},
				OnFocus: func(focused bool) {
					if focused != c.s.PhraseSearchFocus {
						c.SetState(func() { c.s.PhraseSearchFocus = focused })
					}
				},
			}},
			c.btn("Search", control.ButtonNormal, c.ui(func() { c.s.ApplyPhraseSearch(c.s.PhraseSearchDraft) })),
			c.btn("Expand 2-Hop", expandStyle, c.ui(c.s.ToggleExpand2Hop)),
			c.btn(weightLbl, weightStyle, c.ui(func() { c.s.ShowWeightChooser = !c.s.ShowWeightChooser })),
		}},
	}
	if len(senseChips) > 0 {
		barRows = append(barRows, basic.Wrap{Spacing: 4, RunSpacing: 4, Children: senseChips})
	}
	if c.s.ShowWeightChooser {
		barRows = append(barRows, c.weightChooserPanel(th, muted))
	}
	barRows = append(barRows, optionsRow)

	return basic.Container{
		Color:   th.Color.Card,
		Border:  paint.Stroke{Width: 1, Color: th.Color.Accent},
		Radius:  6,
		Padding: geom.Insets{Left: 10, Right: 10, Top: 6, Bottom: 6},
		Child:   basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 5, Children: barRows},
	}
}

// Separate Boost / De-weight Chooser panel: pick terms from the current options
// or enter custom words/phrases to boost (+) or de-weight (−) locally or globally.
func (c *composeState) weightChooserPanel(th *theme.Theme, muted text.Style) gowaku.Widget {
	var quickPicks []gowaku.Widget
	seenPick := map[string]bool{}
	for i, opt := range c.s.PhraseOptions {
		if i >= 6 {
			break
		}
		p := opt.Phrase
		if seenPick[strings.ToLower(p)] {
			continue
		}
		seenPick[strings.ToLower(p)] = true
		pCopy := p
		quickPicks = append(quickPicks,
			c.btn("+ "+pCopy, control.ButtonQuiet, c.ui(func() { c.s.AddLocalBoost(pCopy) })),
			c.btn("− "+pCopy, control.ButtonDanger, c.ui(func() { c.s.AddLocalDeweight(pCopy) })),
		)
	}

	var activePills []gowaku.Widget
	for _, b := range c.s.LocalBoosts {
		b := b
		activePills = append(activePills, c.btn("+ "+b+" ×", control.ButtonPrimary, c.ui(func() { c.s.RemoveLocalBoost(b) })))
	}
	for _, d := range c.s.LocalDeweights {
		d := d
		activePills = append(activePills, c.btn("− "+d+" ×", control.ButtonDanger, c.ui(func() { c.s.RemoveLocalDeweight(d) })))
	}
	for _, b := range c.s.Store.Weights.Boost {
		b := b
		activePills = append(activePills, c.btn("+[book] "+b+" ×", control.ButtonNormal, c.ui(func() { c.s.RemoveGlobalBoost(b) })))
	}
	for _, d := range c.s.Store.Weights.Deweight {
		d := d
		activePills = append(activePills, c.btn("−[book] "+d+" ×", control.ButtonDanger, c.ui(func() { c.s.RemoveGlobalDeweight(d) })))
	}

	etymStyle := control.ButtonQuiet
	etymLbl := "OE+/Grk+/Lat−: OFF"
	if c.s.Store.Weights.Etym.Enabled {
		etymStyle = control.ButtonPrimary
		etymLbl = "OE+/Grk+/Lat−: ON"
	}

	rows := []gowaku.Widget{
		basic.Row{Spacing: 5, Cross: layout.CrossCenter, Children: []gowaku.Widget{
			basic.Text{Content: "Boost / De-weight Chooser:", Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
			c.btn(etymLbl, etymStyle, c.ui(c.s.ToggleEtymBias)),
			basic.Expanded{Child: field.Input{
				Label: "Word or phrase (e.g. steady, animation, firm)",
				Value: c.s.WeightDraft,
				OnChange: func(v string) {
					if v != c.s.WeightDraft {
						c.SetState(func() { c.s.WeightDraft = v })
					}
				},
				OnFocus: func(focused bool) {
					if focused != c.s.WeightFocus {
						c.SetState(func() { c.s.WeightFocus = focused })
					}
				},
			}},
			c.btn("+ Boost", control.ButtonPrimary, c.ui(func() { c.s.AddLocalBoost(c.s.WeightDraft) })),
			c.btn("− De-weight", control.ButtonDanger, c.ui(func() { c.s.AddLocalDeweight(c.s.WeightDraft) })),
			c.btn("+ Book", control.ButtonNormal, c.ui(func() { c.s.AddGlobalBoost(c.s.WeightDraft) })),
			c.btn("− Book", control.ButtonDanger, c.ui(func() { c.s.AddGlobalDeweight(c.s.WeightDraft) })),
		}},
	}
	if len(quickPicks) > 0 {
		rows = append(rows, basic.Wrap{Spacing: 4, RunSpacing: 3, Children: quickPicks})
	}
	if len(activePills) > 0 {
		rows = append(rows, basic.Wrap{Spacing: 4, RunSpacing: 3, Children: activePills})
	} else {
		rows = append(rows, basic.Text{Content: "Pick quick +/− terms above or enter a word/phrase to steer toward (+) or suppress (−).", Style: muted})
	}

	return basic.Container{
		Color:   th.Color.Sidebar,
		Border:  paint.Stroke{Width: 1, Color: th.Color.Hair},
		Radius:  5,
		Padding: geom.Insets{Left: 8, Right: 8, Top: 6, Bottom: 6},
		Child:   basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 4, Children: rows},
	}
}

func (c *composeState) paragraphPreviewCard(th *theme.Theme, body, muted text.Style) gowaku.Widget {
	blk := c.s.ActiveBlock()
	_, texts, _ := c.s.ParagraphTexts()
	pv := c.s.PreviewText()

	spans := make([]text.Span, 0, len(texts)*2)
	for i, t := range texts {
		if i > 0 {
			spans = append(spans, text.Span{Text: " ", Style: body})
		}
		if i == c.s.SentIdx {
			bg := th.Color.AccentTint
			if c.s.HoverText != "" {
				bg = th.Color.Accent.WithAlpha(0.28)
			}
			spans = append(spans, text.Span{
				Text:       pv,
				Style:      text.Style{Size: szBody, Weight: 600, Color: th.Color.Ink},
				Background: bg,
			})
		} else {
			spans = append(spans, text.Span{Text: t, Style: text.Style{Size: szBody, Color: th.Color.Ink2}})
		}
	}

	// Sentence selector pills for quick clicking inside the paragraph.
	var sentPills []gowaku.Widget
	seamSet := map[int]bool{}
	for _, sm := range c.s.Seams {
		seamSet[sm.SentIdx] = true
	}
	if blk != nil {
		for i, ss := range blk.Sentences {
			i := i
			st := c.s.Store.States[ss.ID]
			lbl := fmt.Sprintf("S%d", i+1)
			if i == len(blk.Sentences)-1 {
				lbl += "★"
			}
			if seamSet[i] {
				lbl += "!"
			} else if st != nil && st.Status != "open" {
				lbl += "✓"
			}
			bStyle := control.ButtonQuiet
			if i == c.s.SentIdx {
				bStyle = control.ButtonPrimary
			} else if seamSet[i] {
				bStyle = control.ButtonDanger
			}
			sentPills = append(sentPills, c.btn(lbl, bStyle, c.ui(func() { c.s.JumpSentence(i) })))
		}
	}

	pm := c.s.ParaCadence()
	headerRight := fmt.Sprintf("Lin %.0f · Snap %.0f · Period %.0f · Para CV %.2f · Echoes %d",
		c.s.Plan.Metrics.Score, c.s.Plan.Cadence.Snap, c.s.Plan.Ngram.PeriodFit, pm.LenCV, pm.CloseEchoes)
	goalLine := "Goal (" + string(c.s.Dir) + "): " + c.s.Plan.Goal.Summary
	if c.s.HoverSummary != "" {
		goalLine = c.s.HoverSummary
	}

	return basic.Container{
		Color:   th.Color.Card,
		Border:  paint.Stroke{Width: 1, Color: th.Color.Hair},
		Radius:  6,
		Padding: geom.Insets{Left: 10, Right: 10, Top: 8, Bottom: 8},
		Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 6, Children: []gowaku.Widget{
			basic.Row{Spacing: 6, Cross: layout.CrossCenter, Children: append(
				append([]gowaku.Widget{
					basic.Text{Content: fmt.Sprintf("P%d", c.s.ParaPos+1), Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Accent}},
				}, sentPills...),
				basic.Expanded{Child: basic.SizedBox{}},
				basic.Text{Content: headerRight, Style: muted},
			)},
			basic.RichText{Style: body, Spans: spans},
			basic.Text{Content: goalLine, Style: text.Style{Size: szSmall, Weight: 600, Color: th.Color.Accent}},
		}},
	}
}

func (c *composeState) structureStrip(th *theme.Theme, muted text.Style) gowaku.Widget {
	cards := make([]gowaku.Widget, 0, len(c.s.Plan.Structures))
	for _, sc := range c.s.Plan.Structures {
		sc := sc
		active := sc.ID == c.s.Plan.ActiveStruct
		hovered := c.s.HoverKind == "struct" && c.s.HoverStructID == sc.ID
		bg := th.Color.Sidebar
		borderCol := th.Color.Hair
		if active {
			bg = th.Color.Accent.WithAlpha(0.20)
			borderCol = th.Color.Accent
		} else if hovered {
			bg = th.Color.AccentTint
			borderCol = th.Color.Accent
		}
		card := interact.MouseRegion{
			OnEnter: c.ui(func() { c.s.HoverStructure(sc.ID) }),
			OnExit:  c.ui(c.s.ClearHover),
			Child: interact.GestureDetector{
				OnTap: c.ui(func() { c.s.ToggleStructure(sc.ID) }),
				Child: basic.Container{
					Color:   bg,
					Border:  paint.Stroke{Width: 1, Color: borderCol},
					Radius:  5,
					Padding: geom.Insets{Left: 8, Right: 8, Top: 5, Bottom: 5},
					Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStart, Spacing: 2, Children: []gowaku.Widget{
						basic.Text{Content: sc.Label, Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
						basic.Text{
							Content: fmt.Sprintf("Lin %.0f · Snap %.0f · Score %.1f", sc.Metrics.Score, sc.Cadence.Snap, sc.Combined),
							Style:   text.Style{Size: szTiny, Color: th.Color.Ink2},
						},
					}},
				},
			},
		}
		cards = append(cards, card)
	}
	return basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 4, Children: []gowaku.Widget{
		basic.Text{Content: "Structural Graph Options (hover to preview sentence, click to select/deselect):", Style: muted},
		basic.SizedBox{
			Height: 54,
			Child: basic.Scroll{
				Horizontal: true,
				Child:      basic.Row{Size: layout.MainMin, Cross: layout.CrossCenter, Spacing: 6, Children: cards},
			},
		},
	}}
}

// View 1: Horizontal token/phrase lattice with vertically fanned word options.
func (c *composeState) latticeView(th *theme.Theme, muted text.Style) gowaku.Widget {
	cols := make([]gowaku.Widget, 0, len(c.s.Plan.Slots))
	var totalW float32 = 16
	for idx, slot := range c.s.Plan.Slots {
		idx, slot := idx, slot
		var items []gowaku.Widget

		// Slot header chip (click to select/extend a multi-word phrase; double-click to lock).
		hdrBg := th.Color.Card
		hdrBorder := th.Color.Hair
		badge := slot.POS
		inSel := c.s.HasSelection() && idx >= c.s.SelLo && idx <= c.s.SelHi
		if slot.Locked {
			hdrBg = th.Color.Danger.WithAlpha(0.16)
			hdrBorder = th.Color.Danger
			badge = "LOCK"
		} else if inSel {
			hdrBg = th.Color.Accent.WithAlpha(0.30)
			hdrBorder = th.Color.Accent
			badge = "PHRASE"
		} else if slot.Swapped {
			hdrBg = th.Color.Accent.WithAlpha(0.22)
			hdrBorder = th.Color.Accent
			badge = "SWAP"
		}
		hdr := interact.GestureDetector{
			OnTap:       c.ui(func() { c.s.SelectSlotHeader(idx) }),
			OnDoubleTap: c.ui(func() { c.s.ToggleLockSlot(idx) }),
			Child: basic.Container{
				Color:   hdrBg,
				Border:  paint.Stroke{Width: 1, Color: hdrBorder},
				Radius:  5,
				Padding: geom.Insets{Left: 7, Right: 7, Top: 5, Bottom: 5},
				Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStart, Spacing: 1, Children: []gowaku.Widget{
					basic.Text{Content: slot.Text, Style: text.Style{Size: szBody, Weight: 700, Color: th.Color.Ink}},
					basic.Text{Content: fmt.Sprintf("%s · %s", badge, slot.Rel), Style: text.Style{Size: szTiny, Color: th.Color.Ink2}},
				}},
			},
		}
		items = append(items, hdr)

		if slot.Locked {
			items = append(items, basic.Text{Content: "set phrase", Style: text.Style{Size: szTiny, Color: th.Color.Danger}})
		} else if len(slot.Options) > 0 {
			// Original restore chip when swapped.
			if slot.Swapped {
				origID := fmt.Sprintf("w_%d_orig", slot.Lo)
				items = append(items, interact.GestureDetector{
					OnTap: c.ui(func() { c.s.ToggleWord(idx, origID) }),
					Child: basic.Container{
						Color:   th.Color.Sidebar,
						Border:  paint.Stroke{Width: 1, Color: th.Color.Hair},
						Radius:  4,
						Padding: geom.Insets{Left: 6, Right: 6, Top: 3, Bottom: 3},
						Child:   basic.Text{Content: "↩ " + slot.OrigText, Style: text.Style{Size: szTiny, Color: th.Color.Ink2}},
					},
				})
			}
			for _, opt := range slot.Options {
				opt := opt
				isSel := slot.Text == opt.Surface
				isHov := c.s.HoverKind == "word" && c.s.HoverSlotIdx == idx && c.s.HoverOptID == opt.ID
				bg := th.Color.Sidebar
				bCol := th.Color.Hair
				if isSel {
					bg = th.Color.Accent.WithAlpha(0.24)
					bCol = th.Color.Accent
				} else if isHov {
					bg = th.Color.AccentTint
					bCol = th.Color.Accent
				}
				items = append(items, interact.MouseRegion{
					OnEnter: c.ui(func() { c.s.HoverWord(idx, opt.ID) }),
					OnExit:  c.ui(c.s.ClearHover),
					Child: interact.GestureDetector{
						OnTap: c.ui(func() { c.s.ToggleWord(idx, opt.ID) }),
						Child: basic.Container{
							Color:   bg,
							Border:  paint.Stroke{Width: 1, Color: bCol},
							Radius:  4,
							Padding: geom.Insets{Left: 6, Right: 6, Top: 4, Bottom: 4},
							Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStart, Spacing: 1, Children: []gowaku.Widget{
								basic.Text{Content: opt.Surface, Style: text.Style{Size: szSmall, Weight: 600, Color: th.Color.Ink}},
								basic.Text{
									Content: fmt.Sprintf("%s%s · %.0f", etymPrefix(opt.Etym), shortRel(opt.Relation), opt.Combined),
									Style:   text.Style{Size: szTiny, Color: th.Color.Ink2},
								},
							}},
						},
					},
				})
			}
		}

		colW := float32(92)
		if len(slot.Text) > 10 || slot.Locked {
			colW = 118
		}
		if len(slot.Options) == 0 && !slot.Locked {
			colW = 64
		}
		totalW += colW + 6
		cols = append(cols, basic.SizedBox{
			Width: colW,
			Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 4, Children: items},
		})
	}
	if totalW < 300 {
		totalW = 300
	}

	return basic.Column{Cross: layout.CrossStretch, Spacing: 4, Children: []gowaku.Widget{
		basic.Text{Content: "Word Lattice (hover option to preview · click word to swap · click two headers to select & reconsider a phrase · double-click to lock):", Style: muted},
		basic.Expanded{Child: basic.Scroll{
			Horizontal: true,
			Controller: &c.latticeHScroll,
			OnScroll: func(pos basic.ScrollPosition) {
				if c.suppressScroll {
					c.latticeHPos = pos
					return
				}
				if pos != c.latticeHPos {
					c.SetState(func() { c.latticeHPos = pos })
				}
			},
			Child: basic.SizedBox{
				Width: totalW,
				Child: basic.Scroll{
					Child: basic.Row{Size: layout.MainMin, Cross: layout.CrossStart, Spacing: 6, Children: cols},
				},
			},
		}},
		c.horizScrollBar(muted, &c.latticeHScroll, &c.latticeHPos, totalW, "Lattice horizontal scroll"),
	}}
}

func (c *composeState) horizScrollBar(muted text.Style, ctrl *basic.ScrollController, pos *basic.ScrollPosition, contentW float32, label string) gowaku.Widget {
	maxOff := pos.MaxOffset
	if maxOff <= 0 && ctrl != nil {
		maxOff = ctrl.MaxOffset()
	}
	if maxOff <= 0 {
		maxOff = float64(max(contentW-400, 1))
	}
	cur := pos.Offset
	if cur < 0 {
		cur = 0
	}
	if cur > maxOff {
		cur = maxOff
	}
	pct := 0
	if maxOff > 0 {
		pct = int(cur / maxOff * 100)
	}
	return basic.Row{Spacing: 6, Cross: layout.CrossCenter, Children: []gowaku.Widget{
		c.btn("◀", control.ButtonQuiet, c.ui(func() {
			c.jumpHoriz(ctrl, pos, ctrl.Offset()-220)
		})),
		basic.Expanded{Child: control.Slider{
			Label: label,
			Min:   0,
			Max:   float32(maxOff),
			Value: float32(cur),
			OnChange: func(v float32) {
				c.SetState(func() {
					c.jumpHoriz(ctrl, pos, float64(v))
				})
			},
		}},
		c.btn("▶", control.ButtonQuiet, c.ui(func() {
			c.jumpHoriz(ctrl, pos, ctrl.Offset()+220)
		})),
		basic.Text{Content: fmt.Sprintf("%3d%%", pct), Style: muted},
	}}
}

// View 2 (& top half of View 3): Interactive dependency syntax diagram with drawn Bezier edges.
func (c *composeState) diagramView(th *theme.Theme, muted text.Style, compact bool) gowaku.Widget {
	tr := c.s.Plan.ActiveTree
	if tr == nil || tr.N() == 0 {
		return basic.Center{Child: basic.Text{Content: "No parse tree available.", Style: muted}}
	}
	const (
		nodeW float32 = 92
		nodeH float32 = 40
		gapX  float32 = 14
		gapY  float32 = 58
		pad   float32 = 16
	)
	type pos struct{ x, y float32 }
	coords := make(map[int]pos, len(c.s.Plan.Slots))
	maxDepth := 0
	for _, sl := range c.s.Plan.Slots {
		d := tr.DepthFromRoot(sl.Lo)
		if d > 5 {
			d = 5
		}
		if d > maxDepth {
			maxDepth = d
		}
	}
	for i, sl := range c.s.Plan.Slots {
		d := tr.DepthFromRoot(sl.Lo)
		if d > 5 {
			d = 5
		}
		x := pad + float32(i)*(nodeW+gapX)
		y := pad + float32(d)*gapY
		coords[sl.Lo] = pos{x: x, y: y}
		for t := sl.Lo; t <= sl.Hi; t++ {
			coords[t] = pos{x: x, y: y}
		}
	}
	canvasW := pad*2 + float32(len(c.s.Plan.Slots))*(nodeW+gapX)
	canvasH := pad*2 + float32(maxDepth+1)*gapY + nodeH

	var edges []DiagramEdge
	for i, sl := range c.s.Plan.Slots {
		head := sl.Head
		if head <= 0 {
			continue
		}
		from, ok1 := coords[head]
		to, ok2 := coords[sl.Lo]
		if !ok1 || !ok2 || from == to {
			continue
		}
		hi := i == c.s.DiagramSlot || (c.s.HoverKind == "word" && c.s.HoverSlotIdx == i)
		col := th.Color.Hair
		if hi {
			col = th.Color.Accent
		}
		edges = append(edges, DiagramEdge{
			FromX:     from.x + nodeW/2,
			FromY:     from.y + nodeH,
			ToX:       to.x + nodeW/2,
			ToY:       to.y,
			Color:     col,
			Width:     1.5,
			Highlight: hi,
		})
	}

	stackChildren := []gowaku.Widget{
		EdgeCanvas{Width: canvasW, Height: canvasH, Edges: edges},
	}
	for i, sl := range c.s.Plan.Slots {
		i, sl := i, sl
		p := coords[sl.Lo]
		bg := th.Color.Card
		bCol := th.Color.Hair
		if sl.Locked {
			bg = th.Color.Danger.WithAlpha(0.18)
			bCol = th.Color.Danger
		} else if i == c.s.DiagramSlot {
			bg = th.Color.Accent.WithAlpha(0.24)
			bCol = th.Color.Accent
		} else if sl.Swapped {
			bg = th.Color.AccentTint
			bCol = th.Color.Accent
		}
		node := interact.GestureDetector{
			OnTap: c.ui(func() {
				c.s.DiagramSlot = i
				c.s.SelectSlotHeader(i)
			}),
			OnDoubleTap: c.ui(func() { c.s.ToggleLockSlot(i) }),
			Child: basic.Container{
				Width:   nodeW,
				Height:  nodeH,
				Color:   bg,
				Border:  paint.Stroke{Width: 1, Color: bCol},
				Radius:  6,
				Padding: geom.Insets{Left: 5, Right: 5, Top: 3, Bottom: 3},
				Child: basic.Column{Size: layout.MainMin, Cross: layout.CrossCenter, Spacing: 1, Children: []gowaku.Widget{
					basic.Text{Content: truncUI(sl.Text, 12), Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
					basic.Text{Content: fmt.Sprintf("%s·%s", sl.Rel, sl.POS), Style: text.Style{Size: szTiny, Color: th.Color.Ink2}},
				}},
			},
		}
		stackChildren = append(stackChildren, basic.Positioned{Left: p.x, Top: p.y, Child: node})
	}

	diagramScroll := basic.Scroll{
		Horizontal: true,
		Controller: &c.diagramHScroll,
		OnScroll: func(pos basic.ScrollPosition) {
			if c.suppressScroll {
				c.diagramHPos = pos
				return
			}
			if pos != c.diagramHPos {
				c.SetState(func() { c.diagramHPos = pos })
			}
		},
		Child: basic.SizedBox{
			Width: canvasW,
			Child: basic.Scroll{
				Child: basic.SizedBox{Width: canvasW, Height: canvasH, Child: basic.Stack{Children: stackChildren}},
			},
		},
	}

	if compact {
		return basic.Column{Cross: layout.CrossStretch, Spacing: 3, Children: []gowaku.Widget{
			basic.Text{Content: "Syntax Dependency Diagram (click node to select phrase · double-click to lock):", Style: muted},
			basic.Expanded{Child: diagramScroll},
			c.horizScrollBar(muted, &c.diagramHScroll, &c.diagramHPos, canvasW, "Diagram horizontal scroll"),
		}}
	}

	// In full ViewDiagram, render a focused word-options drawer below the diagram for the selected node.
	var drawerItems []gowaku.Widget
	if c.s.DiagramSlot >= 0 && c.s.DiagramSlot < len(c.s.Plan.Slots) {
		sl := c.s.Plan.Slots[c.s.DiagramSlot]
		drawerItems = append(drawerItems,
			c.btn(lockBtnLabel(sl), control.ButtonQuiet, c.ui(func() { c.s.ToggleLockSlot(c.s.DiagramSlot) })),
		)
		for _, opt := range sl.Options {
			opt := opt
			isSel := sl.Text == opt.Surface
			st := control.ButtonQuiet
			if isSel {
				st = control.ButtonPrimary
			}
			drawerItems = append(drawerItems, interact.MouseRegion{
				OnEnter: c.ui(func() { c.s.HoverWord(c.s.DiagramSlot, opt.ID) }),
				OnExit:  c.ui(c.s.ClearHover),
				Child:   c.btn(fmt.Sprintf("%s (%s %.0f)", opt.Surface, shortRel(opt.Relation), opt.Combined), st, c.ui(func() { c.s.ToggleWord(c.s.DiagramSlot, opt.ID) })),
			})
		}
	}
	return basic.Column{Cross: layout.CrossStretch, Spacing: 6, Children: []gowaku.Widget{
		basic.Text{Content: "Dependency Syntax Diagram (click nodes to select phrase/options · double-click to lock):", Style: muted},
		basic.Expanded{Child: diagramScroll},
		c.horizScrollBar(muted, &c.diagramHScroll, &c.diagramHPos, canvasW, "Diagram horizontal scroll"),
		basic.Container{
			Color:   th.Color.Card,
			Border:  paint.Stroke{Width: 1, Color: th.Color.Hair},
			Radius:  5,
			Padding: geom.Insets{Left: 8, Right: 8, Top: 6, Bottom: 6},
			Child:   basic.Wrap{Spacing: 5, RunSpacing: 5, Children: drawerItems},
		},
	}}
}

func (c *composeState) inspectorPane(th *theme.Theme, body, muted text.Style) gowaku.Widget {
	if c.s.Plan == nil {
		return basic.SizedBox{}
	}
	_, st := c.s.ActiveSentence()

	// Seam issues list for the active paragraph.
	var seamWidgets []gowaku.Widget
	if len(c.s.Seams) == 0 {
		seamWidgets = append(seamWidgets, basic.Text{Content: "All paragraph seams clean.", Style: muted})
	} else {
		for _, sm := range c.s.Seams {
			sm := sm
			seamWidgets = append(seamWidgets, interact.GestureDetector{
				OnTap: c.ui(func() { c.s.JumpSentence(sm.SentIdx) }),
				Child: basic.Container{
					Color:   th.Color.Danger.WithAlpha(0.10),
					Radius:  4,
					Padding: geom.Insets{Left: 6, Right: 6, Top: 4, Bottom: 4},
					Child:   basic.Text{Content: fmt.Sprintf("[%s] %s", sm.Kind, sm.Message), Style: text.Style{Size: szTiny, Color: th.Color.Ink}},
				},
			})
		}
	}

	// Global set-phrases badges.
	var lockChips []gowaku.Widget
	for _, phrase := range c.s.Store.Glossary {
		phrase := phrase
		lockChips = append(lockChips, c.btn("🔒 "+phrase+" ×", control.ButtonQuiet, c.ui(func() {
			_ = c.s.Store.RemoveGlobalPhrase(phrase)
			c.s.refreshParagraph()
		})))
	}

	historyCount := 0
	statusStr := "open"
	if st != nil {
		historyCount = len(st.History)
		statusStr = st.Status
	}

	return basic.Padding{
		Insets: geom.Insets{Left: 8, Right: 10, Top: 8, Bottom: 8},
		Child: basic.Scroll{Child: basic.Column{
			Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 8,
			Children: []gowaku.Widget{
				basic.Text{Content: fmt.Sprintf("Decision (%s · %d tree nodes)", statusStr, historyCount), Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
				basic.Row{Spacing: 4, Children: []gowaku.Widget{
					c.btn("Accept (A)", control.ButtonPrimary, c.ui(c.s.Accept)),
					c.btn("Keep As Is (S)", control.ButtonNormal, c.ui(c.s.KeepAsIs)),
					c.btn("Reject (R)", control.ButtonDanger, c.ui(c.s.Reject)),
				}},
				basic.Row{Spacing: 4, Children: []gowaku.Widget{
					c.btn("Accept Paragraph & Next (Shift+A)", control.ButtonPrimary, c.ui(c.s.AcceptParagraph)),
				}},
				basic.Row{Spacing: 4, Children: []gowaku.Widget{
					c.btn("Prev Sent (K)", control.ButtonQuiet, c.ui(c.s.PrevSentence)),
					c.btn("Next Sent (J)", control.ButtonQuiet, c.ui(c.s.NextSentence)),
				}},
				basic.Text{Content: "Manual Rewrite (becomes new diagrammed base):", Style: muted},
				field.Input{
					Label: "Rewrite base", Multiline: true, MinLines: 3, MaxLines: 6,
					Value: c.s.RewriteDraft,
					OnChange: func(v string) {
						if v != c.s.RewriteDraft {
							c.SetState(func() { c.s.RewriteDraft = v })
						}
					},
					OnFocus: func(focused bool) {
						if focused != c.s.RewriteFocus {
							c.SetState(func() { c.s.RewriteFocus = focused })
						}
					},
				},
				basic.Row{Spacing: 4, Children: []gowaku.Widget{
					c.btn("Use Rewrite as Base (Ctrl+Enter)", control.ButtonNormal, c.ui(func() { c.s.ApplyManualRewrite(c.s.RewriteDraft) })),
				}},
				basic.Text{Content: "Global Set Phrases (never split or swapped):", Style: muted},
				basic.Row{Spacing: 4, Cross: layout.CrossCenter, Children: []gowaku.Widget{
					basic.Expanded{Child: field.Input{
						Label: "Add set phrase", Value: c.s.LockDraft,
						OnChange: func(v string) {
							if v != c.s.LockDraft {
								c.SetState(func() { c.s.LockDraft = v })
							}
						},
						OnFocus: func(focused bool) {
							if focused != c.s.LockFocus {
								c.SetState(func() { c.s.LockFocus = focused })
							}
						},
					}},
					c.btn("Lock", control.ButtonQuiet, c.ui(func() { c.s.LockGlobalPhrase(c.s.LockDraft) })),
				}},
				basic.Wrap{Spacing: 4, RunSpacing: 4, Children: lockChips},
				basic.Text{Content: fmt.Sprintf("Paragraph Seams (%d):", len(c.s.Seams)), Style: text.Style{Size: szSmall, Weight: 700, Color: th.Color.Ink}},
				basic.Column{Size: layout.MainMin, Cross: layout.CrossStretch, Spacing: 4, Children: seamWidgets},
			},
		}},
	}
}

func (c *composeState) statusBar(th *theme.Theme, muted text.Style) gowaku.Widget {
	wnBadge := "WordNet: fallback"
	if c.s.Engine.Lex != nil && c.s.Engine.Lex.WordNetReady {
		wnBadge = "WordNet 3.1"
	}
	embBadge := "BGE: off"
	if c.s.Engine.Embed != nil && c.s.Engine.Embed.Ready() {
		embBadge = "BGE-small: ready"
	}
	return basic.Container{
		Color:   th.Color.Sidebar,
		Border:  paint.Stroke{Width: 1, Color: th.Color.Hair},
		Padding: geom.Insets{Left: 10, Right: 10, Top: 4, Bottom: 4},
		Child: basic.Row{Spacing: 10, Cross: layout.CrossCenter, Children: []gowaku.Widget{
			basic.Expanded{Child: basic.Text{Content: c.s.StatusMsg, Style: muted}},
			basic.Text{Content: c.s.Engine.ParserStatus() + " · " + wnBadge + " · " + embBadge + " · " + fmt.Sprintf("%d tree nodes", len(c.s.Store.Nodes)), Style: muted},
		}},
	}
}

func lockBtnLabel(sl WordSlot) string {
	if sl.Locked {
		return "Unlock " + sl.Text
	}
	return "Lock " + sl.Text
}

func etymPrefix(e string) string {
	if e == "" {
		return ""
	}
	return "[" + e + "] "
}

func shortRel(r string) string {
	switch r {
	case "synonym":
		return "syn"
	case "similar":
		return "sim"
	case "hyponym":
		return "spec"
	case "hypernym":
		return "gen"
	case "fallback":
		return "alt"
	case "wn_phrase":
		return "wn"
	case "template":
		return "tmpl"
	case "idiom":
		return "idiom"
	case "compositional":
		return "comp"
	case "period_ngram":
		return "1850s"
	case "wildcard":
		return "corpus"
	case "gloss":
		return "gloss"
	case "boosted":
		return "+boost"
	case "hop2":
		return "2-hop"
	case "derivation":
		return "deriv"
	default:
		return r
	}
}

func truncUI(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}
