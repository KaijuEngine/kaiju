/******************************************************************************/
/* panel.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"log/slog"
	"sort"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

type PanelScrollDirection = int32
type BorderStyle = int32
type ContentFit = int32
type Overflow = int
type panelBits uint8
type LayoutMode = int
type FlexDirection = int
type FlexWrap = int
type FlexJustify = int
type FlexAlignContent = int

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
	LayoutModeFlow = LayoutMode(iota)
	LayoutModeGrid
	LayoutModeFlex
)

const (
	FlexDirectionRow = FlexDirection(iota)
	FlexDirectionRowReverse
	FlexDirectionColumn
	FlexDirectionColumnReverse
)

const (
	FlexWrapNoWrap = FlexWrap(iota)
	FlexWrapWrap
	FlexWrapWrapReverse
)

const (
	FlexJustifyStart = FlexJustify(iota)
	FlexJustifyEnd
	FlexJustifyCenter
	FlexJustifySpaceBetween
	FlexJustifySpaceAround
	FlexJustifySpaceEvenly
)

const (
	FlexAlignContentStart = FlexAlignContent(iota)
	FlexAlignContentEnd
	FlexAlignContentCenter
	FlexAlignContentStretch
	FlexAlignContentSpaceBetween
	FlexAlignContentSpaceAround
	FlexAlignContentSpaceEvenly
)

const (
	panelBitsIsScrolling panelBits = 1 << iota
	panelBitsIsDragging
	panelBitsIsFrozen
	panelBitsAllowDragScroll
	panelBitsAllowClickThrough
	panelBitsWasDirtied
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
	gridColumns               int
	gridGap                   matrix.Vec2
	// Positive values are fixed pixel widths, negative values are fr units.
	gridTemplateColumns []float32
	gridAutoColumns     float32
	gridAutoRows        float32
	requestScrollX      requestScroll
	requestScrollY      requestScroll
	overflow            Overflow
	layoutMode          LayoutMode
	flexDirection       FlexDirection
	flexWrap            FlexWrap
	flexJustify         FlexJustify
	flexAlignItems      FlexAlign
	flexAlignContent    FlexAlignContent
	enforcedColorStack  []matrix.Color
	flags               panelBits
	minSize             matrix.Vec2
	maxSize             matrix.Vec2
	aspectRatio         float32
	usesBorderBox       bool
}

func (b panelBits) isScrolling() bool        { return b&panelBitsIsScrolling != 0 }
func (b panelBits) isDragging() bool         { return b&panelBitsIsDragging != 0 }
func (b panelBits) isFrozen() bool           { return b&panelBitsIsFrozen != 0 }
func (b panelBits) allowDragScroll() bool    { return b&panelBitsAllowDragScroll != 0 }
func (b panelBits) allowClickThrough() bool  { return b&panelBitsAllowClickThrough != 0 }
func (b panelBits) wasDirtied() bool         { return b&panelBitsWasDirtied != 0 }
func (b *panelBits) setIsScrolling()         { *b |= panelBitsIsScrolling }
func (b *panelBits) setDragging()            { *b |= panelBitsIsDragging }
func (b *panelBits) setFrozen()              { *b |= panelBitsIsFrozen }
func (b *panelBits) setAllowDragScroll()     { *b |= panelBitsAllowDragScroll }
func (b *panelBits) setAllowClickThrough()   { *b |= panelBitsAllowClickThrough }
func (b *panelBits) setWasDirtied()          { *b |= panelBitsWasDirtied }
func (b *panelBits) resetIsScrolling()       { *b &= ^panelBitsIsScrolling }
func (b *panelBits) resetDragging()          { *b &= ^panelBitsIsDragging }
func (b *panelBits) resetFrozen()            { *b &= ^panelBitsIsFrozen }
func (b *panelBits) resetAllowDragScroll()   { *b &= ^panelBitsAllowDragScroll }
func (b *panelBits) resetAllowClickThrough() { *b &= ^panelBitsAllowClickThrough }
func (b *panelBits) resetWasDirtied()        { *b &= ^panelBitsWasDirtied }

func (p *panelData) innerPanelData() *panelData { return p }
func (p *panelData) HasMinWidth() bool          { return p.minSize.X() >= 0 }
func (p *panelData) HasMaxWidth() bool          { return p.maxSize.X() >= 0 }
func (p *panelData) HasMinHeight() bool         { return p.minSize.Y() >= 0 }
func (p *panelData) HasMaxHeight() bool         { return p.maxSize.Y() >= 0 }

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
	pd.minSize = matrix.NewVec2(-1, -1)
	pd.maxSize = matrix.NewVec2(-1, -1)
	pd.scrollEvent = 0
	pd.scrollDirection = PanelScrollDirectionNone
	pd.fitContent = ContentFitBoth
	pd.gridColumns = 0
	pd.gridGap = matrix.Vec2Zero()
	pd.gridTemplateColumns = nil
	pd.gridAutoColumns = 0
	pd.gridAutoRows = 0
	pd.layoutMode = LayoutModeFlow
	pd.flexDirection = FlexDirectionRow
	pd.flexWrap = FlexWrapNoWrap
	pd.flexJustify = FlexJustifyStart
	pd.flexAlignItems = FlexAlignStretch
	pd.flexAlignContent = FlexAlignContentStretch
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

func (p *Panel) MaxScroll() matrix.Vec2   { return p.PanelData().maxScroll }
func (p *Panel) ScrollX() float32         { return p.PanelData().scroll.X() }
func (p *Panel) ScrollY() float32         { return -p.PanelData().scroll.Y() }
func (p *Panel) EnableDragScroll()        { p.PanelData().flags.setAllowDragScroll() }
func (p *Panel) DisableDragScroll()       { p.PanelData().flags.resetAllowDragScroll() }
func (p *Panel) GetMinSize() matrix.Vec2  { return p.PanelData().minSize }
func (p *Panel) GetMaxSize() matrix.Vec2  { return p.PanelData().maxSize }
func (p *Panel) SetMinWidth(w float32)    { p.PanelData().minSize.SetX(w) }
func (p *Panel) SetMaxWidth(w float32)    { p.PanelData().maxSize.SetX(w) }
func (p *Panel) SetMinHeight(h float32)   { p.PanelData().minSize.SetY(h) }
func (p *Panel) SetMaxHeight(h float32)   { p.PanelData().maxSize.SetY(h) }
func (p *Panel) GetAspectRatio() float32  { return p.PanelData().aspectRatio }
func (p *Panel) SetAspectRatio(r float32) { p.PanelData().aspectRatio = r }
func (p *Panel) GetUsesBorderBox() bool   { return p.PanelData().usesBorderBox }
func (p *Panel) SetUsesBorderBox(v bool)  { p.PanelData().usesBorderBox = v }

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
	if !p.Base().hasDirty() {
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
	if !p.Base().hasDirty() {
		p.Base().SetDirty(DirtyTypeLayout)
	} else {
		p.Base().SetDirty(DirtyTypeGenerated)
	}
}

func (p *Panel) FitContent() {
	p.PanelData().fitContent = ContentFitBoth
	if !p.Base().hasDirty() {
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
	if pd.flags.wasDirtied() {
		// Update shader visibility based on scissor clipping
		p.updateShaderVisibility()
		pd.flags.resetWasDirtied()
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

type gridLayoutItem struct {
	ui      *UI
	row     int
	col     int
	rowSpan int
	colSpan int
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

// updateShaderVisibility activates or deactivates the panel's shader data based on
// whether the panel's world‑space rectangle intersects the current UI scissor.
// It does not deactivate the entity itself, preserving layout updates.
func (p *Panel) updateShaderVisibility() {
	// Ensure the entity is active; otherwise nothing to render.
	if !p.entity.IsActive() {
		return
	}
	// Retrieve the current scissor rectangle from the root UI.
	scissor := p.Base().selfScissor()
	// Compute the panel's world‑space bounds.
	pos := p.entity.Transform.WorldPosition()
	size := p.layout.PixelSize()
	half := size.Scale(0.5)
	left := pos.X() - half.X()
	right := pos.X() + half.X()
	bottom := pos.Y() - half.Y()
	top := pos.Y() + half.Y()
	// If the panel is completely outside the scissor, deactivate its shader data.
	if right < scissor.X() || left > scissor.Z() || top < scissor.Y() || bottom > scissor.W() {
		p.shaderData.Deactivate()
	} else {
		p.shaderData.Activate()
	}
}

func (p *Panel) boundsChildren(bounds *matrix.Vec2) {
	// defer tracing.NewRegion("Panel.boundsChildren").End()
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
	ps := p.layout.PixelSize()
	maxSize := matrix.Vec2{}
	maxRowsX := matrix.Float(0)
	if p.IsGrid() && pd.gridColumns > 0 {
		maxSize = p.layoutGridChildren(pd, offsetStart, ps)
		maxRowsX = maxSize.X()
	} else if p.IsFlex() {
		maxSize = p.layoutFlexChildren(pd, offsetStart, ps)
		maxRowsX = maxSize.X()
	} else {
		rows := make([]rowBuilder, 0)
		areaWidth := ps.X()
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
		maxRowsX = matrix.Float(0)
		for i := range rows {
			rows[i].setElements(nextPos[matrix.Vx], nextPos[matrix.Vy])
			addY = rows[i].height + rows[i].maxMarginTop + rows[i].maxMarginBottom
			maxRowsX = max(maxRowsX, rows[i].x)
			nextPos[matrix.Vy] += addY
			maxSize[matrix.Vy] += addY
		}
	}
	bounds := matrix.Vec2{maxSize.X(), maxSize.Y()}
	if p.FittingContent() {
		p.boundsChildren(&bounds)
		w := bounds.X() + p.layout.padding.Horizontal() + p.layout.border.Horizontal()
		h := bounds.Y() + p.layout.padding.Bottom() + p.layout.border.Bottom()
		if pd.HasMinWidth() && w < pd.minSize.X() {
			w = pd.minSize.X()
		}
		if pd.HasMaxWidth() && w > pd.maxSize.X() {
			w = pd.maxSize.X()
		}
		if pd.HasMinHeight() && h < pd.minSize.Y() {
			h = pd.minSize.Y()
		}
		if pd.HasMaxHeight() && h > pd.maxSize.Y() {
			h = pd.maxSize.Y()
		}
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
	pd := p.PanelData()
	host := p.man.Value().Host
	if !pd.drawing.IsValid() {
		defer tracing.NewRegion("Panel.ensureBGExists").End()
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
	p.ensureBGExists(textures[0])
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
		host := p.man.Value().Host
		host.Drawings.AddDrawing(pd.drawing)
		if pd.transparentDrawing.Material != nil {
			pd.transparentDrawing.Material = pd.transparentDrawing.Material.SelectRoot().CreateInstance(t)
			host.Drawings.AddDrawing(pd.transparentDrawing)
		}
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

func (p *Panel) SetFlowLayout() {
	pd := p.PanelData()
	if pd.layoutMode == LayoutModeFlow {
		return
	}
	pd.layoutMode = LayoutModeFlow
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) ClearLayoutStyles() {
	pd := p.PanelData()
	pd.layoutMode = LayoutModeFlow
	pd.gridColumns = 0
	pd.gridGap = matrix.Vec2Zero()
	pd.gridTemplateColumns = nil
	pd.gridAutoColumns = 0
	pd.gridAutoRows = 0
	pd.flexDirection = FlexDirectionRow
	pd.flexWrap = FlexWrapNoWrap
	pd.flexJustify = FlexJustifyStart
	pd.flexAlignItems = FlexAlignStretch
	pd.flexAlignContent = FlexAlignContentStretch
	pd.minSize = matrix.NewVec2(-1, -1)
	pd.maxSize = matrix.NewVec2(-1, -1)
	pd.aspectRatio = 0
	pd.usesBorderBox = false
	pd.fitContent = ContentFitBoth
	// TODO:  This is commented out for now to fix a click-glitch bug
	// in the editor content browser.
	// p.layout.ClearStyles()
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) IsGrid() bool {
	return p.PanelData().layoutMode == LayoutModeGrid && p.GridColumns() > 0
}

func (p *Panel) IsFlex() bool { return p.PanelData().layoutMode == LayoutModeFlex }

func (p *Panel) GridColumns() int { return p.PanelData().gridColumns }

func (p *Panel) GridGap() matrix.Vec2 { return p.PanelData().gridGap }

// GridCellWidth returns the computed width of a single grid column (based on current
// panel dimensions, column count, and gap). Used by CSS width % processing so
// children (e.g. div{width:100%}) fit their grid cell instead of full parent.
func (p *Panel) GridCellWidth() float32 {
	pd := p.PanelData()
	if !p.IsGrid() || pd.gridColumns <= 0 {
		return p.layout.PixelSize().X()
	}
	ps := p.layout.PixelSize()
	innerW := ps.X() - p.layout.padding.Horizontal() - p.layout.border.Horizontal()
	gapX := pd.gridGap.X()
	if gapX < 0 {
		gapX = 0
	}
	colW := (innerW - float32(pd.gridColumns-1)*gapX) / float32(pd.gridColumns)
	if len(pd.gridTemplateColumns) == pd.gridColumns {
		widths := p.computeGridColumnWidths(innerW, gapX)
		if len(widths) > 0 {
			sum := float32(0)
			for i := range widths {
				sum += widths[i]
			}
			colW = sum / float32(len(widths))
		}
	}
	if colW < 1 {
		colW = 1
	}
	return colW
}

// SetGrid enables CSS Grid-like layout with the given number of columns.
// Children will be placed in row-major order into the grid cells.
// Column widths are computed by dividing available width by columns (accounting for gaps).
// Use SetGridGap to control spacing between items. This works with the existing
// fit content and scrolling systems.
func (p *Panel) SetGrid(columns int) {
	pd := p.PanelData()
	if columns <= 0 {
		columns = 3 // default for display: grid or auto
	}
	if pd.layoutMode == LayoutModeGrid && pd.gridColumns == columns {
		return
	}
	pd.layoutMode = LayoutModeGrid
	pd.gridColumns = columns
	if len(pd.gridTemplateColumns) != columns {
		pd.gridTemplateColumns = nil
	}
	if pd.gridGap.X() == 0 && pd.gridGap.Y() == 0 {
		pd.gridGap = matrix.NewVec2(8, 8) // sensible default gap like CSS
	}
	p.Base().SetDirty(DirtyTypeLayout)
}

// SetGridTemplateColumns configures explicit grid column widths.
// Positive values are fixed pixels, negative values are fr units.
func (p *Panel) SetGridTemplateColumns(columns []float32) {
	pd := p.PanelData()
	if len(columns) == 0 {
		pd.gridTemplateColumns = nil
		p.Base().SetDirty(DirtyTypeLayout)
		return
	}
	pd.gridTemplateColumns = append(pd.gridTemplateColumns[:0], columns...)
	pd.layoutMode = LayoutModeGrid
	pd.gridColumns = len(columns)
	if pd.gridGap.X() == 0 && pd.gridGap.Y() == 0 {
		pd.gridGap = matrix.NewVec2(8, 8)
	}
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetGridAutoColumns(width float32) {
	if width < 0 {
		width = 0
	}
	pd := p.PanelData()
	if matrix.Approx(pd.gridAutoColumns, width) {
		return
	}
	pd.gridAutoColumns = width
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetGridAutoRows(height float32) {
	if height < 0 {
		height = 0
	}
	pd := p.PanelData()
	if matrix.Approx(pd.gridAutoRows, height) {
		return
	}
	pd.gridAutoRows = height
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetGridGap(x, y float32) {
	pd := p.PanelData()
	if matrix.Approx(pd.gridGap.X(), x) && matrix.Approx(pd.gridGap.Y(), y) {
		return
	}
	pd.gridGap.SetX(x)
	pd.gridGap.SetY(y)
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetFlex() {
	pd := p.PanelData()
	if pd.layoutMode == LayoutModeFlex {
		return
	}
	pd.layoutMode = LayoutModeFlex
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetFlexDirection(direction FlexDirection) {
	pd := p.PanelData()
	if pd.flexDirection == direction {
		return
	}
	pd.flexDirection = direction
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetFlexWrap(wrap FlexWrap) {
	pd := p.PanelData()
	if pd.flexWrap == wrap {
		return
	}
	pd.flexWrap = wrap
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetFlexJustify(justify FlexJustify) {
	pd := p.PanelData()
	if pd.flexJustify == justify {
		return
	}
	pd.flexJustify = justify
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetFlexAlignItems(align FlexAlign) {
	if align == FlexAlignAuto {
		align = FlexAlignStretch
	}
	pd := p.PanelData()
	if pd.flexAlignItems == align {
		return
	}
	pd.flexAlignItems = align
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetFlexAlignContent(align FlexAlignContent) {
	pd := p.PanelData()
	if pd.flexAlignContent == align {
		return
	}
	pd.flexAlignContent = align
	p.Base().SetDirty(DirtyTypeLayout)
}

type flexLayoutItem struct {
	ui        *UI
	order     int
	baseMain  float32
	finalMain float32
	cross     float32
	margin    matrix.Vec4
}

type flexLayoutLine struct {
	items []*flexLayoutItem
	main  float32
	cross float32
}

func flexItemSize(kui *UI) matrix.Vec2 {
	if kui.elmType == ElementTypeLabel {
		size := kui.ToLabel().Measure()
		size[matrix.Vx] += 0.1
		return size
	}
	return kui.Layout().PixelSize()
}

func flexMainSize(size matrix.Vec2, row bool) float32 {
	if row {
		return size.X()
	}
	return size.Y()
}

func flexCrossSize(size matrix.Vec2, row bool) float32 {
	if row {
		return size.Y()
	}
	return size.X()
}

func flexMainMargin(margin matrix.Vec4, row bool) float32 {
	if row {
		return margin.Horizontal()
	}
	return margin.Vertical()
}

func flexCrossMargin(margin matrix.Vec4, row bool) float32 {
	if row {
		return margin.Vertical()
	}
	return margin.Horizontal()
}

func flexSetMainSize(kui *UI, row bool, size float32) {
	if size < 1 {
		size = 1
	}
	if row {
		kui.Layout().ScaleWidth(size)
	} else {
		kui.Layout().ScaleHeight(size)
	}
}

func flexSetCrossSize(kui *UI, row bool, size float32) {
	if size < 1 {
		size = 1
	}
	if row {
		kui.Layout().ScaleHeight(size)
	} else {
		kui.Layout().ScaleWidth(size)
	}
}

func clampFlexItemSize(kui *UI, row bool, size float32) float32 {
	if kui.IsType(ElementTypeLabel) {
		return size
	}
	p := kui.ToPanel()
	pd := p.PanelData()
	if row {
		if pd.HasMinWidth() && size < pd.minSize.X() {
			size = pd.minSize.X()
		}
		if pd.HasMaxWidth() && size > pd.maxSize.X() {
			size = pd.maxSize.X()
		}
	} else {
		if pd.HasMinHeight() && size < pd.minSize.Y() {
			size = pd.minSize.Y()
		}
		if pd.HasMaxHeight() && size > pd.maxSize.Y() {
			size = pd.maxSize.Y()
		}
	}
	return size
}

func (p *Panel) collectFlexItems(pd *panelData, row bool, containerMain float32) []*flexLayoutItem {
	items := make([]*flexLayoutItem, 0, len(p.entity.Children))
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
		case PositioningAbsolute, PositioningFixed, PositioningSticky:
			continue
		}
		size := flexItemSize(kui)
		baseMain := flexMainSize(size, row)
		if !kLayout.FlexBasisAuto() {
			baseMain = kLayout.FlexBasis()
			if kLayout.FlexBasisPercent() {
				baseMain *= containerMain
			}
		}
		baseMain = clampFlexItemSize(kui, row, baseMain)
		items = append(items, &flexLayoutItem{
			ui:        kui,
			order:     kLayout.FlexOrder(),
			baseMain:  baseMain,
			finalMain: baseMain,
			cross:     flexCrossSize(size, row),
			margin:    kLayout.Margin(),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].order < items[j].order
	})
	if pd.flexDirection == FlexDirectionRowReverse || pd.flexDirection == FlexDirectionColumnReverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}
	return items
}

func appendFlexLine(lines *[]flexLayoutLine, line flexLayoutLine) {
	if len(line.items) > 0 {
		*lines = append(*lines, line)
	}
}

func buildFlexLines(items []*flexLayoutItem, row, wrap bool, containerMain, gapMain float32) []flexLayoutLine {
	lines := make([]flexLayoutLine, 0, 1)
	line := flexLayoutLine{}
	for i := range items {
		item := items[i]
		outerMain := item.baseMain + flexMainMargin(item.margin, row)
		if !wrap {
			outerMain = item.baseMain
		}
		if len(line.items) > 0 {
			outerMain += gapMain
		}
		if wrap && len(line.items) > 0 && line.main+outerMain > containerMain {
			appendFlexLine(&lines, line)
			line = flexLayoutLine{}
			outerMain = item.baseMain + flexMainMargin(item.margin, row)
		}
		line.items = append(line.items, item)
		line.main += outerMain
		line.cross = matrix.Max(line.cross, item.cross)
	}
	appendFlexLine(&lines, line)
	return lines
}

func distributeFlexLine(line *flexLayoutLine, row bool, containerMain, gapMain float32) {
	totalOuter := float32(0)
	totalGrow := float32(0)
	totalShrink := float32(0)
	for i := range line.items {
		item := line.items[i]
		totalOuter += item.baseMain + flexMainMargin(item.margin, row)
		totalGrow += item.ui.Layout().FlexGrow()
		totalShrink += item.ui.Layout().FlexShrink() * item.baseMain
	}
	if len(line.items) > 1 {
		totalOuter += float32(len(line.items)-1) * gapMain
	}
	free := containerMain - totalOuter
	for i := range line.items {
		item := line.items[i]
		item.finalMain = item.baseMain
		if free > 0 && totalGrow > 0 {
			item.finalMain += free * (item.ui.Layout().FlexGrow() / totalGrow)
		} else if free < 0 && totalShrink > 0 {
			item.finalMain += free * ((item.ui.Layout().FlexShrink() * item.baseMain) / totalShrink)
		}
		item.finalMain = clampFlexItemSize(item.ui, row, item.finalMain)
		flexSetMainSize(item.ui, row, item.finalMain)
	}
	line.main = 0
	line.cross = 0
	for i := range line.items {
		item := line.items[i]
		size := flexItemSize(item.ui)
		item.finalMain = flexMainSize(size, row)
		item.cross = flexCrossSize(size, row)
		line.main += item.finalMain + flexMainMargin(item.margin, row)
		line.cross = matrix.Max(line.cross, item.cross+flexCrossMargin(item.margin, row))
	}
	if len(line.items) > 1 {
		line.main += float32(len(line.items)-1) * gapMain
	}
}

func flexDistributedStart(free float32, count int, justify FlexJustify) (float32, float32) {
	if free < 0 {
		free = 0
	}
	switch justify {
	case FlexJustifyEnd:
		return free, 0
	case FlexJustifyCenter:
		return free * 0.5, 0
	case FlexJustifySpaceBetween:
		if count > 1 {
			return 0, free / float32(count-1)
		}
	case FlexJustifySpaceAround:
		if count > 0 {
			gap := free / float32(count)
			return gap * 0.5, gap
		}
	case FlexJustifySpaceEvenly:
		if count > 0 {
			gap := free / float32(count+1)
			return gap, gap
		}
	}
	return 0, 0
}

func flexAlignContentStart(free float32, count int, align FlexAlignContent) (float32, float32) {
	if free < 0 {
		free = 0
	}
	switch align {
	case FlexAlignContentEnd:
		return free, 0
	case FlexAlignContentCenter:
		return free * 0.5, 0
	case FlexAlignContentSpaceBetween:
		if count > 1 {
			return 0, free / float32(count-1)
		}
	case FlexAlignContentSpaceAround:
		if count > 0 {
			gap := free / float32(count)
			return gap * 0.5, gap
		}
	case FlexAlignContentSpaceEvenly:
		if count > 0 {
			gap := free / float32(count+1)
			return gap, gap
		}
	}
	return 0, 0
}

func flexItemCrossOffset(lineCross, itemCross float32, margin matrix.Vec4, row bool, align FlexAlign) float32 {
	outerCross := itemCross + flexCrossMargin(margin, row)
	switch align {
	case FlexAlignEnd:
		if row {
			return lineCross - outerCross + margin.Top()
		}
		return lineCross - outerCross + margin.Left()
	case FlexAlignCenter:
		if row {
			return (lineCross-outerCross)*0.5 + margin.Top()
		}
		return (lineCross-outerCross)*0.5 + margin.Left()
	default:
		if row {
			return margin.Top()
		}
		return margin.Left()
	}
}

func (p *Panel) layoutFlexChildren(pd *panelData, offsetStart matrix.Vec2, ps matrix.Vec2) matrix.Vec2 {
	defer tracing.NewRegion("Panel.layoutFlexChildren").End()
	row := pd.flexDirection == FlexDirectionRow || pd.flexDirection == FlexDirectionRowReverse
	innerLeft := p.layout.padding.Left() + p.layout.border.Left()
	innerTop := p.layout.padding.Top() + p.layout.border.Top()
	startX := offsetStart.X() + innerLeft
	startY := offsetStart.Y() + innerTop
	innerWidth := ps.X() - p.layout.padding.Horizontal() - p.layout.border.Horizontal()
	innerHeight := ps.Y() - p.layout.padding.Vertical() - p.layout.border.Vertical()
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}
	containerMain := innerWidth
	containerCross := innerHeight
	gapMain := pd.gridGap.X()
	gapCross := pd.gridGap.Y()
	if !row {
		containerMain = innerHeight
		containerCross = innerWidth
		gapMain = pd.gridGap.Y()
		gapCross = pd.gridGap.X()
	}
	if gapMain < 0 {
		gapMain = 0
	}
	if gapCross < 0 {
		gapCross = 0
	}
	items := p.collectFlexItems(pd, row, containerMain)
	if len(items) == 0 {
		return matrix.NewVec2(innerLeft+p.layout.padding.Right()+p.layout.border.Right(),
			innerTop+p.layout.padding.Bottom()+p.layout.border.Bottom())
	}
	wrap := pd.flexWrap != FlexWrapNoWrap
	lines := buildFlexLines(items, row, wrap, containerMain, gapMain)
	for i := range lines {
		distributeFlexLine(&lines[i], row, containerMain, gapMain)
	}
	if pd.flexWrap == FlexWrapWrapReverse {
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
	}
	totalCross := float32(0)
	for i := range lines {
		totalCross += lines[i].cross
	}
	if len(lines) > 1 {
		totalCross += float32(len(lines)-1) * gapCross
	}
	fittingCrossContent := (row && p.FittingContentHeight()) || (!row && p.FittingContentWidth())
	if fittingCrossContent {
		containerCross = totalCross
	}
	extraCross := containerCross - totalCross
	if extraCross < 0 {
		extraCross = 0
	}
	if pd.flexAlignContent == FlexAlignContentStretch && len(lines) > 0 {
		add := extraCross / float32(len(lines))
		for i := range lines {
			lines[i].cross += add
		}
		extraCross = 0
	}
	lineStart, lineExtraGap := flexAlignContentStart(extraCross, len(lines), pd.flexAlignContent)
	maxMainUsed := float32(0)
	maxCrossUsed := float32(0)
	crossPos := lineStart
	for lineIdx := range lines {
		line := &lines[lineIdx]
		mainFree := containerMain - line.main
		mainStart, extraMainGap := flexDistributedStart(mainFree, len(line.items), pd.flexJustify)
		mainPos := mainStart
		for itemIdx := range line.items {
			item := line.items[itemIdx]
			align := item.ui.Layout().AlignSelf()
			if align == FlexAlignAuto {
				align = pd.flexAlignItems
			}
			itemCross := item.cross
			if align == FlexAlignStretch {
				itemCross = line.cross - flexCrossMargin(item.margin, row)
				flexSetCrossSize(item.ui, row, itemCross)
				item.cross = flexCrossSize(flexItemSize(item.ui), row)
			}
			crossOffset := flexItemCrossOffset(line.cross, item.cross, item.margin, row, align)
			if row {
				x := startX + mainPos + item.margin.Left()
				y := startY + crossPos + crossOffset
				item.ui.Layout().SetRowLayoutOffset(matrix.NewVec2(x, y))
				mainPos += item.finalMain + item.margin.Horizontal() + gapMain + extraMainGap
			} else {
				x := startX + crossPos + crossOffset
				y := startY + mainPos + item.margin.Top()
				item.ui.Layout().SetRowLayoutOffset(matrix.NewVec2(x, y))
				mainPos += item.finalMain + item.margin.Vertical() + gapMain + extraMainGap
			}
		}
		maxMainUsed = matrix.Max(maxMainUsed, line.main)
		maxCrossUsed = matrix.Max(maxCrossUsed, crossPos+line.cross)
		crossPos += line.cross + gapCross + lineExtraGap
	}
	if row {
		return matrix.NewVec2(maxMainUsed+innerLeft+p.layout.padding.Right()+p.layout.border.Right(),
			maxCrossUsed+innerTop+p.layout.padding.Bottom()+p.layout.border.Bottom())
	}
	return matrix.NewVec2(maxCrossUsed+innerLeft+p.layout.padding.Right()+p.layout.border.Right(),
		maxMainUsed+innerTop+p.layout.padding.Bottom()+p.layout.border.Bottom())
}

func (p *Panel) computeGridColumnWidths(innerWidth, gapX float32, columns ...int) []float32 {
	pd := p.PanelData()
	cols := pd.gridColumns
	if len(columns) > 0 && columns[0] > cols {
		cols = columns[0]
	}
	if cols <= 0 {
		return []float32{}
	}
	out := make([]float32, cols)
	explicitCols := pd.gridColumns
	if explicitCols <= 0 {
		explicitCols = cols
	}
	if len(pd.gridTemplateColumns) != explicitCols {
		colW := (innerWidth - float32(explicitCols-1)*gapX) / float32(explicitCols)
		if colW < 1 {
			colW = 1
		}
		for i := 0; i < cols; i++ {
			if i < explicitCols || pd.gridAutoColumns <= 0 {
				out[i] = colW
			} else {
				out[i] = pd.gridAutoColumns
			}
		}
		return out
	}
	totalFixed := float32(0)
	totalFr := float32(0)
	for i := 0; i < explicitCols; i++ {
		v := pd.gridTemplateColumns[i]
		if v >= 0 {
			totalFixed += v
		} else {
			totalFr += -v
		}
	}
	remaining := innerWidth - totalFixed - float32(cols-1)*gapX
	if remaining < 0 {
		remaining = 0
	}
	for i := 0; i < cols; i++ {
		if i >= explicitCols {
			if pd.gridAutoColumns > 0 {
				out[i] = pd.gridAutoColumns
			} else {
				out[i] = innerWidth / float32(explicitCols)
			}
			if out[i] < 1 {
				out[i] = 1
			}
			continue
		}
		v := pd.gridTemplateColumns[i]
		if v >= 0 {
			out[i] = v
		} else if totalFr > 0 {
			out[i] = remaining * ((-v) / totalFr)
		}
		if out[i] < 1 {
			out[i] = 1
		}
	}
	return out
}

func gridTrackSpan(start, end int) int {
	if start > 0 && end > start {
		return end - start
	}
	return 1
}

func isGridAreaFree(occupied map[int]map[int]bool, columns, row, col, rowSpan, colSpan int) bool {
	if col < 0 || col+colSpan > columns {
		return false
	}
	for y := 0; y < rowSpan; y++ {
		for x := 0; x < colSpan; x++ {
			if occupied[row+y][col+x] {
				return false
			}
		}
	}
	return true
}

func occupyGridArea(occupied map[int]map[int]bool, row, col, rowSpan, colSpan int) {
	for y := 0; y < rowSpan; y++ {
		r := row + y
		if _, ok := occupied[r]; !ok {
			occupied[r] = map[int]bool{}
		}
		for x := 0; x < colSpan; x++ {
			occupied[r][col+x] = true
		}
	}
}

func nextGridCell(occupied map[int]map[int]bool, columns, row, col, rowSpan, colSpan int) (int, int) {
	for {
		if col >= columns || col+colSpan > columns {
			row++
			col = 0
		}
		if _, ok := occupied[row]; !ok {
			occupied[row] = map[int]bool{}
		}
		if isGridAreaFree(occupied, columns, row, col, rowSpan, colSpan) {
			return row, col
		}
		col++
	}
}

func nextGridCellInColumn(occupied map[int]map[int]bool, columns, row, col, rowSpan, colSpan int) (int, int) {
	if col+colSpan > columns {
		return nextGridCell(occupied, columns, row, 0, rowSpan, colSpan)
	}
	for {
		if _, ok := occupied[row]; !ok {
			occupied[row] = map[int]bool{}
		}
		if isGridAreaFree(occupied, columns, row, col, rowSpan, colSpan) {
			return row, col
		}
		row++
	}
}

func firstGridCellInRow(occupied map[int]map[int]bool, columns, row, rowSpan, colSpan int) (int, int) {
	for col := 0; col < columns; col++ {
		if isGridAreaFree(occupied, columns, row, col, rowSpan, colSpan) {
			return row, col
		}
	}
	return nextGridCell(occupied, columns, row+1, 0, rowSpan, colSpan)
}

func requestedGridCell(occupied map[int]map[int]bool, columns, cursorRow, cursorCol, rowStart, colStart, rowSpan, colSpan int) (int, int) {
	if rowStart > 0 && colStart > 0 {
		return nextGridCellInColumn(occupied, columns, rowStart-1, colStart-1, rowSpan, colSpan)
	}
	if rowStart > 0 {
		return firstGridCellInRow(occupied, columns, rowStart-1, rowSpan, colSpan)
	}
	if colStart > 0 {
		return nextGridCellInColumn(occupied, columns, cursorRow, colStart-1, rowSpan, colSpan)
	}
	return nextGridCell(occupied, columns, cursorRow, cursorCol, rowSpan, colSpan)
}

func (p *Panel) layoutGridChildren(pd *panelData, offsetStart matrix.Vec2, ps matrix.Vec2) matrix.Vec2 {
	defer tracing.NewRegion("Panel.layoutGridChildren").End()
	innerLeft := p.layout.padding.Left() + p.layout.border.Left()
	innerTop := p.layout.padding.Top() + p.layout.border.Top()
	startX := offsetStart.X() + innerLeft
	startY := offsetStart.Y() + innerTop
	innerWidth := ps.X() - p.layout.padding.Horizontal() - p.layout.border.Horizontal()
	if innerWidth < 1 {
		innerWidth = 100
	}
	gapX := pd.gridGap.X()
	if gapX < 0 {
		gapX = 0
	}
	gapY := pd.gridGap.Y()
	if gapY < 0 {
		gapY = 0
	}
	effectiveColumns := pd.gridColumns
	for _, kid := range p.entity.Children {
		if !kid.IsActive() || kid.IsDestroyed() {
			continue
		}
		kui := FirstOnEntity(kid)
		if kui == nil {
			continue
		}
		kLayout := kui.Layout()
		switch kLayout.Positioning() {
		case PositioningAbsolute, PositioningFixed, PositioningSticky:
			continue
		}
		if start := kLayout.GridColumnStart(); start > 0 {
			span := gridTrackSpan(start, kLayout.GridColumnEnd())
			effectiveColumns = max(effectiveColumns, start+span-1)
		}
	}
	colWidths := p.computeGridColumnWidths(innerWidth, gapX, effectiveColumns)
	items := make([]gridLayoutItem, 0, len(p.entity.Children))
	occupied := map[int]map[int]bool{}
	cursorRow := 0
	cursorCol := 0
	rowCount := 0
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
		case PositioningAbsolute, PositioningFixed, PositioningSticky:
			continue
		}
		rowSpan := gridTrackSpan(kLayout.GridRowStart(), kLayout.GridRowEnd())
		colSpan := gridTrackSpan(kLayout.GridColumnStart(), kLayout.GridColumnEnd())
		if colSpan > effectiveColumns {
			colSpan = effectiveColumns
		}
		row, col := requestedGridCell(occupied, effectiveColumns, cursorRow, cursorCol,
			kLayout.GridRowStart(), kLayout.GridColumnStart(), rowSpan, colSpan)
		if kLayout.GridRowStart() == 0 && kLayout.GridColumnStart() == 0 {
			cursorRow, cursorCol = row, col+colSpan
		}
		occupyGridArea(occupied, row, col, rowSpan, colSpan)
		rowCount = max(rowCount, row+rowSpan)
		items = append(items, gridLayoutItem{
			ui:      kui,
			row:     row,
			col:     col,
			rowSpan: rowSpan,
			colSpan: colSpan,
		})
	}
	if rowCount == 0 {
		return matrix.Vec2{innerWidth, innerTop + p.layout.padding.Bottom() + p.layout.border.Bottom()}
	}
	rowHeights := make([]float32, rowCount)
	if pd.gridAutoRows > 0 {
		for i := range rowHeights {
			rowHeights[i] = pd.gridAutoRows
		}
	}
	for i := range items {
		kLayout := items[i].ui.Layout()
		kSize := kLayout.PixelSize()
		margin := kLayout.Margin()
		rowHeights[items[i].row] = matrix.Max(rowHeights[items[i].row], kSize.Y()+margin.Vertical())
	}
	rowOffsets := make([]float32, rowCount)
	y := startY
	for i := 0; i < rowCount; i++ {
		rowOffsets[i] = y
		y += rowHeights[i] + gapY
	}
	contentSize := matrix.Vec2{innerWidth, innerTop}
	for i := range items {
		kui := items[i].ui
		kLayout := kui.Layout()
		kSize := kLayout.PixelSize()
		margin := kLayout.Margin()
		cellX := startX
		for col := 0; col < items[i].col && col < len(colWidths); col++ {
			cellX += colWidths[col] + gapX
		}
		x := cellX + margin.X() // left aligned like CSS start
		itemY := rowOffsets[items[i].row] + margin.Y()
		kLayout.SetRowLayoutOffset(matrix.NewVec2(x, itemY))
		right := (x - startX) + kSize.X() + margin.Z()
		contentSize.SetX(matrix.Max(contentSize.X(), right))
		bottom := itemY - offsetStart.Y() + kSize.Y() + margin.W()
		contentSize.SetY(matrix.Max(contentSize.Y(), bottom))
	}
	contentSize.SetY(contentSize.Y() + p.layout.padding.Bottom() + p.layout.border.Bottom())
	return contentSize
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
	if !p.flags.hovering() {
		if pd.scrollBarX != nil {
			pd.scrollBarX.Base().Hide()
		}
		if pd.scrollBarY != nil {
			pd.scrollBarY.Base().Hide()
		}
		return
	}
	ps := p.layout.PixelSize()
	panelW, panelH := ps.X(), ps.Y()
	if pd.scrollBarX != nil && p.flags.hovering() {
		y := panelH - scrollBarWidth
		pd.scrollBarX.layout.SetOffsetY(y)
		maxX := pd.maxScroll.X()
		if !matrix.Approx(pd.scrollBarDrag.X(), 0) {
			mx := p.Base().Host().Window.Cursor.Position().X()
			mouseDelta := pd.scrollBarDrag.X() - mx
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
			pd.scrollBarX.Base().Show()
		} else {
			pd.scrollBarX.Base().Hide()
		}
	}
	if pd.scrollBarY != nil && p.flags.hovering() {
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
			pd.scrollBarY.Base().Show()
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

func (p *Panel) OutlineOutset() float32 {
	return p.shaderData.OutlineSize.X() + p.shaderData.OutlineSize.Y()
}

func (p *Panel) SetBorderRadius(topLeft, topRight, bottomRight, bottomLeft float32) {
	p.shaderData.BorderRadius = matrix.Vec4{
		bottomLeft, bottomRight, topRight, topLeft}
}

func (p *Panel) SetBorderRadiusTopLeft(r float32)     { p.shaderData.BorderRadius.SetW(r) }
func (p *Panel) SetBorderRadiusTopRight(r float32)    { p.shaderData.BorderRadius.SetZ(r) }
func (p *Panel) SetBorderRadiusBottomRight(r float32) { p.shaderData.BorderRadius.SetY(r) }
func (p *Panel) SetBorderRadiusBottomLeft(r float32)  { p.shaderData.BorderRadius.SetX(r) }

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

func (p *Panel) OutlineColor() matrix.Color { return p.shaderData.OutlineColor }

func (p *Panel) OutlineWidth() float32 { return p.shaderData.OutlineSize.X() }

func (p *Panel) OutlineOffset() float32 { return p.shaderData.OutlineSize.Y() }

func (p *Panel) SetOutline(width, offset float32, color matrix.Color) {
	if width < 0 {
		width = 0
	}
	if offset < 0 {
		offset = 0
	}
	if width > 0 && color.A() > 0 {
		p.ensureBGExists(nil)
	}
	p.shaderData.OutlineSize = matrix.NewVec2(width, offset)
	p.shaderData.OutlineColor = color
	p.Base().SetDirty(DirtyTypeLayout)
}

func (p *Panel) SetUseBlending(useBlending bool) {
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

func (p *Panel) AllowClickThrough() {
	pd := p.PanelData()
	pd.flags.setAllowClickThrough()
	p.events[EventTypeDown].Clear()
	p.events[EventTypeUp].Clear()
	p.events[EventTypeClick].Clear()
	p.events[EventTypeRightClick].Clear()
	p.events[EventTypeRightDown].Clear()
	p.events[EventTypeRightUp].Clear()
}
