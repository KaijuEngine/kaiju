package common_workspace

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
)

type CommonWorkspace struct {
	Doc   *document.Document
	UiMan ui.Manager
}

func (w *CommonWorkspace) Initialize(host *engine.Host, htmlPath string, withData any, funcMap map[string]func(*document.Element)) error {
	w.UiMan.Init(host)
	var err error
	w.Doc, err = markup.DocumentFromHTMLAsset(&w.UiMan, htmlPath, withData, funcMap)
	return err
}

func (w *CommonWorkspace) Open() {
	w.UiMan.EnableUpdate()
}

func (w *CommonWorkspace) Close() {
	w.UiMan.DisableUpdate()
}
