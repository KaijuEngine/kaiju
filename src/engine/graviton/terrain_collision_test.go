/******************************************************************************/
/* terrain_collision_test.go                                                  */
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
	"testing"

	"kaijuengine.com/matrix"
)

func TestTerrainCollisionSamplesTerrainLocalCoordinates(t *testing.T) {
	collision := testTerrainCollision(t, 3, matrix.NewVec2(2, 2), []matrix.Float{
		0, 1, 2,
		3, 4, 5,
		6, 7, 8,
	}, 0, 8)
	if got := collision.Height(1, 2); got != 7 {
		t.Fatalf("expected height 7, got %f", got)
	}
	if got := collision.SampleGrid(0.5, 0.5); !matrix.ApproxTo(got, 2, 0.0001) {
		t.Fatalf("expected interpolated grid height 2, got %f", got)
	}
	if got := collision.HeightAtLocal(matrix.NewVec2(0, 0)); !matrix.ApproxTo(got, 4, 0.0001) {
		t.Fatalf("expected center local height 4, got %f", got)
	}
	local := collision.GridToLocal(2, 0)
	if !matrix.Vec3ApproxTo(local, matrix.NewVec3(1, 2, -1), 0.0001) {
		t.Fatalf("expected grid 2,0 to map to local 1,2,-1, got %v", local)
	}
	x, z := collision.LocalToGrid(matrix.NewVec2(-1, 1))
	if !matrix.ApproxTo(x, 0, 0.0001) || !matrix.ApproxTo(z, 2, 0.0001) {
		t.Fatalf("expected local -1,1 to map to grid 0,2, got %f,%f", x, z)
	}
}

func TestTerrainCollisionBoundsAndNormals(t *testing.T) {
	collision := testTerrainCollision(t, 2, matrix.NewVec2(4, 6), []matrix.Float{
		1, 1,
		1, 1,
	}, -2, 8)
	bounds := collision.LocalBounds()
	if !matrix.Vec3ApproxTo(bounds.Center, matrix.NewVec3(0, 3, 0), 0.0001) {
		t.Fatalf("expected bounds center 0,3,0, got %v", bounds.Center)
	}
	if !matrix.Vec3ApproxTo(bounds.Extent, matrix.NewVec3(2, 5, 3), 0.0001) {
		t.Fatalf("expected bounds extent 2,5,3, got %v", bounds.Extent)
	}
	normal := collision.NormalAtLocal(matrix.NewVec2(0, 0))
	if !matrix.Vec3ApproxTo(normal, matrix.Vec3Up(), 0.0001) {
		t.Fatalf("expected flat terrain normal to point up, got %v", normal)
	}
}

func TestTerrainCollisionCellRangeForLocalAABB(t *testing.T) {
	collision := testTerrainCollision(t, 3, matrix.NewVec2(2, 2), make([]matrix.Float, 9), -1, 1)
	bounds := AABBFromMinMax(matrix.NewVec3(-0.1, -0.1, -0.1), matrix.NewVec3(0.1, 0.1, 0.1))
	minX, minZ, maxX, maxZ, ok := collision.CellRangeForLocalAABB(bounds)
	if !ok {
		t.Fatal("expected local bounds to overlap terrain")
	}
	if minX != 0 || minZ != 0 || maxX != 1 || maxZ != 1 {
		t.Fatalf("expected center range [0,0]-[1,1], got [%d,%d]-[%d,%d]", minX, minZ, maxX, maxZ)
	}
	outside := AABBFromMinMax(matrix.NewVec3(3, -0.1, 3), matrix.NewVec3(4, 0.1, 4))
	if _, _, _, _, ok := collision.CellRangeForLocalAABB(outside); ok {
		t.Fatal("expected outside bounds to miss terrain")
	}
}

func TestTerrainCollisionForEachTriangleInLocalAABB(t *testing.T) {
	collision := testTerrainCollision(t, 2, matrix.NewVec2(2, 2), []matrix.Float{
		0, 0,
		0, 0,
	}, -1, 1)
	bounds := AABBFromMinMax(matrix.NewVec3(-1, -0.1, -1), matrix.NewVec3(1, 0.1, 1))
	triangles := make([]DetailedTriangle, 0, 2)
	collision.ForEachTriangleInLocalAABB(bounds, func(triangle DetailedTriangle) bool {
		triangles = append(triangles, triangle)
		return true
	})
	if len(triangles) != 2 {
		t.Fatalf("expected 2 terrain cell triangles, got %d", len(triangles))
	}
	for i := range triangles {
		if !matrix.Vec3ApproxTo(triangles[i].Normal, matrix.Vec3Up(), 0.0001) {
			t.Fatalf("expected triangle %d normal to point up, got %v", i, triangles[i].Normal)
		}
	}
}

func testTerrainCollision(t *testing.T, resolution int, worldSize matrix.Vec2, heights []matrix.Float, minHeight, maxHeight matrix.Float) *TerrainCollision {
	t.Helper()
	collision, err := NewTerrainCollision(resolution, worldSize, heights, minHeight, maxHeight)
	if err != nil {
		t.Fatal(err)
	}
	return collision
}
