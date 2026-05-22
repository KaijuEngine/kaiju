/******************************************************************************/
/* lighting_collection.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package lighting

import (
	"sort"

	"kaijuengine.com/debug"
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
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
	c.setHasChanges()
	return entry
}

func (c *LightCollection) Remove(light *LightEntry) {
	defer tracing.NewRegion("LightCollection.Remove").End()
	c.pools.Remove(light.poolId, light.id)
	c.setHasChanges()
}

func (c *LightCollection) setHasChanges() {
	// Hack to force the next update cache to happen
	c.lastPoint = matrix.NewVec3(matrix.Inf(1), matrix.Inf(1), matrix.Inf(1))
	c.hasChanges = true
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
		clear(writeTo)
	}
	for i := range min(len(writeTo), len(c.itrDists)) {
		writeTo[i] = c.itrDists[i].Light
	}
}
