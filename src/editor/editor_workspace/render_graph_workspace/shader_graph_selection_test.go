package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/windowing"
)

func TestShaderGraphSelectNodesReplaceAppendToggle(t *testing.T) {
	graph := shaderGraph{}
	a := &shaderGraphNode{id: "a"}
	b := &shaderGraphNode{id: "b"}
	graph.nodes = []*shaderGraphNode{a, b}

	graph.SelectNodes([]*shaderGraphNode{a}, shaderGraphSelectionReplace)
	if !graph.IsSelected(a) || graph.IsSelected(b) {
		t.Fatalf("replace selection should select only a")
	}

	graph.SelectNodes([]*shaderGraphNode{b}, shaderGraphSelectionAppend)
	if !graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("append selection should keep a and add b")
	}

	graph.SelectNodes([]*shaderGraphNode{a}, shaderGraphSelectionToggle)
	if graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("toggle selection should remove a and keep b")
	}
}

func TestShaderGraphSelectNodesEmptyReplaceClearsSelection(t *testing.T) {
	graph := shaderGraph{}
	node := &shaderGraphNode{id: "node"}
	graph.nodes = []*shaderGraphNode{node}
	graph.SelectNodes([]*shaderGraphNode{node}, shaderGraphSelectionReplace)

	graph.SelectNodes(nil, shaderGraphSelectionReplace)

	if graph.HasSelection() {
		t.Fatalf("empty replace selection should clear selection")
	}
}

func TestShaderGraphSelectNodesReplaceAlreadySelectedPreservesSelection(t *testing.T) {
	graph := shaderGraph{}
	a := &shaderGraphNode{id: "a"}
	b := &shaderGraphNode{id: "b"}
	graph.nodes = []*shaderGraphNode{a, b}
	graph.SelectNodes([]*shaderGraphNode{a}, shaderGraphSelectionReplace)
	graph.SelectNodes([]*shaderGraphNode{b}, shaderGraphSelectionAppend)

	graph.SelectNodes([]*shaderGraphNode{a}, shaderGraphSelectionReplace)

	if !graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("replace on an already selected node should preserve selection")
	}
}

func TestShaderGraphSelectionHistoryUndoRedo(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph := shaderGraph{history: history}
	a := &shaderGraphNode{id: "a"}
	b := &shaderGraphNode{id: "b"}
	graph.nodes = []*shaderGraphNode{a, b}

	graph.SelectNodes([]*shaderGraphNode{a}, shaderGraphSelectionReplace)
	graph.SelectNodes([]*shaderGraphNode{b}, shaderGraphSelectionReplace)

	history.Undo()
	if !graph.IsSelected(a) || graph.IsSelected(b) {
		t.Fatalf("undo should restore previous selection")
	}

	history.Redo()
	if graph.IsSelected(a) || !graph.IsSelected(b) {
		t.Fatalf("redo should restore next selection")
	}
}

func TestShaderGraphSelectionEventSkipsAltInput(t *testing.T) {
	keyboard := hid.NewKeyboard()
	keyboard.SetKeyDown(hid.KeyboardKeyLeftAlt)
	graph := shaderGraph{
		host: &engine.Host{
			Window: &windowing.Window{Keyboard: keyboard},
		},
	}
	node := &shaderGraphNode{id: "node", graph: &graph}
	graph.nodes = []*shaderGraphNode{node}
	target := &ui.UI{}
	node.bindSelectionEvent(target)

	target.ExecuteEvent(ui.EventTypeDown)

	if graph.IsSelected(node) {
		t.Fatal("alt down should not select the node")
	}
}

func TestShaderGraphBeginBoxSelectionRespectsInputBlocker(t *testing.T) {
	mouse := hid.NewMouse()
	mouse.SetPosition(12, 24, 100, 100)
	graph := shaderGraph{
		host: &engine.Host{
			Window: &windowing.Window{Mouse: mouse},
		},
		inputBlocked: func(matrix.Vec2) bool { return true },
	}

	graph.beginBoxSelectionFromInput()

	if graph.boxSelecting {
		t.Fatal("blocked graph input should not start box selection")
	}
}

func TestShaderGraphSelectedNodeDragMovesSelectionWithHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph := shaderGraph{history: history}
	a := &shaderGraphNode{id: "a", graph: &graph, position: matrix.NewVec2(4, 6)}
	b := &shaderGraphNode{id: "b", graph: &graph, position: matrix.NewVec2(12, 18)}
	graph.nodes = []*shaderGraphNode{a, b}
	graph.setSelectionNodes([]*shaderGraphNode{a, b})

	a.captureDragNodes()
	a.applyDragDelta(matrix.NewVec2(10, -5))
	a.addDragHistory()

	if got := a.position; !matrix.Vec2Approx(got, matrix.NewVec2(14, 1)) {
		t.Fatalf("dragged node position = %v, want [14 1]", got)
	}
	if got := b.position; !matrix.Vec2Approx(got, matrix.NewVec2(22, 13)) {
		t.Fatalf("other selected node position = %v, want [22 13]", got)
	}

	history.Undo()
	if !matrix.Vec2Approx(a.position, matrix.NewVec2(4, 6)) ||
		!matrix.Vec2Approx(b.position, matrix.NewVec2(12, 18)) {
		t.Fatalf("undo should restore selected node positions")
	}

	history.Redo()
	if !matrix.Vec2Approx(a.position, matrix.NewVec2(14, 1)) ||
		!matrix.Vec2Approx(b.position, matrix.NewVec2(22, 13)) {
		t.Fatalf("redo should restore moved selected node positions")
	}
}
