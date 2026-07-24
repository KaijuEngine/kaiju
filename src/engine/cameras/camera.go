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
	SetFOV(fov matrix.Float)
	SetNearPlane(near matrix.Float)
	SetFarPlane(far matrix.Float)
	SetWidth(width matrix.Float)
	SetHeight(height matrix.Float)
	ViewportChanged(width, height matrix.Float)
	SetProperties(fov, nearPlane, farPlane, width, height matrix.Float)
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
	Width() matrix.Float
	Height() matrix.Float
	View() matrix.Mat4
	Projection() matrix.Mat4
	InverseProjection() matrix.Mat4
	LookAt() matrix.Vec3
	NearPlane() matrix.Float
	FarPlane() matrix.Float
	IsOrthographic() bool
	Viewport() matrix.Vec4
	Frustum() graviton.Frustum
	LightFrustumCSMProjections() []matrix.Mat4
	NumCSMCascades() uint8
	CSMCascadeDistances() [4]matrix.Float
	IsDirty() bool
	NewFrame()
}
