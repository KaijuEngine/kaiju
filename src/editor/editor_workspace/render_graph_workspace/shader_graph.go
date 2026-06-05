/******************************************************************************/
/* shader_graph.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"fmt"
	"slices"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

const (
	shadingGraphMenuBarHeight   = float32(24.0)
	shadingGraphStatusBarHeight = float32(20.8)
	shaderGraphMinZoom          = matrix.Float(0.35)
	shaderGraphMaxZoom          = matrix.Float(1.0)
	shaderGraphZoomStep         = matrix.Float(0.1)
)

type shaderGraph struct {
	host          *engine.Host
	history       *memento.History
	uiMan         ui.Manager
	root          *ui.Panel
	nodes         []*shaderGraphNode
	selected      []*shaderGraphNode
	connections   []*shaderGraphConnection
	selectionBox  *ui.Panel
	pendingFrom   *shaderGraphPort
	pendingVisual *shaderGraphSpline
	pan           matrix.Vec2
	zoom          matrix.Float
	zoomBlocked   func(matrix.Vec2) bool
	inputBlocked  func(matrix.Vec2) bool
	panning       bool
	panMouse      matrix.Vec2
	hasPanMouse   bool
	boxSelecting  bool
	boxStart      matrix.Vec2
	viewport      shaderGraphViewport
	selectTexture func(current string, onSelect func(string), onClose func())
	textureName   func(id string) string
}

type shaderGraphViewport struct {
	x, y, width, height float32
}

func (g *shaderGraph) Initialize(host *engine.Host, history *memento.History) {
	g.host = host
	g.history = history
	g.zoom = 1
	g.uiMan.Init(host)
	g.root = g.uiMan.Add().ToPanel()
	g.root.Init(nil, ui.ElementTypePanel)
	g.root.DontFitContent()
	g.root.SetOverflow(ui.OverflowHidden)
	g.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	g.root.Base().Layout().SetZ(4)
	g.root.Base().Hide()
	g.root.Base().AddEvent(ui.EventTypeDown, g.beginBoxSelectionFromInput)

	g.selectionBox = g.uiMan.Add().ToPanel()
	g.selectionBox.Init(nil, ui.ElementTypePanel)
	g.selectionBox.AllowClickThrough()
	g.selectionBox.DontFitContent()
	g.selectionBox.SetColor(matrix.NewColor(0.95, 0.72, 0.28, 0.18))
	g.selectionBox.SetBorderSize(1, 1, 1, 1)
	g.selectionBox.SetBorderColor(
		shaderGraphNodeSelectColor,
		shaderGraphNodeSelectColor,
		shaderGraphNodeSelectColor,
		shaderGraphNodeSelectColor,
	)
	g.selectionBox.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	g.selectionBox.Base().Layout().SetZ(30)
	g.selectionBox.Base().Hide()
	g.root.AddChild(g.selectionBox.Base())
}

func (g *shaderGraph) Shutdown() {
	g.clear()
	if g.pendingVisual != nil {
		g.pendingVisual.Destroy()
	}
	g.pendingFrom = nil
	g.pendingVisual = nil
	g.selectionBox = nil
	g.root = nil
	g.uiMan.Shutdown()
}

func (g *shaderGraph) Open() {
	if g.root == nil {
		return
	}
	g.hasPanMouse = false
	g.root.Base().Show()
	g.applyLayout()
	g.applyViewOffsets()
	for i := range g.connections {
		g.connections[i].Update()
	}
}

func (g *shaderGraph) Close() {
	if g.root != nil {
		g.root.Base().Hide()
	}
	for i := range g.nodes {
		g.nodes[i].stopDrag()
	}
	g.panning = false
	g.hasPanMouse = false
	g.cancelBoxSelection()
	for i := range g.connections {
		g.connections[i].Hide()
	}
	g.cancelPendingConnection()
}

func (g *shaderGraph) Update() {
	if g.root == nil || !g.root.Base().IsActive() {
		return
	}
	if g.applyLayout() {
		g.applyViewOffsets()
	}
	g.updateZoom()
	g.updatePan()
	g.updateBoxSelection()
	for i := range g.nodes {
		g.nodes[i].Update()
	}
	for i := range g.connections {
		g.connections[i].Update()
	}
	g.updatePendingConnection()
}

func (g *shaderGraph) CreateNode(spec shaderGraphNodeSpec, position matrix.Vec2) *shaderGraphNode {
	return g.createNode("", spec, position, "")
}

func (g *shaderGraph) CreateCatalogNode(typeID string, position matrix.Vec2) (*shaderGraphNode, bool) {
	spec, ok := shaderGraphNodeCatalogSpec(typeID)
	if !ok {
		return nil, false
	}
	node := g.createNode(typeID, spec, position, "")
	return node, node != nil
}

func (g *shaderGraph) createNode(typeID string, spec shaderGraphNodeSpec, position matrix.Vec2, id string) *shaderGraphNode {
	if g.root == nil {
		return nil
	}
	node := &shaderGraphNode{}
	node.id = id
	if node.id == "" {
		node.id = g.nextNodeID()
	}
	node.typeID = typeID
	node.Initialize(g, g.host, &g.uiMan, g.root, spec, position)
	g.nodes = append(g.nodes, node)
	return node
}

func (g *shaderGraph) createNodeFromSnapshot(node RenderGraphNode) *shaderGraphNode {
	if g == nil || node.ID == "" {
		return nil
	}
	if existing := g.nodeByID(node.ID); existing != nil {
		return existing
	}
	spec, ok := shaderGraphNodeCatalogSpec(node.Type)
	if !ok {
		return nil
	}
	var created *shaderGraphNode
	if g.root != nil && g.host != nil {
		created = g.createNode(node.Type, spec, node.Position, node.ID)
	} else {
		created = &shaderGraphNode{
			graph:    g,
			id:       node.ID,
			typeID:   node.Type,
			position: node.Position,
			values:   make(map[string]shaderGraphNodeFieldValue, len(spec.Fields)),
		}
		g.nodes = append(g.nodes, created)
	}
	for key, value := range renderGraphFieldValuesToNode(node.Values) {
		created.values[key] = value
	}
	created.applyFieldValues()
	return created
}

func (g *shaderGraph) RemoveNode(id string) bool {
	if g == nil || id == "" {
		return false
	}
	nodeIndex := -1
	for i := range g.nodes {
		if g.nodes[i] != nil && g.nodes[i].id == id {
			nodeIndex = i
			break
		}
	}
	if nodeIndex < 0 {
		return false
	}
	for i := len(g.connections) - 1; i >= 0; i-- {
		if g.connections[i] == nil || !g.connections[i].touchesNode(id) {
			continue
		}
		g.connections[i].Destroy()
		g.connections = slices.Delete(g.connections, i, i+1)
	}
	node := g.nodes[nodeIndex]
	if node != nil && node.root != nil && g.host != nil {
		g.host.DestroyEntity(node.root.Base().Entity())
	}
	g.nodes = slices.Delete(g.nodes, nodeIndex, nodeIndex+1)
	g.setSelectionNodes(g.selected)
	return true
}

func (g *shaderGraph) CreateConnection(a, b *shaderGraphPort) *shaderGraphConnection {
	output, input, ok := shaderGraphConnectionPorts(a, b)
	if !ok || g.root == nil {
		return nil
	}
	outputRef, inputRef, ok := shaderGraphConnectionRefs(output, input)
	if ok {
		if existing := g.connectionByRefs(outputRef, inputRef); existing != nil {
			return existing
		}
	}
	connection := &shaderGraphConnection{}
	connection.Initialize(g.host, g.root, output, input)
	g.connections = append(g.connections, connection)
	return connection
}

func (g *shaderGraph) ConnectPorts(a, b *shaderGraphPort) *shaderGraphConnection {
	output, input, ok := shaderGraphConnectionPorts(a, b)
	if !ok {
		return nil
	}
	outputRef, inputRef, ok := shaderGraphConnectionRefs(output, input)
	if !ok {
		return nil
	}
	if existing := g.connectionByRefs(outputRef, inputRef); existing != nil {
		return existing
	}
	connection := g.CreateConnection(output, input)
	if connection == nil {
		return nil
	}
	if g.history != nil {
		g.history.Add(&shaderGraphConnectionHistory{
			graph:  g,
			output: outputRef,
			input:  inputRef,
		})
	}
	return connection
}

func (g *shaderGraph) RemoveConnection(outputRef, inputRef RenderGraphPortRef) bool {
	if g == nil {
		return false
	}
	for i := range g.connections {
		connection := g.connections[i]
		if connection == nil || !connection.matches(outputRef, inputRef) {
			continue
		}
		connection.Destroy()
		g.connections = slices.Delete(g.connections, i, i+1)
		return true
	}
	return false
}

func (g *shaderGraph) DisconnectPort(port *shaderGraphPort) bool {
	if port == nil {
		return false
	}
	removed := g.removeConnectionsTouchingPort(port)
	if len(removed) == 0 {
		return false
	}
	if g.history != nil {
		g.history.Add(&shaderGraphConnectionDisconnectHistory{
			graph:       g,
			connections: removed,
		})
	}
	return true
}

func (g *shaderGraph) removeConnectionsTouchingPort(port *shaderGraphPort) []RenderGraphConnection {
	if g == nil {
		return nil
	}
	removed := make([]RenderGraphConnection, 0)
	for i := len(g.connections) - 1; i >= 0; i-- {
		connection := g.connections[i]
		if connection == nil || !connection.touchesPort(port) {
			continue
		}
		renderConnection, ok := connection.renderConnection()
		if ok {
			removed = append(removed, renderConnection)
		}
		connection.Destroy()
		g.connections = slices.Delete(g.connections, i, i+1)
	}
	return removed
}

func (g *shaderGraph) removeConnectionRef(connection RenderGraphConnection) bool {
	if g == nil {
		return false
	}
	return g.RemoveConnection(connection.Output, connection.Input)
}

func (g *shaderGraph) createConnectionRef(connection RenderGraphConnection) *shaderGraphConnection {
	if g == nil {
		return nil
	}
	return g.createConnectionFromRefs(connection.Output, connection.Input)
}

func (g *shaderGraph) connectionByRefs(outputRef, inputRef RenderGraphPortRef) *shaderGraphConnection {
	if g == nil {
		return nil
	}
	for i := range g.connections {
		connection := g.connections[i]
		if connection != nil && connection.matches(outputRef, inputRef) {
			return connection
		}
	}
	return nil
}

func (g *shaderGraph) createConnectionFromRefs(outputRef, inputRef RenderGraphPortRef) *shaderGraphConnection {
	if g == nil {
		return nil
	}
	outputNode := g.nodeByID(outputRef.Node)
	inputNode := g.nodeByID(inputRef.Node)
	if outputNode == nil || inputNode == nil {
		return nil
	}
	return g.CreateConnection(outputNode.Output(outputRef.Port), inputNode.Input(inputRef.Port))
}

func (g *shaderGraph) SetViewport(x, y, width, height float32) {
	g.viewport = shaderGraphViewport{
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}
}

func (g *shaderGraph) clear() {
	g.setSelectionNodes(nil)
	g.cancelBoxSelection()
	for i := range g.connections {
		g.connections[i].Destroy()
	}
	for i := range g.nodes {
		if g.nodes[i].root != nil && g.host != nil {
			g.host.DestroyEntity(g.nodes[i].root.Base().Entity())
		}
	}
	g.nodes = nil
	g.connections = nil
	g.cancelPendingConnection()
}

func (g *shaderGraph) nextNodeID() string {
	for i := len(g.nodes) + 1; ; i++ {
		id := "node-" + fmt.Sprint(i)
		found := false
		for j := range g.nodes {
			if g.nodes[j].id == id {
				found = true
				break
			}
		}
		if !found {
			return id
		}
	}
}

func (g *shaderGraph) beginConnection(port *shaderGraphPort) {
	if port == nil || g.root == nil {
		return
	}
	g.pendingFrom = port
	if g.pendingVisual == nil {
		g.pendingVisual = &shaderGraphSpline{}
		g.pendingVisual.Initialize(g.host, g.root)
	}
	g.pendingVisual.Show()
	g.pendingVisual.SetColor(port.Color())
	g.updatePendingConnection()
}

func (g *shaderGraph) finishConnection(port *shaderGraphPort) {
	if g.pendingFrom == nil {
		return
	}
	if port != nil && g.pendingFrom.CanConnect(port) {
		g.ConnectPorts(g.pendingFrom, port)
	}
	g.cancelPendingConnection()
}

func (g *shaderGraph) cancelPendingConnection() {
	g.pendingFrom = nil
	if g.pendingVisual != nil {
		g.pendingVisual.Hide()
	}
}

func (g *shaderGraph) updatePendingConnection() {
	if g.pendingFrom == nil || g.host == nil || g.host.Window == nil || g.pendingVisual == nil {
		return
	}
	mouse := &g.host.Window.Mouse
	if !mouse.Held(hid.MouseButtonLeft) && !mouse.Pressed(hid.MouseButtonLeft) {
		g.cancelPendingConnection()
		return
	}
	end := g.screenToViewPosition(mouse.ScreenPosition())
	start := g.viewPosition(g.pendingFrom.Anchor())
	if g.pendingFrom.output {
		g.pendingVisual.Update(start, end)
	} else {
		g.pendingVisual.Update(end, start)
	}
}

func (g *shaderGraph) applyLayout() bool {
	if g.host == nil || g.host.Window == nil || g.root == nil {
		return false
	}
	viewport := g.viewport
	if viewport.width <= 0 || viewport.height <= 0 {
		windowWidth := float32(g.host.Window.Width())
		windowHeight := float32(g.host.Window.Height())
		contentHeight := max(1, windowHeight-shadingGraphMenuBarHeight-shadingGraphStatusBarHeight)
		stageHeight := max(1, contentHeight*0.5)
		viewport = shaderGraphViewport{
			x:      0,
			y:      shadingGraphMenuBarHeight + stageHeight,
			width:  max(1, windowWidth),
			height: max(1, contentHeight-stageHeight),
		}
	}
	layout := g.root.Base().Layout()
	resized := layout.Scale(max(1, viewport.width), max(1, viewport.height))
	layout.SetOffset(viewport.x, viewport.y)
	return resized
}
