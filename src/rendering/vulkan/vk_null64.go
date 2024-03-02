// +build !386,!arm

package vulkan

import "unsafe"

var (
	// NullHandle defines a platform-specfic NULL handle.
	NullHandle unsafe.Pointer
	// NullSemaphore defines a platform-specfic NULL Semaphore.
	NullSemaphore Semaphore
	// NullFence defines a platform-specfic NULL Fence.
	NullFence Fence
	// NullDeviceMemory defines a platform-specfic NULL DeviceMemory.
	NullDeviceMemory DeviceMemory
	// NullBuffer defines a platform-specfic NULL Buffer.
	NullBuffer Buffer
	// NullImage defines a platform-specfic NULL Image.
	NullImage Image
	// NullEvent defines a platform-specfic NULL Event.
	NullEvent Event
	// NullQueryPool defines a platform-specfic NULL QueryPool.
	NullQueryPool QueryPool
	// NullBufferView defines a platform-specfic NULL BufferView.
	NullBufferView BufferView
	// NullImageView defines a platform-specfic NULL ImageView.
	NullImageView ImageView
	// NullShaderModule defines a platform-specfic NULL ShaderModule.
	NullShaderModule ShaderModule
	// NullPipelineCache defines a platform-specfic NULL PipelineCache.
	NullPipelineCache PipelineCache
	// NullPipelineLayout defines a platform-specfic NULL PipelineLayout.
	NullPipelineLayout PipelineLayout
	// NullRenderPass defines a platform-specfic NULL RenderPass.
	NullRenderPass RenderPass
	// NullPipeline defines a platform-specfic NULL Pipeline.
	NullPipeline Pipeline
	// NullDescriptorSetLayout defines a platform-specfic NULL DescriptorSetLayout.
	NullDescriptorSetLayout DescriptorSetLayout
	// NullSampler defines a platform-specfic NULL Sampler.
	NullSampler Sampler
	// NullDescriptorPool defines a platform-specfic NULL DescriptorPool.
	NullDescriptorPool DescriptorPool
	// NullDescriptorSet defines a platform-specfic NULL DescriptorSet.
	NullDescriptorSet DescriptorSet
	// NullFramebuffer defines a platform-specfic NULL Framebuffer.
	NullFramebuffer Framebuffer
	// NullCommandPool defines a platform-specfic NULL CommandPool.
	NullCommandPool CommandPool
	// NullSurface defines a platform-specfic NULL Surface.
	NullSurface Surface
	// NullSwapchain defines a platform-specfic NULL Swapchain.
	NullSwapchain Swapchain
	// NullDisplay defines a platform-specfic NULL Display.
	NullDisplay Display
	// NullDisplayMode defines a platform-specfic NULL DisplayMode.
	NullDisplayMode DisplayMode
	// NullDebugReportCallback defines a platform-specfic NULL DebugReportCallback.
	NullDebugReportCallback DebugReportCallback
)
