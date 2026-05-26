/******************************************************************************/
/* editor_stage_manager_interact_test.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

func TestSelectionBoundsUsesWorldBVHCenter(t *testing.T) {
	t.Parallel()

	entity := newSelectionBoundsTestEntity(
		matrix.NewVec3(10, 0, -4),
		graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3(1, 1, 1)),
	)
	manager := StageManager{selected: []*StageEntity{entity}}

	got := manager.SelectionBounds()
	wantCenter := matrix.NewVec3(10, 0, -4)
	if !matrix.Vec3ApproxTo(got.Center, wantCenter, 0.0001) {
		t.Fatalf("center = %v, want %v", got.Center, wantCenter)
	}
}

func TestSelectionBoundsPreservesLocalBVHCenterOffset(t *testing.T) {
	t.Parallel()

	entity := newSelectionBoundsTestEntity(
		matrix.NewVec3(10, 0, 0),
		graviton.NewAABB(matrix.NewVec3(2, 0, 0), matrix.NewVec3(1, 1, 1)),
	)
	manager := StageManager{selected: []*StageEntity{entity}}

	got := manager.SelectionBounds()
	wantCenter := matrix.NewVec3(12, 0, 0)
	if !matrix.Vec3ApproxTo(got.Center, wantCenter, 0.0001) {
		t.Fatalf("center = %v, want %v", got.Center, wantCenter)
	}
}

func TestSelectionBoundsUnionsSelectedWorldBounds(t *testing.T) {
	t.Parallel()

	a := newSelectionBoundsTestEntity(
		matrix.NewVec3(10, 0, 0),
		graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3(1, 1, 1)),
	)
	b := newSelectionBoundsTestEntity(
		matrix.NewVec3(20, 0, 0),
		graviton.NewAABB(matrix.Vec3Zero(), matrix.NewVec3(1, 1, 1)),
	)
	manager := StageManager{selected: []*StageEntity{a, b}}

	got := manager.SelectionBounds()
	wantCenter := matrix.NewVec3(15, 0, 0)
	wantExtent := matrix.NewVec3(6, 1, 1)
	if !matrix.Vec3ApproxTo(got.Center, wantCenter, 0.0001) {
		t.Fatalf("center = %v, want %v", got.Center, wantCenter)
	}
	if !matrix.Vec3ApproxTo(got.Extent, wantExtent, 0.0001) {
		t.Fatalf("extent = %v, want %v", got.Extent, wantExtent)
	}
}

func newSelectionBoundsTestEntity(position matrix.Vec3, localBounds graviton.AABB) *StageEntity {
	entity := &StageEntity{}
	entity.Init(nil)
	entity.StageData.Bvh = graviton.NewBVH([]graviton.HitObject{localBounds}, &entity.Transform, entity)
	entity.Transform.SetPosition(position)
	return entity
}
