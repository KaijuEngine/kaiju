package markup

import (
	"kaiju/systems/events"
	"kaiju/ui"
)

func tryMap(attr string, elm *DocElement, evt *events.Event, funcMap map[string]func(*DocElement)) {
	if funcName := elm.HTML.Attribute(attr); len(funcName) > 0 {
		if f, ok := funcMap[funcName]; ok {
			evt.Add(func() { f(elm) })
		}
	}
}

func setupEvents(elm *DocElement, funcMap map[string]func(*DocElement)) {
	tryMap("onclick", elm, elm.UI.Event(ui.EventTypeClick), funcMap)
	tryMap("onmouseover", elm, elm.UI.Event(ui.EventTypeEnter), funcMap)
	tryMap("onmouseenter", elm, elm.UI.Event(ui.EventTypeEnter), funcMap)
	tryMap("onmouseleave", elm, elm.UI.Event(ui.EventTypeExit), funcMap)
	tryMap("onmouseexit", elm, elm.UI.Event(ui.EventTypeExit), funcMap)
	tryMap("onmousedown", elm, elm.UI.Event(ui.EventTypeDown), funcMap)
	tryMap("onmouseup", elm, elm.UI.Event(ui.EventTypeUp), funcMap)
	tryMap("onmousewheel", elm, elm.UI.Event(ui.EventTypeScroll), funcMap)
	tryMap("onchange", elm, elm.UI.Event(ui.EventTypeChange), funcMap)
}
