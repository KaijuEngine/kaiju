/******************************************************************************/
/* html_events.go                                                             */
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

package document

import (
	"kaiju/engine/systems/events"
	"kaiju/engine/ui"
	"log/slog"
)

func tryMap(attr string, elm *Element, evt *events.Event, funcMap map[string]func(*Element)) {
	if funcName := elm.Attribute(attr); len(funcName) > 0 {
		if f, ok := funcMap[funcName]; ok {
			evt.Add(func() { f(elm) })
		} else {
			slog.Warn("Failed to find the event function",
				slog.String("func", funcName),
				slog.String("event", attr))
		}
	}
}

func tryExecute(attr string, elm *Element, funcMap map[string]func(*Element)) {
	if funcName := elm.Attribute(attr); len(funcName) > 0 {
		if f, ok := funcMap[funcName]; ok {
			f(elm)
		} else {
			slog.Warn("Failed to find the event function",
				slog.String("func", funcName),
				slog.String("event", attr))
		}
	}
}

func setupEvents(elm *Element, funcMap map[string]func(*Element)) {
	tryMap("onclick", elm, elm.UI.Event(ui.EventTypeClick), funcMap)
	tryMap("onrightclick", elm, elm.UI.Event(ui.EventTypeRightClick), funcMap)
	tryMap("onmiss", elm, elm.UI.Event(ui.EventTypeMiss), funcMap)
	tryMap("onsubmit", elm, elm.UI.Event(ui.EventTypeSubmit), funcMap)
	tryMap("onkeydown", elm, elm.UI.Event(ui.EventTypeKeyDown), funcMap)
	tryMap("onkeyup", elm, elm.UI.Event(ui.EventTypeKeyUp), funcMap)
	tryMap("ondblclick", elm, elm.UI.Event(ui.EventTypeDoubleClick), funcMap)
	tryMap("onmouseover", elm, elm.UI.Event(ui.EventTypeEnter), funcMap)
	tryMap("onmouseenter", elm, elm.UI.Event(ui.EventTypeEnter), funcMap)
	tryMap("onmouseleave", elm, elm.UI.Event(ui.EventTypeExit), funcMap)
	tryMap("onmousemove", elm, elm.UI.Event(ui.EventTypeMove), funcMap)
	tryMap("onmouseexit", elm, elm.UI.Event(ui.EventTypeExit), funcMap)
	tryMap("onmousedown", elm, elm.UI.Event(ui.EventTypeDown), funcMap)
	tryMap("onmouseup", elm, elm.UI.Event(ui.EventTypeUp), funcMap)
	tryMap("onmousewheel", elm, elm.UI.Event(ui.EventTypeScroll), funcMap)
	tryMap("onchange", elm, elm.UI.Event(ui.EventTypeChange), funcMap)
	tryMap("ondragenter", elm, elm.UI.Event(ui.EventTypeDropEnter), funcMap)
	tryMap("ondragleave", elm, elm.UI.Event(ui.EventTypeDropExit), funcMap)
	tryMap("ondragstart", elm, elm.UI.Event(ui.EventTypeDragStart), funcMap)
	tryMap("ondrop", elm, elm.UI.Event(ui.EventTypeDrop), funcMap)
	tryMap("ondragend", elm, elm.UI.Event(ui.EventTypeDragEnd), funcMap)
	// Special case for onload
	tryExecute("onload", elm, funcMap)
}
