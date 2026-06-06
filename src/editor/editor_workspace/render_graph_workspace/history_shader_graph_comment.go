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

type renderGraphCommentCreateHistory struct {
	graph             *renderGraph
	comment           RenderGraphComment
	previousSelection []string
}

func (h *renderGraphCommentCreateHistory) Redo() {
	defer tracing.NewRegion("renderGraphCommentCreateHistory.Redo").End()
	if h.graph == nil {
		return
	}
	comment := h.graph.createCommentFromSnapshot(h.comment)
	if comment != nil {
		h.graph.setSelectionNodes(nil)
		h.graph.setSelectedComment(comment)
	}
}

func (h *renderGraphCommentCreateHistory) Undo() {
	defer tracing.NewRegion("renderGraphCommentCreateHistory.Undo").End()
	if h.graph == nil {
		return
	}
	h.graph.RemoveComment(h.comment.ID)
	h.graph.setSelectionIDs(h.previousSelection)
}

func (h *renderGraphCommentCreateHistory) Delete() {}
func (h *renderGraphCommentCreateHistory) Exit()   {}

type renderGraphCommentDeleteHistory struct {
	graph   *renderGraph
	comment RenderGraphComment
}

func (h *renderGraphCommentDeleteHistory) Redo() {
	defer tracing.NewRegion("renderGraphCommentDeleteHistory.Redo").End()
	if h.graph != nil {
		h.graph.RemoveComment(h.comment.ID)
	}
}

func (h *renderGraphCommentDeleteHistory) Undo() {
	defer tracing.NewRegion("renderGraphCommentDeleteHistory.Undo").End()
	if h.graph == nil {
		return
	}
	comment := h.graph.createCommentFromSnapshot(h.comment)
	if comment != nil {
		h.graph.setSelectionNodes(nil)
		h.graph.setSelectedComment(comment)
	}
}

func (h *renderGraphCommentDeleteHistory) Delete() {}
func (h *renderGraphCommentDeleteHistory) Exit()   {}

type renderGraphCommentPositionHistory struct {
	graph *renderGraph
	id    string
	from  matrix.Vec2
	to    matrix.Vec2
}

func (h *renderGraphCommentPositionHistory) Redo() {
	defer tracing.NewRegion("renderGraphCommentPositionHistory.Redo").End()
	h.apply(h.to)
}

func (h *renderGraphCommentPositionHistory) Undo() {
	defer tracing.NewRegion("renderGraphCommentPositionHistory.Undo").End()
	h.apply(h.from)
}

func (h *renderGraphCommentPositionHistory) apply(position matrix.Vec2) {
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

func (h *renderGraphCommentPositionHistory) Delete() {}
func (h *renderGraphCommentPositionHistory) Exit()   {}

type renderGraphCommentSizeHistory struct {
	graph *renderGraph
	id    string
	from  matrix.Vec2
	to    matrix.Vec2
}

func (h *renderGraphCommentSizeHistory) Redo() {
	defer tracing.NewRegion("renderGraphCommentSizeHistory.Redo").End()
	h.apply(h.to)
}

func (h *renderGraphCommentSizeHistory) Undo() {
	defer tracing.NewRegion("renderGraphCommentSizeHistory.Undo").End()
	h.apply(h.from)
}

func (h *renderGraphCommentSizeHistory) apply(size matrix.Vec2) {
	if h.graph == nil {
		return
	}
	comment := h.graph.commentByID(h.id)
	if comment == nil {
		return
	}
	comment.setSize(size)
}

func (h *renderGraphCommentSizeHistory) Delete() {}
func (h *renderGraphCommentSizeHistory) Exit()   {}
