/******************************************************************************/
/* standard_camera.go                                                         */
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
	viewWidth        float32
	viewHeight       float32
	isOrthographic   bool
	sizeIsViewSize   bool
	frameDirty       bool
}

// NewStandardCamera creates a new perspective camera using the width/height
// for the viewport and the position to place the camera.
func NewStandardCamera(width, height, viewWidth, viewHeight float32, position matrix.Vec3) *StandardCamera {
	defer tracing.NewRegion("cameras.NewStandardCamera").End()
	c := new(StandardCamera)
	c.initializeValues(position)
	c.initialize(width, height, viewWidth, viewHeight)
	return c
}

// NewStandardCameraOrthographic creates a new orthographic camera using the
// width/height for the viewport and the position to place the camera.
func NewStandardCameraOrthographic(width, height, viewWidth, viewHeight float32, position matrix.Vec3) *StandardCamera {
	defer tracing.NewRegion("cameras.NewStandardCameraOrthographic").End()
	c := new(StandardCamera)
	c.initializeValues(position)
	c.isOrthographic = true
	c.nearPlane = -1
	c.initialize(width, height, viewWidth, viewHeight)
	return c
}

// Frustum will return the camera's view frustum which is updated any time the
// view or project of the camera changes.
func (c *StandardCamera) Frustum() collision.Frustum { return c.frustum }
func (c *StandardCamera) IsDirty() bool              { return c.frameDirty }

func (c *StandardCamera) NewFrame() { c.frameDirty = false }

// SetPosition sets the position of the camera.
func (c *StandardCamera) SetPosition(position matrix.Vec3) {
	defer tracing.NewRegion("StandardCamera.SetPosition").End()
	c.position = position
	c.callUpdateView()
}

// SetFOV sets the field of view for the camera.
func (c *StandardCamera) SetFOV(fov float32) {
	defer tracing.NewRegion("StandardCamera.SetFOV").End()
	c.fieldOfView = fov
	c.callUpdateProjection()
}

// SetNearPlane sets the near plane for the camera.
func (c *StandardCamera) SetNearPlane(near float32) {
	defer tracing.NewRegion("StandardCamera.SetNearPlane").End()
	c.nearPlane = near
	c.callUpdateProjection()
}

// SetFarPlane sets the far plane for the camera.
func (c *StandardCamera) SetFarPlane(far float32) {
	defer tracing.NewRegion("StandardCamera.SetFarPlane").End()
	c.farPlane = far
	c.callUpdateProjection()
}

// SetWidth sets the width of the camera viewport.
func (c *StandardCamera) SetWidth(width float32) {
	defer tracing.NewRegion("StandardCamera.SetWidth").End()
	c.width = width
	c.callUpdateProjection()
}

// SetHeight sets the height of the camera viewport.
func (c *StandardCamera) SetHeight(height float32) {
	defer tracing.NewRegion("StandardCamera.SetHeight").End()
	c.height = height
	c.callUpdateProjection()
}

// Resize sets the width and height of the camera viewport.
func (c *StandardCamera) Resize(width, height float32) {
	defer tracing.NewRegion("StandardCamera.Resize").End()
	c.width = width
	c.height = height
	c.callUpdateProjection()
}

// ViewportChanged will update the camera's projection matrix and should only
// be used when there is a change in the viewport. This is typically done
// internally in the system and should not be called by the end-developer.
func (c *StandardCamera) ViewportChanged(width, height float32) {
	defer tracing.NewRegion("StandardCamera.ViewportChanged").End()
	if c.sizeIsViewSize {
		c.width = width
		c.height = height
	}
	c.viewWidth = width
	c.viewHeight = height
	c.callUpdateProjection()
}

// SetProperties is quick access to set many properties of the camera at once.
// This is typically used for initializing the camera to new values. Calling
// each individual setter for fields would otherwise do needless projection
// matrix updates.
func (c *StandardCamera) SetProperties(fov, nearPlane, farPlane, width, height float32) {
	defer tracing.NewRegion("StandardCamera.SetProperties").End()
	c.fieldOfView = fov
	c.nearPlane = nearPlane
	c.farPlane = farPlane
	c.width = width
	c.height = height
	c.callUpdateProjection()
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
	defer tracing.NewRegion("StandardCamera.SetLookAt").End()
	c.lookAt = position
	c.callUpdateView()
}

// SetLookAtWithUp sets the look at position of the camera and the up vector to use.
func (c *StandardCamera) SetLookAtWithUp(point, up matrix.Vec3) {
	defer tracing.NewRegion("StandardCamera.SetLookAtWithUp").End()
	c.lookAt = point
	c.up = up
	c.callUpdateView()
}

// SetPositionAndLookAt sets the position and look at position of the camera.
// This is often useful for when the camera warps to another location in space
// and avoids needless view matrix updates when setting the position and look
// at separately.
func (c *StandardCamera) SetPositionAndLookAt(position, lookAt matrix.Vec3) {
	defer tracing.NewRegion("StandardCamera.SetPositionAndLookAt").End()
	if matrix.Approx(position.Z(), lookAt.Z()) {
		position[matrix.Vz] += 0.0001
	}
	c.position = position
	c.lookAt = lookAt
	c.callUpdateView()
}

// RayCast will project a ray from the camera's position given a screen position
// using the camera's view and projection matrices.
func (c *StandardCamera) RayCast(cursorPosition matrix.Vec2) collision.Ray {
	defer tracing.NewRegion("StandardCamera.RayCast").End()
	return c.internalRayCast(cursorPosition, c.position)
}

// TryPlaneHit will project a ray from the camera's position given a screen
// position and test if it hits a plane. If it does, it will return the hit
// position and true. If it does not, it will return the zero vector and false.
func (c *StandardCamera) TryPlaneHit(screenPos matrix.Vec2, planePos, planeNml matrix.Vec3) (hit matrix.Vec3, success bool) {
	defer tracing.NewRegion("StandardCamera.TryPlaneHit").End()
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
	defer tracing.NewRegion("StandardCamera.ForwardPlaneHit").End()
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

// InverseProjection will return the inverse projection matrix of the camera.
func (c *StandardCamera) InverseProjection() matrix.Mat4 { return c.iProjection }

// LookAt will return the look at position of the camera.
func (c *StandardCamera) LookAt() matrix.Vec3 { return c.lookAt }

// NearPlane will return the near plane of the camera.
func (c *StandardCamera) NearPlane() float32 { return c.nearPlane }

// FarPlane will return the far plane of the camera.
func (c *StandardCamera) FarPlane() float32 { return c.farPlane }

// IsOrthographic will return if this camera is set to be an orthographic camera
func (c *StandardCamera) IsOrthographic() bool { return c.isOrthographic }

func (c *StandardCamera) Viewport() matrix.Vec4 {
	return matrix.NewVec4(0, 0, c.viewWidth, c.viewHeight)
}

func (c *StandardCamera) initializeValues(position matrix.Vec3) {
	defer tracing.NewRegion("StandardCamera.initializeValues").End()
	c.fieldOfView = 60.0
	c.nearPlane = 0.01
	c.farPlane = 500.0
	c.position = position
	c.view = matrix.Mat4Identity()
	c.projection = matrix.Mat4Identity()
	c.up = matrix.Vec3Up()
	c.lookAt = position.Add(matrix.Vec3Forward())
}

func (c *StandardCamera) initialize(width, height, viewWidth, viewHeight float32) {
	defer tracing.NewRegion("StandardCamera.initialize").End()
	c.updateProjection = c.internalUpdateProjection
	c.updateView = c.internalUpdateView
	c.viewWidth = viewWidth
	c.viewHeight = viewHeight
	c.setProjection(width, height)
	c.callUpdateView()
	c.sizeIsViewSize = width == viewWidth && height == viewHeight
}

func (c *StandardCamera) setProjection(width, height float32) {
	defer tracing.NewRegion("StandardCamera.setProjection").End()
	c.width = width
	c.height = height
	c.callUpdateProjection()
}

func (c *StandardCamera) internalUpdateProjection() {
	defer tracing.NewRegion("StandardCamera.internalUpdateProjection").End()
	if !c.isOrthographic {
		c.projection.Perspective(matrix.Deg2Rad(c.fieldOfView),
			c.width/c.height, c.nearPlane, c.farPlane)
	} else {
		c.projection.Orthographic(-c.width*0.5, c.width*0.5, -c.height*0.5, c.height*0.5, c.nearPlane, c.farPlane)
	}
	c.iProjection = c.projection
	c.iProjection.Inverse()
	c.updateFrustum()
}

func (c *StandardCamera) internalUpdateView() {
	defer tracing.NewRegion("StandardCamera.internalUpdateView").End()
	c.view = matrix.Mat4LookAt(c.position, c.lookAt, c.up)
	c.iView = c.view
	c.iView.Inverse()
	c.updateFrustum()
}

func (c *StandardCamera) updateFrustum() {
	defer tracing.NewRegion("StandardCamera.updateFrustum").End()
	vp := matrix.Mat4Multiply(c.view, c.projection)
	c.frustum.ExtractPlanes(vp)
}

func (c *StandardCamera) internalRayCast(cursorPosition matrix.Vec2, pos matrix.Vec3) collision.Ray {
	defer tracing.NewRegion("StandardCamera.internalRayCast").End()
	x := (2.0*cursorPosition.X())/c.viewWidth - 1.0
	y := 1.0 - (2.0*cursorPosition.Y())/c.viewHeight
	var origin, direction matrix.Vec3
	if !c.isOrthographic {
		origin = pos
		// Normalized Device Coordinates
		rayNds := matrix.Vec3{x, y, 1}
		rayClip := matrix.Vec4{rayNds.X(), rayNds.Y(), -1, 1}
		rayEye := matrix.Vec4MultiplyMat4(rayClip, c.iProjection)
		rayEye = matrix.Vec4{rayEye.X(), rayEye.Y(), -1, 0}
		// Normalize up/down/left/right
		res := matrix.Vec4MultiplyMat4(rayEye, c.view)
		direction = matrix.Vec3{res.X(), res.Y(), res.Z()}
		direction.Normalize()
	} else {
		up := c.Up()
		forward := c.Forward()
		worldX := x * c.width / 2.0
		worldY := y * c.height / 2.0
		right := c.Right()
		origin = c.position.Add(right.Scale(worldX)).Add(up.Scale(worldY))
		direction = forward
	}
	return collision.Ray{Origin: origin, Direction: direction}
}

func (c *StandardCamera) callUpdateView() {
	c.updateView()
	c.frameDirty = true
}

func (c *StandardCamera) callUpdateProjection() {
	c.updateProjection()
	c.frameDirty = true
}
