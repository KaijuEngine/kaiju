//go:build !amd64 && !arm64

/******************************************************************************/
/* matrix.none.go                                                             */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
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
