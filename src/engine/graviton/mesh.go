/******************************************************************************/
/* mesh.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

// MeshCollision stores the heavy collision data for triangle meshes. The flat
// Shape keeps only type and bounds so primitive shape data stays compact.
type MeshCollision struct {
	Triangles []DetailedTriangle
	BVH       *BVH
	Bounds    AABB
}

func (s *Shape) SetMesh(bounds AABB) {
	s.Type = ShapeTypeMesh
	s.Center = bounds.Center
	s.Extent = bounds.Extent
}

func NewMeshShape(bounds AABB) Shape {
	s := Shape{}
	s.SetMesh(bounds)
	return s
}

func NewMeshCollision(triangles []DetailedTriangle) *MeshCollision {
	mesh := &MeshCollision{
		Triangles: append([]DetailedTriangle(nil), triangles...),
	}
	if len(mesh.Triangles) == 0 {
		mesh.Bounds = NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero())
		mesh.BVH = nil
		return mesh
	}
	entries := make([]HitObject, len(mesh.Triangles))
	for i := range mesh.Triangles {
		entries[i] = mesh.Triangles[i]
	}
	mesh.Bounds = AABBFromTriangles(mesh.Triangles)
	mesh.BVH = NewBVH(entries, nil, nil)
	return mesh
}

func NewMeshCollisionFromVertices(vertices []matrix.Vec3, indexes []uint32) *MeshCollision {
	if len(indexes) == 0 {
		triangles := make([]DetailedTriangle, 0, len(vertices)/3)
		for i := 0; i+2 < len(vertices); i += 3 {
			triangles = append(triangles, DetailedTriangleFromPoints([3]matrix.Vec3{
				vertices[i],
				vertices[i+1],
				vertices[i+2],
			}))
		}
		return NewMeshCollision(triangles)
	}
	triCount := len(indexes) / 3
	triangles := make([]DetailedTriangle, 0, triCount)
	for i := 0; i+2 < len(indexes); i += 3 {
		i0, i1, i2 := int(indexes[i]), int(indexes[i+1]), int(indexes[i+2])
		if i0 < 0 || i0 >= len(vertices) ||
			i1 < 0 || i1 >= len(vertices) ||
			i2 < 0 || i2 >= len(vertices) {
			continue
		}
		triangles = append(triangles, DetailedTriangleFromPoints([3]matrix.Vec3{
			vertices[i0],
			vertices[i1],
			vertices[i2],
		}))
	}
	return NewMeshCollision(triangles)
}

func AABBFromTriangles(triangles []DetailedTriangle) AABB {
	if len(triangles) == 0 {
		return NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero())
	}
	bounds := triangles[0].Bounds()
	for i := 1; i < len(triangles); i++ {
		bounds = AABBUnion(bounds, triangles[i].Bounds())
	}
	return bounds
}

func (m *MeshCollision) ForEachWorldTriangle(transform *matrix.Transform, visit func(DetailedTriangle) bool) {
	if m == nil || visit == nil {
		return
	}
	for i := range m.Triangles {
		tri := m.Triangles[i]
		if transform != nil {
			wm := transform.WorldMatrix()
			tri = DetailedTriangleFromPoints([3]matrix.Vec3{
				wm.TransformPoint(tri.Points[0]),
				wm.TransformPoint(tri.Points[1]),
				wm.TransformPoint(tri.Points[2]),
			})
		}
		if !visit(tri) {
			return
		}
	}
}

func (m *MeshCollision) Raycast(ray Ray, length matrix.Float, transform *matrix.Transform) (Hit, bool) {
	if m == nil || m.BVH == nil || length <= contactEpsilon {
		return Hit{}, false
	}
	localRay := ray
	localLength := length
	if transform != nil {
		inv := transform.InverseWorldMatrix()
		localOrigin := inv.TransformPoint(ray.Origin)
		localEnd := inv.TransformPoint(ray.Point(matrix.Float(length)))
		localDelta := localEnd.Subtract(localOrigin)
		localLength = localDelta.Length()
		if localLength <= contactEpsilon {
			return Hit{}, false
		}
		localRay = Ray{
			Origin:    localOrigin,
			Direction: localDelta.Scale(1.0 / localLength),
		}
	}
	data, localPoint, ok := m.BVH.RayIntersect(localRay, matrix.Float(localLength))
	if !ok {
		return Hit{}, false
	}
	point := localPoint
	if transform != nil {
		point = transform.WorldMatrix().TransformPoint(localPoint)
	}
	distance := point.Distance(ray.Origin)
	if distance > length {
		return Hit{}, false
	}
	normal := ray.Direction.Negative()
	if tri, ok := data.(DetailedTriangle); ok {
		normal = meshTriangleWorldNormal(tri, transform, ray.Direction)
	}
	return Hit{
		Point:    point,
		Normal:   normal,
		Distance: distance,
	}, true
}

func meshTriangleWorldNormal(tri DetailedTriangle, transform *matrix.Transform, rayDirection matrix.Vec3) matrix.Vec3 {
	p0, p1, p2 := tri.Points[0], tri.Points[1], tri.Points[2]
	if transform != nil {
		wm := transform.WorldMatrix()
		p0 = wm.TransformPoint(p0)
		p1 = wm.TransformPoint(p1)
		p2 = wm.TransformPoint(p2)
	}
	normal := matrix.Vec3Cross(p1.Subtract(p0), p2.Subtract(p0))
	normal = safeNormal(normal, rayDirection.Negative())
	if matrix.Vec3Dot(normal, rayDirection) > 0 {
		normal = normal.Negative()
	}
	return normal
}
