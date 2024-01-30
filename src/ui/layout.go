package ui

import (
	"kaiju/engine"
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
	innerOffset      matrix.Vec4
	localInnerOffset matrix.Vec4
	left             float32
	top              float32
	right            float32
	bottom           float32
	z                float32
	anchor           matrix.Vec2
	pixelSize        matrix.Vec2
	ui               UI
	screenAnchor     Anchor
	layoutFunction   func(layout *Layout)
	anchorFunction   func(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4
	bounds           matrix.Vec4
	lastParent       *engine.Entity
	mySize           matrix.Vec3
	worldScalar      matrix.Vec2
	border           matrix.Vec4
	padding          matrix.Vec4
	margin           matrix.Vec4
	inset            matrix.Vec4
	positioning      Positioning
	functions        []func(layout *Layout)
	runningFuncs     bool
}

func NewLayout(ui UI) Layout {
	return Layout{
		ui:        ui,
		anchor:    matrix.Vec2{0.0, 0.0},
		pixelSize: matrix.Vec2{1.0, 1.0},
	}
}

func (layout *Layout) AddFunction(fn func(layout *Layout)) {
	layout.functions = append(layout.functions, fn)
}

func anchorTopLeft(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{size.X() * 0.5, h - size.Y()*0.5, 0.0, 0.0}
}

func anchorTopCenter(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{w * 0.5, h - size.Y()*0.5, 0.0, 0.0}
}

func anchorTopRight(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{w - size.X()*0.5, h - size.Y()*0.5, 0.0, 0.0}
}

func anchorLeft(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{size.X() * 0.5, h * 0.5, 0.0, 0.0}
}

func anchorCenter(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{w * 0.5, h * 0.5, 0.0, 0.0}
}

func anchorRight(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{w - size.X()*0.5, h * 0.5, 0.0, 0.0}
}

func anchorBottomLeft(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{size.X() * 0.5, size.Y() * 0.5, 0.0, 0.0}
}

func anchorBottomCenter(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{w * 0.5, size.Y() * 0.5, 0.0, 0.0}
}

func anchorBottomRight(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	return matrix.Vec4{w - size.X()*0.5, size.Y() * 0.5, 0.0, 0.0}
}

func layoutFloating(self *Layout) {
	t := &self.ui.Entity().Transform
	w := self.bounds.Z() - self.bounds.X()
	h := self.bounds.W() - self.bounds.Y()
	pos := self.anchorFunction(self, w, h, self.mySize)
	p := matrix.Vec3{pos.X() + self.offset.X() + self.bounds.X(),
		pos.Y() + self.offset.Y() + self.bounds.Y(), t.Position().Z()}
	t.SetPosition(p)
}

func anchorStretchLeft(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	xSize := self.right
	ySize := h - (self.top + self.bottom)
	xMid := xSize*0.5 + self.left
	yMid := self.bottom + ySize*0.5
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchTop(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	xSize := w - (self.left + self.right)
	ySize := self.bottom
	xMid := self.left + xSize*0.5
	yMid := h - (ySize * 0.5) - self.top
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchRight(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	xSize := self.left
	ySize := h - (self.top + self.bottom)
	xMid := w - (xSize * 0.5) - self.right
	yMid := self.bottom + ySize*0.5
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchBottom(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	xSize := w - (self.left + self.right)
	ySize := self.top
	xMid := self.left + xSize*0.5
	yMid := ySize*0.5 + self.bottom
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func anchorStretchCenter(self *Layout, w, h float32, size matrix.Vec3) matrix.Vec4 {
	xSize := w - (self.left + self.right)
	ySize := h - (self.top + self.bottom)
	xMid := self.left + xSize*0.5
	yMid := self.bottom + ySize*0.5
	return matrix.Vec4{xMid, yMid, xSize, ySize}
}

func layoutStretch(self *Layout) {
	t := &self.ui.Entity().Transform
	width := self.bounds.Z() - self.bounds.X()
	height := self.bounds.W() - self.bounds.Y()
	res := self.anchorFunction(self, width, height, matrix.Vec3Zero())
	x := res.X() + self.offset.X()
	y := res.Y() + self.offset.Y()
	xSize := res.Z()
	ySize := res.W()
	scale := matrix.Vec3{xSize * self.worldScalar.X(), ySize * self.worldScalar.Y(), self.mySize.Z()}
	scale[matrix.Vx] -= (self.inset.X() + self.inset.Z()) / width
	scale[matrix.Vy] -= (self.inset.Y() + self.inset.W()) / height
	self.ui.Entity().ScaleWithoutChildren(scale)
	self.pixelSize = self.ui.Entity().Transform.WorldScale().AsVec2()
	pos := matrix.Vec3{
		x + self.bounds.X() + (self.inset.X()-self.inset.Z())*0.5,
		y + self.bounds.Y() + (self.inset.W()-self.inset.Y())*0.5,
		t.Position().Z(),
	}
	t.SetPosition(pos)
}

func (layout *Layout) prepare() {
	if layout.runningFuncs {
		return
	}
	layout.runningFuncs = true
	for _, fn := range layout.functions {
		fn(layout)
	}
	layout.runningFuncs = false
	layout.setBounds()
}

func (layout *Layout) setBounds() {
	t := &layout.ui.Entity().Transform
	layout.mySize = t.WorldScale()
	if layout.ui.Entity().IsRoot() {
		layout.bounds = matrix.Vec4{0, 0,
			matrix.Float(layout.ui.selfHost().Window.Width()),
			matrix.Float(layout.ui.selfHost().Window.Height()),
		}
		layout.worldScalar = matrix.Vec2One()
		et := &layout.ui.Entity().Transform
		pos := matrix.Vec3{et.Position().X(), et.Position().Y(), layout.z}
		et.SetPosition(pos)
	} else {
		parent := layout.ui.Entity().Parent
		s := parent.Transform.WorldScale()
		p := parent
		layout.inset = matrix.Vec4Zero()
		for ; p != nil; p = p.Parent {
			pUI := FirstOnEntity(p)
			if pUI != nil {
				pLayout := pUI.Layout()
				layout.inset.AddAssign(pLayout.margin)
				layout.inset.AddAssign(pLayout.border)
				layout.inset.AddAssign(pLayout.padding)
			}
		}
		layout.worldScalar = matrix.Vec2{1.0 / s.X(), 1.0 / s.Y()}
		layout.bounds = matrix.Vec4{-s.X() * 0.5, -s.Y() * 0.5, s.X() * 0.5, s.Y() * 0.5}
		// Set this child in front of parent
		et := &layout.ui.Entity().Transform
		pos := matrix.Vec3{et.Position().X(), et.Position().Y(), 0.2 + layout.z}
		et.SetPosition(pos)
	}
	if layout.parentChanged() {
		layout.lastParent = layout.ui.Entity().Parent
		layout.Scale(layout.pixelSize.Width(), layout.pixelSize.Height())
		for _, c := range layout.ui.Entity().Children {
			if ui := FirstOnEntity(c); ui != nil {
				cl := ui.Layout()
				cl.Scale(cl.pixelSize.Width(), cl.pixelSize.Height())
			}
		}
	}
}

func (layout *Layout) initialize(ui UI, anchor Anchor) {
	layout.anchor = matrix.Vec2Zero()
	layout.ui = ui
	layout.AnchorTo(anchor)
	layout.prepare()
	layout.update()
}

func (layout *Layout) SetOffset(x, y float32) {
	if matrix.Vec2Approx(layout.offset, matrix.Vec2{x, y}) {
		return
	}
	layout.offset.SetX(x)
	layout.offset.SetY(y)
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout *Layout) SetInnerOffset(left, top, right, bottom float32) {
	if matrix.Vec4Approx(layout.innerOffset, matrix.Vec4{left, top, right, bottom}) {
		return
	}
	layout.innerOffset = matrix.Vec4{left, top, right, bottom}
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout *Layout) SetInnerOffsetLeft(offset float32) {
	if matrix.Approx(layout.innerOffset.X(), offset) {
		return
	}
	layout.innerOffset.SetX(offset)
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout *Layout) SetInnerOffsetTop(offset float32) {
	if matrix.Approx(layout.innerOffset[matrix.Vy], offset) {
		return
	}
	layout.innerOffset.SetY(offset)
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout *Layout) SetInnerOffsetRight(offset float32) {
	if matrix.Approx(layout.innerOffset.Right(), offset) {
		return
	}
	layout.innerOffset.SetRight(offset)
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout *Layout) SetInnerOffsetBottom(offset float32) {
	if matrix.Approx(layout.innerOffset.Bottom(), offset) {
		return
	}
	layout.innerOffset.SetBottom(offset)
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout *Layout) LocalInnerOffset() matrix.Vec4 {
	return layout.localInnerOffset
}

func (layout *Layout) SetLocalInnerOffset(left, top, right, bottom float32) {
	if matrix.Vec4Approx(layout.localInnerOffset, matrix.Vec4{left, top, right, bottom}) {
		return
	}
	layout.localInnerOffset = matrix.Vec4{left, top, right, bottom}
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout Layout) InnerOffset() matrix.Vec4 {
	return matrix.Vec4{
		layout.localInnerOffset.Left() + layout.innerOffset.Left(),
		layout.localInnerOffset.Top() + layout.innerOffset.Top(),
		layout.localInnerOffset.Right() + layout.innerOffset.Right(),
		layout.localInnerOffset.Bottom() + layout.innerOffset.Bottom(),
	}
}

func (layout *Layout) SetStretch(left, top, right, bottom float32) {
	layout.left = left
	layout.top = top
	layout.right = right
	layout.bottom = bottom
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout *Layout) SetStretchRatio(leftRatio, topRatio, rightRatio, bottomRatio float32) {
	w := layout.bounds.Z() - layout.bounds.X()
	h := layout.bounds.W() - layout.bounds.Y()
	layout.left = w * leftRatio
	layout.top = h * topRatio
	layout.right = w * rightRatio
	layout.bottom = h * bottomRatio
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout *Layout) Scale(width, height float32) {
	width += layout.padding.X() + layout.padding.Z()
	height += layout.padding.Y() + layout.padding.W()
	if matrix.Vec2Approx(layout.pixelSize, matrix.Vec2{width, height}) {
		return
	}
	layout.pixelSize.SetX(width)
	layout.pixelSize.SetY(height)
	size := matrix.Vec3{width, height, 1.0}
	if layout.ui.Entity().Parent != nil {
		size.DivideAssign(layout.ui.Entity().Parent.Transform.WorldScale())
	}
	layout.ui.Entity().ScaleWithoutChildren(size)
	layout.prepare()
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout *Layout) ScaleWidth(width float32) {
	width += layout.padding.X() + layout.padding.Z()
	if matrix.Approx(layout.pixelSize[matrix.Vx], width) {
		return
	}
	layout.pixelSize.SetX(width)
	size := matrix.Vec3{width, layout.pixelSize.Height(), 1.0}
	if layout.ui.Entity().Parent != nil {
		size.DivideAssign(layout.ui.Entity().Parent.Transform.WorldScale())
	}
	layout.ui.Entity().ScaleWithoutChildren(size)
	layout.prepare()
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout *Layout) ScaleHeight(height float32) {
	height += layout.padding.Y() + layout.padding.W()
	if matrix.Approx(layout.pixelSize.Y(), height) {
		return
	}
	layout.pixelSize.SetY(height)
	size := matrix.Vec3{layout.pixelSize.Width(), height, 1.0}
	if layout.ui.Entity().Parent != nil {
		size.DivideAssign(layout.ui.Entity().Parent.Transform.WorldScale())
	}
	layout.ui.Entity().ScaleWithoutChildren(size)
	layout.prepare()
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout Layout) Positioning() Positioning { return layout.positioning }
func (layout Layout) Anchor() Anchor           { return layout.screenAnchor }
func (layout Layout) Border() matrix.Vec4      { return layout.border }
func (layout Layout) Padding() matrix.Vec4     { return layout.padding }
func (layout Layout) Margin() matrix.Vec4      { return layout.margin }
func (layout Layout) Offset() matrix.Vec2      { return matrix.Vec2{layout.offset.X(), layout.offset.Y()} }

func (layout Layout) PixelSize() matrix.Vec2 {
	return matrix.NewVec2(layout.pixelSize.Width(), layout.pixelSize.Height())
}

func (layout Layout) Stretch() matrix.Vec4 {
	return matrix.Vec4{layout.left, layout.top, layout.right, layout.bottom}
}

func (layout *Layout) SetBorder(left, top, right, bottom float32) {
	layout.border = matrix.Vec4{left, top, right, bottom}
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout *Layout) SetPadding(left, top, right, bottom float32) {
	lastPad := layout.padding
	layout.padding = matrix.Vec4{left, top, right, bottom}
	layout.Scale(layout.pixelSize.Width()-lastPad.X()-lastPad.Z(),
		layout.pixelSize.Height()-lastPad.Y()-lastPad.W())
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout *Layout) SetMargin(left, top, right, bottom float32) {
	layout.margin = matrix.Vec4{left, top, right, bottom}
	layout.ui.SetDirty(DirtyTypeResize)
}

func (layout *Layout) AnchorTo(anchorPosition Anchor) {
	if layout.screenAnchor == anchorPosition {
		return
	}
	var lfn func(*Layout) = nil
	var afn func(*Layout, float32, float32, matrix.Vec3) matrix.Vec4 = nil
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
	layout.screenAnchor = anchorPosition
	layout.anchorFunction = afn
	layout.layoutFunction = lfn
	layout.ui.SetDirty(DirtyTypeGenerated)
}

func (layout *Layout) parentChanged() bool {
	return layout.lastParent != layout.ui.Entity().Parent
}

func (layout *Layout) update() {
	if layout.layoutFunction != nil {
		layout.prepare()
		layout.layoutFunction(layout)
		if layout.ui.hasScissor() {
			layout.ui.generateScissor()
		}
	}
}

func (layout Layout) Z() float32 {
	return layout.z
}

func (layout *Layout) SetZ(z float32) {
	layout.z = z
}

func (layout *Layout) SetPositioning(pos Positioning) {
	layout.positioning = pos
	layout.ui.SetDirty(DirtyTypeLayout)
}

func (layout *Layout) ContentSize() (float32, float32) {
	return layout.pixelSize.X() - layout.padding.X() - layout.padding.Z(),
		layout.pixelSize.Y() - layout.padding.Y() - layout.padding.W()
}
