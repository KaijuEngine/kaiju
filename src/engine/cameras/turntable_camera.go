/******************************************************************************/
/* turntable_camera.go                                                        */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package cameras

import (
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
)

type TurntableCamera struct {
	StandardCamera
	pitch float32
	yaw   float32
	zoom  float32
}

// ToTurntable converts a standard camera to a turntable camera.
func ToTurntable(camera *StandardCamera) *TurntableCamera {
	defer tracing.NewRegion("cameras.ToTurntable").End()
	tc := &TurntableCamera{
		StandardCamera: *camera,
		yaw:            0.0,
		pitch:          0.0,
	}
	tc.updateView = tc.internalUpdateView
	tc.updateProjection = tc.internalUpdateProjection
	return tc
}

// Yaw returns the yaw of the camera.
func (c *TurntableCamera) Yaw() float32 { return c.yaw }

// Pitch returns the pitch of the camera.
func (c *TurntableCamera) Pitch() float32 { return c.pitch }

// Zoom returns the zoom of the camera.
func (c *TurntableCamera) Zoom() float32 { return c.zoom }

// SetPosition sets the position of the camera.
func (c *TurntableCamera) SetPosition(position matrix.Vec3) {
	defer tracing.NewRegion("TurntableCamera.SetPosition").End()
	c.position = position
	c.zoom = position.Z()
	c.updateViewAndPosition()
}

// SetLookAt sets the look at position of the camera.
func (c *TurntableCamera) SetLookAt(lookAt matrix.Vec3) {
	defer tracing.NewRegion("TurntableCamera.SetLookAt").End()
	c.lookAt = lookAt
	c.updateViewAndPosition()
}

// SetLookAtWithUp sets the look at position of the camera and the up vector to use.
func (c *TurntableCamera) SetLookAtWithUp(point, up matrix.Vec3) {
	defer tracing.NewRegion("TurntableCamera.SetLookAtWithUp").End()
	c.lookAt = point
	c.up = up
	c.updateViewAndPosition()
}

// Pan pans the camera while keeping the same facing by the given delta.
func (c *TurntableCamera) Pan(delta matrix.Vec3) {
	defer tracing.NewRegion("TurntableCamera.Pan").End()
	d := delta.Scale(c.zoom)
	u := c.Up()
	u.ScaleAssign(-d.Y())
	r := c.Right()
	r.ScaleAssign(-d.X())
	c.lookAt.AddAssign(u)
	c.lookAt.AddAssign(r)
	c.updateViewAndPosition()
}

// Dolly moves the camera closer/further from the look at point by the given delta.
func (c *TurntableCamera) Dolly(delta float32) {
	defer tracing.NewRegion("TurntableCamera.Dolly").End()
	zoom := c.zoom
	diff := c.position.Subtract(c.lookAt)
	length := diff.Length()
	zoom += delta * length
	if c.position.Z() <= 0.0 {
		zoom += 0.001
	}
	c.SetZoom(zoom)
}

// Orbit orbits the camera around the look at point by the given delta.
func (c *TurntableCamera) Orbit(delta matrix.Vec3) {
	defer tracing.NewRegion("TurntableCamera.Orbit").End()
	if delta.Equals(matrix.Vec3Zero()) {
		return
	}
	c.pitch += delta.X()
	c.yaw += delta.Y()
	c.updateViewAndPosition()
}

func (c *TurntableCamera) FlyRotate(yawDelta, pitchDelta float32) {
	defer tracing.NewRegion("TurntableCamera.FlyRotate").End()
	c.yaw += matrix.Deg2Rad(yawDelta)
	c.pitch += matrix.Deg2Rad(pitchDelta)
	const maxPitch = 89.0 * matrix.DegToRadVal
	const minPitch = -89.0 * matrix.DegToRadVal
	if c.pitch > maxPitch {
		c.pitch = maxPitch
	}
	if c.pitch < minPitch {
		c.pitch = minPitch
	}
	c.FlyUpdateView()
}

func (c *TurntableCamera) FlyUpdateView() {
	defer tracing.NewRegion("TurntableCamera.FlyUpdateView").End()
	direction := matrix.Vec3{
		-matrix.Sin(c.yaw) * matrix.Cos(c.pitch),
		matrix.Sin(c.pitch),
		-matrix.Cos(c.yaw) * matrix.Cos(c.pitch),
	}
	direction.Normalize()
	c.lookAt = c.position.Add(direction.Scale(c.zoom))
	c.view = matrix.Mat4LookAt(c.position, c.lookAt, c.up)
	c.iView = c.view
	c.iView.Inverse()
	c.updateFrustum()
}

// SetYaw sets the yaw of the camera.
func (c *TurntableCamera) SetYaw(yaw float32) {
	defer tracing.NewRegion("TurntableCamera.SetYaw").End()
	c.setYaw(yaw)
	c.updateViewAndPosition()
}

// SetPitch sets the pitch of the camera.
func (c *TurntableCamera) SetPitch(pitch float32) {
	defer tracing.NewRegion("TurntableCamera.SetPitch").End()
	c.setPitch(pitch)
	c.updateViewAndPosition()
}

// SetZoom sets the zoom of the camera.
func (c *TurntableCamera) SetZoom(zoom float32) {
	defer tracing.NewRegion("TurntableCamera.SetZoom").End()
	c.setZoom(zoom)
	c.updateViewAndPosition()
}

// SetYawAndPitch sets the yaw and pitch of the camera. This helps skip
// needless view matrix calculations by setting both before updating the view.
func (c *TurntableCamera) SetYawAndPitch(yaw, pitch float32) {
	defer tracing.NewRegion("TurntableCamera.SetYawAndPitch").End()
	c.setYaw(yaw)
	c.setPitch(pitch)
	c.updateViewAndPosition()
}

// SetYawPitchZoom sets the yaw, pitch, and zoom of the camera. This helps skip
// needless view matrix calculations by setting all three before updating the view.
func (c *TurntableCamera) SetYawPitchZoom(yaw, pitch, zoom float32) {
	defer tracing.NewRegion("TurntableCamera.SetYawPitchZoom").End()
	c.setYaw(yaw)
	c.setPitch(pitch)
	c.setZoom(zoom)
	c.updateViewAndPosition()
}

// RayCast will project a ray from the camera's position given a screen position
// using the camera's view and projection matrices.
func (c *TurntableCamera) RayCast(cursorPosition matrix.Vec2) collision.Ray {
	defer tracing.NewRegion("TurntableCamera.RayCast").End()
	return c.internalRayCast(cursorPosition, c.iView.ExtractPosition())
}

func (c *TurntableCamera) internalUpdateView() {
	defer tracing.NewRegion("TurntableCamera.internalUpdateView").End()
	c.view = matrix.Mat4Identity()
	tx := c.lookAt.X()
	ty := c.lookAt.Y()
	tz := c.lookAt.Z()
	rx := c.pitch
	ry := c.yaw
	rz := float32(0.0)
	di := c.zoom
	a := rx * float32(0.5)
	b := ry * float32(0.5)
	cc := rz * float32(0.5)
	d := matrix.Cos(a)
	e := matrix.Sin(a)
	f := matrix.Cos(b)
	g := matrix.Sin(b)
	h := matrix.Cos(cc)
	i := matrix.Sin(cc)
	j := f*e*h + g*d*i
	k := f*-e*i + g*d*h
	l := f*d*i - g*e*h
	m := f*d*h - g*-e*i
	n := j * j
	o := k * k
	p := l * l
	q := m * m
	r := j * k
	s := k * l
	t := j * l
	u := m * j
	v := m * k
	w := m * l
	x := q + n - o - p
	y := (r + w) * 2.0
	z := (t - v) * 2.0
	A := (r - w) * 2.0
	B := q - n + o - p
	C := (s + u) * 2.0
	D := (t + v) * 2.0
	E := (s - u) * 2.0
	F := q - n - o + p
	G := di
	H := -(tx + D*G)
	I := -(ty + E*G)
	J := -(tz + F*G)
	c.view[0] = x
	c.view[1] = A
	c.view[2] = D
	c.view[3] = 0.0
	c.view[4] = y
	c.view[5] = B
	c.view[6] = E
	c.view[7] = 0.0
	c.view[8] = z
	c.view[9] = C
	c.view[10] = F
	c.view[11] = 0.0
	c.view[12] = x*H + y*I + z*J
	c.view[13] = A*H + B*I + C*J
	c.view[14] = D*H + E*I + F*J
	c.view[15] = 1.0
	c.iView = c.view
	c.iView.Inverse()
	c.updateFrustum()
}

func (c *TurntableCamera) updateViewAndPosition() {
	defer tracing.NewRegion("TurntableCamera.updateViewAndPosition").End()
	c.position.SetZ(c.zoom)
	c.callUpdateView()
	c.position = c.iView.ExtractPosition()
}

func (c *TurntableCamera) setYaw(yaw float32) {
	defer tracing.NewRegion("TurntableCamera.setYaw").End()
	c.yaw = matrix.Deg2Rad(yaw)
	direction := matrix.Vec3{
		matrix.Cos(c.yaw) * matrix.Cos(c.pitch),
		matrix.Sin(c.pitch),
		matrix.Sin(c.yaw) * matrix.Cos(c.pitch),
	}
	direction.Normalize()
	c.lookAt = c.position.Add(direction)
}

func (c *TurntableCamera) setPitch(pitch float32) {
	defer tracing.NewRegion("TurntableCamera.setPitch").End()
	c.pitch = matrix.Deg2Rad(pitch)
	direction := matrix.Vec3{
		matrix.Cos(c.yaw) * matrix.Cos(c.pitch),
		matrix.Sin(c.pitch),
		matrix.Sin(c.yaw) * matrix.Cos(c.pitch),
	}
	direction.Normalize()
	c.lookAt = c.position.Add(direction)
}

func (c *TurntableCamera) setZoom(zoom float32) {
	c.zoom = zoom
}
