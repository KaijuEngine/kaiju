package cameras

import (
	"kaiju/collision"
	"kaiju/matrix"
)

type StandardCamera struct {
	view           matrix.Mat4
	iView          matrix.Mat4
	projection     matrix.Mat4
	iProjection    matrix.Mat4
	frustum        collision.Frustum
	position       matrix.Vec3
	lookAt         matrix.Vec3
	up             matrix.Vec3
	fieldOfView    float32
	nearPlane      float32
	farPlane       float32
	pitch          float32
	yaw            float32
	width          float32
	height         float32
	isOrthographic bool
}

func NewStandardCamera(width, height float32, position matrix.Vec3) *StandardCamera {
	c := new(StandardCamera)
	c.initializeValues(position)
	c.initialize(width, height)
	return c
}

func NewStandardCameraOrthographic(width, height float32, position matrix.Vec3) *StandardCamera {
	c := new(StandardCamera)
	c.initializeValues(position)
	c.isOrthographic = true
	c.nearPlane = -1
	c.initialize(width, height)
	return c
}

func (c *StandardCamera) initializeValues(position matrix.Vec3) {
	c.fieldOfView = 60.0
	c.nearPlane = 0.01
	c.farPlane = 500.0
	c.position = position
	c.view = matrix.Mat4Identity()
	c.projection = matrix.Mat4Identity()
	c.yaw = -90.0
	c.pitch = 0.0
	c.up = matrix.Vec3Up()
	c.lookAt = matrix.Vec3Forward()
}

func (c *StandardCamera) initialize(width, height float32) {
	c.setProjection(width, height)
	c.updateView()
}

func (c *StandardCamera) SetPosition(position matrix.Vec3) {
	c.position = position
	c.updateView()
}

func (c *StandardCamera) setProjection(width, height float32) {
	c.width = width
	c.height = height
	c.updateProjection()
}

func (c *StandardCamera) updateProjection() {
	if !c.isOrthographic {
		c.projection.Perspective(matrix.Deg2Rad(c.fieldOfView),
			c.width/c.height, c.nearPlane, c.farPlane)
	} else {
		c.projection.Orthographic(-c.width*0.5, c.width*0.5, -c.height*0.5, c.height*0.5, c.nearPlane, c.farPlane)
	}
	c.iProjection = c.projection
	c.iProjection.Inverse()
}

func (c *StandardCamera) updateView() {
	if !c.isOrthographic {
		c.view.LookAt(c.position, c.lookAt, c.up)
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
	vp := c.view.Multiply(c.projection)
	for i := 3; i >= 0; i-- {
		c.frustum.Planes[0].SetFloatValue(vp[i*4+3]+vp[i*4+0], i)
		c.frustum.Planes[1].SetFloatValue(vp[i*4+3]-vp[i*4+0], i)
		c.frustum.Planes[2].SetFloatValue(vp[i*4+3]+vp[i*4+1], i)
		c.frustum.Planes[3].SetFloatValue(vp[i*4+3]-vp[i*4+1], i)
		c.frustum.Planes[4].SetFloatValue(vp[i*4+3]+vp[i*4+2], i)
		c.frustum.Planes[5].SetFloatValue(vp[i*4+3]-vp[i*4+2], i)
	}
}

func (c *StandardCamera) SetFOV(fov float32) {
	c.fieldOfView = fov
	c.updateProjection()
}

func (c *StandardCamera) SetNearPlane(near float32) {
	c.nearPlane = near
	c.updateProjection()
}

func (c *StandardCamera) SetFarPlane(far float32) {
	c.farPlane = far
	c.updateProjection()
}

func (c *StandardCamera) SetWidth(width float32) {
	c.width = width
	c.updateProjection()
}

func (c *StandardCamera) SetHeight(height float32) {
	c.height = height
	c.updateProjection()
}

func (c *StandardCamera) SetProperties(fov, nearPlane, farPlane, width, height float32) {
	c.fieldOfView = fov
	c.nearPlane = nearPlane
	c.farPlane = farPlane
	c.width = width
	c.height = height
	c.updateProjection()
}

func (c *StandardCamera) SetYaw(yaw float32) {
	c.yaw = yaw
	if yaw > 360 {
		c.yaw -= 360
	} else if yaw < -360 {
		c.yaw += 360
	}
	direction := matrix.Vec3{
		matrix.Cos(matrix.Deg2Rad(c.yaw)) * matrix.Cos(matrix.Deg2Rad(c.pitch)),
		matrix.Sin(matrix.Deg2Rad(c.pitch)),
		matrix.Sin(matrix.Deg2Rad(c.yaw)) * matrix.Cos(matrix.Deg2Rad(c.pitch)),
	}
	direction.Normalize()
	c.lookAt = c.position.Add(direction)
	c.updateView()
}

func (c *StandardCamera) SetPitch(pitch float32) {
	c.pitch = pitch
	if c.pitch > 89.0 {
		c.pitch = 89.0
	} else if c.pitch < -89.0 {
		c.pitch = -89.0
	}
	direction := matrix.Vec3{
		matrix.Cos(matrix.Deg2Rad(c.yaw)) * matrix.Cos(matrix.Deg2Rad(c.pitch)),
		matrix.Sin(matrix.Deg2Rad(c.pitch)),
		matrix.Sin(matrix.Deg2Rad(c.yaw)) * matrix.Cos(matrix.Deg2Rad(c.pitch)),
	}
	direction.Normalize()
	c.lookAt = c.position.Add(direction)
	c.updateView()
}

func (c *StandardCamera) SetYawAndPitch(yaw, pitch float32) {
	c.yaw = yaw
	c.pitch = pitch
	if c.pitch > 89.0 {
		c.pitch = 89.0
	} else if c.pitch < -89.0 {
		c.pitch = -89.0
	}
	if yaw > 360 {
		c.yaw -= 360
	} else if yaw < -360 {
		c.yaw += 360
	}
	direction := matrix.Vec3{
		matrix.Cos(matrix.Deg2Rad(c.yaw)) * matrix.Cos(matrix.Deg2Rad(c.pitch)),
		matrix.Sin(matrix.Deg2Rad(c.pitch)),
		matrix.Sin(matrix.Deg2Rad(c.yaw)) * matrix.Cos(matrix.Deg2Rad(c.pitch)),
	}
	direction.Normalize()
	c.lookAt = c.position.Add(direction)
	c.updateView()
}

func (c StandardCamera) Forward() matrix.Vec3 {
	return matrix.Vec3{
		-c.iView[matrix.Mat4x0y2],
		-c.iView[matrix.Mat4x1y2],
		-c.iView[matrix.Mat4x2y2],
	}
}

func (c StandardCamera) Right() matrix.Vec3 {
	return matrix.Vec3{
		c.iView[matrix.Mat4x0y0],
		c.iView[matrix.Mat4x1y0],
		c.iView[matrix.Mat4x2y0],
	}
}

func (c StandardCamera) Up() matrix.Vec3 {
	return matrix.Vec3{
		c.iView[matrix.Mat4x0y1],
		c.iView[matrix.Mat4x1y1],
		c.iView[matrix.Mat4x2y1],
	}
}

func (c *StandardCamera) SetLookAt(position matrix.Vec3) {
	c.lookAt = position
	c.updateView()
}

func (c *StandardCamera) LookAt(point, up matrix.Vec3) {
	c.lookAt = point
	c.up = up
	c.updateView()
}

func (c *StandardCamera) SetPositionAndLookAt(position, lookAt matrix.Vec3) {
	if matrix.Approx(position.Z(), lookAt.Z()) {
		position[matrix.Vz] += 0.0001
	}
	c.position = position
	c.lookAt = lookAt
	c.updateView()
}

func (c StandardCamera) Raycast(screenPos matrix.Vec2) collision.Ray {
	x := (2.0*screenPos.X())/c.width - 1.0
	y := 1.0 - (2.0*screenPos.Y())/c.height
	rayNdc := matrix.Vec3{x, y, -1.0}
	rayClip := matrix.Vec4{rayNdc.X(), rayNdc.Y(), -1.0, 1.0}
	rayEye := rayClip.MultiplyMat4(c.iProjection)
	origin := matrix.Vec3{rayEye.X() / rayEye.W(), rayEye.Y() / rayEye.W(), rayEye.Z() / rayEye.W()}
	origin.AddAssign(c.position)
	rayEye = matrix.Vec4{rayEye.X(), rayEye.Y(), -1.0, 0.0}
	res := rayEye.MultiplyMat4(c.view)
	rayWorld := res.AsVec3().Normal()
	return collision.Ray{Origin: origin, Direction: rayWorld}
}

func (c *StandardCamera) TryPlaneHit(screenPos matrix.Vec2, planePos, planeNml matrix.Vec3) (hit matrix.Vec3, success bool) {
	r := c.Raycast(screenPos)
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

func (c *StandardCamera) ForwardPlaneHit(screenPos matrix.Vec2, planePos matrix.Vec3) (matrix.Vec3, bool) {
	fwd := c.Forward()
	return c.TryPlaneHit(screenPos, planePos, fwd)
}

func (c StandardCamera) Position() matrix.Vec3    { return c.position }
func (c StandardCamera) Width() float32           { return c.width }
func (c StandardCamera) Height() float32          { return c.height }
func (c *StandardCamera) View() matrix.Mat4       { return c.view }
func (c *StandardCamera) Projection() matrix.Mat4 { return c.projection }
func (c *StandardCamera) Center() matrix.Vec3     { return c.lookAt }
func (c *StandardCamera) Yaw() float32            { return c.yaw }
func (c *StandardCamera) Pitch() float32          { return c.pitch }
func (c *StandardCamera) NearPlane() float32      { return c.nearPlane }
func (c *StandardCamera) FarPlane() float32       { return c.farPlane }
