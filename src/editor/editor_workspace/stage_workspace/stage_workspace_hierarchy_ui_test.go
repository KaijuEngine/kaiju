/******************************************************************************/
/* stage_workspace_hierarchy_ui_test.go                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestViewAlignedRotationMatchesCameraAxes(t *testing.T) {
	view := matrix.Mat4Identity()
	view.Rotate(matrix.NewVec3(20, 35, 10))
	up := view.Up().Normal()
	forward := view.Forward().Negative().Normal()

	var transform matrix.Transform
	transform.SetupRawTransform()
	transform.SetRotation(viewAlignedRotation(up, forward))

	const tolerance = matrix.Float(0.0001)
	if !matrix.Vec3ApproxTo(transform.Up(), up, tolerance) {
		t.Fatalf("expected up axis %v, got %v", up, transform.Up())
	}
	if !matrix.Vec3ApproxTo(transform.Forward(), forward, tolerance) {
		t.Fatalf("expected forward axis %v, got %v", forward, transform.Forward())
	}
	rightHandedForward := matrix.Vec3Cross(transform.Right(), transform.Up()).Normal()
	if !matrix.Vec3ApproxTo(rightHandedForward, transform.Forward(), tolerance) {
		t.Fatalf("expected right-handed transform basis, got %v from right/up and %v forward", rightHandedForward, transform.Forward())
	}
}
