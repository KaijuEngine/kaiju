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

type renderGraphSelectionMode int

const (
	renderGraphSelectionReplace renderGraphSelectionMode = iota
	renderGraphSelectionAppend
	renderGraphSelectionToggle
	renderGraphSelectionSubtract
)

func (g *renderGraph) HasSelection() bool {
	return g != nil && (len(g.selected) > 0 || g.selectedComment != nil)
}

func (g *renderGraph) Selection() []*renderGraphNode {
	return slices.Clone(g.selected)
}

func (g *renderGraph) IsSelected(node *renderGraphNode) bool {
	if g == nil || node == nil {
		return false
	}
	return slices.Contains(g.selected, node)
}

func (g *renderGraph) SelectNodeFromInput(node *renderGraphNode) {
	if g == nil || node == nil || g.host == nil || g.host.Window == nil {
		return
	}
	g.SelectNodes([]*renderGraphNode{node}, renderGraphSelectionModeFromKeyboard(&g.host.Window.Keyboard))
}

func (g *renderGraph) SelectCommentFromInput(comment *renderGraphComment) {
	if g == nil || comment == nil {
		return
	}
	g.setSelectionNodes(nil)
	g.setSelectedComment(comment)
}

func (g *renderGraph) beginBoxSelectionFromInput() {
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

func (g *renderGraph) SelectNodes(nodes []*renderGraphNode, mode renderGraphSelectionMode) {
	defer tracing.NewRegion("renderGraph.SelectNodes").End()
	if g == nil {
		return
	}
	filtered := make([]*renderGraphNode, 0, len(nodes))
	for i := range nodes {
		node := nodes[i]
		if node == nil || !slices.Contains(g.nodes, node) || slices.Contains(filtered, node) {
			continue
		}
		filtered = append(filtered, node)
	}
	if len(filtered) == 0 && mode != renderGraphSelectionReplace {
		return
	}
	from := g.selectionIDs()
	next := slices.Clone(g.selected)
	if mode == renderGraphSelectionReplace && len(filtered) == 1 && g.IsSelected(filtered[0]) {
		next = renderGraphSelectionMoveToTop(next, filtered[0])
		g.setSelectionNodes(next)
		to := g.selectionIDs()
		if slices.Equal(from, to) {
			return
		}
		if g.history != nil {
			g.history.Add(&renderGraphSelectionHistory{
				graph: g,
				from:  from,
				to:    to,
			})
		}
		return
	}
	switch mode {
	case renderGraphSelectionAppend:
		for i := range filtered {
			if index := slices.Index(next, filtered[i]); index >= 0 {
				next = slices.Delete(next, index, index+1)
				next = append(next, filtered[i])
			} else {
				next = append(next, filtered[i])
			}
		}
	case renderGraphSelectionToggle:
		for i := range filtered {
			if index := slices.Index(next, filtered[i]); index >= 0 {
				next = slices.Delete(next, index, index+1)
			} else {
				next = append(next, filtered[i])
			}
		}
	case renderGraphSelectionSubtract:
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
		g.history.Add(&renderGraphSelectionHistory{
			graph: g,
			from:  from,
			to:    to,
		})
	}
}

func (g *renderGraph) setSelectionNodes(nodes []*renderGraphNode) {
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

func (g *renderGraph) setSelectedComment(comment *renderGraphComment) {
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

func (g *renderGraph) setSelectionIDs(ids []string) {
	if g == nil {
		return
	}
	nodes := make([]*renderGraphNode, 0, len(ids))
	for i := range ids {
		if node := g.nodeByID(ids[i]); node != nil {
			nodes = append(nodes, node)
		}
	}
	g.setSelectionNodes(nodes)
}

func (g *renderGraph) selectionIDs() []string {
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

func (g *renderGraph) nodeByID(id string) *renderGraphNode {
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

func (g *renderGraph) commentByID(id string) *renderGraphComment {
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

func (n *renderGraphNode) bindSelectionEvent(target *ui.UI) {
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

func renderGraphSelectionMoveToTop(nodes []*renderGraphNode, node *renderGraphNode) []*renderGraphNode {
	if node == nil {
		return nodes
	}
	if index := slices.Index(nodes, node); index >= 0 {
		nodes = slices.Delete(nodes, index, index+1)
	}
	return append(nodes, node)
}

func renderGraphSelectionModeFromKeyboard(kb *hid.Keyboard) renderGraphSelectionMode {
	if kb == nil {
		return renderGraphSelectionReplace
	}
	if kb.HasShift() {
		return renderGraphSelectionAppend
	}
	if kb.HasCtrlOrMeta() {
		return renderGraphSelectionToggle
	}
	return renderGraphSelectionReplace
}

func renderGraphBoxSelectionModeFromKeyboard(kb *hid.Keyboard) renderGraphSelectionMode {
	if kb == nil {
		return renderGraphSelectionReplace
	}
	if kb.HasShift() {
		return renderGraphSelectionAppend
	}
	if kb.HasCtrlOrMeta() {
		return renderGraphSelectionSubtract
	}
	return renderGraphSelectionReplace
}
