//go:build !js && !OPENGL

package rendering

import vk "github.com/BrentFarris/go-vulkan"

type InstanceDriverData struct {
	descriptorSets        [maxFramesInFlight]vk.DescriptorSet
	instanceBuffers       [maxFramesInFlight]vk.Buffer
	instanceBuffersMemory [maxFramesInFlight]vk.DeviceMemory
	lastInstanceCount     int
}

func (d *DrawInstanceGroup) generateInstanceDriverData(renderer Renderer, shader *Shader) {
	vr := renderer.(*Vulkan)
	d.InstanceDriverData.descriptorSets, _ = vr.createDescriptorSet(
		shader.RenderId.descriptorSetLayout)
}

func (d *DrawInstanceGroup) bindInstanceDriverData() {

}
