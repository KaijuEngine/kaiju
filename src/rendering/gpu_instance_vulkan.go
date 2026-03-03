/******************************************************************************/
/* gpu_instance_vulkan.go                                                     */
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

package rendering

import (
	"fmt"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
	"log/slog"
	"unsafe"
)

/*
#include <stdlib.h>
*/
import "C"

func (g *GPUInstance) setupImpl(window RenderingContainer, app *GPUApplication) error {
	extensions := gatherInstanceExtensions(window)
	appName := C.CString(app.Name)
	defer C.free(unsafe.Pointer(appName))
	appVersion := vk.MakeVersion(app.ApplicationVersion())
	engineVersion := vk.MakeVersion(app.EngineVersion())
	info := vk.ApplicationInfo{
		SType:              vulkan_const.StructureTypeApplicationInfo,
		PApplicationName:   (*vk.Char)(appName),
		ApplicationVersion: appVersion,
		PEngineName:        (*vk.Char)(unsafe.Pointer(&([]byte("Kaiju\x00"))[0])),
		EngineVersion:      engineVersion,
		ApiVersion:         vulkan_const.ApiVersion11,
	}
	createInfo := vk.InstanceCreateInfo{
		SType:            vulkan_const.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &info,
		Flags:            vkInstanceFlags,
	}
	defer createInfo.Free()
	createInfo.SetEnabledExtensionNames(extensions)
	validationLayers := validationLayers()
	if len(validationLayers) > 0 {
		if !checkValidationLayerSupport(validationLayers) {
			slog.Warn("Expected to have validation layers for debugging, but didn't find them")
		} else {
			slog.Info("enabling the validation layers")
			createInfo.SetEnabledLayerNames(validationLayers)
		}
	}
	var instance vk.Instance
	result := vk.CreateInstance(&createInfo, nil, &instance)
	if result != vulkan_const.Success {
		slog.Error("Failed to get the VK instance", slog.Int("code", int(result)))
		return fmt.Errorf("failed to get the VK instance, code: %d", int(result))
	}
	g.handle = unsafe.Pointer(instance)
	vk.InitInstance(instance)
	return nil
}

func gatherInstanceExtensions(window RenderingContainer) []string {
	defer tracing.NewRegion("rendering.gatherInstanceExtensions").End()
	windowExtensions := window.GetInstanceExtensions()
	added := make([]string, 0, 3)
	if useValidationLayers {
		added = append(added, vulkan_const.ExtDebugReportExtensionName+"\x00")
	}
	extensions := make([]string, 0, len(windowExtensions)+len(added))
	extensions = append(extensions, windowExtensions...)
	extensions = append(extensions, added...)
	extensions = append(extensions, vkInstanceExtensions()...)
	return extensions
}

func (g *GPUInstance) destroyImpl() {
	vk.DestroyInstance(vk.Instance(g.handle), nil)
}
