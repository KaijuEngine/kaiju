/******************************************************************************/
/* vk_validation_layers.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"

	"kaijuengine.com/klib"

	vk "kaijuengine.com/rendering/vulkan"
)

func checkValidationLayerSupport(validationLayers []string) bool {
	var layerCount uint32
	vk.EnumerateInstanceLayerProperties(&layerCount, nil)
	if layerCount == 0 {
		return false
	}
	availableLayers := make([]vk.LayerProperties, layerCount)
	vk.EnumerateInstanceLayerProperties(&layerCount, &availableLayers[0])
	available := true
	for i := uint64(0); i < uint64(len(validationLayers)) && available; i++ {
		layerFound := false
		layerName := validationLayers[i]
		for j := uint32(0); j < layerCount; j++ {
			layer := &availableLayers[j]
			end := klib.FindFirstZeroInByteArray(layer.LayerName[:])
			if layerName == string(layer.LayerName[:end+1]) {
				layerFound = true
				break
			}
		}
		if !layerFound {
			available = false
			slog.Error("Could not find validation layer", slog.String("layer", layerName))
		}
	}
	return available
}
