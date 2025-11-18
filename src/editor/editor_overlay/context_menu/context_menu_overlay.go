package context_menu

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"strconv"
)

var existing *ContextMenu

type ContextMenu struct {
	doc     *document.Document
	uiMan   ui.Manager
	options []ContextMenuOption
}

type ContextMenuOption struct {
	Label string
	Call  func()
}

func Show(host *engine.Host, options []ContextMenuOption, screenPos matrix.Vec2) (*ContextMenu, error) {
	defer tracing.NewRegion("context_menu.Show").End()
	// Only allow one context menu open at a time
	if existing != nil {
		existing.Close()
	}
	o := &ContextMenu{options: options}
	o.uiMan.Init(host)
	var err error
	data := make([]string, len(options))
	for i := range options {
		data[i] = options[i].Label
	}
	o.doc, err = markup.DocumentFromHTMLAsset(&o.uiMan, "editor/ui/overlay/context_menu.go.html",
		data, map[string]func(*document.Element){
			"clickMiss":   o.clickMiss,
			"clickOption": o.clickOption,
		})
	if err != nil {
		return o, err
	}
	win, _ := o.doc.GetElementById("window")
	win.UI.Layout().SetOffset(screenPos.X(), screenPos.Y())
	win.UI.Clean()
	// TODO:  Ensure that it is not running off the screen
	ps := win.UI.Layout().PixelSize()
	winWidth := matrix.Float(host.Window.Width())
	winHeight := matrix.Float(host.Window.Height())
	x := min(winWidth-(screenPos.X()+ps.Width()), 0)
	y := min(winHeight-(screenPos.Y()+ps.Height()), 0)
	if x < 0 || y < 0 {
		const xPad, yPad = 10, 30
		newPos := screenPos.Add(matrix.NewVec2(x-xPad, y-yPad))
		win.UI.Layout().SetOffset(newPos.X(), newPos.Y())
	}
	existing = o
	return o, nil
}

func (o *ContextMenu) Close() {
	defer tracing.NewRegion("ConfirmPrompt.Close").End()
	o.doc.Destroy()
	existing = nil
}

func (o *ContextMenu) clickMiss(*document.Element) {
	defer tracing.NewRegion("ConfirmPrompt.clickMiss").End()
	o.Close()
}

func (o *ContextMenu) clickOption(e *document.Element) {
	defer tracing.NewRegion("ConfirmPrompt.clickOption").End()
	idx, err := strconv.Atoi(e.Attribute("data-idx"))
	defer o.Close()
	if err != nil {
		slog.Error("failed to parse the index of the context option", "error", err)
		return
	}
	if idx < 0 || idx >= len(o.options) {
		slog.Error("the index from the context menu was not valid", "index", idx, "len", len(o.options))
		return
	}
	o.options[idx].Call()
}
