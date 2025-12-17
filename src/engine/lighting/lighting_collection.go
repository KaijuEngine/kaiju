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

package lighting

import (
	"kaiju/debug"
	"kaiju/engine/pooling"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"sort"
)

type LightEntry struct {
	rendering.Light
	Transform *matrix.Transform
	lastDist  float32
	poolId    pooling.PoolGroupId
	id        pooling.PoolIndex
}

type LightCollection struct {
	Cache      []rendering.Light
	pools      pooling.PoolGroup[LightEntry]
	lastPoint  matrix.Vec3
	itrDists   []*LightEntry
	hasChanges bool
}

func (c *LightCollection) Add(transform *matrix.Transform, target rendering.Light) *LightEntry {
	defer tracing.NewRegion("LightCollection.Add").End()
	entry, poolId, elmId := c.pools.Add()
	*entry = LightEntry{
		poolId:    poolId,
		id:        elmId,
		Light:     target,
		Transform: transform,
	}
	// Hack to force the next update cache to happen
	c.lastPoint = matrix.NewVec3(matrix.Inf(1), matrix.Inf(1), matrix.Inf(1))
	c.hasChanges = true
	return entry
}

func (c *LightCollection) Clear() {
	c.pools.Clear()
}

func (c *LightCollection) HasChanges() bool {
	changes := c.hasChanges
	c.hasChanges = false
	return changes
}

func (c *LightCollection) UpdateCache(point matrix.Vec3) []rendering.Light {
	defer tracing.NewRegion("Collection[T].UpdateCache").End()
	if len(c.Cache) > 0 {
		c.findLocalLights(point, c.Cache)
	}
	return c.Cache
}

func (c *LightCollection) findLocalLights(point matrix.Vec3, writeTo []rendering.Light) {
	defer tracing.NewRegion("LightCollection.FindClosest").End()
	const moveDistanceToRecalculate = 1
	debug.Assert(len(writeTo) > 0, "you can not use an empty slice for LightCollection.FindClosest")
	if !matrix.Vec3ApproxTo(c.lastPoint, point, moveDistanceToRecalculate) {
		c.itrDists = klib.WipeSlice(c.itrDists)
		c.pools.Each(func(elm *LightEntry) {
			if elm.Type() != rendering.LightTypeDirectional {
				elm.lastDist = point.Subtract(elm.Transform.Position()).Length()
			}
			c.itrDists = append(c.itrDists, elm)
		})
		sort.Slice(c.itrDists, func(i, j int) bool {
			if c.itrDists[i].Light.Type() == rendering.LightTypeDirectional {
				return true
			} else if c.itrDists[j].Light.Type() == rendering.LightTypeDirectional {
				return false
			} else {
				return c.itrDists[i].lastDist < c.itrDists[j].lastDist
			}
		})
		c.lastPoint = point
	}
	for i := range min(len(writeTo), len(c.itrDists)) {
		writeTo[i] = c.itrDists[i].Light
	}
}
