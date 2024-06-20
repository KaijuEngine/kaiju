//go:build android

package vulkan

type UintPointerAble interface {
	DescriptorPool | Semaphore | Fence | CommandPool | Buffer | DeviceMemory | Surface | Framebuffer | ShaderModule | Pipeline | PipelineLayout | DescriptorSetLayout | Image | ImageView | Sampler | Swapchain | RenderPass
}

func TypeToUintPtr[T UintPointerAble](t T) uintptr { return uintptr(t) }
