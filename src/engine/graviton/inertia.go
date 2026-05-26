/******************************************************************************/
/* inertia.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package graviton

import "kaijuengine.com/matrix"

func CalculateLocalInertia(shape Shape, mass matrix.Float) matrix.Vec3 {
	if mass <= 0 {
		return matrix.Vec3Zero()
	}
	switch shape.Type {
	case ShapeTypeSphere:
		return calculateSphereInertia(shape.Radius, mass)
	case ShapeTypeAABB, ShapeTypeOOBB:
		return calculateBoxInertia(shape.Extent, mass)
	case ShapeTypeCapsule:
		return calculateCapsuleInertia(shape.Radius, shape.Height, shape.Direction, mass)
	case ShapeTypeCylinder:
		return calculateCylinderInertia(shape.Radius, shape.Height, shape.Direction, mass)
	case ShapeTypeCone:
		return calculateConeInertia(shape.Radius, shape.Height, shape.Direction, mass)
	case ShapeTypeMesh:
		return calculateBoxInertia(shape.Extent, mass)
	default:
		return matrix.Vec3Zero()
	}
}

func calculateSphereInertia(radius, mass matrix.Float) matrix.Vec3 {
	radius = matrix.Abs(radius)
	if radius <= 0 {
		return matrix.Vec3Zero()
	}
	inertia := matrix.Float(0.4) * mass * radius * radius
	return matrix.NewVec3(inertia, inertia, inertia)
}

func calculateBoxInertia(extent matrix.Vec3, mass matrix.Float) matrix.Vec3 {
	extent = extent.Abs()
	if extent.IsZero() {
		return matrix.Vec3Zero()
	}
	return matrix.NewVec3(
		(mass/3)*(extent.Y()*extent.Y()+extent.Z()*extent.Z()),
		(mass/3)*(extent.X()*extent.X()+extent.Z()*extent.Z()),
		(mass/3)*(extent.X()*extent.X()+extent.Y()*extent.Y()),
	)
}

func calculateCapsuleInertia(radius, height matrix.Float, direction matrix.Vec3, mass matrix.Float) matrix.Vec3 {
	radius = matrix.Abs(radius)
	height = matrix.Abs(height)
	if radius <= 0 {
		return matrix.Vec3Zero()
	}
	if height <= 0 {
		return calculateSphereInertia(radius, mass)
	}
	cylinderVolume := radius * radius * height
	sphereVolume := matrix.Float(4.0/3.0) * radius * radius * radius
	totalVolume := cylinderVolume + sphereVolume
	cylinderMass := mass * cylinderVolume / totalVolume
	sphereMass := mass - cylinderMass
	radiusSquared := radius * radius
	heightSquared := height * height
	axial := (matrix.Float(0.5) * cylinderMass * radiusSquared) +
		(matrix.Float(0.4) * sphereMass * radiusSquared)
	radial := (cylinderMass/matrix.Float(12))*(3*radiusSquared+heightSquared) +
		(matrix.Float(0.4) * sphereMass * radiusSquared) +
		(sphereMass * heightSquared / 4)
	return axisymmetricInertia(direction, axial, radial)
}

func calculateCylinderInertia(radius, height matrix.Float, direction matrix.Vec3, mass matrix.Float) matrix.Vec3 {
	radius = matrix.Abs(radius)
	height = matrix.Abs(height)
	if radius <= 0 || height <= 0 {
		return matrix.Vec3Zero()
	}
	radiusSquared := radius * radius
	heightSquared := height * height
	axial := matrix.Float(0.5) * mass * radiusSquared
	radial := (mass / matrix.Float(12)) * (3*radiusSquared + heightSquared)
	return axisymmetricInertia(direction, axial, radial)
}

func calculateConeInertia(radius, height matrix.Float, direction matrix.Vec3, mass matrix.Float) matrix.Vec3 {
	radius = matrix.Abs(radius)
	height = matrix.Abs(height)
	if radius <= 0 || height <= 0 {
		return matrix.Vec3Zero()
	}
	radiusSquared := radius * radius
	heightSquared := height * height
	axial := matrix.Float(0.3) * mass * radiusSquared
	radial := (matrix.Float(0.15) * mass * radiusSquared) +
		(matrix.Float(0.1) * mass * heightSquared)
	return axisymmetricInertia(direction, axial, radial)
}

func axisymmetricInertia(direction matrix.Vec3, axial, radial matrix.Float) matrix.Vec3 {
	axis := safeNormal(direction, matrix.Vec3Up())
	difference := axial - radial
	return matrix.NewVec3(
		radial+difference*axis.X()*axis.X(),
		radial+difference*axis.Y()*axis.Y(),
		radial+difference*axis.Z()*axis.Z(),
	)
}
