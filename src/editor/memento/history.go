package memento

type History struct {
	undoStack []Memento
	position  int
	limit     int
}

func NewHistory(limit int) History {
	return History{
		undoStack: make([]Memento, 0),
		limit:     limit,
	}
}

func (h *History) Add(m Memento) {
	for i := h.position; i < len(h.undoStack); i++ {
		h.undoStack[i].Delete()
	}
	h.undoStack = h.undoStack[:h.position]
	h.undoStack = append(h.undoStack, m)
	h.position++
	if h.position > h.limit {
		h.position = h.limit
		h.undoStack[0].Exit()
		h.undoStack = h.undoStack[1:]
	}
}

func (h *History) Undo() {
	if h.position == 0 {
		return
	}
	h.position--
	m := h.undoStack[h.position]
	m.Undo()
}

func (h *History) Redo() {
	if h.position == len(h.undoStack) {
		return
	}
	m := h.undoStack[h.position]
	m.Redo()
	h.position++
}
