package tools

import (
	"kaiju/cameras"
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/matrix"
)

type ScaleTool struct {
	HandleTool
	starts []matrix.Vec3
}

func (t *ScaleTool) Initialize(host *engine.Host, selection *selection.Selection) {
	// TODO:  Use a screen plane for scale since the tool doesn't move
	t.init(host, selection, "editor/meshes/scale-pointer.gltf")
}

func (t *ScaleTool) DragUpdate(pointerPos matrix.Vec2, camera cameras.Camera) {
	t.starts = t.starts[:0]
	for _, e := range t.selection.Entities() {
		t.starts = append(t.starts, e.Transform.Scale())
	}
	t.HandleTool.dragUpdate(pointerPos, camera, t.processDelta)
}

func (t *ScaleTool) DragStop() {
	t.HandleTool.dragStop()
	//_engine->history->add_memento(history_transform_scale(_engine,
	//	_selection, _starts, hierarchy_get_scales(_selection)));
}

func (t *ScaleTool) processDelta(length matrix.Vec3) {
	for i := range t.starts {
		t.selection.Entities()[i].Transform.SetScale(t.starts[i].Add(length))
	}
}
