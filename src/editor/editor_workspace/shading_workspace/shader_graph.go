/******************************************************************************/
/* shader_graph.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

const (
	shadingGraphMenuBarHeight   = float32(24.0)
	shadingGraphStatusBarHeight = float32(20.8)
)

type shaderGraph struct {
	host          *engine.Host
	uiMan         ui.Manager
	root          *ui.Panel
	nodes         []*shaderGraphNode
	connections   []*shaderGraphConnection
	pendingFrom   *shaderGraphPort
	pendingVisual *shaderGraphSpline
	pan           matrix.Vec2
	panning       bool
	panMouse      matrix.Vec2
	hasPanMouse   bool
}

func (g *shaderGraph) Initialize(host *engine.Host) {
	g.host = host
	g.uiMan.Init(host)
	g.root = g.uiMan.Add().ToPanel()
	g.root.Init(nil, ui.ElementTypePanel)
	g.root.DontFitContent()
	g.root.SetOverflow(ui.OverflowHidden)
	g.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	g.root.Base().Layout().SetZ(4)
	g.root.Base().Hide()
}

func (g *shaderGraph) Shutdown() {
	for i := range g.connections {
		g.connections[i].Destroy()
	}
	if g.pendingVisual != nil {
		g.pendingVisual.Destroy()
	}
	g.nodes = nil
	g.connections = nil
	g.pendingFrom = nil
	g.pendingVisual = nil
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
	for i := range g.connections {
		g.connections[i].Hide()
	}
	g.cancelPendingConnection()
}

func (g *shaderGraph) Update() {
	if g.root == nil || !g.root.Base().IsActive() {
		return
	}
	g.applyLayout()
	g.updatePan()
	for i := range g.nodes {
		g.nodes[i].Update()
	}
	for i := range g.connections {
		g.connections[i].Update()
	}
	g.updatePendingConnection()
}

func (g *shaderGraph) CreateNode(spec shaderGraphNodeSpec, position matrix.Vec2) *shaderGraphNode {
	if g.root == nil {
		return nil
	}
	node := &shaderGraphNode{}
	node.Initialize(g, g.host, &g.uiMan, g.root, spec, position)
	g.nodes = append(g.nodes, node)
	return node
}

func (g *shaderGraph) CreateConnection(a, b *shaderGraphPort) *shaderGraphConnection {
	output, input, ok := shaderGraphConnectionPorts(a, b)
	if !ok || g.root == nil {
		return nil
	}
	connection := &shaderGraphConnection{}
	connection.Initialize(g.host, g.root, output, input)
	g.connections = append(g.connections, connection)
	return connection
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
		g.CreateConnection(g.pendingFrom, port)
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

func (g *shaderGraph) applyLayout() {
	if g.host == nil || g.host.Window == nil || g.root == nil {
		return
	}
	windowWidth := float32(g.host.Window.Width())
	windowHeight := float32(g.host.Window.Height())
	width := max(1, windowWidth*0.5)
	height := max(1, windowHeight-shadingGraphMenuBarHeight-shadingGraphStatusBarHeight)
	layout := g.root.Base().Layout()
	layout.Scale(width, height)
	layout.SetOffset(windowWidth*0.5, shadingGraphMenuBarHeight)
}
