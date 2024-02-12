package controls

import (
	"kaiju/cameras"
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/matrix"
	"math"
)

const (
	ROT_SCALE  = 0.01
	ZOOM_SCALE = float32(0.05)
)

type EditorCamera struct {
	lastMousePos  matrix.Vec2
	mouseDown     matrix.Vec2
	lastHit       matrix.Vec3
	yawScale      matrix.Float
	mouseLeftDown bool
}

func (e *EditorCamera) pan(tc *cameras.TurntableCamera, mp matrix.Vec2) {
	if hitPoint, ok := tc.ForwardPlaneHit(mp, tc.Center()); ok {
		if matrix.Vec3Approx(e.lastHit, matrix.Vec3Zero()) {
			e.lastHit = hitPoint
		}
		delta := hitPoint.Subtract(e.lastHit)
		tc.SetLookAt(tc.Center().Add(delta))
		e.lastHit, _ = tc.ForwardPlaneHit(mp, tc.Center())
	}
}

func (e *EditorCamera) Update(host *engine.Host, delta float64) {
	tc := host.Camera.(*cameras.TurntableCamera)
	mouse := &host.Window.Mouse
	kb := &host.Window.Keyboard
	mp := mouse.Position()
	if mouse.Pressed(hid.MouseButtonLeft) || mouse.Pressed(hid.MouseButtonMiddle) {
		e.mouseLeftDown = true
		e.mouseDown = mp
		rg := int(math.Abs(float64(int(matrix.Rad2Deg(tc.Pitch())) % 360)))
		if rg < 90 || rg > 270 {
			e.yawScale = ROT_SCALE
		} else {
			e.yawScale = -ROT_SCALE
		}
	} else if mouse.Held(hid.MouseButtonLeft) {
		if kb.KeyHeld(hid.KeyboardKeyLeftAlt) {
			x := (e.lastMousePos.Y() - mp.Y()) * -ROT_SCALE
			y := (e.lastMousePos.X() - mp.X()) * e.yawScale
			tc.Orbit(matrix.Vec3{x, y, 0.0})
		} else if kb.KeyHeld(hid.KeyboardKeySpace) {
			e.pan(tc, mp)
		}
	} else if mouse.Held(hid.MouseButtonMiddle) {
		e.pan(tc, mp)
	} else if mouse.Released(hid.MouseButtonLeft) || mouse.Released(hid.MouseButtonMiddle) {
		e.lastHit = matrix.Vec3Zero()
	}
	if mouse.Scrolled() {
		zoom := tc.Zoom()
		scale := -ZOOM_SCALE
		if zoom < 1.0 {
			scale *= zoom / 1.0
		}
		tc.Dolly(mouse.Scroll().Y() * scale)
	}
	e.lastMousePos = mp
}
