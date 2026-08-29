/******************************************************************************/
/* history_transform_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/matrix"
)

func TestTransformHistoryRestoresChildInWorldSpace(t *testing.T) {
	t.Parallel()

	parent := &editor_stage_manager.StageEntity{}
	child := &editor_stage_manager.StageEntity{}
	parent.Transform.SetupRawTransform()
	child.Transform.SetupRawTransform()
	parent.Transform.SetPosition(matrix.NewVec3(10, 0, 0))
	parent.Transform.SetRotation(matrix.NewVec3(90, 0, 0))
	parent.Transform.SetScale(matrix.NewVec3(2, 2, 2))
	child.SetParent(&parent.Entity)
	child.Transform.SetPosition(matrix.NewVec3(2, 0, 0))
	child.Transform.SetRotation(matrix.Vec3Zero())
	child.Transform.SetScale(matrix.Vec3One())

	original := transformHistoryPRS{
		position: child.Transform.WorldPosition(),
		rotation: child.Transform.WorldRotation(),
		scale:    child.Transform.WorldScale(),
	}
	child.Transform.SetWorldPosition(original.position.Add(matrix.NewVec3(5, 0, 0)))
	history := transformHistory{entities: []*editor_stage_manager.StageEntity{child}}

	history.apply([]transformHistoryPRS{original})

	if !matrix.Vec3Approx(child.Transform.WorldPosition(), original.position) {
		t.Fatalf("restored world position = %v, want %v", child.Transform.WorldPosition(), original.position)
	}
	if !matrix.Vec3Approx(child.Transform.WorldRotation(), original.rotation) {
		t.Fatalf("restored world rotation = %v, want %v", child.Transform.WorldRotation(), original.rotation)
	}
	if !matrix.Vec3Approx(child.Transform.WorldScale(), original.scale) {
		t.Fatalf("restored world scale = %v, want %v", child.Transform.WorldScale(), original.scale)
	}
	if !matrix.Vec3Approx(child.Transform.Rotation(), matrix.Vec3Zero()) {
		t.Fatalf("restored local rotation = %v, want [0 0 0]", child.Transform.Rotation())
	}
}
