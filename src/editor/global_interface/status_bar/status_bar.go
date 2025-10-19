package status_bar

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type StatusBar struct {
	doc   *document.Document
	uiMan ui.Manager
}

func (b *StatusBar) Initialize(host *engine.Host) error {
	defer tracing.NewRegion("TitleBar.Initialize").End()
	b.uiMan.Init(host)
	var err error
	b.doc, err = markup.DocumentFromHTMLAsset(&b.uiMan, "editor/ui/global/status_bar.go.html",
		nil, map[string]func(*document.Element){})
	return err
}

func (b *StatusBar) Focus() { b.uiMan.EnableUpdate() }
func (b *StatusBar) Blur()  { b.uiMan.DisableUpdate() }
