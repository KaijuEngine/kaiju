/******************************************************************************/
/* vk_config.go                                                               */
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
	"unsafe"

	"kaiju/build"
	vk "kaiju/rendering/vulkan"
	"kaiju/rendering/vulkan_const"
)

const (
	useValidationLayers  = build.Debug
	BytesInPixel         = 4
	MaxCommandPools      = 5
	MaxSecondaryCommands = 25
	maxFramesInFlight    = 10
	oitSuffix            = ".oit.spv"
)

func validationLayers() []string {
	if useValidationLayers {
		return []string{"VK_LAYER_KHRONOS_validation\x00"}
	} else {
		return []string{}
	}
}

func requiredDeviceExtensions() []string {
	return append([]string{vulkan_const.KhrSwapchainExtensionName + "\x00"}, vkDeviceExtensions()...)
}

func vertexGetBindingDescription(shader *Shader) []vk.VertexInputBindingDescription {
	desc := []vk.VertexInputBindingDescription{
		{
			Binding:   0,
			Stride:    uint32(unsafe.Sizeof(*(*Vertex)(nil))),
			InputRate: vulkan_const.VertexInputRateVertex,
		},
	}
	if shader.DriverData.Stride > 0 {
		desc = append(desc, vk.VertexInputBindingDescription{
			Binding:   1,
			Stride:    shader.DriverData.Stride,
			InputRate: vulkan_const.VertexInputRateInstance,
		})
	}
	return desc
}

func vertexGetAttributeDescription(shader *Shader) []vk.VertexInputAttributeDescription {
	var desc [8]vk.VertexInputAttributeDescription
	desc[0].Binding = 0
	desc[0].Location = 0
	desc[0].Format = vulkan_const.FormatR32g32b32Sfloat
	desc[0].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Position))
	desc[1].Binding = 0
	desc[1].Location = 1
	desc[1].Format = vulkan_const.FormatR32g32b32Sfloat
	desc[1].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Normal))
	desc[2].Binding = 0
	desc[2].Location = 2
	desc[2].Format = vulkan_const.FormatR32g32b32a32Sfloat
	desc[2].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Tangent))
	desc[3].Binding = 0
	desc[3].Location = 3
	desc[3].Format = vulkan_const.FormatR32g32Sfloat
	desc[3].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).UV0))
	desc[4].Binding = 0
	desc[4].Location = 4
	desc[4].Format = vulkan_const.FormatR32g32b32a32Sfloat
	desc[4].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).Color))
	desc[5].Binding = 0
	desc[5].Location = 5
	desc[5].Format = vulkan_const.FormatR32g32b32a32Sint
	desc[5].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).JointIds))
	desc[6].Binding = 0
	desc[6].Location = 6
	desc[6].Format = vulkan_const.FormatR32g32b32a32Sfloat
	desc[6].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).JointWeights))
	desc[7].Binding = 0
	desc[7].Location = 7
	desc[7].Format = vulkan_const.FormatR32g32b32Sfloat
	desc[7].Offset = uint32(unsafe.Offsetof((*Vertex)(nil).MorphTarget))
	uniformDescriptions := shader.DriverData.AttributeDescriptions
	descriptions := make([]vk.VertexInputAttributeDescription, 0, len(uniformDescriptions)+len(desc))
	descriptions = append(descriptions, desc[:]...)
	descriptions = append(descriptions, uniformDescriptions...)
	return descriptions
}
