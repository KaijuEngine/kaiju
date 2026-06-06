/******************************************************************************/
/* shader_layout_vulkan.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "kaijuengine.com/rendering/vulkan_const"

var (
	defTypes = map[string]shaderFieldType{
		"float":  {uint32(floatSize), vulkan_const.FormatR32Sfloat, 1},
		"vec2":   {uint32(floatSize) * 2, vulkan_const.FormatR32g32Sfloat, 1},
		"vec3":   {uint32(floatSize) * 3, vulkan_const.FormatR32g32b32Sfloat, 1},
		"vec4":   {uint32(vec4Size), vulkan_const.FormatR32g32b32a32Sfloat, 1},
		"mat4":   {uint32(vec4Size), vulkan_const.FormatR32g32b32a32Sfloat, 4},
		"int":    {uint32(int32Size), vulkan_const.FormatR32Sint, 1},
		"int32":  {uint32(int32Size), vulkan_const.FormatR32Sint, 1},
		"uint":   {uint32(uint32Size), vulkan_const.FormatR32Uint, 1},
		"uint32": {uint32(uint32Size), vulkan_const.FormatR32Uint, 1},
	}
)
