/******************************************************************************/
/* reference_viewer_overlay.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package reference_viewer

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

var existing *ReferenceViewer

type ReferenceViewer struct {
	doc           *document.Document
	entryTemplate *document.Element
	uiMan         ui.Manager
	OnClose       events.Event
}

func Show(host *engine.Host, p *project.Project, id string) (*ReferenceViewer, error) {
	defer tracing.NewRegion("reference_viewer.Show").End()
	// Only allow one context menu open at a time
	if existing != nil {
		existing.closeInternal(true)
	}
	referenceViewer := &ReferenceViewer{}
	referenceViewer.uiMan.Init(host)
	var err error
	referenceViewer.doc, err = markup.DocumentFromHTMLAsset(&referenceViewer.uiMan, "editor/ui/overlay/reference_viewer.go.html",
		nil, map[string]func(*document.Element){
			"clickMiss": referenceViewer.clickMiss,
		})
	if err != nil {
		return referenceViewer, err
	}
	referenceViewer.entryTemplate, _ = referenceViewer.doc.GetElementById("entryTemplate")
	referenceViewer.entryTemplate.UI.Hide()
	existing = referenceViewer
	go func() {
		notFoundInfo := referenceViewer.doc.GetElementsByClass("not-found-info")[0]
		notFoundInfo.UI.Hide()
		searchInfo := referenceViewer.doc.GetElementsByClass("search-info")[0]
		var hasReferences = false
		if err := p.FindReferencesWithCallback(id, func(ref project.ContentReference) {
			hasReferences = true
			referenceViewer.onFound(ref)
		}); err != nil {
			slog.Error("failed to find all references for content", "error", err)
		}
		if referenceViewer == existing {
			searchInfo.UI.Hide()
		}
		notFoundInfo.UI.SetVisibility(!hasReferences)
	}()

	return referenceViewer, nil
}

func (o *ReferenceViewer) onFound(newRef project.ContentReference) {
	o.uiMan.Host.RunOnMainThread(func() {
		var nest func(parent *document.Element, ref project.ContentReference, first bool)
		nest = func(parent *document.Element, ref project.ContentReference, first bool) {
			e := o.doc.DuplicateElementToParent(o.entryTemplate, parent)
			e.Children[0].InnerLabel().SetText(fmt.Sprintf("%s (%s)", ref.Name, ref.Source))
			if !first {
				o.doc.SetElementClassesWithoutApply(e, "entry", "entryChild")
			}
			for i := range ref.SubReference {
				nest(e, ref.SubReference[i], false)
			}
		}
		nest(o.entryTemplate.Parent.Value(), newRef, true)
		o.doc.ApplyStyles()
	})
}

func (o *ReferenceViewer) Close() {
	defer tracing.NewRegion("ReferenceViewer.Close").End()
	o.closeInternal(false)
}

func (o *ReferenceViewer) closeInternal(beingReplaced bool) {
	if !beingReplaced {
		o.OnClose.Execute()
	}
	o.doc.Destroy()
	existing = nil
}

func (o *ReferenceViewer) clickMiss(*document.Element) {
	defer tracing.NewRegion("ReferenceViewer.clickMiss").End()
	o.Close()
}
