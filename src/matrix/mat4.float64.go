//go:build F64

/******************************************************************************/
/* mat4.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

type Vec3f = Vec3T[float32]
type Mat4f [16]float32
type Vec4f = Vec4T[float32]

func NewMat4f() Mat4f {
	return Mat4fIdentity()
}

func Mat4fFromSlice(a []float32) Mat4f {
	return Mat4f{
		a[0], a[1], a[2], a[3],
		a[4], a[5], a[6], a[7],
		a[8], a[9], a[10], a[11],
		a[12], a[13], a[14], a[15],
	}
}

func Mat4fZero() Mat4f {
	return Mat4f{}
}

func (m *Mat4f) Mat3f() Mat3f {
	return Mat3f{
		m[x0y0], m[x1y0], m[x2y0],
		m[x0y1], m[x1y1], m[x2y1],
		m[x0y2], m[x1y2], m[x2y2],
	}
}

func (m *Mat4f) Reset() {
	m[x0y0] = 1
	m[x1y0] = 0
	m[x2y0] = 0
	m[x3y0] = 0
	m[x0y1] = 0
	m[x1y1] = 1
	m[x2y1] = 0
	m[x3y1] = 0
	m[x0y2] = 0
	m[x1y2] = 0
	m[x2y2] = 1
	m[x3y2] = 0
	m[x0y3] = 0
	m[x1y3] = 0
	m[x2y3] = 0
	m[x3y3] = 1
}

func (m *Mat4f) Zero() {
	for i := 0; i < len(m); i++ {
		m[i] = 0.0
	}
}

func (m Mat4f) Mat4fProjToVulkan() Mat4f {
	res := m
	res[x1y1] *= -1.0
	return res
}

func (m Mat4f) At(rowIndex, colIndex int) float32 {
	return m[rowIndex*4+colIndex]
}

func (m Mat4f) RowVector(row int) Vec4f {
	return Vec4f{m[row*4+0], m[row*4+1], m[row*4+2], m[row*4+3]}
}

func (m Mat4f) ColumnVector(col int) Vec4f {
	return Vec4f{m[col+0], m[col+4], m[col+8], m[col+12]}
}

func (m Mat4f) Mat4fProject(pos Vec3f, viewport Vec4f) Vec3f {
	pos4 := Vec4f{pos.X(), pos.Y(), pos.Z(), 1.0}
	pos4 = Mat4fMultiplyVec4f(m, pos4)
	z := pos4.Z()
	pos4.Shrink(pos4.W())
	pos4.Scale(0.5)
	pos4.Add(Vec4f{0.5, 0.5, 0.5, 0.5})
	return Vec3f{pos4.X()*viewport.Z() + viewport.X(),
		pos4.Y()*viewport.W() + viewport.Y(), z}
}

func Mat4fToScreenSpace(pos Vec3f, view, projection Mat4f, viewport Vec4f) (Vec3f, bool) {
	clip := Mat4fMultiplyVec4f(projection, Mat4fMultiplyVec4f(view, pos.AsVec4()))
	if clip.W() <= 0.0001 {
		return Vec3f{}, false
	}
	invW := 1.0 / clip.W()
	ndcX := clip.X() * invW
	ndcY := clip.Y() * invW
	ndcZ := clip.Z() * invW
	screenX := (ndcX+1.0)*0.5*viewport.Z() + viewport.X()
	screenY := (ndcY+1.0)*0.5*viewport.W() + viewport.Y()
	return Vec3f{screenX, screenY, ndcZ}, true
}

func (m Mat4f) Mat4fUnProject(p Vec3f, invViewProj Mat4f, viewport Vec4f) Vec3f {
	var v Vec4f
	v.SetX((p.X()-viewport.X())/viewport.Z()*2.0 - 1.0)
	v.SetY(-(p.Y()-viewport.Y())/viewport.W()*2.0 + 1.0)
	v.SetZ(p.Z())
	v.SetW(1.0)
	v = Mat4fMultiplyVec4f(invViewProj, v)
	if v.W() != 0 {
		v.Scale(1.0 / v.W())
	}
	return Vec3f{v.X(), v.Y(), v.Z()}
}

func (m Mat4f) Transpose() Mat4f {
	var res Mat4f
	res[x0y0] = m[x0y0]
	res[x1y0] = m[x0y1]
	res[x2y0] = m[x0y2]
	res[x3y0] = m[x0y3]
	res[x0y1] = m[x1y0]
	res[x1y1] = m[x1y1]
	res[x2y1] = m[x1y2]
	res[x3y1] = m[x1y3]
	res[x0y2] = m[x2y0]
	res[x1y2] = m[x2y1]
	res[x2y2] = m[x2y2]
	res[x3y2] = m[x2y3]
	res[x0y3] = m[x3y0]
	res[x1y3] = m[x3y1]
	res[x2y3] = m[x3y2]
	res[x3y3] = m[x3y3]
	return res
}

func (m *Mat4f) TransposeAssign() {
	result := *m
	m[x0y1] = result[x1y0]
	m[x0y2] = result[x2y0]
	m[x0y3] = result[x3y0]
	m[x1y0] = result[x0y1]
	m[x1y2] = result[x2y1]
	m[x1y3] = result[x3y1]
	m[x2y0] = result[x0y2]
	m[x2y1] = result[x1y2]
	m[x2y3] = result[x3y2]
	m[x3y0] = result[x0y3]
	m[x3y1] = result[x1y3]
	m[x3y2] = result[x2y3]
}

func (m *Mat4f) AddAssign(rhs Mat4f) {
	m[x0y0] += rhs[x0y0]
	m[x1y0] += rhs[x1y0]
	m[x2y0] += rhs[x2y0]
	m[x3y0] += rhs[x3y0]
	m[x0y1] += rhs[x0y1]
	m[x1y1] += rhs[x1y1]
	m[x2y1] += rhs[x2y1]
	m[x3y1] += rhs[x3y1]
	m[x0y2] += rhs[x0y2]
	m[x1y2] += rhs[x1y2]
	m[x2y2] += rhs[x2y2]
	m[x3y2] += rhs[x3y2]
	m[x0y3] += rhs[x0y3]
	m[x1y3] += rhs[x1y3]
	m[x2y3] += rhs[x2y3]
	m[x3y3] += rhs[x3y3]
}

func (m *Mat4f) SubtractAssign(rhs Mat4f) {
	m[x0y0] -= rhs[x0y0]
	m[x1y0] -= rhs[x1y0]
	m[x2y0] -= rhs[x2y0]
	m[x3y0] -= rhs[x3y0]
	m[x0y1] -= rhs[x0y1]
	m[x1y1] -= rhs[x1y1]
	m[x2y1] -= rhs[x2y1]
	m[x3y1] -= rhs[x3y1]
	m[x0y2] -= rhs[x0y2]
	m[x1y2] -= rhs[x1y2]
	m[x2y2] -= rhs[x2y2]
	m[x3y2] -= rhs[x3y2]
	m[x0y3] -= rhs[x0y3]
	m[x1y3] -= rhs[x1y3]
	m[x2y3] -= rhs[x2y3]
	m[x3y3] -= rhs[x3y3]
}

func (m *Mat4f) NegateAssign() {
	for i := 0; i < len(m); i++ {
		m[i] *= -1.0
	}
}

func (a *Mat4f) MultiplyAssign(b Mat4f) {
	*a = Mat4fMultiply(*a, b)
}

func (m *Mat4f) Orthographic(left, right, bottom, top, near, far float32) {
	m.Zero()
	m[x0y0] = 2.0 / (right - left)
	// Vulkan inverts x1y1 (see mat4_projection_gl2vulkan)
	//m[x1y1] = -2.0 / (top - bottom)
	m[x1y1] = mat4X1Y1(2.0 / (top - bottom))
	m[x2y2] = -1.0 / (far - near)
	m[x0y3] = -(right + left) / (right - left)
	m[x1y3] = -(top + bottom) / (top - bottom)
	m[x2y3] = -near / (far - near)
	m[x3y3] = 1.0
}

func (m *Mat4f) Perspective(fovy, aspect, nearVal, farVal float32) {
	var f, fn float32
	m.Zero()
	f = 1.0 / Tan(fovy*0.5)
	fn = 1.0 / (nearVal - farVal)
	m[x0y0] = f / aspect
	m[x1y1] = mat4X1Y1(f)
	m[x2y2] = (nearVal + farVal) * fn
	m[x3y2] = -1.0
	m[x2y3] = 2.0 * nearVal * farVal * fn
}

func (m *Mat4f) Translate(translation Vec3f) {
	(*m)[x0y3] += translation.X()
	(*m)[x1y3] += translation.Y()
	(*m)[x2y3] += translation.Z()
}

func (m *Mat4f) SetTranslation(translation Vec3f) {
	(*m)[x0y3] = translation.X()
	(*m)[x1y3] = translation.Y()
	(*m)[x2y3] = translation.Z()
}

func (m *Mat4f) Scale(scale Vec3f) {
	m[x0y0] *= scale.X()
	m[x1y1] *= scale.Y()
	m[x2y2] *= scale.Z()
}

func (m *Mat4f) LookAt(eye, center, up Vec3f) {
	f := eye.Subtract(center)
	f.Normalize()
	s := Vec3Cross(up, f)
	s.Normalize()
	u := Vec3Cross(f, s)
	ns := s.Negative()
	nu := u.Negative()
	nf := f.Negative()
	*m = Mat4f{
		s.X(), u.X(), f.X(), 0.0,
		s.Y(), u.Y(), f.Y(), 0.0,
		s.Z(), u.Z(), f.Z(), 0.0,
		Vec3Dot(ns, eye), Vec3Dot(nu, eye), Vec3Dot(nf, eye), 1.0}
}

func Mat4fLookAt(eye, center, up Vec3f) Mat4f {
	m := Mat4fIdentity()
	m.LookAt(eye, center, up)
	return m
}

func (m *Mat4f) Rotate(rotate Vec3f) {
	q := QuaternionFromEuler(rotate.AsVec3())
	rm := q.ToMat4f()
	m.MultiplyAssign(rm)
}

func (m *Mat4f) RotateX(angles float32) {
	rot := Mat4fIdentity()
	c := float32(Cos(Deg2Rad(angles)))
	s := float32(Sin(Deg2Rad(angles)))
	rot[x1y1] = c
	rot[x2y1] = -s
	rot[x1y2] = s
	rot[x2y2] = c
	m.MultiplyAssign(rot)
}

func (m *Mat4f) RotateY(angles float32) {
	rot := Mat4f{}
	c := float32(Cos(Deg2Rad(angles)))
	s := float32(Sin(Deg2Rad(angles)))
	rot[x0y0] = c
	rot[x2y0] = -s
	rot[x0y2] = s
	rot[x2y2] = c
	rot[x1y1] = 1
	rot[x3y3] = 1
	m.MultiplyAssign(rot)
}

func (m *Mat4f) RotateZ(angles float32) {
	rot := Mat4fIdentity()
	c := float32(Cos(Deg2Rad(angles)))
	s := float32(Sin(Deg2Rad(angles)))
	rot[x0y0] = c
	rot[x1y0] = -s
	rot[x0y1] = s
	rot[x1y1] = c
	m.MultiplyAssign(rot)
}

func (m *Mat4f) RotateAngles(axis Vec3f, angle float32) {
	a := angle
	c := float32(Cos(Deg2Rad(a)))
	s := float32(Sin(Deg2Rad(a)))
	axisNorm := axis.Normal()
	temp := axisNorm.Scale(1.0 - c)
	var rot Mat4f
	rot[x0y0] = c + temp.X()*axisNorm.X()
	rot[x0y1] = temp.X()*axisNorm.Y() + s*axisNorm.Z()
	rot[x0y2] = temp.X()*axisNorm.Z() - s*axisNorm.Y()
	rot[x1y0] = temp.Y()*axisNorm.X() - s*axisNorm.Z()
	rot[x1y1] = c + temp.Y()*axisNorm.Y()
	rot[x1y2] = temp.Y()*axisNorm.Z() + s*axisNorm.X()
	rot[x2y0] = temp.Z()*axisNorm.X() + s*axisNorm.Y()
	rot[x2y1] = temp.Z()*axisNorm.Y() - s*axisNorm.X()
	rot[x2y2] = c + temp.Z()*axisNorm.Z()
	v0 := Vec4f{m[0], m[4], m[8], m[12]}
	v1 := Vec4f{m[1], m[5], m[9], m[13]}
	v2 := Vec4f{m[2], m[6], m[10], m[14]}
	v3 := Vec4f{m[3], m[7], m[11], m[15]}
	var res Mat4f
	c0x0y0 := v0.Scale(rot[x0y0])
	c0x1y0 := v0.Scale(rot[x1y0])
	c0x2y0 := v0.Scale(rot[x2y0])
	c1x0y1 := v1.Scale(rot[x0y1])
	c1x1y1 := v1.Scale(rot[x1y1])
	c1x2y1 := v1.Scale(rot[x2y1])
	c2x0y2 := v2.Scale(rot[x0y2])
	c2x1y2 := v2.Scale(rot[x1y2])
	c2x2y2 := v2.Scale(rot[x2y2])
	r0 := c0x0y0.Add(c1x0y1)
	r1 := c0x1y0.Add(c1x1y1)
	r2 := c0x2y0.Add(c1x2y1)
	r0.AddAssign(c2x0y2)
	r1.AddAssign(c2x1y2)
	r2.AddAssign(c2x2y2)
	res[0] = r0.X()
	res[4] = r0.Y()
	res[8] = r0.Z()
	res[12] = r0.W()
	res[1] = r1.X()
	res[5] = r1.Y()
	res[9] = r1.Z()
	res[13] = r1.W()
	res[2] = r2.X()
	res[6] = r2.Y()
	res[10] = r2.Z()
	res[14] = r2.W()
	res[3] = v3.X()
	res[7] = v3.Y()
	res[11] = v3.Z()
	res[15] = v3.W()
	*m = res
}

func (m *Mat4f) Inverse() {
	t := [6]float32{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
	a := m[x0y0]
	b := m[x0y1]
	c := m[x0y2]
	d := m[x0y3]
	e := m[x1y0]
	f := m[x1y1]
	g := m[x1y2]
	h := m[x1y3]
	i := m[x2y0]
	j := m[x2y1]
	k := m[x2y2]
	l := m[x2y3]
	n := m[x3y0]
	o := m[x3y1]
	p := m[x3y2]
	q := m[x3y3]
	t[0] = k*q - p*l
	t[1] = j*q - o*l
	t[2] = j*p - o*k
	t[3] = i*q - n*l
	t[4] = i*p - n*k
	t[5] = i*o - n*j
	m[x0y0] = f*t[0] - g*t[1] + h*t[2]
	m[x1y0] = -(e*t[0] - g*t[3] + h*t[4])
	m[x2y0] = e*t[1] - f*t[3] + h*t[5]
	m[x3y0] = -(e*t[2] - f*t[4] + g*t[5])
	m[x0y1] = -(b*t[0] - c*t[1] + d*t[2])
	m[x1y1] = a*t[0] - c*t[3] + d*t[4]
	m[x2y1] = -(a*t[1] - b*t[3] + d*t[5])
	m[x3y1] = a*t[2] - b*t[4] + c*t[5]
	t[0] = g*q - p*h
	t[1] = f*q - o*h
	t[2] = f*p - o*g
	t[3] = e*q - n*h
	t[4] = e*p - n*g
	t[5] = e*o - n*f
	m[x0y2] = b*t[0] - c*t[1] + d*t[2]
	m[x1y2] = -(a*t[0] - c*t[3] + d*t[4])
	m[x2y2] = a*t[1] - b*t[3] + d*t[5]
	m[x3y2] = -(a*t[2] - b*t[4] + c*t[5])
	t[0] = g*l - k*h
	t[1] = f*l - j*h
	t[2] = f*k - j*g
	t[3] = e*l - i*h
	t[4] = e*k - i*g
	t[5] = e*j - i*f
	m[x0y3] = -(b*t[0] - c*t[1] + d*t[2])
	m[x1y3] = a*t[0] - c*t[3] + d*t[4]
	m[x2y3] = -(a*t[1] - b*t[3] + d*t[5])
	m[x3y3] = a*t[2] - b*t[4] + c*t[5]
	det := 1.0 / (a*m[x0y0] + b*m[x1y0] + c*m[x2y0] + d*m[x3y0])
	m[x0y0] *= det
	m[x0y1] *= det
	m[x0y2] *= det
	m[x0y3] *= det
	m[x1y0] *= det
	m[x1y1] *= det
	m[x1y2] *= det
	m[x1y3] *= det
	m[x2y0] *= det
	m[x2y1] *= det
	m[x2y2] *= det
	m[x2y3] *= det
	m[x3y0] *= det
	m[x3y1] *= det
	m[x3y2] *= det
	m[x3y3] *= det
}

func (m Mat4f) TransformPoint(point Vec3f) Vec3f {
	pt0 := Vec4f{point.X(), point.Y(), point.Z(), 1.0}
	res := Mat4fMultiplyVec4f(m, pt0)
	v3 := Vec3f{res.X(), res.Y(), res.Z()}
	v3.ShrinkAssign(res.W())
	return v3
}

func (m Mat4f) Right() Vec3f {
	return Vec3f{m[x0y0], m[x1y0], m[x2y0]}
}

func (m Mat4f) Up() Vec3f {
	return Vec3f{m[x0y1], m[x1y1], m[x2y1]}
}

func (m Mat4f) Forward() Vec3f {
	return Vec3f{m[x0y2], m[x1y2], m[x2y2]}
}

func (m Mat4f) ToQuaternion() Quaternion {
	m00 := Float(m[0])
	m10 := Float(m[1])
	m20 := Float(m[2])
	m01 := Float(m[4])
	m11 := Float(m[5])
	m21 := Float(m[6])
	m02 := Float(m[8])
	m12 := Float(m[9])
	m22 := Float(m[10])
	t := m00 + m11 + m22
	if t > 0 {
		s := 0.5 / Sqrt(t+1.0)
		return Quaternion{0.25 / s, (m12 - m21) * s, (m20 - m02) * s, (m01 - m10) * s}
	} else if m00 > m11 && m00 > m22 {
		s := 2.0 * Sqrt(1.0+m00-m11-m22)
		return Quaternion{(m12 - m21) / s, 0.25 * s, (m10 + m01) / s, (m20 + m02) / s}
	} else if m11 > m22 {
		s := 2.0 * Sqrt(1.0+m11-m00-m22)
		return Quaternion{(m20 - m02) / s, (m10 + m01) / s, 0.25 * s, (m21 + m12) / s}
	} else {
		s := 2.0 * Sqrt(1.0+m22-m00-m11)
		return Quaternion{(m01 - m10) / s, (m20 + m02) / s, (m21 + m12) / s, 0.25 * s}
	}
}

func (m Mat4f) Inverted() Mat4f {
	res := Mat4f{}
	res[x0y0] = m[x0y0]
	res[x1y0] = m[x0y1]
	res[x2y0] = m[x0y2]
	res[x3y0] = -(m[x3y0]*m[x0y0] + m[x3y1]*m[x0y1] + m[x3y2]*m[x0y2])
	res[x0y1] = m[x1y0]
	res[x1y1] = m[x1y1]
	res[x2y1] = m[x1y2]
	res[x3y1] = -(m[x3y0]*m[x1y0] + m[x3y1]*m[x1y1] + m[x3y2]*m[x1y2])
	res[x0y2] = m[x2y0]
	res[x1y2] = m[x2y1]
	res[x2y2] = m[x2y2]
	res[x3y2] = -(m[x3y0]*m[x2y0] + m[x3y1]*m[x2y1] + m[x3y2]*m[x2y2])
	res[x0y3] = 0.0
	res[x1y3] = 0.0
	res[x2y3] = 0.0
	res[x3y3] = 1.0
	return res
}

func (m Mat4f) IsIdentity() bool {
	match := Mat4fIdentity()
	success := true
	for i := 0; i < len(match) && success; i++ {
		success = match[i] == m[i]
	}
	return success
}

func Mat4fApprox(a, b Mat4f) bool {
	res := true
	for i := range a {
		res = res && Abs(a[i]-b[i]) < float32(FloatSmallestNonzero)
	}
	return res
}

func Mat4fApproxTo(a, b Mat4f, delta float32) bool {
	res := true
	for i := range a {
		res = res && Abs(a[i]-b[i]) < delta
	}
	return res
}

func (m Mat4f) Equals(other Mat4f) bool {
	return Mat4fApprox(m, other)
}

func (m Mat4f) ExtractPosition() Vec3f {
	return Vec3f{m[x0y3], m[x1y3], m[x2y3]}
}

func (m Mat4f) ExtractScale() Vec3f {
	sx := Sqrt(m[x0y0]*m[x0y0] + m[x1y0]*m[x1y0] + m[x2y0]*m[x2y0])
	sy := Sqrt(m[x0y1]*m[x0y1] + m[x1y1]*m[x1y1] + m[x2y1]*m[x2y1])
	sz := Sqrt(m[x0y2]*m[x0y2] + m[x1y2]*m[x1y2] + m[x2y2]*m[x2y2])
	det := m[x0y0]*(m[x1y1]*m[x2y2]-m[x1y2]*m[x2y1]) -
		m[x0y1]*(m[x1y0]*m[x2y2]-m[x1y2]*m[x2y0]) +
		m[x0y2]*(m[x1y0]*m[x2y1]-m[x1y1]*m[x2y0])
	if det < 0 {
		if sx >= sy && sx >= sz {
			sx = -sx
		} else if sy >= sx && sy >= sz {
			sy = -sy
		} else {
			sz = -sz
		}
	}
	return Vec3f{sx, sy, sz}
}

func (m Mat4f) ExtractRotation() Quaternion {
	sx := float32(Sqrt(m[x0y0]*m[x0y0] + m[x1y0]*m[x1y0] + m[x2y0]*m[x2y0]))
	sy := float32(Sqrt(m[x0y1]*m[x0y1] + m[x1y1]*m[x1y1] + m[x2y1]*m[x2y1]))
	sz := float32(Sqrt(m[x0y2]*m[x0y2] + m[x1y2]*m[x1y2] + m[x2y2]*m[x2y2]))
	if sx == 0 || sy == 0 || sz == 0 {
		return Quaternion{0, 0, 0, 1}
	}
	r00 := m[x0y0] / sx
	r10 := m[x1y0] / sx
	r20 := m[x2y0] / sx
	r01 := m[x0y1] / sy
	r11 := m[x1y1] / sy
	r21 := m[x2y1] / sy
	r02 := m[x0y2] / sz
	r12 := m[x1y2] / sz
	r22 := m[x2y2] / sz
	det := m[x0y0]*(m[x1y1]*m[x2y2]-m[x1y2]*m[x2y1]) -
		m[x0y1]*(m[x1y0]*m[x2y2]-m[x1y2]*m[x2y0]) +
		m[x0y2]*(m[x1y0]*m[x2y1]-m[x1y1]*m[x2y0])
	flipIndex := -1
	if det < 0 {
		absSx, absSy, absSz := Abs(sx), Abs(sy), Abs(sz)
		if absSx >= absSy && absSx >= absSz {
			flipIndex = 0
		} else if absSy >= absSx && absSy >= absSz {
			flipIndex = 1
		} else {
			flipIndex = 2
		}
	}
	if flipIndex == 0 {
		r00 = -r00
		r10 = -r10
		r20 = -r20
	} else if flipIndex == 1 {
		r01 = -r01
		r11 = -r11
		r21 = -r21
	} else if flipIndex == 2 {
		r02 = -r02
		r12 = -r12
		r22 = -r22
	}
	trace := r00 + r11 + r22
	var q Quaternion
	if trace > 0 {
		t := Sqrt(trace+1) * 2
		q[Qw] = float64(0.25 * t)
		q[Qx] = float64((r21 - r12) / t)
		q[Qy] = float64((r02 - r20) / t)
		q[Qz] = float64((r10 - r01) / t)
	} else if r00 > r11 && r00 > r22 {
		t := Sqrt(1+r00-r11-r22) * 2
		q[Qx] = float64(0.25 * t)
		q[Qw] = float64((r21 - r12) / t)
		q[Qy] = float64((r01 + r10) / t)
		q[Qz] = float64((r02 + r20) / t)
	} else if r11 > r22 {
		t := Sqrt(1+r11-r00-r22) * 2
		q[Qy] = float64(0.25 * t)
		q[Qw] = float64((r02 - r20) / t)
		q[Qx] = float64((r01 + r10) / t)
		q[Qz] = float64((r12 + r21) / t)
	} else {
		t := Sqrt(1+r22-r00-r11) * 2
		q[Qz] = float64(0.25 * t)
		q[Qw] = float64((r10 - r01) / t)
		q[Qx] = float64((r02 + r20) / t)
		q[Qy] = float64((r12 + r21) / t)
	}
	return q
}
