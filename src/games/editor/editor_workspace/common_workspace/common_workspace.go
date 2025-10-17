package common_workspace

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
)

type CommonWorkspace struct {
	Host      *engine.Host
	Doc       *document.Document
	UiMan     ui.Manager
	IsBlurred bool
}

func (w *CommonWorkspace) InitializeWithUI(host *engine.Host, htmlPath string, withData any, funcMap map[string]func(*document.Element)) error {
	w.Host = host
	w.UiMan.Init(host)
	var err error
	w.Doc, err = markup.DocumentFromHTMLAsset(&w.UiMan, htmlPath, withData, funcMap)
	w.Doc.Deactivate()
	return err
}

func (w *CommonWorkspace) CommonOpen() {
	w.Doc.Activate()
	w.UiMan.EnableUpdate()
}

func (w *CommonWorkspace) CommonClose() {
	w.UiMan.DisableUpdate()
	w.Doc.Deactivate()
}

func (w *CommonWorkspace) Focus() {
	w.UiMan.EnableUpdate()
	w.IsBlurred = false
}

func (w *CommonWorkspace) Blur() {
	w.UiMan.DisableUpdate()
	w.IsBlurred = true
}
