/******************************************************************************/
/* schema_graph_layout.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import "kaijuengine.com/matrix"

const (
	schemaGraphPaddingX    = float32(32.0)
	schemaGraphPaddingY    = float32(32.0)
	schemaGraphColumnGap   = float32(80.0)
	schemaGraphVerticalGap = float32(18.0)
)

func (g *schemaGraph) reflow() {
	if g == nil {
		return
	}
	y := schemaGraphPaddingY
	for i := range g.rootNodes {
		node := g.rootNodes[i]
		if node == nil {
			continue
		}
		height := schemaSubtreeHeight(node)
		schemaLayoutSubtree(node, 0, y, height)
		y += height + schemaGraphVerticalGap
	}
	g.layoutDirty = false
}

func schemaSubtreeHeight(node *schemaNode) float32 {
	if node == nil {
		return 0
	}
	if len(node.children) == 0 {
		return node.height
	}
	childrenHeight := float32(0)
	for i := range node.children {
		childHeight := schemaSubtreeHeight(node.children[i])
		if childHeight <= 0 {
			continue
		}
		if childrenHeight > 0 {
			childrenHeight += schemaGraphVerticalGap
		}
		childrenHeight += childHeight
	}
	return max(node.height, childrenHeight)
}

func schemaLayoutSubtree(node *schemaNode, depth int, top, subtreeHeight float32) {
	if node == nil {
		return
	}
	x := schemaGraphPaddingX + float32(depth)*(schemaNodeWidth+schemaGraphColumnGap)
	y := top + max(0, (subtreeHeight-node.height)*0.5)
	node.SetPosition(matrix.NewVec2(matrix.Float(x), matrix.Float(y)))
	if len(node.children) == 0 {
		return
	}
	childY := top
	for i := range node.children {
		child := node.children[i]
		childHeight := schemaSubtreeHeight(child)
		schemaLayoutSubtree(child, depth+1, childY, childHeight)
		childY += childHeight + schemaGraphVerticalGap
	}
}
