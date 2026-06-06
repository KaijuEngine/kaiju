/******************************************************************************/
/* context_menu_overlay.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package context_menu

import (
	"log/slog"
	"strconv"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

var existing *ContextMenu

func IsOpen() bool {
	return existing != nil
}

// IsPointInside returns true when a centered-screen UI point is inside the
// currently open context menu bounds.
func IsPointInside(centeredScreenPos matrix.Vec2) bool {
	if existing == nil || existing.doc == nil {
		return false
	}
	win, ok := existing.doc.GetElementById("window")
	if !ok {
		return false
	}
	return win.UI.Entity().Transform.ContainsPoint2D(centeredScreenPos)
}

type ContextMenu struct {
	doc     *document.Document
	uiMan   ui.Manager
	options []ContextMenuOption
	onClose func()
}

type ContextMenuOption struct {
	Label string
	Call  func()
}

func Show(host *engine.Host, options []ContextMenuOption, screenPos matrix.Vec2, onClose func()) (*ContextMenu, error) {
	defer tracing.NewRegion("context_menu.Show").End()

	if len(options) == 0 {
		options = append(options, ContextMenuOption{
			Label: "No Action possible",
			Call: func() {
				// do nothing , just consume the event
			},
		})
	}

	// Only allow one context menu open at a time
	if existing != nil {
		existing.Close()
	}
	o := &ContextMenu{
		options: options,
		onClose: onClose,
	}
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
	if o.onClose != nil {
		o.onClose()
	}
}

func (o *ContextMenu) clickMiss(*document.Element) {
	defer tracing.NewRegion("ConfirmPrompt.clickMiss").End()
	o.Close()
}

func (o *ContextMenu) clickOption(e *document.Element) {
	defer tracing.NewRegion("ConfirmPrompt.clickOption").End()
	idx, err := strconv.Atoi(e.Attribute("data-idx"))
	if o.onClose != nil {
		o.onClose() // This needs to be called before the option Call
		o.onClose = nil
	}
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
