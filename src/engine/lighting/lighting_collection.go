/******************************************************************************/
/* lighting_collection.go                                                     */
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

package lighting

import (
	"sort"

	"github.com/KaijuEngine/kaiju/debug"
	"github.com/KaijuEngine/kaiju/klib"
	"github.com/KaijuEngine/kaiju/matrix"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

type EntryId = int

type Entry[T any] struct {
	// TODO:  If the target moves, this transform would need to be updated
	id       EntryId
	Position matrix.Vec3
	Target   T
	lastDist float32
}

type Collection[T any] struct {
	Cache     []T
	nextId    EntryId
	lastPoint matrix.Vec3
	entries   []Entry[T]
}

func (c *Collection[T]) Add(position matrix.Vec3, target T) EntryId {
	defer tracing.NewRegion("ShadowCollection.Add").End()
	c.nextId++
	c.entries = append(c.entries, Entry[T]{
		id:       c.nextId,
		Position: position,
		Target:   target,
	})
	return c.nextId
}

func (c *Collection[T]) FindById(id EntryId) *Entry[T] {
	for i := range c.entries {
		if c.entries[i].id == id {
			return &c.entries[i]
		}
	}
	debug.Throw("an invalid id was passed in to search the collection for")
	return nil
}

func (c *Collection[T]) FindClosest(point matrix.Vec3, writeTo []T) {
	const moveDistanceToRecalculate = 1
	defer tracing.NewRegion("Collection[T].FindClosest").End()
	debug.Assert(len(writeTo) > 0, "you can not use an empty slice for Collection[T].FindClosest")
	if matrix.Vec3ApproxTo(c.lastPoint, point, moveDistanceToRecalculate) {
		for i := range c.entries {
			e := &c.entries[i]
			e.lastDist = point.Subtract(e.Position).Length()
		}
		sort.Slice(c.entries, func(i, j int) bool {
			return c.entries[i].lastDist < c.entries[j].lastDist
		})
		c.lastPoint = point
	}
	for i := range min(len(writeTo), len(c.entries)) {
		writeTo[i] = c.entries[i].Target
	}
}

func (c *Collection[T]) Clear() {
	c.entries = klib.WipeSlice(c.entries)
}

func (c *Collection[T]) UpdateCache(point matrix.Vec3) []T {
	if len(c.Cache) > 0 {
		c.FindClosest(point, c.Cache)
	}
	return c.Cache
}
