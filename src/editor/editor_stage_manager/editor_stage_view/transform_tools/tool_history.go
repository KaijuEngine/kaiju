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

package transform_tools

import (
	"kaiju/editor/editor_stage_manager"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
)

type toolHistory struct {
	stage    StageInterface
	entities []*editor_stage_manager.StageEntity
	from     []matrix.Vec3
	to       []matrix.Vec3
	state    ToolState
}

func (h *toolHistory) Redo() {
	defer tracing.NewRegion("toolHistory.Redo").End()
	for i, e := range h.entities {
		switch h.state {
		case ToolStateMove:
			e.Transform.SetPosition(h.to[i])
		case ToolStateRotate:
			e.Transform.SetRotation(h.to[i])
		case ToolStateScale:
			e.Transform.SetScale(h.to[i])
		}
	}
	// TODO:  Use the following when the BVH.Refit function is fixed. Just
	// so there aren't any issues right now, I'm going to use a refit on
	// the first selected entity as it'll go to the root and refit all.
	// goroutine - Update all the BVHs
	//	for _, e := range t.stage.Manager().Selection() {
	//		e.StageData.Bvh.Refit()
	//	}
	h.stage.Manager().RefitBVH(h.entities[0])
}

func (h *toolHistory) Undo() {
	defer tracing.NewRegion("toolHistory.Undo").End()
	for i, e := range h.entities {
		switch h.state {
		case ToolStateMove:
			e.Transform.SetPosition(h.from[i])
		case ToolStateRotate:
			e.Transform.SetRotation(h.from[i])
		case ToolStateScale:
			e.Transform.SetScale(h.from[i])
		}
	}
	// TODO:  Use the following when the BVH.Refit function is fixed. Just
	// so there aren't any issues right now, I'm going to use a refit on
	// the first selected entity as it'll go to the root and refit all.
	// goroutine - Update all the BVHs
	//	for _, e := range t.stage.Manager().Selection() {
	//		e.StageData.Bvh.Refit()
	//	}
	h.stage.Manager().RefitBVH(h.entities[0])
}

func (h *toolHistory) Delete() {}
func (h *toolHistory) Exit()   {}
