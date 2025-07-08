/******************************************************************************/
/* duplicate_history.go                                                       */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package duplicator

import (
	"kaiju/editor/editor_interface"
	"kaiju/engine"
	"kaiju/engine/collision"
	"kaiju/rendering"
)

type duplicateHistory struct {
	entities   []*engine.Entity
	duplicates []*engine.Entity
	editor     editor_interface.Editor
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
				copy.ShaderData = rendering.ReflectDuplicateDrawInstance(draws[i].ShaderData)
				host.Drawings.AddDrawing(copy)
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
			e.EditorRestore(h.editor.BVH())
		}
	}
	h.editor.Selection().UntrackedClear()
	h.editor.ReloadTabs("Hierarchy")
}

func (h *duplicateHistory) Undo() {
	for _, e := range h.duplicates {
		e.EditorDelete()
	}
	h.editor.Selection().UntrackedAdd(h.entities...)
	h.editor.ReloadTabs("Hierarchy")
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
