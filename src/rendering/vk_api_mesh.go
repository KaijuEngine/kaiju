/******************************************************************************/
/* vk_api_mesh.go                                                             */
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

package rendering

import (
	"runtime"

	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
	vk "github.com/KaijuEngine/kaiju/rendering/vulkan"
)

type MeshCleanup struct {
	id       MeshId
	renderer Renderer
}

func (vr *Vulkan) MeshIsReady(mesh Mesh) bool {
	return mesh.MeshId.vertexBuffer != vk.Buffer(vk.NullHandle)
}

func (vr *Vulkan) CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32) {
	defer tracing.NewRegion("Vulkan.CreateMesh").End()
	id := &mesh.MeshId
	vNum := uint32(len(verts))
	iNum := uint32(len(indices))
	id.indexCount = iNum
	id.vertexCount = vNum
	vr.createVertexBuffer(verts, &id.vertexBuffer, &id.vertexBufferMemory)
	vr.createIndexBuffer(indices, &id.indexBuffer, &id.indexBufferMemory)
	runtime.AddCleanup(mesh, func(state MeshCleanup) {
		v := state.renderer.(*Vulkan)
		v.preRuns = append(v.preRuns, func() {
			state.renderer.(*Vulkan).destroyMeshHandle(state.id)
		})
	}, MeshCleanup{mesh.MeshId, vr})
}

func (vr *Vulkan) destroyMeshHandle(handle MeshId) MeshId {
	defer tracing.NewRegion("Vulkan.DestroyMesh").End()
	vk.DeviceWaitIdle(vr.device)
	vk.DestroyBuffer(vr.device, handle.indexBuffer, nil)
	vr.dbg.remove(vk.TypeToUintPtr(handle.indexBuffer))
	vk.FreeMemory(vr.device, handle.indexBufferMemory, nil)
	vr.dbg.remove(vk.TypeToUintPtr(handle.indexBufferMemory))
	vk.DestroyBuffer(vr.device, handle.vertexBuffer, nil)
	vr.dbg.remove(vk.TypeToUintPtr(handle.vertexBuffer))
	vk.FreeMemory(vr.device, handle.vertexBufferMemory, nil)
	vr.dbg.remove(vk.TypeToUintPtr(handle.vertexBufferMemory))
	return MeshId{}
}
