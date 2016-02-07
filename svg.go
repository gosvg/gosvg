package gosvg

import (
	"fmt"
	"io"
	"strings"
)

type renderable interface {
	render(e io.Writer) error
}

func floatAttr(name string, val float64) string {
	return fmt.Sprintf("%s=\"%g\"", name, val)
}

func boolAttr(name string, val bool) string {
	var valStr string
	if val {
		valStr = "true"
	} else {
		valStr = "false"
	}

	return fmt.Sprintf("%s=\"%s\"", name, valStr)
}

func stringAttr(name string, val string) string {
	return fmt.Sprintf("%s=\"%s\"", name, val)
}

type valueMap map[string]string

func (m *valueMap) Get(k string) string {
	if *m == nil {
		*m = make(valueMap)
	}
	return (*m)[k]
}

func (m *valueMap) Set(k string, v string) {
	if *m == nil {
		*m = make(valueMap)
	}
	(*m)[k] = v
}

func (m *valueMap) Unset(k string) {
	if *m == nil {
		*m = make(valueMap)
	}
	delete(*m, k)
}

// Style represents the style attribute for any stylable SVG element.
type Style struct {
	valueMap
}

func (s Style) attrString() string {
	if s.valueMap == nil {
		return ""
	}

	var outs []string

	for k, v := range s.valueMap {
		valStr := fmt.Sprintf("%s:%s", k, v)
		outs = append(outs, valStr)
	}

	out := strings.Join(outs, ";")

	out = fmt.Sprintf("style=\"%s\"", out)

	return out
}

// Transform represents a series of transforms applied to an SVG element.
type Transform struct {
	transforms []string
}

// Matrix adds a matrix SVG transform.
func (t *Transform) Matrix(a, b, c, d, e, f float64) {
	out := fmt.Sprintf("matrix(%g,%g,%g,%g,%g,%g)", a, b, c, d, e, f)

	t.transforms = append(t.transforms, out)
}

// Translate adds a translate SVG transform.
func (t *Transform) Translate(tx, ty float64) {
	out := fmt.Sprintf("translate(%g,%g)", tx, ty)

	t.transforms = append(t.transforms, out)
}

// Scale adds a scale SVG transform.
func (t *Transform) Scale(sx, sy float64) {
	out := fmt.Sprintf("scale(%g,%g)", sx, sy)

	t.transforms = append(t.transforms, out)
}

// Rotate adds a rotate SVG transform of angle degrees around the point (cx, cy).
func (t *Transform) Rotate(angle, cx, cy float64) {
	out := fmt.Sprintf("rotate(%g,%g,%g)", angle, cx, cy)

	t.transforms = append(t.transforms, out)
}

// SkewX adds a skew SVG transform along the x axis.
func (t *Transform) SkewX(angle float64) {
	out := fmt.Sprintf("skewX(%g)", angle)

	t.transforms = append(t.transforms, out)
}

// SkewY adds a skew SVG transform along the y axis.
func (t *Transform) SkewY(angle float64) {
	out := fmt.Sprintf("skewY(%g)", angle)

	t.transforms = append(t.transforms, out)
}

func (t *Transform) attrString() string {
	out := strings.Join(t.transforms, " ")

	out = fmt.Sprintf("transform=\"%s\"", out)

	return out
}

// ViewBox represents the viewBox attribute for an <svg> element.
type ViewBox struct {
	minX   float64
	minY   float64
	width  float64
	height float64
	isSet  bool
}

// Set sets the viewbox to the given values.
func (v *ViewBox) Set(minX, minY, width, height float64) {
	v.minX = minX
	v.minY = minY
	v.width = width
	v.height = height
	v.isSet = true
}

func (v ViewBox) attrString() string {
	if !v.isSet {
		return ""
	}

	out := fmt.Sprintf("%g %g %g %g", v.minX, v.minX, v.width, v.height)

	out = fmt.Sprintf("viewBox=\"%s\"", out)

	return out
}

// BaseAttrs represents the basic attributes for all visual and container elements.
type BaseAttrs struct {
	Style                     Style
	ExternalResourcesRequired bool
	Class                     string
}

func (a *BaseAttrs) attrStrings() []string {
	extString := ""
	if a.ExternalResourcesRequired {
		extString = boolAttr("externalResourcesRequired", a.ExternalResourcesRequired)
	}

	out := []string{
		a.Style.attrString(),
		extString,
		a.Class,
	}

	return out
}

// ShapeAttrs represents attributes specific to shape elements.
type ShapeAttrs struct {
	BaseAttrs
	Transform Transform
}

func (a *ShapeAttrs) attrStrings() []string {
	out := append(a.BaseAttrs.attrStrings(), a.Transform.attrString())

	return out
}

type container struct {
	name     string
	contents []renderable
}

func (c *container) render(w io.Writer, attrs []string) error {
	attrString := strings.Join(attrs, " ")

	opening := fmt.Sprintf("<%s %s>", c.name, attrString)
	if _, err := w.Write([]byte(opening)); err != nil {
		return err
	}
	for _, r := range c.contents {
		if err := r.render(w); err != nil {
			return err
		}
	}
	closing := fmt.Sprintf("</%s>", c.name)
	if _, err := w.Write([]byte(closing)); err != nil {
		return err
	}

	return nil
}

// SVG generates a new SVG embedded in the given container.
func (c *container) SVG(x, y, w, h float64) *SVG {
	r := &SVG{Width: w, Height: h, X: x, Y: y, container: container{name: "svg"}}

	c.contents = append(c.contents, r)

	return r
}

// Circle generates a new circle in the given container.
func (c *container) Circle(cx, cy, r float64) *Circle {
	d := &Circle{Cx: cx, Cy: cy, R: r}
	c.contents = append(c.contents, d)

	return d
}

// Ellipse generates a new ellipse in the given container.
func (c *container) Ellipse(cx, cy, rx, ry float64) *Ellipse {
	e := &Ellipse{Cx: cx, Cy: cy, Rx: rx, Ry: ry}
	c.contents = append(c.contents, e)

	return e
}

// Rect generates a new rect in the given container.
func (c *container) Rect(x, y, w, h float64) *Rect {
	r := &Rect{X: x, Y: y, Width: w, Height: h}
	c.contents = append(c.contents, r)

	return r
}

// Polygon generates a new polygon in the given container.
func (c *container) Polygon(pts ...Point) *Polygon {
	p := &Polygon{Points: pts}
	c.contents = append(c.contents, p)

	return p
}

// Polyline generates a new polyline in the given container.
func (c *container) Polyline(pts ...Point) *Polyline {
	p := &Polyline{Points: pts}
	c.contents = append(c.contents, p)

	return p
}

// Line generates a new line in the given container.
func (c *container) Line(x1, y1, x2, y2 float64) *Line {
	n := &Line{X1: x1, Y1: y1, X2: x2, Y2: y2}
	c.contents = append(c.contents, n)

	return n
}

// Path generates a new in the given container.
func (c *container) Path() *Path {
	p := &Path{}
	c.contents = append(c.contents, p)

	return p
}

// Group generates a new group within the given container.
func (c *SVG) Group() *Group {
	g := &Group{container: container{name: "g"}}
	c.contents = append(c.contents, g)

	return g
}

// SVG represents a complete SVG fragment (the svg element).
type SVG struct {
	BaseAttrs
	container
	ViewBox ViewBox
	Width   float64
	Height  float64
	X       float64
	Y       float64
}

// NewSVG creates a new SVG with the given width and height.
func NewSVG(width, height float64) *SVG {
	return &SVG{
		Width:     width,
		Height:    height,
		container: container{name: "svg"},
	}
}

func (s *SVG) attrStrings() []string {
	base := s.BaseAttrs.attrStrings()
	viewBox := s.ViewBox.attrString()
	width := floatAttr("width", s.Width)
	height := floatAttr("height", s.Height)
	x := floatAttr("x", s.X)
	y := floatAttr("y", s.Y)
	xmlns := stringAttr("xmlns", "http://www.w3.org/2000/svg")

	return append(base, viewBox, width, height, x, y, xmlns)
}

// Render renders the SVG as a complete document with XML version tag.
func (s *SVG) Render(w io.Writer) error {
	top := `<?xml version="1.0"?>`
	if _, err := w.Write([]byte(top)); err != nil {
		return err
	}

	return s.render(w)
}

// RenderFragment renders the SVG as a fragment with no XML version tag.
func (s *SVG) RenderFragment(w io.Writer) error {
	return s.render(w)
}

func (s *SVG) render(w io.Writer) error {
	attrStrings := s.attrStrings()
	return s.container.render(w, attrStrings)
}

// Group represents an SVG group (the g element).
type Group struct {
	ShapeAttrs
	container
}

func (g *Group) render(w io.Writer) error {
	attrStrings := g.attrStrings()
	return g.container.render(w, attrStrings)
}

// Circle represents a circle (the cirlc element).
type Circle struct {
	ShapeAttrs
	Cx float64 `xml:"cx,attr"`
	Cy float64 `xml:"cy,attr"`
	R  float64 `xml:"r,attr"`
}

func (c *Circle) attrStrings() []string {
	cx := floatAttr("cx", c.Cx)
	cy := floatAttr("cy", c.Cy)
	r := floatAttr("r", c.R)

	attrs := c.ShapeAttrs.attrStrings()
	attrs = append(attrs, cx, cy, r)

	return attrs
}

func (c *Circle) render(w io.Writer) error {
	attrStrings := c.attrStrings()
	attrString := strings.Join(attrStrings, " ")

	out := fmt.Sprintf("<circle %s/>", attrString)
	_, err := w.Write([]byte(out))

	return err
}

// Ellipse represents an ellipse (the ellipse element).
type Ellipse struct {
	ShapeAttrs
	Cx float64
	Cy float64
	Rx float64
	Ry float64
}

func (e *Ellipse) attrStrings() []string {
	cx := floatAttr("cx", e.Cx)
	cy := floatAttr("cy", e.Cy)
	rx := floatAttr("rx", e.Rx)
	ry := floatAttr("ry", e.Ry)

	attrs := e.ShapeAttrs.attrStrings()
	attrs = append(attrs, cx, cy, rx, ry)

	return attrs
}

func (e *Ellipse) render(w io.Writer) error {
	attrStrings := e.attrStrings()
	attrString := strings.Join(attrStrings, " ")

	out := fmt.Sprintf("<ellipse %s/>", attrString)
	_, err := w.Write([]byte(out))

	return err
}

// Rect represents a rectangle (the rect element).
type Rect struct {
	ShapeAttrs
	Width  float64
	Height float64
	X      float64
	Y      float64
}

func (r *Rect) attrStrings() []string {
	w := floatAttr("width", r.Width)
	h := floatAttr("height", r.Height)
	x := floatAttr("x", r.X)
	y := floatAttr("y", r.Y)

	return append(r.ShapeAttrs.attrStrings(), w, h, x, y)
}

func (r *Rect) render(w io.Writer) error {
	attrStrings := r.attrStrings()
	attrString := strings.Join(attrStrings, " ")

	out := fmt.Sprintf("<rect %s/>", attrString)
	_, err := w.Write([]byte(out))

	return err
}

// Point represents a Cartesian point.
type Point struct {
	X float64
	Y float64
}

func (p Point) coordString() string {
	return fmt.Sprintf("%g,%g", p.X, p.Y)
}

// Polygon represents a closed, filled polygon (the polygon element).
type Polygon struct {
	ShapeAttrs
	Points []Point
}

func (p *Polygon) attrStrings() []string {
	var pointStrs []string
	for _, pt := range p.Points {
		pointStrs = append(pointStrs, pt.coordString())
	}
	pointStr := strings.Join(pointStrs, " ")

	attrs := p.ShapeAttrs.attrStrings()
	pointAttr := stringAttr("points", pointStr)

	return append(attrs, pointAttr)
}

func (p *Polygon) render(w io.Writer) error {
	attrStrings := p.attrStrings()
	attrString := strings.Join(attrStrings, " ")

	tag := "polygon"

	out := fmt.Sprintf("<%s %s/>", tag, attrString)
	_, err := w.Write([]byte(out))

	return err
}

// Polyline represents an open polygon rendered as line (the polyline element).
type Polyline struct {
	ShapeAttrs
	Points []Point
}

func (p *Polyline) attrStrings() []string {
	var pointStrs []string
	for _, pt := range p.Points {
		pointStrs = append(pointStrs, pt.coordString())
	}
	pointStr := strings.Join(pointStrs, " ")

	attrs := p.ShapeAttrs.attrStrings()
	pointAttr := stringAttr("points", pointStr)

	return append(attrs, pointAttr)
}

func (p *Polyline) render(w io.Writer) error {
	attrStrings := p.attrStrings()
	attrString := strings.Join(attrStrings, " ")

	tag := "polyline"

	out := fmt.Sprintf("<%s %s/>", tag, attrString)
	_, err := w.Write([]byte(out))

	return err
}

// Line represents a straight line (the line element).
type Line struct {
	ShapeAttrs
	X1 float64
	X2 float64
	Y1 float64
	Y2 float64
}

func (n *Line) attrStrings() []string {
	x1 := floatAttr("x1", n.X1)
	y1 := floatAttr("y1", n.Y1)
	x2 := floatAttr("x2", n.X2)
	y2 := floatAttr("y2", n.Y2)

	return append(n.ShapeAttrs.attrStrings(), x1, y1, x2, y2)
}

func (n *Line) render(w io.Writer) error {
	attrStrings := n.attrStrings()
	attrString := strings.Join(attrStrings, " ")

	tag := "line"

	out := fmt.Sprintf("<%s %s/>", tag, attrString)
	_, err := w.Write([]byte(out))

	return err
}

// Path command declarations

type cmdBody interface {
	strings() []string
}

type cmd struct {
	name string
	body cmdBody
}

type mCmd struct {
	pts []Point
}

func (m mCmd) strings() []string {
	var out []string

	for _, p := range m.pts {
		out = append(out, fmt.Sprintf("%g", p.X), fmt.Sprintf("%g", p.Y))
	}

	return out
}

type zCmd struct {
}

func (z zCmd) strings() []string {
	return nil
}

// Don't want to use just lowercase L here
type elCmd struct {
	pts   []Point
	isAbs bool
}

func (e elCmd) strings() []string {
	var out []string

	for _, p := range e.pts {
		out = append(out, fmt.Sprintf("%g", p.X), fmt.Sprintf("%g", p.Y))
	}

	return out
}

type hCmd struct {
	xs    []float64
	isAbs bool
}

func (h hCmd) strings() []string {
	var out []string

	for _, x := range h.xs {
		out = append(out, fmt.Sprintf("%g", x))
	}

	return out
}

type vCmd struct {
	ys    []float64
	isAbs bool
}

func (v vCmd) strings() []string {
	var out []string

	for _, y := range v.ys {
		out = append(out, fmt.Sprintf("%g", y))
	}

	return out
}

// CCurve represents a cubic Bezier curve.
type CCurve struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
	X  float64
	Y  float64
}

// SCurve represents a shorthand cubic Bezier curve.
type SCurve struct {
	X2 float64
	Y2 float64
	X  float64
	Y  float64
}

// QCurve represents a quadratic Bezier curve.
type QCurve struct {
	X1 float64
	Y1 float64
	X  float64
	Y  float64
}

type cCmd struct {
	cvs   []CCurve
	isAbs bool
}

func (c cCmd) strings() []string {
	var out []string

	for _, cv := range c.cvs {
		out = append(out,
			fmt.Sprintf("%g", cv.X1),
			fmt.Sprintf("%g", cv.Y1),
			fmt.Sprintf("%g", cv.X2),
			fmt.Sprintf("%g", cv.Y2),
			fmt.Sprintf("%g", cv.X),
			fmt.Sprintf("%g", cv.Y))
	}

	return out
}

type sCmd struct {
	cvs   []SCurve
	isAbs bool
}

func (c sCmd) strings() []string {
	var out []string

	for _, cv := range c.cvs {
		out = append(out,
			fmt.Sprintf("%g", cv.X2),
			fmt.Sprintf("%g", cv.Y2),
			fmt.Sprintf("%g", cv.X),
			fmt.Sprintf("%g", cv.Y))
	}

	return out
}

type qCmd struct {
	cvs   []QCurve
	isAbs bool
}

func (c qCmd) strings() []string {
	var out []string

	for _, cv := range c.cvs {
		out = append(out,
			fmt.Sprintf("%g", cv.X1),
			fmt.Sprintf("%g", cv.Y1),
			fmt.Sprintf("%g", cv.X),
			fmt.Sprintf("%g", cv.Y))
	}

	return out
}

type tCmd struct {
	pts   []Point
	isAbs bool
}

func (c tCmd) strings() []string {
	var out []string

	for _, pt := range c.pts {
		out = append(out,
			fmt.Sprintf("%g", pt.X),
			fmt.Sprintf("%g", pt.Y))
	}

	return out
}

// Path represents a path through a given coordinate system (the path element).
type Path struct {
	ShapeAttrs
	d          []cmd
	PathLength float64
}

func (p *Path) addCmd(name string, body cmdBody) *Path {
	p.d = append(p.d, cmd{name: name, body: body})

	return p
}

// Ma appends an absolute moveto command to the path.
func (p *Path) Ma(pts ...Point) *Path {
	return p.addCmd("M", mCmd{pts: pts})
}

// Mr appends a relative moveto command to the path.
func (p *Path) Mr(pts ...Point) *Path {
	return p.addCmd("m", mCmd{pts: pts})
}

// Z adds a closepath command to the path.
func (p *Path) Z() *Path {
	return p.addCmd("z", zCmd{})
}

// La appends an absolute lineto command to the path.
func (p *Path) La(pts ...Point) *Path {
	return p.addCmd("L", elCmd{pts: pts})
}

// Lr appends a relative lineto command to the path.
func (p *Path) Lr(pts ...Point) *Path {
	return p.addCmd("l", elCmd{pts: pts})
}

// Ha appends an absolute horizontal lineto command to the path.
func (p *Path) Ha(xs ...float64) *Path {
	return p.addCmd("H", hCmd{xs: xs})
}

// Hr appends a relative horizontal lineto command to the path.
func (p *Path) Hr(xs ...float64) *Path {
	return p.addCmd("h", hCmd{xs: xs})
}

// Va appends an absolute vertical lineto command to the path.
func (p *Path) Va(ys ...float64) *Path {
	return p.addCmd("V", vCmd{ys: ys})
}

// Vr appends a relative vertical lineto command to the path.
func (p *Path) Vr(ys ...float64) *Path {
	return p.addCmd("v", vCmd{ys: ys})
}

// Ca appends an absolute curveto command to the path.
func (p *Path) Ca(cvs ...CCurve) *Path {
	return p.addCmd("C", cCmd{cvs: cvs})
}

// Cr appends a relative curveto command to the path.
func (p *Path) Cr(cvs ...CCurve) *Path {
	return p.addCmd("c", cCmd{cvs: cvs})
}

// Sa appends an absolute shorthand/smooth curveto command to the path.
func (p *Path) Sa(cvs ...SCurve) *Path {
	return p.addCmd("S", sCmd{cvs: cvs})
}

// Sr appends a relative shorthand/smooth curveto command to the path.
func (p *Path) Sr(cvs ...SCurve) *Path {
	return p.addCmd("s", sCmd{cvs: cvs})
}

// Qa appends an absolute quadratic Bezier curveto command to the path.
func (p *Path) Qa(cvs ...QCurve) *Path {
	return p.addCmd("Q", qCmd{cvs: cvs})
}

// Qr appends a relative quadratic Bezier curveto command to the path.
func (p *Path) Qr(cvs ...QCurve) *Path {
	return p.addCmd("q", qCmd{cvs: cvs})
}

// Ta appends an absolute shorthand/smooth quadratic Bezier curveto command to the path.
func (p *Path) Ta(cvs ...Point) *Path {
	return p.addCmd("T", tCmd{pts: cvs})
}

// Tr appends a relative shorthand/smooth quadratic Bezier curveto command to the path.
func (p *Path) Tr(cvs ...Point) *Path {
	return p.addCmd("t", tCmd{pts: cvs})
}

func (p *Path) pathStr() string {
	var outs []string
	accumLen := 0

	for _, cmd := range p.d {
		accumLen++
		if accumLen > 255 {
			outs = append(outs, "\n", cmd.name)
			accumLen = 0
		} else {
			outs = append(outs, " ", cmd.name)
		}

		bodyStrs := cmd.body.strings()
		for _, s := range bodyStrs {
			accumLen += len(s)
			if accumLen > 255 {
				outs = append(outs, "\n", s)
				accumLen = 0
			} else {
				outs = append(outs, " ", s)
			}
		}
	}

	if len(outs) > 0 {
		outs = outs[1:]
	}

	out := strings.Join(outs, "")

	return out
}

func (p *Path) attrStrings() []string {
	pathStr := p.pathStr()
	pathAttr := stringAttr("d", pathStr)

	attrs := p.ShapeAttrs.attrStrings()

	return append(attrs, pathAttr)
}

func (p *Path) render(w io.Writer) error {
	attrStrings := p.attrStrings()
	attrString := strings.Join(attrStrings, " ")

	tag := "path"

	out := fmt.Sprintf("<%s %s/>", tag, attrString)
	_, err := w.Write([]byte(out))

	return err
}
