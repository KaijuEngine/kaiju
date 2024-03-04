package details_window

import (
	"kaiju/engine"
	"kaiju/markup/document"
	"kaiju/ui"
	"reflect"
	"strconv"
	"strings"
)

func (d *Details) elmToReflectedValue(elm *document.DocElement) (reflect.Value, bool) {
	id := elm.HTML.Attribute("id")
	lr := strings.Split(id, "_")
	if len(lr) != 2 {
		return reflect.Value{}, false
	}
	dataIdx, _ := strconv.Atoi(lr[0])
	fieldIdx, _ := strconv.Atoi(lr[1])
	data := d.viewData.Data[dataIdx]
	return data.entityData.(reflect.Value).Elem().Field(fieldIdx), true
}

func inputString(input *document.DocElement) string { return input.UI.(*ui.Input).Text() }

func toInt(str string) int64 {
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toUint(str string) uint64 {
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseUint(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toFloat(str string) float64 {
	if str == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}
	return 0
}

func entityDragData(host *engine.Host) (engine.EntityId, bool) {
	eid, ok := host.Window.Mouse.DragData().(engine.EntityId)
	return eid, ok
}
