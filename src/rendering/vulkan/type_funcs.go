//go:build !android

package vulkan

import "unsafe"

type UintPointerAble interface {
	DescriptorPool | Semaphore | Fence | CommandPool | Buffer | DeviceMemory | Surface | Framebuffer | ShaderModule | Pipeline | PipelineLayout | DescriptorSetLayout | Image | ImageView | Sampler | Swapchain | RenderPass
}

func TypeToUintPtr[T UintPointerAble](t T) uintptr { return uintptr(unsafe.Pointer(t)) }
