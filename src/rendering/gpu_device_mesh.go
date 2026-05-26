/******************************************************************************/
/* gpu_device_mesh.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"runtime"
	"weak"

	"kaijuengine.com/platform/profiler/tracing"
)

type MeshCleanup struct {
	id     MeshId
	device weak.Pointer[GPUDevice]
}

func (g *GPUDevice) CreateVertexBuffer(verts []Vertex) (GPUBuffer, GPUDeviceMemory, error) {
	defer tracing.NewRegion("GPUDevice.CreateVertexBuffer").End()
	return g.createVertexBufferImpl(verts)
}

func (g *GPUDevice) CreateIndexBuffer(indices []uint32) (GPUBuffer, GPUDeviceMemory, error) {
	defer tracing.NewRegion("GPUDevice.CreateIndexBuffer").End()
	return g.createIndexBufferImpl(indices)
}

func (g *GPUDevice) MeshIsReady(mesh Mesh) bool {
	return mesh.MeshId.vertexBuffer.IsValid()
}

// UpdateMeshVertices re-uploads vertex data to an existing mesh's GPU
// buffer via a staging buffer copy. The vertex count must match the
// original; use CreateMesh for topology changes.
func (g *GPUDevice) UpdateMeshVertices(mesh *Mesh, verts []Vertex) {
	defer tracing.NewRegion("GPUDevice.UpdateMeshVertices").End()
	if uint32(len(verts)) != mesh.MeshId.vertexCount {
		return
	}
	g.updateVertexBufferImpl(mesh.MeshId.vertexBuffer, verts)
}

// CreateDynamicMesh creates a mesh with a HOST_VISIBLE vertex buffer that
// can be updated directly via MapMemory without staging buffers or GPU
// synchronization. Use for meshes that change frequently (e.g. terrain
// during brush edits). The index buffer is still DEVICE_LOCAL.
func (g *GPUDevice) CreateDynamicMesh(mesh *Mesh, verts []Vertex, indices []uint32) {
	defer tracing.NewRegion("GPUDevice.CreateDynamicMesh").End()
	id := &mesh.MeshId
	id.vertexCount = uint32(len(verts))
	id.indexCount = uint32(len(indices))
	id.vertexBuffer, id.vertexBufferMemory, _ = g.createDynamicVertexBufferImpl(verts)
	id.indexBuffer, id.indexBufferMemory, _ = g.CreateIndexBuffer(indices)
	runtime.AddCleanup(mesh, func(state MeshCleanup) {
		d := state.device.Value()
		if d == nil {
			return
		}
		d.Painter.preRuns = append(d.Painter.preRuns, func() {
			d.destroyMeshHandle(state.id)
		})
	}, MeshCleanup{mesh.MeshId, weak.Make(g)})
}

// UpdateDynamicMeshVertices writes vertex data directly to a HOST_VISIBLE
// vertex buffer. No staging buffer, no command buffer, no fence wait.
func (g *GPUDevice) UpdateDynamicMeshVertices(mesh *Mesh, verts []Vertex) {
	defer tracing.NewRegion("GPUDevice.UpdateDynamicMeshVertices").End()
	if uint32(len(verts)) != mesh.MeshId.vertexCount {
		return
	}
	g.updateDynamicVertexBufferImpl(mesh.MeshId.vertexBufferMemory, verts)
}

func (g *GPUDevice) CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32) {
	defer tracing.NewRegion("GPUDevice.CreateMesh").End()
	id := &mesh.MeshId
	vNum := uint32(len(verts))
	iNum := uint32(len(indices))
	id.indexCount = iNum
	id.vertexCount = vNum
	id.vertexBuffer, id.vertexBufferMemory, _ = g.CreateVertexBuffer(verts) // TODO:  Don't discard
	id.indexBuffer, id.indexBufferMemory, _ = g.CreateIndexBuffer(indices)  // TODO:  Don't discard
	runtime.AddCleanup(mesh, func(state MeshCleanup) {
		d := state.device.Value()
		if d == nil {
			return
		}
		d.Painter.preRuns = append(d.Painter.preRuns, func() {
			d.destroyMeshHandle(state.id)
		})
	}, MeshCleanup{mesh.MeshId, weak.Make(g)})
}

func (g *GPUDevice) destroyMeshHandle(handle MeshId) MeshId {
	defer tracing.NewRegion("GPUDevice.DestroyMesh").End()
	g.LogicalDevice.WaitIdle()
	g.DestroyBuffer(handle.indexBuffer)
	g.LogicalDevice.dbg.remove(handle.indexBuffer.handle)
	g.FreeMemory(handle.indexBufferMemory)
	g.LogicalDevice.dbg.remove(handle.indexBufferMemory.handle)
	g.DestroyBuffer(handle.vertexBuffer)
	g.LogicalDevice.dbg.remove(handle.vertexBuffer.handle)
	g.FreeMemory(handle.vertexBufferMemory)
	g.LogicalDevice.dbg.remove(handle.vertexBufferMemory.handle)
	return MeshId{}
}
