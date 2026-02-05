/******************************************************************************/
/* context_menu_overlay.go                                                    */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
