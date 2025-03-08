/******************************************************************************/
/* standard_camera.go                                                         */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package cameras

import (
	"kaiju/collision"
	"kaiju/matrix"
)

type StandardCamera struct {
	view             matrix.Mat4
	iView            matrix.Mat4
	projection       matrix.Mat4
	iProjection      matrix.Mat4
	frustum          collision.Frustum
	position         matrix.Vec3
	lookAt           matrix.Vec3
	up               matrix.Vec3
	updateProjection func()
	updateView       func()
	fieldOfView      float32
	nearPlane        float32
	farPlane         float32
	width            float32
	height           float32
	isOrthographic   bool
}

// NewStandardCamera creates a new perspective camera using the width/height
// for the viewport and the position to place the camera.
func NewStandardCamera(width, height float32, position matrix.Vec3) *StandardCamera {
	c := new(StandardCamera)
	c.initializeValues(position)
	c.initialize(width, height)
	return c
}

// NewStandardCameraOrthographic creates a new orthographic camera using the
// width/height for the viewport and the position to place the camera.
func NewStandardCameraOrthographic(width, height float32, position matrix.Vec3) *StandardCamera {
	c := new(StandardCamera)
	c.initializeValues(position)
	c.isOrthographic = true
	c.nearPlane = -1
	c.initialize(width, height)
	return c
}

// SetPosition sets the position of the camera.
func (c *StandardCamera) SetPosition(position matrix.Vec3) {
	c.position = position
	c.updateView()
}

// SetFOV sets the field of view for the camera.
func (c *StandardCamera) SetFOV(fov float32) {
	c.fieldOfView = fov
	c.updateProjection()
}

// SetNearPlane sets the near plane for the camera.
func (c *StandardCamera) SetNearPlane(near float32) {
	c.nearPlane = near
	c.updateProjection()
}

// SetFarPlane sets the far plane for the camera.
func (c *StandardCamera) SetFarPlane(far float32) {
	c.farPlane = far
	c.updateProjection()
}

// SetWidth sets the width of the camera viewport.
func (c *StandardCamera) SetWidth(width float32) {
	c.width = width
	c.updateProjection()
}

// SetHeight sets the height of the camera viewport.
func (c *StandardCamera) SetHeight(height float32) {
	c.height = height
	c.updateProjection()
}

// Resize sets the width and height of the camera viewport.
func (c *StandardCamera) Resize(width, height float32) {
	c.width = width
	c.height = height
	c.updateProjection()
}

// ViewportChanged will update the camera's projection matrix and should only
// be used when there is a change in the viewport. This is typically done
// internally in the system and should not be called by the end-developer.
func (c *StandardCamera) ViewportChanged(width, height float32) {
	c.width = width
	c.height = height
	c.updateProjection()
}

// SetProperties is quick access to set many properties of the camera at once.
// This is typically used for initializing the camera to new values. Calling
// each individual setter for fields would otherwise do needless projection
// matrix updates.
func (c *StandardCamera) SetProperties(fov, nearPlane, farPlane, width, height float32) {
	c.fieldOfView = fov
	c.nearPlane = nearPlane
	c.farPlane = farPlane
	c.width = width
	c.height = height
	c.updateProjection()
}

// Forward returns the forward vector of the camera.
func (c *StandardCamera) Forward() matrix.Vec3 {
	return matrix.Vec3{
		-c.iView[matrix.Mat4x0y2],
		-c.iView[matrix.Mat4x1y2],
		-c.iView[matrix.Mat4x2y2],
	}
}

// Right returns the right vector of the camera.
func (c *StandardCamera) Right() matrix.Vec3 {
	return matrix.Vec3{
		c.iView[matrix.Mat4x0y0],
		c.iView[matrix.Mat4x1y0],
		c.iView[matrix.Mat4x2y0],
	}
}

// Up returns the up vector of the camera.
func (c *StandardCamera) Up() matrix.Vec3 {
	return matrix.Vec3{
		c.iView[matrix.Mat4x0y1],
		c.iView[matrix.Mat4x1y1],
		c.iView[matrix.Mat4x2y1],
	}
}

// SetLookAt sets the look at position of the camera.
func (c *StandardCamera) SetLookAt(position matrix.Vec3) {
	c.lookAt = position
	c.updateView()
}

// SetLookAtWithUp sets the look at position of the camera and the up vector to use.
func (c *StandardCamera) SetLookAtWithUp(point, up matrix.Vec3) {
	c.lookAt = point
	c.up = up
	c.updateView()
}

// SetPositionAndLookAt sets the position and look at position of the camera.
// This is often useful for when the camera warps to another location in space
// and avoids needless view matrix updates when setting the position and look
// at separately.
func (c *StandardCamera) SetPositionAndLookAt(position, lookAt matrix.Vec3) {
	if matrix.Approx(position.Z(), lookAt.Z()) {
		position[matrix.Vz] += 0.0001
	}
	c.position = position
	c.lookAt = lookAt
	c.updateView()
}

// RayCast will project a ray from the camera's position given a screen position
// using the camera's view and projection matrices.
func (c *StandardCamera) RayCast(screenPos matrix.Vec2) collision.Ray {
	return c.internalRayCast(screenPos, c.position)
}

// TryPlaneHit will project a ray from the camera's position given a screen
// position and test if it hits a plane. If it does, it will return the hit
// position and true. If it does not, it will return the zero vector and false.
func (c *StandardCamera) TryPlaneHit(screenPos matrix.Vec2, planePos, planeNml matrix.Vec3) (hit matrix.Vec3, success bool) {
	r := c.RayCast(screenPos)
	d := matrix.Vec3Dot(planeNml, r.Direction)
	if matrix.Abs(d) < matrix.FloatSmallestNonzero {
		return hit, success
	}
	diff := planePos.Subtract(r.Origin)
	distance := matrix.Vec3Dot(diff, planeNml) / d
	if distance < 0 {
		return hit, success
	}
	hit = r.Point(distance)
	return hit, true
}

// ForwardPlaneHit will project a ray from the camera's position given a screen
// position and test if it hits a plane directly facing the cameras position.
func (c *StandardCamera) ForwardPlaneHit(screenPos matrix.Vec2, planePos matrix.Vec3) (matrix.Vec3, bool) {
	fwd := c.Forward()
	return c.TryPlaneHit(screenPos, planePos, fwd)
}

// Position will return the position of the camera.
func (c *StandardCamera) Position() matrix.Vec3 { return c.position }

// Width will return the width of the camera's viewport.
func (c *StandardCamera) Width() float32 { return c.width }

// Height will return the height of the camera's viewport.
func (c *StandardCamera) Height() float32 { return c.height }

// View will return the view matrix of the camera.
func (c *StandardCamera) View() matrix.Mat4 { return c.view }

// Projection will return the projection matrix of the camera.
func (c *StandardCamera) Projection() matrix.Mat4 { return c.projection }

// LookAt will return the look at position of the camera.
func (c *StandardCamera) LookAt() matrix.Vec3 { return c.lookAt }

// NearPlane will return the near plane of the camera.
func (c *StandardCamera) NearPlane() float32 { return c.nearPlane }

// FarPlane will return the far plane of the camera.
func (c *StandardCamera) FarPlane() float32 { return c.farPlane }

func (c *StandardCamera) initializeValues(position matrix.Vec3) {
	c.fieldOfView = 60.0
	c.nearPlane = 0.01
	c.farPlane = 500.0
	c.position = position
	c.view = matrix.Mat4Identity()
	c.projection = matrix.Mat4Identity()
	c.up = matrix.Vec3Up()
	c.lookAt = matrix.Vec3Forward()
}

func (c *StandardCamera) initialize(width, height float32) {
	c.updateProjection = c.internalUpdateProjection
	c.updateView = c.internalUpdateView
	c.setProjection(width, height)
	c.updateView()
}

func (c *StandardCamera) setProjection(width, height float32) {
	c.width = width
	c.height = height
	c.updateProjection()
}

func (c *StandardCamera) internalUpdateProjection() {
	if !c.isOrthographic {
		c.projection.Perspective(matrix.Deg2Rad(c.fieldOfView),
			c.width/c.height, c.nearPlane, c.farPlane)
	} else {
		c.projection.Orthographic(-c.width*0.5, c.width*0.5, -c.height*0.5, c.height*0.5, c.nearPlane, c.farPlane)
	}
	c.iProjection = c.projection
	c.iProjection.Inverse()
}

func (c *StandardCamera) internalUpdateView() {
	if !c.isOrthographic {
		c.view = matrix.Mat4LookAt(c.position, c.lookAt, c.up)
	} else {
		iPos := c.position
		iPos.ScaleAssign(-1.0)
		c.view.Reset()
		c.view.Translate(iPos)
	}
	c.iView = c.view
	c.iView.Inverse()
	c.updateFrustum()
}

func (c *StandardCamera) updateFrustum() {
	vp := matrix.Mat4Multiply(c.view, c.projection)
	for i := 3; i >= 0; i-- {
		c.frustum.Planes[0].SetFloatValue(vp[i*4+3]+vp[i*4+0], i)
		c.frustum.Planes[1].SetFloatValue(vp[i*4+3]-vp[i*4+0], i)
		c.frustum.Planes[2].SetFloatValue(vp[i*4+3]+vp[i*4+1], i)
		c.frustum.Planes[3].SetFloatValue(vp[i*4+3]-vp[i*4+1], i)
		c.frustum.Planes[4].SetFloatValue(vp[i*4+3]+vp[i*4+2], i)
		c.frustum.Planes[5].SetFloatValue(vp[i*4+3]-vp[i*4+2], i)
	}
}

func (c *StandardCamera) internalRayCast(screenPos matrix.Vec2, pos matrix.Vec3) collision.Ray {
	x := (2.0*screenPos.X())/c.width - 1.0
	y := 1.0 - (2.0*screenPos.Y())/c.height
	// Normalized Device Coordinates
	rayNds := matrix.Vec3{x, y, 1}
	rayClip := matrix.Vec4{rayNds.X(), rayNds.Y(), -1, 1}
	rayEye := matrix.Vec4MultiplyMat4(rayClip, c.iProjection)
	rayEye = matrix.Vec4{rayEye.X(), rayEye.Y(), -1, 0}
	// Normalize up/down/left/right
	res := matrix.Vec4MultiplyMat4(rayEye, c.view)
	rayWorld := matrix.Vec3{res.X(), res.Y(), res.Z()}
	rayWorld.Normalize()
	return collision.Ray{Origin: pos, Direction: rayWorld}
}
