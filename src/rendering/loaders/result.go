/******************************************************************************/
/* result.go                                                                  */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package loaders

import (
	"kaiju/collision"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering/loaders/load_result"
)

func TrySelectResultMesh(mesh *load_result.Mesh,
	e *engine.Entity, ray collision.Ray) (matrix.Float, bool) {

	const rayLen = 10000.0
	p, _, s := e.Transform.WorldTransform()
	rad := mesh.ScaledRadius(s)
	if ray.SphereHit(p, rad, rayLen) {
		mat := e.Transform.Matrix()
		for j := 0; j < len(mesh.Indexes); j += 3 {
			a := mat.TransformPoint(mesh.Verts[mesh.Indexes[j]].Position)
			b := mat.TransformPoint(mesh.Verts[mesh.Indexes[j+1]].Position)
			c := mat.TransformPoint(mesh.Verts[mesh.Indexes[j+2]].Position)
			if ray.TriangleHit(rayLen, a, b, c) {
				center := matrix.Vec3{
					(a.X() + b.X() + c.X()) / 3.0,
					(a.Y() + b.Y() + c.Y()) / 3.0,
					(a.Z() + b.Z() + c.Z()) / 3.0,
				}
				return center.Distance(ray.Origin), true
			}
		}
	}
	return 0, false
}
