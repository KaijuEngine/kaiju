/******************************************************************************/
/* tool_history.go                                                            */
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

package transform_tools

import (
	"kaiju/editor/editor_interface"
	"kaiju/engine"
	"kaiju/matrix"
)

type toolHistory struct {
	editor   editor_interface.Editor
	entities []*engine.Entity
	from     []matrix.Vec3
	to       []matrix.Vec3
	state    ToolState
}

func (h *toolHistory) Redo() {
	for i, e := range h.entities {
		if h.state == ToolStateMove {
			e.Transform.SetPosition(h.to[i])
		} else if h.state == ToolStateRotate {
			e.Transform.SetRotation(h.to[i])
		} else if h.state == ToolStateScale {
			e.Transform.SetScale(h.to[i])
		}
	}
	h.editor.BVHEntityUpdates(h.editor.Selection().Entities()...)
}

func (h *toolHistory) Undo() {
	for i, e := range h.entities {
		if h.state == ToolStateMove {
			e.Transform.SetPosition(h.from[i])
		} else if h.state == ToolStateRotate {
			e.Transform.SetRotation(h.from[i])
		} else if h.state == ToolStateScale {
			e.Transform.SetScale(h.from[i])
		}
	}
	h.editor.BVHEntityUpdates(h.editor.Selection().Entities()...)
}

func (h *toolHistory) Delete() {}
func (h *toolHistory) Exit()   {}
