/******************************************************************************/
/* snapshot_camera.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package cameras

import (
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

type SnapshotCamera struct {
	view             matrix.Mat4
	iView            matrix.Mat4
	projection       matrix.Mat4
	iProjection      matrix.Mat4
	frustum          graviton.Frustum
	position         matrix.Vec3
	lookAt           matrix.Vec3
	nearPlane        float32
	farPlane         float32
	width            float32
	height           float32
	viewport         matrix.Vec4
	orthographic     bool
	csmProjections   []matrix.Mat4
	cascadeCount     uint8
	cascadeDistances [4]float32
}

func NewSnapshotCamera(camera Camera) *SnapshotCamera {
	if camera == nil {
		return nil
	}
	projections := camera.LightFrustumCSMProjections()
	csm := make([]matrix.Mat4, len(projections))
	copy(csm, projections)
	view := camera.View()
	iView := view
	iView.Inverse()
	return &SnapshotCamera{
		view:             view,
		iView:            iView,
		projection:       camera.Projection(),
		iProjection:      camera.InverseProjection(),
		frustum:          camera.Frustum(),
		position:         camera.Position(),
		lookAt:           camera.LookAt(),
		nearPlane:        camera.NearPlane(),
		farPlane:         camera.FarPlane(),
		width:            camera.Width(),
		height:           camera.Height(),
		viewport:         camera.Viewport(),
		orthographic:     camera.IsOrthographic(),
		csmProjections:   csm,
		cascadeCount:     camera.NumCSMCascades(),
		cascadeDistances: camera.CSMCascadeDistances(),
	}
}

func (c *SnapshotCamera) SetPosition(position matrix.Vec3)      { c.position = position }
func (c *SnapshotCamera) SetFOV(float32)                        {}
func (c *SnapshotCamera) SetNearPlane(near float32)             { c.nearPlane = near }
func (c *SnapshotCamera) SetFarPlane(far float32)               { c.farPlane = far }
func (c *SnapshotCamera) SetWidth(width float32)                { c.width = width }
func (c *SnapshotCamera) SetHeight(height float32)              { c.height = height }
func (c *SnapshotCamera) ViewportChanged(width, height float32) { c.width, c.height = width, height }
func (c *SnapshotCamera) SetProperties(_, near, far, width, height float32) {
	c.nearPlane, c.farPlane, c.width, c.height = near, far, width, height
}
func (c *SnapshotCamera) SetLookAt(position matrix.Vec3)       { c.lookAt = position }
func (c *SnapshotCamera) SetLookAtWithUp(point, _ matrix.Vec3) { c.lookAt = point }
func (c *SnapshotCamera) SetPositionAndLookAt(position, lookAt matrix.Vec3) {
	c.position, c.lookAt = position, lookAt
}

func (c *SnapshotCamera) Forward() matrix.Vec3 {
	return matrix.Vec3{-c.iView[matrix.Mat4x0y2], -c.iView[matrix.Mat4x1y2], -c.iView[matrix.Mat4x2y2]}
}

func (c *SnapshotCamera) Right() matrix.Vec3 {
	return matrix.Vec3{c.iView[matrix.Mat4x0y0], c.iView[matrix.Mat4x1y0], c.iView[matrix.Mat4x2y0]}
}

func (c *SnapshotCamera) Up() matrix.Vec3 {
	return matrix.Vec3{c.iView[matrix.Mat4x0y1], c.iView[matrix.Mat4x1y1], c.iView[matrix.Mat4x2y1]}
}

func (c *SnapshotCamera) RayCast(cursorPosition matrix.Vec2) graviton.Ray {
	x := (2.0*cursorPosition.X())/c.viewport.Z() - 1.0
	y := 1.0 - (2.0*cursorPosition.Y())/c.viewport.W()
	if c.orthographic {
		origin := c.position.Add(c.Right().Scale(x * c.width / 2.0)).Add(c.Up().Scale(y * c.height / 2.0))
		return graviton.Ray{Origin: origin, Direction: c.Forward()}
	}
	rayNds := matrix.Vec3{x, y, 1}
	rayClip := matrix.Vec4{rayNds.X(), rayNds.Y(), -1, 1}
	rayEye := matrix.Vec4MultiplyMat4(rayClip, c.iProjection)
	rayEye = matrix.Vec4{rayEye.X(), rayEye.Y(), -1, 0}
	res := matrix.Vec4MultiplyMat4(rayEye, c.view)
	direction := matrix.Vec3{res.X(), res.Y(), res.Z()}
	direction.Normalize()
	return graviton.Ray{Origin: c.position, Direction: direction}
}

func (c *SnapshotCamera) TryPlaneHit(cursorPosition matrix.Vec2, planePos, planeNml matrix.Vec3) (matrix.Vec3, bool) {
	r := c.RayCast(cursorPosition)
	d := matrix.Vec3Dot(planeNml, r.Direction)
	if matrix.Abs(d) < matrix.FloatSmallestNonzero {
		return matrix.Vec3Zero(), false
	}
	distance := matrix.Vec3Dot(planePos.Subtract(r.Origin), planeNml) / d
	if distance < 0 {
		return matrix.Vec3Zero(), false
	}
	return r.Point(distance), true
}

func (c *SnapshotCamera) ForwardPlaneHit(cursorPosition matrix.Vec2, planePos matrix.Vec3) (matrix.Vec3, bool) {
	return c.TryPlaneHit(cursorPosition, planePos, c.Forward())
}

func (c *SnapshotCamera) Position() matrix.Vec3                     { return c.position }
func (c *SnapshotCamera) Width() float32                            { return c.width }
func (c *SnapshotCamera) Height() float32                           { return c.height }
func (c *SnapshotCamera) View() matrix.Mat4                         { return c.view }
func (c *SnapshotCamera) Projection() matrix.Mat4                   { return c.projection }
func (c *SnapshotCamera) InverseProjection() matrix.Mat4            { return c.iProjection }
func (c *SnapshotCamera) LookAt() matrix.Vec3                       { return c.lookAt }
func (c *SnapshotCamera) NearPlane() float32                        { return c.nearPlane }
func (c *SnapshotCamera) FarPlane() float32                         { return c.farPlane }
func (c *SnapshotCamera) IsOrthographic() bool                      { return c.orthographic }
func (c *SnapshotCamera) Viewport() matrix.Vec4                     { return c.viewport }
func (c *SnapshotCamera) Frustum() graviton.Frustum                 { return c.frustum }
func (c *SnapshotCamera) LightFrustumCSMProjections() []matrix.Mat4 { return c.csmProjections }
func (c *SnapshotCamera) NumCSMCascades() uint8                     { return c.cascadeCount }
func (c *SnapshotCamera) CSMCascadeDistances() [4]float32           { return c.cascadeDistances }
func (c *SnapshotCamera) IsDirty() bool                             { return true }
func (c *SnapshotCamera) NewFrame()                                 {}
