package memento

type History struct {
	undoStack []Memento
	position  int
}

func NewHistory() History {
	return History{
		undoStack: make([]Memento, 0),
	}
}

func (h *History) Add(m Memento) {
	for i := h.position; i < len(h.undoStack); i++ {
		h.undoStack[i].Delete()
	}
	h.undoStack = h.undoStack[:h.position]
	h.undoStack = append(h.undoStack, m)
	h.position++
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
	m.Do()
	h.position++
}
