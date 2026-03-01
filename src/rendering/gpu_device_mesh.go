/******************************************************************************/
/* gpu_device_mesh.go                                                         */
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

package rendering

import (
	"kaijuengine.com/platform/profiler/tracing"
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
