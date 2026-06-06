package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/editor/memento"
	"kaijuengine.com/matrix"
)

func TestRenderGraphCreateCommentFromActionCreatesUndoableComment(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	workspace := &RenderGraphWorkspace{}
	workspace.graph.history = history

	history.BeginTransaction()
	comment, ok := workspace.CreateCommentFromAction(CreateCommentActionArgs{
		Label:       "Lighting",
		X:           20,
		Y:           30,
		Width:       260,
		Height:      140,
		UsePosition: true,
		UseSize:     true,
	})
	history.CommitTransaction()

	if !ok || comment == nil {
		t.Fatal("CreateCommentFromAction() failed")
	}
	if comment.label != "Lighting" {
		t.Fatalf("comment label = %q, want Lighting", comment.label)
	}
	if !matrix.Vec2Approx(comment.position, matrix.NewVec2(20, 30)) {
		t.Fatalf("comment position = %v, want [20 30]", comment.position)
	}
	if !matrix.Vec2Approx(comment.size, matrix.NewVec2(260, 140)) {
		t.Fatalf("comment size = %v, want [260 140]", comment.size)
	}
	if workspace.graph.selectedComment != comment {
		t.Fatal("created comment should be selected")
	}

	history.Undo()
	if got := workspace.graph.commentByID(comment.id); got != nil {
		t.Fatalf("created comment still exists after undo: %#v", got)
	}

	history.Redo()
	created := workspace.graph.commentByID(comment.id)
	if created == nil {
		t.Fatal("created comment was not restored by redo")
	}
	if workspace.graph.selectedComment != created {
		t.Fatal("redo should select the recreated comment")
	}
}

func TestShaderGraphDeleteSelectedCommentAddsUndoableHistory(t *testing.T) {
	history := &memento.History{}
	history.Initialize(8)
	graph := shaderGraph{history: history}
	comment := graph.createCommentFromSnapshot(RenderGraphComment{
		ID:       "comment-a",
		Label:    "Group",
		Position: matrix.NewVec2(2, 3),
		Size:     matrix.NewVec2(240, 120),
	})
	graph.setSelectedComment(comment)

	if !graph.DeleteSelectedNodes() {
		t.Fatal("DeleteSelectedNodes() should delete a selected comment")
	}
	if got := graph.commentByID("comment-a"); got != nil {
		t.Fatalf("deleted comment still exists: %#v", got)
	}

	history.Undo()
	restored := graph.commentByID("comment-a")
	if restored == nil {
		t.Fatal("undo should restore deleted comment")
	}
	if restored.label != "Group" || !matrix.Vec2Approx(restored.size, matrix.NewVec2(240, 120)) {
		t.Fatalf("restored comment = %#v", restored)
	}
}

func TestShaderGraphCommentSizeClampsToMinimum(t *testing.T) {
	got := shaderGraphCommentSizeOrDefault(matrix.NewVec2(1, 2))

	if got.X() != shaderGraphCommentMinWidth || got.Y() != shaderGraphCommentMinHeight {
		t.Fatalf("clamped size = %v, want min %v x %v", got, shaderGraphCommentMinWidth, shaderGraphCommentMinHeight)
	}
}
