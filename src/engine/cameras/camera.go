/******************************************************************************/
/* camera.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package cameras

import (
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

type Camera interface {
	SetPosition(position matrix.Vec3)
	SetFOV(fov float32)
	SetNearPlane(near float32)
	SetFarPlane(far float32)
	SetWidth(width float32)
	SetHeight(height float32)
	ViewportChanged(width, height float32)
	SetProperties(fov, nearPlane, farPlane, width, height float32)
	Forward() matrix.Vec3
	Right() matrix.Vec3
	Up() matrix.Vec3
	SetLookAt(position matrix.Vec3)
	SetLookAtWithUp(point, up matrix.Vec3)
	SetPositionAndLookAt(position, lookAt matrix.Vec3)
	RayCast(cursorPosition matrix.Vec2) graviton.Ray
	TryPlaneHit(cursorPosition matrix.Vec2, planePos, planeNml matrix.Vec3) (hit matrix.Vec3, success bool)
	ForwardPlaneHit(cursorPosition matrix.Vec2, planePos matrix.Vec3) (matrix.Vec3, bool)
	Position() matrix.Vec3
	Width() float32
	Height() float32
	View() matrix.Mat4
	Projection() matrix.Mat4
	InverseProjection() matrix.Mat4
	LookAt() matrix.Vec3
	NearPlane() float32
	FarPlane() float32
	IsOrthographic() bool
	Viewport() matrix.Vec4
	Frustum() graviton.Frustum
	LightFrustumCSMProjections() []matrix.Mat4
	NumCSMCascades() uint8
	CSMCascadeDistances() [4]float32
	IsDirty() bool
	NewFrame()
}
