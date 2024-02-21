/******************************************************************************/
/* move_tool.go                                                               */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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

package tools

import (
	"kaiju/cameras"
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
)

type MoveTool struct {
	HandleToolBase
	toolStart matrix.Vec3
	starts    []matrix.Vec3
}

func (t *MoveTool) Initialize(host *engine.Host, selection *selection.Selection, renderTarget rendering.Canvas) {
	t.init(host, selection, renderTarget, "editor/meshes/move-pointer.gltf")
}

func (t *MoveTool) Update() (changed bool) {
	return t.HandleToolBase.internalUpdate(t)
}

func (t *MoveTool) DragStart(pointerPos matrix.Vec2, camera cameras.Camera) {
	t.HandleToolBase.DragStart(pointerPos, camera)
	t.starts = t.starts[:0]
	for _, e := range t.selection.Entities() {
		t.starts = append(t.starts, e.Transform.Position())
	}
	t.toolStart = t.tool.Transform.Position()
}

func (t *MoveTool) DragUpdate(pointerPos matrix.Vec2, camera cameras.Camera) {
	t.HandleToolBase.dragUpdate(pointerPos, camera, t.processDelta)
}

func (t *MoveTool) DragStop() {
	t.HandleToolBase.dragStop()
	//_engine->history->add_memento(history_transform_move(
	//	_engine, _selection, _starts, hierarchy_get_positions(_selection)));
}

func (t *MoveTool) processDelta(length matrix.Vec3) {
	t.tool.Transform.SetPosition(t.toolStart.Add(length))
	for i := range t.starts {
		t.selection.Entities()[i].Transform.SetPosition(t.starts[i].Add(length))
	}
}
