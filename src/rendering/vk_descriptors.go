package rendering

import (
	"errors"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) createDescriptorSetLayout(device vk.Device, structure DescriptorSetLayoutStructure) (vk.DescriptorSetLayout, error) {
	structureCount := len(structure.Types)
	bindings := make([]vk.DescriptorSetLayoutBinding, structureCount)
	for i := 0; i < structureCount; i++ {
		bindings[i].Binding = structure.Types[i].Binding
		bindings[i].DescriptorType = structure.Types[i].Type
		bindings[i].DescriptorCount = structure.Types[i].Count
		bindings[i].PImmutableSamplers = nil // Optional
		bindings[i].StageFlags = vk.ShaderStageFlags(structure.Types[i].Flags)
	}

	info := vk.DescriptorSetLayoutCreateInfo{}
	info.SType = vk.StructureTypeDescriptorSetLayoutCreateInfo
	info.BindingCount = uint32(structureCount)
	info.PBindings = &bindings[0]
	var layout vk.DescriptorSetLayout
	if vk.CreateDescriptorSetLayout(device, &info, nil, &layout) != vk.Success {
		return layout, errors.New("failed to create descriptor set layout")
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(layout)))
	}
	return layout, nil
}

func bufferInfo(buffer vk.Buffer, bufferSize vk.DeviceSize) vk.DescriptorBufferInfo {
	bufferInfo := vk.DescriptorBufferInfo{}
	bufferInfo.Buffer = buffer
	bufferInfo.Offset = 0
	bufferInfo.Range = bufferSize
	return bufferInfo
}

func prepareSetWriteBuffer(set vk.DescriptorSet, bufferInfos []vk.DescriptorBufferInfo, bindingIndex uint32, descriptorType vk.DescriptorType) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{}
	write.SType = vk.StructureTypeWriteDescriptorSet
	write.DstSet = set
	write.DstBinding = bindingIndex
	write.DstArrayElement = 0
	write.DescriptorType = descriptorType
	write.DescriptorCount = uint32(len(bufferInfos))
	write.PBufferInfo = &bufferInfos[0]
	return write
}

func imageInfo(view vk.ImageView, sampler vk.Sampler) vk.DescriptorImageInfo {
	imageInfo := vk.DescriptorImageInfo{}
	imageInfo.ImageLayout = vk.ImageLayoutShaderReadOnlyOptimal
	imageInfo.ImageView = view
	imageInfo.Sampler = sampler
	return imageInfo
}

func prepareSetWriteImage(set vk.DescriptorSet, imageInfos []vk.DescriptorImageInfo, bindingIndex uint32, asAttachment bool) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{}
	write.SType = vk.StructureTypeWriteDescriptorSet
	write.DstSet = set
	write.DstBinding = bindingIndex
	write.DstArrayElement = 0
	if asAttachment {
		write.DescriptorType = vk.DescriptorTypeInputAttachment
	} else {
		write.DescriptorType = vk.DescriptorTypeCombinedImageSampler
	}
	write.DescriptorCount = uint32(len(imageInfos))
	write.PImageInfo = &imageInfos[0]
	return write
}
