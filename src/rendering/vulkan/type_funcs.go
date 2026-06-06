/******************************************************************************/
/* type_funcs.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package vulkan

import "unsafe"

type UintPointerAble interface {
	DescriptorPool | Semaphore | Fence | CommandPool | Buffer | DeviceMemory | Surface | Framebuffer | ShaderModule | Pipeline | PipelineLayout | DescriptorSetLayout | Image | ImageView | Sampler | Swapchain | RenderPass | CommandBuffer
}

func TypeToUintPtr[T UintPointerAble](t T) uintptr { return uintptr(unsafe.Pointer(t)) }
