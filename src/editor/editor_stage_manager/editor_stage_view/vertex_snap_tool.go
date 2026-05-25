/******************************************************************************/
/* vertex_snap_tool.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"log/slog"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	vertexSnapPickRadiusPixels = matrix.Float(12)
	vertexSnapMarkerScale      = matrix.Float(0.012)
)

type VertexSnapTool struct {
	view             *StageView
	transformManager *TransformationManager
	sourceMarker     vertexSnapMarker
	targetMarker     vertexSnapMarker
	source           vertexSnapCandidate
	sourceStartWorld matrix.Vec3
	dragging         bool
}

type vertexSnapMarker struct {
	transform  matrix.Transform
	shaderData rendering.DrawInstance
}

type vertexSnapCandidate struct {
	Entity       *editor_stage_manager.StageEntity
	Local        matrix.Vec3
	World        matrix.Vec3
	Screen       matrix.Vec3
	ScreenDistSq matrix.Float
}

func (t *VertexSnapTool) Initialize(host *engine.Host, view *StageView, transformManager *TransformationManager) {
	defer tracing.NewRegion("VertexSnapTool.Initialize").End()
	t.view = view
	t.transformManager = transformManager
	t.sourceMarker.Initialize(host, matrix.ColorYellow())
	t.targetMarker.Initialize(host, matrix.ColorCyan())
	t.Hide()
}

func (t *VertexSnapTool) IsBusy() bool { return t.dragging }

func (t *VertexSnapTool) Update(host *engine.Host) bool {
	defer tracing.NewRegion("VertexSnapTool.Update").End()
	if t.view == nil || t.transformManager == nil {
		return false
	}
	if t.dragging {
		t.updateDrag(host)
		if host.Window.Keyboard.KeyDown(hid.KeyboardKeyEscape) {
			t.cancelDrag()
			return true
		}
		if host.Window.Cursor.Released() {
			t.commitDrag(host)
			return true
		}
		return true
	}
	if t.view.camera.Mode() != editor_controls.EditorCameraMode3d ||
		!host.Window.Keyboard.KeyHeld(hid.KeyboardKeyV) ||
		!t.transformManager.manager.HasSelection() {
		t.Hide()
		return false
	}
	source, ok := t.pickSource(host)
	if ok {
		t.source = source
		t.sourceMarker.Show(source.World, vertexSnapMarkerWorldScale(host.PrimaryCamera(), source.World))
	} else {
		t.sourceMarker.Hide()
	}
	t.targetMarker.Hide()
	if ok && host.Window.Cursor.Pressed() {
		t.beginDrag(host, source)
	}
	return true
}

func (t *VertexSnapTool) Hide() {
	t.sourceMarker.Hide()
	t.targetMarker.Hide()
}

func (t *VertexSnapTool) pickSource(host *engine.Host) (vertexSnapCandidate, bool) {
	manager := t.transformManager.manager
	if !manager.HasSelection() {
		return vertexSnapCandidate{}, false
	}
	return closestSnapVertexOnEntity(manager.LastSelected(), host.Window.Cursor.Position(),
		host.PrimaryCamera().View(), host.PrimaryCamera().Projection(), host.PrimaryCamera().Viewport(),
		vertexSnapPickRadiusPixels)
}

func (t *VertexSnapTool) beginDrag(host *engine.Host, source vertexSnapCandidate) {
	t.dragging = true
	t.source = source
	t.sourceStartWorld = source.World
	t.transformManager.setupMemento()
	t.updateDrag(host)
}

func (t *VertexSnapTool) updateDrag(host *engine.Host) {
	cam := host.PrimaryCamera()
	mouse := host.Window.Cursor.Position()
	delta := matrix.Vec3Zero()
	target, hasTarget := t.pickTarget(host, mouse)
	if hasTarget {
		delta = target.World.Subtract(t.sourceStartWorld)
		t.targetMarker.Show(target.World, vertexSnapMarkerWorldScale(cam, target.World))
	} else {
		t.targetMarker.Hide()
		if hit, ok := cam.ForwardPlaneHit(mouse, t.sourceStartWorld); ok {
			delta = hit.Subtract(t.sourceStartWorld)
		}
	}
	applyVertexSnapDelta(t.transformManager.memento, delta)
	sourceWorld := t.sourceStartWorld.Add(delta)
	t.sourceMarker.Show(sourceWorld, vertexSnapMarkerWorldScale(cam, sourceWorld))
}

func (t *VertexSnapTool) pickTarget(host *engine.Host, mouse matrix.Vec2) (vertexSnapCandidate, bool) {
	manager := t.transformManager.manager
	selection := manager.HierarchyRespectiveSelection()
	targets := manager.VertexSnapTargetEntities(selection)
	cam := host.PrimaryCamera()
	return closestSnapVertexOnEntities(targets, mouse, cam.View(), cam.Projection(),
		cam.Viewport(), vertexSnapPickRadiusPixels)
}

func (t *VertexSnapTool) commitDrag(host *engine.Host) {
	t.updateDrag(host)
	t.dragging = false
	if t.transformManager.memento != nil && len(t.transformManager.memento.entities) > 0 {
		t.transformManager.history.Add(t.transformManager.memento)
		t.transformManager.manager.RefitBVH(t.transformManager.memento.entities[0])
	}
	t.transformManager.memento = nil
	t.targetMarker.Hide()
}

func (t *VertexSnapTool) cancelDrag() {
	t.dragging = false
	memento := t.transformManager.memento
	if memento != nil {
		for i, e := range memento.entities {
			e.Transform.SetWorldPosition(memento.from[i].position)
			e.Transform.SetWorldRotation(memento.from[i].rotation)
			e.Transform.SetWorldScale(memento.from[i].scale)
		}
		if len(memento.entities) > 0 {
			t.transformManager.manager.RefitBVH(memento.entities[0])
		}
	}
	t.transformManager.memento = nil
	t.Hide()
}

func (m *vertexSnapMarker) Initialize(host *engine.Host, color matrix.Color) {
	m.transform.Initialize(host.WorkGroup())
	mesh := rendering.NewMeshSphere(host.MeshCache(), 1, 8, 8)
	mat, err := host.MaterialCache().Material("gizmo_overlay.material")
	if err != nil {
		slog.Error("failed to load vertex snap marker material", "error", err)
		return
	}
	m.shaderData = shader_data_registry.Create("unlit")
	m.shaderData.(*shader_data_registry.ShaderDataUnlit).Color = color
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: m.shaderData,
		Transform:  &m.transform,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
	m.Hide()
}

func (m *vertexSnapMarker) Show(pos matrix.Vec3, scale matrix.Float) {
	if m.shaderData == nil {
		return
	}
	m.transform.SetPosition(pos)
	m.transform.SetScale(matrix.NewVec3XYZ(scale))
	m.shaderData.Activate()
}

func (m *vertexSnapMarker) Hide() {
	if m.shaderData != nil {
		m.shaderData.Deactivate()
	}
}

func vertexSnapMarkerWorldScale(cam cameras.Camera, pos matrix.Vec3) matrix.Float {
	if cam == nil {
		return vertexSnapMarkerScale
	}
	if cam.IsOrthographic() {
		return matrix.Float(matrix.Max(cam.Width(), cam.Height())) * vertexSnapMarkerScale
	}
	viewPos := matrix.Mat4MultiplyVec4(cam.View(), pos.AsVec4())
	dist := matrix.Abs(viewPos.Z())
	if dist <= matrix.FloatSmallestNonzero {
		return vertexSnapMarkerScale
	}
	return dist * vertexSnapMarkerScale
}

func closestSnapVertexOnEntities(entities []*editor_stage_manager.StageEntity, mouse matrix.Vec2, view, projection matrix.Mat4, viewport matrix.Vec4, radius matrix.Float) (vertexSnapCandidate, bool) {
	best := vertexSnapCandidate{}
	found := false
	bestDistSq := radius * radius
	for _, e := range entities {
		candidate, ok := closestSnapVertexOnEntity(e, mouse, view, projection, viewport, radius)
		if !ok || candidate.ScreenDistSq > bestDistSq {
			continue
		}
		best = candidate
		bestDistSq = candidate.ScreenDistSq
		found = true
	}
	return best, found
}

func closestSnapVertexOnEntity(e *editor_stage_manager.StageEntity, mouse matrix.Vec2, view, projection matrix.Mat4, viewport matrix.Vec4, radius matrix.Float) (vertexSnapCandidate, bool) {
	if e == nil || e.IsDeleted() || len(e.StageData.SnapVertices) == 0 {
		return vertexSnapCandidate{}, false
	}
	best := vertexSnapCandidate{}
	found := false
	bestDistSq := radius * radius
	world := e.Transform.WorldMatrix()
	for _, local := range e.StageData.SnapVertices {
		worldPos := world.TransformPoint(local)
		screen, ok := matrix.Mat4ToScreenSpace(worldPos, view, projection, viewport)
		if !ok {
			continue
		}
		screen = vertexSnapToCursorScreenSpace(screen, viewport)
		dx := screen.X() - mouse.X()
		dy := screen.Y() - mouse.Y()
		distSq := dx*dx + dy*dy
		if distSq > bestDistSq {
			continue
		}
		best = vertexSnapCandidate{
			Entity:       e,
			Local:        local,
			World:        worldPos,
			Screen:       screen,
			ScreenDistSq: distSq,
		}
		bestDistSq = distSq
		found = true
	}
	return best, found
}

func vertexSnapToCursorScreenSpace(screen matrix.Vec3, viewport matrix.Vec4) matrix.Vec3 {
	screen.SetY(viewport.Y() + viewport.W() - (screen.Y() - viewport.Y()))
	return screen
}

func applyVertexSnapDelta(memento *transformHistory, delta matrix.Vec3) {
	if memento == nil {
		return
	}
	for i, e := range memento.entities {
		pos := memento.from[i].position.Add(delta)
		memento.to[i].position = pos
		e.Transform.SetWorldPosition(pos)
	}
}
