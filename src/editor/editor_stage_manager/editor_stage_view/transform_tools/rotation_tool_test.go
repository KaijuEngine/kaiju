/******************************************************************************/
/* rotation_tool_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestRotationGizmoDragRotationPreservesSnapForCommit(t *testing.T) {
	t.Parallel()

	tool := RotationTool{
		TransformGizmo: TransformGizmo{
			currentAxis: matrix.Vz,
			dragging:    true,
		},
		rotationDelta: matrix.Deg2Rad(93.061),
	}
	committed := matrix.Vec4Zero()
	tool.OnDragEnd.Add(func(rotation matrix.Vec4) {
		committed = rotation
	})

	tool.endDrag(true, 15)

	if !matrix.Approx(committed.W(), 90) {
		t.Fatalf("committed drag rotation = %v degrees, want 90", committed.W())
	}
	if committed.AsVec3() != matrix.Vec3Backward() {
		t.Fatalf("committed drag axis = %v, want %v", committed.AsVec3(), matrix.Vec3Backward())
	}
	if tool.dragging {
		t.Fatal("rotation gizmo remained in dragging state after commit")
	}
	if !matrix.Approx(tool.rotationDelta, 0) {
		t.Fatalf("rotation delta after commit = %v, want 0", tool.rotationDelta)
	}
}
