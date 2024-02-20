/******************************************************************************/
/* selection.go                                                               */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package selection

import (
	"kaiju/assets"
	"kaiju/collision"
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/systems/events"
	"kaiju/systems/visual2d/sprite"
)

const (
	minDragDistance = 5
	rayCastLength   = 10000
)

type Selection struct {
	box      *sprite.Sprite
	entities []*engine.Entity
	downPos  matrix.Vec2
	Changed  events.Event
}

func (s *Selection) isBoxDrag() bool { return s.box.Entity.IsActive() }

func New(host *engine.Host) Selection {
	tex := klib.MustReturn(host.TextureCache().Texture(
		assets.TextureSquare, rendering.TextureFilterLinear))
	host.CreatingEditorEntities()
	b := sprite.NewSprite(0, 0, 0, 0, host, tex, matrix.Color{0.7, 0.7, 0.7, 0.5})
	host.DoneCreatingEditorEntities()
	b.Deactivate()
	return Selection{
		box:      b,
		entities: make([]*engine.Entity, 0),
		Changed:  events.New(),
	}
}

func (s *Selection) Entities() []*engine.Entity {
	return s.entities
}

func (s *Selection) IsEmpty() bool { return len(s.entities) == 0 }

func (s *Selection) deactivateBox() {
	s.box.Resize(0, 0)
	s.box.Deactivate()
}

func (s *Selection) Update(host *engine.Host) {
	if s.isBoxDrag() {
		s.boxDrag(host)
	} else {
		s.checkForBoxDrag(&host.Window.Mouse)
	}
}

func (s *Selection) boxDrag(host *engine.Host) {
	mouse := &host.Window.Mouse
	pos := mouse.Position()
	if mouse.Released(hid.MouseButtonLeft) {
		s.deactivateBox()
		if pos.Distance(s.downPos) < minDragDistance {
			s.clickSelect(host)
		} else {
			s.unProjectSelect(host, pos)
		}
		return
	}
	box := matrix.Vec4{s.downPos.X(), s.downPos.Y(), pos.X(), pos.Y()}
	w := box.Right() - box.Left()
	h := box.Top() - box.Bottom()
	s.box.SetPosition(box.Left()+w*0.5, box.Bottom()+h*0.5)
	s.box.Resize(w, h)
}

func (s *Selection) clickSelect(host *engine.Host) {
	changed := false
	if !host.Window.Keyboard.HasCtrl() {
		changed = len(s.entities) > 0
		s.entities = s.entities[:0]
	}
	ray := host.Camera.RayCast(s.downPos)
	all := host.Entities()
	for i := range all {
		pos := all[i].Transform.Position()
		// TODO:  Use BVH or other acceleration structure. The sphere check
		// here is just to get testing quickly
		if ray.SphereHit(pos, 0.5, rayCastLength) {
			s.entities = append(s.entities, all[i])
			changed = true
			break
		}
	}
	if changed {
		s.Changed.Execute()
	}
}

func (s *Selection) unProjectSelect(host *engine.Host, endPos matrix.Vec2) {
	changed := false
	if !host.Window.Keyboard.HasCtrl() {
		changed = len(s.entities) > 0
		s.entities = s.entities[:0]
	}
	all := host.Entities()
	pts := make([]matrix.Vec3, len(all))
	vp := host.Window.Viewport()
	view := host.Camera.View()
	proj := host.Camera.Projection()
	// TODO:  Parallel
	for i := range all {
		point := all[i].Transform.Position()
		pts[i] = matrix.Mat4ToScreenSpace(point, view, proj, vp)
	}
	box := matrix.Vec4Box(s.downPos.X(), s.downPos.Y(), endPos.X(), endPos.Y())
	for i := range pts {
		if box.BoxContains(pts[i].X(), pts[i].Y()) {
			s.entities = append(s.entities, all[i])
			changed = true
		}
	}
	if changed {
		s.Changed.Execute()
	}
}

func (s *Selection) checkForBoxDrag(mouse *hid.Mouse) {
	if mouse.Pressed(hid.MouseButtonLeft) {
		// TODO:  Don't click through top menu
		s.downPos = mouse.Position()
		s.box.Activate()
	}
}

func (s *Selection) Center() matrix.Vec3 {
	centroid := matrix.Vec3Zero()
	for _, e := range s.entities {
		centroid.AddAssign(e.Transform.Position())
	}
	centroid.ScaleAssign(1 / matrix.Float(len(s.entities)))
	return centroid
}

func (s *Selection) Bounds() collision.AABB {
	min := matrix.Vec3Inf(1)
	max := matrix.Vec3Inf(-1)
	for _, e := range s.entities {
		p := e.Transform.Position()
		min = matrix.Vec3Min(min, p)
		max = matrix.Vec3Max(max, p)
	}
	return collision.AABBFromMinMax(min, max)
}
