package memento

type HistoryTransaction struct {
	stack []Memento
}

func (h *HistoryTransaction) Undo() {
	for i := len(h.stack) - 1; i >= 0; i-- {
		h.stack[i].Undo()
	}
}

func (h *HistoryTransaction) Redo() {
	for i := range h.stack {
		h.stack[i].Redo()
	}
}

func (h *HistoryTransaction) Delete() {
	for i := len(h.stack) - 1; i >= 0; i-- {
		h.stack[i].Delete()
	}
}

func (h *HistoryTransaction) Exit() {
	for i := range h.stack {
		h.stack[i].Exit()
	}
}
