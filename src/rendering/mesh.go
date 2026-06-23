/******************************************************************************/
/* mesh.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"fmt"
	"math"
	"slices"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type MeshDrawMode = int
type MeshCullMode = int
type QuadPivot = int32
type PrimitiveMesh string

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

const (
	PrimitiveMeshSphere         PrimitiveMesh = "sphere_1.00_32_32"
	PrimitiveMeshTexturableCube PrimitiveMesh = "texturable_cube"
	PrimitiveMeshCapsule        PrimitiveMesh = "capsule_0.50_1.00_32_8"
	PrimitiveMeshPlane          PrimitiveMesh = "plane"
	PrimitiveMeshCylinder       PrimitiveMesh = "cylinder_1.00_0.50_32_true"
	PrimitiveMeshCone           PrimitiveMesh = "cone_1.00_0.50_32_true"
	PrimitiveMeshArrow          PrimitiveMesh = "arrow_0.75_0.05_0.25_0.15_32_"
)

type Mesh struct {
	MeshId         MeshId
	key            string
	pendingVerts   []Vertex
	pendingIndexes []uint32
	bounds         graviton.AABB
	dynamic        bool
}

func NewMesh(key string, verts []Vertex, indexes []uint32) *Mesh {
	defer tracing.NewRegion("rendering.NewMesh").End()
	m := &Mesh{
		key:            key,
		pendingVerts:   verts,
		pendingIndexes: indexes,
	}
	if len(verts) > 0 {
		low, high := verts[0].Position, verts[0].Position
		for i := 1; i < len(verts); i++ {
			low = matrix.Vec3Min(low, verts[i].Position)
			high = matrix.Vec3Max(high, verts[i].Position)
		}
		m.bounds = graviton.AABBFromMinMax(low, high)
	}
	return m
}

func NewDynamicMesh(key string, verts []Vertex, indexes []uint32) *Mesh {
	m := NewMesh(key, verts, indexes)
	m.dynamic = true
	return m
}

func (m *Mesh) SetKey(key string) {
	m.key = key
}

func (m *Mesh) DelayedCreate(device *GPUDevice) {
	defer tracing.NewRegion("Mesh.DelayedCreate").End()
	if len(m.pendingVerts) == 0 {
		return
	}
	if m.IsReady() {
		if m.dynamic {
			device.UpdateDynamicMeshVertices(m, m.pendingVerts)
		} else {
			device.UpdateMeshVertices(m, m.pendingVerts)
		}
	} else {
		if m.dynamic {
			device.CreateDynamicMesh(m, m.pendingVerts, m.pendingIndexes)
		} else {
			device.CreateMesh(m, m.pendingVerts, m.pendingIndexes)
		}
	}
	m.pendingVerts = make([]Vertex, 0)
	m.pendingIndexes = make([]uint32, 0)
}

func (m Mesh) Key() string           { return m.key }
func (m Mesh) IsReady() bool         { return m.MeshId.IsValid() }
func (m Mesh) Bounds() graviton.AABB { return m.bounds }

func (m *Mesh) SetPendingVertices(verts []Vertex) {
	m.pendingVerts = slices.Clone(verts)
	if len(verts) == 0 {
		return
	}
	low, high := verts[0].Position, verts[0].Position
	for i := 1; i < len(verts); i++ {
		low = matrix.Vec3Min(low, verts[i].Position)
		high = matrix.Vec3Max(high, verts[i].Position)
	}
	m.bounds = graviton.AABBFromMinMax(low, high)
}

func NewMeshPrimitive(cache *MeshCache, primitive PrimitiveMesh) *Mesh {
	switch primitive {
	case PrimitiveMeshSphere:
		return NewMeshSphere(cache, 1, 32, 32)
	case PrimitiveMeshTexturableCube:
		return NewMeshTexturableCube(cache)
	case PrimitiveMeshCapsule:
		return NewMeshCapsule(cache, 0.5, 1, 32, 8)
	case PrimitiveMeshPlane:
		return NewMeshPlane(cache)
	case PrimitiveMeshCylinder:
		return NewMeshCylinder(cache, 1, 0.5, 32, true)
	case PrimitiveMeshCone:
		return NewMeshCone(cache, 1, 0.5, 32, true)
	case PrimitiveMeshArrow:
		return NewMeshArrow(cache, 0.75, 0.05, 0.25, 0.15, 32)
	default:
		return nil
	}
}

func BuiltInMeshData(key string) ([]Vertex, []uint32, bool) {
	switch key {
	case "quad":
		verts, indexes := MeshQuadData()
		return verts, slices.Clone(indexes), true
	case "plane":
		verts, indexes := MeshPlaneData()
		return verts, slices.Clone(indexes), true
	case "cube":
		return builtInGeneratedMeshData(func(cache *MeshCache) *Mesh {
			return NewMeshCube(cache)
		})
	case string(PrimitiveMeshSphere):
		return builtInGeneratedMeshData(func(cache *MeshCache) *Mesh {
			return NewMeshSphere(cache, 1, 32, 32)
		})
	case string(PrimitiveMeshTexturableCube):
		return builtInGeneratedMeshData(func(cache *MeshCache) *Mesh {
			return NewMeshTexturableCube(cache)
		})
	case string(PrimitiveMeshCapsule):
		return builtInGeneratedMeshData(func(cache *MeshCache) *Mesh {
			return NewMeshCapsule(cache, 0.5, 1, 32, 8)
		})
	case string(PrimitiveMeshCylinder):
		return builtInGeneratedMeshData(func(cache *MeshCache) *Mesh {
			return NewMeshCylinder(cache, 1, 0.5, 32, true)
		})
	case string(PrimitiveMeshCone):
		return builtInGeneratedMeshData(func(cache *MeshCache) *Mesh {
			return NewMeshCone(cache, 1, 0.5, 32, true)
		})
	case string(PrimitiveMeshArrow):
		return builtInGeneratedMeshData(func(cache *MeshCache) *Mesh {
			return NewMeshArrow(cache, 0.75, 0.05, 0.25, 0.15, 32)
		})
	default:
		return nil, nil, false
	}
}

func builtInGeneratedMeshData(create func(*MeshCache) *Mesh) ([]Vertex, []uint32, bool) {
	cache := NewMeshCache(nil, nil)
	mesh := create(&cache)
	return slices.Clone(mesh.pendingVerts), slices.Clone(mesh.pendingIndexes), true
}

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

func MeshQuadData() ([]Vertex, []uint32) {
	verts := make([]Vertex, len(meshQuadCenter))
	for i := range meshQuadCenter {
		verts[i].Position = meshQuadCenter[i]
		verts[i].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[i].UV0 = meshQuadUvs[i]
		verts[i].Color = matrix.ColorWhite()
	}
	return verts, meshQuadIndexes[:]
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
		verts, indexes := MeshPlaneData()
		return cache.Mesh(key, verts, indexes)
	}
}

func MeshPlaneData() ([]Vertex, []uint32) {
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
	return verts, []uint32{0, 2, 1, 0, 3, 2}
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

func NewMeshFrustumBox(cache *MeshCache, inverseProjection matrix.Mat4) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshFrustum").End()
	const key = "frustum_box"
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	} else {
		verts := meshCube(matrix.ColorWhite())
		for i := 0; i < 8; i++ {
			verts[i].Position.ScaleAssign(2.0)
		}
		indexes := []uint32{
			0, 1, 1, 2, 2, 3, 3, 0,
			4, 5, 5, 6, 6, 7, 7, 4,
			0, 4, 1, 5, 2, 6, 3, 7,
		}
		return cache.Mesh(key, verts, indexes)
	}
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
			0, 1, 1, 2, 2, 3, 3, 0,
			4, 5, 5, 6, 6, 7, 7, 4,
			0, 4, 1, 5, 2, 6, 3, 7,
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

// NewMeshCapsule creates a capsule mesh (cylinder with hemispherical ends) with the specified radius and height.
// The capsule is aligned along the Y-axis, with hemispheres at y=height/2 and y=-height/2.
// segments controls the number of subdivisions around the circumference, rings controls the number of rings per hemisphere.
func NewMeshCapsule(cache *MeshCache, radius, height float32, segments, rings int) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshCapsule").End()
	if segments < 3 {
		segments = 3
	}
	if rings < 1 {
		rings = 1
	}
	key := fmt.Sprintf("capsule_%.2f_%.2f_%d_%d", radius, height, segments, rings)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}

	rowCount := rings*2 + 2
	verts := make([]Vertex, rowCount*(segments+1))
	indices := make([]uint32, (rowCount-1)*segments*6)

	vIndex := 0
	topCenter := height / 2
	bottomCenter := -height / 2
	for row := 0; row < rowCount; row++ {
		var y, ringRadius, normalRadius, normalY float32
		if row <= rings {
			theta := float32(row) * math.Pi / (2.0 * float32(rings))
			normalRadius = matrix.Sin(theta)
			normalY = matrix.Cos(theta)
			y = topCenter + radius*normalY
			ringRadius = radius * normalRadius
		} else if row == rings+1 {
			normalRadius = 1
			normalY = 0
			y = bottomCenter
			ringRadius = radius
		} else {
			hemisphereRow := row - (rings + 1)
			theta := math.Pi/2 + float32(hemisphereRow)*math.Pi/(2.0*float32(rings))
			normalRadius = matrix.Sin(theta)
			normalY = matrix.Cos(theta)
			y = bottomCenter + radius*normalY
			ringRadius = radius * normalRadius
		}

		for j := 0; j <= segments; j++ {
			phi := float32(j) * 2.0 * math.Pi / float32(segments)
			sinPhi := matrix.Sin(phi)
			cosPhi := matrix.Cos(phi)
			verts[vIndex].Position = matrix.Vec3{
				ringRadius * cosPhi,
				y,
				ringRadius * sinPhi,
			}
			normal := matrix.Vec3{cosPhi * normalRadius, normalY, sinPhi * normalRadius}
			verts[vIndex].Normal = normal.Normal()
			verts[vIndex].UV0 = matrix.NewVec2(
				float32(j)/float32(segments),
				float32(row)/float32(rowCount-1),
			)
			verts[vIndex].Color = matrix.ColorWhite()
			vIndex++
		}
	}

	iIndex := 0
	for row := 0; row < rowCount-1; row++ {
		rowStart := row * (segments + 1)
		nextRowStart := (row + 1) * (segments + 1)
		for j := 0; j < segments; j++ {
			v00 := uint32(rowStart + j)
			v10 := uint32(rowStart + j + 1)
			v01 := uint32(nextRowStart + j)
			v11 := uint32(nextRowStart + j + 1)
			indices[iIndex] = v00
			indices[iIndex+1] = v10
			indices[iIndex+2] = v01
			indices[iIndex+3] = v01
			indices[iIndex+4] = v10
			indices[iIndex+5] = v11
			iIndex += 6
		}
	}
	return cache.Mesh(key, verts, indices)
}

func NewMeshSphere(cache *MeshCache, radius float32, latitudeBands, longitudeBands int) *Mesh {
	if latitudeBands < 2 {
		latitudeBands = 2
	}
	if longitudeBands < 3 {
		longitudeBands = 3
	}
	key := fmt.Sprintf("sphere_%.2f_%d_%d", radius, latitudeBands, longitudeBands)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	numVerts := (latitudeBands + 1) * (longitudeBands + 1)
	verts := make([]Vertex, numVerts)
	vIdx := 0
	for lat := 0; lat <= latitudeBands; lat++ {
		theta := float32(lat) * math.Pi / float32(latitudeBands) // 0..π
		sinTheta := matrix.Sin(theta)
		cosTheta := matrix.Cos(theta)
		for lon := 0; lon <= longitudeBands; lon++ {
			phi := float32(lon) * 2.0 * math.Pi / float32(longitudeBands) // 0..2π
			sinPhi := matrix.Sin(phi)
			cosPhi := matrix.Cos(phi)
			x := radius * sinTheta * cosPhi
			y := radius * cosTheta
			z := radius * sinTheta * sinPhi
			verts[vIdx].Position = matrix.Vec3{x, y, z}
			if radius != 0 {
				invR := 1.0 / radius
				verts[vIdx].Normal = matrix.Vec3{x * invR, y * invR, z * invR}
			} else {
				verts[vIdx].Normal = matrix.Vec3{0, 0, 0}
			}
			verts[vIdx].UV0 = matrix.NewVec2(
				float32(lon)/float32(longitudeBands),
				float32(lat)/float32(latitudeBands),
			)
			verts[vIdx].Color = matrix.ColorWhite()
			vIdx++
		}
	}
	numIndices := latitudeBands * longitudeBands * 6
	indices := make([]uint32, numIndices)
	iIdx := 0
	for lat := 0; lat < latitudeBands; lat++ {
		for lon := 0; lon < longitudeBands; lon++ {
			first := uint32(lat*(longitudeBands+1) + lon)
			second := first + uint32(longitudeBands+1)
			indices[iIdx] = first
			indices[iIdx+1] = first + 1
			indices[iIdx+2] = second
			indices[iIdx+3] = second
			indices[iIdx+4] = first + 1
			indices[iIdx+5] = second + 1
			iIdx += 6
		}
	}
	return cache.Mesh(key, verts, indices)
}

func NewMeshWireSphere(cache *MeshCache, radius float32, latitudeBands, longitudeBands int) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshWireSphereLatLon").End()
	key := fmt.Sprintf("wire_sphere_latlon_%.2f_%d_%d", radius, latitudeBands, longitudeBands)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	if latitudeBands < 4 {
		latitudeBands = 4
	}
	if longitudeBands < 8 {
		longitudeBands = 8
	}
	numRings := latitudeBands - 1
	numVerts := numRings * longitudeBands
	verts := make([]Vertex, numVerts)
	vIdx := 0
	for ring := 1; ring < latitudeBands; ring++ {
		theta := float32(ring) * math.Pi / float32(latitudeBands)
		sinTheta := matrix.Sin(theta)
		cosTheta := matrix.Cos(theta)
		for lon := 0; lon < longitudeBands; lon++ {
			phi := float32(lon) * 2.0 * math.Pi / float32(longitudeBands)
			sinPhi := matrix.Sin(phi)
			cosPhi := matrix.Cos(phi)
			verts[vIdx].Position = matrix.Vec3{
				radius * cosPhi * sinTheta,
				radius * cosTheta,
				radius * sinPhi * sinTheta,
			}
			verts[vIdx].Normal = matrix.Vec3{0.0, 0.0, 1.0}
			verts[vIdx].UV0 = matrix.Vec2{0.0, 0.0}
			verts[vIdx].Color = matrix.ColorWhite()
			vIdx++
		}
	}
	var indices []uint32
	for ring := 0; ring < numRings; ring++ {
		base := uint32(ring * longitudeBands)
		for lon := 0; lon < longitudeBands; lon++ {
			curr := base + uint32(lon)
			next := base + uint32((lon+1)%longitudeBands)
			indices = append(indices, curr, next)
		}
	}
	for lon := 0; lon < longitudeBands; lon++ {
		for ring := 0; ring < numRings-1; ring++ {
			curr := uint32(ring*longitudeBands + lon)
			next := uint32((ring+1)*longitudeBands + lon)
			indices = append(indices, curr, next)
		}
	}
	return cache.Mesh(key, verts, indices)
}

func NewMeshWireCylinder(cache *MeshCache, radius, height float32, segments, heightSegments int) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshWireCylinder").End()
	key := fmt.Sprintf("wire_cylinder_%.2f_%.2f_%d_%d", radius, height, segments, heightSegments)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	if segments < 8 {
		segments = 8
	}
	if heightSegments < 1 {
		heightSegments = 1
	}
	numRings := heightSegments + 1
	numVerts := numRings * segments
	verts := make([]Vertex, numVerts)
	vIdx := 0
	halfHeight := height / 2.0
	for ring := 0; ring < numRings; ring++ {
		y := -halfHeight + float32(ring)/float32(heightSegments)*height
		for s := 0; s < segments; s++ {
			phi := float32(s) * 2.0 * math.Pi / float32(segments)
			cosPhi := matrix.Cos(phi)
			sinPhi := matrix.Sin(phi)
			verts[vIdx].Position = matrix.Vec3{
				radius * cosPhi,
				y,
				radius * sinPhi,
			}
			verts[vIdx].Normal = matrix.Vec3{0.0, 0.0, 1.0} // not used for wireframe
			verts[vIdx].UV0 = matrix.Vec2{0.0, 0.0}
			verts[vIdx].Color = matrix.ColorWhite()
			vIdx++
		}
	}
	var indices []uint32
	for ring := 0; ring < numRings; ring++ {
		base := uint32(ring * segments)
		for s := 0; s < segments; s++ {
			curr := base + uint32(s)
			next := base + uint32((s+1)%segments)
			indices = append(indices, curr, next)
		}
	}
	for s := 0; s < segments; s++ {
		for ring := 0; ring < heightSegments; ring++ {
			curr := uint32(ring*segments + s)
			next := uint32((ring+1)*segments + s)
			indices = append(indices, curr, next)
		}
	}
	return cache.Mesh(key, verts, indices)
}

func NewMeshWireCone(cache *MeshCache, radius, height float32, segments, heightSegments int) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshWireCone").End()
	key := fmt.Sprintf("wire_cone_%.2f_%.2f_%d_%d", radius, height, segments, heightSegments)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	if segments < 8 {
		segments = 8
	}
	if heightSegments < 1 {
		heightSegments = 1
	}
	numRings := heightSegments + 1
	numVerts := 1 + numRings*segments
	verts := make([]Vertex, numVerts)
	verts[0].Position = matrix.NewVec3(0.0, height/2.0, 0.0)
	verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[0].UV0 = matrix.Vec2{0.0, 0.0}
	verts[0].Color = matrix.ColorWhite()
	halfHeight := height / 2.0
	apexIdx := uint32(0)
	vIdx := 1
	for ring := 0; ring < numRings; ring++ {
		t := float32(ring) / float32(heightSegments)
		y := -halfHeight + t*height
		currentRadius := radius * (1.0 - t)
		for s := 0; s < segments; s++ {
			phi := float32(s) * 2.0 * math.Pi / float32(segments)
			cosPhi := matrix.Cos(phi)
			sinPhi := matrix.Sin(phi)

			verts[vIdx].Position = matrix.Vec3{
				currentRadius * cosPhi,
				y,
				currentRadius * sinPhi,
			}
			verts[vIdx].Normal = matrix.Vec3{0.0, 0.0, 1.0}
			verts[vIdx].UV0 = matrix.Vec2{0.0, 0.0}
			verts[vIdx].Color = matrix.ColorWhite()
			vIdx++
		}
	}
	var indices []uint32
	for ring := 0; ring < numRings; ring++ {
		t := float32(ring) / float32(heightSegments)
		if t >= 1.0 {
			continue
		}
		base := uint32(1 + ring*segments)
		for s := 0; s < segments; s++ {
			curr := base + uint32(s)
			next := base + uint32((s+1)%segments)
			indices = append(indices, curr, next)
		}
	}
	baseRingStart := uint32(1 + (numRings-1)*segments)
	for s := 0; s < segments; s++ {
		baseVert := baseRingStart + uint32(s)
		indices = append(indices, apexIdx, baseVert)
	}
	for s := 0; s < segments; s++ {
		for ring := 0; ring < numRings-1; ring++ {
			curr := uint32(1 + ring*segments + s)
			next := uint32(1 + (ring+1)*segments + s)
			indices = append(indices, curr, next)
		}
		topOfLine := uint32(1 + (numRings-1)*segments + s)
		indices = append(indices, topOfLine, apexIdx)
	}
	return cache.Mesh(key, verts, indices)
}

// NewMeshCircleWire creates a wireframe circle (line loop) mesh.
// It follows the same pattern as the other wireframe mesh generators
// (e.g. NewMeshWireCube, NewMeshWireSphereLatLon, etc.).
func NewMeshCircleWire(cache *MeshCache, radius float32, segments int) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshCircleWire").End()
	if segments < 3 {
		segments = 3
	}
	// Use a distinct cache key for the wireframe version.
	key := fmt.Sprintf("circle_wire_%.2f_%d", radius, segments)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	// Vertex layout: one vertex per segment (no center vertex needed for a line loop).
	verts := make([]Vertex, segments)
	for i := 0; i < segments; i++ {
		phi := float32(i) * 2.0 * math.Pi / float32(segments)
		cosPhi := matrix.Cos(phi)
		sinPhi := matrix.Sin(phi)
		verts[i].Position = matrix.Vec3{
			radius * cosPhi,
			0.0,
			radius * sinPhi,
		}
		verts[i].Normal = matrix.Vec3{0.0, 0.0, 1.0}
		verts[i].UV0 = matrix.Vec2{
			0.5 + 0.5*cosPhi,
			0.5 + 0.5*sinPhi,
		}
		verts[i].Color = matrix.ColorWhite()
	}
	// Line‑loop indices: each vertex connects to the next, wrapping at the end.
	indices := make([]uint32, segments*2)
	idx := 0
	for i := 0; i < segments; i++ {
		indices[idx] = uint32(i)                    // current vertex
		indices[idx+1] = uint32((i + 1) % segments) // next vertex (wrap)
		idx += 2
	}
	return cache.Mesh(key, verts, indices)
}

func NewMeshCylinder(cache *MeshCache, height, radius float32, segments int, capped bool) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshCylinder").End()
	if segments < 3 {
		segments = 3
	}
	key := fmt.Sprintf("cylinder_%.2f_%.2f_%d_%v", height, radius, segments, capped)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	verts, indices := meshCylinder(height, radius, segments, capped)
	for i := range verts {
		verts[i].Position[matrix.Vy] += height * 0.5
	}
	return cache.Mesh(key, verts, indices)
}

func NewMeshCone(cache *MeshCache, height, baseRadius float32, segments int, capped bool) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshCone").End()
	if segments < 3 {
		segments = 3
	}
	key := fmt.Sprintf("cone_%.2f_%.2f_%d_%v", height, baseRadius, segments, capped)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	verts, indices := meshCone(height, baseRadius, segments, capped)
	return cache.Mesh(key, verts, indices)
}

func NewMeshArrowWithTransform(cache *MeshCache, shaftLength, shaftRadius, tipHeight, tipRadius float32, segments int, transform matrix.Mat4, uid string) *Mesh {
	defer tracing.NewRegion("rendering.NewMeshArrow").End()
	if segments < 3 {
		segments = 3
	}
	key := fmt.Sprintf("arrow_%.2f_%.2f_%.2f_%.2f_%d_%s", shaftLength, shaftRadius, tipHeight, tipRadius, segments, uid)
	if mesh, ok := cache.FindMesh(key); ok {
		return mesh
	}
	shaftVerts, shaftIndices := meshCylinder(shaftLength, shaftRadius, segments, true)
	tipVerts, tipIndices := meshCone(tipHeight, tipRadius, segments, true)
	for i := range shaftVerts {
		shaftVerts[i].Position[matrix.Vy] += shaftLength * 0.5
	}
	for i := range tipVerts {
		tipVerts[i].Position[matrix.Vy] += shaftLength
	}
	verts := append(shaftVerts, tipVerts...)

	for i := range verts {
		verts[i].Position = matrix.Mat4MultiplyVec4(transform, verts[i].Position.AsVec4()).AsVec3()
	}

	indices := make([]uint32, 0, len(shaftIndices)+len(tipIndices))
	indices = append(indices, shaftIndices...)
	offset := uint32(len(shaftVerts))
	for i := range tipIndices {
		indices = append(indices, tipIndices[i]+offset)
	}
	return cache.Mesh(key, verts, indices)
}

func NewMeshArrow(cache *MeshCache, shaftLength, shaftRadius, tipHeight, tipRadius float32, segments int) *Mesh {
	return NewMeshArrowWithTransform(cache, shaftLength, shaftRadius,
		tipHeight, tipRadius, segments, matrix.Mat4Identity(), "")
}

func meshCylinder(height, radius float32, segments int, capped bool) ([]Vertex, []uint32) {
	if segments < 3 {
		segments = 3
	}
	numVerts := segments * 2   // Bottom and top rings
	numIndices := segments * 6 // Sides (2 triangles per segment)
	if capped {
		numVerts += segments * 2       // Additional rings for caps, but actually triangles for caps
		numIndices += segments * 3 * 2 // Two caps, each with segments triangles
	}
	verts := make([]Vertex, numVerts)
	indices := make([]uint32, numIndices)
	vIndex := 0
	iIndex := 0
	// Generate bottom and top rings
	halfHeight := height / 2.0
	for i := 0; i < 2; i++ { // 0: bottom, 1: top
		y := -halfHeight + float32(i)*height
		for j := 0; j < segments; j++ {
			phi := float32(j) * 2.0 * math.Pi / float32(segments)
			cosPhi := matrix.Cos(phi)
			sinPhi := matrix.Sin(phi)
			verts[vIndex].Position = matrix.Vec3{radius * cosPhi, y, radius * sinPhi}
			verts[vIndex].Normal = matrix.Vec3{cosPhi, 0.0, sinPhi} // Side normal
			verts[vIndex].UV0 = matrix.NewVec2(float32(j)/float32(segments), float32(i))
			verts[vIndex].Color = matrix.ColorWhite()
			vIndex++
		}
	}
	// Generate side indices
	for j := 0; j < segments; j++ {
		v00 := uint32(j)
		v10 := uint32((j + 1) % segments)
		v01 := uint32(j + segments)
		v11 := uint32((j+1)%segments + segments)
		indices[iIndex] = v00
		indices[iIndex+1] = v01
		indices[iIndex+2] = v10
		indices[iIndex+3] = v01
		indices[iIndex+4] = v11
		indices[iIndex+5] = v10
		iIndex += 6
	}
	if capped {
		// For caps, we need center points or triangulate
		// Add bottom center
		bottomCenter := vIndex
		verts[vIndex].Position = matrix.NewVec3(0.0, -halfHeight, 0.0)
		verts[vIndex].Normal = matrix.Vec3{0.0, -1.0, 0.0}
		verts[vIndex].UV0 = matrix.Vec2{0.5, 0.5}
		verts[vIndex].Color = matrix.ColorWhite()
		vIndex++
		// Add top center
		topCenter := vIndex
		verts[vIndex].Position = matrix.NewVec3(0.0, halfHeight, 0.0)
		verts[vIndex].Normal = matrix.Vec3{0.0, 1.0, 0.0}
		verts[vIndex].UV0 = matrix.Vec2{0.5, 0.5}
		verts[vIndex].Color = matrix.ColorWhite()
		vIndex++
		// Bottom cap indices (note: adjust for winding)
		for j := 0; j < segments; j++ {
			v0 := uint32(j)
			v1 := uint32((j + 1) % segments)
			indices[iIndex] = uint32(bottomCenter)
			indices[iIndex+1] = v0
			indices[iIndex+2] = v1
			iIndex += 3
		}
		// Top cap indices
		for j := 0; j < segments; j++ {
			v0 := uint32(j + segments)
			v1 := uint32((j+1)%segments + segments)
			indices[iIndex] = uint32(topCenter)
			indices[iIndex+1] = v1
			indices[iIndex+2] = v0
			iIndex += 3
		}
	}
	return verts, indices
}

func meshCone(height, radius float32, segments int, capped bool) ([]Vertex, []uint32) {
	if segments < 3 {
		segments = 3
	}
	numVerts := segments + 1   // Base ring + apex
	numIndices := segments * 3 // Side triangles
	if capped {
		numIndices += segments * 3 // Base cap
		numVerts++                 // Base center
	}
	verts := make([]Vertex, numVerts)
	indices := make([]uint32, numIndices)
	vIndex := 0
	iIndex := 0
	// Apex
	apex := vIndex
	verts[vIndex].Position = matrix.NewVec3(0.0, height, 0.0)
	verts[vIndex].Normal = matrix.Vec3{0.0, 1.0, 0.0} // Will compute better normals later if needed
	verts[vIndex].UV0 = matrix.Vec2{0.5, 0.5}
	verts[vIndex].Color = matrix.ColorWhite()
	vIndex++
	// Base ring
	for j := 0; j < segments; j++ {
		phi := float32(j) * 2.0 * math.Pi / float32(segments)
		cosPhi := matrix.Cos(phi)
		sinPhi := matrix.Sin(phi)
		verts[vIndex].Position = matrix.Vec3{radius * cosPhi, 0.0, radius * sinPhi}
		// Normal for side: slant
		slantLength := matrix.Sqrt(radius*radius + height*height)
		normalXz := radius / slantLength
		normalY := height / slantLength
		verts[vIndex].Normal = matrix.Vec3{normalXz * cosPhi, normalY, normalXz * sinPhi}
		verts[vIndex].UV0 = matrix.NewVec2(float32(j)/float32(segments), 0.0)
		verts[vIndex].Color = matrix.ColorWhite()
		vIndex++
	}
	// Side indices
	for j := 0; j < segments; j++ {
		v0 := uint32(apex)
		v1 := uint32(1 + j)
		v2 := uint32(1 + (j+1)%segments)
		indices[iIndex] = v0
		indices[iIndex+1] = v2
		indices[iIndex+2] = v1
		iIndex += 3
	}
	if capped {
		// Base center
		baseCenter := vIndex
		verts[vIndex].Position = matrix.Vec3{0.0, 0.0, 0.0}
		verts[vIndex].Normal = matrix.Vec3{0.0, -1.0, 0.0}
		verts[vIndex].UV0 = matrix.Vec2{0.5, 0.5}
		verts[vIndex].Color = matrix.ColorWhite()
		vIndex++
		// Base cap indices (winding opposite for backface)
		for j := 0; j < segments; j++ {
			v0 := uint32(1 + j)
			v1 := uint32(1 + (j+1)%segments)
			indices[iIndex] = uint32(baseCenter)
			indices[iIndex+1] = v0
			indices[iIndex+2] = v1
			iIndex += 3
		}
	}
	return verts, indices
}
