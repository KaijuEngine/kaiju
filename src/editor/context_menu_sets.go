package editor

import (
	"kaiju/editor/ui/context_menu"
	"kaiju/editor/viewport/tools/deleter"
)

func hierarchyContextMenuActions(ed *Editor) context_menu.ContextMenuSet {
	entries := []context_menu.ContextMenuEntry{
		{Id: "delete", Label: "Delete", OnClick: func() {
			deleter.DeleteSelected(ed)
		}},
	}
	return context_menu.NewSet(ed.contextMenu, entries)
}
