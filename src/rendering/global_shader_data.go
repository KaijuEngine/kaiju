package rendering

import "kaiju/matrix"

type GlobalShaderData struct {
	View             matrix.Mat4
	Projection       matrix.Mat4
	UIView           matrix.Mat4
	UIProjection     matrix.Mat4
	CameraPosition   matrix.Vec3
	UICameraPosition matrix.Vec3
	Time             float32
}
