package duplicator

import (
	"kaiju/collision"
	"kaiju/editor/interfaces"
	"kaiju/engine"
	"kaiju/rendering"
	"slices"
)

type duplicateHistory struct {
	entities   []*engine.Entity
	duplicates []*engine.Entity
	editor     interfaces.Editor
	sparse     bool
}

func (h *duplicateHistory) doDuplication() {
	for _, e := range h.entities {
		host := h.editor.Host()
		dupe := e.Duplicate(h.sparse, func(from, to *engine.Entity) {
			to.GenerateId()
			// Duplicate the drawings
			draws := from.EditorBindings.Drawings()
			for i := range draws {
				copy := draws[i]
				copy.Transform = &to.Transform
				copy.Textures = slices.Clone(draws[i].Textures)
				copy.ShaderData = rendering.ReflectDuplicateDrawInstance(draws[i].ShaderData)
				host.Drawings.AddDrawing(&copy)
				to.EditorBindings.AddDrawing(copy)
			}
			// Duplicate the BVH
			bvh := from.EditorBindings.Data("bvh")
			dupeBVH := bvh.(*collision.BVH).Duplicate()
			dupeBVH.Transform = &to.Transform
			h.editor.BVH().Insert(dupeBVH)
		})
		host.AddEntity(dupe)
		h.duplicates = append(h.duplicates, dupe)
	}
}

func (h *duplicateHistory) Redo() {
	if len(h.duplicates) == 0 {
		h.doDuplication()
	} else {
		for _, e := range h.duplicates {
			draws := e.EditorBindings.Drawings()
			e.EditorBindings.IsDeleted = false
			for _, d := range draws {
				d.ShaderData.Activate()
			}
			e.Activate()
			bvh := e.EditorBindings.Data("bvh")
			if bvh != nil {
				h.editor.BVH().Insert(bvh.(*collision.BVH))
			}
		}
	}
	h.editor.Selection().UntrackedClear()
	h.editor.Hierarchy().Reload()
}

func (h *duplicateHistory) Undo() {
	for _, e := range h.duplicates {
		draws := e.EditorBindings.Drawings()
		e.EditorBindings.IsDeleted = true
		for _, d := range draws {
			d.ShaderData.Deactivate()
		}
		e.Deactivate()
		bvh := e.EditorBindings.Data("bvh")
		if bvh != nil {
			bvh.(*collision.BVH).RemoveNode()
		}
	}
	h.editor.Selection().UntrackedAdd(h.entities...)
	h.editor.Hierarchy().Reload()
}

func (h *duplicateHistory) Delete() {
	for _, e := range h.duplicates {
		drawings := e.EditorBindings.Drawings()
		for _, d := range drawings {
			d.ShaderData.Destroy()
		}
		e.Destroy()
	}
}

func (h *duplicateHistory) Exit() {}
