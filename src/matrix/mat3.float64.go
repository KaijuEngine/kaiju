//go:build F64

/******************************************************************************/
/* mat3.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

type Mat3f [9]float32

func NewMat3f() Mat3f {
	return Mat3fIdentity()
}

func (m Mat3f) RowVector(row int) Vec3f {
	return Vec3f{m[row*3+0], m[row*3+1], m[row*3+2]}
}

func (m Mat3f) ColumnVector(col int) Vec3f {
	return Vec3f{m[col+0], m[col+3], m[col+6]}
}

func Mat3fFromSlice(a []float32) Mat3f {
	return Mat3f{
		a[0], a[1], a[2],
		a[3], a[4], a[5],
		a[6], a[7], a[8],
	}
}

func Mat3fFromVec3f(v Vec3f) Mat3f {
	return Mat3f{
		v.X(), 0, 0,
		0, v.Y(), 0,
		0, 0, v.Z(),
	}
}

func Mat3fIdentity() Mat3f {
	return Mat3f{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

func Mat3fZero() Mat3f {
	return Mat3f{}
}

func (m *Mat3f) Reset() {
	m[x0y0_3] = 1
	m[x1y0_3] = 0
	m[x2y0_3] = 0
	m[x0y1_3] = 0
	m[x1y1_3] = 1
	m[x2y1_3] = 0
	m[x0y2_3] = 0
	m[x1y2_3] = 0
	m[x2y2_3] = 1
}

func (m *Mat3f) Zero() {
	for i := 0; i < len(m); i++ {
		m[i] = 0.0
	}
}

func Mat3fFromMat4(m Mat4f) Mat3f { return m.Mat3f() }

func (m Mat3f) ToMat4() Mat4f {
	return Mat4f{
		m[0], m[1], m[2], 0,
		m[3], m[4], m[5], 0,
		m[6], m[7], m[8], 0,
		0, 0, 0, 1,
	}
}

func (m Mat3f) At(rowIndex, colIndex int) float32 {
	return m[rowIndex*3+colIndex]
}

func (m Mat3f) Transpose() Mat3f {
	return Mat3f{
		m[x0y0_3], m[x0y1_3], m[x0y2_3],
		m[x1y0_3], m[x1y1_3], m[x1y2_3],
		m[x2y0_3], m[x2y1_3], m[x2y2_3],
	}
}

func (m *Mat3f) TransposeAssign() {
	t := *m
	m[x1y0_3] = t[x0y1_3]
	m[x2y0_3] = t[x0y2_3]
	m[x0y1_3] = t[x1y0_3]
	m[x2y1_3] = t[x1y2_3]
	m[x0y2_3] = t[x2y0_3]
	m[x1y2_3] = t[x2y1_3]
}

func Mat3fMultiply(a, b Mat3f) Mat3f {
	return Mat3f{
		a[x0y0_3]*b[x0y0_3] + a[x1y0_3]*b[x0y1_3] + a[x2y0_3]*b[x0y2_3],
		a[x0y0_3]*b[x1y0_3] + a[x1y0_3]*b[x1y1_3] + a[x2y0_3]*b[x1y2_3],
		a[x0y0_3]*b[x2y0_3] + a[x1y0_3]*b[x2y1_3] + a[x2y0_3]*b[x2y2_3],

		a[x0y1_3]*b[x0y0_3] + a[x1y1_3]*b[x0y1_3] + a[x2y1_3]*b[x0y2_3],
		a[x0y1_3]*b[x1y0_3] + a[x1y1_3]*b[x1y1_3] + a[x2y1_3]*b[x1y2_3],
		a[x0y1_3]*b[x2y0_3] + a[x1y1_3]*b[x2y1_3] + a[x2y1_3]*b[x2y2_3],

		a[x0y2_3]*b[x0y0_3] + a[x1y2_3]*b[x0y1_3] + a[x2y2_3]*b[x0y2_3],
		a[x0y2_3]*b[x1y0_3] + a[x1y2_3]*b[x1y1_3] + a[x2y2_3]*b[x1y2_3],
		a[x0y2_3]*b[x2y0_3] + a[x1y2_3]*b[x2y1_3] + a[x2y2_3]*b[x2y2_3],
	}
}

func (m Mat3f) Multiply(rhs Mat3f) Mat3f {
	return Mat3fMultiply(m, rhs)
}

func (m *Mat3f) MultiplyAssign(rhs Mat3f) {
	*m = Mat3fMultiply(*m, rhs)
}

func Mat3fMultiplyVec3f(m Mat3f, v Vec3f) Vec3f {
	return Vec3f{
		m[x0y0_3]*v[0] + m[x1y0_3]*v[1] + m[x2y0_3]*v[2],
		m[x0y1_3]*v[0] + m[x1y1_3]*v[1] + m[x2y1_3]*v[2],
		m[x0y2_3]*v[0] + m[x1y2_3]*v[1] + m[x2y2_3]*v[2],
	}
}

func (m Mat3f) MultiplyVec3f(v Vec3f) Vec3f {
	return Mat3fMultiplyVec3f(m, v)
}

func (m *Mat3f) AddAssign(rhs Mat3f) {
	m[x0y0_3] += rhs[x0y0_3]
	m[x1y0_3] += rhs[x1y0_3]
	m[x2y0_3] += rhs[x2y0_3]
	m[x0y1_3] += rhs[x0y1_3]
	m[x1y1_3] += rhs[x1y1_3]
	m[x2y1_3] += rhs[x2y1_3]
	m[x0y2_3] += rhs[x0y2_3]
	m[x1y2_3] += rhs[x1y2_3]
	m[x2y2_3] += rhs[x2y2_3]
}

func (m Mat3f) Add(rhs Mat3f) Mat3f {
	res := m
	res.AddAssign(rhs)
	return res
}

func (m *Mat3f) SubtractAssign(rhs Mat3f) {
	m[x0y0_3] -= rhs[x0y0_3]
	m[x1y0_3] -= rhs[x1y0_3]
	m[x2y0_3] -= rhs[x2y0_3]
	m[x0y1_3] -= rhs[x0y1_3]
	m[x1y1_3] -= rhs[x1y1_3]
	m[x2y1_3] -= rhs[x2y1_3]
	m[x0y2_3] -= rhs[x0y2_3]
	m[x1y2_3] -= rhs[x1y2_3]
	m[x2y2_3] -= rhs[x2y2_3]
}

func (m Mat3f) Subtract(rhs Mat3f) Mat3f {
	res := m
	res.SubtractAssign(rhs)
	return res
}

func (m *Mat3f) NegateAssign() {
	for i := 0; i < len(m); i++ {
		m[i] *= -1.0
	}
}

func (m Mat3f) Negate() Mat3f {
	res := m
	res.NegateAssign()
	return res
}

func (m *Mat3f) ScaleAssign(s float32) {
	for i := 0; i < len(m); i++ {
		m[i] *= s
	}
}

func (m Mat3f) Scale(s float32) Mat3f {
	res := m
	res.ScaleAssign(s)
	return res
}

func Mat3fDeterminant(m Mat3f) float32 {
	return m[x0y0_3]*(m[x1y1_3]*m[x2y2_3]-m[x1y2_3]*m[x2y1_3]) -
		m[x1y0_3]*(m[x0y1_3]*m[x2y2_3]-m[x0y2_3]*m[x2y1_3]) +
		m[x2y0_3]*(m[x0y1_3]*m[x1y2_3]-m[x0y2_3]*m[x1y1_3])
}

func (m Mat3f) Determinant() float32 {
	return Mat3fDeterminant(m)
}

func (m *Mat3f) Inverse() {
	det := m.Determinant()
	if det == 0 {
		*m = Mat3fZero()
		return
	}
	invDet := 1.0 / det
	res := Mat3f{
		(m[x1y1_3]*m[x2y2_3] - m[x1y2_3]*m[x2y1_3]) * invDet,
		(m[x0y2_3]*m[x2y1_3] - m[x0y1_3]*m[x2y2_3]) * invDet,
		(m[x0y1_3]*m[x1y2_3] - m[x0y2_3]*m[x1y1_3]) * invDet,

		(m[x1y2_3]*m[x2y0_3] - m[x1y0_3]*m[x2y2_3]) * invDet,
		(m[x0y0_3]*m[x2y2_3] - m[x0y2_3]*m[x2y0_3]) * invDet,
		(m[x0y2_3]*m[x1y0_3] - m[x0y0_3]*m[x1y2_3]) * invDet,

		(m[x1y0_3]*m[x2y1_3] - m[x1y1_3]*m[x2y0_3]) * invDet,
		(m[x0y1_3]*m[x2y0_3] - m[x0y0_3]*m[x2y1_3]) * invDet,
		(m[x0y0_3]*m[x1y1_3] - m[x0y1_3]*m[x1y0_3]) * invDet,
	}
	*m = res
}

func (m Mat3f) Inverted() Mat3f {
	res := m
	res.Inverse()
	return res
}

func (m Mat3f) IsIdentity() bool {
	match := Mat3fIdentity()
	success := true
	for i := 0; i < len(match) && success; i++ {
		success = match[i] == m[i]
	}
	return success
}

func Mat3fApprox(a, b Mat3f) bool {
	res := true
	for i := range a {
		res = res && Abs(a[i]-b[i]) < float32(FloatSmallestNonzero)
	}
	return res
}

func Mat3fApproxTo(a, b Mat3f, delta float32) bool {
	res := true
	for i := range a {
		res = res && Abs(a[i]-b[i]) < delta
	}
	return res
}

func (m Mat3f) Equals(other Mat3f) bool {
	return Mat3fApprox(m, other)
}
