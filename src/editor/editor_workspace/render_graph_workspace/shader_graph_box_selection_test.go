package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestShaderGraphNodesTouchedByBoxIncludesIntersectingNodes(t *testing.T) {
	graph := shaderGraph{}
	a := &shaderGraphNode{id: "a", position: matrix.NewVec2(10, 10), height: 80}
	b := &shaderGraphNode{id: "b", position: matrix.NewVec2(260, 10), height: 80}
	graph.nodes = []*shaderGraphNode{a, b}

	touched := graph.nodesTouchedByBox(matrix.NewVec4(0, 0, 20, 20))

	if len(touched) != 1 || touched[0] != a {
		t.Fatalf("touched nodes = %v, want only a", touched)
	}
}

func TestShaderGraphNodesTouchedByBoxIncludesEdgeTouches(t *testing.T) {
	graph := shaderGraph{}
	node := &shaderGraphNode{id: "node", position: matrix.NewVec2(10, 10), height: 80}
	graph.nodes = []*shaderGraphNode{node}

	touched := graph.nodesTouchedByBox(matrix.NewVec4(220, 90, 240, 120))

	if len(touched) != 1 || touched[0] != node {
		t.Fatalf("edge-touching box should include node")
	}
}

func TestShaderGraphBoxSelectionModes(t *testing.T) {
	graph := shaderGraph{}
	a := &shaderGraphNode{id: "a"}
	b := &shaderGraphNode{id: "b"}
	c := &shaderGraphNode{id: "c"}
	graph.nodes = []*shaderGraphNode{a, b, c}

	graph.SelectNodes([]*shaderGraphNode{a}, shaderGraphSelectionReplace)
	graph.SelectNodes([]*shaderGraphNode{b}, shaderGraphSelectionAppend)
	if !graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("append box selection should add touched nodes")
	}

	graph.SelectNodes([]*shaderGraphNode{a}, shaderGraphSelectionSubtract)
	if graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("subtract box selection should remove touched nodes")
	}

	graph.SelectNodes([]*shaderGraphNode{c}, shaderGraphSelectionReplace)
	if graph.IsSelected(a) || graph.IsSelected(b) || !graph.IsSelected(c) {
		t.Fatalf("replace box selection should select only touched nodes")
	}
}
