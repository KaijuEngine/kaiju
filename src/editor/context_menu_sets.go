package editor

import (
	"kaiju/editor/ui/context_menu"
	"kaiju/editor/viewport/tools/deleter"
	"kaiju/editor/viewport/tools/duplicator"
)

func hierarchyContextMenuActions(ed *Editor) context_menu.ContextMenuSet {
	entries := []context_menu.ContextMenuEntry{
		{Id: "delete", Label: "Delete", OnClick: func() {
			deleter.DeleteSelected(ed)
		}},
		{Id: "duplicate", Label: "Duplicate", OnClick: func() {
			duplicator.DeleteSelected(ed)
		}},
		{Id: "details", Label: "Details", OnClick: func() {
			ed.detailsWindow.Show()
		}},
		{Id: "focus", Label: "Focus", OnClick: func() {
			ed.selection.Focus(ed.Host().Camera)
		}},
	}
	return context_menu.NewSet(ed.contextMenu, entries)
}
