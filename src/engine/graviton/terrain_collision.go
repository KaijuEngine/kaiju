/******************************************************************************/
/* terrain_collision.go                                                       */
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

package graviton

import (
	"errors"
	"fmt"

	"kaijuengine.com/matrix"
)

type TerrainCollision struct {
	Resolution int
	WorldSize  matrix.Vec2
	Heights    []matrix.Float
	MinHeight  matrix.Float
	MaxHeight  matrix.Float
	Bounds     AABB
}

func NewTerrainCollision(resolution int, worldSize matrix.Vec2, heights []matrix.Float, minHeight, maxHeight matrix.Float) (*TerrainCollision, error) {
	return newTerrainCollision(resolution, worldSize, heights, minHeight, maxHeight, false)
}

func NewTerrainCollisionCopy(resolution int, worldSize matrix.Vec2, heights []matrix.Float, minHeight, maxHeight matrix.Float) (*TerrainCollision, error) {
	return newTerrainCollision(resolution, worldSize, heights, minHeight, maxHeight, true)
}

func newTerrainCollision(resolution int, worldSize matrix.Vec2, heights []matrix.Float, minHeight, maxHeight matrix.Float, copyHeights bool) (*TerrainCollision, error) {
	if resolution < 2 {
		return nil, errors.New("terrain collision resolution must be at least 2")
	}
	if worldSize.X() <= 0 || worldSize.Y() <= 0 {
		return nil, errors.New("terrain collision world size must be greater than zero")
	}
	expected := resolution * resolution
	if len(heights) != expected {
		return nil, fmt.Errorf("terrain collision expected %d heights, got %d", expected, len(heights))
	}
	if minHeight > maxHeight {
		minHeight, maxHeight = maxHeight, minHeight
	}
	if copyHeights {
		heights = append([]matrix.Float(nil), heights...)
	}
	c := &TerrainCollision{
		Resolution: resolution,
		WorldSize:  worldSize,
		Heights:    heights,
		MinHeight:  minHeight,
		MaxHeight:  maxHeight,
	}
	c.RefreshBounds()
	return c, nil
}

func (c *TerrainCollision) Height(x, z int) matrix.Float {
	if c == nil || !c.inBounds(x, z) {
		return 0
	}
	return c.Heights[c.index(x, z)]
}

func (c *TerrainCollision) SampleGrid(x, z matrix.Float) matrix.Float {
	if c == nil || !c.valid() {
		return 0
	}
	maxGrid := matrix.Float(c.Resolution - 1)
	x = matrix.Clamp(x, 0, maxGrid)
	z = matrix.Clamp(z, 0, maxGrid)
	x0 := int(matrix.Floor(x))
	z0 := int(matrix.Floor(z))
	x1 := min(x0+1, c.Resolution-1)
	z1 := min(z0+1, c.Resolution-1)
	tx := x - matrix.Float(x0)
	tz := z - matrix.Float(z0)
	h00 := c.Height(x0, z0)
	h10 := c.Height(x1, z0)
	h01 := c.Height(x0, z1)
	h11 := c.Height(x1, z1)
	return matrix.Lerp(matrix.Lerp(h00, h10, tx), matrix.Lerp(h01, h11, tx), tz)
}

func (c *TerrainCollision) HeightAtLocal(localXZ matrix.Vec2) matrix.Float {
	x, z := c.LocalToGrid(localXZ)
	return c.SampleGrid(x, z)
}

func (c *TerrainCollision) NormalAtLocal(localXZ matrix.Vec2) matrix.Vec3 {
	if c == nil || !c.valid() {
		return matrix.Vec3Up()
	}
	x, z := c.LocalToGrid(localXZ)
	cellX := c.WorldSize.X() / matrix.Float(c.Resolution-1)
	cellZ := c.WorldSize.Y() / matrix.Float(c.Resolution-1)
	left := c.SampleGrid(x-1, z)
	right := c.SampleGrid(x+1, z)
	front := c.SampleGrid(x, z-1)
	back := c.SampleGrid(x, z+1)
	dx := (right - left) / (cellX * 2)
	dz := (back - front) / (cellZ * 2)
	return matrix.NewVec3(-dx, 1, -dz).Normal()
}

func (c *TerrainCollision) GridToLocal(x, z matrix.Float) matrix.Vec3 {
	if c == nil || !c.valid() {
		return matrix.Vec3Zero()
	}
	gx := x / matrix.Float(c.Resolution-1)
	gz := z / matrix.Float(c.Resolution-1)
	return matrix.NewVec3(
		(gx-0.5)*c.WorldSize.X(),
		c.SampleGrid(x, z),
		(gz-0.5)*c.WorldSize.Y(),
	)
}

func (c *TerrainCollision) LocalToGrid(localXZ matrix.Vec2) (matrix.Float, matrix.Float) {
	if c == nil || c.Resolution < 2 || c.WorldSize.X() == 0 || c.WorldSize.Y() == 0 {
		return 0, 0
	}
	x := ((localXZ.X() / c.WorldSize.X()) + 0.5) * matrix.Float(c.Resolution-1)
	z := ((localXZ.Y() / c.WorldSize.Y()) + 0.5) * matrix.Float(c.Resolution-1)
	return x, z
}

func (c *TerrainCollision) LocalBounds() AABB {
	if c == nil {
		return NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero())
	}
	return c.Bounds
}

func (c *TerrainCollision) RefreshBounds() AABB {
	if c == nil {
		return NewAABB(matrix.Vec3Zero(), matrix.Vec3Zero())
	}
	minPoint := matrix.NewVec3(-c.WorldSize.X()*0.5, c.MinHeight, -c.WorldSize.Y()*0.5)
	maxPoint := matrix.NewVec3(c.WorldSize.X()*0.5, c.MaxHeight, c.WorldSize.Y()*0.5)
	c.Bounds = AABBFromMinMax(minPoint, maxPoint)
	return c.Bounds
}

func (c *TerrainCollision) CellRangeForLocalAABB(bounds AABB) (minX, minZ, maxX, maxZ int, ok bool) {
	if c == nil || !c.valid() {
		return 0, 0, 0, 0, false
	}
	terrainBounds := c.LocalBounds()
	if !terrainBounds.AABBIntersect(bounds) {
		return 0, 0, 0, 0, false
	}
	queryMin := bounds.Min()
	queryMax := bounds.Max()
	terrainMin := terrainBounds.Min()
	terrainMax := terrainBounds.Max()
	localMin := matrix.NewVec2(
		matrix.Clamp(queryMin.X(), terrainMin.X(), terrainMax.X()),
		matrix.Clamp(queryMin.Z(), terrainMin.Z(), terrainMax.Z()),
	)
	localMax := matrix.NewVec2(
		matrix.Clamp(queryMax.X(), terrainMin.X(), terrainMax.X()),
		matrix.Clamp(queryMax.Z(), terrainMin.Z(), terrainMax.Z()),
	)
	gridMinX, gridMinZ := c.LocalToGrid(localMin)
	gridMaxX, gridMaxZ := c.LocalToGrid(localMax)
	minX = int(matrix.Floor(matrix.Min(gridMinX, gridMaxX)))
	minZ = int(matrix.Floor(matrix.Min(gridMinZ, gridMaxZ)))
	maxX = int(matrix.Floor(matrix.Max(gridMinX, gridMaxX)))
	maxZ = int(matrix.Floor(matrix.Max(gridMinZ, gridMaxZ)))
	lastCell := c.Resolution - 2
	minX = max(0, min(minX, lastCell))
	minZ = max(0, min(minZ, lastCell))
	maxX = max(0, min(maxX, lastCell))
	maxZ = max(0, min(maxZ, lastCell))
	return minX, minZ, maxX, maxZ, minX <= maxX && minZ <= maxZ
}

func (c *TerrainCollision) ForEachTriangleInLocalAABB(bounds AABB, visit func(DetailedTriangle) bool) {
	if visit == nil {
		return
	}
	minX, minZ, maxX, maxZ, ok := c.CellRangeForLocalAABB(bounds)
	if !ok {
		return
	}
	for z := minZ; z <= maxZ; z++ {
		for x := minX; x <= maxX; x++ {
			p0 := c.GridToLocal(matrix.Float(x), matrix.Float(z))
			p1 := c.GridToLocal(matrix.Float(x), matrix.Float(z+1))
			p2 := c.GridToLocal(matrix.Float(x+1), matrix.Float(z+1))
			p3 := c.GridToLocal(matrix.Float(x+1), matrix.Float(z))
			tri := DetailedTriangleFromPoints([3]matrix.Vec3{p0, p1, p2})
			if bounds.TriangleIntersect(tri) && !visit(tri) {
				return
			}
			tri = DetailedTriangleFromPoints([3]matrix.Vec3{p0, p2, p3})
			if bounds.TriangleIntersect(tri) && !visit(tri) {
				return
			}
		}
	}
}

func (c *TerrainCollision) valid() bool {
	return c.Resolution >= 2 &&
		c.WorldSize.X() > 0 &&
		c.WorldSize.Y() > 0 &&
		len(c.Heights) == c.Resolution*c.Resolution
}

func (c *TerrainCollision) inBounds(x, z int) bool {
	return c.valid() && x >= 0 && z >= 0 && x < c.Resolution && z < c.Resolution
}

func (c *TerrainCollision) index(x, z int) int { return x + z*c.Resolution }
