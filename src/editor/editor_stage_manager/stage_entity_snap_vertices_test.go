/******************************************************************************/
/* stage_entity_snap_vertices_test.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestSnapVerticesFromMeshDedupesExactPositions(t *testing.T) {
	verts := []rendering.Vertex{
		{Position: matrix.NewVec3(1, 2, 3)},
		{Position: matrix.NewVec3(1, 2, 3)},
		{Position: matrix.NewVec3(1, 2, 3.0001)},
		{Position: matrix.NewVec3(-1, 0, 2)},
		{Position: matrix.NewVec3(-1, 0, 2)},
	}
	got := snapVerticesFromMesh(verts)
	want := []matrix.Vec3{
		matrix.NewVec3(1, 2, 3),
		matrix.NewVec3(1, 2, 3.0001),
		matrix.NewVec3(-1, 0, 2),
	}
	if len(got) != len(want) {
		t.Fatalf("snapVerticesFromMesh returned %d vertices, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("snapVerticesFromMesh[%d] = %v; want %v", i, got[i], want[i])
		}
	}
}

func TestVertexSnapTargetEntitiesFiltersIneligibleEntities(t *testing.T) {
	selected := newVertexSnapTargetTestEntity(true, matrix.Vec3Zero())
	selectedChild := newVertexSnapTargetTestEntity(true, matrix.NewVec3(1, 0, 0))
	selectedChild.SetParent(&selected.Entity)
	locked := newVertexSnapTargetTestEntity(true, matrix.NewVec3(2, 0, 0))
	locked.SetLocked(true)
	deleted := newVertexSnapTargetTestEntity(true, matrix.NewVec3(3, 0, 0))
	deleted.isDeleted = true
	meshless := newVertexSnapTargetTestEntity(false, matrix.NewVec3(4, 0, 0))
	empty := newVertexSnapTargetTestEntity(true, matrix.NewVec3(5, 0, 0))
	empty.StageData.SnapVertices = nil
	target := newVertexSnapTargetTestEntity(true, matrix.NewVec3(6, 0, 0))
	manager := StageManager{
		entities: []*StageEntity{
			selected,
			selectedChild,
			locked,
			deleted,
			meshless,
			empty,
			target,
		},
	}
	got := manager.VertexSnapTargetEntities([]*StageEntity{selected})
	if len(got) != 1 || got[0] != target {
		t.Fatalf("VertexSnapTargetEntities = %v; want only target %p", got, target)
	}
}

func newVertexSnapTargetTestEntity(hasMesh bool, vertex matrix.Vec3) *StageEntity {
	e := &StageEntity{}
	e.Init(nil)
	e.StageData.SnapVertices = []matrix.Vec3{vertex}
	if hasMesh {
		e.StageData.Mesh = &rendering.Mesh{}
	}
	return e
}
