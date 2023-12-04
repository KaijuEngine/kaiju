package rendering

import "kaiju/matrix"

type GlobalShaderData struct {
	View           matrix.Mat4
	Projection     matrix.Mat4
	CameraPosition matrix.Vec3
	Time           float32
}
