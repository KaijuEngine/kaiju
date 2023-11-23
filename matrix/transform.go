package matrix

import "math"

type Transform struct {
	matrix                    Mat4
	parent                    *Transform
	position, rotation, scale Vec3
	isDirty, isLive           bool
}

func NewTransform() Transform {
	return Transform{
		matrix:   Mat4Identity(),
		position: Vec3Zero(),
		rotation: Vec3Zero(),
		scale:    Vec3One(),
		isDirty:  true,
	}
}

func (t *Transform) SetParent(parent *Transform) {
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
	}
	t.SetPosition(pos)
	t.SetRotation(rot)
	t.SetScale(scale)
}

func (t Transform) Right() Vec3 {
	return t.matrix.Right().Normal()
}

func (t Transform) Up() Vec3 {
	return t.matrix.Up().Normal()
}

func (t Transform) Forward() Vec3 {
	return t.matrix.Forward().Normal()
}

func (t *Transform) SetDirty() {
	t.isDirty = true
}

func (t *Transform) ResetDirty() {
	if t.isDirty {
		t.UpdateMatrix()
		t.isDirty = false
	}
}

func (t Transform) IsDirty() bool {
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
		t.matrix.Reset()
		t.matrix.Scale(t.scale)
		t.matrix.Rotate(t.rotation)
		t.matrix.Translate(t.position)
	}
}

func (t Transform) Matrix() Mat4 {
	return t.matrix
}

func (t *Transform) WorldMatrix(base *Mat4) {
	m := t.matrix
	m.Reset()
	m.Scale(t.scale)
	m.Rotate(t.rotation)
	m.Translate(t.position)
	bPos := base.Position()
	dPos := t.position.Add(bPos)
	base.MultiplyAssign(m)
	base.SetTranslation(dPos)
	if t.parent != nil {
		t.parent.WorldMatrix(base)
	}
}

func (t *Transform) Copy(other Transform) {
	t.position = other.position
	t.rotation = other.rotation
	t.scale = other.scale
	t.matrix = other.matrix
	t.isDirty = other.isDirty
}

func (t Transform) WorldTransform() (Vec3, Vec3, Vec3) {
	pos, rot, scale := Vec3Zero(), Vec3Zero(), Vec3One()
	p := &t
	for p != nil {
		pp, rr, ss := p.position, p.rotation, p.scale
		pos.AddAssign(pp)
		rot.AddAssign(rr)
		scale.MultiplyAssign(ss)
		p = p.parent
	}
	return pos, rot, scale
}

func (t Transform) WorldPosition() Vec3 {
	pos := Vec3Zero()
	p := &t
	for p != nil {
		pp := p.position
		pos.AddAssign(pp)
		p = p.parent
	}
	return pos
}

func (t Transform) WorldRotation() Vec3 {
	rot := Vec3Zero()
	p := &t
	for p != nil {
		rr := p.rotation
		rot.AddAssign(rr)
		p = p.parent
	}
	return rot
}

func (t Transform) WorldScale() Vec3 {
	scale := Vec3One()
	p := &t
	for p != nil {
		ss := p.scale
		scale.MultiplyAssign(ss)
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
	p.SetPosition(position)
}

func (t *Transform) SetWorldRotation(rotation Vec3) {
	p := t.parent
	for p != nil {
		rr := p.rotation
		rotation.SubtractAssign(rr)
		p = p.parent
	}
	p.SetRotation(rotation)
}

func (t *Transform) SetWorldScale(scale Vec3) {
	p := t.parent
	for p != nil {
		ss := p.scale
		scale.DivideAssign(ss)
		p = p.parent
	}
	p.SetScale(scale)
}
