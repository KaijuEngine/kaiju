/******************************************************************************/
/* panel.go                                                                   */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/engine/assets"
	"kaiju/engine/systems/events"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"log/slog"
)

type PanelScrollDirection = int32
type BorderStyle = int32
type ContentFit = int32
type Overflow = int
type panelBits uint8

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
	scrollBarWidth = 8
)

const (
	OverflowScroll = iota
	OverflowVisible
	OverflowHidden
)

const (
	panelBitsIsScrolling panelBits = 1 << iota
	panelBitsIsDragging
	panelBitsIsFrozen
	panelBitsAllowDragScroll
	panelBitsAllowClickThrough
)

var UIScrollSpeed float32 = 20

type requestScroll struct {
	to        float32
	requested bool
}

type panelData struct {
	scrollBarX, scrollBarY    *Panel
	scrollBarStart            float32
	scrollBarDrag             matrix.Vec2
	scroll, offset, maxScroll matrix.Vec2
	scrollDirection           PanelScrollDirection
	scrollEvent               events.Id
	borderStyle               [4]BorderStyle
	drawing                   rendering.Drawing
	transparentDrawing        rendering.Drawing
	fitContent                ContentFit
	requestScrollX            requestScroll
	requestScrollY            requestScroll
	overflow                  Overflow
	enforcedColorStack        []matrix.Color
	flags                     panelBits
}

func (b panelBits) isScrolling() bool        { return b&panelBitsIsScrolling != 0 }
func (b panelBits) isDragging() bool         { return b&panelBitsIsDragging != 0 }
func (b panelBits) isFrozen() bool           { return b&panelBitsIsFrozen != 0 }
func (b panelBits) allowDragScroll() bool    { return b&panelBitsAllowDragScroll != 0 }
func (b panelBits) allowClickThrough() bool  { return b&panelBitsAllowClickThrough != 0 }
func (b *panelBits) setIsScrolling()         { *b |= panelBitsIsScrolling }
func (b *panelBits) setDragging()            { *b |= panelBitsIsDragging }
func (b *panelBits) setFrozen()              { *b |= panelBitsIsFrozen }
func (b *panelBits) setAllowDragScroll()     { *b |= panelBitsAllowDragScroll }
func (b *panelBits) setAllowClickThrough()   { *b |= panelBitsAllowClickThrough }
func (b *panelBits) resetIsScrolling()       { *b &= ^panelBitsIsScrolling }
func (b *panelBits) resetDragging()          { *b &= ^panelBitsIsDragging }
func (b *panelBits) resetFrozen()            { *b &= ^panelBitsIsFrozen }
func (b *panelBits) resetAllowDragScroll()   { *b &= ^panelBitsAllowDragScroll }
func (b *panelBits) resetAllowClickThrough() { *b &= ^panelBitsAllowClickThrough }

func (p *panelData) innerPanelData() *panelData { return p }

type Panel UI

func (u *UI) ToPanel() *Panel { return (*Panel)(u) }
func (p *Panel) Base() *UI    { return (*UI)(p) }

func (p *Panel) PanelData() *panelData { return p.elmData.innerPanelData() }

func (panel *Panel) Init(texture *rendering.Texture, elmType ElementType) {
	defer tracing.NewRegion("Panel.Init").End()
	var pd *panelData
	panel.elmType = elmType
	if panel.elmData == nil {
		panel.elmData = &panelData{}
	}
	pd = panel.elmData.innerPanelData()
	pd.scrollEvent = 0
	pd.scrollDirection = PanelScrollDirectionNone
	pd.fitContent = ContentFitBoth
	pd.enforcedColorStack = make([]matrix.Color, 0)
	panel.postLayoutUpdate = panel.panelPostLayoutUpdate
	panel.render = panel.panelRender
	ts := matrix.Vec2Zero()
	if texture != nil {
		ts = texture.Size()
	}
	base := panel.Base()
	base.init(ts)
	panel.shaderData.FgColor = matrix.Color{1.0, 1.0, 1.0, 1.0}
	panel.entity.SetChildrenOrdered()
	if texture != nil {
		panel.ensureBGExists(texture)
	}
	panel.entity.OnActivate.Add(func() {
		panel.shaderData.Activate()
		base.SetDirty(DirtyTypeLayout)
	})
	panel.entity.OnDeactivate.Add(func() { panel.shaderData.Deactivate() })
}

func (p *Panel) MaxScroll() matrix.Vec2 { return p.PanelData().maxScroll }
func (p *Panel) ScrollX() float32       { return p.PanelData().scroll.X() }
func (p *Panel) ScrollY() float32       { return -p.PanelData().scroll.Y() }
func (p *Panel) EnableDragScroll()      { p.PanelData().flags.setAllowDragScroll() }
func (p *Panel) DisableDragScroll()     { p.PanelData().flags.resetAllowDragScroll() }

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

func (p *Panel) FittingContentWidth() bool {
	return (p.PanelData().fitContent & ContentFitWidth) != 0
}

func (p *Panel) FittingContentHeight() bool {
	return (p.PanelData().fitContent & ContentFitHeight) != 0
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
	defer tracing.NewRegion("Panel.onScroll").End()
	pd := p.PanelData()
	host := p.man.Value().Host
	mouse := &host.Window.Mouse
	delta := mouse.Scroll()
	scroll := pd.scroll
	base := p.Base()
	if !mouse.Scrolled() {
		pos := base.cursorPos(&host.Window.Cursor)
		delta = pos.Subtract(p.downPos)
		delta[matrix.Vy] *= -1.0
	} else {
		pd.offset = pd.scroll
		delta.ScaleAssign(UIScrollSpeed)
	}
	// If the panel can only scroll horizontally, use the Y scroll if there is no X
	if pd.scrollDirection == PanelScrollDirectionHorizontal {
		if matrix.ApproxTo(delta.X(), 0, matrix.Tiny) {
			delta.SetX(-delta.Y())
		}
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
		pd.flags.setIsScrolling()
	}
}

func (p *Panel) update(deltaTime float64) {
	defer tracing.NewRegion("Panel.update").End()
	base := p.Base()
	if !base.IsActive() {
		return
	}
	base.eventUpdates()
	base.Update(deltaTime)
	pd := p.PanelData()
	if base.Host().Window.Cursor.Released() {
		pd.scrollBarDrag = matrix.Vec2Zero()
		pd.scrollBarStart = -1
	}
	p.updateScrollBars()
	if !pd.flags.isFrozen() {
		if p.flags.isDown() && pd.flags.isDragging() {
			if pd.flags.allowClickThrough() {
				p.onScroll()
			}
		} else if pd.flags.isDragging() {
			pd.flags.resetDragging()
		} else {
			pd.flags.resetIsScrolling()
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
	h := eSize.Height()
	w := eSize.Width() + e.layout.margin.Horizontal()
	if len(rb.elements) > 0 && rb.x+w > areaWidth {
		return false
	}
	rb.elements = append(rb.elements, e)
	rb.maxMarginTop = matrix.Max(rb.maxMarginTop, e.Layout().margin.Y())
	rb.maxMarginBottom = matrix.Max(rb.maxMarginBottom, e.Layout().margin.W())
	rb.x += w
	rb.height = matrix.Max(rb.height, h)
	return true
}

func (rb rowBuilder) Width() float32 {
	return rb.x
}

func (rb rowBuilder) Height() float32 {
	return rb.height + rb.maxMarginTop + rb.maxMarginBottom
}

func (rb rowBuilder) setElements(offsetX, offsetY float32) {
	defer tracing.NewRegion("Panel.Init").End()
	for _, e := range rb.elements {
		layout := e.Layout()
		x, y := offsetX, offsetY
		switch e.Layout().Positioning() {
		case PositioningAbsolute:
			fallthrough
		case PositioningRelative:
			x += layout.InnerOffset().Left()
		}
		x += layout.margin.X()
		y += rb.maxMarginTop
		layout.SetRowLayoutOffset(matrix.Vec2{x, y})
		offsetX += layout.PixelSize().Width() + layout.margin.X() + layout.margin.Z()
	}
}

func (p *Panel) boundsChildren(bounds *matrix.Vec2) {
	defer tracing.NewRegion("Panel.boundsChildren").End()
	for _, kid := range p.entity.Children {
		kui := FirstOnEntity(kid)
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
		//pos := kid.Transform.Position()
		//r := pos.X() + size.X()
		//b := pos.Y() + size.Y()
		r := size.X()
		b := size.Y()
		*bounds = matrix.Vec2{max(bounds.X(), r), max(bounds.Y(), b)}
	}
}

func (p *Panel) panelPostLayoutUpdate() {
	defer tracing.NewRegion("Panel.panelPostLayoutUpdate").End()
	if !p.Base().IsActive() {
		return
	}
	if p.PanelData().drawing.IsValid() {
		p.shaderData.setSize2d(p.Base())
	}
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
	areaWidth := ps.X()
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
	nextPos[matrix.Vx] += p.layout.padding.Left() + p.layout.border.Left()
	addY := p.layout.padding.Top() + p.layout.border.Top()
	nextPos[matrix.Vy] += addY
	maxSize[matrix.Vy] += addY
	maxRowsX := matrix.Float(0)
	for i := range rows {
		rows[i].setElements(nextPos[matrix.Vx], nextPos[matrix.Vy])
		addY = rows[i].height + rows[i].maxMarginTop + rows[i].maxMarginBottom
		maxRowsX = max(maxRowsX, rows[i].x)
		nextPos[matrix.Vy] += addY
		maxSize[matrix.Vy] += addY
	}
	bounds := matrix.Vec2{maxSize.X(), maxSize.Y()}
	if p.FittingContent() {
		p.boundsChildren(&bounds)
		w := bounds.X() + p.layout.padding.Horizontal() + p.layout.border.Horizontal()
		h := bounds.Y() + p.layout.padding.Bottom() + p.layout.border.Bottom()
		switch pd.fitContent {
		case ContentFitWidth:
			p.layout.ScaleWidth(max(1, w))
		case ContentFitHeight:
			p.layout.ScaleHeight(max(1, h))
		case ContentFitBoth:
			p.layout.Scale(max(1, w), max(1, h))
		}
	} else {
		bounds.SetX(maxRowsX)
	}
	last := pd.maxScroll
	ws := p.entity.Transform.WorldScale()
	yScroll := bounds.Y() - ws.Y()
	if pd.scrollBarX != nil {
		yScroll += scrollBarWidth
	}
	pd.maxScroll = matrix.NewVec2(max(0, bounds.X()-ws.X()), max(0.0, yScroll))
	if !matrix.Vec2Roughly(last, pd.maxScroll) {
		if pd.scrollBarX != nil {
			pd.scrollBarX.Base().Show()
		}
		if pd.scrollBarY != nil {
			pd.scrollBarY.Base().Show()
		}
		p.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) panelRender() {
	pd := p.PanelData()
	//p.Base().render() ---v
	p.events[EventTypeRender].Execute()
	if pd.drawing.IsValid() {
		p.shaderData.setSize2d(p.Base())
	}
	pd.requestScrollX.requested = false
	pd.requestScrollY.requested = false
}

func (p *Panel) AddChild(target *UI) {
	if target.entity.Parent == &p.entity {
		return
	}
	if target.entity.Parent != nil {
		FirstPanelOnEntity(target.entity.Parent).RemoveChild(target)
	}
	target.Entity().SetParent(&p.entity)
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
	p.Base().SetDirty(DirtyTypeGenerated)
	target.SetDirty(DirtyTypeGenerated)
}

func (p *Panel) Child(index int) *UI {
	return FirstOnEntity(p.entity.Children[index])
}

func (p *Panel) recreateDrawing() {
	p.shaderData.Destroy()
	proxy := *p.shaderData
	proxy.CancelDestroy()
	p.shaderData = &ShaderData{}
	*p.shaderData = proxy
	p.PanelData().drawing.ShaderData = p.shaderData
	p.PanelData().transparentDrawing.ShaderData = p.shaderData
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

func (p *Panel) Color() matrix.Color { return p.shaderData.FgColor }

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

func (p *Panel) ScrollToChild(child *UI) {
	pps := p.Base().Layout().PixelSize()
	cps := child.layout.PixelSize()
	y := child.entity.Transform.Position().Y()
	parent := child.entity.Parent
	for parent != &p.entity {
		y += parent.Transform.Position().Y()
		parent = parent.Parent
		if parent == nil {
			slog.Error("invalid child supplied to ScrollToChild, it is not a child or grandchild")
			return
		}
	}
	top := pps.Y()*0.5 - cps.Y()*0.5 - y
	bottom := -(pps.Y()*0.5 - cps.Y()*0.5) - y
	if top < 0 {
		p.SetScrollY(p.ScrollY() + top)
	} else if bottom >= 0 {
		p.SetScrollY(p.ScrollY() + bottom)
	}
}

func (p *Panel) ResetScroll() {
	p.PanelData().scroll = matrix.Vec2Zero()
}

func (p *Panel) ensureBGExists(tex *rendering.Texture) {
	defer tracing.NewRegion("Panel.ensureBGExists").End()
	pd := p.PanelData()
	host := p.man.Value().Host
	if !pd.drawing.IsValid() {
		if tex == nil {
			tex, _ = host.TextureCache().Texture(
				assets.TextureSquare, rendering.TextureFilterLinear)
		}
		tex.MipLevels = 1
		material, err := host.MaterialCache().Material(assets.MaterialDefinitionUI)
		if err != nil {
			slog.Error("failed to load the ui material for panel", "error", err)
			return
		}
		p.shaderData.BorderLen = matrix.Vec2{8.0, 8.0}
		p.shaderData.UVs = matrix.Vec4{0.0, 0.0, 1.0, 1.0}
		p.shaderData.Size2D = matrix.Vec4{0.0, 0.0,
			float32(tex.Width), float32(tex.Height)}
		p.textureSize = tex.Size()
		p.shaderData.resetSize2D(p.Base())
		material = material.CreateInstance([]*rendering.Texture{tex})
		pd.drawing = rendering.Drawing{
			Material:   material,
			Mesh:       rendering.NewMeshQuad(host.MeshCache()),
			ShaderData: p.shaderData,
			Transform:  &p.entity.Transform,
			ViewCuller: &host.Cameras.UI,
		}
		host.Drawings.AddDrawing(pd.drawing)
	} else if tex != nil {
		p.SetBackground(tex)
		p.textureSize = tex.Size()
		p.shaderData.setSize2d(p.Base())
	}
	// TODO:  Allow this to be overridable for transparent overlays?
	// Panels that have a background shouldn't be click-through-able (probably)
	if !pd.flags.allowClickThrough() {
		if p.events[EventTypeDown].IsEmpty() {
			p.Base().AddEvent(EventTypeDown, func() { /* Do nothing, but block things */ })
		}
		if p.events[EventTypeUp].IsEmpty() {
			p.Base().AddEvent(EventTypeUp, func() { /* Do nothing, but block things */ })
		}
		if p.events[EventTypeRightDown].IsEmpty() {
			p.Base().AddEvent(EventTypeRightDown, func() { /* Do nothing, but block things */ })
		}
		if p.events[EventTypeRightUp].IsEmpty() {
			p.Base().AddEvent(EventTypeRightUp, func() { /* Do nothing, but block things */ })
		}
	}
}

func (p *Panel) HasBackground() bool { return p.PanelData().drawing.IsValid() }

func (p *Panel) Background() *rendering.Texture {
	pd := p.PanelData()
	if pd.drawing.IsValid() {
		return pd.drawing.Material.Textures[0]
	}
	return nil
}

func (p *Panel) SetMaterial(mat *rendering.Material) {
	defer tracing.NewRegion("Panel.SetMaterial").End()
	pd := p.PanelData()
	if !pd.drawing.IsValid() {
		p.ensureBGExists(nil)
	}
	textures := pd.drawing.Material.Textures
	pd.drawing.Material = mat.SelectRoot().CreateInstance(textures)
	p.recreateDrawing()
}

func (p *Panel) SetBackground(tex *rendering.Texture) {
	defer tracing.NewRegion("Panel.SetBackground").End()
	pd := p.PanelData()
	if pd.drawing.IsValid() {
		p.recreateDrawing()
		t := []*rendering.Texture{tex}
		// TODO:  Should this setting of mips be here?
		tex.MipLevels = 1
		p.textureSize = matrix.NewVec2(float32(tex.Width), float32(tex.Height))
		p.shaderData.resetSize2D(p.Base())
		pd.drawing.Material = pd.drawing.Material.SelectRoot().CreateInstance(t)
		if pd.transparentDrawing.Material != nil {
			pd.transparentDrawing.Material = pd.transparentDrawing.Material.SelectRoot().CreateInstance(t)
		}
		host := p.man.Value().Host
		host.Drawings.AddDrawing(pd.drawing)
	} else {
		p.ensureBGExists(tex)
	}
}

func (p *Panel) RemoveBackground() {
	p.recreateDrawing()
}

func (p *Panel) IsScrolling() bool {
	return p.PanelData().flags.isScrolling()
}

func (p *Panel) Freeze() {
	p.PanelData().flags.setFrozen()
}

func (p *Panel) UnFreeze() {
	p.PanelData().flags.resetFrozen()
}

func (p *Panel) IsFrozen() bool {
	return p.PanelData().flags.isFrozen()
}

func (p *Panel) SetScrollDirection(direction PanelScrollDirection) {
	pd := p.PanelData()
	if pd.scrollDirection == direction {
		return
	}
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
	if pd.scrollBarX == nil && (pd.scrollDirection&PanelScrollDirectionHorizontal) != 0 {
		pd.scrollBarX = p.createScrollBar()
		p.AddChild((*UI)(pd.scrollBarX))
	}
	if pd.scrollBarY == nil && (pd.scrollDirection&PanelScrollDirectionVertical) != 0 {
		pd.scrollBarY = p.createScrollBar()
		p.AddChild((*UI)(pd.scrollBarY))
	}
}

func (p *Panel) createScrollBar() *Panel {
	man := p.man.Value()
	scrollBarTex, _ := man.Host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear)
	sb := man.Add().ToPanel()
	sb.Init(scrollBarTex, ElementTypePanel)
	sb.DontFitContent()
	sb.SetColor(matrix.ColorGray())
	sb.layout.SetPositioning(PositioningAbsolute)
	sb.layout.Scale(scrollBarWidth, scrollBarWidth)
	sb.layout.SetZ(10)
	sb.Base().AddEvent(EventTypeEnter, func() {
		sb.EnforceColor(matrix.NewColor(0.575, 0.575, 0.575, 1.0))
	})
	sb.Base().AddEvent(EventTypeExit, func() {
		sb.UnEnforceColor()
	})
	sb.Base().AddEvent(EventTypeDown, func() {
		pd := p.PanelData()
		cp := p.Base().Host().Window.Cursor.Position()
		switch sb {
		case pd.scrollBarX:
			pd.scrollBarStart = sb.layout.offset.X()
			pd.scrollBarDrag.SetX(cp.X())
		case pd.scrollBarY:
			pd.scrollBarStart = sb.layout.offset.Y()
			pd.scrollBarDrag.SetY(cp.Y())
		}
	})
	return sb
}

func (p *Panel) updateScrollBars() {
	pd := p.PanelData()
	if pd.scrollBarX == nil && pd.scrollBarY == nil {
		return
	}
	bars := [...]*Panel{pd.scrollBarX, pd.scrollBarY}
	for i := range bars {
		if bars[i] != nil {
			if p.flags.hovering() {
				bars[i].Base().Show()
			} else {
				bars[i].Base().Hide()
			}
		}
	}
	ps := p.layout.PixelSize()
	panelW, panelH := ps.X(), ps.Y()
	if pd.scrollBarX != nil && pd.scrollBarX.Base().IsActive() {
		y := panelH - scrollBarWidth
		pd.scrollBarX.layout.SetOffsetY(y)
		maxX := pd.maxScroll.X()
		if !matrix.Approx(pd.scrollBarDrag.X(), 0) {
			mx := p.Base().Host().Window.Cursor.Position().X()
			mouseDelta := pd.scrollBarDrag.Y() - mx
			startOffset := pd.scrollBarStart
			newOffset := startOffset + mouseDelta
			maxX := pd.maxScroll.X()
			barW := panelW * (panelW / (panelW + maxX))
			if barW < 1 {
				barW = 1
			}
			newOffset = matrix.Clamp(newOffset, 0, panelW-barW)
			scrollX := -(newOffset / (panelW - barW)) * maxX
			pd.scroll.SetX(matrix.Clamp(scrollX, 0, pd.maxScroll.X()))
		}
		if maxX > 0 {
			barW := panelW * (panelW / (panelW + maxX))
			if barW < 1 {
				barW = 1
			}
			offsetX := (pd.scroll.X() / maxX) * (panelW - barW)
			pd.scrollBarX.layout.Scale(barW, 12)
			pd.scrollBarX.layout.SetOffsetX(offsetX)
		} else {
			pd.scrollBarX.Base().Hide()
		}
	}
	if pd.scrollBarY != nil && pd.scrollBarY.Base().IsActive() {
		x := panelW - scrollBarWidth
		pd.scrollBarY.layout.SetOffsetX(x)
		maxY := pd.maxScroll.Y()
		if !matrix.Approx(pd.scrollBarDrag.Y(), 0) {
			my := p.Base().Host().Window.Cursor.Position().Y()
			mouseDelta := pd.scrollBarDrag.Y() - my
			startOffset := pd.scrollBarStart
			newOffset := startOffset + mouseDelta
			maxY := pd.maxScroll.Y()
			barH := panelH * (panelH / (panelH + maxY))
			if barH < 1 {
				barH = 1
			}
			newOffset = matrix.Clamp(newOffset, 0, panelH-barH)
			scrollY := -(newOffset / (panelH - barH)) * maxY
			pd.scroll.SetY(matrix.Clamp(scrollY, -pd.maxScroll.Y(), 0))
		}
		if maxY > 0 {
			barH := panelH * (panelH / (panelH + maxY))
			if barH < 1 {
				barH = 1
			}
			offsetY := (-pd.scroll.Y() / maxY) * (panelH - barH)
			pd.scrollBarY.layout.Scale(12, barH)
			pd.scrollBarY.layout.SetOffsetY(offsetY)
		} else {
			pd.scrollBarY.Base().Hide()
		}
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
	defer tracing.NewRegion("Panel.SetUseBlending").End()
	p.recreateDrawing()
	pd := p.PanelData()
	host := p.man.Value().Host
	host.Drawings.AddDrawing(pd.drawing)
	if useBlending {
		pd.transparentDrawing = pd.drawing
		m, err := host.MaterialCache().Material(assets.MaterialDefinitionUITransparent)
		if err != nil {
			slog.Error("failed to load the material",
				"material", assets.MaterialDefinitionUITransparent, "error", err)
			return
		}
		pd.transparentDrawing.Material = m.CreateInstance(pd.drawing.Material.Textures)
		host.Drawings.AddDrawing(pd.transparentDrawing)
	}
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
	defer tracing.NewRegion("Panel.setColorInternal").End()
	p.ensureBGExists(nil)
	if p.shaderData.FgColor.Equals(bgColor) {
		return
	}
	host := p.man.Value().Host
	hasBlending := p.shaderData.FgColor.A() < 1.0
	shouldBlend := bgColor.A() < 1.0
	if hasBlending != shouldBlend {
		p.recreateDrawing()
		pd := p.PanelData()
		host.Drawings.AddDrawing(pd.drawing)
		if shouldBlend {
			sd := pd.transparentDrawing.ShaderData
			pd.transparentDrawing = pd.drawing
			pd.transparentDrawing.ShaderData = sd
			m, err := host.MaterialCache().Material(assets.MaterialDefinitionUITransparent)
			if err != nil {
				slog.Error("failed to load the material",
					"material", assets.MaterialDefinitionUITransparent, "error", err)
			} else {
				pd.transparentDrawing.Material = m.CreateInstance(pd.drawing.Material.Textures)
				host.Drawings.AddDrawing(pd.transparentDrawing)
			}
		}
	}
	p.shaderData.FgColor = bgColor
}

func (p *Panel) allowClickThrough() {
	pd := p.PanelData()
	pd.flags.setAllowClickThrough()
	p.events[EventTypeDown].Clear()
	p.events[EventTypeUp].Clear()
}
