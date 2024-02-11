package matrix

import "math"

type Transform struct {
	localMatrix               Mat4
	worldMatrix               Mat4
	parent                    *Transform
	children                  []*Transform
	position, rotation, scale Vec3
	isDirty, isLive           bool
}

func NewTransform() Transform {
	return Transform{
		localMatrix: Mat4Identity(),
		worldMatrix: Mat4Identity(),
		position:    Vec3Zero(),
		rotation:    Vec3Zero(),
		scale:       Vec3One(),
		isDirty:     true,
	}
}

func (t *Transform) Position() Vec3 {
	return t.position
}

func (t *Transform) Rotation() Vec3 {
	return t.rotation
}

func (t *Transform) Scale() Vec3 {
	return t.scale
}

func (t *Transform) removeChild(child *Transform) {
	for i, c := range t.children {
		if c == child {
			last := len(t.children) - 1
			t.children[i] = t.children[last]
			t.children = t.children[:last]
			break
		}
	}
}

func (t *Transform) SetParent(parent *Transform) {
	if t.parent == parent {
		return
	} else if t.parent != nil {
		t.parent.removeChild(t)
	}
	pos, rot, scale := t.WorldTransform()
	t.parent = parent
	if t.parent != nil {
		p, r, s := t.parent.WorldTransform()
		pos.SubtractAssign(p)
		rot.SubtractAssign(r)
		if Abs(s.X()) <= math.SmallestNonzeroFloat64 {
			scale.SetX(0)
		} else {
			scale.SetX(scale.X() / s.X())
		}
		if Abs(s.Y()) <= math.SmallestNonzeroFloat64 {
			scale.SetY(0)
		} else {
			scale.SetY(scale.Y() / s.Y())
		}
		if Abs(s.Z()) <= math.SmallestNonzeroFloat64 {
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

func (t *Transform) Right() Vec3 {
	return t.localMatrix.Right().Normal()
}

func (t *Transform) Up() Vec3 {
	return t.localMatrix.Up().Normal()
}

func (t *Transform) Forward() Vec3 {
	return t.localMatrix.Forward().Normal()
}

func (t *Transform) SetDirty() {
	t.isDirty = true
	for _, child := range t.children {
		child.SetDirty()
	}
}

func (t *Transform) ResetDirty() {
	if t.isDirty {
		t.UpdateMatrix()
		t.isDirty = false
	}
}

func (t *Transform) IsDirty() bool {
	return t.isDirty
}

func (t *Transform) SetPosition(position Vec3) {
	if !t.position.Equals(position) {
		t.position = position
		t.SetDirty()
	}
}

func (t *Transform) SetRotation(rotation Vec3) {
	if !t.rotation.Equals(rotation) {
		t.rotation = rotation
		t.SetDirty()
	}
}

func (t *Transform) SetScale(scale Vec3) {
	if !t.scale.Equals(scale) {
		t.scale = scale
		t.SetDirty()
	}
}

func (t *Transform) StartLive() {
	t.isLive = true
}

func (t *Transform) StopLive() {
	t.isLive = false
}

func (t *Transform) UpdateMatrix() {
	if t.isDirty || t.isLive {
		t.localMatrix.Reset()
		t.localMatrix.Scale(t.scale)
		t.localMatrix.Rotate(t.rotation)
		t.localMatrix.Translate(t.position)
	}
}

func (t *Transform) UpdateWorldMatrix() {
	if t.isDirty || t.isLive {
		t.worldMatrix.Reset()
		t.CalcWorldMatrix(&t.worldMatrix)
	}
}

func (t *Transform) updateMatrices() {
	t.UpdateMatrix()
	t.UpdateWorldMatrix()
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

func (t *Transform) CalcWorldMatrix(base *Mat4) {
	m := Mat4Identity()
	m.Scale(t.scale)
	m.Rotate(t.rotation)
	m.Translate(t.position)
	dPos := t.position.Add(base.Position())
	base.MultiplyAssign(m)
	base.SetTranslation(dPos)
	if t.parent != nil {
		t.parent.CalcWorldMatrix(base)
	}
}

func (t *Transform) Copy(other Transform) {
	t.position = other.position
	t.rotation = other.rotation
	t.scale = other.scale
	t.localMatrix = other.localMatrix
	t.worldMatrix = other.worldMatrix
	t.isDirty = other.isDirty
}

func (t *Transform) WorldTransform() (Vec3, Vec3, Vec3) {
	pos, rot, scale := Vec3{}, Vec3{}, Vec3One()
	p := t
	for p != nil {
		pp, rr, ss := p.position, p.rotation, p.scale
		pos.AddAssign(pp)
		rot.AddAssign(rr)
		scale.MultiplyAssign(ss)
		p = p.parent
	}
	return pos, rot, scale
}

func (t *Transform) WorldPosition() Vec3 {
	pos := Vec3{}
	p := t
	for p != nil {
		pp := p.position
		pos.AddAssign(pp)
		p = p.parent
	}
	return pos
}

func (t *Transform) WorldRotation() Vec3 {
	rot := Vec3{}
	p := t
	for p != nil {
		r := p.rotation
		rot.AddAssign(r)
		p = p.parent
	}
	return rot
}

func (t *Transform) WorldScale() Vec3 {
	scale := Vec3One()
	p := t
	for p != nil {
		s := p.scale
		scale.MultiplyAssign(s)
		p = p.parent
	}
	return scale
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
