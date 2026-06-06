/******************************************************************************/
/* render_graph_selection.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"slices"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
)

type shaderGraphSelectionMode int

const (
	shaderGraphSelectionReplace shaderGraphSelectionMode = iota
	shaderGraphSelectionAppend
	shaderGraphSelectionToggle
	shaderGraphSelectionSubtract
)

func (g *shaderGraph) HasSelection() bool {
	return g != nil && (len(g.selected) > 0 || g.selectedComment != nil)
}

func (g *shaderGraph) Selection() []*shaderGraphNode {
	return slices.Clone(g.selected)
}

func (g *shaderGraph) IsSelected(node *shaderGraphNode) bool {
	if g == nil || node == nil {
		return false
	}
	return slices.Contains(g.selected, node)
}

func (g *shaderGraph) SelectNodeFromInput(node *shaderGraphNode) {
	if g == nil || node == nil || g.host == nil || g.host.Window == nil {
		return
	}
	g.SelectNodes([]*shaderGraphNode{node}, shaderGraphSelectionModeFromKeyboard(&g.host.Window.Keyboard))
}

func (g *shaderGraph) SelectCommentFromInput(comment *shaderGraphComment) {
	if g == nil || comment == nil {
		return
	}
	g.setSelectionNodes(nil)
	g.setSelectedComment(comment)
}

func (g *shaderGraph) beginBoxSelectionFromInput() {
	if g == nil || g.host == nil || g.host.Window == nil || g.isPanInputHeld() {
		return
	}
	mousePosition := g.host.Window.Mouse.ScreenPosition()
	if g.inputBlocked != nil && g.inputBlocked(mousePosition) {
		return
	}
	g.boxSelecting = true
	g.boxStart = g.graphPositionFromView(g.screenToViewPosition(mousePosition))
	g.updateSelectionBoxVisual(g.boxStart)
}

func (g *shaderGraph) SelectNodes(nodes []*shaderGraphNode, mode shaderGraphSelectionMode) {
	defer tracing.NewRegion("shaderGraph.SelectNodes").End()
	if g == nil {
		return
	}
	filtered := make([]*shaderGraphNode, 0, len(nodes))
	for i := range nodes {
		node := nodes[i]
		if node == nil || !slices.Contains(g.nodes, node) || slices.Contains(filtered, node) {
			continue
		}
		filtered = append(filtered, node)
	}
	if len(filtered) == 0 && mode != shaderGraphSelectionReplace {
		return
	}
	from := g.selectionIDs()
	next := slices.Clone(g.selected)
	if mode == shaderGraphSelectionReplace && len(filtered) == 1 && g.IsSelected(filtered[0]) {
		next = shaderGraphSelectionMoveToTop(next, filtered[0])
		g.setSelectionNodes(next)
		to := g.selectionIDs()
		if slices.Equal(from, to) {
			return
		}
		if g.history != nil {
			g.history.Add(&shaderGraphSelectionHistory{
				graph: g,
				from:  from,
				to:    to,
			})
		}
		return
	}
	switch mode {
	case shaderGraphSelectionAppend:
		for i := range filtered {
			if index := slices.Index(next, filtered[i]); index >= 0 {
				next = slices.Delete(next, index, index+1)
				next = append(next, filtered[i])
			} else {
				next = append(next, filtered[i])
			}
		}
	case shaderGraphSelectionToggle:
		for i := range filtered {
			if index := slices.Index(next, filtered[i]); index >= 0 {
				next = slices.Delete(next, index, index+1)
			} else {
				next = append(next, filtered[i])
			}
		}
	case shaderGraphSelectionSubtract:
		for i := range filtered {
			if index := slices.Index(next, filtered[i]); index >= 0 {
				next = slices.Delete(next, index, index+1)
			}
		}
	default:
		next = filtered
	}
	g.setSelectionNodes(next)
	to := g.selectionIDs()
	if slices.Equal(from, to) {
		return
	}
	if g.history != nil {
		g.history.Add(&shaderGraphSelectionHistory{
			graph: g,
			from:  from,
			to:    to,
		})
	}
}

func (g *shaderGraph) setSelectionNodes(nodes []*shaderGraphNode) {
	if g == nil {
		return
	}
	if g.selectedComment != nil {
		g.setSelectedComment(nil)
	}
	previous := slices.Clone(g.selected)
	g.selected = klib.WipeSlice(g.selected)
	for i := range nodes {
		node := nodes[i]
		if node == nil || !slices.Contains(g.nodes, node) || slices.Contains(g.selected, node) {
			continue
		}
		g.selected = append(g.selected, node)
	}
	for i := range previous {
		if !slices.Contains(g.selected, previous[i]) {
			if slices.Contains(g.nodes, previous[i]) {
				g.assignNodeZSlot(previous[i])
			}
			previous[i].setSelected(false)
		}
	}
	for i := range g.selected {
		g.releaseNodeZSlot(g.selected[i])
		g.selected[i].setSelected(true)
	}
	g.applySelectionZOrder()
}

func (g *shaderGraph) setSelectedComment(comment *shaderGraphComment) {
	if g == nil || g.selectedComment == comment {
		return
	}
	if g.selectedComment != nil {
		g.selectedComment.setSelected(false)
	}
	g.selectedComment = nil
	if comment == nil || !slices.Contains(g.comments, comment) {
		return
	}
	g.selectedComment = comment
	g.selectedComment.setSelected(true)
}

func (g *shaderGraph) setSelectionIDs(ids []string) {
	if g == nil {
		return
	}
	nodes := make([]*shaderGraphNode, 0, len(ids))
	for i := range ids {
		if node := g.nodeByID(ids[i]); node != nil {
			nodes = append(nodes, node)
		}
	}
	g.setSelectionNodes(nodes)
}

func (g *shaderGraph) selectionIDs() []string {
	if g == nil || len(g.selected) == 0 {
		return nil
	}
	ids := make([]string, 0, len(g.selected))
	for i := range g.selected {
		if g.selected[i] != nil && g.selected[i].id != "" {
			ids = append(ids, g.selected[i].id)
		}
	}
	return ids
}

func (g *shaderGraph) nodeByID(id string) *shaderGraphNode {
	if g == nil || id == "" {
		return nil
	}
	for i := range g.nodes {
		if g.nodes[i] != nil && g.nodes[i].id == id {
			return g.nodes[i]
		}
	}
	return nil
}

func (g *shaderGraph) commentByID(id string) *shaderGraphComment {
	if g == nil || id == "" {
		return nil
	}
	for i := range g.comments {
		if g.comments[i] != nil && g.comments[i].id == id {
			return g.comments[i]
		}
	}
	return nil
}

func (n *shaderGraphNode) bindSelectionEvent(target *ui.UI) {
	if n == nil || target == nil {
		return
	}
	target.AddEvent(ui.EventTypeDown, func() {
		if n.graph != nil && n.graph.isAltInputHeld() {
			return
		}
		if n.graph != nil {
			n.graph.SelectNodeFromInput(n)
		}
	})
}

func shaderGraphSelectionMoveToTop(nodes []*shaderGraphNode, node *shaderGraphNode) []*shaderGraphNode {
	if node == nil {
		return nodes
	}
	if index := slices.Index(nodes, node); index >= 0 {
		nodes = slices.Delete(nodes, index, index+1)
	}
	return append(nodes, node)
}

func shaderGraphSelectionModeFromKeyboard(kb *hid.Keyboard) shaderGraphSelectionMode {
	if kb == nil {
		return shaderGraphSelectionReplace
	}
	if kb.HasShift() {
		return shaderGraphSelectionAppend
	}
	if kb.HasCtrlOrMeta() {
		return shaderGraphSelectionToggle
	}
	return shaderGraphSelectionReplace
}

func shaderGraphBoxSelectionModeFromKeyboard(kb *hid.Keyboard) shaderGraphSelectionMode {
	if kb == nil {
		return shaderGraphSelectionReplace
	}
	if kb.HasShift() {
		return shaderGraphSelectionAppend
	}
	if kb.HasCtrlOrMeta() {
		return shaderGraphSelectionSubtract
	}
	return shaderGraphSelectionReplace
}
