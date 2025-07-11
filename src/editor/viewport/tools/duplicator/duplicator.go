/******************************************************************************/
/* duplicator.go                                                              */
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
