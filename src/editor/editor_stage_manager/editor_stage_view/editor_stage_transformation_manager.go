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
}

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
	t.rotationTool.OnDragStart.Add(t.translateStart)
	t.rotationTool.OnDragRotate.Add(t.translateMove)
	t.rotationTool.OnDragEnd.Add(t.translateEnd)
}

func (t *TransformationManager) Update(host *engine.Host) bool {
	return t.translateTool.Update(host) || t.rotationTool.Update(host)
}

func (t *TransformationManager) translateStart(pos matrix.Vec3) {
	tracing.NewRegion("TransformationManager.translateStart")
	t.transformStart = pos
	sel := t.manager.HierarchyRespectiveSelection()
	t.memento = &transformHistory{
		tman:       t,
		toolTarget: t.manager.LastSelected(),
		state:      ToolStateMove,
	}
	t.memento.entities = make([]*editor_stage_manager.StageEntity, len(sel))
	t.memento.from = make([]matrix.Vec3, len(sel))
	t.memento.to = make([]matrix.Vec3, len(sel))
	for i := range sel {
		t.memento.entities[i] = sel[i]
		t.memento.from[i] = sel[i].Transform.WorldPosition()
		t.memento.to[i] = sel[i].Transform.WorldPosition()
	}
}

func (t *TransformationManager) translateMove(pos matrix.Vec3) {
	tracing.NewRegion("TransformationManager.translateMove")
	sel := t.manager.HierarchyRespectiveSelection()
	delta := pos.Subtract(t.transformStart)
	for i := range sel {
		t.memento.to[i] = t.memento.from[i].Add(delta)
		sel[i].Transform.SetWorldPosition(t.memento.to[i])
	}
}

func (t *TransformationManager) translateEnd(pos matrix.Vec3) {
	tracing.NewRegion("TransformationManager.translateEnd")
	t.translateMove(pos)
	t.history.Add(t.memento)
}
