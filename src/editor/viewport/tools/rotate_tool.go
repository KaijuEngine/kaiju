package tools

import (
	"kaiju/cameras"
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/matrix"
)

type RotateTool struct {
	HandleTool
	starts []matrix.Vec3
}

func (t *RotateTool) Initialize(host *engine.Host, selection *selection.Selection) {
	// TODO:  Use a screen plane for rotation since the tool doesn't move
	t.init(host, selection, "editor/meshes/rotate-pointer.gltf")
}

func (t *RotateTool) DragUpdate(pointerPos matrix.Vec2, camera cameras.Camera) {
	t.starts = t.starts[:0]
	for _, e := range t.selection.Entities() {
		t.starts = append(t.starts, e.Transform.Rotation())
	}
	t.HandleTool.dragUpdate(pointerPos, camera, t.processDelta)
}

func (t *RotateTool) DragStop() {
	t.HandleTool.dragStop()
	//_engine->history->add_memento(history_transform_rotate(_engine,
	//	_selection, _starts, hierarchy_get_scales(_selection)));
}

func (t *RotateTool) processDelta(length matrix.Vec3) {
	length.ScaleAssign(rotateScale)
	for i := range t.starts {
		t.selection.Entities()[i].Transform.SetRotation(t.starts[i].Add(length))
	}
}
