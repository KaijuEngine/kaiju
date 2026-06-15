/******************************************************************************/
/* schema_graph.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
)

type schemaGraphViewport struct {
	x, y, width, height float32
}

type schemaGraph struct {
	host        *engine.Host
	uiMan       ui.Manager
	root        *ui.Panel
	nodes       []*schemaNode
	rootNodes   []*schemaNode
	viewport    schemaGraphViewport
	layoutDirty bool
}

func (g *schemaGraph) Initialize(host *engine.Host) {
	g.host = host
	g.uiMan.Init(host)
	g.root = g.uiMan.Add().ToPanel()
	g.root.Init(nil, ui.ElementTypePanel)
	g.root.DontFitContent()
	g.root.SetOverflow(ui.OverflowHidden)
	g.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	g.root.Base().Layout().SetZ(4)
	g.root.Base().Hide()
	g.layoutDirty = true
}

func (g *schemaGraph) Shutdown() {
	g.clear()
	g.root = nil
	g.uiMan.Shutdown()
	g.host = nil
}

func (g *schemaGraph) Open() {
	if g.root == nil {
		return
	}
	g.root.Base().Show()
	g.applyViewport()
	g.reflow()
}

func (g *schemaGraph) Close() {
	if g.root != nil {
		g.root.Base().Hide()
	}
}

func (g *schemaGraph) Update() {
	if g.root == nil || !g.root.Base().IsActive() {
		return
	}
	if g.applyViewport() {
		g.layoutDirty = true
	}
	if g.layoutDirty {
		g.reflow()
	}
}

func (g *schemaGraph) SetViewport(x, y, width, height float32) {
	viewport := schemaGraphViewport{x: x, y: y, width: width, height: height}
	if g.viewport == viewport {
		return
	}
	g.viewport = viewport
	g.layoutDirty = true
}

func (g *schemaGraph) CreateRootNode(kind schemaNodeKind) *schemaNode {
	return g.createNode(kind, nil)
}

func (g *schemaGraph) AddProperty(parent *schemaNode) *schemaNode {
	if parent == nil || parent.kind != schemaNodeKindProperties {
		return nil
	}
	return g.createNode(schemaNodeKindProperty, parent)
}

func (g *schemaGraph) NodeCount() int {
	return len(g.nodes)
}

func (g *schemaGraph) createNode(kind schemaNodeKind, parent *schemaNode) *schemaNode {
	if g.root == nil || g.host == nil {
		return nil
	}
	spec, ok := schemaNodeSpecForKind(kind)
	if !ok {
		return nil
	}
	node := &schemaNode{
		graph:  g,
		id:     g.nextNodeID(),
		kind:   kind,
		parent: parent,
	}
	node.Initialize(&g.uiMan, g.root, spec)
	if parent == nil {
		g.rootNodes = append(g.rootNodes, node)
	} else {
		parent.children = append(parent.children, node)
	}
	g.nodes = append(g.nodes, node)
	g.layoutDirty = true
	return node
}

func (g *schemaGraph) clear() {
	if g.host != nil {
		for i := range g.nodes {
			if g.nodes[i] != nil && g.nodes[i].root != nil {
				g.host.DestroyEntity(g.nodes[i].root.Base().Entity())
			}
		}
	}
	g.nodes = nil
	g.rootNodes = nil
	g.layoutDirty = true
}

func (g *schemaGraph) nextNodeID() string {
	for i := len(g.nodes) + 1; ; i++ {
		id := "schema-node-" + fmt.Sprint(i)
		found := false
		for j := range g.nodes {
			if g.nodes[j] != nil && g.nodes[j].id == id {
				found = true
				break
			}
		}
		if !found {
			return id
		}
	}
}

func (g *schemaGraph) applyViewport() bool {
	if g.root == nil || g.host == nil || g.host.Window == nil {
		return false
	}
	viewport := g.viewport
	if viewport.width <= 0 || viewport.height <= 0 {
		windowWidth := float32(g.host.Window.Width())
		windowHeight := float32(g.host.Window.Height())
		contentHeight := max(1, windowHeight-schemaWorkspaceMenuBarHeight-schemaWorkspaceStatusBarHeight)
		viewport = schemaGraphViewport{
			x:      0,
			y:      schemaWorkspaceMenuBarHeight,
			width:  max(1, windowWidth),
			height: max(1, contentHeight-schemaWorkspaceActionBarHeight),
		}
	}
	layout := g.root.Base().Layout()
	resized := layout.Scale(max(1, viewport.width), max(1, viewport.height))
	layout.SetOffset(viewport.x, viewport.y)
	return resized
}
