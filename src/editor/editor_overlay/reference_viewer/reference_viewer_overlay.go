package reference_viewer

import (
	"kaiju/editor/project"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type ReferenceViewer struct {
	doc   *document.Document
	uiMan ui.Manager
}

func Show(host *engine.Host, references []project.ContentReference) (*ReferenceViewer, error) {
	defer tracing.NewRegion("reference_viewer.Show").End()
	o := &ReferenceViewer{}
	o.uiMan.Init(host)
	var err error
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/reference_viewer.go.html",
		references, map[string]func(*document.Element){
			"clickMiss": o.clickMiss,
		})
	if err != nil {
		return o, err
	}
	return o, nil
}

func (o *ReferenceViewer) Close() {
	defer tracing.NewRegion("ReferenceViewer.Close").End()
	o.doc.Destroy()
}

func (o *ReferenceViewer) clickMiss(*document.Element) {
	defer tracing.NewRegion("ReferenceViewer.clickMiss").End()
	o.Close()
}
