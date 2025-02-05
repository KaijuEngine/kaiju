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
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/systems/events"
	"log/slog"
)

type PanelScrollDirection = int32
type BorderStyle = int32
type ContentFit = int32
type Overflow = int

const (
	PanelScrollDirectionNone       = 0x00
	PanelScrollDirectionVertical   = 0x01
	PanelScrollDirectionHorizontal = 0x02
	PanelScrollDirectionBoth       = 0x03
)

const (
	BorderStyleNone = iota
	BorderStyleHidden
	BorderStyleDotted
	BorderStyleDashed
	BorderStyleSolid
	BorderStyleDouble
	BorderStyleGroove
	BorderStyleRidge
	BorderStyleInset
	BorderStyleOutset
)

const (
	ContentFitNone = iota
	ContentFitWidth
	ContentFitHeight
	ContentFitBoth
)

const (
	OverflowScroll = iota
	OverflowVisible
	OverflowHidden
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

type Panel struct {
	uiBase
	scroll, offset, maxScroll matrix.Vec2
	scrollSpeed               float32
	scrollDirection           PanelScrollDirection
	scrollEvent               events.Id
	borderStyle               [4]BorderStyle
	color                     matrix.Color
	drawing                   rendering.Drawing
	localData                 localData
	innerUpdate               func(deltaTime float64)
	fitContent                ContentFit
	requestScrollX            requestScroll
	requestScrollY            requestScroll
	overflow                  Overflow
	enforcedColorStack        []matrix.Color
	isScrolling               bool
	dragging                  bool
	frozen                    bool
	allowDragScroll           bool
}

func NewPanel(host *engine.Host, texture *rendering.Texture, anchor Anchor) *Panel {
	panel := &Panel{
		scrollEvent:        0,
		scrollSpeed:        20.0,
		scrollDirection:    PanelScrollDirectionVertical,
		color:              matrix.Color{1.0, 1.0, 1.0, 1.0},
		fitContent:         ContentFitBoth,
		enforcedColorStack: make([]matrix.Color, 0),
	}
	ts := matrix.Vec2Zero()
	if texture != nil {
		ts = texture.Size()
	}
	panel.updateId = host.Updater.AddUpdate(panel.update)
	panel.init(host, ts, anchor, panel)
	panel.entity.SetChildrenOrdered()
	if texture != nil {
		panel.ensureBGExists(texture)
	}
	panel.entity.OnActivate.Add(func() {
		panel.shaderData.Activate()
		panel.updateId = host.Updater.AddUpdate(panel.update)
		panel.SetDirty(DirtyTypeLayout)
		panel.Clean()
	})
	panel.entity.OnDeactivate.Add(func() {
		panel.shaderData.Deactivate()
		host.Updater.RemoveUpdate(panel.updateId)
		panel.updateId = 0
	})
	panel.entity.OnDestroy.Add(func() {
		panel.shaderData.Destroy()
	})
	return panel
}

func (p *Panel) ScrollX() float32   { return p.scroll.X() }
func (p *Panel) ScrollY() float32   { return -p.scroll.Y() }
func (p *Panel) EnableDragScroll()  { p.allowDragScroll = true }
func (p *Panel) DisableDragScroll() { p.allowDragScroll = false }

func (p *Panel) DontFitContentWidth() {
	switch p.fitContent {
	case ContentFitBoth:
		p.fitContent = ContentFitHeight
	case ContentFitWidth:
		p.fitContent = ContentFitNone
	}
}

func (p *Panel) DontFitContentHeight() {
	switch p.fitContent {
	case ContentFitBoth:
		p.fitContent = ContentFitWidth
	case ContentFitHeight:
		p.fitContent = ContentFitNone
	}
}

func (p *Panel) DontFitContent() {
	p.fitContent = ContentFitNone
}

func (p *Panel) FittingContent() bool {
	return p.fitContent != ContentFitNone
}

func (p *Panel) FitContentWidth() {
	switch p.fitContent {
	case ContentFitNone:
		p.fitContent = ContentFitWidth
	case ContentFitHeight:
		p.fitContent = ContentFitBoth
	}
	if p.dirtyType == DirtyTypeNone {
		p.SetDirty(DirtyTypeLayout)
	} else {
		p.SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) FitContentHeight() {
	switch p.fitContent {
	case ContentFitNone:
		p.fitContent = ContentFitHeight
	case ContentFitWidth:
		p.fitContent = ContentFitBoth
	}
	if p.dirtyType == DirtyTypeNone {
		p.SetDirty(DirtyTypeLayout)
	} else {
		p.SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) FitContent() {
	p.fitContent = ContentFitBoth
	if p.dirtyType == DirtyTypeNone {
		p.SetDirty(DirtyTypeLayout)
	} else {
		p.SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) onScroll() {
	mouse := &p.host.Window.Mouse
	delta := mouse.Scroll()
	scroll := p.scroll
	if !mouse.Scrolled() {
		pos := p.cursorPos(&p.host.Window.Cursor)
		delta = pos.Subtract(p.downPos)
		delta[matrix.Vy] *= -1.0
	} else {
		p.offset = p.scroll
		delta.ScaleAssign(p.scrollSpeed)
	}
	if (p.scrollDirection & PanelScrollDirectionHorizontal) != 0 {
		x := matrix.Clamp(delta.X()+p.offset.X(), 0.0, p.maxScroll.X())
		scroll.SetX(x)
	}
	if (p.scrollDirection & PanelScrollDirectionVertical) != 0 {
		y := matrix.Clamp(delta.Y()+p.offset.Y(), -p.maxScroll.Y(), 0)
		scroll.SetY(y)
	}
	if !matrix.Vec2Approx(scroll, p.scroll) {
		p.scroll = scroll
		p.SetDirty(DirtyTypeLayout)
		p.isScrolling = true
	}
}

func panelOnDown(ui UI) {
	var target UI = ui
	ok := false
	var panel *Panel
	for !ok {
		target = FirstOnEntity(target.Entity().Parent)
		panel, ok = target.(*Panel)
	}
	panel.offset = panel.scroll
	panel.dragging = true
	if !panel.allowDragScroll {
		// TODO:  Change the mouse cursor to look like it's dragging something
	}
}

func (p *Panel) update(deltaTime float64) {
	p.uiBase.eventUpdates()
	p.uiBase.Update(deltaTime)
	if !p.frozen {
		if p.isDown && p.dragging {
			if p.allowDragScroll {
				p.onScroll()
			}
		} else if p.dragging {
			p.dragging = false
		} else {
			p.isScrolling = false
		}
	}
	if p.innerUpdate != nil {
		p.innerUpdate(deltaTime)
	}
}

type rowBuilder struct {
	elements        []UI
	maxMarginTop    float32
	maxMarginBottom float32
	x               float32
	height          float32
}

func (rb *rowBuilder) addElement(areaWidth float32, e UI) bool {
	eSize := e.Layout().PixelSize()
	w := eSize.Width()
	if len(rb.elements) > 0 && rb.x+w > areaWidth {
		return false
	}
	rb.elements = append(rb.elements, e)
	rb.maxMarginTop = matrix.Max(rb.maxMarginTop, e.Layout().margin.Y())
	rb.maxMarginBottom = matrix.Max(rb.maxMarginBottom, e.Layout().margin.W())
	rb.x += w
	rb.height = matrix.Max(rb.height, eSize.Height())
	return true
}

func (rb rowBuilder) Width() float32 {
	return rb.x
}

func (rb rowBuilder) Height() float32 {
	return rb.height + rb.maxMarginTop + rb.maxMarginBottom
}

func (rb rowBuilder) setElements(offsetX, offsetY float32) {
	for _, e := range rb.elements {
		layout := e.Layout()
		x, y := offsetX, offsetY
		switch e.Layout().Positioning() {
		case PositioningAbsolute:
			fallthrough
		case PositioningRelative:
			if layout.Anchor().IsLeft() {
				x += layout.InnerOffset().Left()
			} else {
				x += layout.InnerOffset().Right()
			}
			if layout.Anchor().IsTop() {
				y += layout.InnerOffset().Top()
			} else {
				y += layout.InnerOffset().Bottom()
			}
		}
		x += layout.margin.X()
		y += rb.maxMarginTop
		layout.SetRowLayoutOffset(matrix.Vec2{x, y})
		offsetX += layout.PixelSize().Width() + layout.margin.X() + layout.margin.Z()
	}
}

func (p *Panel) boundsChildren(bounds *matrix.Vec2) {
	for _, kid := range p.entity.Children {
		pos := kid.Transform.Position()
		kui := FirstOnEntity(kid)
		if kui.Layout().Positioning() == PositioningAbsolute {
			continue
		}
		var size matrix.Vec2
		if lbl, ok := kui.(*Label); ok {
			size = lbl.Measure()
			// Give a little margin for error on text
			size[matrix.Vx] += 0.1
		} else {
			size = kid.Transform.WorldScale().AsVec2()
			kui.(*Panel).boundsChildren(bounds)
		}
		r := pos.X() + size.X()
		b := pos.Y() + size.Y()
		*bounds = matrix.Vec2{max(bounds.X(), r), max(bounds.Y(), b)}
	}
}

func (p *Panel) postLayoutUpdate() {
	if len(p.entity.Children) == 0 {
		return
	}
	if p.requestScrollX.requested {
		x := matrix.Clamp(p.requestScrollX.to, 0.0, p.maxScroll.X())
		p.scroll.SetX(x)
	}
	if p.requestScrollY.requested {
		y := matrix.Clamp(-p.requestScrollY.to, -p.maxScroll.Y(), 0)
		p.scroll.SetY(y)
	}
	offsetStart := matrix.Vec2{-p.scroll.X(), p.scroll.Y()}
	rows := make([]rowBuilder, 0)
	ps := p.layout.PixelSize()
	areaWidth := ps.X() - p.layout.padding.X() - p.layout.padding.Z() -
		p.layout.border.X() - p.layout.border.Z()
	for _, kid := range p.entity.Children {
		if !kid.IsActive() || kid.IsDestroyed() {
			continue
		}
		kui := FirstOnEntity(kid)
		if kui == nil {
			slog.Error("No UI component on entity")
			continue
		}
		kLayout := kui.Layout()
		switch kLayout.Positioning() {
		case PositioningAbsolute:
			if kLayout.Anchor().IsTop() {
				kLayout.rowLayoutOffset.SetY(p.layout.InnerOffset().Top() +
					p.layout.padding.Top() + p.layout.border.Top())
			} else if kLayout.Anchor().IsBottom() {
				kLayout.rowLayoutOffset.SetY(p.layout.InnerOffset().Bottom() +
					p.layout.padding.Bottom() + p.layout.border.Bottom())
			}
			if kLayout.Anchor().IsLeft() {
				kLayout.rowLayoutOffset.SetX(p.layout.InnerOffset().Left() +
					p.layout.padding.Left() + p.layout.border.Left())
			} else if kLayout.Anchor().IsRight() {
				kLayout.rowLayoutOffset.SetX(p.layout.InnerOffset().Right() +
					p.layout.padding.Right() + p.layout.border.Right())
			}
		case PositioningRelative:
			fallthrough
		case PositioningStatic:
			if len(rows) == 0 || !rows[len(rows)-1].addElement(areaWidth, kui) {
				rows = append(rows, rowBuilder{})
				rows[len(rows)-1].addElement(areaWidth, kui)
			}
		case PositioningFixed:
		case PositioningSticky:
		}
	}
	nextPos := offsetStart
	nextPos[matrix.Vy] += p.layout.padding.Y() + p.layout.border.Y()
	for i := range rows {
		rows[i].setElements(p.layout.padding.X()+p.layout.border.X(), nextPos[matrix.Vy])
		nextPos[matrix.Vy] += rows[i].Height()
	}
	if p.FittingContent() {
		bounds := matrix.Vec2{0, nextPos[matrix.Vy]}
		p.boundsChildren(&bounds)
		if p.fitContent == ContentFitWidth {
			p.layout.ScaleWidth(max(1, bounds.X()))
		} else if p.fitContent == ContentFitHeight {
			p.layout.ScaleHeight(max(1, bounds.Y()))
		} else if p.fitContent == ContentFitBoth {
			p.layout.Scale(max(1, bounds.X()), max(1, bounds.Y()))
		}
	}
	length := nextPos.Subtract(offsetStart)
	last := p.maxScroll
	ws := p.entity.Transform.WorldScale()
	p.maxScroll = matrix.Vec2{
		matrix.Max(0.0, length.X()-ws.X()),
		matrix.Max(0.0, length.Y()-ws.Y())}
	if !matrix.Vec2Approx(last, p.maxScroll) {
		p.SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) render() {
	p.uiBase.render()
	p.shaderData.setSize2d(p, p.textureSize.X(), p.textureSize.Y())
	p.requestScrollX.requested = false
	p.requestScrollY.requested = false
}

func (p *Panel) AddChild(target UI) {
	target.Entity().SetParent(p.entity)
	if p.group != nil {
		target.SetGroup(p.group)
	}
	target.Layout().update()
	p.SetDirty(DirtyTypeGenerated)
}

func (p *Panel) InsertChild(target UI, idx int) {
	p.AddChild(target)
	kidLen := len(p.entity.Children)
	idx = max(idx, 0)
	for i := idx; i < kidLen-1; i++ {
		p.entity.Children[i], p.entity.Children[kidLen-1] = p.entity.Children[kidLen-1], p.entity.Children[i]
	}
}

func (p *Panel) RemoveChild(target UI) {
	target.Entity().SetParent(nil)
	target.setScissor(matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax})
	target.Layout().update()
	p.layout.update()
	p.SetDirty(DirtyTypeGenerated)
}

func (p *Panel) Child(index int) UI {
	return FirstOnEntity(p.entity.Children[index])
}

func (p *Panel) SetSpeed(speed float32) {
	p.scrollSpeed = speed
}

func (p *Panel) recreateDrawing() {
	p.shaderData.Destroy()
	proxy := p.shaderData
	proxy.CancelDestroy()
	p.shaderData = proxy
}

func (p *Panel) removeDrawing() {
	p.recreateDrawing()
	p.drawing = rendering.Drawing{}
}

func (p *Panel) EnforceColor(color matrix.Color) {
	p.enforcedColorStack = append(p.enforcedColorStack, p.shaderData.FgColor)
	p.setColorInternal(color)
}

func (p *Panel) UnEnforceColor() {
	if !p.HasEnforcedColor() {
		return
	}
	last := len(p.enforcedColorStack) - 1
	p.setColorInternal(p.enforcedColorStack[last])
	p.enforcedColorStack = p.enforcedColorStack[:last]
}

func (p *Panel) SetColor(bgColor matrix.Color) {
	if p.HasEnforcedColor() {
		p.enforcedColorStack[0] = bgColor
		return
	}
	p.setColorInternal(bgColor)
}

func (p *Panel) SetScrollX(value float32) {
	p.requestScrollX.to = value
	p.requestScrollX.requested = true
	p.SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetScrollY(value float32) {
	p.requestScrollY.to = value
	p.requestScrollY.requested = true
	p.SetDirty(DirtyTypeLayout)
}

func (p *Panel) ResetScroll() {
	p.scroll = matrix.Vec2Zero()
}

func (p *Panel) ensureBGExists(tex *rendering.Texture) {
	if !p.drawing.IsValid() {
		if tex == nil {
			tex, _ = p.host.TextureCache().Texture(
				assets.TextureSquare, rendering.TextureFilterLinear)
		}
		shader := p.host.ShaderCache().ShaderFromDefinition(
			assets.ShaderDefinitionUI)
		p.shaderData.BorderLen = matrix.Vec2{8.0, 8.0}
		p.shaderData.BgColor = p.color
		p.shaderData.FgColor = p.color
		p.shaderData.UVs = matrix.Vec4{0.0, 0.0, 1.0, 1.0}
		p.shaderData.Size2D = matrix.Vec4{0.0, 0.0,
			float32(tex.Width), float32(tex.Height)}
		p.textureSize = tex.Size()
		p.shaderData.setSize2d(p, p.textureSize.X(), p.textureSize.Y())
		p.drawing = rendering.Drawing{
			Renderer:   p.host.Window.Renderer,
			Shader:     shader,
			Mesh:       rendering.NewMeshQuad(p.host.MeshCache()),
			Textures:   []*rendering.Texture{tex},
			ShaderData: &p.shaderData,
			Transform:  &p.entity.Transform,
			CanvasId:   "default",
		}
		p.host.Drawings.AddDrawing(&p.drawing)
	} else if tex != nil {
		p.SetBackground(tex)
	}
}

func (p *Panel) Background() *rendering.Texture {
	if p.drawing.IsValid() {
		return p.drawing.Textures[0]
	}
	return nil
}

func (p *Panel) SetBackground(tex *rendering.Texture) {
	if p.drawing.IsValid() {
		p.recreateDrawing()
		p.drawing.Textures[0] = tex
		p.host.Drawings.AddDrawing(&p.drawing)
	}
}

func (p *Panel) RemoveBackground() {
	p.recreateDrawing()
}

func (p *Panel) IsScrolling() bool {
	return p.isScrolling
}

func (p *Panel) Freeze() {
	p.frozen = true
}

func (p *Panel) UnFreeze() {
	p.frozen = false
}

func (p *Panel) IsFrozen() bool {
	return p.frozen
}

func (p *Panel) SetScrollDirection(direction PanelScrollDirection) {
	p.scrollDirection = direction
	p.SetDirty(DirtyTypeLayout)
	if p.scrollDirection == PanelScrollDirectionNone {
		if p.scrollEvent != 0 {
			p.RemoveEvent(EventTypeScroll, p.scrollEvent)
			p.scrollEvent = 0
		}
	} else if p.scrollEvent == 0 {
		p.scrollEvent = p.AddEvent(EventTypeScroll, p.onScroll)
	}
}

func (p *Panel) ScrollDirection() PanelScrollDirection { return p.scrollDirection }
func (p *Panel) BorderSize() matrix.Vec4               { return p.layout.Border() }
func (p *Panel) BorderStyle() [4]BorderStyle           { return p.borderStyle }

func (p *Panel) BorderColor() [4]matrix.Color {
	return p.shaderData.BorderColor
}

func (p *Panel) SetBorderRadius(topLeft, topRight, bottomRight, bottomLeft float32) {
	p.shaderData.BorderRadius = matrix.Vec4{
		topLeft, topRight, bottomRight, bottomLeft}
}

func (p *Panel) SetBorderSize(left, top, right, bottom float32) {
	p.layout.SetBorder(left, top, right, bottom)
	// TODO:  If there isn't a border, it should be transparent when created
	p.ensureBGExists(nil)
	p.shaderData.BorderSize = p.layout.Border()
}

func (p *Panel) SetBorderStyle(left, top, right, bottom BorderStyle) {
	p.borderStyle = [4]BorderStyle{left, top, right, bottom}
}

func (p *Panel) SetBorderColor(left, top, right, bottom matrix.Color) {
	p.shaderData.BorderColor = [4]matrix.Color{left, top, right, bottom}
}

func (p *Panel) SetUseBlending(useBlending bool) {
	p.recreateDrawing()
	p.drawing.UseBlending = useBlending
	p.host.Drawings.AddDrawing(&p.drawing)
}

func (p *Panel) Overflow() Overflow { return p.overflow }

func (p *Panel) SetOverflow(overflow Overflow) {
	if p.overflow != overflow {
		p.overflow = overflow
		p.SetDirty(DirtyTypeLayout)
	}
}

func (p *Panel) HasEnforcedColor() bool { return len(p.enforcedColorStack) > 0 }

func (p *Panel) setColorInternal(bgColor matrix.Color) {
	if p.shaderData.FgColor.Equals(bgColor) {
		return
	}
	p.ensureBGExists(nil)
	hasBlending := p.shaderData.FgColor.A() < 1.0
	shouldBlend := bgColor.A() < 1.0
	if hasBlending != shouldBlend {
		p.recreateDrawing()
		p.drawing.UseBlending = shouldBlend
		p.host.Drawings.AddDrawing(&p.drawing)
	}
	p.shaderData.FgColor = bgColor
}
