// Command cover writes the civ book cover as a one-page vector PDF.
//
// The page is 6in x 9in (432 x 648 pt), the book's trim size. Everything on
// it is a path in one content stream: a double frame with notched corners,
// seven concentric arches, a door, and the title and author as glyph
// outlines. No font is embedded, so the file is a few kilobytes.
//
// The text is shaped with HarfBuzz (go-text/typesetting), so the font's
// kerning and small caps apply, and each glyph's outline is written as path
// operators.
//
// It also writes the cover as a PNG for the EPUB, rendered from the PDF by
// Ghostscript and kept to four grays: black, white, and two for the
// antialiased edges. That is a 2-bit PNG, half the size of 8-bit gray.
//
// Usage:
//
//	go run . [-font file] [-plain] [-o cover.pdf] [-png cover.png]
//
// build.fish puts cover.pdf in front of the book's title page and gives
// cover.png to the EPUB as its cover image.
package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

// Page, in pt.
const (
	pageW = 432.0
	pageH = 648.0
	midX  = pageW / 2
)

// A line of text, centered on its ink.
type textLine struct {
	text     string
	size     float64  // pt
	baseline float64  // pt from the page bottom
	tracking float64  // space added after each glyph but the last, em
	weight   float64  // stroke added around each glyph, pt; 0 draws the font as is
	features []string // OpenType features to turn on, such as "smcp"
}

// The title, three lines 54.2 pt apart, then the author in the open band
// between the arches' ground and the frame.
var textLines = []textLine{
	{text: "Civilization’s", size: 61, baseline: 541.5, weight: 0.5},
	{text: "Common", size: 61, baseline: 487.3, weight: 0.5},
	{text: "Principles", size: 61, baseline: 433.1, weight: 0.5},
	{text: "Daniel Theophanes", size: 22, baseline: 70, tracking: 0.08, weight: 0.15, features: []string{"smcp"}},
}

// Frame: two rules whose corners are cut by circles centered on the corners
// of the outer rule. The inner rule's cut is the parallel of the outer's, so
// the gap holds around the corner.
const (
	frameInset  = 17.5 // outer rule's centerline, from the page edge
	frameGap    = 4.2  // outer to inner rule, centerline to centerline
	frameNotch  = 11.0 // radius of the outer rule's corner cut
	frameOuterW = 1.4
	frameInnerW = 0.7
)

// Door and arches. Arch i = 0 is the innermost. Each arch is an ellipse
// about the page's center line, cut off at the ground. The innermost ellipse
// rests on the ground; each step outward widens, heightens, and lifts it.
const (
	ground = 133.5 // the door and the arches stand on this line
	doorW  = 104.5
	doorH  = 127.2

	arches   = 7
	archA0   = 70.9  // half width of the innermost arch
	archDA   = 11.77 // added per step outward
	archB0   = 82.0  // half height of the innermost arch
	archDB   = 13.07 // added per step outward
	archDC   = 3.45  // each step outward lifts the ellipse center this much
	archTop0 = 4.4   // ink at the crown of the innermost arch
	archDTop = -0.5  // added per step outward
	archSide = 0.5   // ink at the widest point, as a fraction of the crown's
	archFoot = 0.4   // ink the crescent would close to at the ellipse bottom
)

// kappa places the control points of a quarter-circle cubic.
const kappa = 4.0 / 3 * (math.Sqrt2 - 1)

// num formats v rounded to places decimals, in the fewest characters.
func num(v float64, places int) string {
	p := math.Pow10(places)
	s := strconv.FormatFloat(math.Round(v*p)/p, 'f', -1, 64)
	switch {
	case s == "-0":
		return "0"
	case strings.HasPrefix(s, "0."):
		return s[1:]
	case strings.HasPrefix(s, "-0."):
		return "-" + s[2:]
	}
	return s
}

// content is a PDF content stream.
type content struct{ strings.Builder }

// op writes the operands, rounded to 0.01, then the operator, as one line.
func (c *content) op(op string, vs ...float64) {
	for _, v := range vs {
		c.WriteString(num(v, 2))
		c.WriteByte(' ')
	}
	c.WriteString(op)
	c.WriteByte('\n')
}

// arc appends one cubic for a circular arc of under 90°, from angle a0 to a1.
func (c *content) arc(cx, cy, r, a0, a1 float64) {
	k := 4.0 / 3 * math.Tan((a1-a0)/4)
	c0, s0 := math.Cos(a0), math.Sin(a0)
	c1, s1 := math.Cos(a1), math.Sin(a1)
	c.op("c",
		cx+r*(c0-k*s0), cy+r*(s0+k*c0),
		cx+r*(c1+k*s1), cy+r*(s1-k*c1),
		cx+r*c1, cy+r*s1)
}

// notchedRect appends a rule at inset from the page edge whose corners are
// cut by circles of radius r centered on the outer rule's corners.
func (c *content) notchedRect(inset, r float64) {
	const q = math.Pi / 2
	lo, hx, hy := frameInset, pageW-frameInset, pageH-frameInset
	x0, y0, x1, y1 := inset, inset, pageW-inset, pageH-inset
	d := inset - frameInset   // rule's offset from the circle centers
	e := math.Sqrt(r*r - d*d) // where a circle meets the rule
	t := math.Atan2(d, e)     // that point's angle off the axis
	c.op("m", lo+e, y0)       // counterclockwise from the bottom left
	c.op("l", hx-e, y0)
	c.arc(hx, lo, r, 2*q-t, q+t)
	c.op("l", x1, hy-e)
	c.arc(hx, hy, r, 3*q-t, 2*q+t)
	c.op("l", lo+e, y1)
	c.arc(lo, hy, r, -t, -q+t)
	c.op("l", x0, lo+e)
	c.arc(lo, lo, r, q-t, t)
	c.op("h")
}

func (c *content) frame() {
	c.op("w", frameOuterW)
	c.notchedRect(frameInset, frameNotch)
	c.op("S")
	c.op("w", frameInnerW)
	c.notchedRect(frameInset+frameGap, frameNotch+frameGap)
	c.op("S")
}

// ellipse appends a closed ellipse centered on (0, cy).
func (c *content) ellipse(cy, a, b float64) {
	ka, kb := kappa*a, kappa*b
	c.op("m", a, cy)
	c.op("c", a, cy+kb, ka, cy+b, 0, cy+b)
	c.op("c", -ka, cy+b, -a, cy+kb, -a, cy)
	c.op("c", -a, cy-kb, -ka, cy-b, 0, cy-b)
	c.op("c", ka, cy-b, a, cy-kb, a, cy)
}

// arches draws each arch as the crescent between two ellipses, filled
// even-odd. The inner edge's center is dropped so the ink is archTop thick
// at the crown, about archSide of that at the sides, and would close to
// archFoot at the ellipse bottom. The clip keeps what stands above the
// ground; the door then covers the inner arches' feet.
func (c *content) arches() {
	c.op("q")
	c.op("re", 0, ground, pageW, pageH-ground)
	c.op("W n")
	c.op("cm", 1, 0, 0, 1, midX, 0)
	for i := range arches {
		f := float64(i)
		a := archA0 + f*archDA
		b := archB0 + f*archDB
		cy := ground + archB0 + f*archDC
		top := archTop0 + f*archDTop
		side := archSide * top
		drop := (top - archFoot) / 2
		c.ellipse(cy, a+side/2, b+top/2)
		c.ellipse(cy-drop, a-side/2, b-top/2+drop)
	}
	c.op("f*")
	c.op("Q")
	c.op("re", midX-doorW/2, ground, doorW, doorH)
	c.op("f")
}

// outline returns glyph g as path operators in whole font units.
func outline(face *font.Face, g font.GID) string {
	var b strings.Builder
	put := func(op string, pts ...ot.SegmentPoint) {
		for _, p := range pts {
			fmt.Fprintf(&b, "%d %d ", int(math.Round(float64(p.X))), int(math.Round(float64(p.Y))))
		}
		b.WriteString(op)
		b.WriteByte('\n')
	}
	data, ok := face.GlyphData(g).(font.GlyphOutline)
	if !ok {
		log.Fatalf("glyph %d is not an outline", g)
	}
	var cur ot.SegmentPoint
	for i, s := range data.Segments {
		a := s.Args
		switch s.Op {
		case ot.SegmentOpMoveTo:
			if i > 0 {
				b.WriteString("h\n")
			}
			put("m", a[0])
			cur = a[0]
		case ot.SegmentOpLineTo:
			put("l", a[0])
			cur = a[0]
		case ot.SegmentOpQuadTo:
			// TrueType quadratic, raised to a cubic
			c1 := ot.SegmentPoint{X: cur.X + 2*(a[0].X-cur.X)/3, Y: cur.Y + 2*(a[0].Y-cur.Y)/3}
			c2 := ot.SegmentPoint{X: a[1].X + 2*(a[0].X-a[1].X)/3, Y: a[1].Y + 2*(a[0].Y-a[1].Y)/3}
			put("c", c1, c2, a[1])
			cur = a[1]
		case ot.SegmentOpCubeTo:
			put("c", a[0], a[1], a[2])
			cur = a[2]
		}
	}
	if len(data.Segments) > 0 {
		b.WriteString("h\n")
	}
	return b.String()
}

// text draws each line as glyph outlines, centered on its ink. A line is
// drawn in font units under one scaling cm; each glyph moves the origin to
// its own pen position, so a repeated glyph repeats the same bytes and
// compresses to almost nothing.
func (c *content) text(face *font.Face, lines []textLine) {
	upem := float64(face.Upem())
	fu := func(v fixed.Int26_6) float64 { return float64(v) / 64 }
	outlines := map[font.GID]string{}
	var shaper shaping.HarfbuzzShaper

	for _, line := range lines {
		text := []rune(line.text)
		var feats []shaping.FontFeature
		for _, f := range line.features {
			feats = append(feats, shaping.FontFeature{Tag: ot.MustNewTag(f), Value: 1})
		}
		out := shaper.Shape(shaping.Input{
			Text:         text,
			RunEnd:       len(text),
			Direction:    di.DirectionLTR,
			Face:         face,
			Size:         fixed.I(int(upem)), // positions come back in font units
			Script:       language.Latin,
			Language:     language.NewLanguage("en"),
			FontFeatures: feats,
		})
		gl := out.Glyphs
		xs := make([]float64, len(gl))
		pen := 0.0
		for i, g := range gl {
			xs[i] = pen + fu(g.XOffset)
			pen += fu(g.Advance) + line.tracking*upem
		}
		last := len(gl) - 1
		left := xs[0] + fu(gl[0].XBearing)
		right := xs[last] + fu(gl[last].XBearing) + fu(gl[last].Width)
		scale := line.size / upem
		x0 := midX - (left+right)/2*scale
		fmt.Fprintf(c, "q %s 0 0 %s %s %s cm\n", num(scale, 6), num(scale, 6), num(x0, 2), num(line.baseline, 2))
		paint := "f"
		if line.weight > 0 {
			paint = "B"
			c.op("j", 1)
			c.op("w", line.weight/scale)
		}
		at := 0.0
		for i, g := range gl {
			s, ok := outlines[g.GlyphID]
			if !ok {
				s = outline(face, g.GlyphID)
				outlines[g.GlyphID] = s
			}
			if s == "" { // a space: no ink, only its advance
				continue
			}
			if dx := xs[i] - at; dx != 0 {
				c.op("cm", 1, 0, 0, 1, dx, 0)
				at = xs[i]
			}
			c.WriteString(s)
			c.op(paint)
		}
		c.op("Q")
	}
}

// document wraps one content stream as a single-page PDF.
func document(stream []byte, compress bool) []byte {
	filter := ""
	if compress {
		var z bytes.Buffer
		w, _ := zlib.NewWriterLevel(&z, zlib.BestCompression)
		w.Write(stream)
		w.Close()
		stream = z.Bytes()
		filter = "/Filter/FlateDecode"
	}
	objs := []string{
		"<</Type/Catalog/Pages 2 0 R>>",
		"<</Type/Pages/Kids[3 0 R]/Count 1>>",
		fmt.Sprintf("<</Type/Page/Parent 2 0 R/MediaBox[0 0 %s %s]/Resources<<>>/Contents 4 0 R>>",
			num(pageW, 2), num(pageH, 2)),
		fmt.Sprintf("<</Length %d%s>>stream\n%s\nendstream", len(stream), filter, stream),
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	offs := make([]int, len(objs))
	for i, o := range objs {
		offs[i] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, o := range offs {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<</Size %d/Root 1 0 R>>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)
	return b.Bytes()
}

// PNG cover size, px: 2:3 like the page, the size e-book stores ask for.
const pngW, pngH = 1600, 2400

// writePNG renders the PDF at pdfPath with Ghostscript and writes it to
// pngPath in four grays.
func writePNG(pdfPath, pngPath string) error {
	gs := exec.Command("gs", "-q", "-dNOPAUSE", "-dBATCH", "-dSAFER",
		"-sDEVICE=pgmraw", fmt.Sprintf("-g%dx%d", pngW, pngH), "-dPDFFitPage",
		"-dGraphicsAlphaBits=4", "-sOutputFile=-", pdfPath)
	gs.Stderr = os.Stderr
	out, err := gs.Output()
	if err != nil {
		return fmt.Errorf("gs: %w", err)
	}
	// PGM: "P5", width, height, maxval, one whitespace byte, then pixels.
	// Ghostscript puts a "#" comment line after "P5".
	r := bufio.NewReader(bytes.NewReader(out))
	var head []string
	for len(head) < 4 {
		line, err := r.ReadString('\n')
		if err != nil {
			return fmt.Errorf("gs: short PGM header: %w", err)
		}
		if !strings.HasPrefix(line, "#") {
			head = append(head, strings.Fields(line)...)
		}
	}
	w, _ := strconv.Atoi(head[1])
	h, _ := strconv.Atoi(head[2])
	if head[0] != "P5" || head[3] != "255" || w == 0 || h == 0 {
		return fmt.Errorf("gs: unexpected PGM header %q", head)
	}
	pix := make([]byte, w*h)
	if _, err := io.ReadFull(r, pix); err != nil {
		return fmt.Errorf("gs: short PGM: %w", err)
	}

	img := image.NewPaletted(image.Rect(0, 0, w, h), color.Palette{
		color.Gray{0}, color.Gray{85}, color.Gray{170}, color.Gray{255},
	})
	for i, v := range pix {
		img.Pix[i] = uint8((int(v)*3 + 127) / 255) // nearest of the four
	}
	var b bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&b, img); err != nil {
		return err
	}
	if err := os.WriteFile(pngPath, b.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Printf("%s: %dx%d, %d bytes\n", pngPath, w, h, b.Len())
	return nil
}

// findFont returns name if it is a path, else asks kpsewhich for it.
func findFont(name string) (string, error) {
	if strings.ContainsRune(name, os.PathSeparator) {
		return name, nil
	}
	out, err := exec.Command("kpsewhich", name).Output()
	p := strings.TrimSpace(string(out))
	if err != nil || p == "" {
		return "", fmt.Errorf("kpsewhich cannot find %s", name)
	}
	return p, nil
}

func main() {
	fontName := flag.String("font", "texgyretermes-regular.otf", "title font: a path, or a file name kpsewhich can find")
	plain := flag.Bool("plain", false, "leave the content stream uncompressed, to read or edit by hand")
	outPath := flag.String("o", "cover.pdf", "output file")
	pngPath := flag.String("png", "cover.png", "PNG cover for the EPUB, rendered from the PDF with gs; empty skips it")
	flag.Parse()

	path, err := findFont(*fontName)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	face, err := font.ParseTTF(f)
	if err != nil {
		log.Fatalf("%s: %v", path, err)
	}

	var c content
	c.frame()
	c.arches()
	c.text(face, textLines)
	doc := document([]byte(c.String()), !*plain)
	if err := os.WriteFile(*outPath, doc, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %d bytes\n", *outPath, len(doc))

	if *pngPath != "" {
		if err := writePNG(*outPath, *pngPath); err != nil {
			log.Fatalf("%s: %v (-png '' skips the PNG)", *pngPath, err)
		}
	}
}
