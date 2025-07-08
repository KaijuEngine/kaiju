/******************************************************************************/
/* mesh.go                                                                    */
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

package rendering

import (
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
)

type MeshDrawMode = int
type MeshCullMode = int
type QuadPivot = int32

const (
	MeshDrawModePoints MeshDrawMode = iota
	MeshDrawModeLines
	MeshDrawModeTriangles
	MeshDrawModePatches
)

const (
	MeshCullModeBack MeshCullMode = iota
	MeshCullModeFront
	MeshCullModeNone
)

const (
	QuadPivotCenter = QuadPivot(iota)
	QuadPivotLeft
	QuadPivotTop
	QuadPivotRight
	QuadPivotBottom
	QuadPivotBottomLeft
	QuadPivotBottomRight
	QuadPivotTopLeft
	QuadPivotTopRight
)

type Mesh struct {
	MeshId         MeshId
	key            string
	pendingVerts   []Vertex
	pendingIndexes []uint32
	Details        meshDetails
	bvh            *collision.BVH
}

func (m *Mesh) BVH() *collision.BVH { return m.bvh }

func NewMesh(key string, verts []Vertex, indexes []uint32) *Mesh {
	defer tracing.NewRegion("rendering.NewMesh").End()
	m := &Mesh{
		key:            key,
		pendingVerts:   verts,
		pendingIndexes: indexes,
	}
	m.generateMeshBVH(verts, indexes)
	// TODO:  Is the following line needed anymore since we're creating a bvh?
	m.Details.Set(verts, indexes)
	return m
}

func (m *Mesh) Destroy(renderer Renderer) {
	defer tracing.NewRegion("Mesh.Destroy").End()
	renderer.DestroyMesh(m)
}

func (m *Mesh) generateMeshBVH(verts []Vertex, indexes []uint32) {
	defer tracing.NewRegion("Mesh.generateMeshBVH").End()
	idxLen := len(indexes)
	if idxLen == 0 || idxLen%3 != 0 {
		// We're doing some special stuff here, probably lines or grids
		return
	}
	tris := make([]collision.DetailedTriangle, len(indexes)/3)
	construct := func(from, to int) {
		for i := from; i < to; i += 3 {
			for i := 0; i < len(indexes); i += 3 {
				points := [3]matrix.Vec3{
					verts[indexes[i]].Position,
					verts[indexes[i+1]].Position,
					verts[indexes[i+2]].Position,
				}
				tris[i/3] = collision.DetailedTriangleFromPoints(points)
			}
		}
	}
	for i := range len(tris) {
		construct(i*3, (i+3)*3)
	}
	m.bvh = collision.BVHBottomUp(tris)
}

func (m *Mesh) SetKey(key string) {
	m.key = key
}

func (m *Mesh) DelayedCreate(renderer Renderer) {
	defer tracing.NewRegion("Mesh.DelayedCreate").End()
	if len(m.pendingVerts) > 0 {
		renderer.CreateMesh(m, m.pendingVerts, m.pendingIndexes)
		m.pendingVerts = m.pendingVerts[:0]
		m.pendingIndexes = m.pendingIndexes[:0]
	}
}

func (m Mesh) Key() string   { return m.key }
func (m Mesh) IsReady() bool { return m.MeshId.IsValid() }

var (
	meshQuadUvs         = [4]matrix.Vec2{{0, 1}, {0, 0}, {1, 0}, {1, 1}}
	meshQuadIndexes     = [6]uint32{0, 2, 1, 0, 3, 2}
	meshQuadCenter      = [4]matrix.Vec3{{-0.5, -0.5, 0}, {-0.5, 0.5, 0}, {0.5, 0.5, 0}, {0.5, -0.5, 0}}
	meshQuadLeft        = [4]matrix.Vec3{{0, -0.5, 0}, {0, 0.5, 0}, {1, 0.5, 0}, {1, -0.5, 0}}
	meshQuadTop         = [4]matrix.Vec3{{-0.5, -1, 0}, {-0.5, 0, 0}, {0.5, 0, 0}, {0.5, -1, 0}}
	meshQuadRight       = [4]matrix.Vec3{{-1, -0.5, 0}, {-1, 0.5, 0}, {0, 0.5, 0}, {0, -0.5, 0}}
	meshQuadBottom      = [4]matrix.Vec3{{-0.5, 0, 0}, {-0.5, 1, 0}, {0.5, 1, 0}, {0.5, 0, 0}}
	meshQuadBottomLeft  = [4]matrix.Vec3{{0, 0, 0}, {0, 1, 0}, {1, 1, 0}, {1, 0, 0}}
	meshQuadBottomRight = [4]matrix.Vec3{{-1, 0, 0}, {-1, 1, 0}, {0, 1, 0}, {0, 0, 0}}
	meshQuadTopLeft     = [4]matrix.Vec3{{0, -1, 0}, {0, 0, 0}, {1, 0, 0}, {1, -1, 0}}
	meshQuadTopRight    = [4]matrix.Vec3{{-1, -1, 0}, {-1, 0, 0}, {0, 0, 0}, {0, -1, 0}}
)

func newMeshQuad(key string, points [4]matrix.Vec3, cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.newMeshQuad").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, len(points))
		for i := range points {
			verts[i].Position = points[i]
			verts[i].Normal = matrix.Vec3{0.0, 0.0, 1.0}
			verts[i].UV0 = meshQuadUvs[i]
			verts[i].Color = matrix.ColorWhite()
		}
		return cache.Mesh(key, verts, meshQuadIndexes[:])
	}
}

func NewMeshQuad(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshQuad").End()
	return NewMeshQuadAnchored(QuadPivotCenter, cache)
}

func NewMeshQuadAnchored(anchor QuadPivot, cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshQuadAnchored").End()
	switch anchor {
	case QuadPivotLeft:
		return newMeshQuad("quad_left", meshQuadLeft, cache)
	case QuadPivotTop:
		return newMeshQuad("quad_top", meshQuadTop, cache)
	case QuadPivotRight:
		return newMeshQuad("quad_right", meshQuadRight, cache)
	case QuadPivotBottom:
		return newMeshQuad("quad_bottom", meshQuadBottom, cache)
	case QuadPivotBottomLeft:
		return newMeshQuad("quad_bottom_left", meshQuadBottomLeft, cache)
	case QuadPivotBottomRight:
		return newMeshQuad("quad_bottom_right", meshQuadBottomRight, cache)
	case QuadPivotTopLeft:
		return newMeshQuad("quad_top_left", meshQuadTopLeft, cache)
	case QuadPivotTopRight:
		return newMeshQuad("quad_top_right", meshQuadTopRight, cache)
	case QuadPivotCenter:
		fallthrough
	default:
		return newMeshQuad("quad", meshQuadCenter, cache)
	}
}

func NewMeshTriangle(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshTriangle").End()
	const key = "triangle"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, 3)
		verts[0].Position = matrix.Vec3{-0.5, -0.5, 0.0}
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = matrix.ColorWhite()
		verts[1].Position = matrix.Vec3{0.0, 0.5, 0.0}
		verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[1].Color = matrix.ColorWhite()
		verts[2].Position = matrix.Vec3{0.5, -0.5, 0.0}
		verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[2].Color = matrix.ColorWhite()
		indexes := []uint32{0, 2, 1}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshUnitQuad(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshUnitQuad").End()
	const key = "unit_quad"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, 4)
		verts[0].Position = matrix.Vec3{-1.0, -1.0, 0.0}
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = matrix.ColorWhite()
		verts[1].Position = matrix.Vec3{-1.0, 1.0, 0.0}
		verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[1].Color = matrix.ColorWhite()
		verts[2].Position = matrix.Vec3{1.0, 1.0, 0.0}
		verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[2].Color = matrix.ColorWhite()
		verts[3].Position = matrix.Vec3{1.0, -1.0, 0.0}
		verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[3].UV0 = matrix.Vec2{1.0, 1.0}
		verts[3].Color = matrix.ColorWhite()
		indexes := []uint32{0, 2, 1, 0, 3, 2}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshScreenQuad(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshScreenQuad").End()
	const key = "screen_quad"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, 4)
		verts[0].Position = matrix.Vec3{0.0, 0.0, 0.0}
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = matrix.ColorWhite()
		verts[1].Position = matrix.Vec3{0.0, 1.0, 0.0}
		verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[1].Color = matrix.ColorWhite()
		verts[2].Position = matrix.Vec3{1.0, 1.0, 0.0}
		verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[2].Color = matrix.ColorWhite()
		verts[3].Position = matrix.Vec3{1.0, 0.0, 0.0}
		verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[3].UV0 = matrix.Vec2{1.0, 1.0}
		verts[3].Color = matrix.ColorWhite()
		indexes := []uint32{0, 2, 1, 0, 3, 2}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshPlane(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshPlane").End()
	const key = "plane"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, 4)
		verts[0].Position = matrix.Vec3{-0.5, 0.0, 0.5}
		verts[0].Normal = matrix.Vec3{0.0, 1.0, 0.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = matrix.ColorWhite()
		verts[1].Position = matrix.Vec3{-0.5, 0.0, -0.5}
		verts[1].Normal = matrix.Vec3{0.0, 1.0, 0.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[1].Color = matrix.ColorWhite()
		verts[2].Position = matrix.Vec3{0.5, 0.0, -0.5}
		verts[2].Normal = matrix.Vec3{0.0, 1.0, 0.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[2].Color = matrix.ColorWhite()
		verts[3].Position = matrix.Vec3{0.5, 0.0, 0.5}
		verts[3].Normal = matrix.Vec3{0.0, 1.0, 0.0}
		verts[3].UV0 = matrix.Vec2{1.0, 1.0}
		verts[3].Color = matrix.ColorWhite()
		indexes := []uint32{0, 2, 1, 0, 3, 2}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshCube(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshCube").End()
	const key = "cube"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, 8)
		verts[0].Position = matrix.Vec3{-0.5, -0.5, 0.5} // 0 - lbf
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = matrix.ColorWhite()
		verts[1].Position = matrix.Vec3{-0.5, 0.5, 0.5} // 1 - ltf
		verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[1].Color = matrix.ColorWhite()
		verts[2].Position = matrix.Vec3{0.5, 0.5, 0.5} // 2 - rtf
		verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[2].Color = matrix.ColorWhite()
		verts[3].Position = matrix.Vec3{0.5, -0.5, 0.5} // 3 - rbf
		verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[3].UV0 = matrix.Vec2{1.0, 1.0}
		verts[3].Color = matrix.ColorWhite()
		verts[4].Position = matrix.Vec3{-0.5, -0.5, -0.5} // 4 - lbb
		verts[4].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[4].UV0 = matrix.Vec2{1.0, 1.0}
		verts[4].Color = matrix.ColorWhite()
		verts[5].Position = matrix.Vec3{-0.5, 0.5, -0.5} // 5 - ltb
		verts[5].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[5].UV0 = matrix.Vec2{1.0, 0.0}
		verts[5].Color = matrix.ColorWhite()
		verts[6].Position = matrix.Vec3{0.5, 0.5, -0.5} // 6 - rtb
		verts[6].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[6].UV0 = matrix.Vec2{0.0, 0.0}
		verts[6].Color = matrix.ColorWhite()
		verts[7].Position = matrix.Vec3{0.5, -0.5, -0.5} // 7 - rbb
		verts[7].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[7].UV0 = matrix.Vec2{0.0, 1.0}
		verts[7].Color = matrix.ColorWhite()
		indexes := []uint32{
			5, 2, 6, 2, 0, 3,
			1, 4, 0, 7, 0, 4,
			6, 3, 7, 5, 7, 4,
			5, 1, 2, 2, 1, 0,
			1, 5, 4, 7, 3, 0,
			6, 2, 3, 5, 6, 7,
		}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshTexturableCube(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshTexturableCube").End()
	const key = "texturable_cube"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, 36)
		for i := 0; i < 36; i++ {
			verts[i].Color = matrix.ColorWhite()
			if i%4 == 0 {
				verts[i+0].UV0 = matrix.Vec2{0.0, 1.0}
				verts[i+1].UV0 = matrix.Vec2{0.0, 0.0}
				verts[i+2].UV0 = matrix.Vec2{1.0, 0.0}
				verts[i+3].UV0 = matrix.Vec2{1.0, 1.0}
			}
		}
		for i := 0; i < 4; i++ {
			verts[i].Normal = matrix.Vec3Forward()
		}
		for i := 4; i < 8; i++ {
			verts[i].Normal = matrix.Vec3Left()
		}
		for i := 8; i < 12; i++ {
			verts[i].Normal = matrix.Vec3Backward()
		}
		for i := 12; i < 16; i++ {
			verts[i].Normal = matrix.Vec3Right()
		}
		for i := 16; i < 20; i++ {
			verts[i].Normal = matrix.Vec3Up()
		}
		for i := 20; i < 24; i++ {
			verts[i].Normal = matrix.Vec3Down()
		}

		offset := 0
		verts[offset].Position = matrix.Vec3{-0.5, -0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, 0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, 0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, -0.5, 0.5}
		offset++

		verts[offset].Position = matrix.Vec3{-0.5, -0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, 0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, 0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, -0.5, 0.5}
		offset++

		verts[offset].Position = matrix.Vec3{0.5, -0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, 0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, 0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, -0.5, -0.5}
		offset++

		verts[offset].Position = matrix.Vec3{0.5, -0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, 0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, 0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, -0.5, -0.5}
		offset++

		verts[offset].Position = matrix.Vec3{-0.5, 0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, 0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, 0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, 0.5, 0.5}
		offset++

		verts[offset].Position = matrix.Vec3{-0.5, -0.5, -0.5}
		offset++
		verts[offset].Position = matrix.Vec3{-0.5, -0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, -0.5, 0.5}
		offset++
		verts[offset].Position = matrix.Vec3{0.5, -0.5, -0.5}
		offset++

		indexes := []uint32{
			0, 2, 1, 0, 3, 2,
			4, 6, 5, 4, 7, 6,
			8, 10, 9, 8, 11, 10,
			12, 14, 13, 12, 15, 14,
			16, 18, 17, 16, 19, 18,
			20, 22, 21, 20, 23, 22,
		}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshSkyboxCube(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshSkyboxCube").End()
	const key = "skybox"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := make([]Vertex, 8)
		verts[0].Position = matrix.Vec3{-0.5, -0.5, 0.5} // 0 - lbf
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = matrix.ColorWhite()
		verts[1].Position = matrix.Vec3{-0.5, 0.5, 0.5} // 1 - ltf
		verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[1].Color = matrix.ColorWhite()
		verts[2].Position = matrix.Vec3{0.5, 0.5, 0.5} // 2 - rtf
		verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[2].Color = matrix.ColorWhite()
		verts[3].Position = matrix.Vec3{0.5, -0.5, 0.5} // 3 - rbf
		verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[3].UV0 = matrix.Vec2{1.0, 1.0}
		verts[3].Color = matrix.ColorWhite()
		verts[4].Position = matrix.Vec3{-0.5, -0.5, -0.5} // 4 - lbb
		verts[4].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[4].UV0 = matrix.Vec2{1.0, 1.0}
		verts[4].Color = matrix.ColorWhite()
		verts[5].Position = matrix.Vec3{-0.5, 0.5, -0.5} // 5 - ltb
		verts[5].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[5].UV0 = matrix.Vec2{1.0, 0.0}
		verts[5].Color = matrix.ColorWhite()
		verts[6].Position = matrix.Vec3{0.5, 0.5, -0.5} // 6 - rtb
		verts[6].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[6].UV0 = matrix.Vec2{0.0, 0.0}
		verts[6].Color = matrix.ColorWhite()
		verts[7].Position = matrix.Vec3{0.5, -0.5, -0.5} // 7 - rbb
		verts[7].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[7].UV0 = matrix.Vec2{0.0, 1.0}
		verts[7].Color = matrix.ColorWhite()
		indexes := []uint32{
			5, 4, 7, 7, 6, 5,
			0, 4, 5, 5, 1, 0,
			7, 3, 2, 2, 6, 7,
			0, 1, 2, 2, 3, 0,
			5, 6, 2, 2, 1, 5,
			4, 0, 7, 7, 0, 3,
		}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshCubeInverse(cache *MeshCache) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshCubeInverse").End()
	const key = "cube_inverse"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := meshCube(matrix.ColorWhite())
		indexes := []uint32{
			5, 4, 7, 7, 6, 5,
			0, 4, 5, 5, 1, 0,
			7, 3, 2, 2, 6, 7,
			0, 1, 2, 2, 3, 0,
			5, 6, 2, 2, 1, 5,
			4, 0, 7, 7, 0, 3,
		}
		return cache.Mesh(key, verts, indexes)
	}
}

func meshCube(vertColor matrix.Color) []Vertex {
	defer tracing.NewRegion("rendering.meshCube").End()
	var verts = make([]Vertex, 8)
	verts[0].Position = matrix.Vec3{-0.5, -0.5, 0.5} // 0
	verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[0].UV0 = matrix.Vec2{0.0, 1.0}
	verts[0].Color = vertColor
	verts[1].Position = matrix.Vec3{-0.5, 0.5, 0.5} // 1
	verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[1].UV0 = matrix.Vec2{0.0, 0.0}
	verts[1].Color = vertColor
	verts[2].Position = matrix.Vec3{0.5, 0.5, 0.5} // 2
	verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[2].UV0 = matrix.Vec2{1.0, 0.0}
	verts[2].Color = vertColor
	verts[3].Position = matrix.Vec3{0.5, -0.5, 0.5} // 3
	verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[3].UV0 = matrix.Vec2{1.0, 1.0}
	verts[3].Color = vertColor
	verts[4].Position = matrix.Vec3{-0.5, -0.5, -0.5} // 4
	verts[4].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[4].UV0 = matrix.Vec2{0.0, 1.0}
	verts[4].Color = vertColor
	verts[5].Position = matrix.Vec3{-0.5, 0.5, -0.5} // 5
	verts[5].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[5].UV0 = matrix.Vec2{0.0, 0.0}
	verts[5].Color = vertColor
	verts[6].Position = matrix.Vec3{0.5, 0.5, -0.5} // 6
	verts[6].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[6].UV0 = matrix.Vec2{1.0, 0.0}
	verts[6].Color = vertColor
	verts[7].Position = matrix.Vec3{0.5, -0.5, -0.5} // 7
	verts[7].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[7].UV0 = matrix.Vec2{1.0, 1.0}
	verts[7].Color = vertColor
	return verts
}

func NewMeshFrustum(cache *MeshCache, key string, inverseProjection matrix.Mat4) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshFrustum").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := meshCube(matrix.ColorWhite())
		for i := 0; i < 8; i++ {
			verts[i].Position.ScaleAssign(2.0)
			verts[i].Position = inverseProjection.TransformPoint(verts[i].Position)
		}
		indexes := []uint32{
			0, 1, 2, 3, 0,
			4, 5, 6, 7, 4,
			0, 1, 5, 6,
			2, 3, 7,
		}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshOffsetQuad(cache *MeshCache, key string, sideOffsets matrix.Vec4) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshOffsetQuad").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		l := sideOffsets.X()
		t := -sideOffsets.Y()
		r := -sideOffsets.Z()
		b := sideOffsets.W()
		var verts = make([]Vertex, 4)
		verts[0].Position = matrix.Vec3{-0.5 + l, -0.5 + b, 0.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[1].Position = matrix.Vec3{-0.5 + l, 0.5 + t, 0.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[2].Position = matrix.Vec3{0.5 + r, 0.5 + t, 0.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[3].Position = matrix.Vec3{0.5 + r, -0.5 + b, 0.0}
		verts[3].UV0 = matrix.Vec2{1.0, 1.0}
		for i := 0; i < 4; i++ {
			verts[i].Normal = matrix.Vec3{0.0, 0.0, 1.0}
			verts[i].Color = matrix.ColorWhite()
		}
		indexes := []uint32{0, 1, 2, 0, 2, 3}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshGrid(cache *MeshCache, key string, points []matrix.Vec3, vertColor matrix.Color) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshGrid").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		length := uint32(len(points))
		if length%2 != 0 {
			panic("points length must be even")
		}
		var verts = make([]Vertex, length)
		var indexes = make([]uint32, length)
		for i := uint32(0); i < length; i++ {
			verts[i].Position = points[i]
			verts[i].Normal = matrix.Vec3{0.0, 0.0, 1.0}
			verts[i].UV0 = matrix.Vec2{0.0, 1.0}
			verts[i].Color = vertColor
			indexes[i] = i
		}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshPoint(cache *MeshCache, key string, position matrix.Vec3, vertColor matrix.Color) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshPoint").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		var verts = make([]Vertex, 1)
		verts[0].Position = position
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = vertColor
		indexes := []uint32{0}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshLine(cache *MeshCache, key string, p0, p1 matrix.Vec3, vertColor matrix.Color) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshLine").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		var verts = make([]Vertex, 2)
		verts[0].Position = p0
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = vertColor
		verts[1].Position = p1
		verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[1].UV0 = matrix.Vec2{0.0, 1.0}
		verts[1].Color = vertColor
		indexes := []uint32{0, 1}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshWireQuad(cache *MeshCache, key string, vertColor matrix.Color) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshWireQuad").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		var verts = make([]Vertex, 4)
		verts[0].Position = matrix.Vec3{-0.5, -0.5, 0.0}
		verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[0].UV0 = matrix.Vec2{0.0, 1.0}
		verts[0].Color = vertColor
		verts[1].Position = matrix.Vec3{-0.5, 0.5, 0.0}
		verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[1].UV0 = matrix.Vec2{0.0, 0.0}
		verts[1].Color = vertColor
		verts[2].Position = matrix.Vec3{0.5, 0.5, 0.0}
		verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[2].UV0 = matrix.Vec2{1.0, 0.0}
		verts[2].Color = vertColor
		verts[3].Position = matrix.Vec3{0.5, -0.5, 0.0}
		verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[3].UV0 = matrix.Vec2{1.0, 1.0}
		verts[3].Color = vertColor
		var indexes = []uint32{0, 1, 1, 2, 2, 3, 3, 0}
		return cache.Mesh(key, verts, indexes)
	}
}

func NewMeshWireCube(cache *MeshCache, key string, vertColor matrix.Color) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshWireCube").End()
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := meshCube(vertColor)
		indexes := []uint32{
			0, 1, 1, 2, 2, 3, 3, 0,
			0, 4, 1, 5, 2, 6, 3, 7,
			4, 5, 5, 6, 6, 7, 7, 4,
		}
		return cache.Mesh(key, verts, indexes)
	}
}
