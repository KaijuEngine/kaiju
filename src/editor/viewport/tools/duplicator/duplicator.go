package duplicator

import (
	"kaiju/editor/editor_interface"
	"kaiju/engine"
	"slices"
)

func doDuplicate(h *duplicateHistory, editor editor_interface.Editor) {
	for i := 0; i < len(h.entities) && !h.sparse; i++ {
		for j := i + 1; j < len(h.entities) && !h.sparse; j++ {
			h.sparse = h.entities[i].HasChildRecursive(h.entities[j]) ||
				h.entities[j].HasChildRecursive(h.entities[i])
		}
	}
	h.Redo()
	editor.History().Add(h)
}

func Delete(editor editor_interface.Editor, entities []*engine.Entity) {
	h := &duplicateHistory{entities, []*engine.Entity{}, editor, false}
	doDuplicate(h, editor)
}

func DeleteSelected(editor editor_interface.Editor) {
	h := &duplicateHistory{slices.Clone(editor.Selection().Entities()), []*engine.Entity{}, editor, false}
	doDuplicate(h, editor)
}
