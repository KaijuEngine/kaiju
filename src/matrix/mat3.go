/******************************************************************************/
/* mat3.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

const (
	Mat3Row0 = iota
	Mat3Row1
	Mat3Row2
)

const (
	Mat3Col0 = iota
	Mat3Col1
	Mat3Col2
)

const (
	x0y0_3 = iota // 0
	x1y0_3        // 1
	x2y0_3        // 2
	x0y1_3        // 3
	x1y1_3        // 4
	x2y1_3        // 5
	x0y2_3        // 6
	x1y2_3        // 7
	x2y2_3        // 8
)

const (
	Mat3x0y0 = x0y0_3
	Mat3x1y0 = x1y0_3
	Mat3x2y0 = x2y0_3
	Mat3x0y1 = x0y1_3
	Mat3x1y1 = x1y1_3
	Mat3x2y1 = x2y1_3
	Mat3x0y2 = x0y2_3
	Mat3x1y2 = x1y2_3
	Mat3x2y2 = x2y2_3
)

type Mat3 [9]Float

func NewMat3() Mat3 {
	return Mat3Identity()
}

func (m Mat3) RowVector(row int) Vec3 {
	return Vec3{m[row*3+0], m[row*3+1], m[row*3+2]}
}

func (m Mat3) ColumnVector(col int) Vec3 {
	return Vec3{m[col+0], m[col+3], m[col+6]}
}

func Mat3FromSlice(a []Float) Mat3 {
	return Mat3{
		a[0], a[1], a[2],
		a[3], a[4], a[5],
		a[6], a[7], a[8],
	}
}

func Mat3FromVec3(v Vec3) Mat3 {
	return Mat3{
		v.X(), 0, 0,
		0, v.Y(), 0,
		0, 0, v.Z(),
	}
}

func Mat3Identity() Mat3 {
	return Mat3{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

func Mat3Zero() Mat3 {
	return Mat3{}
}

func (m *Mat3) Reset() {
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

func (m *Mat3) Zero() {
	for i := 0; i < len(m); i++ {
		m[i] = 0.0
	}
}

func Mat3FromMat4(m Mat4) Mat3 { return m.Mat3() }

func (m Mat3) ToMat4() Mat4 {
	return Mat4{
		m[0], m[1], m[2], 0,
		m[3], m[4], m[5], 0,
		m[6], m[7], m[8], 0,
		0, 0, 0, 1,
	}
}

func (m Mat3) At(rowIndex, colIndex int) Float {
	return m[rowIndex*3+colIndex]
}

func (m Mat3) Transpose() Mat3 {
	return Mat3{
		m[x0y0_3], m[x0y1_3], m[x0y2_3],
		m[x1y0_3], m[x1y1_3], m[x1y2_3],
		m[x2y0_3], m[x2y1_3], m[x2y2_3],
	}
}

func (m *Mat3) TransposeAssign() {
	t := *m
	m[x1y0_3] = t[x0y1_3]
	m[x2y0_3] = t[x0y2_3]
	m[x0y1_3] = t[x1y0_3]
	m[x2y1_3] = t[x1y2_3]
	m[x0y2_3] = t[x2y0_3]
	m[x1y2_3] = t[x2y1_3]
}

func Mat3Multiply(a, b Mat3) Mat3 {
	return Mat3{
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

func (m Mat3) Multiply(rhs Mat3) Mat3 {
	return Mat3Multiply(m, rhs)
}

func (m *Mat3) MultiplyAssign(rhs Mat3) {
	*m = Mat3Multiply(*m, rhs)
}

func Mat3MultiplyVec3(m Mat3, v Vec3) Vec3 {
	return Vec3{
		m[x0y0_3]*v[0] + m[x1y0_3]*v[1] + m[x2y0_3]*v[2],
		m[x0y1_3]*v[0] + m[x1y1_3]*v[1] + m[x2y1_3]*v[2],
		m[x0y2_3]*v[0] + m[x1y2_3]*v[1] + m[x2y2_3]*v[2],
	}
}

func (m Mat3) MultiplyVec3(v Vec3) Vec3 {
	return Mat3MultiplyVec3(m, v)
}

func (m *Mat3) AddAssign(rhs Mat3) {
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

func (m Mat3) Add(rhs Mat3) Mat3 {
	res := m
	res.AddAssign(rhs)
	return res
}

func (m *Mat3) SubtractAssign(rhs Mat3) {
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

func (m Mat3) Subtract(rhs Mat3) Mat3 {
	res := m
	res.SubtractAssign(rhs)
	return res
}

func (m *Mat3) NegateAssign() {
	for i := 0; i < len(m); i++ {
		m[i] *= -1.0
	}
}

func (m Mat3) Negate() Mat3 {
	res := m
	res.NegateAssign()
	return res
}

func (m *Mat3) ScaleAssign(s Float) {
	for i := 0; i < len(m); i++ {
		m[i] *= s
	}
}

func (m Mat3) Scale(s Float) Mat3 {
	res := m
	res.ScaleAssign(s)
	return res
}

func Mat3Determinant(m Mat3) Float {
	return m[x0y0_3]*(m[x1y1_3]*m[x2y2_3]-m[x1y2_3]*m[x2y1_3]) -
		m[x1y0_3]*(m[x0y1_3]*m[x2y2_3]-m[x0y2_3]*m[x2y1_3]) +
		m[x2y0_3]*(m[x0y1_3]*m[x1y2_3]-m[x0y2_3]*m[x1y1_3])
}

func (m Mat3) Determinant() Float {
	return Mat3Determinant(m)
}

func (m *Mat3) Inverse() {
	det := m.Determinant()
	if det == 0 {
		*m = Mat3Zero()
		return
	}
	invDet := 1.0 / det
	res := Mat3{
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

func (m Mat3) Inverted() Mat3 {
	res := m
	res.Inverse()
	return res
}

func (m Mat3) IsIdentity() bool {
	match := Mat3Identity()
	success := true
	for i := 0; i < len(match) && success; i++ {
		success = match[i] == m[i]
	}
	return success
}

func Mat3Approx(a, b Mat3) bool {
	res := true
	for i := range a {
		res = res && Abs(a[i]-b[i]) < FloatSmallestNonzero
	}
	return res
}

func Mat3ApproxTo(a, b Mat3, delta Float) bool {
	res := true
	for i := range a {
		res = res && Abs(a[i]-b[i]) < delta
	}
	return res
}

func (m Mat3) Equals(other Mat3) bool {
	return Mat3Approx(m, other)
}
