/******************************************************************************/
/* quaternion.go                                                              */
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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package matrix

import (
	"math"
)

type Quaternion [4]Float

func (q Quaternion) W() Float { return q[Qw] }
func (q Quaternion) X() Float { return q[Qx] }
func (q Quaternion) Y() Float { return q[Qy] }
func (q Quaternion) Z() Float { return q[Qz] }

func NewQuaternion(w, x, y, z Float) Quaternion {
	return Quaternion{w, x, y, z}
}

func QuaternionIdentity() Quaternion {
	return Quaternion{1, 0, 0, 0}
}

func QuaternionFromXYZW(xyzw [4]Float) Quaternion {
	return Quaternion{xyzw[3], xyzw[0], xyzw[1], xyzw[2]}
}

func QuaternionFromXYZWSlice(xyzw []Float) Quaternion {
	return Quaternion{xyzw[3], xyzw[0], xyzw[1], xyzw[2]}
}

func QuaternionFromArray(xyzw [4]Float) Quaternion {
	return Quaternion{xyzw[0], xyzw[1], xyzw[2], xyzw[3]}
}

func QuaternionFromSlice(xyzw []Float) Quaternion {
	return Quaternion{xyzw[0], xyzw[1], xyzw[2], xyzw[3]}
}

func QuaternionFromVec4(v Vec4) Quaternion {
	return Quaternion{v.X(), v.Y(), v.Z(), v.W()}
}

func QuaternionApprox(a, b Quaternion) bool {
	return Abs(a.W()-b.W()) < math.SmallestNonzeroFloat32 &&
		Abs(a.X()-b.X()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Y()-b.Y()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Z()-b.Z()) < math.SmallestNonzeroFloat32
}

func QuaternionFromMat4(m Mat4) Quaternion {
	m00 := m[x0y0]
	m10 := m[x1y0]
	m20 := m[x2y0]
	m01 := m[x0y1]
	m11 := m[x1y1]
	m21 := m[x2y1]
	m02 := m[x0y2]
	m12 := m[x1y2]
	m22 := m[x2y2]
	t := m00 + m11 + m22
	if t > 0 {
		s := 0.5 / Sqrt(t+Float(1.0))
		return Quaternion{Float(0.25) / s, (m12 - m21) * s, (m20 - m02) * s, (m01 - m10) * s}
	} else if m00 > m11 && m00 > m22 {
		s := 2.0 * Sqrt(1.0+m00-m11-m22)
		return Quaternion{(m12 - m21) / s, Float(0.25) * s, (m10 + m01) / s, (m20 + m02) / s}
	} else if m11 > m22 {
		s := 2.0 * Sqrt(1.0+m11-m00-m22)
		return Quaternion{(m20 - m02) / s, (m10 + m01) / s, Float(0.25) * s, (m21 + m12) / s}
	} else {
		s := 2.0 * Sqrt(1.0+m22-m00-m11)
		return Quaternion{(m01 - m10) / s, (m20 + m02) / s, (m21 + m12) / s, Float(0.25) * s}
	}
}

func (q Quaternion) ToMat4() Mat4 {
	out := Mat4Identity()
	sqw := q.W() * q.W()
	sqx := q.X() * q.X()
	sqy := q.Y() * q.Y()
	sqz := q.Z() * q.Z()
	invs := 1.0 / (sqx + sqy + sqz + sqw)
	out[x0y0] = (sqx - sqy - sqz + sqw) * invs
	out[x1y1] = (-sqx + sqy - sqz + sqw) * invs
	out[x2y2] = (-sqx - sqy + sqz + sqw) * invs
	tmp1 := q.X() * q.Y()
	tmp2 := q.Z() * q.W()
	out[x1y0] = 2.0 * (tmp1 + tmp2) * invs
	out[x0y1] = 2.0 * (tmp1 - tmp2) * invs
	tmp1 = q.X() * q.Z()
	tmp2 = q.Y() * q.W()
	out[x2y0] = 2.0 * (tmp1 - tmp2) * invs
	out[x0y2] = 2.0 * (tmp1 + tmp2) * invs
	tmp1 = q.Y() * q.Z()
	tmp2 = q.X() * q.W()
	out[x2y1] = 2.0 * (tmp1 + tmp2) * invs
	out[x1y2] = 2.0 * (tmp1 - tmp2) * invs
	return out
}

func QuaternionFromEuler(v Vec3) Quaternion {
	x := Deg2Rad(v.X())
	y := Deg2Rad(v.Y())
	z := Deg2Rad(v.Z())
	c1 := Cos(x / 2.0)
	c2 := Cos(y / 2.0)
	c3 := Cos(z / 2.0)
	s1 := Sin(x / 2.0)
	s2 := Sin(y / 2.0)
	s3 := Sin(z / 2.0)
	return Quaternion{
		c1*c2*c3 - s1*s2*s3,
		s1*c2*c3 + c1*s2*s3,
		c1*s2*c3 - s1*c2*s3,
		c1*c2*s3 + s1*s2*c3,
	}
}

func (q Quaternion) ToEuler() Vec3 {
	out := Vec3{}
	m := q.ToMat4()
	out[Vy] = Rad2Deg(Asin(Clamp(m[x0y2], -1.0, 1.0)))
	if Abs(m[x0y2]) < 0.9999999 {
		out.SetX(Rad2Deg(Atan2(-m[x1y2], m[x2y2])))
		out.SetZ(Rad2Deg(Atan2(-m[x0y1], m[x0y0])))
	} else {
		out.SetX(Rad2Deg(Atan2(m[x2y1], m[x1y1])))
		out.SetZ(0.0)
	}
	return out
}

func (q Quaternion) scale(scalar Float) Quaternion {
	return Quaternion{q.W() * scalar, q.X() * scalar, q.Y() * scalar, q.Z() * scalar}
}

func quaternionDot(q Quaternion, other Quaternion) Float {
	return q.W()*other.W() + q.X()*other.X() + q.Y()*other.Y() + q.Z()*other.Z()
}

func (q Quaternion) length() Float {
	return Sqrt(quaternionDot(q, q))
}

func (q *Quaternion) scaleAssign(scalar Float) {
	q[Qw] *= scalar
	q[Qx] *= scalar
	q[Qy] *= scalar
	q[Qz] *= scalar
}

func (q Quaternion) Normal() Quaternion {
	return q.scale(1.0 / q.length())
}

func (q *Quaternion) Normalize() {
	q.scaleAssign(1.0 / q.length())
}

func QuaternionLerp(from, to Quaternion, factor Float) Quaternion {
	var r Quaternion
	t := 1.0 - factor
	r[Qx] = t*from.X() + factor*to.X()
	r[Qy] = t*from.Y() + factor*to.Y()
	r[Qz] = t*from.Z() + factor*to.Z()
	r[Qw] = t*from.W() + factor*to.W()
	r.Normalize()
	return r
}

func QuaternionSlerp(from, to Quaternion, factor Float) Quaternion {
	if factor <= math.SmallestNonzeroFloat32 {
		return from
	} else if factor >= 1.0 {
		return to
	} else {
		var r Quaternion
		x := from.X()
		y := from.Y()
		z := from.Z()
		w := from.W()
		cosHalfTheta := w*to.W() + x*to.X() + y*to.Y() + z*to.Z()
		if cosHalfTheta < 0 {
			r[Qw] = -to.W()
			r[Qx] = -to.X()
			r[Qy] = -to.Y()
			r[Qz] = -to.Z()
			cosHalfTheta = -cosHalfTheta
		} else {
			r = to
		}
		if cosHalfTheta >= 1.0 {
			r[Qw] = w
			r[Qx] = x
			r[Qy] = y
			r[Qz] = z
			return r
		}
		sqrSinHalfTheta := 1.0 - cosHalfTheta*cosHalfTheta
		if sqrSinHalfTheta <= math.SmallestNonzeroFloat32 {
			s := 1.0 - factor
			r[Qw] = s*w + factor*r.W()
			r[Qx] = s*x + factor*r.X()
			r[Qy] = s*y + factor*r.Y()
			r[Qz] = s*z + factor*r.Z()
			r.Normalize()
			return r
		}
		sinHalfTheta := Sqrt(sqrSinHalfTheta)
		halfTheta := Atan2(sinHalfTheta, cosHalfTheta)
		ratioA := Sin((1.0-factor)*halfTheta) / sinHalfTheta
		ratioB := Sin(factor*halfTheta) / sinHalfTheta
		r[Qw] = w*ratioA + r.W()*ratioB
		r[Qx] = x*ratioA + r.X()*ratioB
		r[Qy] = y*ratioA + r.Y()*ratioB
		r[Qz] = z*ratioA + r.Z()*ratioB
		return r
	}
}

func QuaternionAxisAngle(axis Vec3, angle Float) Quaternion {
	cpy := axis.Scale(Sin(angle * 0.5))
	return Quaternion{Cos(angle * 0.5), cpy.X(), cpy.Y(), cpy.Z()}
}

func (q *Quaternion) Inverse() {
	d := q.W()*q.W() + q.X()*q.X() + q.Y()*q.Y() + q.Z()*q.Z()
	q[Qw] = q.W() / d
	q[Qx] = -q.X() / d
	q[Qy] = -q.Y() / d
	q[Qz] = -q.Z() / d
}

func (q *Quaternion) Conjugate() {
	q[Qx] = -q.X()
	q[Qy] = -q.Y()
	q[Qz] = -q.Z()
}

func (q Quaternion) Multiply(rhs Quaternion) Quaternion {
	return NewQuaternion(
		q.W()*rhs.W()-q.X()*rhs.X()-q.Y()*rhs.Y()-q.Z()*rhs.Z(),
		q.W()*rhs.X()+q.X()*rhs.W()+q.Y()*rhs.Z()-q.Z()*rhs.Y(),
		q.W()*rhs.Y()-q.X()*rhs.Z()+q.Y()*rhs.W()+q.Z()*rhs.X(),
		q.W()*rhs.Z()+q.X()*rhs.Y()-q.Y()*rhs.X()+q.Z()*rhs.W(),
	)
}

func (q *Quaternion) MultiplyAssign(rhs Quaternion) {
	q[Qw] = q.W()*rhs.W() - q.X()*rhs.X() - q.Y()*rhs.Y() - q.Z()*rhs.Z()
	q[Qx] = q.W()*rhs.X() + q.X()*rhs.W() + q.Y()*rhs.Z() - q.Z()*rhs.Y()
	q[Qy] = q.W()*rhs.Y() - q.X()*rhs.Z() + q.Y()*rhs.W() + q.Z()*rhs.X()
	q[Qz] = q.W()*rhs.Z() + q.X()*rhs.Y() - q.Y()*rhs.X() + q.Z()*rhs.W()
}

func (q Quaternion) MultiplyVec3(rhs Vec3) Vec3 {
	v0 := q.X() * 2.0
	v1 := q.Y() * 2.0
	v2 := q.Z() * 2.0
	v := [12]Float{
		v0, v1, v2,
		q.X() * v0,
		q.Y() * v1,
		q.Z() * v2,
		q.X() * v1,
		q.X() * v2,
		q.Y() * v2,
		q.W() * v0,
		q.W() * v1,
		q.W() * v2}
	return Vec3{
		(1.0-(v[4]+v[5]))*rhs.X() + (v[6]-v[11])*rhs.Y() + (v[7]+v[10])*rhs.Z(),
		(v[6]+v[11])*rhs.X() + (1.0-(v[3]+v[5]))*rhs.Y() + (v[8]-v[9])*rhs.Z(),
		(v[7]-v[10])*rhs.X() + (v[8]+v[9])*rhs.Y() + (1.0-(v[3]+v[4]))*rhs.Z()}
}

func (q *Quaternion) AddAssign(rhs Quaternion) {
	q[Qw] = q.W() + rhs.W()
	q[Qx] = q.X() + rhs.X()
	q[Qy] = q.Y() + rhs.Y()
	q[Qz] = q.Z() + rhs.Z()
}

func QuatAngleBetween(lhs, rhs Vec3) Quaternion {
	// It is important that the inputs are of equal length when
	// calculating the half-way vector.
	kCosTheta := Vec3Dot(lhs, rhs)
	k := Sqrt(Pow(lhs.Length(), 2) * Pow(rhs.Length(), 2))
	// TODO:  Approx here
	if kCosTheta/k == -1.0 {
		// 180 degree rotation around any orthogonal vector
		o := lhs.Orthogonal()
		oNorm := o.Normal()
		return Quaternion{0, oNorm.X(), oNorm.Y(), oNorm.Z()}
	}
	c := Vec3Cross(lhs, rhs)
	q := Quaternion{kCosTheta + k, c.X(), c.Y(), c.Z()}
	q.Normalize()
	return q
}

func QuaternionLookAt(from, to Vec3) Quaternion {
	diff := to.Subtract(from)
	direction := diff.Normal()
	back := Vec3Backward()
	dot := Vec3Dot(back, direction)
	if Abs(dot-(-1.0)) < 0.000001 {
		u := Vec3Up()
		return QuaternionAxisAngle(u, Rad2Deg(Float(math.Pi)))
	} else if Abs(dot-(1.0)) < 0.000001 {
		return QuaternionIdentity()
	}
	angle := -Rad2Deg(Acos(dot))
	cross := Vec3Cross(back, direction)
	nmlCross := cross.Normal()
	return QuaternionAxisAngle(nmlCross, angle)
}
