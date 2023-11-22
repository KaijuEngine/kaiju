package matrix

type Mat3 [9]Float

func NewMat3() Mat3 {
	return Mat3Identity()
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

func Mat3FromMat4(m Mat4) Mat3 {
	return Mat3{
		m[0], m[1], m[2],
		m[4], m[5], m[6],
		m[8], m[9], m[10],
	}
}

func (m Mat3) ToMat4() Mat4 {
	return Mat4{
		m[0], m[1], m[2], 0,
		m[3], m[4], m[5], 0,
		m[6], m[7], m[8], 0,
		0, 0, 0, 1,
	}
}

func (m *Mat3) Reset() {
	m[0] = 1
	m[1] = 0
	m[2] = 0
	m[3] = 0
	m[4] = 1
	m[5] = 0
	m[6] = 0
	m[7] = 0
	m[8] = 1
}

func (m Mat3) RowVector(row int) Vec3 {
	return Vec3{m[row*3+0], m[row*3+1], m[row*3+2]}
}

func (m Mat3) ColumnVector(col int) Vec3 {
	return Vec3{m[col+0], m[col+3], m[col+6]}
}

func (m Mat3) Multiply(rhs Mat3) Mat3 {
	return Mat3{
		m[0]*rhs[0] + m[1]*rhs[3] + m[2]*rhs[6],
		m[0]*rhs[1] + m[1]*rhs[4] + m[2]*rhs[7],
		m[0]*rhs[2] + m[1]*rhs[5] + m[2]*rhs[8],
		m[3]*rhs[0] + m[4]*rhs[3] + m[5]*rhs[6],
		m[3]*rhs[1] + m[4]*rhs[4] + m[5]*rhs[7],
		m[3]*rhs[2] + m[4]*rhs[5] + m[5]*rhs[8],
		m[6]*rhs[0] + m[7]*rhs[3] + m[8]*rhs[6],
		m[6]*rhs[1] + m[7]*rhs[4] + m[8]*rhs[7],
		m[6]*rhs[2] + m[7]*rhs[5] + m[8]*rhs[8],
	}
}

func (m *Mat3) MultiplyAssign(rhs Mat3) {
	m[0] = m[0]*rhs[0] + m[1]*rhs[3] + m[2]*rhs[6]
	m[1] = m[0]*rhs[1] + m[1]*rhs[4] + m[2]*rhs[7]
	m[2] = m[0]*rhs[2] + m[1]*rhs[5] + m[2]*rhs[8]
	m[3] = m[3]*rhs[0] + m[4]*rhs[3] + m[5]*rhs[6]
	m[4] = m[3]*rhs[1] + m[4]*rhs[4] + m[5]*rhs[7]
	m[5] = m[3]*rhs[2] + m[4]*rhs[5] + m[5]*rhs[8]
	m[6] = m[6]*rhs[0] + m[7]*rhs[3] + m[8]*rhs[6]
	m[7] = m[6]*rhs[1] + m[7]*rhs[4] + m[8]*rhs[7]
	m[8] = m[6]*rhs[2] + m[7]*rhs[5] + m[8]*rhs[8]
}

func (m Mat3) MultiplyVec3(v Vec3) Vec3 {
	return Vec3{
		m[0]*v[0] + m[1]*v[1] + m[2]*v[2],
		m[3]*v[0] + m[4]*v[1] + m[5]*v[2],
		m[6]*v[0] + m[7]*v[1] + m[8]*v[2],
	}
}
