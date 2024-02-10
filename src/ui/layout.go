package ui

import (
	"kaiju/matrix"
)

type Anchor int32
type Positioning = int

const (
	AnchorTopLeft = Anchor(1 + iota)
	AnchorTopCenter
	AnchorTopRight
	AnchorLeft
	AnchorCenter
	AnchorRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
	AnchorStretchLeft
	AnchorStretchTop
	AnchorStretchRight
	AnchorStretchBottom
	AnchorStretchCenter
)

const (
	PositioningStatic = Positioning(iota)
	PositioningAbsolute
	PositioningFixed
	PositioningRelative
	PositioningSticky
)

func (a Anchor) ConvertToTop() Anchor {
	switch a {
	case AnchorBottomLeft:
		return AnchorTopLeft
	case AnchorBottomCenter:
		return AnchorTopCenter
	case AnchorBottomRight:
		return AnchorTopRight
	case AnchorStretchTop:
		return AnchorStretchBottom
	default:
		return a
	}
}

func (a Anchor) ConvertToBottom() Anchor {
	switch a {
	case AnchorTopLeft:
		return AnchorBottomLeft
	case AnchorTopCenter:
		return AnchorBottomCenter
	case AnchorTopRight:
		return AnchorBottomRight
	case AnchorStretchBottom:
		return AnchorStretchTop
	default:
		return a
	}
}

func (a Anchor) ConvertToLeft() Anchor {
	switch a {
	case AnchorTopRight:
		return AnchorTopLeft
	case AnchorCenter:
		return AnchorLeft
	case AnchorBottomRight:
		return AnchorBottomLeft
	case AnchorStretchRight:
		return AnchorStretchLeft
	default:
		return a
	}
}

func (a Anchor) ConvertToRight() Anchor {
	switch a {
	case AnchorTopLeft:
		return AnchorTopRight
	case AnchorLeft:
		return AnchorRight
	case AnchorBottomLeft:
		return AnchorBottomRight
	case AnchorStretchLeft:
		return AnchorStretchRight
	default:
		return a
	}
}

func (a Anchor) ConvertToCenter() Anchor {
	switch a {
	case AnchorTopLeft:
		fallthrough
	case AnchorTopRight:
		return AnchorTopCenter
	case AnchorLeft:
		fallthrough
	case AnchorRight:
		return AnchorCenter
	case AnchorBottomLeft:
		fallthrough
	case AnchorBottomRight:
		return AnchorBottomCenter
	default:
		return a
	}
}

func (a Anchor) IsLeft() bool {
	return a == AnchorLeft || a == AnchorTopLeft || a == AnchorBottomLeft || a == AnchorStretchLeft
}

func (a Anchor) IsRight() bool {
	return a == AnchorRight || a == AnchorTopRight || a == AnchorBottomRight || a == AnchorStretchRight
}

func (a Anchor) IsTop() bool {
	return a == AnchorTopLeft || a == AnchorTopCenter || a == AnchorTopRight || a == AnchorStretchTop
}

func (a Anchor) IsBottom() bool {
	return a == AnchorBottomLeft || a == AnchorBottomCenter || a == AnchorBottomRight || a == AnchorStretchBottom
}

type Layout struct {
	offset           matrix.Vec2
	rowLayoutOffset  matrix.Vec2
	innerOffset      matrix.Vec4
	localInnerOffset matrix.Vec4
	left             float32
	top              float32
	right            float32
	bottom           float32
	z                float32
	anchor           matrix.Vec2
	ui               UI
	screenAnchor     Anchor
	layoutFunction   func(layout *Layout)
	anchorFunction   func(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4
	worldScalar      matrix.Vec2
	border           matrix.Vec4
	padding          matrix.Vec4
	margin           matrix.Vec4
	inset            matrix.Vec4
	positioning      Positioning
	functions        LayoutFunctions
	runningFuncs     bool
}

func NewLayout(ui UI) Layout {
	return Layout{
		ui:          ui,
		anchor:      matrix.Vec2{0.0, 0.0},
		positioning: PositioningAbsolute,
	}
}

func (l *Layout) AddFunction(fn func(layout *Layout)) LayoutFuncId {
	return l.functions.Add(fn)
}

func (l *Layout) RemoveFunction(id LayoutFuncId) {
	l.functions.Remove(id)
}

func (l *Layout) ClearFunctions() {
	l.functions.Clear()
}

func (l *Layout) PixelSize() matrix.Vec2 {
	return l.ui.Entity().Transform.WorldScale().AsVec2()
}

func al(edges matrix.Vec4, w float32, size matrix.Vec2) float32 {
	return -w*0.5 + size.X()*0.5 + edges.Left()
}

func ar(edges matrix.Vec4, w float32, size matrix.Vec2) float32 {
	return w*0.5 - size.X()*0.5 - edges.Right()
}

func at(edges matrix.Vec4, h float32, size matrix.Vec2) float32 {
	return h*0.5 - size.Y()*0.5 - edges.Top()
}

func ab(edges matrix.Vec4, h float32, size matrix.Vec2) float32 {
	return -h*0.5 + size.Y()*0.5 + edges.Bottom()
}

func (l *Layout) CalcOffset() matrix.Vec2 {
	return l.rowLayoutOffset.Add(l.offset)
}

func anchorTopLeft(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{al(edges, w, size) + inner.Left(), at(edges, h, size) + inner.Top(), 0.0, 0.0}
}

func anchorTopCenter(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{self.CalcOffset().X(), at(edges, h, size) + inner.Top(), 0.0, 0.0}
}

func anchorTopRight(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{ar(edges, w, size) + inner.Right(), at(edges, h, size) + inner.Top(), 0.0, 0.0}
}

func anchorLeft(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{al(edges, w, size) + inner.Left(), self.CalcOffset().Y(), 0.0, 0.0}
}

func anchorCenter(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	return matrix.Vec4{self.CalcOffset().X(), self.CalcOffset().Y(), 0.0, 0.0}
}

func anchorRight(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{ar(edges, w, size) + inner.Right(), self.CalcOffset().Y(), 0.0, 0.0}
}

func anchorBottomLeft(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{al(edges, w, size) + inner.Left(), ab(edges, h, size) + inner.Bottom(), 0.0, 0.0}
}

func anchorBottomCenter(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{self.CalcOffset().X(), ab(edges, h, size) + inner.Bottom(), 0.0, 0.0}
}

func anchorBottomRight(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{ar(edges, w, size) + inner.Right(), ab(edges, h, size) + inner.Bottom(), 0.0, 0.0}
}

func layoutFloating(self *Layout) {
	t := &self.ui.Entity().Transform
	s := self.PixelSize()
	bounds := self.bounds()
	pos := self.anchorFunction(self, bounds.X(), bounds.Y(), s)
	pos.SetZ(self.z + 0.01)
	t.SetPosition(pos.AsVec3())
}

func anchorStretchLeft(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	xSize := self.right
	ySize := h - (self.top + self.bottom)
	xMid := xSize*0.5 + self.left
	yMid := self.bottom + ySize*0.5
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchTop(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	xSize := w - (self.left + self.right)
	ySize := self.bottom
	xMid := self.left + xSize*0.5
	yMid := h - (ySize * 0.5) - self.top
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchRight(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	xSize := self.left
	ySize := h - (self.top + self.bottom)
	xMid := w - (xSize * 0.5) - self.right
	yMid := self.bottom + ySize*0.5
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchBottom(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	xSize := w - (self.left + self.right)
	ySize := self.top
	xMid := self.left + xSize*0.5
	yMid := ySize*0.5 + self.bottom
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchCenter(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	xSize := w - (self.left + self.right)
	ySize := h - (self.top + self.bottom)
	xMid := self.left + xSize*0.5
	yMid := self.bottom + ySize*0.5
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func layoutStretch(self *Layout) {
	t := &self.ui.Entity().Transform
	bounds := self.bounds()
	res := self.anchorFunction(self, bounds.X(), bounds.Y(), matrix.Vec2Zero())
	x := res.X() + self.CalcOffset().X()
	y := res.Y() + self.CalcOffset().Y()
	xSize := res.Z()
	ySize := res.W()
	scale := matrix.Vec3{xSize * self.worldScalar.X(), ySize * self.worldScalar.Y(), 1}
	scale[matrix.Vx] -= (self.inset.X() + self.inset.Z()) / bounds.X()
	scale[matrix.Vy] -= (self.inset.Y() + self.inset.W()) / bounds.Y()
	self.ui.Entity().ScaleWithoutChildren(scale)
	pos := matrix.Vec3{
		x + bounds.X() + (self.inset.X()-self.inset.Z())*0.5,
		y + bounds.Y() + (self.inset.W()-self.inset.Y())*0.5,
		t.Position().Z(),
	}
	t.SetPosition(pos)
}

func (l *Layout) prepare() {
	if l.runningFuncs {
		return
	}
	l.runningFuncs = true
	l.functions.Execute(l)
	l.runningFuncs = false
}

func (l *Layout) bounds() matrix.Vec2 {
	if l.ui.Entity().IsRoot() {
		return matrix.Vec2{
			matrix.Float(l.ui.Host().Window.Width()),
			matrix.Float(l.ui.Host().Window.Height()),
		}
	} else {
		parent := l.ui.Entity().Parent
		s := parent.Transform.WorldScale()
		return matrix.Vec2{s.X(), s.Y()}
	}
}

func (l *Layout) initialize(ui UI, anchor Anchor) {
	l.anchor = matrix.Vec2Zero()
	l.ui = ui
	l.AnchorTo(anchor)
	l.prepare()
	l.update()
}

func (l *Layout) SetOffset(x, y float32) {
	if matrix.Vec2Approx(l.offset, matrix.Vec2{x, y}) {
		return
	}
	l.offset.SetX(x)
	l.offset.SetY(y)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetInnerOffset(left, top, right, bottom float32) {
	if matrix.Vec4Approx(l.innerOffset, matrix.Vec4{left, top, right, bottom}) {
		return
	}
	l.innerOffset = matrix.Vec4{left, top, right, bottom}
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetInnerOffsetLeft(offset float32) {
	if matrix.Approx(l.innerOffset.X(), offset) {
		return
	}
	l.innerOffset.SetX(offset)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetInnerOffsetTop(offset float32) {
	if matrix.Approx(l.innerOffset[matrix.Vy], offset) {
		return
	}
	l.innerOffset.SetY(offset)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetInnerOffsetRight(offset float32) {
	if matrix.Approx(l.innerOffset.Right(), offset) {
		return
	}
	l.innerOffset.SetRight(offset)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetInnerOffsetBottom(offset float32) {
	if matrix.Approx(l.innerOffset.Bottom(), offset) {
		return
	}
	l.innerOffset.SetBottom(offset)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) LocalInnerOffset() matrix.Vec4 {
	return l.localInnerOffset
}

func (l *Layout) SetLocalInnerOffset(left, top, right, bottom float32) {
	if matrix.Vec4Approx(l.localInnerOffset, matrix.Vec4{left, top, right, bottom}) {
		return
	}
	l.localInnerOffset = matrix.Vec4{left, top, right, bottom}
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) InnerOffset() matrix.Vec4 {
	return matrix.Vec4{
		l.localInnerOffset.Left() + l.innerOffset.Left(),
		l.localInnerOffset.Top() + l.innerOffset.Top(),
		l.localInnerOffset.Right() + l.innerOffset.Right(),
		l.localInnerOffset.Bottom() + l.innerOffset.Bottom(),
	}
}

func (l *Layout) SetStretch(left, top, right, bottom float32) {
	l.left = left
	l.top = top
	l.right = right
	l.bottom = bottom
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) SetStretchRatio(leftRatio, topRatio, rightRatio, bottomRatio float32) {
	bounds := l.bounds()
	w := bounds.X()
	h := bounds.Y()
	l.left = w * leftRatio
	l.top = h * topRatio
	l.right = w * rightRatio
	l.bottom = h * bottomRatio
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) Scale(width, height float32) bool {
	width += l.padding.X() + l.padding.Z()
	height += l.padding.Y() + l.padding.W()
	ps := l.PixelSize()
	if matrix.Vec2ApproxTo(ps, matrix.Vec2{width, height}, 0.001) {
		return false
	}
	if width == 0 || height == 0 {
		return false
	}
	size := matrix.Vec3{width, height, 1.0}
	if l.ui.Entity().Parent != nil {
		size.DivideAssign(l.ui.Entity().Parent.Transform.WorldScale())
	}
	l.ui.Entity().ScaleWithoutChildren(size)
	l.ui.layoutChanged(DirtyTypeResize)
	return true
}

func (l *Layout) ScaleWidth(width float32) bool {
	width += l.padding.X() + l.padding.Z()
	ps := l.PixelSize()
	if matrix.ApproxTo(ps[matrix.Vx], width, 0.001) {
		return false
	}
	size := matrix.Vec3{width, ps.Height(), 1.0}
	if l.ui.Entity().Parent != nil {
		size.DivideAssign(l.ui.Entity().Parent.Transform.WorldScale())
	}
	l.ui.Entity().ScaleWithoutChildren(size)
	l.prepare()
	l.ui.layoutChanged(DirtyTypeResize)
	return true
}

func (l *Layout) ScaleHeight(height float32) bool {
	height += l.padding.Y() + l.padding.W()
	ps := l.PixelSize()
	if matrix.ApproxTo(ps.Y(), height, 0.001) {
		return false
	}
	if height == 0 {
		return false
	}
	size := matrix.Vec3{ps.Width(), height, 1.0}
	if l.ui.Entity().Parent != nil {
		size.DivideAssign(l.ui.Entity().Parent.Transform.WorldScale())
	}
	l.ui.Entity().ScaleWithoutChildren(size)
	l.prepare()
	l.ui.layoutChanged(DirtyTypeResize)
	return true
}

func (l *Layout) Positioning() Positioning { return l.positioning }
func (l *Layout) Anchor() Anchor           { return l.screenAnchor }
func (l *Layout) Border() matrix.Vec4      { return l.border }
func (l *Layout) Padding() matrix.Vec4     { return l.padding }
func (l *Layout) Margin() matrix.Vec4      { return l.margin }
func (l *Layout) Offset() matrix.Vec2      { return matrix.Vec2{l.offset.X(), l.offset.Y()} }

func (l *Layout) totalOffsetBounds() matrix.Vec4 {
	return matrix.Vec4{
		l.CalcOffset().X(),
		l.CalcOffset().Y(),
		l.CalcOffset().X(),
		l.CalcOffset().Y(),
	}
}

func (l *Layout) Stretch() matrix.Vec4 {
	return matrix.Vec4{l.left, l.top, l.right, l.bottom}
}

func (l *Layout) SetBorder(left, top, right, bottom float32) {
	l.border = matrix.Vec4{left, top, right, bottom}
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) SetPadding(left, top, right, bottom float32) {
	lastPad := l.padding
	l.padding = matrix.Vec4{left, top, right, bottom}
	ps := l.PixelSize()
	l.Scale(ps.Width()-lastPad.X()-lastPad.Z(),
		ps.Height()-lastPad.Y()-lastPad.W())
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) SetMargin(left, top, right, bottom float32) {
	l.margin = matrix.Vec4{left, top, right, bottom}
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) AnchorTo(anchorPosition Anchor) {
	if l.screenAnchor == anchorPosition {
		return
	}
	var lfn func(*Layout) = nil
	var afn func(*Layout, float32, float32, matrix.Vec2) matrix.Vec4 = nil
	if anchorPosition == AnchorTopLeft {
		afn = anchorTopLeft
		lfn = layoutFloating
	} else if anchorPosition == AnchorTopCenter {
		afn = anchorTopCenter
		lfn = layoutFloating
	} else if anchorPosition == AnchorTopRight {
		afn = anchorTopRight
		lfn = layoutFloating
	} else if anchorPosition == AnchorLeft {
		afn = anchorLeft
		lfn = layoutFloating
	} else if anchorPosition == AnchorCenter {
		afn = anchorCenter
		lfn = layoutFloating
	} else if anchorPosition == AnchorRight {
		afn = anchorRight
		lfn = layoutFloating
	} else if anchorPosition == AnchorBottomLeft {
		afn = anchorBottomLeft
		lfn = layoutFloating
	} else if anchorPosition == AnchorBottomCenter {
		afn = anchorBottomCenter
		lfn = layoutFloating
	} else if anchorPosition == AnchorBottomRight {
		afn = anchorBottomRight
		lfn = layoutFloating
	} else if anchorPosition == AnchorStretchLeft {
		afn = anchorStretchLeft
		lfn = layoutStretch
	} else if anchorPosition == AnchorStretchTop {
		afn = anchorStretchTop
		lfn = layoutStretch
	} else if anchorPosition == AnchorStretchRight {
		afn = anchorStretchRight
		lfn = layoutStretch
	} else if anchorPosition == AnchorStretchBottom {
		afn = anchorStretchBottom
		lfn = layoutStretch
	} else if anchorPosition == AnchorStretchCenter {
		afn = anchorStretchCenter
		lfn = layoutStretch
	} else {
		panic("Invalid anchor position")
	}
	l.screenAnchor = anchorPosition
	l.anchorFunction = afn
	l.layoutFunction = lfn
	//layout.ui.layoutChanged(DirtyTypeLayout)
	l.ui.layoutChanged(DirtyTypeGenerated)
}

func (l *Layout) update() {
	if l.layoutFunction != nil {
		l.prepare()
		l.layoutFunction(l)
		if l.ui.hasScissor() {
			l.ui.GenerateScissor()
		}
	}
}

func (l *Layout) Z() float32 {
	return l.z
}

func (l *Layout) SetZ(z float32) {
	l.z = z
}

func (l *Layout) SetPositioning(pos Positioning) {
	l.positioning = pos
	l.ui.SetDirty(DirtyTypeLayout)
}

func (l *Layout) ContentSize() (float32, float32) {
	ps := l.PixelSize()
	return ps.X() - l.padding.X() - l.padding.Z(),
		ps.Y() - l.padding.Y() - l.padding.W()
}
