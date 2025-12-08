//go:build 386 || arm
// +build 386 arm

/******************************************************************************/
/* vk_null32.go                                                               */
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

package vulkan

var (
	// NullHandle defines a platform-specfic NULL handle.
	NullHandle         = 0
	NullInstance       Instance
	NullPhysicalDevice PhysicalDevice
	NullDevice         Device
	NullCommandBuffer  CommandBuffer
	// NullSemaphore defines a platform-specfic NULL Semaphore.
	NullSemaphore Semaphore = 0
	// NullFence defines a platform-specfic NULL Fence.
	NullFence Fence = 0
	// NullDeviceMemory defines a platform-specfic NULL DeviceMemory.
	NullDeviceMemory DeviceMemory = 0
	// NullBuffer defines a platform-specfic NULL Buffer.
	NullBuffer Buffer = 0
	// NullImage defines a platform-specfic NULL Image.
	NullImage Image = 0
	// NullEvent defines a platform-specfic NULL Event.
	NullEvent Event = 0
	// NullQueryPool defines a platform-specfic NULL QueryPool.
	NullQueryPool QueryPool = 0
	// NullBufferView defines a platform-specfic NULL BufferView.
	NullBufferView BufferView = 0
	// NullImageView defines a platform-specfic NULL ImageView.
	NullImageView ImageView = 0
	// NullShaderModule defines a platform-specfic NULL ShaderModule.
	NullShaderModule ShaderModule = 0
	// NullPipelineCache defines a platform-specfic NULL PipelineCache.
	NullPipelineCache PipelineCache = 0
	// NullPipelineLayout defines a platform-specfic NULL PipelineLayout.
	NullPipelineLayout PipelineLayout = 0
	// NullRenderPass defines a platform-specfic NULL RenderPass.
	NullRenderPass RenderPass = 0
	// NullPipeline defines a platform-specfic NULL Pipeline.
	NullPipeline Pipeline = 0
	// NullDescriptorSetLayout defines a platform-specfic NULL DescriptorSetLayout.
	NullDescriptorSetLayout DescriptorSetLayout = 0
	// NullSampler defines a platform-specfic NULL Sampler.
	NullSampler Sampler = 0
	// NullDescriptorPool defines a platform-specfic NULL DescriptorPool.
	NullDescriptorPool DescriptorPool = 0
	// NullDescriptorSet defines a platform-specfic NULL DescriptorSet.
	NullDescriptorSet DescriptorSet = 0
	// NullFramebuffer defines a platform-specfic NULL Framebuffer.
	NullFramebuffer Framebuffer = 0
	// NullCommandPool defines a platform-specfic NULL CommandPool.
	NullCommandPool CommandPool = 0
	// NullSurface defines a platform-specfic NULL Surface.
	NullSurface Surface = 0
	// NullSwapchain defines a platform-specfic NULL Swapchain.
	NullSwapchain Swapchain = 0
	// NullDisplay defines a platform-specfic NULL Display.
	NullDisplay Display = 0
	// NullDisplayMode defines a platform-specfic NULL DisplayMode.
	NullDisplayMode DisplayMode = 0
	// NullDebugReportCallback defines a platform-specfic NULL DebugReportCallback.
	NullDebugReportCallback DebugReportCallback = 0
)
