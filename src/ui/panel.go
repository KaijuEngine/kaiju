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

type requestScroll struct {
	to        float32
	requested bool
}

type panelData struct {
	scroll, offset, maxScroll matrix.Vec2
	scrollSpeed               float32
	scrollDirection           PanelScrollDirection
	scrollEvent               events.Id
	borderStyle               [4]BorderStyle
	drawing                   rendering.Drawing
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

func (p *panelData) innerPanelData() *panelData { return p }

type Panel UI

func (u *UI) ToPanel() *Panel { return (*Panel)(u) }
func (p *Panel) Base() *UI    { return (*UI)(p) }

func (p *Panel) PanelData() *panelData { return p.elmData.innerPanelData() }

func (panel *Panel) Init(texture *rendering.Texture, anchor Anchor, elmType ElementType) {
	var pd *panelData
	panel.elmType = elmType
	if panel.elmData == nil {
		panel.elmData = &panelData{}
	}
	pd = panel.elmData.innerPanelData()
	pd.scrollEvent = 0
	pd.scrollSpeed = 20.0
	pd.scrollDirection = PanelScrollDirectionVertical
	pd.fitContent = ContentFitBoth
	pd.enforcedColorStack = make([]matrix.Color, 0)
	panel.postLayoutUpdate = panel.panelPostLayoutUpdate
	panel.render = panel.panelRender
	ts := matrix.Vec2Zero()
	if texture != nil {
		ts = texture.Size()
	}
	base := panel.Base()
	base.init(ts, anchor)
	panel.shaderData.FgColor = matrix.Color{1.0, 1.0, 1.0, 1.0}
	panel.entity.SetChildrenOrdered()
	if texture != nil {
		panel.ensureBGExists(texture)
	}
	panel.entity.OnActivate.Add(func() {
		panel.shaderData.Activate()
		base.SetDirty(DirtyTypeLayout)
		base.Clean()
	})
	panel.entity.OnDeactivate.Add(func() {
		panel.shaderData.Deactivate()
	})
}

func (p *Panel) MaxScroll() matrix.Vec2 { return p.PanelData().maxScroll }
func (p *Panel) ScrollX() float32       { return p.PanelData().scroll.X() }
func (p *Panel) ScrollY() float32       { return -p.PanelData().scroll.Y() }
func (p *Panel) EnableDragScroll()      { p.PanelData().allowDragScroll = true }
func (p *Panel) DisableDragScroll()     { p.PanelData().allowDragScroll = false }

func (p *Panel) DontFitContentWidth() {
	pd := p.PanelData()
	switch pd.fitContent {
	case ContentFitBoth:
		pd.fitContent = ContentFitHeight
	case ContentFitWidth:
		pd.fitContent = ContentFitNone
	}
}

func (p *Panel) DontFitContentHeight() {
	pd := p.PanelData()
	switch pd.fitContent {
	case ContentFitBoth:
		pd.fitContent = ContentFitWidth
	case ContentFitHeight:
		pd.fitContent = ContentFitNone
	}
}

func (p *Panel) DontFitContent() {
	p.PanelData().fitContent = ContentFitNone
}

func (p *Panel) FittingContent() bool {
	return p.PanelData().fitContent != ContentFitNone
}

func (p *Panel) FitContentWidth() {
	pd := p.PanelData()
	switch pd.fitContent {
	case ContentFitNone:
		pd.fitContent = ContentFitWidth
	case ContentFitHeight:
		pd.fitContent = ContentFitBoth
	}
	if p.dirtyType == DirtyTypeNone {
		p.Base().SetDirty(DirtyTypeLayout)
	} else {
		p.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) FitContentHeight() {
	pd := p.PanelData()
	switch pd.fitContent {
	case ContentFitNone:
		pd.fitContent = ContentFitHeight
	case ContentFitWidth:
		pd.fitContent = ContentFitBoth
	}
	if p.dirtyType == DirtyTypeNone {
		p.Base().SetDirty(DirtyTypeLayout)
	} else {
		p.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) FitContent() {
	p.PanelData().fitContent = ContentFitBoth
	if p.dirtyType == DirtyTypeNone {
		p.Base().SetDirty(DirtyTypeLayout)
	} else {
		p.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) onScroll() {
	pd := p.PanelData()
	mouse := &p.man.Host.Window.Mouse
	delta := mouse.Scroll()
	scroll := pd.scroll
	base := p.Base()
	if !mouse.Scrolled() {
		pos := base.cursorPos(&p.man.Host.Window.Cursor)
		delta = pos.Subtract(p.downPos)
		delta[matrix.Vy] *= -1.0
	} else {
		pd.offset = pd.scroll
		delta.ScaleAssign(pd.scrollSpeed)
	}
	if (pd.scrollDirection & PanelScrollDirectionHorizontal) != 0 {
		x := matrix.Clamp(delta.X()+pd.offset.X(), 0.0, pd.maxScroll.X())
		scroll.SetX(x)
	}
	if (pd.scrollDirection & PanelScrollDirectionVertical) != 0 {
		y := matrix.Clamp(delta.Y()+pd.offset.Y(), -pd.maxScroll.Y(), 0)
		scroll.SetY(y)
	}
	if !matrix.Vec2Approx(scroll, pd.scroll) {
		pd.scroll = scroll
		base.SetDirty(DirtyTypeLayout)
		pd.isScrolling = true
	}
}

func panelOnDown(ui *UI) {
	var target *UI = ui
	ok := false
	var panel *Panel
	for !ok {
		target = FirstOnEntity(target.Entity().Parent)
		if target.elmType == ElementTypePanel {
			panel = target.ToPanel()
		}
	}
	pd := panel.PanelData()
	pd.offset = pd.scroll
	pd.dragging = true
	if !pd.allowDragScroll {
		// TODO:  Change the mouse cursor to look like it's dragging something
	}
}

func (p *Panel) update(deltaTime float64) {
	base := p.Base()
	base.eventUpdates()
	base.Update(deltaTime)
	pd := p.PanelData()
	if !pd.frozen {
		if p.isDown && pd.dragging {
			if pd.allowDragScroll {
				p.onScroll()
			}
		} else if pd.dragging {
			pd.dragging = false
		} else {
			pd.isScrolling = false
		}
	}
}

type rowBuilder struct {
	elements        []*UI
	maxMarginTop    float32
	maxMarginBottom float32
	x               float32
	height          float32
}

func (rb *rowBuilder) addElement(areaWidth float32, e *UI) bool {
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
		kui := FirstOnEntity(kid)
		if kui.layout.screenAnchor.IsStretch() {
			continue
		}
		pos := kid.Transform.Position()
		if kui.Layout().Positioning() == PositioningAbsolute {
			continue
		}
		var size matrix.Vec2
		if kui.elmType == ElementTypeLabel {
			size = kui.ToLabel().Measure()
			// Give a little margin for error on text
			size[matrix.Vx] += 0.1
		} else {
			size = kid.Transform.WorldScale().AsVec2()
			kui.ToPanel().boundsChildren(bounds)
		}
		r := pos.X() + size.X()
		b := pos.Y() + size.Y()
		*bounds = matrix.Vec2{max(bounds.X(), r), max(bounds.Y(), b)}
	}
}

func (p *Panel) panelPostLayoutUpdate() {
	if len(p.entity.Children) == 0 {
		return
	}
	pd := p.PanelData()
	if pd.requestScrollX.requested {
		x := matrix.Clamp(pd.requestScrollX.to, 0, pd.maxScroll.X())
		pd.scroll.SetX(x)
		pd.requestScrollX.requested = false
	}
	if pd.requestScrollY.requested {
		y := matrix.Clamp(-pd.requestScrollY.to, -pd.maxScroll.Y(), 0)
		pd.scroll.SetY(y)
		pd.requestScrollY.requested = false
	}
	offsetStart := matrix.Vec2{-pd.scroll.X(), pd.scroll.Y()}
	rows := make([]rowBuilder, 0)
	ps := p.layout.PixelSize()
	areaWidth := ps.X() - p.layout.padding.X() - p.layout.padding.Z() -
		p.layout.border.X() - p.layout.border.Z()
	maxSize := matrix.Vec2{}
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
			kws := kid.Transform.WorldScale()
			maxSize[matrix.Vx] = max(maxSize.X(), kLayout.left+kLayout.offset.X()+kws.Width())
			maxSize[matrix.Vy] = max(maxSize.Y(), kLayout.top+kLayout.offset.Y()+kws.Height())
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
	addY := p.layout.padding.Y() + p.layout.border.Y()
	nextPos[matrix.Vy] += addY
	maxSize[matrix.Vy] += addY
	for i := range rows {
		rows[i].setElements(p.layout.padding.X()+p.layout.border.X(), nextPos[matrix.Vy])
		addY = rows[i].height + rows[i].maxMarginTop + rows[i].maxMarginBottom
		nextPos[matrix.Vy] += addY
		maxSize[matrix.Vy] += addY
	}
	bounds := matrix.Vec2{maxSize.X(), maxSize.Y()}
	if p.FittingContent() {
		p.boundsChildren(&bounds)
		border := p.layout.border
		if pd.fitContent == ContentFitWidth {
			p.layout.ScaleWidth(max(1, bounds.X()+border.Left()+border.Right()))
		} else if pd.fitContent == ContentFitHeight {
			p.layout.ScaleHeight(max(1, bounds.Y()+border.Top()+border.Bottom()))
		} else if pd.fitContent == ContentFitBoth {
			p.layout.Scale(max(1, bounds.X()+border.Left()+border.Right()),
				max(1, bounds.Y()+border.Top()+border.Bottom()))
		}
	}
	last := pd.maxScroll
	ws := p.entity.Transform.WorldScale()
	pd.maxScroll = matrix.NewVec2(max(0, bounds.X()-ws.X()), max(0.0, bounds.Y()-ws.Y()))
	if !matrix.Vec2Roughly(last, pd.maxScroll) {
		p.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) panelRender() {
	pd := p.PanelData()
	//p.Base().render() ---v
	p.events[EventTypeRender].Execute()
	p.shaderData.setSize2d(p.Base(), p.textureSize.X(), p.textureSize.Y())
	pd.requestScrollX.requested = false
	pd.requestScrollY.requested = false
}

func (p *Panel) AddChild(target *UI) {
	target.Entity().SetParent(&p.entity)
	// No need to set the group on the target as it's set by the UI Manager
	target.Layout().update()
	p.Base().SetDirty(DirtyTypeGenerated)
}

func (p *Panel) InsertChild(target *UI, idx int) {
	p.AddChild(target)
	kidLen := len(p.entity.Children)
	idx = max(idx, 0)
	for i := idx; i < kidLen-1; i++ {
		p.entity.Children[i], p.entity.Children[kidLen-1] = p.entity.Children[kidLen-1], p.entity.Children[i]
	}
}

func (p *Panel) RemoveChild(target *UI) {
	target.Entity().SetParent(nil)
	target.setScissor(matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax})
	target.Layout().update()
	p.layout.update()
	p.Base().SetDirty(DirtyTypeGenerated)
}

func (p *Panel) Child(index int) *UI {
	return FirstOnEntity(p.entity.Children[index])
}

func (p *Panel) SetSpeed(speed float32) {
	p.PanelData().scrollSpeed = speed
}

func (p *Panel) recreateDrawing() {
	p.shaderData.Destroy()
	proxy := *p.shaderData
	proxy.CancelDestroy()
	p.shaderData = &ShaderData{}
	*p.shaderData = proxy
	p.PanelData().drawing.ShaderData = p.shaderData
}

func (p *Panel) removeDrawing() {
	p.recreateDrawing()
	p.PanelData().drawing = rendering.Drawing{}
}

func (p *Panel) EnforceColor(color matrix.Color) {
	pd := p.PanelData()
	pd.enforcedColorStack = append(pd.enforcedColorStack, p.shaderData.FgColor)
	p.setColorInternal(color)
}

func (p *Panel) UnEnforceColor() {
	if !p.HasEnforcedColor() {
		return
	}
	pd := p.PanelData()
	last := len(pd.enforcedColorStack) - 1
	p.setColorInternal(pd.enforcedColorStack[last])
	pd.enforcedColorStack = pd.enforcedColorStack[:last]
}

func (p *Panel) SetColor(bgColor matrix.Color) {
	if p.HasEnforcedColor() {
		p.PanelData().enforcedColorStack[0] = bgColor
		return
	}
	p.setColorInternal(bgColor)
}

func (p *Panel) SetScrollX(value float32) {
	pd := p.PanelData()
	pd.requestScrollX.to = value
	pd.requestScrollX.requested = true
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetScrollY(value float32) {
	pd := p.PanelData()
	pd.requestScrollY.to = value
	pd.requestScrollY.requested = true
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) ResetScroll() {
	p.PanelData().scroll = matrix.Vec2Zero()
}

func (p *Panel) ensureBGExists(tex *rendering.Texture) {
	pd := p.PanelData()
	if !pd.drawing.IsValid() {
		if tex == nil {
			tex, _ = p.man.Host.TextureCache().Texture(
				assets.TextureSquare, rendering.TextureFilterLinear)
		}
		shader := p.man.Host.ShaderCache().ShaderFromDefinition(
			assets.ShaderDefinitionUI)
		p.shaderData.BorderLen = matrix.Vec2{8.0, 8.0}
		p.shaderData.UVs = matrix.Vec4{0.0, 0.0, 1.0, 1.0}
		p.shaderData.Size2D = matrix.Vec4{0.0, 0.0,
			float32(tex.Width), float32(tex.Height)}
		p.textureSize = tex.Size()
		p.shaderData.setSize2d(p.Base(), p.textureSize.X(), p.textureSize.Y())
		pd.drawing = rendering.Drawing{
			Renderer:   p.man.Host.Window.Renderer,
			Shader:     shader,
			Mesh:       rendering.NewMeshQuad(p.man.Host.MeshCache()),
			Textures:   []*rendering.Texture{tex},
			ShaderData: p.shaderData,
			Transform:  &p.entity.Transform,
			CanvasId:   "default",
		}
		p.man.Host.Drawings.AddDrawing(&pd.drawing)
	} else if tex != nil {
		p.SetBackground(tex)
	}
}

func (p *Panel) Background() *rendering.Texture {
	pd := p.PanelData()
	if pd.drawing.IsValid() {
		return pd.drawing.Textures[0]
	}
	return nil
}

func (p *Panel) SetBackground(tex *rendering.Texture) {
	pd := p.PanelData()
	if pd.drawing.IsValid() {
		p.recreateDrawing()
		pd.drawing.Textures[0] = tex
		p.man.Host.Drawings.AddDrawing(&pd.drawing)
	}
}

func (p *Panel) RemoveBackground() {
	p.recreateDrawing()
}

func (p *Panel) IsScrolling() bool {
	return p.PanelData().isScrolling
}

func (p *Panel) Freeze() {
	p.PanelData().frozen = true
}

func (p *Panel) UnFreeze() {
	p.PanelData().frozen = false
}

func (p *Panel) IsFrozen() bool {
	return p.PanelData().frozen
}

func (p *Panel) SetScrollDirection(direction PanelScrollDirection) {
	pd := p.PanelData()
	pd.scrollDirection = direction
	p.Base().SetDirty(DirtyTypeLayout)
	if pd.scrollDirection == PanelScrollDirectionNone {
		if pd.scrollEvent != 0 {
			p.Base().RemoveEvent(EventTypeScroll, pd.scrollEvent)
			pd.scrollEvent = 0
		}
	} else if pd.scrollEvent == 0 {
		pd.scrollEvent = p.Base().AddEvent(EventTypeScroll, p.onScroll)
	}
}

func (p *Panel) ScrollDirection() PanelScrollDirection {
	return p.PanelData().scrollDirection
}

func (p *Panel) BorderSize() matrix.Vec4     { return p.layout.Border() }
func (p *Panel) BorderStyle() [4]BorderStyle { return p.PanelData().borderStyle }

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
	p.PanelData().borderStyle = [4]BorderStyle{left, top, right, bottom}
}

func (p *Panel) SetBorderColor(left, top, right, bottom matrix.Color) {
	p.shaderData.BorderColor = [4]matrix.Color{left, top, right, bottom}
}

func (p *Panel) SetUseBlending(useBlending bool) {
	p.recreateDrawing()
	pd := p.PanelData()
	pd.drawing.UseBlending = useBlending
	p.man.Host.Drawings.AddDrawing(&pd.drawing)
}

func (p *Panel) Overflow() Overflow { return p.PanelData().overflow }

func (p *Panel) SetOverflow(overflow Overflow) {
	pd := p.PanelData()
	if pd.overflow != overflow {
		pd.overflow = overflow
		p.Base().SetDirty(DirtyTypeLayout)
	}
}

func (p *Panel) HasEnforcedColor() bool {
	return len(p.PanelData().enforcedColorStack) > 0
}

func (p *Panel) setColorInternal(bgColor matrix.Color) {
	if p.shaderData.FgColor.Equals(bgColor) {
		return
	}
	p.ensureBGExists(nil)
	hasBlending := p.shaderData.FgColor.A() < 1.0
	shouldBlend := bgColor.A() < 1.0
	if hasBlending != shouldBlend {
		p.recreateDrawing()
		pd := p.PanelData()
		pd.drawing.UseBlending = shouldBlend
		p.man.Host.Drawings.AddDrawing(&pd.drawing)
	}
	p.shaderData.FgColor = bgColor
}
