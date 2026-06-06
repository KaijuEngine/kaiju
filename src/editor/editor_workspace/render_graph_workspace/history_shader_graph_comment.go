/******************************************************************************/
/* history_render_graph_comment.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type shaderGraphCommentCreateHistory struct {
	graph             *shaderGraph
	comment           RenderGraphComment
	previousSelection []string
}

func (h *shaderGraphCommentCreateHistory) Redo() {
	defer tracing.NewRegion("shaderGraphCommentCreateHistory.Redo").End()
	if h.graph == nil {
		return
	}
	comment := h.graph.createCommentFromSnapshot(h.comment)
	if comment != nil {
		h.graph.setSelectionNodes(nil)
		h.graph.setSelectedComment(comment)
	}
}

func (h *shaderGraphCommentCreateHistory) Undo() {
	defer tracing.NewRegion("shaderGraphCommentCreateHistory.Undo").End()
	if h.graph == nil {
		return
	}
	h.graph.RemoveComment(h.comment.ID)
	h.graph.setSelectionIDs(h.previousSelection)
}

func (h *shaderGraphCommentCreateHistory) Delete() {}
func (h *shaderGraphCommentCreateHistory) Exit()   {}

type shaderGraphCommentDeleteHistory struct {
	graph   *shaderGraph
	comment RenderGraphComment
}

func (h *shaderGraphCommentDeleteHistory) Redo() {
	defer tracing.NewRegion("shaderGraphCommentDeleteHistory.Redo").End()
	if h.graph != nil {
		h.graph.RemoveComment(h.comment.ID)
	}
}

func (h *shaderGraphCommentDeleteHistory) Undo() {
	defer tracing.NewRegion("shaderGraphCommentDeleteHistory.Undo").End()
	if h.graph == nil {
		return
	}
	comment := h.graph.createCommentFromSnapshot(h.comment)
	if comment != nil {
		h.graph.setSelectionNodes(nil)
		h.graph.setSelectedComment(comment)
	}
}

func (h *shaderGraphCommentDeleteHistory) Delete() {}
func (h *shaderGraphCommentDeleteHistory) Exit()   {}

type shaderGraphCommentPositionHistory struct {
	graph *shaderGraph
	id    string
	from  matrix.Vec2
	to    matrix.Vec2
}

func (h *shaderGraphCommentPositionHistory) Redo() {
	defer tracing.NewRegion("shaderGraphCommentPositionHistory.Redo").End()
	h.apply(h.to)
}

func (h *shaderGraphCommentPositionHistory) Undo() {
	defer tracing.NewRegion("shaderGraphCommentPositionHistory.Undo").End()
	h.apply(h.from)
}

func (h *shaderGraphCommentPositionHistory) apply(position matrix.Vec2) {
	if h.graph == nil {
		return
	}
	comment := h.graph.commentByID(h.id)
	if comment == nil {
		return
	}
	comment.position = position
	comment.applyViewOffset()
}

func (h *shaderGraphCommentPositionHistory) Delete() {}
func (h *shaderGraphCommentPositionHistory) Exit()   {}

type shaderGraphCommentSizeHistory struct {
	graph *shaderGraph
	id    string
	from  matrix.Vec2
	to    matrix.Vec2
}

func (h *shaderGraphCommentSizeHistory) Redo() {
	defer tracing.NewRegion("shaderGraphCommentSizeHistory.Redo").End()
	h.apply(h.to)
}

func (h *shaderGraphCommentSizeHistory) Undo() {
	defer tracing.NewRegion("shaderGraphCommentSizeHistory.Undo").End()
	h.apply(h.from)
}

func (h *shaderGraphCommentSizeHistory) apply(size matrix.Vec2) {
	if h.graph == nil {
		return
	}
	comment := h.graph.commentByID(h.id)
	if comment == nil {
		return
	}
	comment.setSize(size)
}

func (h *shaderGraphCommentSizeHistory) Delete() {}
func (h *shaderGraphCommentSizeHistory) Exit()   {}
