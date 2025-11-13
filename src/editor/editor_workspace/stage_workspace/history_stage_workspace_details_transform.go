/******************************************************************************/
/* history_stage_workspace_details_transform.go                               */
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

package stage_workspace

import (
	"kaiju/editor/editor_stage_manager"
	"kaiju/matrix"
)

type transformHistoryType = int

const (
	transformHistoryTypePosition = transformHistoryType(iota)
	transformHistoryTypeRotation
	transformHistoryTypeScale
)

type detailTransformHistory struct {
	entities      []*editor_stage_manager.StageEntity
	transformType transformHistoryType
	prevValues    []matrix.Vec3
	nextValues    []matrix.Vec3
}

func (h *detailTransformHistory) Redo() {
	for i, e := range h.entities {
		switch h.transformType {
		case transformHistoryTypePosition:
			e.Transform.SetPosition(h.nextValues[i])
		case transformHistoryTypeRotation:
			e.Transform.SetRotation(h.nextValues[i])
		case transformHistoryTypeScale:
			e.Transform.SetScale(h.nextValues[i])
		}
	}
}

func (h *detailTransformHistory) Undo() {
	for i, e := range h.entities {
		switch h.transformType {
		case transformHistoryTypePosition:
			e.Transform.SetPosition(h.prevValues[i])
		case transformHistoryTypeRotation:
			e.Transform.SetRotation(h.prevValues[i])
		case transformHistoryTypeScale:
			e.Transform.SetScale(h.prevValues[i])
		}
	}
}

func (h *detailTransformHistory) Delete() {}
func (h *detailTransformHistory) Exit()   {}
