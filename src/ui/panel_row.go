package ui

import "kaiju/matrix"

type rowBuilder struct {
	elements        []*UI
	maxMarginTop    float32
	maxMarginBottom float32
	x               float32
	height          float32
}

func (r *rowBuilder) addElement(areaWidth float32, e *UI) bool {
	eSize := e.layout.PixelSize()
	w := eSize.Width()
	if len(r.elements) > 0 && r.x+w > areaWidth {
		return false
	}
	r.elements = append(r.elements, e)
	r.maxMarginTop = max(r.maxMarginTop, e.layout.margin.Y())
	r.maxMarginBottom = max(r.maxMarginBottom, e.layout.margin.W())
	if e.layout.screenAnchor.IsStretch() {
		if e.layout.screenAnchor.IsTop() {
			r.maxMarginTop += e.layout.top
		} else if e.layout.screenAnchor.IsBottom() {
			r.maxMarginBottom += e.layout.bottom
		}
	}
	r.x += w
	r.height = max(r.height, eSize.Height())
	return true
}

func (r *rowBuilder) Width() float32 { return r.x }

func (r *rowBuilder) Height() float32 {
	return r.height + r.maxMarginTop + r.maxMarginBottom
}

func (r *rowBuilder) setElements(offsetX, offsetY float32) {
	for i := range r.elements {
		e := r.elements[i]
		layout := &e.layout
		x := offsetX
		y := offsetY
		switch layout.positioning {
		case PositioningAbsolute:
			fallthrough
		case PositioningRelative:
			if layout.screenAnchor.IsLeft() {
				x += layout.innerOffset.Left()
			} else if layout.screenAnchor.IsRight() {
				x += layout.innerOffset.Right()
			}
			if layout.screenAnchor.IsTop() {
				y += layout.innerOffset.Top()
			} else if layout.screenAnchor.IsBottom() {
				y += layout.innerOffset.Bottom()
			}
		}
		x += layout.margin.X()
		y += r.maxMarginTop
		layout.rowLayoutOffset = matrix.NewVec2(x, y)
		offsetX += layout.PixelSize().Width() + layout.margin.X() + layout.margin.Z()
	}
}
