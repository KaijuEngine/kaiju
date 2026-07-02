/******************************************************************************/
/* render_graph_workspace_create_menu.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"strings"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

const (
	renderGraphCreateMenuWidth  = matrix.Float(320)
	renderGraphCreateMenuHeight = matrix.Float(360)
)

type renderGraphCreateNodeMenu struct {
	workspace      *RenderGraphWorkspace
	root           *document.Element
	search         *document.Element
	list           *document.Element
	empty          *document.Element
	items          []*document.Element
	createPosition matrix.Vec2
	connection     renderGraphCreateNodeConnection
	open           bool
}

type renderGraphCreateNodeConnection struct {
	Active       bool
	SourceNode   string
	SourcePort   int
	SourceOutput bool
	SourceType   string
}

func (m *renderGraphCreateNodeMenu) Initialize(workspace *RenderGraphWorkspace) {
	m.workspace = workspace
	if workspace == nil || workspace.Doc == nil {
		return
	}
	m.root, _ = workspace.Doc.GetElementById("createNodeMenu")
	m.search, _ = workspace.Doc.GetElementById("createNodeSearch")
	m.list, _ = workspace.Doc.GetElementById("createNodeList")
	m.empty, _ = workspace.Doc.GetElementById("createNodeEmpty")
	m.items = workspace.Doc.GetElementsByClass("createNodeMenuItem")
	for _, item := range m.items {
		renderGraphCreateMenuAllowChildrenClickThrough(item)
	}
	m.Hide()
}

func (m *renderGraphCreateNodeMenu) Show(position, createPosition matrix.Vec2) {
	m.connection = renderGraphCreateNodeConnection{}
	m.show(position, createPosition)
}

func (m *renderGraphCreateNodeMenu) ShowForConnection(position, createPosition matrix.Vec2, source *renderGraphPort) {
	ref, ok := renderGraphPortRef(source)
	if !ok || source == nil {
		m.Show(position, createPosition)
		return
	}
	m.connection = renderGraphCreateNodeConnection{
		Active:       true,
		SourceNode:   ref.Node,
		SourcePort:   ref.Port,
		SourceOutput: source.output,
		SourceType:   source.spec.Type,
	}
	m.show(position, createPosition)
}

func (m *renderGraphCreateNodeMenu) show(position, createPosition matrix.Vec2) {
	if m.root == nil {
		return
	}
	m.open = true
	m.createPosition = createPosition
	m.root.UI.Show()
	m.positionRoot(position)
	if m.search != nil && m.search.UI != nil {
		input := m.search.UI.ToInput()
		input.SetTextWithoutEvent("")
		input.Focus()
	}
	m.Filter("")
}

func (m *renderGraphCreateNodeMenu) Hide() {
	m.open = false
	m.connection = renderGraphCreateNodeConnection{}
	if m.root != nil && m.root.UI != nil {
		m.root.UI.Hide()
	}
}

func (m *renderGraphCreateNodeMenu) Update() {
	if !m.open || m.workspace == nil || m.workspace.Host == nil || m.workspace.Host.Window == nil {
		return
	}
	if m.workspace.Host.Window.Keyboard.KeyDown(hid.KeyboardKeyEscape) {
		m.Hide()
	}
}

func (m *renderGraphCreateNodeMenu) CreatePosition() matrix.Vec2 {
	return m.createPosition
}

func (m *renderGraphCreateNodeMenu) Connection() renderGraphCreateNodeConnection {
	return m.connection
}

func (m *renderGraphCreateNodeMenu) BlocksGraphZoom(position matrix.Vec2) bool {
	return m.BlocksGraphInput(position)
}

func (m *renderGraphCreateNodeMenu) BlocksGraphInput(position matrix.Vec2) bool {
	if !m.open || m.root == nil || m.root.UI == nil || !m.root.UI.IsActive() {
		return false
	}
	layout := m.root.UI.Layout()
	offset := layout.Offset()
	size := layout.PixelSize()
	return position.X() >= offset.X() && position.Y() >= offset.Y() &&
		position.X() <= offset.X()+size.X() &&
		position.Y() <= offset.Y()+size.Y()
}

func (m *renderGraphCreateNodeMenu) Filter(query string) {
	query = strings.ToLower(strings.TrimSpace(query))
	visible := 0
	for _, item := range m.items {
		if item == nil || item.UI == nil {
			continue
		}
		matches := m.itemCompatible(item) &&
			(query == "" || renderGraphCreateMenuMatches(item.Attribute("data-search"), query))
		if matches {
			item.UI.Show()
			visible++
		} else {
			item.UI.Hide()
		}
	}
	if m.empty != nil && m.empty.UI != nil {
		if visible == 0 {
			m.empty.UI.Show()
		} else {
			m.empty.UI.Hide()
		}
	}
	if m.list != nil && m.list.UI != nil {
		m.list.UI.ToPanel().SetScrollY(0)
		m.list.UI.SetDirty(ui.DirtyTypeLayout)
	}
}

func (m *renderGraphCreateNodeMenu) itemCompatible(item *document.Element) bool {
	if item == nil || !m.connection.Active {
		return true
	}
	if item.Attribute("data-comment") == "true" {
		return false
	}
	for _, entry := range renderGraphNodeCatalog() {
		if entry.ID != item.Attribute("data-node-id") {
			continue
		}
		return renderGraphNodeCatalogEntryCompatible(entry, renderGraphNodePortCompatibility{
			Active:       true,
			SourceOutput: m.connection.SourceOutput,
			Type:         m.connection.SourceType,
		})
	}
	return false
}

func (m *renderGraphCreateNodeMenu) positionRoot(position matrix.Vec2) {
	if m.workspace == nil || m.workspace.renderGraphArea == nil {
		return
	}
	areaLayout := m.workspace.renderGraphArea.UI.Layout()
	areaSize := areaLayout.PixelSize()
	areaOffset := areaLayout.Offset()
	x := matrix.Clamp(position.X(), 8, max(8, areaSize.X()-renderGraphCreateMenuWidth-8))
	y := matrix.Clamp(position.Y(), 8, max(8, areaSize.Y()-renderGraphCreateMenuHeight-8))
	layout := m.root.UI.Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.SetOffset(float32(areaOffset.X()+x), float32(areaOffset.Y()+y))
	layout.SetZ(40)
}

func renderGraphCreateMenuMatches(search, query string) bool {
	search = strings.ToLower(search)
	for _, token := range strings.Fields(query) {
		if !strings.Contains(search, token) {
			return false
		}
	}
	return true
}

func renderGraphCreateMenuAllowChildrenClickThrough(element *document.Element) {
	if element == nil {
		return
	}
	for _, child := range element.Children {
		if child.UI != nil && child.UI.IsType(ui.ElementTypePanel) {
			child.UI.ToPanel().AllowClickThrough()
		}
		renderGraphCreateMenuAllowChildrenClickThrough(child)
	}
}

func (w *RenderGraphWorkspace) filterCreateNodeMenu(e *document.Element) {
	if e == nil || e.UI == nil {
		return
	}
	w.createNodeMenu.Filter(e.UI.ToInput().Text())
}

func (w *RenderGraphWorkspace) selectCreateNode(e *document.Element) {
	if e == nil {
		return
	}
	w.runCreateNodeAction(e.Attribute("data-node-id"))
}

func (w *RenderGraphWorkspace) selectCreateComment(e *document.Element) {
	w.runCreateCommentAction()
}

func (w *RenderGraphWorkspace) closeCreateNodeMenu(*document.Element) {
	w.createNodeMenu.Hide()
}

func (w *RenderGraphWorkspace) createNodeMenuPosition() matrix.Vec2 {
	mousePosition, ok := w.graphLocalMousePosition()
	if ok {
		return mousePosition
	}
	return w.defaultCreateNodeViewPosition()
}

func (w *RenderGraphWorkspace) defaultCreateNodePosition() matrix.Vec2 {
	return w.graph.graphPositionFromView(w.defaultCreateNodeViewPosition())
}

func (w *RenderGraphWorkspace) defaultCreateNodeViewPosition() matrix.Vec2 {
	if w.graph.root == nil {
		return matrix.NewVec2(48, 48)
	}
	size := w.graph.root.Base().Layout().PixelSize()
	offset := matrix.Float(w.createNodeCount % 10 * 18)
	return matrix.NewVec2(
		max(24, size.X()*0.5-renderGraphNodeWidth*0.5+offset),
		max(24, size.Y()*0.35+offset),
	)
}

func (w *RenderGraphWorkspace) graphLocalMousePosition() (matrix.Vec2, bool) {
	if w.ed == nil || w.ed.Host() == nil || w.ed.Host().Window == nil || w.graph.root == nil {
		return matrix.Vec2Zero(), false
	}
	mouse := w.ed.Host().Window.Mouse.ScreenPosition()
	offset := w.graph.root.Base().Layout().Offset()
	size := w.graph.root.Base().Layout().PixelSize()
	local := mouse.Subtract(offset)
	if local.X() < 0 || local.Y() < 0 || local.X() > size.X() || local.Y() > size.Y() {
		return matrix.Vec2Zero(), false
	}
	return local, true
}
