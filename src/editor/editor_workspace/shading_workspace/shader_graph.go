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
)

const (
	shadingGraphMenuBarHeight   = float32(24.0)
	shadingGraphStatusBarHeight = float32(20.8)
)

type shaderGraph struct {
	host  *engine.Host
	uiMan ui.Manager
	root  *ui.Panel
	nodes []*shaderGraphNode
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
	g.nodes = nil
	g.root = nil
	g.uiMan.Shutdown()
}

func (g *shaderGraph) Open() {
	if g.root == nil {
		return
	}
	g.root.Base().Show()
	g.applyLayout()
}

func (g *shaderGraph) Close() {
	if g.root != nil {
		g.root.Base().Hide()
	}
	for i := range g.nodes {
		g.nodes[i].stopDrag()
	}
}

func (g *shaderGraph) Update() {
	if g.root == nil || !g.root.Base().IsActive() {
		return
	}
	g.applyLayout()
	for i := range g.nodes {
		g.nodes[i].Update()
	}
}

func (g *shaderGraph) CreateNode(spec shaderGraphNodeSpec, position matrix.Vec2) *shaderGraphNode {
	if g.root == nil {
		return nil
	}
	node := &shaderGraphNode{}
	node.Initialize(g.host, &g.uiMan, g.root, spec, position)
	g.nodes = append(g.nodes, node)
	return node
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
