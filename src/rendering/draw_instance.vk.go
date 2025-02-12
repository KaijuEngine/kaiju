/******************************************************************************/
/* draw_instance.vk.go                                                       */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package rendering

import vk "kaiju/rendering/vulkan"

type InstanceGroupSkinningData struct {
}

type ShaderBuffer struct {
	bindingId int
	size      vk.DeviceSize
	buffers   [maxFramesInFlight]vk.Buffer
	memories  [maxFramesInFlight]vk.DeviceMemory
	stride    int
	capacity  int
}

type InstanceDriverData struct {
	descriptorPool    vk.DescriptorPool
	descriptorSets    [maxFramesInFlight]vk.DescriptorSet
	instanceBuffer    ShaderBuffer
	imageInfos        []vk.DescriptorImageInfo
	namedBuffers      map[string]ShaderBuffer
	lastInstanceCount int
	generatedSets     bool
}

func (d *DrawInstanceGroup) generateInstanceDriverData(renderer Renderer, shader *Shader) {
	if !d.generatedSets {
		vr := renderer.(*Vulkan)
		d.descriptorSets, d.descriptorPool, _ = vr.createDescriptorSet(
			shader.RenderId.descriptorSetLayout, 0)
		d.imageInfos = make([]vk.DescriptorImageInfo, len(d.Textures))
		d.generatedSets = true
		d.instanceBuffer.bindingId = 1
		d.namedBuffers = make(map[string]ShaderBuffer)
		if shader.definition != nil {
			for i := range shader.definition.LayoutGroups {
				g := &shader.definition.LayoutGroups[i]
				for j := range g.Layouts {
					if g.Layouts[j].IsBuffer() {
						d.namedBuffers[g.Layouts[j].FullName()] = ShaderBuffer{
							bindingId: g.Layouts[j].Binding,
							stride:    g.Layouts[j].Stride(),
							capacity:  g.Layouts[j].Capacity(),
						}
					}
				}
			}
		}
	}
}

func (d *DrawInstanceGroup) bindInstanceDriverData() {
}
