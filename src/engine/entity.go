/******************************************************************************/
/* entity.go                                                                  */
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

package engine

import (
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/systems/events"
	"log/slog"
	"slices"
)

// EntityId is a string that represents a unique identifier for an entity. The
// identifier is only valid for entities that are not generated through template
// instantiation. The identifier may also be stripped during game runtime if the
// entity is never externally referenced by any other part of the system.
type EntityId string

// Entity is a struct that represents an arbitrary object in the host system.
// It contains a 3D transformation and can be a parent of, or a child to, other
// entities. Entities can also contain arbitrary named data to make it easier
// to access data that is specific to the entity.
//
// Child entities are unordered by default, you'll need to call
// #Entity.SetChildrenOrdered to make them ordered. It is recommended to leave
// children unordered unless you have a specific reason to order them.
type Entity struct {
	id                    EntityId
	Transform             matrix.Transform
	Parent                *Entity
	Children              []*Entity
	matrix                matrix.Mat4
	namedData             map[string][]any
	data                  []EntityData
	OnDestroy             events.Event
	OnActivate            events.Event
	OnDeactivate          events.Event
	name                  string
	EditorBindings        entityEditorBindings
	destroyedFrames       int8
	isDestroyed           bool
	isActive              bool
	deactivatedFromParent bool
	orderedChildren       bool
}

// NewEntity creates a new #Entity struct and returns it
func NewEntity() *Entity {
	e := &Entity{
		isActive:     true,
		Children:     make([]*Entity, 0),
		Transform:    matrix.NewTransform(),
		matrix:       matrix.Mat4Identity(),
		namedData:    make(map[string][]interface{}),
		name:         "Entity",
		OnDestroy:    events.New(),
		OnActivate:   events.New(),
		OnDeactivate: events.New(),
		data:         make([]EntityData, 0),
	}
	e.EditorBindings.init()
	return e
}

// ID returns the unique identifier of the entity. The Id is only valid for
// entities that are not generated through template instantiation. The Id may
// also be stripped during game runtime if the entity is never externally
// referenced by any other part of the system.
func (e *Entity) Id() EntityId { return e.id }

// IsRoot returns true if the entity is the root entity in the hierarchy
func (e *Entity) IsRoot() bool { return e.Parent == nil }

// HasChildren returns true if the entity has any children
func (e *Entity) HasChildren() bool { return len(e.Children) > 0 }

// ChildCount returns the number of children the entity has
func (e *Entity) ChildCount() int { return len(e.Children) }

// ChildAt returns the child entity at the specified index
func (e *Entity) ChildAt(idx int) *Entity { return e.Children[idx] }

// Name returns the name of the entity
func (e *Entity) Name() string { return e.name }

// SetName sets the name of the entity
func (e *Entity) SetName(name string) { e.name = name }

// IsActive will return true if the entity is active, false otherwise
func (e *Entity) IsActive() bool { return e.isActive }

// IsDestroyed will return true if the entity is destroyed, false otherwise
func (e *Entity) IsDestroyed() bool { return e.isDestroyed }

// SetChildrenOrdered sets the children of the entity to be ordered
func (e *Entity) SetChildrenOrdered() {
	e.orderedChildren = true
	e.Transform.SetChildrenOrdered()
}

// SetChildrenUnordered sets the children of the entity to be unordered
func (e *Entity) SetChildrenUnordered() {
	e.orderedChildren = false
	e.Transform.SetChildrenUnordered()
}

// Activate will set an active flag on the entity that can be queried with
// #Entity.IsActive. It will also set the active flag on all children of the
// entity. If the entity is already active, this function will do nothing.
func (e *Entity) Activate() {
	if e.isActive {
		return
	}
	e.Transform.SetDirty()
	e.isActive = true
	for i := range e.Children {
		e.Children[i].activateFromParent()
	}
	e.OnActivate.Execute()
}

// Deactivate will set an active flag on the entity that can be queried with
// #Entity.IsActive. It will also set the active flag on all children of the
// entity. If the entity is already inactive, this function will do nothing.
func (e *Entity) Deactivate() {
	if !e.isActive {
		return
	}
	e.deactivatedFromParent = false
	e.isActive = false
	for i := range e.Children {
		e.Children[i].deactivateFromParent()
	}
	e.OnDeactivate.Execute()
}

// SetActive will set the active flag on the entity that can be queried with
// #Entity.IsActive. It will also set the active flag on all children of the
// entity. If the entity is already active, this function will do nothing.
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

// Destroy will set the destroyed flag on the entity, this can be queried with
// #Entity.IsDestroyed. The entity is not immediately destroyed as it may be
// in use for the current frame. The #Entity.TickCleanup should be called for
// each frame to check if the entity is ready to be completely destroyed.
//
// Destroying a parent will also destroy all children of the entity.
func (e *Entity) Destroy() {
	if !e.isDestroyed {
		e.isDestroyed = true
		e.destroyedFrames = 1
		for i := len(e.Children) - 1; i >= 0; i-- {
			e.Children[i].Destroy()
		}
		e.removeFromParent()
		e.Transform.SetParent(nil)
	}
}

// SetParent will set the parent of the entity. If the entity already has a
// parent, it will be removed from the parent's children list. If the new parent
// is nil, the entity will be removed from the hierarchy and will become the
// root entity. If the new parent is not nil, the entity will be added to the
// new parent's children list. If the new parent is not active, the entity will
// be deactivated as well.
//
// This will also handle the transformation parenting internally
func (e *Entity) SetParent(newParent *Entity) {
	if e == newParent {
		slog.Error("Can't set an entity to parent itself")
		return
	}
	if newParent != nil && newParent.isDestroyed {
		slog.Error("Can't set an entity to a destroyed parent")
		return
	}
	if newParent == e.Parent {
		return
	}
	p := newParent
	for p != nil {
		if p.Parent == e {
			slog.Error("Can't set an entity to a child of itself")
			return
		}
		p = p.Parent
	}
	e.removeFromParent()
	e.Parent = newParent
	if newParent != nil {
		e.Transform.SetParent(&newParent.Transform)
	} else {
		e.Transform.SetParent(nil)
	}
	if e.Parent != nil {
		e.Parent.Children = append(e.Parent.Children, e)
	}
	if e.Parent != nil && !e.Parent.isActive {
		e.deactivateFromParent()
	}
}

// FindByName will search the entity and the tree of children for the first
// entity with the specified name. If no entity is found, nil will be returned.
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

// ScaleWithoutChildren will temporarily remove all children from the entity,
// scale the entity, and then re-add the children. This is useful when you want
// to scale an entity without scaling its children. When the children are
// re-added, they keep the world transformations they had before being removed.
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

// TickCleanup will check if the entity is ready to be completely destroyed. If
// the entity is ready to be destroyed, it will execute the #Entity.OnDestroy
// event and return true. If the entity is not ready to be destroyed, it will
// return false.
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

// Root will return the root entity of the entity's hierarchy
func (e *Entity) Root() *Entity {
	root := e
	for root.Parent != nil {
		root = root.Parent
	}
	return root
}

// AddNamedData allows you to add arbitrary data to the entity that can be
// accessed by a string key. This is useful for storing data that is specific
// to the entity.
//
// Named data is stored in a map of slices, so you can add multiple pieces of
// data to the same key. It is recommended to compile the data into a single
// structure so the slice length is 1, but sometimes that's not reasonable.
func (e *Entity) AddNamedData(key string, data interface{}) {
	if _, ok := e.namedData[key]; !ok {
		e.namedData[key] = make([]interface{}, 0)
	}
	e.namedData[key] = append(e.namedData[key], data)
}

// RemoveNamedData will remove the specified data from the entity's named data
// map. If the key does not exist, this function will do nothing.
//
// *This will remove the entire slice and all of it's data*
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

// NamedData will return the data associated with the specified key. If the key
// does not exist, nil will be returned.
func (e *Entity) NamedData(key string) []interface{} {
	if _, ok := e.namedData[key]; ok {
		return e.namedData[key]
	}
	return nil
}

func (e *Entity) removeFromParent() {
	if e.Parent == nil {
		return
	}
	for i := range e.Parent.Children {
		me := e.Parent.Children[i]
		if me == e {
			if e.orderedChildren {
				e.Parent.Children = slices.Delete(e.Parent.Children, i, i+1)
			} else {
				e.Parent.Children = klib.RemoveUnordered(e.Parent.Children, i)
			}
			break
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
