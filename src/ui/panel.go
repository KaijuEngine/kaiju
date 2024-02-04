package ui

import (
	"kaiju/assets"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
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

type localData interface {
}

type ContentFit = int32

const (
	ContentFitNone = iota
	ContentFitWidth
	ContentFitHeight
	ContentFitBoth
)

type Panel struct {
	uiBase
	scroll, offset, maxScroll     matrix.Vec2
	scrollSpeed                   float32
	scrollDirection               PanelScrollDirection
	scrollEvent                   engine.EventId
	borderStyle                   [4]BorderStyle
	color                         matrix.Color
	drawing                       rendering.Drawing
	localData                     localData
	innerUpdate                   func(deltaTime float64)
	isScrolling, dragging, frozen bool
	isButton                      bool
	fitContent                    ContentFit
}

func NewPanel(host *engine.Host, texture *rendering.Texture, anchor Anchor) *Panel {
	panel := &Panel{
		scrollEvent:     -1,
		scrollSpeed:     30.0,
		scrollDirection: PanelScrollDirectionVertical,
		color:           matrix.Color{1.0, 1.0, 1.0, 1.0},
		fitContent:      ContentFitBoth,
	}
	ts := matrix.Vec2Zero()
	if texture != nil {
		ts = texture.Size()
	}
	panel.updateId = host.Updater.AddUpdate(panel.update)
	panel.init(host, ts, anchor, panel)
	panel.scrollEvent = panel.AddEvent(EventTypeScroll, panel.onScroll)
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

func (panel *Panel) DontFitContentWidth() {
	switch panel.fitContent {
	case ContentFitBoth:
		panel.fitContent = ContentFitHeight
	case ContentFitWidth:
		panel.fitContent = ContentFitNone
	}
}

func (panel *Panel) DontFitContentHeight() {
	switch panel.fitContent {
	case ContentFitBoth:
		panel.fitContent = ContentFitWidth
	case ContentFitHeight:
		panel.fitContent = ContentFitNone
	}
}

func (panel *Panel) DontFitContent() {
	panel.fitContent = ContentFitNone
}

func (panel *Panel) FittingContent() bool {
	return panel.fitContent != ContentFitNone
}

func (panel *Panel) FitContentWidth() {
	switch panel.fitContent {
	case ContentFitNone:
		panel.fitContent = ContentFitWidth
	case ContentFitHeight:
		panel.fitContent = ContentFitBoth
	}
	if panel.dirtyType == DirtyTypeNone {
		panel.SetDirty(DirtyTypeLayout)
	} else {
		panel.SetDirty(DirtyTypeGenerated)
	}
}

func (panel *Panel) FitContentHeight() {
	switch panel.fitContent {
	case ContentFitNone:
		panel.fitContent = ContentFitHeight
	case ContentFitWidth:
		panel.fitContent = ContentFitBoth
	}
	if panel.dirtyType == DirtyTypeNone {
		panel.SetDirty(DirtyTypeLayout)
	} else {
		panel.SetDirty(DirtyTypeGenerated)
	}
}

func (panel *Panel) FitContent() {
	panel.fitContent = ContentFitBoth
	if panel.dirtyType == DirtyTypeNone {
		panel.SetDirty(DirtyTypeLayout)
	} else {
		panel.SetDirty(DirtyTypeGenerated)
	}
}

func (panel *Panel) onScroll() {
	mouse := &panel.host.Window.Mouse
	delta := mouse.Scroll()
	scroll := panel.scroll
	if !mouse.Scrolled() {
		pos := panel.cursorPos(&panel.host.Window.Cursor)
		delta = pos.Subtract(panel.downPos)
		delta[matrix.Vy] *= -1.0
	} else {
		panel.offset = panel.scroll
		delta.ScaleAssign(1.0 * panel.scrollSpeed)
	}
	if (panel.scrollDirection & PanelScrollDirectionHorizontal) != 0 {
		x := matrix.Clamp(delta.X()+panel.offset.X(), 0.0, panel.maxScroll.X())
		scroll.SetX(x)
	}
	if (panel.scrollDirection & PanelScrollDirectionVertical) != 0 {
		y := matrix.Clamp(delta.Y()+panel.offset.Y(), -panel.maxScroll.Y(), 0)
		scroll.SetY(y)
	}
	if !matrix.Vec2Approx(scroll, panel.scroll) {
		panel.scroll = scroll
		panel.SetDirty(DirtyTypeLayout)
		panel.isScrolling = true
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
		x, y := offsetX, offsetY
		switch e.Layout().Positioning() {
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
		e.Layout().rowLayoutOffset = matrix.Vec2{x, y}
		offsetX += e.Layout().PixelSize().Width() + e.Layout().margin.X() + e.Layout().margin.Z()
	}
}

func (panel *Panel) postLayoutUpdate() {
	if len(panel.entity.Children) == 0 {
		return
	}
	offsetStart := matrix.Vec2{-panel.scroll.X(), panel.scroll.Y()}
	rows := make([]rowBuilder, 0)
	ps := panel.layout.PixelSize()
	areaWidth := ps.X() - panel.layout.padding.X() - panel.layout.padding.Z()
	for _, kid := range panel.entity.Children {
		if !kid.IsActive() || kid.IsDestroyed() {
			continue
		}
		kui := FirstOnEntity(kid)
		if kui == nil {
			panic("No UI component on entity")
		}
		kLayout := kui.Layout()
		switch kLayout.Positioning() {
		case PositioningAbsolute:
			if kLayout.Anchor().IsTop() {
				kLayout.rowLayoutOffset.SetY(panel.layout.InnerOffset().Top() + panel.layout.padding.Top())
			} else if kLayout.Anchor().IsBottom() {
				kLayout.rowLayoutOffset.SetY(panel.layout.InnerOffset().Bottom() + panel.layout.padding.Bottom())
			}
			if kLayout.Anchor().IsLeft() {
				kLayout.rowLayoutOffset.SetX(panel.layout.InnerOffset().Left() + panel.layout.padding.Left())
			} else if kLayout.Anchor().IsRight() {
				kLayout.rowLayoutOffset.SetX(panel.layout.InnerOffset().Right() + panel.layout.padding.Right())
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
	xyOffset := matrix.Vec2{panel.layout.padding.X(), panel.layout.padding.Y()}
	nextPos := offsetStart.Add(xyOffset)
	for i := range rows {
		rows[i].setElements(panel.layout.padding.X(), nextPos[matrix.Vy])
		nextPos[matrix.Vy] += rows[i].Height()
	}
	nextPos[matrix.Vy] += panel.layout.padding.W()
	if panel.FittingContent() {
		bounds := matrix.Vec2{0, nextPos[matrix.Vy]}
		panelScale := panel.entity.Transform.WorldScale().Scale(0.5)
		for _, kid := range panel.entity.Children {
			pos := kid.Transform.Position()
			pos[matrix.Vx] += panelScale.X()
			pos[matrix.Vy] -= panelScale.Y()
			kui := FirstOnEntity(kid)
			var r, b matrix.Float
			if lbl, ok := kui.(*Label); ok {
				maxWidth := matrix.Float(1000000.0)
				if !panel.entity.IsRoot() {
					maxWidth = panel.entity.Parent.Transform.WorldScale().X()
				}
				size := lbl.measure(maxWidth)
				r = matrix.Abs(pos.X()) + size.X()
				b = matrix.Abs(pos.Y()) + size.Y()
			} else {
				size := kid.Transform.WorldScale().Scale(0.5)
				r = matrix.Abs(pos.X()) + size.X()
				b = matrix.Abs(pos.Y()) + size.Y()
			}
			bounds = matrix.Vec2{max(bounds.X(), r), max(bounds.Y(), b)}
		}
		if panel.fitContent == ContentFitWidth {
			panel.layout.ScaleWidth(max(1, bounds.X()))
		} else if panel.fitContent == ContentFitHeight {
			panel.layout.ScaleHeight(max(1, bounds.Y()))
		} else if panel.fitContent == ContentFitBoth {
			panel.layout.Scale(max(1, bounds.X()), max(1, bounds.Y()))
		}
	}
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
}

func (panel *Panel) AddChild(target UI) {
	target.Entity().SetParent(panel.entity)
	if panel.group != nil {
		target.SetGroup(panel.group)
	}
	target.Layout().update()
	panel.SetDirty(DirtyTypeGenerated)
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
	target.setScissor(matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax})
	target.Layout().update()
	panel.layout.update()
	panel.SetDirty(DirtyTypeGenerated)
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
		panel.textureSize = tex.Size()
		panel.shaderData.setSize2d(panel, panel.textureSize.X(), panel.textureSize.Y())
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
func (panel *Panel) BorderSize() matrix.Vec4               { return panel.layout.Border() }
func (panel *Panel) BorderStyle() [4]BorderStyle           { return panel.borderStyle }
func (panel *Panel) BorderColor() [4]matrix.Color          { return panel.shaderData.BorderColor }
func (panel *Panel) SetBorderRadius(topLeft, topRight, bottomRight, bottomLeft float32) {
	panel.shaderData.BorderRadius = matrix.Vec4{topLeft, topRight, bottomRight, bottomLeft}
}

func (panel *Panel) SetBorderSize(left, top, right, bottom float32) {
	panel.layout.SetBorder(left, top, right, bottom)
	// TODO:  If there isn't a border, it should be transparent when created
	panel.ensureBGExists(nil)
	panel.shaderData.BorderSize = panel.layout.Border()
}

func (panel *Panel) SetBorderStyle(left, top, right, bottom BorderStyle) {
	panel.borderStyle = [4]BorderStyle{left, top, right, bottom}
}

func (panel *Panel) SetBorderColor(left, top, right, bottom matrix.Color) {
	panel.shaderData.BorderColor = [4]matrix.Color{left, top, right, bottom}
}

func (panel *Panel) SetUseBlending(useBlending bool) {
	panel.recreateDrawing()
	panel.drawing.UseBlending = useBlending
	panel.host.Drawings.AddDrawing(panel.drawing)
}
