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
	"kaijuengine.com/registry/shader_data_registry"
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

func TestSelectionOutlineGroupsJoinHierarchiesAndSeparateObjects(t *testing.T) {
	t.Parallel()

	parent := newSelectionOutlineTestEntity()
	child := newSelectionOutlineTestEntity()
	child.SetParent(&parent.Entity)
	other := newSelectionOutlineTestEntity()
	manager := StageManager{}

	// Select the child first so the parent selection has to regroup it
	manager.SelectEntity(child)
	manager.SelectEntity(parent)
	manager.SelectEntity(other)

	parentGroup := selectionOutlineGroup(parent)
	if parentGroup == 0 {
		t.Fatalf("selected entity should have a nonzero outline group")
	}
	if got := selectionOutlineGroup(child); got != parentGroup {
		t.Fatalf("child outline group = %#x, want parent group %#x", got, parentGroup)
	}
	if got := selectionOutlineGroup(other); got == parentGroup || got == 0 {
		t.Fatalf("separate object outline group = %#x, want a distinct nonzero group", got)
	}

	manager.DeselectEntity(other)
	if shader_data_registry.StandardShaderDataFlagsTest(other.StageData.ShaderData,
		shader_data_registry.ShaderDataStandardFlagOutline) {
		t.Fatalf("deselected entity should not keep the outline flag")
	}
	if got := selectionOutlineGroup(other); got != 0 {
		t.Fatalf("deselected entity outline group = %#x, want 0", got)
	}
}

func newSelectionOutlineTestEntity() *StageEntity {
	entity := &StageEntity{}
	entity.Init(nil)
	entity.StageData.ShaderData = &shader_data_registry.ShaderDataStandard{}
	return entity
}

func selectionOutlineGroup(e *StageEntity) shader_data_registry.StandardShaderDataFlags {
	return e.StageData.ShaderData.(*shader_data_registry.ShaderDataStandard).Flags &
		shader_data_registry.ShaderDataStandardOutlineGroupMask
}

func newSelectionBoundsTestEntity(position matrix.Vec3, localBounds graviton.AABB) *StageEntity {
	entity := &StageEntity{}
	entity.Init(nil)
	entity.StageData.Bvh = graviton.NewBVH([]graviton.HitObject{localBounds}, &entity.Transform, entity)
	entity.Transform.SetPosition(position)
	return entity
}
