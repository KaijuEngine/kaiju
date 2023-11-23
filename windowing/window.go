package windowing

type Window struct {
	evtSharedMem evtMem
	isClosed     bool
	isCrashed    bool
}

func New(windowName string) *Window {
	w := &Window{}
	createWindow(windowName, &w.evtSharedMem)
	return w
}

func (w Window) IsClosed() bool {
	return w.isClosed
}

func (w Window) IsCrashed() bool {
	return w.isCrashed
}

func (w *Window) Poll() {
	w.evtSharedMem.MakeAvailable()
	for !w.evtSharedMem.IsQuit() && !w.evtSharedMem.IsFatal() {
		for !w.evtSharedMem.IsReady() {
		}
		if w.evtSharedMem.IsWritten() {
			if w.evtSharedMem.HasEvent() {
				// TODO:  Read event here
				w.evtSharedMem.MakeAvailable()
			} else {
				break
			}
		}
	}
	w.isClosed = w.isClosed || w.evtSharedMem.IsQuit()
	w.isCrashed = w.isCrashed || w.evtSharedMem.IsFatal()
}
