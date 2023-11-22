package windowing

func New(windowName string) {
	var evtSharedMem evtMem
	createWindow(windowName, &evtSharedMem)
	for !evtSharedMem.IsQuit() && !evtSharedMem.IsFatal() {
		for !evtSharedMem.IsReady() {
		}
		if evtSharedMem.IsWritten() {
			// TODO:  Read event here
			evtSharedMem.MakeAvailable()
		}
	}
}
