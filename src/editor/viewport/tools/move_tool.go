package tools

import (
	"kaiju/cameras"
	"kaiju/editor/selection"
	"kaiju/engine"
	"kaiju/matrix"
)

type MoveTool struct {
	HandleTool
	toolStart matrix.Vec3
	starts    []matrix.Vec3
}

func (t *MoveTool) Initialize(host *engine.Host, selection *selection.Selection) {
	t.init(host, selection, "editor/meshes/move-pointer.gltf")
}

func (t *MoveTool) Update() {
	mp := t.host.Window.Mouse.Position()
	if t.isDragging {
		t.DragUpdate(mp, t.host.Camera)
	} else {
		t.updateScale(t.host.Camera.Position())
		if t.CheckHover(mp, t.host.Camera) {

		}
	}
}

func (t *MoveTool) DragUpdate(pointerPos matrix.Vec2, camera cameras.Camera) {
	t.starts = t.starts[:0]
	for _, e := range t.selection.Entities() {
		t.starts = append(t.starts, e.Transform.Position())
	}
	t.toolStart = t.tool.Transform.Position()
	t.HandleTool.dragUpdate(pointerPos, camera, t.processDelta)
}

func (t *MoveTool) DragStop() {
	t.HandleTool.dragStop()
	//_engine->history->add_memento(history_transform_move(
	//	_engine, _selection, _starts, hierarchy_get_positions(_selection)));
}

func (t *MoveTool) processDelta(length matrix.Vec3) {
	t.tool.Transform.SetPosition(t.toolStart.Add(length))
	for i := range t.starts {
		t.selection.Entities()[i].Transform.SetPosition(t.starts[i].Add(length))
	}
}
