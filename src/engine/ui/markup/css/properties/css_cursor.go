/******************************************************************************/
/* css_cursor.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
)

func (p Cursor) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return errors.New("no cursor value")
	}

	panel.Base().AddEvent(ui.EventTypeEnter, func() {
		switch values[0].Str {
		case "text":
			host.Window.CursorIbeam()
		case "default":
			host.Window.CursorStandard()
		case "context-menu":
			fallthrough
		case "help":
			fallthrough
		case "pointer":
			host.Window.CursorHand()
		case "progress":
			fallthrough
		case "wait":
			fallthrough
		case "cell":
			fallthrough
		case "crosshair":
			fallthrough
		case "vertical-text":
			fallthrough
		case "alias":
			fallthrough
		case "copy":
			fallthrough
		case "move":
			fallthrough
		case "no-drop":
			fallthrough
		case "not-allowed":
			fallthrough
		case "grab":
			fallthrough
		case "grabbing":
			fallthrough
		case "all-scroll":
			fallthrough
		case "col-resize":
			fallthrough
		case "row-resize":
			fallthrough
		case "n-resize":
			fallthrough
		case "e-resize":
			fallthrough
		case "s-resize":
			fallthrough
		case "w-resize":
			fallthrough
		case "ne-resize":
			fallthrough
		case "nw-resize":
			fallthrough
		case "se-resize":
			fallthrough
		case "sw-resize":
			fallthrough
		case "ew-resize":
			fallthrough
		case "ns-resize":
			fallthrough
		case "nesw-resize":
			fallthrough
		case "nwse-resize":
			fallthrough
		case "zoom-in":
			fallthrough
		case "zoom-out":
			klib.NotYetImplemented(180)
		default:
			host.Window.CursorStandard()
		}
	})

	panel.Base().AddEvent(ui.EventTypeExit, func() {
		host.Window.CursorStandard()
	})

	return nil
}
