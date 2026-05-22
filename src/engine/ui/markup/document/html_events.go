/******************************************************************************/
/* html_events.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package document

import (
	"log/slog"

	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui"
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
	tryMap("onfocus", elm, elm.UI.Event(ui.EventTypeFocus), funcMap)
	tryMap("onblur", elm, elm.UI.Event(ui.EventTypeBlur), funcMap)
	tryMap("onclick", elm, elm.UI.Event(ui.EventTypeClick), funcMap)
	tryMap("onrightclick", elm, elm.UI.Event(ui.EventTypeRightClick), funcMap)
	tryMap("oncontextmenu", elm, elm.UI.Event(ui.EventTypeRightClick), funcMap)
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
