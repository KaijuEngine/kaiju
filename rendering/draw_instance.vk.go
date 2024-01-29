//go:build !js && !OPENGL

package rendering

import vk "github.com/BrentFarris/go-vulkan"

type InstanceDriverData struct {
	descriptorSets        [maxFramesInFlight]vk.DescriptorSet
	instanceBuffers       [maxFramesInFlight]vk.Buffer
	instanceBuffersMemory [maxFramesInFlight]vk.DeviceMemory
	lastInstanceCount     int
	generatedSets         bool
}

func (d *DrawInstanceGroup) generateInstanceDriverData(renderer Renderer, shader *Shader) {
	if !d.generatedSets {
		vr := renderer.(*Vulkan)
		d.InstanceDriverData.descriptorSets, _ = vr.createDescriptorSet(
			shader.RenderId.descriptorSetLayout)
		d.generatedSets = true
	}
}

func (d *DrawInstanceGroup) bindInstanceDriverData() {

}
