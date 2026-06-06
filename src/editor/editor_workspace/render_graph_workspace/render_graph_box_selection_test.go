package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestRenderGraphNodesTouchedByBoxIncludesIntersectingNodes(t *testing.T) {
	graph := renderGraph{}
	a := &renderGraphNode{id: "a", position: matrix.NewVec2(10, 10), height: 80}
	b := &renderGraphNode{id: "b", position: matrix.NewVec2(260, 10), height: 80}
	graph.nodes = []*renderGraphNode{a, b}

	touched := graph.nodesTouchedByBox(matrix.NewVec4(0, 0, 20, 20))

	if len(touched) != 1 || touched[0] != a {
		t.Fatalf("touched nodes = %v, want only a", touched)
	}
}

func TestRenderGraphNodesTouchedByBoxIncludesEdgeTouches(t *testing.T) {
	graph := renderGraph{}
	node := &renderGraphNode{id: "node", position: matrix.NewVec2(10, 10), height: 80}
	graph.nodes = []*renderGraphNode{node}

	touched := graph.nodesTouchedByBox(matrix.NewVec4(220, 90, 240, 120))

	if len(touched) != 1 || touched[0] != node {
		t.Fatalf("edge-touching box should include node")
	}
}

func TestRenderGraphBoxSelectionModes(t *testing.T) {
	graph := renderGraph{}
	a := &renderGraphNode{id: "a"}
	b := &renderGraphNode{id: "b"}
	c := &renderGraphNode{id: "c"}
	graph.nodes = []*renderGraphNode{a, b, c}

	graph.SelectNodes([]*renderGraphNode{a}, renderGraphSelectionReplace)
	graph.SelectNodes([]*renderGraphNode{b}, renderGraphSelectionAppend)
	if !graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("append box selection should add touched nodes")
	}

	graph.SelectNodes([]*renderGraphNode{a}, renderGraphSelectionSubtract)
	if graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("subtract box selection should remove touched nodes")
	}

	graph.SelectNodes([]*renderGraphNode{c}, renderGraphSelectionReplace)
	if graph.IsSelected(a) || graph.IsSelected(b) || !graph.IsSelected(c) {
		t.Fatalf("replace box selection should select only touched nodes")
	}
}
