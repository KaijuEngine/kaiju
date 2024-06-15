//go:build !amd64

package matrix

// TODO:  Convert this to use NEON instructions

func Vec4MultiplyMat4(v Vec4, m Mat4) Vec4 {
	var result Vec4
	row := m.RowVector(0)
	result[Vx] = Vec4Dot(row, v)
	row = m.RowVector(1)
	result[Vy] = Vec4Dot(row, v)
	row = m.RowVector(2)
	result[Vz] = Vec4Dot(row, v)
	row = m.RowVector(3)
	result[Vw] = Vec4Dot(row, v)
	return result
}
