package main

import (
	"github.com/bankai-tensei/gowaku/core"
	"github.com/bankai-tensei/gowaku/geom"
	"github.com/bankai-tensei/gowaku/layout"
	"github.com/bankai-tensei/gowaku/paint"
)

// DiagramEdge is one directed dependency arc drawn behind positioned syntax nodes.
type DiagramEdge struct {
	FromX, FromY float32
	ToX, ToY     float32
	Color        geom.Color
	Width        float32
	Highlight    bool
}

// EdgeCanvas is a lightweight gowaku render widget that draws curved Bezier
// dependency connectors between positioned nodes in ViewDiagram and ViewHybrid.
type EdgeCanvas struct {
	Width, Height float32
	Edges         []DiagramEdge
}

func (EdgeCanvas) Key() core.Key { return nil }

func (e EdgeCanvas) CreateRenderObject() core.RenderObject {
	return &renderEdgeCanvas{w: e.Width, h: e.Height, edges: e.Edges}
}

func (e EdgeCanvas) UpdateRenderObject(ro core.RenderObject) {
	r := ro.(*renderEdgeCanvas)
	if r.w != e.Width || r.h != e.Height {
		r.w, r.h = e.Width, e.Height
		r.MarkNeedsLayout()
	}
	r.edges = e.Edges
	r.MarkNeedsPaint()
}

type renderEdgeCanvas struct {
	core.RenderBase
	w, h  float32
	edges []DiagramEdge
}

func (r *renderEdgeCanvas) Layout(c layout.Constraints) geom.Size {
	return c.Constrain(geom.Size{W: r.w, H: r.h})
}

func (r *renderEdgeCanvas) Paint(cv *paint.Canvas, off geom.Offset) {
	if len(r.edges) == 0 {
		return
	}
	cv.Save()
	cv.Translate(off.X, off.Y)
	for _, e := range r.edges {
		var p paint.Path
		p.MoveTo(e.FromX, e.FromY)
		midY := (e.FromY + e.ToY) / 2
		if e.FromY == e.ToY {
			midY = e.FromY - 22
		}
		p.CubeTo(e.FromX, midY, e.ToX, midY, e.ToX, e.ToY)
		w := e.Width
		if w <= 0 {
			w = 1.4
		}
		if e.Highlight {
			w = 2.4
		}
		cv.DrawPath(&p, paint.Paint{
			Stroke: &paint.Stroke{
				Width: w,
				Color: e.Color,
			},
		})
	}
	cv.Restore()
}
