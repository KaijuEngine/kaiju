/******************************************************************************/
/* delete_history.go                                                          */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package deleter

import (
	"kaiju/collision"
	"kaiju/editor/interfaces"
	"kaiju/engine"
)

type deleteHistory struct {
	entities []*engine.Entity
	editor   interfaces.Editor
}

func (h *deleteHistory) Redo() {
	for _, e := range h.entities {
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
	h.editor.Selection().UntrackedClear()
	h.editor.Hierarchy().Reload()
}

func (h *deleteHistory) Undo() {
	for _, e := range h.entities {
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
	h.editor.Selection().UntrackedAdd(h.entities...)
	h.editor.Hierarchy().Reload()
}

func (h *deleteHistory) Delete() {}

func (h *deleteHistory) Exit() {
	for _, e := range h.entities {
		drawings := e.EditorBindings.Drawings()
		for _, d := range drawings {
			d.ShaderData.Destroy()
		}
		e.Destroy()
	}
}
