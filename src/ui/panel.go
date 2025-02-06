/******************************************************************************/
/* panel.go                                                                   */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
	"kaiju/assets"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/systems/events"
	"log/slog"
)

const (
	baseScrollSpeed = 24
)

type childScrollEvent struct {
	down   events.Id
	scroll events.Id
}

type localData interface {
}

type requestScroll struct {
	to        float32
	requested bool
}

type Panel = UI

// TODO:  On destroy call removeDrawings

func (p *Panel) EnforcingColor() bool {
	return len(p.PanelData().enforcedColorStack) > 0
}

func (p *Panel) ScrollX() float32        { return p.PanelData().scroll.X() }
func (p *Panel) ScrollY() float32        { return -p.PanelData().scroll.Y() }
func (p *Panel) EnableDragScroll()       { p.PanelData().allowDragScroll = true }
func (p *Panel) DisableDragScroll()      { p.PanelData().allowDragScroll = false }
func (p *Panel) DontFitContent()         { p.PanelData().fitContent = ContentFitNone }
func (p *Panel) FittingContent() bool    { return p.PanelData().fitContent != ContentFitNone }
func (p *Panel) SetSpeed(speed float32)  { p.PanelData().scrollSpeed = speed }
func (p *Panel) ResetScroll()            { p.PanelData().scroll = matrix.Vec2Zero() }
func (p *Panel) Freeze()                 { p.PanelData().frozen = true }
func (p *Panel) Unfreeze()               { p.PanelData().frozen = false }
func (p *Panel) BorderSize() matrix.Vec4 { return p.layout.border }

func (p *Panel) DontFitContentWidth() {
	switch p.PanelData().fitContent {
	case ContentFitBoth:
		p.PanelData().fitContent = ContentFitHeight
	case ContentFitWidth:
		p.PanelData().fitContent = ContentFitNone
	}
}

func (p *Panel) DontFitContentHeight() {
	switch p.PanelData().fitContent {
	case ContentFitBoth:
		p.PanelData().fitContent = ContentFitWidth
	case ContentFitHeight:
		p.PanelData().fitContent = ContentFitNone
	}
}

func (p *Panel) FitContentWidth() {
	switch p.PanelData().fitContent {
	case ContentFitNone:
		p.PanelData().fitContent = ContentFitWidth
	case ContentFitHeight:
		p.PanelData().fitContent = ContentFitBoth
	}
	if p.dirtyType == DirtyTypeNone {
		p.SetDirty(DirtyTypeLayout)
	} else {
		p.SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) FitContentHeight() {
	switch p.PanelData().fitContent {
	case ContentFitNone:
		p.PanelData().fitContent = ContentFitHeight
	case ContentFitWidth:
		p.PanelData().fitContent = ContentFitBoth
	}
	if p.dirtyType == DirtyTypeNone {
		p.SetDirty(DirtyTypeLayout)
	} else {
		p.SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) FitContent() {
	p.PanelData().fitContent = ContentFitBoth
	if p.dirtyType == DirtyTypeNone {
		p.SetDirty(DirtyTypeLayout)
	} else {
		p.SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) SetScrollX(value float32) {
	p.PanelData().requestScrollX.to = value
	p.PanelData().requestScrollX.requested = true
	p.SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetScrollY(value float32) {
	p.PanelData().requestScrollY.to = value
	p.PanelData().requestScrollY.requested = true
	p.SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetOverflow(overflow Overflow) {
	if p.PanelData().overflow != overflow {
		p.PanelData().overflow = overflow
		p.SetDirty(DirtyTypeLayout)
	}
}

func (p *Panel) Background(drawing int) *rendering.Texture {
	if p.PanelData().drawings[drawing].IsValid() {
		return p.PanelData().drawings[drawing].Textures[0]
	}
	return nil
}

func (p *Panel) recreateDrawing(drawing int) {
	var shader *rendering.Shader
	pd := p.PanelData()
	if len(p.overrideShaderDefinition) > 0 {
		shader = p.man.host.ShaderCache().ShaderFromDefinition(p.overrideShaderDefinition)
	} else {
		switch pd.shaderType {
		case PanelShaderTypeNine:
			shader = p.man.host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionUI)
		case PanelShaderTypeImage:
			shader = p.man.host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionUIImage)
		}
	}
	mesh := rendering.NewMeshQuad(p.man.host.MeshCache())
	d := &pd.drawings[drawing]
	var sd *ShaderData
	d.UseBlending = p.useBlending || pd.color.A() < 1
	if d.ShaderData != nil {
		sd = d.ShaderData.(*ShaderData)
		proxy := *sd
		sd.Destroy()
		d.ShaderData = nil
		*sd = proxy
	} else {
		d.Shader = shader
		d.Mesh = mesh
		d.Transform = &p.entity.Transform
		sd.ShaderDataBase = rendering.NewShaderDataBase()
		sd.BorderLen = matrix.NewVec2(8, 8)
		sd.BgColor = pd.color
		sd.FgColor = pd.color
		sd.UVs = matrix.NewVec4(0, 0, 1, 1)
		if len(d.Textures) > 0 {
			sd.setSize2d(p, float32(d.Textures[0].Width), float32(d.Textures[0].Height))
		}
		p.scissor = matrix.NewVec4(-matrix.FloatMax, -matrix.FloatMax,
			matrix.FloatMax, matrix.FloatMax)
		sd.Scissor = p.scissor
	}
	d.ShaderData = sd
}

func (p *Panel) recreateDrawings() {
	pd := p.PanelData()
	for i := range pd.drawings {
		p.recreateDrawing(i)
	}
}

func (p *Panel) removeDrawings() {
	pd := p.PanelData()
	for i := range pd.drawings {
		// TODO:  Does anything need to be released for the drawing?
		pd.drawings[i].ShaderData.Destroy()
	}
	pd.drawings = pd.drawings[:0]
}

func (p *Panel) ShaderData(drawing int) *ShaderData {
	return p.PanelData().drawings[drawing].ShaderData.(*ShaderData)
}

func (p *Panel) onScroll() {
	pd := p.PanelData()
	if pd.frozen {
		return
	}
	mouse := &p.man.host.Window.Mouse
	delta := mouse.Scroll()
	scroll := pd.scroll
	if !mouse.Scrolled() {
		pos := p.cursorPos(&p.man.host.Window.Cursor)
		delta = pos.Subtract(p.downPos)
		delta[matrix.Vy] *= -1.0
	} else {
		pd.offset = pd.scroll
		delta.ScaleAssign(pd.scrollSpeed)
	}
	// If the panel can only scroll horizontally, use the Y scroll if there is no X
	if pd.scrollDirection == PanelScrollDirectionHorizontal {
		if matrix.ApproxTo(delta.X(), 0, matrix.RealTiny) {
			delta.SetX(-delta.Y())
		}
	}
	if (pd.scrollDirection & PanelScrollDirectionHorizontal) != 0 {
		x := klib.Clamp(delta.X()+pd.offset.X(), 0, pd.maxScroll.X())
		scroll.SetX(x)
	}
	if (pd.scrollDirection & PanelScrollDirectionVertical) != 0 {
		y := klib.Clamp(delta.Y()+pd.offset.Y(), -pd.maxScroll.Y(), 0.0)
		scroll.SetY(y)
	}
	if scroll.Equals(pd.scroll) {
		pd.scroll = scroll
		p.SetDirty(DirtyTypeLayout)
		pd.isScrolling = true
	}
}

func (p *Panel) onDown() {
	if len(p.man.host.Window.Touch.Pointers) != 1 {
		return
	}
	target := p
	for target != nil && target.PanelData().scrollDirection == PanelScrollDirectionNone {
		found := FirstOnEntity(target.entity.Parent)
		target = found
	}
	if target == nil {
		return
	}
	pd := target.PanelData()
	pd.offset = pd.scroll
	pd.dragging = true
	if !pd.allowDragScroll {
		// TODO:  Change the mouse cursor to look like it's dragging something
	}
}

func (p *Panel) onUI() {
	p.PanelData().dragging = false
}

func (p *Panel) boundsChildren(bounds *matrix.Vec2) {
	for i := range p.entity.Children {
		kid := p.entity.Children[i]
		pos := kid.Transform.Position()
		kui := FirstOnEntity(kid)
		if kui == nil {
			slog.Error("child of ui is not a ui element")
			continue
		}
		if kui.layout.positioning == PositioningAbsolute {
			continue
		}
		var size matrix.Vec2
		if kui.elmType == ElementTypeLabel {
			size = (*Label)(kui).Measure()
			// Give a little margin for error on text
			size[matrix.Vx] += 0.1
		} else {
			size = kid.Transform.WorldScale().AsVec2()
			kui.boundsChildren(bounds)
		}
		r := pos.X() + size.X()
		b := pos.Y() + size.Y()
		*bounds = matrix.NewVec2(max(bounds.X(), r), max(bounds.Y(), b))
	}
}

func (p *Panel) postLayoutUpdate() {
	if len(p.entity.Children) == 0 {
		return
	}
	pd := p.PanelData()
	if pd.requestScrollX.requested {
		x := klib.Clamp(pd.requestScrollX.to, 0, pd.maxScroll.X())
		pd.scroll.SetX(x)
		pd.requestScrollX.requested = false
	}
	if pd.requestScrollY.requested {
		y := klib.Clamp(-pd.requestScrollY.to, -pd.maxScroll.Y(), 0)
		pd.scroll.SetY(y)
		pd.requestScrollY.requested = true
	}
	offsetStart := matrix.NewVec2(-pd.scroll.X(), pd.scroll.Y())
	rows := []rowBuilder{}
	ps := p.layout.PixelSize()
	areaWidth := ps.X() - p.layout.padding.X() - p.layout.padding.Z() -
		p.layout.border.X() - p.layout.border.Z()
	maxSize := matrix.Vec2{}
	for i := range p.entity.Children {
		kid := p.entity.Children[i]
		if !kid.IsActive() || kid.IsDestroyed() {
			continue
		}
		kui := FirstOnEntity(kid)
		if kui == nil {
			slog.Error("child of ui is not a ui element")
			continue
		}
		kLayout := &kui.layout
		switch kLayout.positioning {
		case PositioningAbsolute:
			if kLayout.screenAnchor.IsTop() {
				kLayout.rowLayoutOffset.SetY(p.layout.innerOffset.Top() +
					p.layout.padding.Top() + p.layout.border.Top())
			} else if kLayout.screenAnchor.IsBottom() {
				kLayout.rowLayoutOffset.SetY(p.layout.innerOffset.Bottom() +
					p.layout.padding.Bottom() + p.layout.border.Bottom())
			}
			if kLayout.screenAnchor.IsLeft() {
				kLayout.rowLayoutOffset.SetX(p.layout.innerOffset.Left() +
					p.layout.padding.Left() + p.layout.border.Left() -
					pd.scroll.X())
			} else if kLayout.screenAnchor.IsRight() {
				kLayout.rowLayoutOffset.SetX(p.layout.innerOffset.Right() +
					p.layout.padding.Right() + p.layout.border.Right() -
					pd.scroll.X())
			}
			kws := kid.Transform.WorldScale()
			maxSize.SetX(max(maxSize.X(), kLayout.left+kLayout.offset.X()+kws.Width()))
			maxSize.SetY(max(maxSize.Y(), kLayout.top+kLayout.offset.Y()+kws.Height()))
		case PositioningRelative:
		case PositioningStatic:
			if len(rows) == 0 || !rows[len(rows)-1].addElement(areaWidth, kui) {
				rows = append(rows, rowBuilder{})
				rows[len(rows)-1].addElement(areaWidth, kui)
			}
		}
	}
	nextPos := offsetStart
	addY := p.layout.padding.Y() + p.layout.border.Y()
	nextPos[matrix.Vy] += addY
	maxSize[matrix.Vy] += addY
	for i := range rows {
		rows[i].setElements(p.layout.padding.X()+p.layout.border.X(), nextPos.Y())
		addY = rows[i].height + rows[i].maxMarginTop + rows[i].maxMarginBottom
		nextPos[matrix.Vy] += addY
		maxSize[matrix.Vy] += addY
	}
	bounds := matrix.NewVec2(maxSize.X(), maxSize.Y())
	if p.FittingContent() {
		p.boundsChildren(&bounds)
		if pd.fitContent == ContentFitWidth {
			p.layout.ScaleWidth(max(1, bounds.X()))
		} else if pd.fitContent == ContentFitHeight {
			p.layout.ScaleHeight(max(1, bounds.Y()))
		} else if pd.fitContent == ContentFitBoth {
			p.layout.Scale(max(1, bounds.X()), max(1, bounds.Y()))
		}
	}
	last := pd.maxScroll
	ws := p.entity.Transform.WorldScale()
	pd.maxScroll = matrix.NewVec2(
		max(0, bounds.X()-ws.X()),
		max(0, bounds.Y()-ws.Y()))
	if !last.Roughly(pd.maxScroll) {
		p.SetDirty(DirtyTypeGenerated)
	}
}
