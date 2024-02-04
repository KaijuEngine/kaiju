package hierarchy

import (
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/uimarkup"
	"kaiju/uimarkup/markup"
)

type Hierarchy struct {
	doc *markup.Document
}

type entityEntry struct {
	Entity *engine.Entity
}

func New() *Hierarchy {
	return &Hierarchy{}
}

func (h *Hierarchy) Destroy() {
	if h.doc != nil {
		for _, elm := range h.doc.Elements {
			elm.UI.Entity().Destroy()
		}
	}
}

func (h *Hierarchy) Create(host *engine.Host) {
	allEntities := host.Entities()
	entries := make([]entityEntry, 0, len(allEntities))
	for _, entity := range allEntities {
		entries = append(entries, entityEntry{Entity: entity})
	}
	html := klib.MustReturn(host.AssetDatabase().ReadText("ui/hierarchy/hierarchy.html"))
	h.doc = uimarkup.DocumentFromHTMLString(host, html, "", entries, nil)
}
