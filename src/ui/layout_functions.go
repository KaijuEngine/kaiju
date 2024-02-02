package ui

type LayoutFuncId = int64

type layoutFuncEntry struct {
	id   LayoutFuncId
	call func(layout *Layout)
}

type LayoutFunctions struct {
	nextId LayoutFuncId
	calls  []layoutFuncEntry
}

func NewEvent() LayoutFunctions {
	return LayoutFunctions{
		nextId: 1,
		calls:  make([]layoutFuncEntry, 0),
	}
}

func (lf *LayoutFunctions) Clear()        { lf.calls = lf.calls[:0] }
func (lf *LayoutFunctions) IsEmpty() bool { return len(lf.calls) == 0 }

func (lf *LayoutFunctions) Add(call func(layout *Layout)) LayoutFuncId {
	id := lf.nextId
	lf.nextId++
	lf.calls = append(lf.calls, layoutFuncEntry{id, call})
	return id
}

func (e *LayoutFunctions) Remove(id LayoutFuncId) {
	for i := range e.calls {
		if e.calls[i].id == id {
			last := len(e.calls) - 1
			e.calls[i], e.calls[last] = e.calls[last], e.calls[i]
			e.calls = e.calls[:last]
			return
		}
	}
}

func (lf *LayoutFunctions) Execute(layout *Layout) {
	for i := range lf.calls {
		lf.calls[i].call(layout)
	}
}
