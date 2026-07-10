/******************************************************************************/
/* render_graph_z_order.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"slices"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
)

const (
	renderGraphNodeZBase = matrix.Float(5)
	renderGraphNodeZStep = matrix.Float(0.5)
)

func (g *renderGraph) assignNodeZSlot(node *renderGraphNode) {
	if g == nil || node == nil || node.zSlotAssigned {
		return
	}
	slot := 0
	if len(g.availableNodeZSlots) > 0 {
		slices.Sort(g.availableNodeZSlots)
		slot = g.availableNodeZSlots[0]
		g.availableNodeZSlots = slices.Delete(g.availableNodeZSlots, 0, 1)
	} else {
		for {
			inUse := false
			for i := range g.nodes {
				if g.nodes[i] != nil && g.nodes[i].zSlotAssigned && g.nodes[i].zSlot == slot {
					inUse = true
					break
				}
			}
			if !inUse {
				break
			}
			slot++
		}
	}
	node.zSlot = slot
	node.zSlotAssigned = true
	node.setZDepth(g.nodeZDepthForSlot(slot))
}

func (g *renderGraph) releaseNodeZSlot(node *renderGraphNode) {
	if g == nil || node == nil || !node.zSlotAssigned {
		return
	}
	slot := node.zSlot
	node.zSlotAssigned = false
	node.zSlot = 0
	if !slices.Contains(g.availableNodeZSlots, slot) {
		g.availableNodeZSlots = append(g.availableNodeZSlots, slot)
	}
}

func (g *renderGraph) applySelectionZOrder() {
	if g == nil {
		return
	}
	baseSlot := 0
	for i := range g.nodes {
		node := g.nodes[i]
		if node != nil && node.zSlotAssigned {
			baseSlot = max(baseSlot, node.zSlot+1)
		}
	}
	for i := range g.selected {
		node := g.selected[i]
		if node == nil {
			continue
		}
		node.setZDepth(g.nodeZDepthForSlot(baseSlot + i))
	}
}

func (g *renderGraph) nodeZDepthForSlot(slot int) matrix.Float {
	if slot < 0 {
		slot = 0
	}
	return renderGraphNodeZBase + matrix.Float(slot)*renderGraphNodeZStep
}

func (n *renderGraphNode) setZDepth(z matrix.Float) {
	if n == nil || matrix.Approx(n.zDepth, z) {
		return
	}
	n.zDepth = z
	if n.root == nil {
		return
	}
	layout := n.root.Base().Layout()
	layout.SetZ(z)
	n.root.Base().SetDirty(ui.DirtyTypeLayout)
}
