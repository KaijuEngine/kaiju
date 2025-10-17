/******************************************************************************/
/* layout.go                                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package ui

import (
	"kaiju/matrix"
)

const (
	fractionOfPixel = 0.2
)

type Positioning = int

const (
	PositioningStatic = Positioning(iota)
	PositioningAbsolute
	PositioningFixed
	PositioningRelative
	PositioningSticky
)

type Layout struct {
	offset           matrix.Vec2
	rowLayoutOffset  matrix.Vec2
	innerOffset      matrix.Vec4
	localInnerOffset matrix.Vec4
	left             float32
	top              float32
	z                float32
	ui               *UI
	border           matrix.Vec4
	padding          matrix.Vec4
	margin           matrix.Vec4
	positioning      Positioning
	Stylizer         LayoutStylizer
	runningStylizer  bool
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
		size.DivideAssign(l.ui.Entity().Parent.Transform.WorldScale())
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
	size := matrix.Vec3{width, ps.Height(), 1.0}
	if l.ui.Entity().Parent != nil {
		size.DivideAssign(l.ui.Entity().Parent.Transform.WorldScale())
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
		size.DivideAssign(l.ui.Entity().Parent.Transform.WorldScale())
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
	return matrix.Vec4{
		l.CalcOffset().X(),
		l.CalcOffset().Y(),
		l.CalcOffset().X(),
		l.CalcOffset().Y(),
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
	case ElementTypeSlider:
		if l.runningStylizer {
			l.ui.ToSlider().onLayoutUpdating()
		} else {
			l.runningStylizer = true
			l.ui.ToSlider().onLayoutUpdating()
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
	//l.prepare()
	//l.update()
}
