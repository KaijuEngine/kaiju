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
