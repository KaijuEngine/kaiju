/******************************************************************************/
/* history_stage_manager_spawn.go                                             */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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

type objectSpawnHistory struct {
	m *StageManager
	e *StageEntity
}

func (h *objectSpawnHistory) Redo() {
	defer tracing.NewRegion("objectSpawnHistory.Redo").End()
	h.e.Activate()
	h.e.isDeleted = false
	h.m.host.AddEntity(&h.e.Entity)
	if h.e.StageData.ShaderData != nil {
		h.e.StageData.ShaderData.Activate()
	}
	if h.e.StageData.Bvh != nil {
		h.m.AddBVH(h.e.StageData.Bvh, &h.e.Transform)
	}
	h.m.OnEntitySpawn.Execute(h.e)
}

func (h *objectSpawnHistory) Undo() {
	defer tracing.NewRegion("objectSpawnHistory.Undo").End()
	h.e.Deactivate()
	h.e.isDeleted = true
	h.m.host.RemoveEntity(&h.e.Entity)
	if h.e.StageData.ShaderData != nil {
		h.e.StageData.ShaderData.Deactivate()
	}
	if h.e.StageData.Bvh != nil {
		collision.RemoveAllLeavesMatchingTransform(&h.m.worldBVH, &h.e.Transform)
	}
	h.m.OnEntityDestroy.Execute(h.e)
}

func (h *objectSpawnHistory) Delete() {
	if h.e.StageData.ShaderData != nil {
		h.e.StageData.ShaderData.Destroy()
	}
	h.e.Destroy()
}

func (h *objectSpawnHistory) Exit() {}
