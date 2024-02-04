//go:build !js && !OPENGL

package rendering

import vk "github.com/KaijuEngine/go-vulkan"

type InstanceDriverData struct {
	descriptorPool        vk.DescriptorPool
	descriptorSets        [maxFramesInFlight]vk.DescriptorSet
	instanceBuffers       [maxFramesInFlight]vk.Buffer
	instanceBuffersMemory [maxFramesInFlight]vk.DeviceMemory
	lastInstanceCount     int
	generatedSets         bool
}

func (d *InstanceDriverData) IsReady() bool {
	return d.descriptorSets[0] != nil
}

func (d *InstanceDriverData) Reset() {
	d.generatedSets = false
}

func (d *DrawInstanceGroup) generateInstanceDriverData(renderer Renderer, shader *Shader) {
	if !d.generatedSets {
		vr := renderer.(*Vulkan)
		d.descriptorSets, d.descriptorPool, _ = vr.createDescriptorSet(
			shader.RenderId.descriptorSetLayout, 0)
		d.generatedSets = true
	}
}

func (d *DrawInstanceGroup) bindInstanceDriverData() {

}
