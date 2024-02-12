package cameras

import (
	"kaiju/collision"
	"kaiju/matrix"
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
	SetYaw(yaw float32)
	SetPitch(pitch float32)
	SetYawAndPitch(yaw, pitch float32)
	Forward() matrix.Vec3
	Right() matrix.Vec3
	Up() matrix.Vec3
	SetLookAt(position matrix.Vec3)
	LookAt(point, up matrix.Vec3)
	SetPositionAndLookAt(position, lookAt matrix.Vec3)
	Raycast(screenPos matrix.Vec2) collision.Ray
	TryPlaneHit(screenPos matrix.Vec2, planePos, planeNml matrix.Vec3) (hit matrix.Vec3, success bool)
	ForwardPlaneHit(screenPos matrix.Vec2, planePos matrix.Vec3) (matrix.Vec3, bool)
	Position() matrix.Vec3
	Width() float32
	Height() float32
	View() matrix.Mat4
	Projection() matrix.Mat4
	Center() matrix.Vec3
	Yaw() float32
	Pitch() float32
	NearPlane() float32
	FarPlane() float32
	Zoom() float32
}
