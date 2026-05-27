/******************************************************************************/
/* transform.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type fbxBasisConverter struct {
	settings GlobalSettings
}

func newFBXBasisConverter(settings GlobalSettings) fbxBasisConverter {
	return fbxBasisConverter{settings: settings}
}

func (c fbxBasisConverter) ConvertPosition(v matrix.Vec3) matrix.Vec3 {
	return matrix.NewVec3(
		-fbxAxisValue(v, c.settings.CoordAxis, c.settings.CoordAxisSign),
		fbxAxisValue(v, c.settings.UpAxis, c.settings.UpAxisSign),
		-fbxAxisValue(v, c.settings.FrontAxis, c.settings.FrontAxisSign),
	)
}

func (c fbxBasisConverter) ConvertDirection(v matrix.Vec3) matrix.Vec3 {
	converted := c.ConvertPosition(v)
	if converted.IsZero() {
		return converted
	}
	return converted.Normal()
}

func (c fbxBasisConverter) ConvertScale(v matrix.Vec3) matrix.Vec3 {
	return matrix.NewVec3(
		fbxAxisScale(v, c.settings.CoordAxis),
		fbxAxisScale(v, c.settings.UpAxis),
		fbxAxisScale(v, c.settings.FrontAxis),
	)
}

func (c fbxBasisConverter) ConvertRotation(degrees matrix.Vec3) matrix.Quaternion {
	converted := c.ConvertPosition(degrees)
	return matrix.QuaternionFromEuler(converted)
}

func fbxAxisValue(v matrix.Vec3, axis int, sign int) matrix.Float {
	if sign == 0 {
		sign = 1
	}
	switch axis {
	case 0:
		return v.X() * matrix.Float(sign)
	case 1:
		return v.Y() * matrix.Float(sign)
	case 2:
		return v.Z() * matrix.Float(sign)
	default:
		return 0
	}
}

func fbxAxisScale(v matrix.Vec3, axis int) matrix.Float {
	switch axis {
	case 0:
		return v.X()
	case 1:
		return v.Y()
	case 2:
		return v.Z()
	default:
		return 1
	}
}

func fbxPropertyVec3(props PropertyTable, name string, fallback matrix.Vec3) matrix.Vec3 {
	value, ok := props.Vec3(name)
	if !ok {
		return fallback
	}
	return matrix.NewVec3(matrix.Float(value[0]), matrix.Float(value[1]), matrix.Float(value[2]))
}

func fbxGeometricTRS(model *Object, converter fbxBasisConverter, unitScale matrix.Float) (matrix.Vec3, matrix.Vec3, matrix.Vec3) {
	if model == nil {
		return matrix.Vec3Zero(), matrix.Vec3Zero(), matrix.Vec3One()
	}
	translation := fbxPropertyVec3(model.Properties, "GeometricTranslation", matrix.Vec3Zero())
	translation = converter.ConvertPosition(translation.Scale(unitScale))
	rotation := fbxPropertyVec3(model.Properties, "GeometricRotation", matrix.Vec3Zero())
	rotation = converter.ConvertPosition(rotation)
	scale := converter.ConvertScale(fbxPropertyVec3(model.Properties, "GeometricScaling", matrix.Vec3One()))
	return translation, rotation, scale
}

func bakeGeometricTransform(verts []rendering.Vertex, model *Object, converter fbxBasisConverter, unitScale matrix.Float) {
	translation, rotation, scale := fbxGeometricTRS(model, converter, unitScale)
	if translation.IsZero() && rotation.IsZero() && scale.Equals(matrix.Vec3One()) {
		return
	}
	pointMat := matrix.Mat4Identity()
	pointMat.Scale(scale)
	pointMat.Rotate(rotation)
	pointMat.Translate(translation)
	normalMat := matrix.Mat4Identity()
	normalMat.Scale(fbxInverseScale(scale))
	normalMat.Rotate(rotation)
	for i := range verts {
		verts[i].Position = pointMat.TransformPoint(verts[i].Position)
		verts[i].MorphTarget = pointMat.TransformPoint(verts[i].MorphTarget)
		verts[i].Normal = transformDirection(normalMat, verts[i].Normal)
		if verts[i].Tangent.X() != 0 || verts[i].Tangent.Y() != 0 || verts[i].Tangent.Z() != 0 {
			tangent := transformDirection(normalMat, matrix.NewVec3(
				verts[i].Tangent.X(),
				verts[i].Tangent.Y(),
				verts[i].Tangent.Z(),
			))
			verts[i].Tangent.SetX(tangent.X())
			verts[i].Tangent.SetY(tangent.Y())
			verts[i].Tangent.SetZ(tangent.Z())
		}
	}
}

func fbxInverseScale(scale matrix.Vec3) matrix.Vec3 {
	inv := matrix.Vec3One()
	if scale.X() != 0 {
		inv.SetX(1 / scale.X())
	}
	if scale.Y() != 0 {
		inv.SetY(1 / scale.Y())
	}
	if scale.Z() != 0 {
		inv.SetZ(1 / scale.Z())
	}
	return inv
}

func transformDirection(transform matrix.Mat4, direction matrix.Vec3) matrix.Vec3 {
	if direction.IsZero() {
		return direction
	}
	out := matrix.Mat4MultiplyVec4(transform, direction.AsVec4WithW(0)).AsVec3()
	if out.IsZero() {
		return out
	}
	return out.Normal()
}
