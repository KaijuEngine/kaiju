/******************************************************************************/
/* editor_stage_transformation_manager.go                                     */
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

package editor_stage_view

import (
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_stage_manager/editor_stage_view/transform_tools"
	"kaiju/editor/memento"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"weak"
)

type ToolState = uint8

const (
	ToolStateNone ToolState = iota
	ToolStateMove
	ToolStateRotate
	ToolStateScale
)

type TransformationManager struct {
	view           weak.Pointer[StageView]
	translateTool  transform_tools.TranslationTool
	rotationTool   transform_tools.RotationTool
	scalingTool    transform_tools.ScalingTool
	transformStart matrix.Vec3
	scalingStart   matrix.Vec3
	manager        *editor_stage_manager.StageManager
	history        *memento.History
	memento        *transformHistory
	snapSettings   *editor_settings.SnapSettings
	currentTool    ToolState
	isBusy         bool
}

func (t *TransformationManager) IsBusy() bool { return t.isBusy }

func (t *TransformationManager) Initialize(stageView *StageView, history *memento.History, snapSettings *editor_settings.SnapSettings) {
	t.view = weak.Make(stageView)
	t.snapSettings = snapSettings
	t.translateTool.Initialize(stageView.host)
	t.rotationTool.Initialize(stageView.host)
	t.scalingTool.Initialize(stageView.host)
	t.manager = &stageView.manager
	t.history = history
	t.manager.OnEntitySelected.Add(func(e *editor_stage_manager.StageEntity) {
		t.currentTool = ToolStateNone
		t.setToolState(t.currentTool, e.Transform.Position())
	})
	t.manager.OnEntityDeselected.Add(func(e *editor_stage_manager.StageEntity) {
		if !t.manager.HasSelection() {
			t.translateTool.Hide()
			t.rotationTool.Hide()
			t.scalingTool.Hide()
		}
	})
	t.translateTool.OnDragStart.Add(t.translateStart)
	t.translateTool.OnDragMove.Add(t.translateMove)
	t.translateTool.OnDragEnd.Add(t.translateEnd)
	t.rotationTool.OnDragStart.Add(t.rotateStart)
	t.rotationTool.OnDragRotate.Add(t.rotateSpin)
	t.rotationTool.OnDragEnd.Add(t.rotateEnd)
	t.scalingTool.OnDragStart.Add(t.scaleStart)
	t.scalingTool.OnDragScale.Add(t.scaleScale)
	t.scalingTool.OnDragEnd.Add(t.scaleEnd)
}

func (t *TransformationManager) Update(host *engine.Host) {
	kb := &host.Window.Keyboard
	if !t.isBusy {
		pos := matrix.Vec3NaN()
		if t.manager.HasSelection() {
			pos = t.manager.LastSelected().Transform.Position()
		}
		if kb.KeyDown(hid.KeyboardKey1) {
			t.setToolState(ToolStateMove, pos)
		} else if kb.KeyDown(hid.KeyboardKey2) {
			t.setToolState(ToolStateRotate, pos)
		} else if kb.KeyDown(hid.KeyboardKey3) {
			t.setToolState(ToolStateScale, pos)
		}
	}
	ss := t.snapSettings
	snap := kb.HasCtrl()
	t.isBusy = t.translateTool.Update(host, snap, ss.TranslateIncrement) ||
		t.rotationTool.Update(host, snap, ss.RotateIncrement) ||
		t.scalingTool.Update(host, snap, ss.ScaleIncrement)
}

func (t *TransformationManager) setToolState(state ToolState, pos matrix.Vec3) {
	if t.currentTool == state {
		state = ToolStateNone
	}
	t.translateTool.Hide()
	t.rotationTool.Hide()
	t.scalingTool.Hide()
	t.currentTool = state
	if !pos.IsNaN() {
		switch t.currentTool {
		case ToolStateNone:
		case ToolStateMove:
			t.translateTool.Show(pos)
		case ToolStateRotate:
			t.rotationTool.Show(pos)
		case ToolStateScale:
			t.scalingTool.Show(pos)
		}
	}
}

func (t *TransformationManager) translateStart(pos matrix.Vec3) {
	defer tracing.NewRegion("TransformationManager.translateStart").End()
	t.transformStart = pos
	t.setupMemento()
}

func (t *TransformationManager) translateMove(pos matrix.Vec3) {
	defer tracing.NewRegion("TransformationManager.translateMove").End()
	sel := t.manager.HierarchyRespectiveSelection()
	delta := pos.Subtract(t.transformStart)
	for i := range sel {
		t.memento.to[i].position = t.memento.from[i].position.Add(delta)
		sel[i].Transform.SetWorldPosition(t.memento.to[i].position)
	}
}

func (t *TransformationManager) translateEnd(pos matrix.Vec3) {
	defer tracing.NewRegion("TransformationManager.translateEnd").End()
	t.translateMove(pos)
	t.history.Add(t.memento)
}

func (t *TransformationManager) rotateStart(rot matrix.Vec4) {
	defer tracing.NewRegion("TransformationManager.rotateStart").End()
	t.setupMemento()
}

func (t *TransformationManager) rotateSpin(rot matrix.Vec4) {
	defer tracing.NewRegion("TransformationManager.rotateSpin").End()
	if matrix.Approx(rot.W(), 0) || rot.Equals(matrix.Vec4Zero()) {
		return
	}
	angle := matrix.Deg2Rad(rot.W())
	pivot := t.memento.toolTarget.Transform.Position()
	rotMat := matrix.Mat4Identity()
	rotMat.Rotate(rot.AsVec3().Scale(rot.W()))
	sel := t.manager.HierarchyRespectiveSelection()
	for i := range sel {
		offset := t.memento.from[i].position.Subtract(pivot)
		rotated := rotMat.TransformPoint(offset)
		newPos := rotated.Add(pivot)
		t.memento.to[i].position = newPos
		sel[i].Transform.SetWorldPosition(newPos)
		currentQuat := matrix.QuaternionFromEuler(t.memento.from[i].rotation)
		incrementalQuat := matrix.QuaternionAxisAngle(rot.AsVec3(), angle)
		// newQuat := currentQuat.Multiply(incrementalQuat) // Local space
		newQuat := incrementalQuat.Multiply(currentQuat) // World space
		newRot := newQuat.ToEuler()
		t.memento.to[i].rotation = newRot
		sel[i].Transform.SetWorldRotation(newRot)
	}
}

func (t *TransformationManager) rotateEnd(rot matrix.Vec4) {
	defer tracing.NewRegion("TransformationManager.rotateEnd").End()
	t.rotateSpin(rot)
	t.history.Add(t.memento)
}

func (t *TransformationManager) scaleStart(scale matrix.Vec3) {
	defer tracing.NewRegion("TransformationManager.scaleStart").End()
	t.scalingStart = scale
	t.setupMemento()
}

func (t *TransformationManager) scaleScale(scale matrix.Vec3) {
	defer tracing.NewRegion("TransformationManager.translateMove").End()
	sel := t.manager.HierarchyRespectiveSelection()
	delta := scale.Subtract(t.scalingStart)
	for i := range sel {
		fm := matrix.Mat4Identity()
		fm.Rotate(sel[i].Transform.WorldRotation())
		worldAxis := matrix.Vec3Zero()
		maxDelta := matrix.Float(0)
		axisIndex := -1
		for k := 0; k < 3; k++ {
			absD := matrix.Abs(delta[k])
			if absD > maxDelta {
				maxDelta = absD
				axisIndex = k
			}
		}
		if axisIndex >= 0 {
			worldAxis[axisIndex] = 1.0
			relativeChange := delta[axisIndex] // Relative multiplier change
			localDir := fm.Transpose().TransformPoint(worldAxis)
			for k := 0; k < 3; k++ {
				proj := matrix.Abs(localDir[k])
				newScale := t.memento.from[i].scale[k] * max(0.001, 1+relativeChange*proj)
				t.memento.to[i].scale[k] = newScale
			}
			sel[i].Transform.SetWorldScale(t.memento.to[i].scale)
		}
	}
}

func (t *TransformationManager) scaleEnd(scale matrix.Vec3) {
	defer tracing.NewRegion("TransformationManager.scaleEnd").End()
	t.scaleScale(scale)
	t.history.Add(t.memento)
}

func (t *TransformationManager) setupMemento() {
	defer tracing.NewRegion("TransformationManager.setupMemento").End()
	sel := t.manager.HierarchyRespectiveSelection()
	t.memento = &transformHistory{
		tman:       t,
		toolTarget: t.manager.LastSelected(),
	}
	t.memento.entities = make([]*editor_stage_manager.StageEntity, len(sel))
	t.memento.from = make([]transformHistoryPRS, len(sel))
	t.memento.to = make([]transformHistoryPRS, len(sel))
	for i := range sel {
		t.memento.entities[i] = sel[i]
		t.memento.from[i].position = sel[i].Transform.WorldPosition()
		t.memento.to[i].position = t.memento.from[i].position
		t.memento.from[i].rotation = sel[i].Transform.WorldRotation()
		t.memento.to[i].rotation = t.memento.from[i].rotation
		t.memento.from[i].scale = sel[i].Transform.WorldScale()
		t.memento.to[i].scale = t.memento.from[i].scale
	}
}
