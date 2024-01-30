package matrix

const (
	Mat4Row0 = iota
	Mat4Row1
	Mat4Row2
	Mat4Row3
)

const (
	Mat4Col0 = iota
	Mat4Col1
	Mat4Col2
	Mat4Col3
)

const (
	x0y0 = iota
	x1y0
	x2y0
	x3y0
	x0y1
	x1y1
	x2y1
	x3y1
	x0y2
	x1y2
	x2y2
	x3y2
	x0y3
	x1y3
	x2y3
	x3y3
)

const (
	Mat4x0y0 = x0y0
	Mat4x1y0 = x1y0
	Mat4x2y0 = x2y0
	Mat4x3y0 = x3y0
	Mat4x0y1 = x0y1
	Mat4x1y1 = x1y1
	Mat4x2y1 = x2y1
	Mat4x3y1 = x3y1
	Mat4x0y2 = x0y2
	Mat4x1y2 = x1y2
	Mat4x2y2 = x2y2
	Mat4x3y2 = x3y2
	Mat4x0y3 = x0y3
	Mat4x1y3 = x1y3
	Mat4x2y3 = x2y3
	Mat4x3y3 = x3y3
)

type Mat4 [16]Float

func NewMat4() Mat4 {
	return Mat4Identity()
}

func Mat4Identity() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func Mat4Zero() Mat4 {
	return Mat4{}
}

func (m *Mat4) Reset() {
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

func (m *Mat4) Zero() {
	for i := 0; i < len(m); i++ {
		m[i] = 0.0
	}
}

func (m Mat4) Mat4ProjToVulkan() Mat4 {
	res := m
	res[x1y1] *= -1.0
	return res
}

func (m Mat4) At(rowIndex, colIndex int) Float {
	return m[rowIndex*4+colIndex]
}

func (m Mat4) RowVector(row int) Vec4 {
	return Vec4{m[row*4+0], m[row*4+1], m[row*4+2], m[row*4+3]}
}

func (m Mat4) ColumnVector(col int) Vec4 {
	return Vec4{m[col+0], m[col+4], m[col+8], m[col+12]}
}

func (m Mat4) Mat4Project(pos Vec3, viewport Vec4) Vec3 {
	pos4 := Vec4{pos.X(), pos.Y(), pos.Z(), 1.0}
	pos4 = m.MultiplyVec4(pos4)
	z := pos4.Z()
	pos4.Shrink(pos4.W())
	pos4.Scale(0.5)
	pos4.Add(Vec4{0.5, 0.5, 0.5, 0.5})
	return Vec3{pos4.X()*viewport.Z() + viewport.X(),
		pos4.Y()*viewport.W() + viewport.Y(), z}
}

func (m Mat4) Mat4Unproject(pos Vec3, viewport Vec4) Vec3 {
	var v Vec4
	v.SetX(2.0*(pos.X()-viewport.X())/viewport.Z() - 1.0)
	v.SetY(2.0*(pos.Y()-viewport.Y())/viewport.W() - 1.0)
	v.SetZ(2.0*pos.Z() - 1.0)
	v.SetW(1.0)
	v = m.MultiplyVec4(v)
	v.Scale(1.0 / v.W())
	return Vec3{v.X(), v.Y(), v.Z()}
}

func (m Mat4) Transpose() Mat4 {
	var res Mat4
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

func (m *Mat4) TransposeAssign() {
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

func (m *Mat4) AddAssign(rhs Mat4) {
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

func (m *Mat4) SubtractAssign(rhs Mat4) {
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

func (m *Mat4) NegateAssign() {
	for i := 0; i < len(m); i++ {
		m[i] *= -1.0
	}
}

func (m Mat4) Multiply(rhs Mat4) Mat4 {
	return Mat4{
		m[x0y0]*rhs[x0y0] + m[x1y0]*rhs[x0y1] + m[x2y0]*rhs[x0y2] + m[x3y0]*rhs[x0y3],
		m[x0y0]*rhs[x1y0] + m[x1y0]*rhs[x1y1] + m[x2y0]*rhs[x1y2] + m[x3y0]*rhs[x1y3],
		m[x0y0]*rhs[x2y0] + m[x1y0]*rhs[x2y1] + m[x2y0]*rhs[x2y2] + m[x3y0]*rhs[x2y3],
		m[x0y0]*rhs[x3y0] + m[x1y0]*rhs[x3y1] + m[x2y0]*rhs[x3y2] + m[x3y0]*rhs[x3y3],
		m[x0y1]*rhs[x0y0] + m[x1y1]*rhs[x0y1] + m[x2y1]*rhs[x0y2] + m[x3y1]*rhs[x0y3],
		m[x0y1]*rhs[x1y0] + m[x1y1]*rhs[x1y1] + m[x2y1]*rhs[x1y2] + m[x3y1]*rhs[x1y3],
		m[x0y1]*rhs[x2y0] + m[x1y1]*rhs[x2y1] + m[x2y1]*rhs[x2y2] + m[x3y1]*rhs[x2y3],
		m[x0y1]*rhs[x3y0] + m[x1y1]*rhs[x3y1] + m[x2y1]*rhs[x3y2] + m[x3y1]*rhs[x3y3],
		m[x0y2]*rhs[x0y0] + m[x1y2]*rhs[x0y1] + m[x2y2]*rhs[x0y2] + m[x3y2]*rhs[x0y3],
		m[x0y2]*rhs[x1y0] + m[x1y2]*rhs[x1y1] + m[x2y2]*rhs[x1y2] + m[x3y2]*rhs[x1y3],
		m[x0y2]*rhs[x2y0] + m[x1y2]*rhs[x2y1] + m[x2y2]*rhs[x2y2] + m[x3y2]*rhs[x2y3],
		m[x0y2]*rhs[x3y0] + m[x1y2]*rhs[x3y1] + m[x2y2]*rhs[x3y2] + m[x3y2]*rhs[x3y3],
		m[x0y3]*rhs[x0y0] + m[x1y3]*rhs[x0y1] + m[x2y3]*rhs[x0y2] + m[x3y3]*rhs[x0y3],
		m[x0y3]*rhs[x1y0] + m[x1y3]*rhs[x1y1] + m[x2y3]*rhs[x1y2] + m[x3y3]*rhs[x1y3],
		m[x0y3]*rhs[x2y0] + m[x1y3]*rhs[x2y1] + m[x2y3]*rhs[x2y2] + m[x3y3]*rhs[x2y3],
		m[x0y3]*rhs[x3y0] + m[x1y3]*rhs[x3y1] + m[x2y3]*rhs[x3y2] + m[x3y3]*rhs[x3y3],
	}
}

func (m *Mat4) MultiplyAssign(rhs Mat4) {
	m[x0y0] = m[x0y0]*rhs[x0y0] + m[x1y0]*rhs[x0y1] + m[x2y0]*rhs[x0y2] + m[x3y0]*rhs[x0y3]
	m[x1y0] = m[x0y0]*rhs[x1y0] + m[x1y0]*rhs[x1y1] + m[x2y0]*rhs[x1y2] + m[x3y0]*rhs[x1y3]
	m[x2y0] = m[x0y0]*rhs[x2y0] + m[x1y0]*rhs[x2y1] + m[x2y0]*rhs[x2y2] + m[x3y0]*rhs[x2y3]
	m[x3y0] = m[x0y0]*rhs[x3y0] + m[x1y0]*rhs[x3y1] + m[x2y0]*rhs[x3y2] + m[x3y0]*rhs[x3y3]
	m[x0y1] = m[x0y1]*rhs[x0y0] + m[x1y1]*rhs[x0y1] + m[x2y1]*rhs[x0y2] + m[x3y1]*rhs[x0y3]
	m[x1y1] = m[x0y1]*rhs[x1y0] + m[x1y1]*rhs[x1y1] + m[x2y1]*rhs[x1y2] + m[x3y1]*rhs[x1y3]
	m[x2y1] = m[x0y1]*rhs[x2y0] + m[x1y1]*rhs[x2y1] + m[x2y1]*rhs[x2y2] + m[x3y1]*rhs[x2y3]
	m[x3y1] = m[x0y1]*rhs[x3y0] + m[x1y1]*rhs[x3y1] + m[x2y1]*rhs[x3y2] + m[x3y1]*rhs[x3y3]
	m[x0y2] = m[x0y2]*rhs[x0y0] + m[x1y2]*rhs[x0y1] + m[x2y2]*rhs[x0y2] + m[x3y2]*rhs[x0y3]
	m[x1y2] = m[x0y2]*rhs[x1y0] + m[x1y2]*rhs[x1y1] + m[x2y2]*rhs[x1y2] + m[x3y2]*rhs[x1y3]
	m[x2y2] = m[x0y2]*rhs[x2y0] + m[x1y2]*rhs[x2y1] + m[x2y2]*rhs[x2y2] + m[x3y2]*rhs[x2y3]
	m[x3y2] = m[x0y2]*rhs[x3y0] + m[x1y2]*rhs[x3y1] + m[x2y2]*rhs[x3y2] + m[x3y2]*rhs[x3y3]
	m[x0y3] = m[x0y3]*rhs[x0y0] + m[x1y3]*rhs[x0y1] + m[x2y3]*rhs[x0y2] + m[x3y3]*rhs[x0y3]
	m[x1y3] = m[x0y3]*rhs[x1y0] + m[x1y3]*rhs[x1y1] + m[x2y3]*rhs[x1y2] + m[x3y3]*rhs[x1y3]
	m[x2y3] = m[x0y3]*rhs[x2y0] + m[x1y3]*rhs[x2y1] + m[x2y3]*rhs[x2y2] + m[x3y3]*rhs[x2y3]
	m[x3y3] = m[x0y3]*rhs[x3y0] + m[x1y3]*rhs[x3y1] + m[x2y3]*rhs[x3y2] + m[x3y3]*rhs[x3y3]
}

func (m *Mat4) Orthographic(left Float, right Float, bottom Float, top Float, near Float, far Float) {
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

func (m *Mat4) Perspective(fovy Float, aspect Float, nearVal Float, farVal Float) {
	var f, fn Float
	m.Zero()
	f = 1.0 / Tan(fovy*0.5)
	fn = 1.0 / (nearVal - farVal)
	m[x0y0] = f / aspect
	m[x1y1] = mat4X1Y1(f)
	m[x2y2] = (nearVal + farVal) * fn
	m[x3y2] = -1.0
	m[x2y3] = 2.0 * nearVal * farVal * fn
}

func (m Mat4) Position() Vec3 {
	return Vec3{m[x0y3], m[x1y3], m[x2y3]}
}

func (m *Mat4) Translate(translation Vec3) {
	(*m)[x0y3] += translation.X()
	(*m)[x1y3] += translation.Y()
	(*m)[x2y3] += translation.Z()
}

func (m *Mat4) SetTranslation(translation Vec3) {
	(*m)[x0y3] = translation.X()
	(*m)[x1y3] = translation.Y()
	(*m)[x2y3] = translation.Z()
}

func (m *Mat4) Scale(scale Vec3) {
	m[x0y0] *= scale.X()
	m[x1y1] *= scale.Y()
	m[x2y2] *= scale.Z()
}

func (m *Mat4) LookAt(eye Vec3, center Vec3, up Vec3) {
	f := eye.Subtract(center)
	f.Normalize()
	s := Vec3Cross(up, f)
	s.Normalize()
	u := Vec3Cross(f, s)
	ns := s.Negative()
	nu := u.Negative()
	nf := f.Negative()
	*m = Mat4{
		s.X(), u.X(), f.X(), 0.0,
		s.Y(), u.Y(), f.Y(), 0.0,
		s.Z(), u.Z(), f.Z(), 0.0,
		Vec3Dot(ns, eye), Vec3Dot(nu, eye), Vec3Dot(nf, eye), 1.0}
}

func (m *Mat4) Rotate(rotate Vec3) {
	q := QuaternionFromEuler(rotate)
	rm := q.ToMat4()
	m.MultiplyAssign(rm)
}

func (m *Mat4) RotateX(angles Float) {
	rot := Mat4Identity()
	rot[x1y1] = Cos(Deg2Rad(angles))
	rot[x2y1] = -Sin(Deg2Rad(angles))
	rot[x1y2] = Sin(Deg2Rad(angles))
	rot[x2y2] = Cos(Deg2Rad(angles))
	m.MultiplyAssign(rot)
}

func (m *Mat4) RotateY(angles Float) {
	rot := Mat4Identity()
	rot[x0y0] = Cos(Deg2Rad(angles))
	rot[x2y0] = Sin(Deg2Rad(angles))
	rot[x0y2] = -Sin(Deg2Rad(angles))
	rot[x2y2] = Cos(Deg2Rad(angles))
	m.MultiplyAssign(rot)
}

func (m *Mat4) RotateZ(angles Float) {
	rot := Mat4Identity()
	rot[x0y0] = Cos(Deg2Rad(angles))
	rot[x1y0] = -Sin(Deg2Rad(angles))
	rot[x0y1] = Sin(Deg2Rad(angles))
	rot[x1y1] = Cos(Deg2Rad(angles))
	m.MultiplyAssign(rot)
}

func (m *Mat4) RotateAngles(axis Vec3, angle Float) {
	a := angle
	c := Cos(a)
	s := Sin(a)
	axisNorm := axis.Normal()
	temp := axisNorm.Scale(1.0 - c)
	var rot Mat4
	rot[x0y0] = c + temp.X()*axisNorm.X()
	rot[x0y1] = temp.X()*axisNorm.Y() + s*axisNorm.Z()
	rot[x0y2] = temp.X()*axisNorm.Z() - s*axisNorm.Y()
	rot[x1y0] = temp.Y()*axisNorm.X() - s*axisNorm.Z()
	rot[x1y1] = c + temp.Y()*axisNorm.Y()
	rot[x1y2] = temp.Y()*axisNorm.Z() + s*axisNorm.X()
	rot[x2y0] = temp.Z()*axisNorm.X() + s*axisNorm.Y()
	rot[x2y1] = temp.Z()*axisNorm.Y() - s*axisNorm.X()
	rot[x2y2] = c + temp.Z()*axisNorm.Z()
	v0 := Vec4{m[0], m[4], m[8], m[12]}
	v1 := Vec4{m[1], m[5], m[9], m[13]}
	v2 := Vec4{m[2], m[6], m[10], m[14]}
	v3 := Vec4{m[3], m[7], m[11], m[15]}
	var res Mat4
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

func (m *Mat4) Inverse() {
	t := [6]Float{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
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

func (m Mat4) MultiplyVec4(rhs Vec4) Vec4 {
	var result Vec4
	row := m.ColumnVector(0)
	result[Vx] = Vec4Dot(row, rhs)
	row = m.ColumnVector(1)
	result[Vy] = Vec4Dot(row, rhs)
	row = m.ColumnVector(2)
	result[Vz] = Vec4Dot(row, rhs)
	row = m.ColumnVector(3)
	result[Vw] = Vec4Dot(row, rhs)
	return result
}

func (m Mat4) TransformPoint(point Vec3) Vec3 {
	pt0 := Vec4{point.X(), point.Y(), point.Z(), 1.0}
	res := m.MultiplyVec4(pt0)
	v3 := Vec3{res.X(), res.Y(), res.Z()}
	v3.Shrink(res.W())
	return v3
}

func (m Mat4) Right() Vec3 {
	return Vec3{m[x0y0], m[x1y0], m[x2y0]}.Normal()
}

func (m Mat4) Up() Vec3 {
	return Vec3{m[x0y1], m[x1y1], m[x2y1]}.Normal()
}

func (m Mat4) Forward() Vec3 {
	return Vec3{m[x0y2], m[x1y2], m[x2y2]}.Normal()
}

func (m Mat4) ToQuaternion() Quaternion {
	m00 := m[0]
	m10 := m[1]
	m20 := m[2]
	m01 := m[4]
	m11 := m[5]
	m21 := m[6]
	m02 := m[8]
	m12 := m[9]
	m22 := m[10]
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

func (m Mat4) Invert() Mat4 {
	res := Mat4{}
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
