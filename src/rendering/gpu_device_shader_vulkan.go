/******************************************************************************/
/* gpu_device_shader_vulkan.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"log/slog"
	"runtime"
	"unsafe"
	"weak"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering/gpu_types"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

type ShaderCleanup struct {
	id     ShaderId
	device weak.Pointer[GPUDevice]
}

type FuncPipeline func(device *GPUDevice, shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo) bool

func (g *GPUDevice) CreateShader(shader *Shader, assetDB assets.Database) error {
	defer tracing.NewRegion("GPUDevice.CreateShader").End()
	switch shader.Type {
	case ShaderTypeGraphics:
		return g.createGraphicsShader(shader, assetDB)
	case ShaderTypeCompute:
		return g.createComputeShader(shader, assetDB)
	}
	return errors.New("unhandled shader type")
}

func (g *GPUDevice) createGraphicsShader(shader *Shader, assetDB assets.Database) error {
	defer tracing.NewRegion("GPUDevice.createGraphicsShader").End()
	var vert, frag, geom, tesc, tese vk.ShaderModule
	var vMem, fMem, gMem, cMem, eMem []byte
	vertStage := vk.PipelineShaderStageCreateInfo{}
	vMem, err := assetDB.Read(shader.data.Vertex)
	if err != nil {
		slog.Error("Failed to load vertex shader", "module", shader.data.Vertex, "error", err)
		return err
	}
	vert, ok := g.createSpvModule(vMem)
	if !ok {
		slog.Error("Failed to load vertex module", "module", shader.data.Vertex)
		return err
	}
	vertStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
	vertStage.Stage = vulkan_const.ShaderStageVertexBit
	vertStage.Module = vert
	vertStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.vertModule.Handle = unsafe.Pointer(vert)

	fragStage := vk.PipelineShaderStageCreateInfo{}
	fMem, err = assetDB.Read(shader.data.Fragment)
	if err != nil {
		slog.Error("Failed to load fragment shader", "module", shader.data.Fragment, "error", err)
		return err
	}
	frag, ok = g.createSpvModule(fMem)
	if !ok {
		slog.Error("Failed to load fragment module", "module", shader.data.Fragment)
		return err
	}
	fragStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
	fragStage.Stage = vulkan_const.ShaderStageFragmentBit
	fragStage.Module = frag
	fragStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.fragModule.Handle = unsafe.Pointer(frag)

	geomStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.Geometry) > 0 {
		gMem, err = assetDB.Read(shader.data.Geometry)
		if err != nil {
			slog.Error("Failed to load geometry shader", "module", shader.data.Geometry, "error", err)
			return err
		}
		geom, ok = g.createSpvModule(gMem)
		if !ok {
			slog.Error("Failed to load geometry module", "module", shader.data.Geometry)
			return err
		}
		geomStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
		geomStage.Stage = vulkan_const.ShaderStageGeometryBit
		geomStage.Module = geom
		geomStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	}

	tescStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.TessellationControl) > 0 {
		cMem, err = assetDB.Read(shader.data.TessellationControl)
		if err != nil {
			slog.Error("Failed to load tessellation control shader", "module", shader.data.TessellationControl, "error", err)
			return err
		}
		tesc, ok = g.createSpvModule(cMem)
		if !ok {
			slog.Error("Failed to load tessellation control module", "module", shader.data.TessellationControl)
			return err
		}
		tescStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
		tescStage.Stage = vulkan_const.ShaderStageTessellationControlBit
		tescStage.Module = tesc
		tescStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.tescModule.Handle = unsafe.Pointer(tesc)
	}

	teseStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.data.TessellationEvaluation) > 0 {
		eMem, err = assetDB.Read(shader.data.TessellationEvaluation)
		if err != nil {
			slog.Error("Failed to load tessellation evaluation shader", "module", shader.data.TessellationEvaluation, "error", err)
			return err
		}
		tese, ok = g.createSpvModule(eMem)
		if !ok {
			slog.Error("Failed to load tessellation evaluation module", "module", shader.data.TessellationEvaluation)
			return err
		}
		teseStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
		teseStage.Stage = vulkan_const.ShaderStageTessellationEvaluationBit
		teseStage.Module = tese
		teseStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.teseModule.Handle = unsafe.Pointer(tese)
	}

	id := &shader.RenderId
	shader.DriverData.setup(&shader.data)
	id.descriptorSetLayout, err = g.createDescriptorSetLayout(shader.DriverData.DescriptorSetLayoutStructure)
	if err != nil {
		// TODO:  Handle this error p
		slog.Error(err.Error())
	}
	stages := make([]vk.PipelineShaderStageCreateInfo, 0)
	if vertStage.SType != 0 {
		stages = append(stages, vertStage)
	}
	if fragStage.SType != 0 {
		stages = append(stages, fragStage)
	}
	if geomStage.SType != 0 {
		stages = append(stages, geomStage)
	}
	if tescStage.SType != 0 {
		stages = append(stages, tescStage)
	}
	if teseStage.SType != 0 {
		stages = append(stages, teseStage)
	}
	shader.pipelineInfo.ConstructPipeline(g, shader, shader.renderPass.Value(), stages)
	runtime.AddCleanup(shader, func(state ShaderCleanup) {
		d := state.device.Value()
		if d == nil || !d.LogicalDevice.IsValid() {
			return
		}
		d.Painter.preRuns = append(d.Painter.preRuns, func() {
			d.DestroyShaderHandle(state.id)
		})
	}, ShaderCleanup{shader.RenderId, weak.Make(g)})
	return nil
}

func (g *GPUDevice) createComputeShader(shader *Shader, assetDB assets.Database) error {
	defer tracing.NewRegion("GPUDevice.createComputeShader").End()
	compStage := vk.PipelineShaderStageCreateInfo{}
	cMem, err := assetDB.Read(shader.data.Compute)
	if err != nil {
		return err
	}
	comp, ok := g.createSpvModule(cMem)
	if !ok {
		slog.Error("Failed to load compute module", "module", shader.data.Compute)
		return err
	}
	compStage.SType = vulkan_const.StructureTypePipelineShaderStageCreateInfo
	compStage.Stage = vulkan_const.ShaderStageComputeBit
	compStage.Module = comp
	compStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.compModule.Handle = unsafe.Pointer(comp)
	id := &shader.RenderId
	shader.DriverData.setup(&shader.data)
	vkDevice := vk.Device(g.LogicalDevice.Handle)
	id.descriptorSetLayout, err = g.createDescriptorSetLayout(shader.DriverData.DescriptorSetLayoutStructure)
	if err != nil {
		return err
	}
	pSetLayouts := vk.DescriptorSetLayout(shader.RenderId.descriptorSetLayout.Handle)
	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vulkan_const.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            &pSetLayouts,
		PushConstantRangeCount: 0, // Adjust if push constants are used
		PPushConstantRanges:    nil,
	}
	var pLayout vk.PipelineLayout
	if vk.CreatePipelineLayout(vkDevice, &pipelineLayoutInfo, nil, &pLayout) != vulkan_const.Success {
		slog.Error("Failed to create pipeline layout")
		return errors.New("failed to create pipeline layout")
	} else {
		g.LogicalDevice.dbg.track(unsafe.Pointer(pLayout))
	}
	shader.RenderId.pipelineLayout.Handle = unsafe.Pointer(pLayout)
	pipelineInfo := vk.ComputePipelineCreateInfo{
		SType:  vulkan_const.StructureTypeComputePipelineCreateInfo,
		Stage:  compStage,
		Layout: vk.PipelineLayout(id.pipelineLayout.Handle),
	}
	pipelines := [1]vk.Pipeline{}
	if vk.CreateComputePipelines(vkDevice, vk.NullPipelineCache, 1, &pipelineInfo, nil, &pipelines[0]) != vulkan_const.Success {
		slog.Error("Failed to create compute pipeline")
		return errors.New("failed to create compute pipeline")
	} else {
		g.LogicalDevice.dbg.track(unsafe.Pointer(pipelines[0]))
	}
	id.computePipeline.Handle = unsafe.Pointer(pipelines[0])
	return nil
}

func (g *GPUDevice) createSpvModule(mem []byte) (vk.ShaderModule, bool) {
	defer tracing.NewRegion("GPUDevice.createSpvModule").End()
	info := vk.ShaderModuleCreateInfo{}
	info.SType = vulkan_const.StructureTypeShaderModuleCreateInfo
	info.CodeSize = uint(len(mem))
	info.PCode = (*uint32)(unsafe.Pointer(&mem[0]))
	var outModule vk.ShaderModule
	vkDevice := vk.Device(g.LogicalDevice.Handle)
	if vk.CreateShaderModule(vkDevice, &info, nil, &outModule) != vulkan_const.Success {
		slog.Error("Failed to create shader module", slog.String("module", string(mem)))
		return outModule, false
	} else {
		g.LogicalDevice.dbg.track(unsafe.Pointer(outModule))
		return outModule, true
	}
}

func (g *GPUDevice) destroyShaderHandleImpl(id ShaderId) {
	defer tracing.NewRegion("GPUDevice.destroyShaderHandleImpl").End()
	vkDevice := vk.Device(g.LogicalDevice.Handle)
	g.LogicalDevice.WaitIdle()
	vk.DestroyPipeline(vkDevice, vk.Pipeline(id.graphicsPipeline.Handle), nil)
	g.LogicalDevice.dbg.remove(id.graphicsPipeline.Handle)
	vk.DestroyPipeline(vkDevice, vk.Pipeline(id.computePipeline.Handle), nil)
	g.LogicalDevice.dbg.remove(id.computePipeline.Handle)
	vk.DestroyPipelineLayout(vkDevice, vk.PipelineLayout(id.pipelineLayout.Handle), nil)
	g.LogicalDevice.dbg.remove(id.pipelineLayout.Handle)
	vk.DestroyShaderModule(vkDevice, vk.ShaderModule(id.vertModule.Handle), nil)
	g.LogicalDevice.dbg.remove(id.vertModule.Handle)
	vk.DestroyShaderModule(vkDevice, vk.ShaderModule(id.fragModule.Handle), nil)
	g.LogicalDevice.dbg.remove(id.fragModule.Handle)
	if id.geomModule.IsValid() {
		vk.DestroyShaderModule(vkDevice, vk.ShaderModule(id.geomModule.Handle), nil)
		g.LogicalDevice.dbg.remove(id.geomModule.Handle)
	}
	if id.tescModule.IsValid() {
		vk.DestroyShaderModule(vkDevice, vk.ShaderModule(id.tescModule.Handle), nil)
		g.LogicalDevice.dbg.remove(id.tescModule.Handle)
	}
	if id.teseModule.IsValid() {
		vk.DestroyShaderModule(vkDevice, vk.ShaderModule(id.teseModule.Handle), nil)
		g.LogicalDevice.dbg.remove(id.teseModule.Handle)
	}
	if id.compModule.IsValid() {
		vk.DestroyShaderModule(vkDevice, vk.ShaderModule(id.compModule.Handle), nil)
		g.LogicalDevice.dbg.remove(id.compModule.Handle)
	}
	vk.DestroyDescriptorSetLayout(vkDevice, vk.DescriptorSetLayout(id.descriptorSetLayout.Handle), nil)
	g.LogicalDevice.dbg.remove(id.descriptorSetLayout.Handle)
}

func (g *GPUDevice) createDescriptorSetLayout(structure DescriptorSetLayoutStructure) (gpu_types.DescriptorSetLayout, error) {
	defer tracing.NewRegion("GPUDevice.createDescriptorSetLayout").End()
	structureCount := len(structure.Types)
	bindings := make([]vk.DescriptorSetLayoutBinding, structureCount)
	for i := 0; i < structureCount; i++ {
		bindings[i].Binding = structure.Types[i].Binding
		bindings[i].DescriptorType = structure.Types[i].Type
		bindings[i].DescriptorCount = structure.Types[i].Count
		bindings[i].PImmutableSamplers = nil // Optional
		bindings[i].StageFlags = vk.ShaderStageFlags(structure.Types[i].Flags)
	}

	info := vk.DescriptorSetLayoutCreateInfo{
		SType:        vulkan_const.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: uint32(structureCount),
	}
	if structureCount > 0 {
		info.PBindings = &bindings[0]
	}
	var layout vk.DescriptorSetLayout
	if vk.CreateDescriptorSetLayout(vk.Device(g.LogicalDevice.Handle), &info, nil, &layout) != vulkan_const.Success {
		return gpu_types.DescriptorSetLayout{}, errors.New("failed to create descriptor set layout")
	}
	g.LogicalDevice.dbg.track(unsafe.Pointer(layout))
	return gpu_types.DescriptorSetLayout{gpu_types.GpuHandle{unsafe.Pointer(layout)}}, nil
}

func bufferInfo(buffer vk.Buffer, bufferSize vk.DeviceSize) vk.DescriptorBufferInfo {
	bufferInfo := vk.DescriptorBufferInfo{}
	bufferInfo.Buffer = buffer
	bufferInfo.Offset = 0
	bufferInfo.Range = bufferSize
	return bufferInfo
}

func prepareSetWriteBuffer(set vk.DescriptorSet, bufferInfos []vk.DescriptorBufferInfo, bindingIndex uint32, descriptorType vulkan_const.DescriptorType) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{}
	write.SType = vulkan_const.StructureTypeWriteDescriptorSet
	write.DstSet = set
	write.DstBinding = bindingIndex
	write.DstArrayElement = 0
	write.DescriptorType = descriptorType
	write.DescriptorCount = uint32(len(bufferInfos))
	write.PBufferInfo = &bufferInfos[0]
	return write
}

func imageInfo(view vk.ImageView, sampler vk.Sampler) gpu_types.DescriptorImageInfo {
	imageInfo := gpu_types.DescriptorImageInfo{}
	imageInfo.ImageLayout.FromVulkan(vulkan_const.ImageLayoutShaderReadOnlyOptimal)
	imageInfo.ImageView.Handle = unsafe.Pointer(view)
	imageInfo.Sampler.Handle = unsafe.Pointer(sampler)
	return imageInfo
}

func imageInfoVk(view vk.ImageView, sampler vk.Sampler) vk.DescriptorImageInfo {
	return vk.DescriptorImageInfo{
		ImageLayout: vulkan_const.ImageLayoutShaderReadOnlyOptimal,
		ImageView:   view,
		Sampler:     sampler,
	}
}

func prepareSetWriteImage(set vk.DescriptorSet, imageInfos []vk.DescriptorImageInfo, bindingIndex uint32, asAttachment bool) vk.WriteDescriptorSet {
	write := vk.WriteDescriptorSet{
		SType:           vulkan_const.StructureTypeWriteDescriptorSet,
		DstSet:          set,
		DstBinding:      bindingIndex,
		DstArrayElement: 0,
	}
	if asAttachment {
		write.DescriptorType = vulkan_const.DescriptorTypeInputAttachment
	} else {
		write.DescriptorType = vulkan_const.DescriptorTypeCombinedImageSampler
	}
	write.DescriptorCount = uint32(len(imageInfos))
	write.PImageInfo = &imageInfos[0]
	return write
}
