package hierarchy

import (
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/ui"
	"kaiju/uimarkup"
	"kaiju/uimarkup/markup"
	"strings"
)

type Hierarchy struct {
	doc        *markup.Document
	input      *ui.Input
	onChangeId engine.EventId
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
	searchInputElement, _ := h.doc.GetElementById("hierarchyInput")
	entityList, _ := h.doc.GetElementById("entityList")
	h.input = searchInputElement.UI.(*ui.Input)

	h.input.Data().OnChange.Remove(h.onChangeId)
	h.onChangeId = h.input.Data().OnChange.Add(func() {
		activeText := h.input.Text()

		for idx := range entityList.HTML.Children {
			label := entityList.HTML.Children[idx].Children[0].DocumentElement.UI.(*ui.Label)

			if strings.Contains(label.Text(), activeText) {
				entityList.HTML.Children[idx].DocumentElement.UI.Entity().Activate()
			} else {
				entityList.HTML.Children[idx].DocumentElement.UI.Entity().Deactivate()
			}
		}
	})
}
