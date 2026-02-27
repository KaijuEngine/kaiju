package rendering

import (
	"kaiju/platform/profiler/tracing"
	"runtime"
	"weak"
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
