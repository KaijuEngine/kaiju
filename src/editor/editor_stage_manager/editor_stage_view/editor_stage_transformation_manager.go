package editor_stage_view

import (
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_stage_manager/editor_stage_view/transform_tools"
	"kaiju/editor/memento"
	"kaiju/engine"
	"kaiju/matrix"
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
	transformStart matrix.Vec3
	manager        *editor_stage_manager.StageManager
	history        *memento.History
	memento        *transformHistory
	isBusy         bool
}

func (t *TransformationManager) IsBusy() bool { return t.isBusy }

func (t *TransformationManager) Initialize(stageView *StageView, history *memento.History) {
	t.view = weak.Make(stageView)
	t.translateTool.Initialize(stageView.host)
	t.rotationTool.Initialize(stageView.host)
	t.manager = &stageView.manager
	t.history = history
	t.manager.OnEntitySelected.Add(func(e *editor_stage_manager.StageEntity) {
		// t.translateTool.Show(e.Transform.Position())
		// t.rotationTool.Show(e.Transform.Position())
	})
	t.manager.OnEntityDeselected.Add(func(e *editor_stage_manager.StageEntity) {
		if !t.manager.HasSelection() {
			// t.translateTool.Hide()
			// t.rotationTool.Hide()
		}
	})
	t.translateTool.OnDragStart.Add(t.translateStart)
	t.translateTool.OnDragMove.Add(t.translateMove)
	t.translateTool.OnDragEnd.Add(t.translateEnd)
	t.rotationTool.OnDragStart.Add(t.rotateStart)
	t.rotationTool.OnDragRotate.Add(t.rotateSpin)
	t.rotationTool.OnDragEnd.Add(t.rotateEnd)
}

	t.isBusy = t.translateTool.Update(host) || t.rotationTool.Update(host)
}

func (t *TransformationManager) translateStart(pos matrix.Vec3) {
	tracing.NewRegion("TransformationManager.translateStart")
	t.transformStart = pos
	t.setupMemento()
}

func (t *TransformationManager) translateMove(pos matrix.Vec3) {
	tracing.NewRegion("TransformationManager.translateMove")
	sel := t.manager.HierarchyRespectiveSelection()
	delta := pos.Subtract(t.transformStart)
	for i := range sel {
		t.memento.to[i].position = t.memento.from[i].position.Add(delta)
		sel[i].Transform.SetWorldPosition(t.memento.to[i].position)
	}
}

func (t *TransformationManager) translateEnd(pos matrix.Vec3) {
	tracing.NewRegion("TransformationManager.translateEnd")
	t.translateMove(pos)
	t.history.Add(t.memento)
}

func (t *TransformationManager) rotateStart(rot matrix.Vec3) {
	tracing.NewRegion("TransformationManager.rotateStart")
	t.transformStart = rot
	t.setupMemento()
}

func (t *TransformationManager) rotateSpin(rot matrix.Vec3) {
	tracing.NewRegion("TransformationManager.rotateSpin")
	if rot.Equals(matrix.Vec3Zero()) {
		return
	}
	sel := t.manager.HierarchyRespectiveSelection()
	rotMat := matrix.Mat4Identity()
	rotMat.Rotate(rot)
	rotMat.Translate(t.memento.toolTarget.Transform.Position())
	for i := range sel {
		t.memento.to[i].position = rotMat.TransformPoint(t.memento.from[i].position)
	}
}

func (t *TransformationManager) rotateEnd(rot matrix.Vec3) {
	tracing.NewRegion("TransformationManager.rotateEnd")
	t.rotateSpin(rot)
	t.history.Add(t.memento)
}

func (t *TransformationManager) setupMemento() {
	tracing.NewRegion("TransformationManager.setupMemento")
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
