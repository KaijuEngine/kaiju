/******************************************************************************/
/* render_graph.go                                                            */
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
	renderGraphMenuBarHeight   = float32(24.0)
	renderGraphStatusBarHeight = float32(20.8)
	renderGraphMinZoom         = matrix.Float(0.35)
	renderGraphMaxZoom         = matrix.Float(1.0)
	renderGraphZoomStep        = matrix.Float(0.1)
)

type renderGraph struct {
	host                  *engine.Host
	history               *memento.History
	uiMan                 ui.Manager
	root                  *ui.Panel
	nodes                 []*renderGraphNode
	comments              []*renderGraphComment
	selected              []*renderGraphNode
	selectedComment       *renderGraphComment
	availableNodeZSlots   []int
	connections           []*renderGraphConnection
	selectionBox          *ui.Panel
	pendingFrom           *renderGraphPort
	pendingVisual         *renderGraphSpline
	pan                   matrix.Vec2
	zoom                  matrix.Float
	zoomBlocked           func(matrix.Vec2) bool
	inputBlocked          func(matrix.Vec2) bool
	connectionDropOnBlank func(*renderGraphPort, matrix.Vec2, matrix.Vec2)
	panning               bool
	panMouse              matrix.Vec2
	hasPanMouse           bool
	boxSelecting          bool
	boxStart              matrix.Vec2
	viewport              renderGraphViewport
	selectTexture         func(current string, onSelect func(string), onClose func())
	textureName           func(id string) string
}

type renderGraphViewport struct {
	x, y, width, height float32
}

func (g *renderGraph) Initialize(host *engine.Host, history *memento.History) {
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
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
	)
	g.selectionBox.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	g.selectionBox.Base().Layout().SetZ(30)
	g.selectionBox.Base().Hide()
	g.root.AddChild(g.selectionBox.Base())
}

func (g *renderGraph) Shutdown() {
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

func (g *renderGraph) Open() {
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

func (g *renderGraph) Close() {
	if g.root != nil {
		g.root.Base().Hide()
	}
	for i := range g.nodes {
		g.nodes[i].stopDrag()
	}
	for i := range g.comments {
		g.comments[i].stopInteraction()
	}
	g.panning = false
	g.hasPanMouse = false
	g.cancelBoxSelection()
	for i := range g.connections {
		g.connections[i].Hide()
	}
	g.cancelPendingConnection()
}

func (g *renderGraph) IsFocusedOnInput() bool {
	return g != nil && g.uiMan.Group.IsFocusedOnInput()
}

func (g *renderGraph) Update() {
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
	for i := range g.comments {
		g.comments[i].Update()
	}
	for i := range g.connections {
		g.connections[i].Update()
	}
	g.updatePendingConnection()
}

func (g *renderGraph) CreateNode(spec renderGraphNodeSpec, position matrix.Vec2) *renderGraphNode {
	return g.createNode("", spec, position, "")
}

func (g *renderGraph) CreateCatalogNode(typeID string, position matrix.Vec2) (*renderGraphNode, bool) {
	spec, ok := renderGraphNodeCatalogSpec(typeID)
	if !ok {
		return nil, false
	}
	if g.root == nil || g.host == nil {
		node := g.createNodeFromSnapshot(RenderGraphNode{
			ID:       g.nextNodeID(),
			Type:     typeID,
			Position: position,
		})
		return node, node != nil
	}
	node := g.createNode(typeID, spec, position, "")
	return node, node != nil
}

func (g *renderGraph) createNode(typeID string, spec renderGraphNodeSpec, position matrix.Vec2, id string) *renderGraphNode {
	if g.root == nil {
		return nil
	}
	node := &renderGraphNode{}
	node.id = id
	if node.id == "" {
		node.id = g.nextNodeID()
	}
	node.typeID = typeID
	node.Initialize(g, g.host, &g.uiMan, g.root, spec, position)
	g.assignNodeZSlot(node)
	g.nodes = append(g.nodes, node)
	g.applySelectionZOrder()
	return node
}

func (g *renderGraph) CreateComment(position matrix.Vec2, size matrix.Vec2, label string) *renderGraphComment {
	comment := RenderGraphComment{
		ID:       g.nextCommentID(),
		Label:    label,
		Position: position,
		Size:     renderGraphCommentSizeOrDefault(size),
	}
	return g.createCommentFromSnapshot(comment)
}

func (g *renderGraph) createCommentFromSnapshot(comment RenderGraphComment) *renderGraphComment {
	if g == nil || comment.ID == "" {
		return nil
	}
	if existing := g.commentByID(comment.ID); existing != nil {
		return existing
	}
	comment.Size = renderGraphCommentSizeOrDefault(comment.Size)
	var created *renderGraphComment
	if g.root != nil && g.host != nil {
		created = &renderGraphComment{}
		created.Initialize(g, g.host, &g.uiMan, g.root, comment)
	} else {
		created = &renderGraphComment{
			graph:    g,
			id:       comment.ID,
			label:    comment.Label,
			position: comment.Position,
			size:     comment.Size,
		}
	}
	g.comments = append(g.comments, created)
	return created
}

func (g *renderGraph) createNodeFromSnapshot(node RenderGraphNode) *renderGraphNode {
	if g == nil || node.ID == "" {
		return nil
	}
	if existing := g.nodeByID(node.ID); existing != nil {
		return existing
	}
	spec, ok := renderGraphNodeCatalogSpec(node.Type)
	if !ok {
		return nil
	}
	var created *renderGraphNode
	if g.root != nil && g.host != nil {
		created = g.createNode(node.Type, spec, node.Position, node.ID)
	} else {
		created = &renderGraphNode{
			graph:    g,
			id:       node.ID,
			typeID:   node.Type,
			position: node.Position,
			values:   make(map[string]renderGraphNodeFieldValue, len(spec.Fields)),
		}
		for i := range spec.Inputs {
			created.inputs = append(created.inputs, &renderGraphPort{
				graph: g,
				node:  created,
				spec:  spec.Inputs[i],
				index: i,
			})
		}
		for i := range spec.Outputs {
			created.outputs = append(created.outputs, &renderGraphPort{
				graph:  g,
				node:   created,
				spec:   spec.Outputs[i],
				output: true,
				index:  i,
			})
		}
		g.nodes = append(g.nodes, created)
	}
	for key, value := range renderGraphFieldValuesToNode(node.Values) {
		created.values[key] = value
	}
	created.applyFieldValues()
	return created
}

func (g *renderGraph) RemoveNode(id string) bool {
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
	g.releaseNodeZSlot(node)
	if node != nil && node.root != nil && g.host != nil {
		g.host.DestroyEntity(node.root.Base().Entity())
	}
	g.nodes = slices.Delete(g.nodes, nodeIndex, nodeIndex+1)
	g.setSelectionNodes(g.selected)
	return true
}

func (g *renderGraph) RemoveComment(id string) bool {
	if g == nil || id == "" {
		return false
	}
	commentIndex := -1
	for i := range g.comments {
		if g.comments[i] != nil && g.comments[i].id == id {
			commentIndex = i
			break
		}
	}
	if commentIndex < 0 {
		return false
	}
	comment := g.comments[commentIndex]
	if g.selectedComment == comment {
		g.setSelectedComment(nil)
	}
	if comment != nil && comment.root != nil && g.host != nil {
		g.host.DestroyEntity(comment.root.Base().Entity())
	}
	g.comments = slices.Delete(g.comments, commentIndex, commentIndex+1)
	return true
}

func (g *renderGraph) DeleteSelectedNodes() bool {
	if g == nil || len(g.selected) == 0 {
		if g == nil || g.selectedComment == nil {
			return false
		}
		comment := renderGraphCommentFromRenderGraphComment(g.selectedComment)
		if !g.RemoveComment(comment.ID) {
			return false
		}
		if g.history != nil {
			g.history.Add(&renderGraphCommentDeleteHistory{
				graph:   g,
				comment: comment,
			})
		}
		return true
	}
	nodes := make([]RenderGraphNode, 0, len(g.selected))
	nodeIDs := make(map[string]struct{}, len(g.selected))
	for i := range g.selected {
		node := g.selected[i]
		if node == nil || node.id == "" {
			continue
		}
		nodes = append(nodes, renderGraphNodeFromRenderGraphNode(node))
		nodeIDs[node.id] = struct{}{}
	}
	if len(nodes) == 0 {
		return false
	}
	connections := g.connectionsTouchingNodes(nodeIDs)
	for i := range nodes {
		g.RemoveNode(nodes[i].ID)
	}
	g.setSelectionNodes(nil)
	if g.history != nil {
		g.history.Add(&renderGraphNodeDeleteHistory{
			graph:       g,
			nodes:       nodes,
			connections: connections,
		})
	}
	return true
}

func (g *renderGraph) connectionsTouchingNodes(nodeIDs map[string]struct{}) []RenderGraphConnection {
	if g == nil || len(nodeIDs) == 0 {
		return nil
	}
	connections := make([]RenderGraphConnection, 0)
	seen := make(map[RenderGraphConnection]struct{})
	for i := range g.connections {
		connection := g.connections[i]
		if connection == nil {
			continue
		}
		renderConnection, ok := connection.renderConnection()
		if !ok {
			continue
		}
		_, outputDeleted := nodeIDs[renderConnection.Output.Node]
		_, inputDeleted := nodeIDs[renderConnection.Input.Node]
		if !outputDeleted && !inputDeleted {
			continue
		}
		if _, exists := seen[renderConnection]; exists {
			continue
		}
		seen[renderConnection] = struct{}{}
		connections = append(connections, renderConnection)
	}
	return connections
}

func (g *renderGraph) CreateConnection(a, b *renderGraphPort) *renderGraphConnection {
	output, input, ok := renderGraphConnectionPorts(a, b)
	if !ok {
		return nil
	}
	outputRef, inputRef, ok := renderGraphConnectionRefs(output, input)
	if ok {
		if existing := g.connectionByRefs(outputRef, inputRef); existing != nil {
			return existing
		}
	}
	g.removeConnectionsTouchingPort(input)
	connection := &renderGraphConnection{}
	if g.root != nil {
		connection.Initialize(g.host, g.root, output, input)
	} else {
		connection.output = output
		connection.input = input
	}
	g.connections = append(g.connections, connection)
	return connection
}

func (g *renderGraph) ConnectPorts(a, b *renderGraphPort) *renderGraphConnection {
	output, input, ok := renderGraphConnectionPorts(a, b)
	if !ok {
		return nil
	}
	outputRef, inputRef, ok := renderGraphConnectionRefs(output, input)
	if !ok {
		return nil
	}
	if existing := g.connectionByRefs(outputRef, inputRef); existing != nil {
		return existing
	}
	replaced := g.removeConnectionsTouchingPort(input)
	connection := g.CreateConnection(output, input)
	if connection == nil {
		for i := range replaced {
			g.createConnectionRef(replaced[i])
		}
		return nil
	}
	if g.history != nil {
		g.history.Add(&renderGraphConnectionHistory{
			graph:    g,
			output:   outputRef,
			input:    inputRef,
			replaced: replaced,
		})
	}
	return connection
}

func (g *renderGraph) RemoveConnection(outputRef, inputRef RenderGraphPortRef) bool {
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

func (g *renderGraph) DisconnectPort(port *renderGraphPort) bool {
	if port == nil {
		return false
	}
	removed := g.removeConnectionsTouchingPort(port)
	if len(removed) == 0 {
		return false
	}
	if g.history != nil {
		g.history.Add(&renderGraphConnectionDisconnectHistory{
			graph:       g,
			connections: removed,
		})
	}
	return true
}

func (g *renderGraph) removeConnectionsTouchingPort(port *renderGraphPort) []RenderGraphConnection {
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

func (g *renderGraph) removeConnectionRef(connection RenderGraphConnection) bool {
	if g == nil {
		return false
	}
	return g.RemoveConnection(connection.Output, connection.Input)
}

func (g *renderGraph) createConnectionRef(connection RenderGraphConnection) *renderGraphConnection {
	if g == nil {
		return nil
	}
	return g.createConnectionFromRefs(connection.Output, connection.Input)
}

func (g *renderGraph) connectionByRefs(outputRef, inputRef RenderGraphPortRef) *renderGraphConnection {
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

func (g *renderGraph) createConnectionFromRefs(outputRef, inputRef RenderGraphPortRef) *renderGraphConnection {
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

func (g *renderGraph) portByRef(nodeID string, portIndex int, output bool) *renderGraphPort {
	if g == nil {
		return nil
	}
	node := g.nodeByID(nodeID)
	if node == nil {
		return nil
	}
	if output {
		return node.Output(portIndex)
	}
	return node.Input(portIndex)
}

func (g *renderGraph) SetViewport(x, y, width, height float32) {
	g.viewport = renderGraphViewport{
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}
}

func (g *renderGraph) clear() {
	g.setSelectionNodes(nil)
	g.setSelectedComment(nil)
	g.cancelBoxSelection()
	for i := range g.connections {
		g.connections[i].Destroy()
	}
	for i := range g.comments {
		if g.comments[i].root != nil && g.host != nil {
			g.host.DestroyEntity(g.comments[i].root.Base().Entity())
		}
	}
	for i := range g.nodes {
		if g.nodes[i].root != nil && g.host != nil {
			g.host.DestroyEntity(g.nodes[i].root.Base().Entity())
		}
	}
	g.comments = nil
	g.nodes = nil
	g.availableNodeZSlots = nil
	g.connections = nil
	g.cancelPendingConnection()
}

func (g *renderGraph) nextCommentID() string {
	for i := len(g.comments) + 1; ; i++ {
		id := "comment-" + fmt.Sprint(i)
		if g.itemIDAvailable(id) {
			return id
		}
	}
}

func (g *renderGraph) itemIDAvailable(id string) bool {
	if g == nil || id == "" {
		return false
	}
	for i := range g.nodes {
		if g.nodes[i] != nil && g.nodes[i].id == id {
			return false
		}
	}
	for i := range g.comments {
		if g.comments[i] != nil && g.comments[i].id == id {
			return false
		}
	}
	return true
}

func (g *renderGraph) nextNodeID() string {
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

func (g *renderGraph) beginConnection(port *renderGraphPort) {
	if port == nil || g.root == nil {
		return
	}
	g.pendingFrom = port
	if g.pendingVisual == nil {
		g.pendingVisual = &renderGraphSpline{}
		g.pendingVisual.Initialize(g.host, g.root)
	}
	g.pendingVisual.Show()
	g.pendingVisual.SetColor(port.Color())
	g.updatePendingConnection()
}

func (g *renderGraph) finishConnection(port *renderGraphPort) {
	if g.pendingFrom == nil {
		return
	}
	if port != nil {
		if g.pendingFrom.CanConnect(port) {
			g.ConnectPorts(g.pendingFrom, port)
		}
		g.cancelPendingConnection()
		return
	}
	if g.host != nil && g.host.Window != nil && g.screenPositionInside(g.host.Window.Mouse.ScreenPosition()) {
		source := g.pendingFrom
		viewPosition := g.screenToViewPosition(g.host.Window.Mouse.ScreenPosition())
		createPosition := g.graphPositionFromView(viewPosition)
		if g.connectionDropOnBlank != nil {
			g.connectionDropOnBlank(source, viewPosition, createPosition)
		}
	}
	g.cancelPendingConnection()
}

func (g *renderGraph) cancelPendingConnection() {
	g.pendingFrom = nil
	if g.pendingVisual != nil {
		g.pendingVisual.Hide()
	}
}

func (g *renderGraph) updatePendingConnection() {
	if g.pendingFrom == nil || g.host == nil || g.host.Window == nil || g.pendingVisual == nil {
		return
	}
	mouse := &g.host.Window.Mouse
	if mouse.Released(hid.MouseButtonLeft) {
		g.finishConnection(nil)
		return
	}
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

func (g *renderGraph) applyLayout() bool {
	if g.host == nil || g.host.Window == nil || g.root == nil {
		return false
	}
	viewport := g.viewport
	if viewport.width <= 0 || viewport.height <= 0 {
		windowWidth := float32(g.host.Window.Width())
		windowHeight := float32(g.host.Window.Height())
		contentHeight := max(1, windowHeight-renderGraphMenuBarHeight-renderGraphStatusBarHeight)
		stageHeight := max(1, contentHeight*0.5)
		viewport = renderGraphViewport{
			x:      0,
			y:      renderGraphMenuBarHeight + stageHeight,
			width:  max(1, windowWidth),
			height: max(1, contentHeight-stageHeight),
		}
	}
	layout := g.root.Base().Layout()
	resized := layout.Scale(max(1, viewport.width), max(1, viewport.height))
	layout.SetOffset(viewport.x, viewport.y)
	return resized
}
