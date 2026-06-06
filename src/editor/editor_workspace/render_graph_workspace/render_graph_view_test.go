package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestRenderGraphViewPositionUsesPan(t *testing.T) {
	graph := renderGraph{pan: matrix.NewVec2(24, -16)}
	position := matrix.NewVec2(100, 80)
	want := matrix.NewVec2(124, 64)

	if got := graph.viewPosition(position); !matrix.Vec2Approx(got, want) {
		t.Fatalf("viewPosition() = %v, want %v", got, want)
	}
}

func TestRenderGraphGraphPositionFromViewRemovesPan(t *testing.T) {
	graph := renderGraph{pan: matrix.NewVec2(24, -16)}
	position := matrix.NewVec2(124, 64)
	want := matrix.NewVec2(100, 80)

	if got := graph.graphPositionFromView(position); !matrix.Vec2Approx(got, want) {
		t.Fatalf("graphPositionFromView() = %v, want %v", got, want)
	}
}

func TestRenderGraphViewPositionUsesZoom(t *testing.T) {
	graph := renderGraph{
		pan:  matrix.NewVec2(24, -16),
		zoom: 0.5,
	}
	position := matrix.NewVec2(100, 80)
	want := matrix.NewVec2(74, 24)

	if got := graph.viewPosition(position); !matrix.Vec2Approx(got, want) {
		t.Fatalf("viewPosition() = %v, want %v", got, want)
	}
}

func TestRenderGraphGraphPositionFromViewRemovesZoom(t *testing.T) {
	graph := renderGraph{
		pan:  matrix.NewVec2(24, -16),
		zoom: 0.5,
	}
	position := matrix.NewVec2(74, 24)
	want := matrix.NewVec2(100, 80)

	if got := graph.graphPositionFromView(position); !matrix.Vec2Approx(got, want) {
		t.Fatalf("graphPositionFromView() = %v, want %v", got, want)
	}
}

func TestRenderGraphSetZoomAroundViewPositionKeepsAnchorStable(t *testing.T) {
	graph := renderGraph{
		pan:  matrix.NewVec2(24, -16),
		zoom: 0.5,
	}
	anchor := matrix.NewVec2(250, 120)
	before := graph.graphPositionFromView(anchor)

	graph.setZoomAroundViewPosition(0.75, anchor)

	after := graph.graphPositionFromView(anchor)
	if !matrix.Vec2Approx(after, before) {
		t.Fatalf("anchored graph position = %v, want %v", after, before)
	}
}

func TestRenderGraphSetZoomClampsToDefaultZoom(t *testing.T) {
	graph := renderGraph{
		pan:  matrix.NewVec2(24, -16),
		zoom: 0.75,
	}

	graph.setZoomAroundViewPosition(2, matrix.NewVec2(250, 120))

	if !matrix.Approx(graph.zoom, 1) {
		t.Fatalf("zoom = %v, want default zoom", graph.zoom)
	}
}

func TestRenderGraphNodesBoundsUnionsSelectedNodes(t *testing.T) {
	a := &renderGraphNode{position: matrix.NewVec2(10, 20), height: 80}
	b := &renderGraphNode{position: matrix.NewVec2(260, 120), height: 140}

	bounds, ok := renderGraphNodesBounds([]*renderGraphNode{nil, a, b})

	if !ok {
		t.Fatal("renderGraphNodesBounds() should find bounds")
	}
	want := matrix.NewVec4(10, 20, 470, 260)
	if !matrix.Vec4Approx(bounds, want) {
		t.Fatalf("bounds = %v, want %v", bounds, want)
	}
}

func TestRenderGraphFocusBoundsCentersBoundsAtCurrentZoom(t *testing.T) {
	graph := renderGraph{zoom: 0.5}
	bounds := matrix.NewVec4(50, 100, 250, 200)

	graph.focusBounds(bounds, matrix.NewVec2(400, 300))

	center := matrix.NewVec2(150, 150)
	if got := graph.viewPosition(center); !matrix.Vec2Approx(got, matrix.NewVec2(200, 150)) {
		t.Fatalf("focused center view position = %v, want viewport center", got)
	}
}

func TestRenderGraphCenterViewResetsPanAndZoom(t *testing.T) {
	graph := renderGraph{
		pan:  matrix.NewVec2(24, -16),
		zoom: 0.5,
	}

	graph.CenterView()

	if !matrix.Vec2Approx(graph.pan, matrix.Vec2Zero()) {
		t.Fatalf("pan = %v, want zero", graph.pan)
	}
	if !matrix.Approx(graph.zoom, 1) {
		t.Fatalf("zoom = %v, want default zoom", graph.zoom)
	}
}
