package vector_graphics

import (
	"time"
)

type AnimatedValueType int

const (
	AnimatedValueTypeNone AnimatedValueType = iota
	AnimatedValueTypePosition
	AnimatedValueTypePositionX
	AnimatedValueTypePositionY
	AnimatedValueTypePositionZ
	AnimatedValueTypeRotation
	AnimatedValueTypeWidth
	AnimatedValueTypeHeight
	AnimatedValueTypeRadius
	AnimatedValueTypeRadiusX
	AnimatedValueTypeRadiusY
	AnimatedValueTypeColorR
	AnimatedValueTypeColorG
	AnimatedValueTypeColorB
	AnimatedValueTypeColorA
	AnimatedValueTypeFromX
	AnimatedValueTypeFromY
	AnimatedValueTypeFromZ
	AnimatedValueTypeToX
	AnimatedValueTypeToY
	AnimatedValueTypeToZ
	AnimatedValueTypeStrokeR
	AnimatedValueTypeStrokeG
	AnimatedValueTypeStrokeB
	AnimatedValueTypeStrokeA
	AnimatedValueTypeFillR
	AnimatedValueTypeFillG
	AnimatedValueTypeFillB
	AnimatedValueTypeFillA
	AnimatedValueTypeStrokeWidth
	AnimatedValueTypeOpacity
	AnimatedValueTypePolygonPoint
	// Any number greater than this point is the point index to animate, points
	// are 2-component vectors, so it alternates: p0.x, p0.y, p1.x, p1.y, etc...
)

type Animation struct {
	Target Shape
	Type   AnimatedValueType
	Keys   []AnimationKeyFrame
}

type AnimationKeyFrame struct {
	// TimeCode is represented in seconds from the start of the animation
	TimeCode float64
	Curve    PointCurve
	Value    float64
}

func (a *Animation) Animate(timeCode float64) {
	value := a.GetValueFromTime(timeCode)
	a.Target.Animate(a.Type, value)
}

func (a *Animation) Seconds() time.Duration {
	if len(a.Keys) == 0 {
		return 0
	}
	return time.Duration(a.Keys[len(a.Keys)-1].TimeCode)
}

// GetValueFromTime returns the interpolated value for the animation at the
// given timeCode (seconds). It uses cubic Bézier interpolation between the two
// surrounding keyframes. The curve handles (LeftCharacter / RightCharacter) are
// interpreted as offsets from the keyframe point (time, value).
func (a *Animation) GetValueFromTime(timeCode float64) float64 {
	// No keyframes – return zero.
	if len(a.Keys) == 0 {
		return 0
	}
	// Clamp to first / last keyframe values when out of range.
	if timeCode <= a.Keys[0].TimeCode {
		return a.Keys[0].Value
	}
	if timeCode >= a.Keys[len(a.Keys)-1].TimeCode {
		return a.Keys[len(a.Keys)-1].Value
	}
	// Find the segment [p, q] where p.TimeCode <= timeCode <= q.TimeCode.
	var p, q AnimationKeyFrame
	for i := 0; i < len(a.Keys)-1; i++ {
		if a.Keys[i].TimeCode <= timeCode && timeCode <= a.Keys[i+1].TimeCode {
			p = a.Keys[i]
			q = a.Keys[i+1]
			break
		}
	}
	// Bézier control points (x = time, y = value).
	// Offsets are taken from the keyframe's Curve vectors.
	p0x, p0y := p.TimeCode, p.Value
	p1x := p.TimeCode + float64(p.Curve.RightCharacter.X())
	p1y := p.Value + float64(p.Curve.RightCharacter.Y())
	p2x := q.TimeCode + float64(q.Curve.LeftCharacter.X())
	p2y := q.Value + float64(q.Curve.LeftCharacter.Y())
	p3x, p3y := q.TimeCode, q.Value
	// Helper to evaluate cubic Bézier X (or Y) at parameter u.
	bezier := func(u, x0, x1, x2, x3 float64) float64 {
		// (1-u)^3*x0 + 3*(1-u)^2*u*x1 + 3*(1-u)*u*u*x2 + u^3*x3
		om := 1 - u
		return om*om*om*x0 + 3*om*om*u*x1 + 3*om*u*u*x2 + u*u*u*x3
	}
	// Derivative of cubic Bézier X with respect to u.
	bezierDeriv := func(u, x0, x1, x2, x3 float64) float64 {
		om := 1 - u
		return 3*om*om*(x1-x0) + 6*om*u*(x2-x1) + 3*u*u*(x3-x2)
	}
	// Solve for u such that Bézier X(u) == timeCode using Newton‑Raphson.
	// Initial guess based on linear interpolation.
	u := (timeCode - p0x) / (p3x - p0x)
	if u < 0 {
		u = 0
	} else if u > 1 {
		u = 1
	}
	const maxIter = 10
	const epsilon = 1e-5
	for i := 0; i < maxIter; i++ {
		bx := bezier(u, p0x, p1x, p2x, p3x)
		diff := bx - timeCode
		if diff < 0 {
			diff = -diff
		}
		if diff < epsilon {
			break
		}
		dbx := bezierDeriv(u, p0x, p1x, p2x, p3x)
		if dbx == 0 {
			break
		}
		u = u - (bx-timeCode)/dbx
		if u < 0 {
			u = 0
		} else if u > 1 {
			u = 1
		}
	}
	// Interpolated value is Bézier Y at the solved u.
	value := bezier(u, p0y, p1y, p2y, p3y)
	return value
}
