/******************************************************************************/
/* selection.go                                                               */
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

package selection

import (
	"kaiju/editor/memento"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/engine/systems/events"
	"kaiju/engine/systems/visual2d/sprite"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/rendering"
	"log/slog"
	"slices"
)

const (
	minDragDistance = 5
	rayCastLength   = 10000
)

type Selection struct {
	host        *engine.Host
	box         *sprite.Sprite
	entities    []*engine.Entity
	downPos     matrix.Vec2
	Changed     events.Event
	shaderDatas map[*engine.Entity][]*rendering.ShaderDataBasic
	history     *memento.History
}

func (s *Selection) isBoxDrag() bool { return s.box.Entity.IsActive() }

func New(host *engine.Host, history *memento.History) Selection {
	host.CreatingEditorEntities()
	b := &sprite.Sprite{}
	b.Init(0, 0, 0, 0, host, assets.TextureSquare, matrix.Color{0.7, 0.7, 0.7, 0.5})
	host.DoneCreatingEditorEntities()
	b.Deactivate()
	return Selection{
		host:        host,
		box:         b,
		entities:    make([]*engine.Entity, 0),
		shaderDatas: make(map[*engine.Entity][]*rendering.ShaderDataBasic),
		history:     history,
	}
}

func (s *Selection) Entities() []*engine.Entity { return s.entities }
func (s *Selection) HasSelection() bool         { return len(s.entities) > 0 }

func (s *Selection) Contains(e *engine.Entity) bool {
	for i := range s.entities {
		if s.entities[i] == e {
			return true
		}
	}
	return false
}

func (s *Selection) IsEmpty() bool { return len(s.entities) == 0 }

func (s *Selection) deactivateBox() {
	s.box.SetSize(0, 0)
	s.box.Deactivate()
}

func (s *Selection) clearInternal() {
	if len(s.entities) == 0 {
		return
	}
	for k, v := range s.shaderDatas {
		for i := range v {
			s.shaderDatas[k][i].Destroy()
		}
	}
	clear(s.shaderDatas)
	s.entities = klib.WipeSlice(s.entities)
	s.Changed.Execute()
}

func (s *Selection) Clear() {
	if len(s.entities) == 0 {
		return
	}
	s.history.Add(&selectHistory{
		selection: s,
		from:      slices.Clone(s.entities),
		to:        make([]*engine.Entity, 0),
	})
	s.clearInternal()
}

func (s *Selection) setInternal(entities []*engine.Entity) {
	for _, e := range s.entities {
		if !slices.Contains(entities, e) {
			s.removeInternal(e)
		}
	}
	for _, e := range entities {
		s.addInternal(e)
	}
}

func (s *Selection) Set(e ...*engine.Entity) {
	from := slices.Clone(s.entities)
	s.setInternal(e)
	s.history.Add(&selectHistory{
		selection: s,
		from:      from,
		to:        slices.Clone(s.entities),
	})
	s.Changed.Execute()
}

func (s *Selection) addInternal(e *engine.Entity) {
	if slices.Contains(s.entities, e) {
		return
	}
	s.entities = append(s.entities, e)
	outline, err := s.host.MaterialCache().Material(assets.MaterialDefinitionOutline)
	if err != nil {
		slog.Error("failed to load the outline material", "error", err)
		return
	}
	draws := e.EditorBindings.Drawings()
	s.shaderDatas[e] = []*rendering.ShaderDataBasic{}
	for _, d := range draws {
		ds := &rendering.ShaderDataBasic{
			ShaderDataBase: rendering.NewShaderDataBase(),
			Color:          matrix.ColorCrimson(),
		}
		ds.Color.SetA(0.01) // Line width
		d.Transform.SetDirty()
		s.shaderDatas[e] = append(s.shaderDatas[e], ds)
		d.Material = outline
		d.ShaderData = ds
		s.host.Drawings.AddDrawing(d)
	}
	s.host.RunAfterFrames(1, func() {
		// Make drawings snap to transform
		for _, d := range draws {
			d.Transform.SetDirty()
		}
	})
}

func (s *Selection) removeInternal(e *engine.Entity) {
	for i := range s.entities {
		if s.entities[i] == e {
			s.entities = slices.Delete(s.entities, i, i+1)
			for j := range s.shaderDatas[e] {
				s.shaderDatas[e][j].Destroy()
			}
			delete(s.shaderDatas, e)
			break
		}
	}
}

func (s *Selection) Remove(e ...*engine.Entity) {
	if len(e) == 0 {
		return
	}
	from := slices.Clone(s.entities)
	for i := range e {
		s.removeInternal(e[i])
	}
	s.history.Add(&selectHistory{
		selection: s,
		from:      from,
		to:        slices.Clone(s.entities),
	})
	s.Changed.Execute()
}

func (s *Selection) Toggle(e ...*engine.Entity) {
	sub := make([]*engine.Entity, 0, len(e))
	for i := range e {
		if !slices.Contains(s.entities, e[i]) {
			sub = append(sub, e[i])
		}
	}
	if len(sub) > 0 {
		s.Add(sub...)
	} else {
		s.Remove(e...)
	}
}

func (s *Selection) UntrackedClear() {
	s.clearInternal()
}

func (s *Selection) UntrackedAdd(e ...*engine.Entity) {
	for i := range e {
		s.addInternal(e[i])
	}
}

func (s *Selection) Add(e ...*engine.Entity) {
	if len(e) == 0 {
		return
	}
	from := slices.Clone(s.entities)
	for i := range e {
		s.addInternal(e[i])
	}
	s.history.Add(&selectHistory{
		selection: s,
		from:      from,
		to:        slices.Clone(s.entities),
	})
	s.Changed.Execute()
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
	s.box.SetSize(w, h)
}

func (s *Selection) clickSelect(host *engine.Host) {
	ray := host.Camera.RayCast(s.downPos)
	all := host.Entities()
	found := false
	for i := 0; i < len(all) && !found; i++ {
		pos := all[i].Transform.WorldPosition()
		volume := all[i].EditorBindings.Data("bvh")
		hit := false
		if volume != nil {
			_, hit = volume.(*collision.BVH).RayHit(ray, rayCastLength)
		} else {
			hit = ray.SphereHit(pos, 0.5, rayCastLength)
		}
		if hit {
			if host.Window.Keyboard.HasCtrl() {
				s.Toggle(all[i])
			} else if host.Window.Keyboard.HasShift() {
				s.Add(all[i])
			} else {
				s.Set(all[i])
			}
			found = true
		}
	}
	if !found {
		s.Clear()
	}
}

func (s *Selection) unProjectSelect(host *engine.Host, endPos matrix.Vec2) {
	all := host.Entities()
	pts := make([]matrix.Vec3, len(all))
	vp := host.Window.Viewport()
	view := host.Camera.View()
	proj := host.Camera.Projection()
	// TODO:  Parallel
	for i := range all {
		point := all[i].Transform.WorldPosition()
		pts[i] = matrix.Mat4ToScreenSpace(point, view, proj, vp)
	}
	box := matrix.Vec4Area(s.downPos.X(), s.downPos.Y(), endPos.X(), endPos.Y())
	adding := make([]*engine.Entity, 0, len(all))
	for i := range pts {
		if box.AreaContains(pts[i].X(), pts[i].Y()) {
			adding = append(adding, all[i])
		}
	}
	if len(adding) > 0 {
		if host.Window.Keyboard.HasCtrl() {
			s.Toggle(adding...)
		} else if host.Window.Keyboard.HasShift() {
			s.Add(adding...)
		} else {
			s.Set(adding...)
		}
	} else {
		s.Clear()
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
		centroid.AddAssign(e.Transform.WorldPosition())
	}
	centroid.ScaleAssign(1 / matrix.Float(len(s.entities)))
	return centroid
}

func (s *Selection) Bounds() collision.AABB {
	min := matrix.Vec3Inf(1)
	max := matrix.Vec3Inf(-1)
	for _, e := range s.entities {
		p := e.Transform.Position()
		ex := matrix.Vec3Zero()
		draws := e.EditorBindings.Drawings()
		for _, d := range draws {
			ex = matrix.Vec3Max(ex, d.Mesh.Details.Extents)
		}
		min = matrix.Vec3Min(min, p.Subtract(ex))
		max = matrix.Vec3Max(max, p.Add(ex))
	}
	return collision.AABBFromMinMax(min, max)
}

func (s *Selection) Parent(history *memento.History) {
	if s.IsEmpty() {
		return
	}
	var h selectParentingHistory
	if len(s.entities) == 1 {
		h = selectParentingHistory{
			targets:     []*engine.Entity{s.entities[0]},
			lastParents: []*engine.Entity{s.entities[0].Parent},
			newParent:   nil,
		}
	} else {
		children := s.entities[:len(s.entities)-1]
		lastParents := make([]*engine.Entity, 0, len(children))
		for _, e := range children {
			lastParents = append(lastParents, e.Parent)
		}
		newParent := s.entities[len(s.entities)-1]
		h = selectParentingHistory{
			targets:     slices.Clone(children),
			lastParents: lastParents,
			newParent:   newParent,
		}
	}
	history.Add(&h)
	h.Redo()
}

func (s *Selection) Focus(camera cameras.Camera) {
	b := s.Bounds()
	z := b.Extent.Length()
	if z <= 0.01 {
		z = 5
	} else {
		z *= 2
	}
	if camera.IsOrthographic() {
		c := camera.(*cameras.StandardCamera)
		p := c.Position()
		p.SetX(b.Center.X())
		p.SetY(b.Center.Y())
		c.SetPositionAndLookAt(p, b.Center.Negative())
		r := c.Width() / c.Height()
		if c.Width() > c.Height() {
			c.Resize(z*r, z)
		} else {
			c.Resize(z, z*r)
		}
	} else {
		c := camera.(*cameras.TurntableCamera)
		c.SetLookAt(b.Center.Negative())
		c.SetZoom(z)
	}
}
