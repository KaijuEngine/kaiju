/******************************************************************************/
/* turntable_camera_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package cameras

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestTurntableCameraFlyRotateMarksViewChanged(t *testing.T) {
	camera := ToTurntable(NewStandardCamera(1280, 720, 1280, 720, matrix.Vec3Backward()))
	camera.NewFrame()
	camera.csmDirty = false

	camera.FlyRotate(10, 5)

	if !camera.IsDirty() {
		t.Fatalf("FlyRotate should mark the camera dirty so view culling is recalculated")
	}
	if !camera.csmDirty {
		t.Fatalf("FlyRotate should dirty CSM projections because the view changed")
	}
}
