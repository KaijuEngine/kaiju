package ui

import (
	"kaiju/assets"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
	"math"
)

type PanelScrollDirection = int32

const (
	PanelScrollDirectionNone       = 0x00
	PanelScrollDirectionVertical   = 0x01
	PanelScrollDirectionHorizontal = 0x02
	PanelScrollDirectionBoth       = 0x03
)

type BorderStyle = int32

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

type childScrollEvent struct {
	down   engine.EventId
	scroll engine.EventId
}

type Panel struct {
	uiBase
	scroll, offset, maxScroll     matrix.Vec2
	scrollSpeed                   float32
	scrollDirection               PanelScrollDirection
	scrollEvent                   engine.EventId
	childScrollEvents             map[UI]childScrollEvent
	color                         matrix.Color
	drawing                       rendering.Drawing
	localData                     any
	innerUpdate                   func(deltaTime float64)
	isScrolling, dragging, frozen bool
	isButton                      bool
	fitContent                    bool
}

func NewPanel(host *engine.Host, texture *rendering.Texture, anchor Anchor) *Panel {
	panel := &Panel{
		scrollEvent:       -1,
		scrollSpeed:       30.0,
		childScrollEvents: make(map[UI]childScrollEvent),
		scrollDirection:   PanelScrollDirectionVertical,
		color:             matrix.Color{1.0, 1.0, 1.0, 1.0},
		fitContent:        false,
	}
	panel.init(host, texture.Size(), anchor, panel)
	panel.updateId = panel.host.Updater.AddUpdate(panel.update)
	panel.entity.Transform.SetScale(matrix.Vec3{1.0, 1.0, 1.0})
	panel.Clean()
	panel.scrollEvent = panel.AddEvent(EventTypeScroll, panel.onScroll)
	panel.ensureBGExists(texture)
	panel.AddEvent(EventTypeRebuild, panel.onRebuild)
	return panel
}

func (panel *Panel) DontFitContent() {
	panel.fitContent = false
}

func (panel *Panel) FittingContent() bool {
	return panel.fitContent
}

func (panel *Panel) FitContent() {
	panel.fitContent = true
	if panel.dirtyType == DirtyTypeNone {
		panel.SetDirty(DirtyTypeLayout)
	} else {
		panel.SetDirty(DirtyTypeReGenerated)
	}
}

func (panel *Panel) onScroll() {
	mouse := &panel.host.Window.Mouse
	delta := mouse.Scroll()
	if !mouse.Scrolled() {
		pos := panel.host.Window.Cursor.ScreenPosition()
		delta = pos.Subtract(panel.downPos)
		delta[matrix.Vx] *= -1.0
	} else {
		panel.offset = panel.scroll
		delta.ScaleAssign(-1.0 * panel.scrollSpeed)
	}
	if (panel.scrollDirection & PanelScrollDirectionHorizontal) != 0 {
		x := matrix.Clamp(delta.X()+panel.offset.X(), 0.0, panel.maxScroll.X())
		panel.scroll.SetX(x)
	}
	if (panel.scrollDirection & PanelScrollDirectionVertical) != 0 {
		y := matrix.Clamp(delta.Y()+panel.offset.Y(), 0.0, panel.maxScroll.Y())
		panel.scroll.SetY(y)
	}
	panel.SetDirty(DirtyTypeLayout)
	panel.isScrolling = true
}

func panelOnDown(ui UI) {
	var target UI = ui
	ok := false
	for !ok {
		target = FirstOnEntity(target.Entity().Parent)
		_, ok = target.(*Panel)
	}
	panel := target.(*Panel)
	panel.offset = panel.scroll
	panel.dragging = true
}

func (panel *Panel) shouldAddScrollEvents() bool {
	switch panel.scrollDirection {
	case PanelScrollDirectionNone:
		return false
	case PanelScrollDirectionHorizontal:
		return panel.maxScroll.X() > math.SmallestNonzeroFloat32
	case PanelScrollDirectionVertical:
		return panel.maxScroll.Y() > math.SmallestNonzeroFloat32
	case PanelScrollDirectionBoth:
		return panel.maxScroll.X() > math.SmallestNonzeroFloat32 || panel.maxScroll.Y() > math.SmallestNonzeroFloat32
	default:
		panic("Invalid scroll direction")
	}
}

func (panel *Panel) disableScrollEvents() {
	panel.RemoveEvent(EventTypeScroll, panel.scrollEvent)
	for i := 0; i < len(panel.entity.Children); i++ {
		c := FirstOnEntity(panel.entity.Children[i])
		// TODO:  Nested scroll panels drag...
		if _, isPanel := c.(*Panel); !isPanel {
			if cse, ok := panel.childScrollEvents[c]; ok {
				c.RemoveEvent(EventTypeDown, cse.down)
				c.RemoveEvent(EventTypeScroll, cse.scroll)
			}
		}
	}
}

func (panel *Panel) tryEnableScrollEvents() {
	if panel.shouldAddScrollEvents() {
		panel.scrollEvent = panel.AddEvent(EventTypeScroll, panel.onScroll)
		for i := 0; i < len(panel.entity.Children); i++ {
			c := FirstOnEntity(panel.entity.Children[i])
			// TODO:  Nested scroll panels drag...
			if _, isPanel := c.(*Panel); !isPanel {
				cse := childScrollEvent{}
				cse.down = c.AddEvent(EventTypeDown, func() { panelOnDown(c) })
				cse.scroll = c.AddEvent(EventTypeScroll, func() { panelOnDown(c) })
				panel.childScrollEvents[c] = cse
			}
		}
	}
}

func (panel *Panel) update(deltaTime float64) {
	panel.uiBase.Update(deltaTime)
	if !panel.frozen {
		if panel.isDown && panel.dragging {
			panel.onScroll()
		} else if panel.dragging {
			panel.dragging = false
		} else {
			panel.isScrolling = false
		}
	}
	if panel.innerUpdate != nil {
		panel.innerUpdate(deltaTime)
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
	eSize := e.Layout().pixelSize
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
		x, y := offsetX, offsetY
		switch e.Layout().positioning {
		case PositioningAbsolute:
			fallthrough
		case PositioningRelative:
			if e.Layout().Anchor().IsLeft() {
				x += e.Layout().InnerOffset().Left()
			} else {
				x += e.Layout().InnerOffset().Right()
			}
			if e.Layout().Anchor().IsTop() {
				y += e.Layout().InnerOffset().Top()
			} else {
				y += e.Layout().InnerOffset().Bottom()
			}
		}
		x += e.Layout().margin.X()
		y += rb.maxMarginTop
		e.Layout().SetOffset(x, -y)
		offsetX += e.Layout().pixelSize.Width() + e.Layout().margin.X() + e.Layout().margin.Z()
	}
}

func (panel *Panel) onRebuild() {
	panel.disableScrollEvents()
	if len(panel.entity.Children) == 0 {
		return
	}
	offsetStart := matrix.Vec2{-panel.scroll.X(), panel.scroll.Y()}
	rows := make([]rowBuilder, 0)
	areaWidth := panel.Layout().mySize.X() - panel.Layout().padding.X() - panel.Layout().padding.Z()
	for _, kid := range panel.entity.Children {
		if !kid.IsActive() || kid.IsDestroyed() {
			continue
		}
		target := FirstOnEntity(kid)
		if target == nil {
			panic("No UI component on entity")
		}
		kui := target
		panel.adjustKidsOnRebuild(kui)
		switch kui.Layout().Positioning() {
		case PositioningAbsolute:
			if kui.Layout().Anchor().IsTop() {
				kui.Layout().SetOffset(kui.Layout().left+kui.Layout().InnerOffset().Left(),
					kui.Layout().top+kui.Layout().InnerOffset().Top())
			} else if kui.Layout().Anchor().IsBottom() {
				kui.Layout().SetOffset(kui.Layout().left+kui.Layout().InnerOffset().Left(),
					kui.Layout().bottom-kui.Layout().InnerOffset().Bottom())
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
	xyOffset := matrix.Vec2{panel.Layout().padding.X(), panel.Layout().padding.Y()}
	nextPos := offsetStart.Add(xyOffset)
	for _, row := range rows {
		row.setElements(panel.Layout().padding.X(), nextPos[matrix.Vy])
		nextPos[matrix.Vy] += row.Height()
	}
	nextPos[matrix.Vy] += panel.Layout().padding.W()
	if panel.fitContent {
		ph := panel.Layout().pixelSize.Height()
		if !matrix.Approx(ph, nextPos.Y()) {
			w := panel.Layout().pixelSize.Width() - panel.Layout().padding.Left() - panel.Layout().padding.Right()
			h := nextPos.Y() - panel.Layout().padding.Top() - panel.Layout().padding.Bottom()
			panel.Layout().Scale(w, h)
			panel.SetDirty(DirtyTypeReGenerated)
			if pui := FirstOnEntity(panel.entity.Parent); pui != nil {
				if p, ok := pui.(*Panel); ok {
					p.FitContent()
				}
			}
		}
	}
	if panel.dirtyType != DirtyTypeReGenerated {
		length := nextPos.Subtract(offsetStart)
		last := panel.maxScroll
		ws := panel.entity.Transform.WorldScale()
		panel.maxScroll = matrix.Vec2{
			matrix.Max(0.0, length.X()-ws.X()),
			matrix.Max(0.0, length.Y()-ws.Y())}
		if !matrix.Vec2Approx(last, panel.maxScroll) {
			panel.SetScrollX(panel.scroll.X())
			panel.SetScrollY(panel.scroll.Y())
		}
		panel.tryEnableScrollEvents()
	}
}

func (panel *Panel) adjustKidsOnRebuild(target UI) {
	// TODO:  Only do this if the panel's values have changed
	pos := panel.entity.Transform.WorldPosition()
	size := panel.entity.Transform.WorldScale()
	bounds := matrix.Vec4{
		pos.X() - size.X()*0.5,
		pos.Y() - size.Y()*0.5,
		pos.X() + size.X()*0.5,
		pos.Y() + size.Y()*0.5,
	}
	if panel.entity.Parent != nil {
		pUI := FirstOnEntity(panel.entity.Parent)
		if pUI != nil && pUI.selfScissor().Z() < math.MaxFloat32 {
			parentScissor := pUI.selfScissor()
			bounds.SetLeft(matrix.Max(parentScissor.X(), bounds.X()))
			bounds.SetTop(matrix.Max(parentScissor.Y(), bounds.Y()))
			bounds.SetRight(matrix.Min(parentScissor.Z(), bounds.Z()))
			bounds.SetBottom(matrix.Min(parentScissor.W(), bounds.W()))
		}
	}
	panel.setScissor(bounds)
	//panel.ui.isDirty = DirtyTypeNone;
}

func (panel *Panel) AddChild(target UI) {
	target.Entity().SetParent(panel.entity)
	panel.Layout().update()
	if panel.group != nil {
		target.SetGroup(panel.group)
	}
	panel.SetDirty(DirtyTypeGenerated)
	panel.tryEnableScrollEvents()
}

func (panel *Panel) InsertChild(target UI, idx int) {
	panel.AddChild(target)
	kidLen := len(panel.entity.Children)
	idx = max(idx, 0)
	for i := idx; i < kidLen-1; i++ {
		panel.entity.Children[i], panel.entity.Children[kidLen-1] = panel.entity.Children[kidLen-1], panel.entity.Children[i]
	}
}

func (panel *Panel) RemoveChild(target UI) {
	target.Entity().SetParent(nil)
	target.setScissor(matrix.Vec4{-math.MaxFloat32, -math.MaxFloat32, math.MaxFloat32, math.MaxFloat32})
	target.Layout().update()
	panel.Layout().update()
	panel.SetDirty(DirtyTypeGenerated)
	cse := panel.childScrollEvents[target]
	target.RemoveEvent(EventTypeDown, cse.down)
	target.RemoveEvent(EventTypeScroll, cse.scroll)
	delete(panel.childScrollEvents, target)
}

func (panel *Panel) Child(index int) UI {
	return FirstOnEntity(panel.entity.Children[index])
}

func (panel *Panel) SetSpeed(speed float32) {
	panel.scrollSpeed = speed
}

func (panel *Panel) recreateDrawing() {
	panel.shaderData.Destroy()
	proxy := panel.shaderData
	proxy.CancelDestroy()
	panel.shaderData = proxy
}

func (panel *Panel) removeDrawing() {
	panel.recreateDrawing()
	panel.drawing = rendering.Drawing{}
}

func (panel *Panel) SetColor(bgColor matrix.Color) {
	panel.ensureBGExists(nil)
	hasBlending := panel.shaderData.FgColor.A() < 1.0
	shouldBlend := bgColor.A() < 1.0
	if hasBlending != shouldBlend {
		panel.recreateDrawing()
		panel.drawing.UseBlending = shouldBlend
		panel.host.Drawings.AddDrawing(panel.drawing)
	}
	panel.shaderData.FgColor = bgColor
}

func (panel *Panel) SetScrollX(value float32) {
	panel.scroll.SetX(max(0.0, min(panel.maxScroll.X(), value)))
	panel.SetDirty(DirtyTypeLayout)
}

func (panel *Panel) SetScrollY(value float32) {
	panel.scroll.SetY(max(0.0, min(panel.maxScroll.Y(), value)))
	panel.SetDirty(DirtyTypeLayout)
}

func (panel *Panel) ResetScroll() {
	panel.scroll = matrix.Vec2Zero()
}

func (panel *Panel) ensureBGExists(tex *rendering.Texture) {
	if !panel.drawing.IsValid() {
		if tex == nil {
			tex, _ = panel.host.TextureCache().Texture(
				assets.TextureSquare, rendering.TextureFilterLinear)
		}
		shader := panel.host.ShaderCache().ShaderFromDefinition(
			assets.ShaderDefinitionUI)
		panel.shaderData.BorderLen = matrix.Vec2{8.0, 8.0}
		panel.shaderData.BgColor = panel.color
		panel.shaderData.FgColor = panel.color
		panel.shaderData.UVs = matrix.Vec4{0.0, 0.0, 1.0, 1.0}
		panel.shaderData.Size2D = matrix.Vec4{0.0, 0.0,
			float32(tex.Width), float32(tex.Height)}
		panel.drawing = rendering.Drawing{
			Renderer:   panel.host.Window.Renderer,
			Shader:     shader,
			Mesh:       rendering.NewMeshQuad(panel.host.MeshCache()),
			Textures:   []*rendering.Texture{tex},
			ShaderData: &panel.shaderData,
			Transform:  &panel.entity.Transform,
		}
		panel.host.Drawings.AddDrawing(panel.drawing)
	}
}

func (panel *Panel) SetBackground(tex *rendering.Texture) {
	if panel.drawing.IsValid() {
		panel.recreateDrawing()
		panel.drawing.Textures[0] = tex
		panel.host.Drawings.AddDrawing(panel.drawing)
	}
}

func (panel *Panel) RemoveBackground() {
	panel.recreateDrawing()
}

func (panel *Panel) IsScrolling() bool {
	return panel.isScrolling
}

func (panel *Panel) Freeze() {
	panel.frozen = true
}

func (panel *Panel) UnFreeze() {
	panel.frozen = false
}

func (panel *Panel) IsFrozen() bool {
	return panel.frozen
}

func (panel *Panel) SetScrollDirection(direction PanelScrollDirection) {
	panel.scrollDirection = direction
	panel.SetDirty(DirtyTypeLayout)
}

func (panel *Panel) ScrollDirection() PanelScrollDirection { return panel.scrollDirection }

func (panel *Panel) SetUseBlending(useBlending bool) {
	panel.recreateDrawing()
	panel.drawing.UseBlending = useBlending
	panel.host.Drawings.AddDrawing(panel.drawing)
}
