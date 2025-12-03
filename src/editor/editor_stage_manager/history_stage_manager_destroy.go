/******************************************************************************/
/* history_stage_manager_destroy.go                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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

package editor_stage_manager

import (
	"kaiju/engine/collision"
	"kaiju/platform/profiler/tracing"
)

type objectDeleteHistory struct {
	m *StageManager
	// TODO:  Only add the root-most entities to this list
	entities []*StageEntity
}

func (h *objectDeleteHistory) Redo() {
	defer tracing.NewRegion("objectDeleteHistory.Redo").End()
	for _, e := range h.entities {
		h.m.host.RemoveEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Deactivate()
		}
		h.m.OnEntityDestroy.Execute(e)
		e.isDeleted = true
		if e.StageData.Bvh != nil {
			collision.RemoveAllLeavesMatchingTransform(&h.m.worldBVH, &e.Transform)
		}
	}
}

func (h *objectDeleteHistory) Undo() {
	defer tracing.NewRegion("objectDeleteHistory.Undo").End()
	for _, e := range h.entities {
		h.m.host.AddEntity(&e.Entity)
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Activate()
		}
		h.m.OnEntitySpawn.Execute(e)
		e.isDeleted = false
		if e.StageData.Bvh != nil {
			h.m.AddBVH(e.StageData.Bvh, &e.Transform)
		}
	}
	for _, e := range h.entities {
		if e.Parent != nil {
			h.m.OnEntityChangedParent.Execute(e)
		}
	}
}

func (h *objectDeleteHistory) Delete() {}

func (h *objectDeleteHistory) Exit() {
	for _, e := range h.entities {
		e.StageData.ShaderData.Destroy()
		e.Destroy()
	}
}
