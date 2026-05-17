/******************************************************************************/
/* query.go                                                                   */
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

package graviton

import "kaijuengine.com/matrix"

type Hit struct {
	Body     *RigidBody
	Point    matrix.Vec3
	Normal   matrix.Vec3
	Distance matrix.Float
}

func (s *System) Raycast(from, to matrix.Vec3) (Hit, bool) {
	rayDelta := to.Subtract(from)
	length := rayDelta.Length()
	if length <= contactEpsilon {
		return Hit{}, false
	}
	ray := Ray{
		Origin:    from,
		Direction: rayDelta.Scale(1.0 / length),
	}
	closest := Hit{Distance: matrix.Inf(1)}
	found := false
	s.bodies.Each(func(body *RigidBody) {
		if body == nil || !body.Active {
			return
		}
		if _, ok := raycastAABB(ray, body.WorldAABB(), length); !ok {
			return
		}
		hit, ok := raycastBody(ray, body, length)
		if !ok || hit.Distance >= closest.Distance {
			return
		}
		hit.Body = body
		closest = hit
		found = true
	})
	if !found {
		return Hit{}, false
	}
	return closest, true
}

func raycastBody(ray Ray, body *RigidBody, length matrix.Float) (Hit, bool) {
	if body == nil {
		return Hit{}, false
	}
	if body.Collision.Shape.Type == ShapeTypeMesh {
		return body.Collision.Mesh.Raycast(ray, length, &body.Transform)
	}
	if body.Collision.Shape.Type == ShapeTypeTerrain {
		return body.Collision.Terrain.Raycast(ray, length, &body.Transform)
	}
	return raycastShape(ray, worldShape(body), length)
}

func (s *System) SphereSweep(from, to matrix.Vec3, radius matrix.Float) (Hit, bool) {
	if radius < 0 {
		return Hit{}, false
	}
	rayDelta := to.Subtract(from)
	length := rayDelta.Length()
	rayDirection := matrix.Vec3Right()
	if length > contactEpsilon {
		rayDirection = rayDelta.Scale(1.0 / length)
	}
	ray := Ray{
		Origin:    from,
		Direction: rayDirection,
	}
	closest := Hit{Distance: matrix.Inf(1)}
	found := false
	s.bodies.Each(func(body *RigidBody) {
		if body == nil || !body.Active {
			return
		}
		shape := worldShape(body)
		if shape.Type == ShapeTypeMesh {
			return
		}
		if hit, ok := sphereSweepStartOverlap(from, radius, shape, rayDirection); ok {
			hit.Body = body
			if !found || hit.Distance < closest.Distance {
				closest = hit
				found = true
			}
			return
		}
		if length <= contactEpsilon {
			return
		}
		if _, ok := raycastAABB(ray, expandAABB(body.WorldAABB(), radius), length); !ok {
			return
		}
		hit, ok := sphereSweepShape(ray, shape, length, radius)
		if !ok || hit.Distance >= closest.Distance {
			return
		}
		hit.Body = body
		closest = hit
		found = true
	})
	if !found {
		return Hit{}, false
	}
	return closest, true
}

func raycastShape(ray Ray, shape Shape, length matrix.Float) (Hit, bool) {
	switch shape.Type {
	case ShapeTypeSphere:
		return raycastSphere(ray, Sphere(shape), length)
	case ShapeTypeAABB:
		return raycastAABB(ray, AABB(shape), length)
	case ShapeTypeOOBB:
		return raycastOOBB(ray, OOBB(shape), length)
	case ShapeTypeCapsule:
		return raycastCapsule(ray, Capsule(shape), length)
	case ShapeTypeCylinder:
		return raycastCylinder(ray, Cylinder(shape), length)
	case ShapeTypeCone:
		return raycastCone(ray, Cone(shape), length)
	default:
		return Hit{}, false
	}
}

func sphereSweepShape(ray Ray, shape Shape, length, radius matrix.Float) (Hit, bool) {
	switch shape.Type {
	case ShapeTypeSphere:
		sphere := Sphere(shape)
		sphere.Radius += radius
		return sphereSweepFromExpandedRaycast(ray, Shape(sphere), length, radius)
	case ShapeTypeAABB:
		box := expandAABB(AABB(shape), radius)
		return sphereSweepAABB(ray, box, length, radius)
	case ShapeTypeOOBB:
		box := OOBB(shape)
		box.Extent = box.Extent.Add(matrix.NewVec3XYZ(radius))
		return sphereSweepOOBB(ray, box, length, radius)
	case ShapeTypeCapsule:
		capsule := Capsule(shape)
		capsule.Radius += radius
		return sphereSweepFromExpandedRaycast(ray, Shape(capsule), length, radius)
	case ShapeTypeMesh:
		return Hit{}, false
	default:
		return sphereSweepAABB(ray, expandAABB(shapeWorldAABB(shape), radius), length, radius)
	}
}

func sphereSweepStartOverlap(center matrix.Vec3, radius matrix.Float, shape Shape, sweepDirection matrix.Vec3) (Hit, bool) {
	sweepSphere := Shape{}
	sweepSphere.SetSphere(center, radius)
	contact, ok := collideShapes(sweepSphere, shape)
	if !ok {
		return Hit{}, false
	}
	normal := safeNormal(contact.Normal.Negative(), sweepDirection.Negative())
	return Hit{
		Point:    center.Subtract(normal.Scale(radius)),
		Normal:   normal,
		Distance: 0,
	}, true
}

func sphereSweepFromExpandedRaycast(ray Ray, shape Shape, length, radius matrix.Float) (Hit, bool) {
	hit, ok := raycastShape(ray, shape, length)
	if !ok {
		return Hit{}, false
	}
	hit.Point = sphereSweepContactPoint(ray.Point(hit.Distance), hit.Normal, radius)
	return hit, true
}

func sphereSweepAABB(ray Ray, box AABB, length, radius matrix.Float) (Hit, bool) {
	hit, ok := raycastAABB(ray, box, length)
	if !ok {
		return Hit{}, false
	}
	hit.Point = sphereSweepContactPoint(ray.Point(hit.Distance), hit.Normal, radius)
	return hit, true
}

func sphereSweepOOBB(ray Ray, box OOBB, length, radius matrix.Float) (Hit, bool) {
	hit, ok := raycastOOBB(ray, box, length)
	if !ok {
		return Hit{}, false
	}
	hit.Point = sphereSweepContactPoint(ray.Point(hit.Distance), hit.Normal, radius)
	return hit, true
}

func sphereSweepContactPoint(center, normal matrix.Vec3, radius matrix.Float) matrix.Vec3 {
	return center.Subtract(normal.Scale(radius))
}

func expandAABB(box AABB, radius matrix.Float) AABB {
	box.Extent = box.Extent.Add(matrix.NewVec3XYZ(radius))
	return box
}

func raycastSphere(ray Ray, sphere Sphere, length matrix.Float) (Hit, bool) {
	ok, distance := sphere.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	return Hit{
		Point:    point,
		Normal:   safeNormal(point.Subtract(sphere.Center), ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}

func raycastAABB(ray Ray, box AABB, length matrix.Float) (Hit, bool) {
	point, ok := box.RayHit(ray)
	if !ok {
		return Hit{}, false
	}
	distance := point.Distance(ray.Origin)
	if distance > length {
		return Hit{}, false
	}
	normal, _ := closestAABBFaceNormal(point, box)
	if distance <= contactEpsilon {
		normal = ray.Direction.Negative()
	}
	return Hit{
		Point:    point,
		Normal:   safeNormal(normal, ray.Direction.Negative()),
		Distance: distance,
	}, true
}

func raycastOOBB(ray Ray, box OOBB, length matrix.Float) (Hit, bool) {
	inverseOrientation := box.Orientation.Transpose()
	localRay := Ray{
		Origin:    inverseOrientation.MultiplyVec3(ray.Origin.Subtract(box.Center)),
		Direction: inverseOrientation.MultiplyVec3(ray.Direction),
	}
	localBox := NewAABB(matrix.Vec3Zero(), box.Extent)
	localPoint, ok := localBox.RayHit(localRay)
	if !ok {
		return Hit{}, false
	}
	distance := localPoint.Distance(localRay.Origin)
	if distance > length {
		return Hit{}, false
	}
	localNormal, _ := closestAABBFaceNormal(localPoint, localBox)
	if distance <= contactEpsilon {
		localNormal = localRay.Direction.Negative()
	}
	return Hit{
		Point:    box.Orientation.MultiplyVec3(localPoint).Add(box.Center),
		Normal:   safeNormal(box.Orientation.MultiplyVec3(localNormal), ray.Direction.Negative()),
		Distance: distance,
	}, true
}

func raycastCapsule(ray Ray, capsule Capsule, length matrix.Float) (Hit, bool) {
	ok, distance := capsule.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	a, b := capsuleSegment(capsule)
	closest := closestPointOnSegment(point, a, b)
	return Hit{
		Point:    point,
		Normal:   safeNormal(point.Subtract(closest), ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}

func raycastCylinder(ray Ray, cylinder Cylinder, length matrix.Float) (Hit, bool) {
	ok, distance := cylinder.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	direction := safeNormal(cylinder.Direction, matrix.Vec3Up())
	centerToPoint := point.Subtract(cylinder.Center)
	axisPoint := cylinder.Center.Add(direction.Scale(matrix.Vec3Dot(centerToPoint, direction)))
	return Hit{
		Point:    point,
		Normal:   safeNormal(point.Subtract(axisPoint), ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}

func raycastCone(ray Ray, cone Cone, length matrix.Float) (Hit, bool) {
	ok, distance := cone.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	direction := safeNormal(cone.Direction, matrix.Vec3Up())
	halfHeight := cone.Height * 0.5
	baseCenter := cone.Center.Add(direction.Scale(halfHeight))
	toBasePlane := matrix.Abs(matrix.Vec3Dot(point.Subtract(baseCenter), direction))
	normal := point.Subtract(cone.Center)
	if toBasePlane <= contactEpsilon {
		normal = direction
	}
	return Hit{
		Point:    point,
		Normal:   safeNormal(normal, ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}
