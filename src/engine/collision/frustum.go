/******************************************************************************/
/* frustum.go                                                                 */
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

package collision

import "kaiju/matrix"

type FrustumCorners [8]matrix.Vec4

type Frustum struct {
	Planes [6]Plane
}

func (f *Frustum) ExtractPlanes(vp matrix.Mat4) {
	// Left
	f.Planes[0].Normal = matrix.Vec3{vp[3] + vp[0], vp[7] + vp[4], vp[11] + vp[8]}
	f.Planes[0].Dot = vp[15] + vp[12]
	// Right
	f.Planes[1].Normal = matrix.Vec3{vp[3] - vp[0], vp[7] - vp[4], vp[11] - vp[8]}
	f.Planes[1].Dot = vp[15] - vp[12]
	// Bottom
	f.Planes[2].Normal = matrix.Vec3{vp[3] + vp[1], vp[7] + vp[5], vp[11] + vp[9]}
	f.Planes[2].Dot = vp[15] + vp[13]
	// Top
	f.Planes[3].Normal = matrix.Vec3{vp[3] - vp[1], vp[7] - vp[5], vp[11] - vp[9]}
	f.Planes[3].Dot = vp[15] - vp[13]
	// Near
	f.Planes[4].Normal = matrix.Vec3{vp[3] + vp[2], vp[7] + vp[6], vp[11] + vp[10]}
	f.Planes[4].Dot = vp[15] + vp[14]
	// Far
	f.Planes[5].Normal = matrix.Vec3{vp[3] - vp[2], vp[7] - vp[6], vp[11] - vp[10]}
	f.Planes[5].Dot = vp[15] - vp[14]
}

func FrustumExtractCorners(view, projection matrix.Mat4) FrustumCorners {
	vp := matrix.Mat4Multiply(view, projection)
	inv := vp
	inv.Inverse()
	ndcCorners := [8]matrix.Vec4{
		{-1, -1, 0, 1}, // Near plane corners
		{+1, -1, 0, 1},
		{+1, +1, 0, 1},
		{-1, +1, 0, 1},
		{-1, -1, 1, 1}, // Far plane corners
		{+1, -1, 1, 1},
		{+1, +1, 1, 1},
		{-1, +1, 1, 1},
	}
	var corners FrustumCorners
	for i, ndc := range ndcCorners {
		worldH := matrix.Mat4MultiplyVec4(inv, ndc)
		if worldH.W() != 0 {
			invW := 1.0 / worldH.W()
			corners[i] = matrix.NewVec4(
				worldH.X()*invW,
				worldH.Y()*invW,
				worldH.Z()*invW,
				1,
			)
		} else {
			corners[i] = worldH
		}
	}
	return corners
}

func (c FrustumCorners) Center() matrix.Vec3 {
	center := matrix.Vec3Zero()
	for i := range c {
		center.AddAssign(c[i].AsVec3())
	}
	center.ShrinkAssign(matrix.Float(len(c)))
	return center
}
