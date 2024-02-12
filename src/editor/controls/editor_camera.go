/*****************************************************************************/
/* editor_camera.go                                                          */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

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
