/*****************************************************************************/
/* vk_instance.go                                                            */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package rendering

import (
	"log/slog"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) createSwapChainFrameBuffer() bool {
	count := vr.swapChainImageViewCount
	vr.swapChainFrameBufferCount = count
	vr.swapChainFrameBuffers = make([]vk.Framebuffer, count)
	success := true
	for i := uint32(0); i < count && success; i++ {
		attachments := []vk.ImageView{
			vr.color.View,
			vr.depth.View,
			vr.swapImages[i].View,
		}
		vr.swapChainFrameBuffers[i], success = vr.CreateFrameBuffer(vr.swapChainRenderPass, attachments,
			vr.swapChainExtent.Width, vr.swapChainExtent.Height)
	}
	return success
}

func (vr *Vulkan) createVulkanInstance(appInfo vk.ApplicationInfo) bool {
	windowExtensions := vr.window.GetInstanceExtensions()
	added := make([]string, 0, 3)
	if useValidationLayers {
		added = append(added, vk.ExtDebugReportExtensionName+"\x00")
	}
	//	const char* added[] = {
	//#ifdef ANDROID
	//		VK_KHR_SURFACE_EXTENSION_NAME,
	//		VK_KHR_ANDROID_SURFACE_EXTENSION_NAME,
	//#elif defined(USE_VALIDATION_LAYERS)
	//		VK_EXT_DEBUG_REPORT_EXTENSION_NAME,
	//#endif
	//	};
	extensions := make([]string, 0, len(windowExtensions)+len(added))
	extensions = append(extensions, windowExtensions...)
	extensions = append(extensions, added...)
	extensions = append(extensions, vkInstanceExtensions()...)

	createInfo := vk.InstanceCreateInfo{
		SType:            vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &appInfo,
		Flags:            vkInstanceFlags,
	}
	defer createInfo.Free()
	createInfo.SetEnabledExtensionNames(extensions)

	validationLayers := validationLayers()
	if len(validationLayers) > 0 {
		if !checkValidationLayerSupport(validationLayers) {
			slog.Error("Expected to have validation layers for debugging, but didn't find them")
			return false
		}
		createInfo.SetEnabledLayerNames(validationLayers)
	}

	var instance vk.Instance
	result := vk.CreateInstance(&createInfo, nil, &instance)
	if result != vk.Success {
		slog.Error("Failed to get the VK instance", slog.Int("code", int(result)))
		return false
	} else {
		vr.instance = instance
		vk.InitInstance(vr.instance)
		return true
	}
}
