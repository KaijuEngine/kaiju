package engine

import (
	"kaiju/matrix"
	"kaiju/systems/events"
	"slices"
)

type Entity struct {
	Transform                       matrix.Transform
	Parent                          *Entity
	Children                        []*Entity
	matrix                          matrix.Mat4
	namedData                       map[string][]interface{}
	OnDestroy                       events.Event
	OnActivate                      events.Event
	OnDeactivate                    events.Event
	name                            string
	destroyedFrames                 int8
	isDestroyed                     bool
	isActive, deactivatedFromParent bool
	relativeTransformations         bool
}

func NewEntity() *Entity {
	return &Entity{
		isActive:     true,
		Children:     make([]*Entity, 0),
		Transform:    matrix.NewTransform(),
		matrix:       matrix.Mat4Identity(),
		namedData:    make(map[string][]interface{}),
		name:         "Entity",
		OnDestroy:    events.New(),
		OnActivate:   events.New(),
		OnDeactivate: events.New(),
	}
}

func (e *Entity) IsRoot() bool {
	return e.Parent == nil
}

func (e *Entity) ChildCount() int {
	return len(e.Children)
}

func (e *Entity) ChildAt(idx int) *Entity {
	return e.Children[idx]
}

func (e *Entity) Activate() {
	e.Transform.SetDirty()
	e.isActive = true
	for i := range e.Children {
		e.Children[i].activateFromParent()
	}
	e.OnActivate.Execute()
}

func (e *Entity) Deactivate() {
	e.deactivatedFromParent = false
	e.isActive = false
	for i := range e.Children {
		e.Children[i].deactivateFromParent()
	}
	e.OnDeactivate.Execute()
}

func (e *Entity) RemoveFromParent() {
	if e.Parent != nil {
		for i := range e.Parent.Children {
			me := e.Parent.Children[i]
			if me == e {
				last := len(e.Parent.Children) - 1
				e.Parent.Children[i] = e.Parent.Children[last]
				e.Parent.Children = e.Parent.Children[:last]
				break
			}
		}
	}
}

func (e *Entity) activateFromParent() {
	if e.deactivatedFromParent {
		e.Activate()
		e.deactivatedFromParent = false
	}
}

func (e *Entity) deactivateFromParent() {
	fromParent := e.deactivatedFromParent || e.isActive
	e.Deactivate()
	e.deactivatedFromParent = fromParent
}

func (e *Entity) SetActive(isActive bool) {
	if e.isActive != isActive {
		if isActive {
			e.Activate()
		} else {
			e.Deactivate()
		}
	} else if e.deactivatedFromParent {
		e.deactivatedFromParent = false
	}
}

func (e *Entity) Destroy() {
	if !e.isDestroyed {
		e.isDestroyed = true
		e.destroyedFrames = 1
		for i := len(e.Children) - 1; i >= 0; i-- {
			e.Children[i].Destroy()
		}
		e.RemoveFromParent()
	}
}

func (e *Entity) SetParent(newParent *Entity) {
	if e == newParent {
		panic("Can't set an entity to parent itself")
	}
	if newParent != nil && newParent.isDestroyed {
		panic("Can't set an entity to a destroyed parent")
	}
	if newParent == e.Parent {
		return
	}
	e.RemoveFromParent()
	e.Parent = newParent
	if newParent != nil {
		e.Transform.SetParent(&newParent.Transform)
	} else {
		e.Transform.SetParent(nil)
	}
	if e.Parent != nil {
		e.Parent.Children = append(e.Parent.Children, e)
		e.SetRelativeTransformations(e.Parent.relativeTransformations)
	}
	if e.Parent != nil && !e.Parent.isActive {
		e.deactivateFromParent()
	}
}

func (e *Entity) MatrixRelative(base *matrix.Mat4) {
	e.Transform.UpdateMatrix()
	m := e.Transform.Matrix()
	base.MultiplyAssign(m)
	if !e.IsRoot() {
		e.Parent.MatrixRelative(base)
	}
}

func (e *Entity) Matrix(base *matrix.Mat4) {
	if e.relativeTransformations {
		e.MatrixRelative(base)
	}
	e.Transform.CalcWorldMatrix(base)
}

func (e *Entity) Clone(parentOverride *Entity) *Entity {
	clone := NewEntity()
	if parentOverride == nil {
		clone.SetParent(parentOverride)
	} else {
		clone.SetParent(e.Parent)
	}
	for _, c := range e.Children {
		c.Clone(clone)
	}
	// TODO: Clone named data
	clone.Transform.Copy(e.Transform)
	clone.isDestroyed = e.isDestroyed
	clone.name = e.name
	clone.Transform.SetDirty()
	clone.isActive = e.isActive
	return clone
}

func (e *Entity) Name() string {
	return e.name
}

func (e *Entity) SetName(name string) {
	e.name = name
}

func (e *Entity) FindByName(name string) *Entity {
	if e.name == name {
		return e
	}
	for _, c := range e.Children {
		if found := c.FindByName(name); found != nil {
			return found
		}
	}
	return nil
}

func (e *Entity) SetRelativeTransformations(transformRelative bool) {
	e.relativeTransformations = transformRelative
	for _, c := range e.Children {
		c.SetRelativeTransformations(transformRelative)
	}
}

func (e *Entity) ScaleWithoutChildren(scale matrix.Vec3) {
	count := len(e.Children)
	arr := make([]*Entity, count)
	for i := count - 1; i >= 0; i-- {
		c := e.Children[i]
		c.SetParent(nil)
		arr[i] = c
	}
	e.Children = e.Children[:0]
	e.Transform.SetScale(scale)
	for i := 0; i < count; i++ {
		c := arr[i]
		c.SetParent(e)
	}
}

func (e *Entity) WorldForward() matrix.Vec3 {
	m := matrix.Mat4Identity()
	e.Matrix(&m)
	f := m.Forward()
	f.Normalize()
	return f
}

func (e *Entity) WorldRight() matrix.Vec3 {
	m := matrix.Mat4Identity()
	e.Matrix(&m)
	r := m.Right()
	r.Normalize()
	return r
}

func (e *Entity) WorldUp() matrix.Vec3 {
	m := matrix.Mat4Identity()
	e.Matrix(&m)
	u := m.Up()
	u.Normalize()
	return u
}

func (e *Entity) LookAt(point matrix.Vec3) {
	eye := e.Transform.WorldPosition()
	var rot matrix.Mat4
	rot.LookAt(eye, point, matrix.Vec3Up())
	q := matrix.QuaternionFromMat4(rot)
	r := q.ToEuler()
	e.Transform.SetWorldRotation(r)
}

func (e *Entity) TickCleanup() bool {
	if e.isDestroyed {
		if e.destroyedFrames <= 0 {
			e.OnDestroy.Execute()
			return true
		}
		e.destroyedFrames--
	}
	return false
}

func (e *Entity) CanUpdate() bool {
	return e.isActive && !e.isDestroyed
}

func (e *Entity) LocalPosition() matrix.Vec3 {
	return e.Transform.Position()
}

func (e *Entity) LocalRotation() matrix.Vec3 {
	return e.Transform.Rotation()
}

func (e *Entity) LocalScale() matrix.Vec3 {
	return e.Transform.Scale()
}

func (e *Entity) LocalForward() matrix.Vec3 {
	return e.Transform.Forward()
}

func (e *Entity) LocalRight() matrix.Vec3 {
	return e.Transform.Right()
}

func (e *Entity) LocalUp() matrix.Vec3 {
	return e.Transform.Up()
}

func (e *Entity) Root() *Entity {
	root := e
	for root.Parent != nil {
		root = root.Parent
	}
	return root
}

func (e *Entity) IsActive() bool {
	return e.isActive
}

func (e *Entity) IsDestroyed() bool {
	return e.isDestroyed
}

func (e *Entity) AddNamedData(key string, data interface{}) {
	if _, ok := e.namedData[key]; !ok {
		e.namedData[key] = make([]interface{}, 0)
	}
	e.namedData[key] = append(e.namedData[key], data)
}

func (e *Entity) RemoveNamedData(key string, data interface{}) {
	if _, ok := e.namedData[key]; ok {
		for i := range e.namedData[key] {
			if e.namedData[key][i] == data {
				e.namedData[key] = slices.Delete(e.namedData[key], i, i+1)
				break
			}
		}
	}
}

func (e *Entity) NamedData(key string) []interface{} {
	if _, ok := e.namedData[key]; ok {
		return e.namedData[key]
	}
	return nil
}
