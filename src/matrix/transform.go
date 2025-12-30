/******************************************************************************/
/* transform.go                                                               */
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

import (
	"kaiju/klib"
	"kaiju/platform/concurrent"
	"kaiju/platform/profiler/tracing"
	"slices"
)

const (
	TransformWorkGroup      = "transform"
	TransformResetWorkGroup = "transformReset"
)

type Transform struct {
	localMatrix               Mat4
	worldMatrix               Mat4
	parent                    *Transform
	children                  []*Transform
	workGroup                 *concurrent.WorkGroup
	position, rotation, scale Vec3
	relativePosition          Vec3
	framePosition             Vec3
	frameRotation             Vec3
	frameScale                Vec3
	isDirty                   bool
	frameDirty                bool
	orderedChildren           bool
}

func (t *Transform) setup() {
	t.localMatrix = Mat4Identity()
	t.worldMatrix = Mat4Identity()
	t.position = Vec3Zero()
	t.rotation = Vec3Zero()
	t.scale = Vec3One()
	t.SetDirty()
}

func (t *Transform) Initialize(workGroup *concurrent.WorkGroup) {
	defer tracing.NewRegion("matrix.Initialize").End()
	t.workGroup = workGroup
	t.setup()
}

func (t *Transform) SetupRawTransform() { t.setup() }

func (t *Transform) SetChildrenOrdered()   { t.orderedChildren = true }
func (t *Transform) SetChildrenUnordered() { t.orderedChildren = false }

func (t *Transform) LocalPosition() Vec3 { return t.position }
func (t *Transform) Rotation() Vec3      { return t.rotation }
func (t *Transform) Scale() Vec3         { return t.scale }
func (t *Transform) Parent() *Transform  { return t.parent }
func (t *Transform) Position() Vec3      { return t.relativePosition }

func (t *Transform) IsDirty() bool {
	if !t.frameDirty {
		return false
	}
	return !t.framePosition.Equals(t.position) ||
		!t.frameRotation.Equals(t.rotation) ||
		!t.frameScale.Equals(t.scale)
}

func (t *Transform) Right() Vec3 {
	t.updateMatrices()
	return t.localMatrix.Right().Normal()
}

func (t *Transform) Up() Vec3 {
	t.updateMatrices()
	return t.localMatrix.Up().Normal()
}

func (t *Transform) Forward() Vec3 {
	t.updateMatrices()
	return t.localMatrix.Forward().Normal()
}

func (t *Transform) removeChild(child *Transform) {
	for i, c := range t.children {
		if c == child {
			if t.orderedChildren {
				t.children = slices.Delete(t.children, i, i+1)
			} else {
				t.children = klib.RemoveUnordered(t.children, i)
			}
			break
		}
	}
}

func (t *Transform) SetParent(parent *Transform) {
	if t.parent == parent {
		return
	}
	pos, rot, scale := t.WorldTransform()
	if t.parent != nil {
		t.parent.removeChild(t)
	}
	t.parent = parent
	if t.parent != nil {
		p, r, s := t.parent.WorldTransform()
		pos.SubtractAssign(p)
		rot.SubtractAssign(r)
		if Abs(s.X()) <= FloatSmallestNonzero {
			scale.SetX(0)
		} else {
			scale.SetX(scale.X() / s.X())
		}
		if Abs(s.Y()) <= FloatSmallestNonzero {
			scale.SetY(0)
		} else {
			scale.SetY(scale.Y() / s.Y())
		}
		if Abs(s.Z()) <= FloatSmallestNonzero {
			scale.SetZ(0)
		} else {
			scale.SetZ(scale.Z() / s.Z())
		}
		t.parent.children = append(t.parent.children, t)
	}
	t.SetPosition(pos)
	t.SetRotation(rot)
	t.SetScale(scale)
}

func (t *Transform) SetDirty() {
	if !t.isDirty {
		if !t.frameDirty && t.workGroup != nil {
			t.workGroup.Add(TransformWorkGroup, t.updateMatrices)
			t.workGroup.Add(TransformResetWorkGroup, t.ResetDirty)
		}
		t.isDirty = true
		t.frameDirty = true
	}
	for _, child := range t.children {
		child.SetDirty()
	}
}

func (t *Transform) ResetDirty() {
	if t.isDirty {
		t.updateMatrices()
		t.isDirty = false
	}
	t.frameDirty = false
}

func (t *Transform) SetLocalPosition(position Vec3) {
	if !t.position.Equals(position) {
		if !t.frameDirty {
			t.framePosition = t.position
			t.frameRotation = t.rotation
			t.frameScale = t.scale
		}
		t.position = position
		t.relativePosition = position
		if t.parent != nil {
			wm := t.parent.WorldMatrix()
			t.relativePosition.MultiplyAssign(wm.ExtractScale())
		}
		t.SetDirty()
	}
}

func (t *Transform) SetPosition(position Vec3) {
	if t.parent == nil {
		t.SetLocalPosition(position)
		return
	}
	wm := t.parent.WorldMatrix()
	invParent := wm
	invParent.Inverse()
	localPos := invParent.TransformPoint(wm.ExtractPosition().Add(position))
	if !t.position.Equals(localPos) {
		if !t.frameDirty {
			t.framePosition = t.position
			t.frameRotation = t.rotation
			t.frameScale = t.scale
		}
		t.position = localPos
		t.relativePosition = position
		t.SetDirty()
	}
}

func (t *Transform) SetRotation(rotation Vec3) {
	if !t.rotation.Equals(rotation) {
		if !t.frameDirty {
			t.framePosition = t.position
			t.frameRotation = t.rotation
			t.frameScale = t.scale
		}
		t.rotation = rotation
		t.SetDirty()
	}
}

func (t *Transform) SetScale(scale Vec3) {
	if !t.scale.Equals(scale) {
		if !t.frameDirty {
			t.framePosition = t.position
			t.frameRotation = t.rotation
			t.frameScale = t.scale
		}
		t.scale = scale
		t.SetDirty()
	}
}

func (t *Transform) updateMatrix() {
	if t.isDirty {
		t.localMatrix.Reset()
		t.localMatrix.Scale(t.scale)
		t.localMatrix.Rotate(t.rotation)
		t.localMatrix.Translate(t.position)
	}
}

func (t *Transform) updateWorldMatrix() {
	if t.isDirty {
		if t.parent != nil {
			t.worldMatrix = t.localMatrix
			t.worldMatrix.MultiplyAssign(t.parent.WorldMatrix())
		} else {
			t.worldMatrix = t.localMatrix
		}
	}
}

func (t *Transform) updateMatrices() {
	t.updateMatrix()
	t.updateWorldMatrix()
	t.isDirty = false
}

func (t *Transform) Matrix() Mat4 {
	if t.isDirty {
		t.updateMatrices()
	}
	return t.localMatrix
}

func (t *Transform) WorldMatrix() Mat4 {
	if t.isDirty {
		t.updateMatrices()
	}
	return t.worldMatrix
}

func (t *Transform) Copy(other Transform) {
	var p, r, s Vec3
	if t.parent == nil {
		p, r, s = other.WorldTransform()
	} else {
		p, r, s = other.position, other.rotation, other.scale
	}
	t.SetPosition(p)
	t.SetRotation(r)
	t.SetScale(s)
}

func (t *Transform) WorldTransform() (Vec3, Vec3, Vec3) {
	wm := t.WorldMatrix()
	return wm.ExtractPosition(), wm.ExtractRotation().ToEuler(), wm.ExtractScale()
}

func (t *Transform) WorldPosition() Vec3 {
	if t.parent == nil {
		return t.position
	}
	return t.WorldMatrix().ExtractPosition()
}

func (t *Transform) WorldRotation() Vec3 {
	if t.parent == nil {
		return t.rotation
	}
	return t.WorldMatrix().ExtractRotation().ToEuler()
}

func (t *Transform) WorldScale() Vec3 {
	if t.parent == nil {
		return t.scale
	}
	return t.WorldMatrix().ExtractScale()
}

func (t *Transform) SetWorldPosition(position Vec3) {
	p := t.parent
	for p != nil {
		pp := p.position
		position.SubtractAssign(pp)
		p = p.parent
	}
	t.SetPosition(position)
}

func (t *Transform) SetWorldRotation(rotation Vec3) {
	p := t.parent
	for p != nil {
		r := p.rotation
		rotation.SubtractAssign(r)
		p = p.parent
	}
	t.SetRotation(rotation)
}

func (t *Transform) SetWorldScale(scale Vec3) {
	p := t.parent
	for p != nil {
		s := p.scale
		scale.DivideAssign(s)
		p = p.parent
	}
	t.SetScale(scale)
}

func (t *Transform) ContainsPoint2D(point Vec2) bool {
	p, _, s := t.WorldTransform()
	l := p.X() - (s.X() * 0.5)
	r := p.X() + (s.X() * 0.5)
	u := p.Y() + (s.Y() * 0.5)
	d := p.Y() - (s.Y() * 0.5)
	return point.X() >= l && point.X() <= r && point.Y() >= d && point.Y() <= u
}

func (t *Transform) ContainsPoint(point Vec3) bool {
	p, _, s := t.WorldTransform()
	l := p.X() - (s.X() * 0.5)
	r := p.X() + (s.X() * 0.5)
	u := p.Y() + (s.Y() * 0.5)
	d := p.Y() - (s.Y() * 0.5)
	f := p.Z() - (s.Z() * 0.5)
	b := p.Z() + (s.Z() * 0.5)
	return point.X() >= l && point.X() <= r && point.Y() >= d &&
		point.Y() <= u && point.Z() >= f && point.Z() <= b
}

func (t *Transform) LookAt(point Vec3) {
	eye := t.WorldPosition()
	rot := Mat4LookAt(point, eye, Vec3Up())
	q := QuaternionFromMat4(rot)
	r := q.ToEuler()
	t.SetWorldRotation(r)
}

// ScaleWithoutChildren will scale a transform without changing the world scale
// of the children of this transform. That is to say, it will update all the
// child transform scales to return to their original world scale after scaling
// this transform.
func (t *Transform) ScaleWithoutChildren(scale Vec3) {
	count := len(t.children)
	type tmp struct{ pos, rot, scale Vec3 }
	arr := make([]tmp, count)
	for i := range t.children {
		arr[i].pos, arr[i].rot, arr[i].scale = t.children[i].WorldTransform()
	}
	t.SetScale(scale)
	p, r, s := t.WorldTransform()
	for i := range count {
		arr[i].pos.SubtractAssign(p)
		arr[i].rot.SubtractAssign(r)
		if Abs(s[Vx]) <= FloatSmallestNonzero {
			arr[i].scale[Vx] = 0
		} else {
			arr[i].scale[Vx] = arr[i].scale[Vx] / s[Vx]
		}
		if Abs(s[Vy]) <= FloatSmallestNonzero {
			arr[i].scale[Vy] = 0
		} else {
			arr[i].scale[Vy] = arr[i].scale[Vy] / s[Vy]
		}
		if Abs(s[Vz]) <= FloatSmallestNonzero {
			arr[i].scale[Vz] = 0
		} else {
			arr[i].scale[Vz] = arr[i].scale[Vz] / s[Vz]
		}
		t.children[i].SetPosition(arr[i].pos)
		t.children[i].SetRotation(arr[i].rot)
		t.children[i].SetScale(arr[i].scale)
	}
}

func (t *Transform) AddPosition(add Vec3) { t.SetPosition(t.position.Add(add)) }
func (t *Transform) AddRotation(add Vec3) { t.SetRotation(t.rotation.Add(add)) }
func (t *Transform) AddScale(add Vec3)    { t.SetScale(t.scale.Add(add)) }
