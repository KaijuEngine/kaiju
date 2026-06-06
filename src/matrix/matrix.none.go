//go:build !amd64 && !arm64

/******************************************************************************/
/* matrix.none.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

func Mat4Multiply(a, b Mat4) Mat4 {
	var result Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var sum float32 = 0
			for k := 0; k < 4; k++ {
				sum += a[i*4+k] * b[k*4+j]
			}
			result[i*4+j] = sum
		}
	}
	return result
}

func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4 {
	var result Vec4
	c := a.ColumnVector(0)
	result[Vx] = Vec4Dot(c, b)
	c = a.ColumnVector(1)
	result[Vy] = Vec4Dot(c, b)
	c = a.ColumnVector(2)
	result[Vz] = Vec4Dot(c, b)
	c = a.ColumnVector(3)
	result[Vw] = Vec4Dot(c, b)
	return result
}

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
