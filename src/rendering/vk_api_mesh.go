package rendering

import (
	"kaiju/matrix"
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) CreateMesh(mesh *Mesh, verts []Vertex, indices []uint32) {
	id := &mesh.MeshId
	vNum := uint32(len(verts))
	iNum := uint32(len(indices))
	id.indexCount = iNum
	id.vertexCount = vNum
	vr.createVertexBuffer(verts, &id.vertexBuffer, &id.vertexBufferMemory)
	vr.createIndexBuffer(indices, &id.indexBuffer, &id.indexBufferMemory)
}

func (vr *Vulkan) MeshIsReady(mesh Mesh) bool {
	return mesh.MeshId.vertexBuffer != vk.Buffer(vk.NullHandle)
}

func (vr *Vulkan) CreateFrameBuffer(renderPass vk.RenderPass, attachments []vk.ImageView, width, height uint32, frameBuffer *vk.Framebuffer) bool {
	framebufferInfo := vk.FramebufferCreateInfo{}
	framebufferInfo.SType = vk.StructureTypeFramebufferCreateInfo
	framebufferInfo.RenderPass = renderPass
	framebufferInfo.AttachmentCount = uint32(len(attachments))
	framebufferInfo.PAttachments = &attachments[0]
	framebufferInfo.Width = width
	framebufferInfo.Height = height
	framebufferInfo.Layers = 1
	var fb vk.Framebuffer
	if vk.CreateFramebuffer(vr.device, &framebufferInfo, nil, &fb) != vk.Success {
		slog.Error("Failed to create framebuffer")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(fb)))
	}
	*frameBuffer = fb
	return true
}

func (vr *Vulkan) resizeUniformBuffer(shader *Shader, group *DrawInstanceGroup) {
	currentCount := len(group.Instances)
	lastCount := group.InstanceDriverData.lastInstanceCount
	if currentCount > lastCount {
		if group.instanceBuffers[0] != vk.Buffer(vk.NullHandle) {
			pd := bufferTrash{delay: maxFramesInFlight}
			for i := 0; i < maxFramesInFlight; i++ {
				pd.buffers[i] = group.instanceBuffers[i]
				pd.memories[i] = group.instanceBuffersMemory[i]
				group.instanceBuffers[i] = vk.Buffer(vk.NullHandle)
				group.instanceBuffersMemory[i] = vk.DeviceMemory(vk.NullHandle)
			}
			vr.bufferTrash.Add(pd)
		}
		if currentCount > 0 {
			group.generateInstanceDriverData(vr, shader)
			iSize := vr.padUniformBufferSize(vk.DeviceSize(shader.DriverData.Stride))
			for i := 0; i < maxFramesInFlight; i++ {
				vr.CreateBuffer(iSize*vk.DeviceSize(currentCount),
					vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit|vk.BufferUsageTransferDstBit),
					vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
					&group.instanceBuffers[i], &group.instanceBuffersMemory[i])
			}
			group.AlterPadding(int(iSize))
		}
		group.InstanceDriverData.lastInstanceCount = currentCount
	}
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

func (vr *Vulkan) DestroyTexture(texture *Texture) {
	vk.DeviceWaitIdle(vr.device)
	vr.textureIdFree(&texture.RenderId)
	texture.RenderId = TextureId{}
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
	if shader.SubShader != nil {
		vr.DestroyShader(shader.SubShader)
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

func (vr *Vulkan) Destroy() {
	vk.DeviceWaitIdle(vr.device)
	vr.bufferTrash.Purge()
	if vr.device != vk.Device(vk.NullHandle) {
		vr.defaultTarget.reset(vr)
		vr.oitPass.reset(vr)
		vr.defaultTexture = nil
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroySemaphore(vr.device, vr.imageSemaphores[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.imageSemaphores[i])))
			vk.DestroySemaphore(vr.device, vr.renderSemaphores[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderSemaphores[i])))
			vk.DestroyFence(vr.device, vr.renderFences[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderFences[i])))
		}
		vk.DestroyCommandPool(vr.device, vr.commandPool, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.commandPool)))
		for i := 0; i < maxFramesInFlight; i++ {
			vk.DestroyBuffer(vr.device, vr.globalUniformBuffers[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.globalUniformBuffers[i])))
			vk.FreeMemory(vr.device, vr.globalUniformBuffersMemory[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.globalUniformBuffersMemory[i])))
		}
		for i := range vr.descriptorPools {
			vk.DestroyDescriptorPool(vr.device, vr.descriptorPools[i], nil)
			vr.dbg.remove(uintptr(unsafe.Pointer(vr.descriptorPools[i])))
		}
		vk.DestroyRenderPass(vr.device, vr.renderPass, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.renderPass)))
		vr.swapChainCleanup()
		vk.DestroyDevice(vr.device, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.device)))
	}
	if vr.instance != vk.Instance(vk.NullHandle) {
		vk.DestroySurface(vr.instance, vr.surface, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.surface)))
		vk.DestroyInstance(vr.instance, nil)
		vr.dbg.remove(uintptr(unsafe.Pointer(vr.instance)))
	}
	vr.dbg.print()
}
