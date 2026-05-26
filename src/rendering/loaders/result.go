/******************************************************************************/
/* result.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package loaders

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/loaders/load_result"
)

func TrySelectResultMesh(mesh *load_result.Mesh,
	e *engine.Entity, ray graviton.Ray) (matrix.Float, bool) {
	const rayLen = 10000.0
	p, _, s := e.Transform.WorldTransform()
	rad := mesh.ScaledRadius(s)
	if ray.SphereHit(p, rad, rayLen) {
		mat := e.Transform.Matrix()
		if !mat.Equals(matrix.Mat4Identity()) {
			for j := 0; j < len(mesh.Indexes); j += 3 {
				a := mat.TransformPoint(mesh.Verts[mesh.Indexes[j]].Position)
				b := mat.TransformPoint(mesh.Verts[mesh.Indexes[j+1]].Position)
				c := mat.TransformPoint(mesh.Verts[mesh.Indexes[j+2]].Position)
				if _, ok := ray.TriangleHit(rayLen, a, b, c); ok {
					center := matrix.Vec3{
						(a.X() + b.X() + c.X()) / 3.0,
						(a.Y() + b.Y() + c.Y()) / 3.0,
						(a.Z() + b.Z() + c.Z()) / 3.0,
					}
					return center.Distance(ray.Origin), true
				}
			}
		}
	}
	return 0, false
}
