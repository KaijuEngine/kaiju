/******************************************************************************/
/* action_palette_overlay.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package action_palette

import (
	"log/slog"
	"strconv"
	"strings"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
)

var existing *ActionPalette

type ActionPalette struct {
	doc     *document.Document
	uiMan   ui.Manager
	service *editor_action.Service
	keyKb   hid.KeyCallbackId
	onClose func()
	input   *document.Element
	list    *document.Element
	entries []editor_action.Entry
}

type paletteData struct {
	Entries []paletteEntry
}

type paletteEntry struct {
	Index       int
	Label       string
	Description string
	Category    string
	Search      string
}

func Show(host *engine.Host, service *editor_action.Service, onClose func()) (*ActionPalette, error) {
	defer tracing.NewRegion("action_palette.Show").End()
	if existing != nil {
		existing.Close()
	}
	p := &ActionPalette{
		service: service,
		onClose: onClose,
	}
	p.uiMan.Init(host)
	p.entries = service.Search("")
	data := paletteData{Entries: make([]paletteEntry, len(p.entries))}
	for i, entry := range p.entries {
		data.Entries[i] = paletteEntry{
			Index:       i,
			Label:       entry.Label,
			Description: entry.Description,
			Category:    entry.Category,
			Search:      entrySearchText(entry),
		}
	}
	var err error
	p.doc, err = markup.DocumentFromHTMLAsset(&p.uiMan,
		"editor/ui/overlay/action_palette.go.html", data,
		map[string]func(*document.Element){
			"search":      p.search,
			"clickAction": p.clickAction,
			"clickMiss":   p.clickMiss,
		})
	if err != nil {
		return p, err
	}
	p.input, _ = p.doc.GetElementById("search")
	p.list, _ = p.doc.GetElementById("list")
	if p.input != nil {
		box := p.input.UI.ToInput()
		box.Focus()
		box.SelectAll()
	}
	p.keyKb = host.Window.Keyboard.AddKeyCallback(func(keyId int, keyState hid.KeyState) {
		if keyState != hid.KeyStateDown && keyState != hid.KeyStatePressedAndReleased {
			return
		}
		switch keyId {
		case hid.KeyboardKeyEscape:
			p.Close()
		case hid.KeyboardKeyEnter, hid.KeyboardKeyReturn:
			p.runFirstVisible()
		}
	})
	existing = p
	return p, nil
}

func (p *ActionPalette) Close() {
	defer tracing.NewRegion("ActionPalette.Close").End()
	if p.doc != nil {
		p.doc.Destroy()
		p.doc = nil
	}
	if p.uiMan.Host != nil {
		p.uiMan.Host.Window.Keyboard.RemoveKeyCallback(p.keyKb)
	}
	if existing == p {
		existing = nil
	}
	if p.onClose != nil {
		p.onClose()
		p.onClose = nil
	}
}

func (p *ActionPalette) search(e *document.Element) {
	defer tracing.NewRegion("ActionPalette.search").End()
	if p.list == nil {
		return
	}
	query := strings.ToLower(strings.TrimSpace(e.UI.ToInput().Text()))
	for _, child := range p.list.Children {
		search := strings.ToLower(child.Attribute("data-search"))
		if query == "" || containsAllTokens(search, query) {
			child.UI.Show()
		} else {
			child.UI.Hide()
		}
	}
}

func (p *ActionPalette) clickMiss(*document.Element) {
	defer tracing.NewRegion("ActionPalette.clickMiss").End()
	p.Close()
}

func (p *ActionPalette) clickAction(e *document.Element) {
	defer tracing.NewRegion("ActionPalette.clickAction").End()
	p.runElement(e)
}

func (p *ActionPalette) runFirstVisible() {
	defer tracing.NewRegion("ActionPalette.runFirstVisible").End()
	if p.list == nil {
		return
	}
	for _, child := range p.list.Children {
		if child.UI.Entity().IsActive() {
			p.runElement(child)
			return
		}
	}
}

func (p *ActionPalette) runElement(e *document.Element) {
	idx, err := strconv.Atoi(e.Attribute("data-idx"))
	if err != nil {
		slog.Error("failed to parse action palette index", "error", err)
		return
	}
	if idx < 0 || idx >= len(p.entries) {
		slog.Error("action palette index is out of range", "index", idx, "len", len(p.entries))
		return
	}
	entry := p.entries[idx]
	p.Close()
	p.service.Run(editor_action.Request{
		ID:     entry.ID,
		Params: entry.Params,
		Source: editor_action.SourcePalette,
	})
}

func entrySearchText(entry editor_action.Entry) string {
	parts := []string{
		string(entry.ID),
		entry.Label,
		entry.Description,
		entry.Category,
	}
	parts = append(parts, entry.Tags...)
	parts = append(parts, entry.Aliases...)
	return strings.ToLower(strings.Join(parts, " "))
}

func containsAllTokens(haystack, query string) bool {
	for _, token := range strings.Fields(query) {
		if !strings.Contains(haystack, token) {
			return false
		}
	}
	return true
}
