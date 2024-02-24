/******************************************************************************/
/* vk_api_mesh.go                                                             */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaiju/matrix"
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) MeshIsReady(mesh Mesh) bool {
	return mesh.MeshId.vertexBuffer != vk.Buffer(vk.NullHandle)
}

func (vr *Vulkan) CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32) {
	id := &mesh.MeshId
	vNum := uint32(len(verts))
	iNum := uint32(len(indices))
	id.indexCount = iNum
	id.vertexCount = vNum
	vr.createVertexBuffer(verts, &id.vertexBuffer, &id.vertexBufferMemory)
	vr.createIndexBuffer(indices, &id.indexBuffer, &id.indexBufferMemory)
}

func (vr *Vulkan) CreateFrameBuffer(renderPass *RenderPass, attachments []vk.ImageView, width, height uint32) (vk.Framebuffer, bool) {
	framebufferInfo := vk.FramebufferCreateInfo{}
	framebufferInfo.SType = vk.StructureTypeFramebufferCreateInfo
	framebufferInfo.RenderPass = renderPass.Handle
	framebufferInfo.AttachmentCount = uint32(len(attachments))
	framebufferInfo.PAttachments = &attachments[0]
	framebufferInfo.Width = width
	framebufferInfo.Height = height
	framebufferInfo.Layers = 1
	var fb vk.Framebuffer
	if vk.CreateFramebuffer(vr.device, &framebufferInfo, nil, &fb) != vk.Success {
		slog.Error("Failed to create framebuffer")
		return nil, false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(fb)))
	}
	return fb, true
}

func (vr *Vulkan) TextureReadPixel(texture *Texture, x, y int) matrix.Color {
	panic("not implemented")
}

func (vr *Vulkan) Resize(width, height int) {
	vr.remakeSwapChain()
}

func (vr *Vulkan) AddPreRun(preRun func()) {
	vr.preRuns = append(vr.preRuns, preRun)
}

func (vr *Vulkan) DestroyGroup(group *DrawInstanceGroup) {
	vk.DeviceWaitIdle(vr.device)
	pd := bufferTrash{delay: maxFramesInFlight}
	pd.pool = group.descriptorPool
	for i := 0; i < maxFramesInFlight; i++ {
		pd.buffers[i] = group.instanceBuffers[i]
		pd.memories[i] = group.instanceBuffersMemory[i]
		pd.sets[i] = group.descriptorSets[i]
	}
	vr.bufferTrash.Add(pd)
}

func (vr *Vulkan) DestroyShader(shader *Shader) {
	vk.DeviceWaitIdle(vr.device)
	vk.DestroyPipeline(vr.device, shader.RenderId.graphicsPipeline, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.graphicsPipeline)))
	vk.DestroyPipelineLayout(vr.device, shader.RenderId.pipelineLayout, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.pipelineLayout)))
	vk.DestroyShaderModule(vr.device, shader.RenderId.vertModule, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.vertModule)))
	vk.DestroyShaderModule(vr.device, shader.RenderId.fragModule, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.fragModule)))
	if shader.RenderId.geomModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.geomModule, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.geomModule)))
	}
	if shader.RenderId.tescModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.tescModule, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.tescModule)))
	}
	if shader.RenderId.teseModule != vk.ShaderModule(vk.NullHandle) {
		vk.DestroyShaderModule(vr.device, shader.RenderId.teseModule, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.teseModule)))
	}
	vk.DestroyDescriptorSetLayout(vr.device, shader.RenderId.descriptorSetLayout, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(shader.RenderId.descriptorSetLayout)))
	for _, ss := range shader.subShaders {
		vr.DestroyShader(ss)
	}
}

func (vr *Vulkan) DestroyMesh(mesh *Mesh) {
	vk.DeviceWaitIdle(vr.device)
	id := &mesh.MeshId
	vk.DestroyBuffer(vr.device, id.indexBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.indexBuffer)))
	vk.FreeMemory(vr.device, id.indexBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.indexBufferMemory)))
	vk.DestroyBuffer(vr.device, id.vertexBuffer, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.vertexBuffer)))
	vk.FreeMemory(vr.device, id.vertexBufferMemory, nil)
	vr.dbg.remove(uintptr(unsafe.Pointer(id.vertexBufferMemory)))
	mesh.MeshId = MeshId{}
}
