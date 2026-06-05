/******************************************************************************/
/* layout.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"kaijuengine.com/matrix"
)

const (
	fractionOfPixel = 0.2
)

type Positioning = int
type FlexAlign = int

const (
	PositioningStatic = Positioning(iota)
	PositioningAbsolute
	PositioningFixed
	PositioningRelative
	PositioningSticky
)

const (
	FlexAlignAuto = FlexAlign(iota)
	FlexAlignStart
	FlexAlignEnd
	FlexAlignCenter
	FlexAlignStretch
)

type Layout struct {
	offset           matrix.Vec2
	rowLayoutOffset  matrix.Vec2
	innerOffset      matrix.Vec4
	localInnerOffset matrix.Vec4
	z                float32
	ui               *UI
	border           matrix.Vec4
	padding          matrix.Vec4
	margin           matrix.Vec4
	gridRowStart     int
	gridRowEnd       int
	gridColumnStart  int
	gridColumnEnd    int
	flexGrow         float32
	flexShrink       float32
	flexBasis        float32
	flexBasisAuto    bool
	flexBasisPercent bool
	flexOrder        int
	alignSelf        FlexAlign
	positioning      Positioning
	Stylizer         LayoutStylizer
	runningStylizer  bool
}

func (l *Layout) ClearStyles() {
	ps := l.PixelSize()
	if l.padding.Horizontal() != 0 || l.padding.Vertical() != 0 ||
		l.border.Horizontal() != 0 || l.border.Vertical() != 0 {
		ps.SetX(ps.X() - l.padding.Horizontal() - l.border.Horizontal())
		ps.SetY(ps.Y() - l.padding.Vertical() - l.border.Vertical())
		// Write the box-model-stripped size back to the transform, mirroring how
		// SetPadding/SetBorder apply it. Without this the inflated scale persists
		// and re-applying padding after ClearStyles accumulates.
		l.Scale(ps.Width(), ps.Height())
	}
	l.offset = matrix.Vec2{}
	l.rowLayoutOffset = matrix.Vec2{}
	l.innerOffset = matrix.Vec4{}
	l.localInnerOffset = matrix.Vec4{}
	l.z = 0
	l.border = matrix.Vec4{}
	l.padding = matrix.Vec4{}
	l.margin = matrix.Vec4{}
	l.gridRowStart = 0
	l.gridRowEnd = 0
	l.gridColumnStart = 0
	l.gridColumnEnd = 0
	l.flexGrow = 0
	l.flexShrink = 1
	l.flexBasis = 0
	l.flexBasisAuto = true
	l.flexBasisPercent = false
	l.flexOrder = 0
	l.alignSelf = FlexAlignAuto
	l.positioning = PositioningStatic
}

func (l *Layout) PixelSize() matrix.Vec2 {
	return l.ui.Entity().Transform.WorldScale().AsVec2()
}

func (l *Layout) Ui() *UI { return l.ui }

func (l *Layout) CalcOffset() matrix.Vec2 {
	return l.rowLayoutOffset.Add(l.offset)
}

func (l *Layout) SetOffset(x, y float32) {
	if matrix.Vec2ApproxTo(l.offset, matrix.Vec2{x, y}, fractionOfPixel) {
		return
	}
	l.offset.SetX(x)
	l.offset.SetY(y)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetOffsetX(x float32) {
	if matrix.ApproxTo(l.offset.X(), x, fractionOfPixel) {
		return
	}
	l.offset.SetX(x)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetOffsetY(y float32) {
	if matrix.ApproxTo(l.offset.Y(), y, fractionOfPixel) {
		return
	}
	l.offset.SetY(y)
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetInnerOffset(left, top, right, bottom float32) {
	io := matrix.Vec4{left, top, right, bottom}
	if matrix.Vec4ApproxTo(l.innerOffset, io, fractionOfPixel) {
		return
	}
	l.innerOffset = io
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
	if matrix.Vec4ApproxTo(l.localInnerOffset, matrix.Vec4{left, top, right, bottom}, fractionOfPixel) {
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

func (l *Layout) Scale(width, height float32) bool {
	ps := l.PixelSize()
	if matrix.Vec2ApproxTo(ps, matrix.Vec2{width, height}, fractionOfPixel) {
		return false
	}
	if matrix.Approx(width, 0) || matrix.Approx(height, 0) {
		return false
	}
	size := matrix.Vec3{width, height, 1.0}
	if l.ui.Entity().Parent != nil {
		parentScale := l.ui.Entity().Parent.Transform.WorldScale()
		if matrix.Approx(parentScale.X(), 0) || matrix.Approx(parentScale.Y(), 0) {
			return false
		}
		size.DivideAssign(parentScale)
	}
	l.ui.Entity().Transform.ScaleWithoutChildren(size)
	l.ui.layoutChanged(DirtyTypeResize)
	return true
}

func (l *Layout) ScaleWidth(width float32) bool {
	ps := l.PixelSize()
	if matrix.ApproxTo(ps[matrix.Vx], width, fractionOfPixel) {
		return false
	}
	if matrix.Approx(width, 0) {
		return false
	}
	size := matrix.Vec3{width, ps.Height(), 1.0}
	if l.ui.Entity().Parent != nil {
		parentScale := l.ui.Entity().Parent.Transform.WorldScale()
		if matrix.Approx(parentScale.X(), 0) || matrix.Approx(parentScale.Y(), 0) {
			return false
		}
		size.DivideAssign(parentScale)
	}
	l.ui.Entity().Transform.ScaleWithoutChildren(size)
	l.prepare()
	l.ui.layoutChanged(DirtyTypeResize)
	return true
}

func (l *Layout) ScaleHeight(height float32) bool {
	ps := l.PixelSize()
	if matrix.ApproxTo(ps.Y(), height, fractionOfPixel) {
		return false
	}
	if matrix.Approx(height, 0) {
		return false
	}
	size := matrix.Vec3{ps.Width(), height, 1.0}
	if l.ui.Entity().Parent != nil {
		parentScale := l.ui.Entity().Parent.Transform.WorldScale()
		if matrix.Approx(parentScale.X(), 0) || matrix.Approx(parentScale.Y(), 0) {
			return false
		}
		size.DivideAssign(parentScale)
	}
	l.ui.Entity().Transform.ScaleWithoutChildren(size)
	l.prepare()
	l.ui.layoutChanged(DirtyTypeResize)
	return true
}

func (l *Layout) Positioning() Positioning { return l.positioning }
func (l *Layout) Border() matrix.Vec4      { return l.border }
func (l *Layout) Padding() matrix.Vec4     { return l.padding }
func (l *Layout) Margin() matrix.Vec4      { return l.margin }
func (l *Layout) Offset() matrix.Vec2      { return matrix.Vec2{l.offset.X(), l.offset.Y()} }
func (l *Layout) GridRowStart() int        { return l.gridRowStart }
func (l *Layout) GridRowEnd() int          { return l.gridRowEnd }
func (l *Layout) GridColumnStart() int     { return l.gridColumnStart }
func (l *Layout) GridColumnEnd() int       { return l.gridColumnEnd }
func (l *Layout) FlexGrow() float32        { return l.flexGrow }
func (l *Layout) FlexShrink() float32      { return l.flexShrink }
func (l *Layout) FlexBasis() float32       { return l.flexBasis }
func (l *Layout) FlexBasisAuto() bool      { return l.flexBasisAuto }
func (l *Layout) FlexBasisPercent() bool   { return l.flexBasisPercent }
func (l *Layout) FlexOrder() int           { return l.flexOrder }
func (l *Layout) AlignSelf() FlexAlign     { return l.alignSelf }

func (l *Layout) SetFlexGrow(grow float32) {
	if grow < 0 {
		grow = 0
	}
	if matrix.Approx(l.flexGrow, grow) {
		return
	}
	l.flexGrow = grow
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetFlexShrink(shrink float32) {
	if shrink < 0 {
		shrink = 0
	}
	if matrix.Approx(l.flexShrink, shrink) {
		return
	}
	l.flexShrink = shrink
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetFlexBasisAuto() {
	if l.flexBasisAuto && matrix.Approx(l.flexBasis, 0) && !l.flexBasisPercent {
		return
	}
	l.flexBasis = 0
	l.flexBasisAuto = true
	l.flexBasisPercent = false
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetFlexBasis(basis float32, percent bool) {
	if basis < 0 {
		basis = 0
	}
	if !l.flexBasisAuto && matrix.Approx(l.flexBasis, basis) && l.flexBasisPercent == percent {
		return
	}
	l.flexBasis = basis
	l.flexBasisAuto = false
	l.flexBasisPercent = percent
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetFlexOrder(order int) {
	if l.flexOrder == order {
		return
	}
	l.flexOrder = order
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetAlignSelf(align FlexAlign) {
	if l.alignSelf == align {
		return
	}
	l.alignSelf = align
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetGridRow(start, end int) {
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if l.gridRowStart == start && l.gridRowEnd == end {
		return
	}
	l.gridRowStart = start
	l.gridRowEnd = end
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetGridColumn(start, end int) {
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if l.gridColumnStart == start && l.gridColumnEnd == end {
		return
	}
	l.gridColumnStart = start
	l.gridColumnEnd = end
	l.ui.layoutChanged(DirtyTypeLayout)
}

func (l *Layout) SetBorder(left, top, right, bottom float32) {
	b := matrix.Vec4{left, top, right, bottom}
	if matrix.Vec4ApproxTo(l.border, b, fractionOfPixel) {
		return
	}
	ps := l.PixelSize()
	// Undo last border applied to the size
	ps.SetX(ps.X() - l.border.Horizontal())
	ps.SetY(ps.Y() - l.border.Vertical())
	l.border = b
	ps.SetX(ps.X() + l.border.Horizontal())
	ps.SetY(ps.Y() + l.border.Vertical())
	l.Scale(ps.Width(), ps.Height())
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) SetPadding(left, top, right, bottom float32) {
	newPadding := matrix.Vec4{left, top, right, bottom}
	if matrix.Vec4ApproxTo(l.padding, newPadding, fractionOfPixel) {
		return
	}
	ps := l.PixelSize()
	// Undo last padding applied to the size
	ps.SetX(ps.X() - l.padding.Horizontal())
	ps.SetY(ps.Y() - l.padding.Vertical())
	l.padding = newPadding
	ps.SetX(ps.X() + l.padding.Horizontal())
	ps.SetY(ps.Y() + l.padding.Vertical())
	l.Scale(ps.Width(), ps.Height())
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) SetMargin(left, top, right, bottom float32) {
	m := matrix.Vec4{left, top, right, bottom}
	if matrix.Vec4ApproxTo(m, l.margin, fractionOfPixel) {
		return
	}
	l.margin = m
	l.ui.layoutChanged(DirtyTypeResize)
}

func (l *Layout) Z() float32 {
	return l.z
}

func (l *Layout) SetZ(z float32) {
	l.z = z
}

func (l *Layout) SetPositioning(pos Positioning) {
	if l.positioning != pos {
		l.positioning = pos
		l.SetInnerOffset(0, 0, 0, 0)
		l.SetRowLayoutOffset(matrix.Vec2Zero())
		l.ui.SetDirty(DirtyTypeLayout)
	}
}

func (l *Layout) ContentSize() matrix.Vec2 {
	ps := l.PixelSize()
	return matrix.NewVec2(ps.X()-l.padding.Horizontal()-l.border.Horizontal(),
		ps.Y()-l.padding.Vertical()-l.border.Vertical())
}

func (l *Layout) SetRowLayoutOffset(offset matrix.Vec2) {
	if matrix.Vec2ApproxTo(l.rowLayoutOffset, offset, fractionOfPixel) {
		return
	}
	l.rowLayoutOffset = offset
	l.ui.SetDirty(DirtyTypeLayout)
}

func (l *Layout) update() {
	l.prepare()
	layoutFloating(l)
}

func (l *Layout) totalOffsetBounds() matrix.Vec4 {
	offset := l.CalcOffset()
	if l.positioning == PositioningAbsolute && !l.ui.Entity().IsRoot() {
		if parentUI := FirstOnEntity(l.ui.Entity().Parent); parentUI != nil {
			border := parentUI.Layout().Border()
			offset.SetX(offset.X() + border.Left())
			offset.SetY(offset.Y() + border.Top())
		}
	}
	return matrix.Vec4{
		offset.X(),
		offset.Y(),
		offset.X(),
		offset.Y(),
	}
}

func al(edges matrix.Vec4, w float32, size matrix.Vec2) float32 {
	return -w*0.5 + size.X()*0.5 + edges.Left()
}

func at(edges matrix.Vec4, h float32, size matrix.Vec2) float32 {
	return h*0.5 - size.Y()*0.5 - edges.Top()
}

func anchorTopLeft(self *Layout, w, h float32, size matrix.Vec2) matrix.Vec4 {
	edges := self.totalOffsetBounds()
	inner := self.InnerOffset()
	return matrix.Vec4{al(edges, w, size) + inner.Left(), at(edges, h, size) + inner.Top(), 0, 0}
}

func layoutFloating(self *Layout) {
	t := &self.ui.Entity().Transform
	s := self.PixelSize()
	bounds := self.bounds()
	pos := anchorTopLeft(self, bounds.X(), bounds.Y(), s)
	pos.SetZ(self.z + 0.01)
	t.SetPosition(pos.AsVec3())
}

func (l *Layout) prepare() {
	switch l.ui.elmType {
	case ElementTypeInput:
		if l.runningStylizer {
			l.ui.ToInput().onLayoutUpdating()
		} else {
			l.runningStylizer = true
			l.ui.ToInput().onLayoutUpdating()
			l.runningStylizer = false
		}
	case ElementTypeTextArea:
		if l.runningStylizer {
			l.ui.ToTextArea().onLayoutUpdating()
		} else {
			l.runningStylizer = true
			l.ui.ToTextArea().onLayoutUpdating()
			l.runningStylizer = false
		}
	case ElementTypeSlider:
		if l.runningStylizer {
			l.ui.ToSlider().onLayoutUpdating()
		} else {
			l.runningStylizer = true
			l.ui.ToSlider().onLayoutUpdating()
			l.runningStylizer = false
		}
	case ElementTypeVirtualList:
		if l.runningStylizer {
			l.ui.ToVirtualList().onLayoutUpdating()
		} else {
			l.runningStylizer = true
			l.ui.ToVirtualList().onLayoutUpdating()
			l.runningStylizer = false
		}
	}
	if !l.runningStylizer && l.Stylizer != nil {
		l.runningStylizer = true
		l.Stylizer.ProcessStyle(l)
		l.runningStylizer = false
	}
}

func (l *Layout) bounds() matrix.Vec2 {
	if l.ui.Entity().IsRoot() {
		w := l.ui.Host().Window
		return matrix.Vec2{
			matrix.Float(w.Width()),
			matrix.Float(w.Height()),
		}
	} else {
		parent := l.ui.Entity().Parent
		s := parent.Transform.WorldScale()
		return matrix.Vec2{s.X(), s.Y()}
	}
}

func (l *Layout) initialize(ui *UI) {
	l.ui = ui
	l.flexShrink = 1
	l.flexBasisAuto = true
	l.alignSelf = FlexAlignAuto
	//l.prepare()
	//l.update()
}
