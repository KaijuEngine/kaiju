/******************************************************************************/
/* editor_camera.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor_controls

import (
	"kaiju/engine"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/engine/systems/events"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"math"
)

const (
	ROT_SCALE     = 0.005
	ZOOM_SCALE_3D = float32(0.05)
	ZOOM_SCALE_2D = float32(1.0)
)

type EditorCameraMode = int

const (
	EditorCameraModeNone = EditorCameraMode(iota)
	EditorCameraMode3d
	EditorCameraMode2d
)

type EditorCamera struct {
	OnModeChange events.Event
	camera       cameras.Camera
	lastMousePos matrix.Vec2
	mouseDown    matrix.Vec2
	lastHit      matrix.Vec3
	yawScale     matrix.Float
	dragging     bool
	mode         EditorCameraMode
	resizeId     events.Id
}

func (e *EditorCamera) Mode() EditorCameraMode { return e.mode }

func (e *EditorCamera) LookAtPoint() matrix.Vec3 {
	defer tracing.NewRegion("EditorCamera.LookAtPoint").End()
	switch e.mode {
	case EditorCameraMode3d:
		// Something is backwards about the lookat for the turntable camera...
		return e.camera.LookAt().Negative()
	default:
		return e.camera.LookAt()
	}
}

func (e *EditorCamera) SetMode(mode EditorCameraMode, host *engine.Host) {
	defer tracing.NewRegion("EditorCamera.SetMode").End()
	if e.mode == mode {
		return
	}
	e.mode = mode
	switch e.mode {
	case EditorCameraMode3d:
		w := float32(host.Window.Width())
		h := float32(host.Window.Height())
		cam := cameras.NewStandardCamera(w, h, w, h, matrix.Vec3Backward())
		tc := cameras.ToTurntable(cam)
		tc.SetYawPitchZoom(0, -25, 16)
		tc.SetLookAt(matrix.Vec3Zero())
		tc.SetZoom(15)
		e.camera = tc
		host.Camera = e.camera
	case EditorCameraMode2d:
		cw := host.Camera.Width()
		ch := host.Camera.Height()
		ratio := cw / ch
		w := (cw / cw) * ratio * 10
		h := (ch / cw) * ratio * 10
		oc := cameras.NewStandardCameraOrthographic(w, h, cw, ch, matrix.NewVec3(0, 0, 100))
		e.camera = oc
		host.Camera = e.camera
		host.Window.OnResize.Remove(e.resizeId)
		e.resizeId = host.Window.OnResize.Add(e.OnWindowResize)
	}
	e.OnModeChange.Execute()
}

func (e *EditorCamera) OnWindowResize() {
	defer tracing.NewRegion("EditorCamera.OnWindowResize").End()
	klib.NotYetImplemented(309)
}

func (e *EditorCamera) Update(host *engine.Host, delta float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.Update").End()
	switch e.mode {
	case EditorCameraMode3d:
		return e.update3d(host, delta)
	case EditorCameraMode2d:
		return e.update2d(host, delta)
	case EditorCameraModeNone:
		fallthrough
	default:
		return false
	}
}

func (e *EditorCamera) RayCast(mouse *hid.Mouse) collision.Ray {
	defer tracing.NewRegion("EditorCamera.RayCast").End()
	return e.camera.RayCast(mouse.Position())
}

func (e *EditorCamera) Focus(bounds collision.AABB) {
	defer tracing.NewRegion("EditorCamera.Focus").End()
	z := bounds.Extent.Length()
	if z <= 0.01 {
		z = 5
	} else {
		z *= 2
	}
	if e.camera.IsOrthographic() {
		c := e.camera.(*cameras.StandardCamera)
		p := c.Position()
		p.SetX(bounds.Center.X())
		p.SetY(bounds.Center.Y())
		c.SetPositionAndLookAt(p, bounds.Center.Negative())
		r := c.Width() / c.Height()
		if c.Width() > c.Height() {
			c.Resize(z*r, z)
		} else {
			c.Resize(z, z*r)
		}
	} else {
		c := e.camera.(*cameras.TurntableCamera)
		c.SetLookAt(bounds.Center.Negative())
		c.SetZoom(z)
	}
}

func (e *EditorCamera) pan3d(tc *cameras.TurntableCamera, mp matrix.Vec2) {
	defer tracing.NewRegion("EditorCamera.pan3d").End()
	if hitPoint, ok := tc.ForwardPlaneHit(mp, tc.LookAt()); ok {
		if matrix.Vec3Approx(e.lastHit, matrix.Vec3Zero()) {
			e.lastHit = hitPoint
		}
		delta := hitPoint.Subtract(e.lastHit)
		tc.SetLookAt(tc.LookAt().Add(delta))
		e.lastHit, _ = tc.ForwardPlaneHit(mp, tc.LookAt())
	}
}

func (e *EditorCamera) pan2d(oc *cameras.StandardCamera, mp matrix.Vec2, host *engine.Host) {
	defer tracing.NewRegion("EditorCamera.pan2d").End()
	hitPoint := matrix.NewVec3(mp.X(), mp.Y(), 0)
	if matrix.Vec3Approx(e.lastHit, matrix.Vec3Zero()) {
		e.lastHit = hitPoint
	}
	cw := oc.Width() / float32(host.Window.Width())
	ch := oc.Height() / float32(host.Window.Height())
	delta := e.lastHit.Subtract(hitPoint).Multiply(matrix.NewVec3(cw, ch, 0))
	oc.SetPositionAndLookAt(oc.Position().Add(delta), oc.LookAt().Add(delta))
	e.lastHit = hitPoint.Add(delta)
}

func (e *EditorCamera) update3d(host *engine.Host, _ float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.update3d").End()
	tc := e.camera.(*cameras.TurntableCamera)
	mouse := &host.Window.Mouse
	kb := &host.Window.Keyboard
	mp := mouse.Position()
	if kb.HasAlt() || kb.KeyHeld(hid.KeyboardKeySpace) {
		changed = true
	}
	if mouse.Pressed(hid.MouseButtonLeft) || mouse.Pressed(hid.MouseButtonMiddle) {
		e.dragging = true
		e.mouseDown = mp
		rg := int(math.Abs(float64(int(matrix.Rad2Deg(tc.Pitch())) % 360)))
		if rg < 90 || rg > 270 {
			e.yawScale = ROT_SCALE
		} else {
			e.yawScale = -ROT_SCALE
		}
		if mouse.Pressed(hid.MouseButtonMiddle) {
			changed = true
		}
	} else if e.dragging && mouse.Held(hid.MouseButtonLeft) {
		if kb.HasAlt() {
			x := (e.lastMousePos.Y() - mp.Y()) * -ROT_SCALE
			y := (e.lastMousePos.X() - mp.X()) * e.yawScale
			tc.Orbit(matrix.Vec3{x, y, 0.0})
			changed = true
		} else if kb.KeyHeld(hid.KeyboardKeySpace) {
			e.pan3d(tc, mp)
			changed = true
		}
	} else if e.dragging && mouse.Held(hid.MouseButtonMiddle) {
		e.pan3d(tc, mp)
		changed = true
	} else if mouse.Released(hid.MouseButtonLeft) || mouse.Released(hid.MouseButtonMiddle) {
		e.lastHit = matrix.Vec3Zero()
		if mouse.Released(hid.MouseButtonMiddle) {
			changed = true
		}
		e.dragging = false
	}
	if mouse.Scrolled() {
		zoom := tc.Zoom()
		scale := -ZOOM_SCALE_3D
		if zoom < 1.0 {
			scale *= zoom / 1.0
		}
		tc.Dolly(mouse.Scroll().Y() * scale)
		changed = true
	}
	e.lastMousePos = mp
	return changed
}

func (e *EditorCamera) update2d(host *engine.Host, _ float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.update2d").End()
	oc := e.camera.(*cameras.StandardCamera)
	mouse := &host.Window.Mouse
	kb := &host.Window.Keyboard
	mp := mouse.Position()
	if mouse.Pressed(hid.MouseButtonMiddle) {
		e.dragging = true
		e.mouseDown = mp
		if mouse.Pressed(hid.MouseButtonMiddle) {
			changed = true
		}
	} else if e.dragging && mouse.Held(hid.MouseButtonMiddle) {
		e.pan2d(oc, mp, host)
		changed = true
	} else if mouse.Released(hid.MouseButtonMiddle) {
		e.lastHit = matrix.Vec3Zero()
		if mouse.Released(hid.MouseButtonMiddle) {
			changed = true
		}
		e.dragging = false
	} else if kb.KeyHeld(hid.KeyboardKeySpace) {
		e.pan2d(oc, mp, host)
		changed = true
	}
	if mouse.Scrolled() {
		cw := host.Camera.Width()
		ch := host.Camera.Height()
		w := oc.Width()
		h := oc.Height()
		r := cw / ch
		w += (cw / cw) * r * -ZOOM_SCALE_2D * mouse.Scroll().Y()
		h += (ch / cw) * r * -ZOOM_SCALE_2D * mouse.Scroll().Y()
		if w > matrix.FloatSmallestNonzero && h > matrix.FloatSmallestNonzero {
			oc.Resize(w, h)
			changed = true
		}
	}
	e.lastMousePos = mp
	return changed
}
