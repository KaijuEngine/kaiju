/******************************************************************************/
/* vertex_snap_tool_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/matrix"
)

func TestClosestSnapVertexOnEntitySelectsNearestProjectedVertex(t *testing.T) {
	entity := newVertexSnapToolTestEntity(matrix.Vec3Zero(),
		matrix.NewVec3(0, 0, 0),
		matrix.NewVec3(0.1, 0, 0),
		matrix.NewVec3(-0.4, 0, 0))
	got, ok := closestSnapVertexOnEntity(entity, matrix.NewVec2(54, 50),
		matrix.Mat4Identity(), matrix.Mat4Identity(), matrix.NewVec4(0, 0, 100, 100), 12)
	if !ok {
		t.Fatal("expected a source vertex candidate")
	}
	if got.Local != matrix.NewVec3(0.1, 0, 0) {
		t.Fatalf("closest source vertex = %v; want %v", got.Local, matrix.NewVec3(0.1, 0, 0))
	}
}

func TestClosestSnapVertexOnEntityComparesAgainstCursorScreenY(t *testing.T) {
	entity := newVertexSnapToolTestEntity(matrix.Vec3Zero(),
		matrix.NewVec3(0, 0.2, 0),
		matrix.NewVec3(0, -0.2, 0))
	got, ok := closestSnapVertexOnEntity(entity, matrix.NewVec2(50, 40),
		matrix.Mat4Identity(), matrix.Mat4Identity(), matrix.NewVec4(0, 0, 100, 100), 12)
	if !ok {
		t.Fatal("expected a source vertex candidate")
	}
	if got.Local != matrix.NewVec3(0, 0.2, 0) {
		t.Fatalf("closest source vertex = %v; want %v", got.Local, matrix.NewVec3(0, 0.2, 0))
	}
}

func TestClosestSnapVertexOnEntityRejectsOutOfRadiusAndEmptyMeshes(t *testing.T) {
	entity := newVertexSnapToolTestEntity(matrix.Vec3Zero(), matrix.Vec3Zero())
	if _, ok := closestSnapVertexOnEntity(entity, matrix.NewVec2(90, 90),
		matrix.Mat4Identity(), matrix.Mat4Identity(), matrix.NewVec4(0, 0, 100, 100), 12); ok {
		t.Fatal("expected out-of-radius vertex to be ignored")
	}
	empty := newVertexSnapToolTestEntity(matrix.Vec3Zero())
	if _, ok := closestSnapVertexOnEntity(empty, matrix.NewVec2(50, 50),
		matrix.Mat4Identity(), matrix.Mat4Identity(), matrix.NewVec4(0, 0, 100, 100), 12); ok {
		t.Fatal("expected entity without snap vertices to be ignored")
	}
}

func TestClosestSnapVertexOnEntitiesSelectsNearestTarget(t *testing.T) {
	far := newVertexSnapToolTestEntity(matrix.Vec3Zero(), matrix.NewVec3(-0.2, 0, 0))
	near := newVertexSnapToolTestEntity(matrix.Vec3Zero(), matrix.NewVec3(0.08, 0, 0))
	got, ok := closestSnapVertexOnEntities([]*editor_stage_manager.StageEntity{far, near}, matrix.NewVec2(55, 50),
		matrix.Mat4Identity(), matrix.Mat4Identity(), matrix.NewVec4(0, 0, 100, 100), 12)
	if !ok {
		t.Fatal("expected a target vertex candidate")
	}
	if got.Entity != near {
		t.Fatalf("closest target entity = %p; want %p", got.Entity, near)
	}
}

func TestApplyVertexSnapDeltaMovesSelectionByWorldDelta(t *testing.T) {
	a := newVertexSnapToolTestEntity(matrix.NewVec3(1, 2, 3), matrix.Vec3Zero())
	b := newVertexSnapToolTestEntity(matrix.NewVec3(-2, 0, 5), matrix.Vec3Zero())
	memento := &transformHistory{
		entities: []*editor_stage_manager.StageEntity{a, b},
		from: []transformHistoryPRS{
			{position: a.Transform.WorldPosition(), rotation: a.Transform.WorldRotation(), scale: a.Transform.WorldScale()},
			{position: b.Transform.WorldPosition(), rotation: b.Transform.WorldRotation(), scale: b.Transform.WorldScale()},
		},
		to: make([]transformHistoryPRS, 2),
	}
	delta := matrix.NewVec3(3, -1, 2)
	applyVertexSnapDelta(memento, delta)
	if got := a.Transform.WorldPosition(); got != matrix.NewVec3(4, 1, 5) {
		t.Fatalf("first entity position = %v; want %v", got, matrix.NewVec3(4, 1, 5))
	}
	if got := b.Transform.WorldPosition(); got != matrix.NewVec3(1, -1, 7) {
		t.Fatalf("second entity position = %v; want %v", got, matrix.NewVec3(1, -1, 7))
	}
	if memento.to[0].position != matrix.NewVec3(4, 1, 5) || memento.to[1].position != matrix.NewVec3(1, -1, 7) {
		t.Fatalf("memento targets not updated: %v", memento.to)
	}
}

func newVertexSnapToolTestEntity(position matrix.Vec3, vertices ...matrix.Vec3) *editor_stage_manager.StageEntity {
	e := &editor_stage_manager.StageEntity{}
	e.Init(nil)
	e.Transform.SetPosition(position)
	e.StageData.SnapVertices = vertices
	return e
}
