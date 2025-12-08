//go:build !386 && !arm
// +build !386,!arm

/******************************************************************************/
/* vk_null64.go                                                               */
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

package vulkan

import "unsafe"

var (
	// NullHandle defines a platform-specfic NULL handle.
	NullHandle         unsafe.Pointer
	NullInstance       Instance
	NullPhysicalDevice PhysicalDevice
	NullDevice         Device
	NullCommandBuffer  CommandBuffer
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
