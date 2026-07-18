/******************************************************************************/
/* shader_pipeline_vulkan.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

func (s *ShaderPipelineDataCompiled) ConstructPipeline(device *GPUDevice, shader *Shader, renderPass *RenderPass, stages []vk.PipelineShaderStageCreateInfo) bool {
	defer tracing.NewRegion("ShaderPipelineDataCompiled.ConstructPipeline").End()
	pSetLayout := vk.DescriptorSetLayout(shader.RenderId.descriptorSetLayout.handle)
	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType:          vulkan_const.StructureTypePipelineLayoutCreateInfo,
		Flags:          0, // PipelineLayoutCreateFlags
		SetLayoutCount: 1,
		PSetLayouts:    &pSetLayout,
	}
	if s.PushConstant.Size > 0 {
		pushRanges := [1]vk.PushConstantRange{{
			StageFlags: s.PushConstant.StageFlags.toVulkan(),
			Offset:     0,
			Size:       s.PushConstant.Size,
		}}
		pipelineLayoutInfo.PushConstantRangeCount = 1
		pipelineLayoutInfo.PPushConstantRanges = &pushRanges[0]
	}
	var pLayout vk.PipelineLayout
	if vk.CreatePipelineLayout(vk.Device(device.LogicalDevice.handle), &pipelineLayoutInfo, nil, &pLayout) != vulkan_const.Success {
		slog.Error("Failed to create pipeline layout")
		return false
	} else {
		device.LogicalDevice.dbg.track(unsafe.Pointer(pLayout))
	}
	shader.RenderId.pipelineLayout.handle = unsafe.Pointer(pLayout)
	bDesc := vertexGetBindingDescription(shader)
	bDescCount := uint32(len(bDesc))
	for i := uint32(1); i < bDescCount; i++ {
		bDesc[i].Stride = uint32(device.PhysicalDevice.PadBufferSize(uintptr(bDesc[i].Stride)))
	}
	aDesc := vertexGetAttributeDescription(shader)
	vertexInputInfo := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vulkan_const.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   bDescCount,
		VertexAttributeDescriptionCount: uint32(len(aDesc)),
		PVertexBindingDescriptions:      &bDesc[0],
		PVertexAttributeDescriptions:    &aDesc[0],
	}
	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vulkan_const.StructureTypePipelineInputAssemblyStateCreateInfo,
		Flags:                  0, // PipelineInputAssemblyStateCreateFlags
		Topology:               s.InputAssembly.Topology.toVulkan(),
		PrimitiveRestartEnable: boolToVkBool(s.InputAssembly.PrimitiveRestart),
	}
	sce := device.LogicalDevice.SwapChain.Extent
	viewport := vk.Viewport{
		X:        0.0,
		Y:        0.0,
		Width:    float32(sce.Width()),
		Height:   float32(sce.Height()),
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: vk.Extent2D{
			Width:  uint32(sce.Width()),
			Height: uint32(sce.Height()),
		},
	}
	dynamicStates := []vulkan_const.DynamicState{
		vulkan_const.DynamicStateViewport,
		vulkan_const.DynamicStateScissor,
	}
	dynamicState := vk.PipelineDynamicStateCreateInfo{
		SType:             vulkan_const.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynamicStates)),
		PDynamicStates:    &dynamicStates[0],
	}
	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vulkan_const.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    &viewport,
		ScissorCount:  1,
		PScissors:     &scissor,
	}
	rasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vulkan_const.StructureTypePipelineRasterizationStateCreateInfo,
		Flags:                   0, // PipelineRasterizationStateCreateFlags
		DepthClampEnable:        boolToVkBool(s.Rasterization.DepthClampEnable),
		RasterizerDiscardEnable: boolToVkBool(s.Rasterization.DiscardEnable),
		PolygonMode:             s.Rasterization.PolygonMode.toVulkan(),
		LineWidth:               s.Rasterization.LineWidth,
		CullMode:                s.Rasterization.CullMode.toVulkan(),
		FrontFace:               s.Rasterization.FrontFace.toVulkan(),
		DepthBiasEnable:         boolToVkBool(s.Rasterization.DepthBiasEnable),
		DepthBiasConstantFactor: s.Rasterization.DepthBiasConstantFactor,
		DepthBiasClamp:          s.Rasterization.DepthBiasClamp,
		DepthBiasSlopeFactor:    s.Rasterization.DepthBiasSlopeFactor,
	}
	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vulkan_const.StructureTypePipelineMultisampleStateCreateInfo,
		Flags:                 0, // PipelineMultisampleStateCreateFlags
		SampleShadingEnable:   boolToVkBool(s.Multisample.SampleShadingEnable),
		RasterizationSamples:  vulkan_const.SampleCountFlagBits(s.Multisample.RasterizationSamples.toVulkan()),
		MinSampleShading:      s.Multisample.MinSampleShading,
		PSampleMask:           nil,
		AlphaToCoverageEnable: boolToVkBool(s.Multisample.AlphaToCoverageEnable),
		AlphaToOneEnable:      boolToVkBool(s.Multisample.AlphaToOneEnable),
	}
	compiledBlendAttachments := s.colorBlendAttachmentsForRenderPass(renderPass)
	colorBlendAttachment := make([]vk.PipelineColorBlendAttachmentState, len(compiledBlendAttachments))
	for i := range compiledBlendAttachments {
		colorBlendAttachment[i].BlendEnable = boolToVkBool(compiledBlendAttachments[i].BlendEnable)
		colorBlendAttachment[i].SrcColorBlendFactor = compiledBlendAttachments[i].SrcColorBlendFactor.toVulkan()
		colorBlendAttachment[i].DstColorBlendFactor = compiledBlendAttachments[i].DstColorBlendFactor.toVulkan()
		colorBlendAttachment[i].ColorBlendOp = compiledBlendAttachments[i].ColorBlendOp.toVulkan()
		colorBlendAttachment[i].SrcAlphaBlendFactor = compiledBlendAttachments[i].SrcAlphaBlendFactor.toVulkan()
		colorBlendAttachment[i].DstAlphaBlendFactor = compiledBlendAttachments[i].DstAlphaBlendFactor.toVulkan()
		colorBlendAttachment[i].AlphaBlendOp = compiledBlendAttachments[i].AlphaBlendOp.toVulkan()
		colorBlendAttachment[i].ColorWriteMask = compiledBlendAttachments[i].ColorWriteMask.toVulkan()
	}
	colorBlendAttachmentCount := len(colorBlendAttachment)
	colorBlending := vk.PipelineColorBlendStateCreateInfo{
		SType:           vulkan_const.StructureTypePipelineColorBlendStateCreateInfo,
		Flags:           0, // PipelineColorBlendStateCreateFlags
		LogicOpEnable:   boolToVkBool(s.ColorBlend.LogicOpEnable),
		LogicOp:         s.ColorBlend.LogicOp.toVulkan(),
		AttachmentCount: uint32(colorBlendAttachmentCount),
		BlendConstants:  s.ColorBlend.BlendConstants,
	}
	if colorBlendAttachmentCount > 0 {
		colorBlending.PAttachments = &colorBlendAttachment[0]
	}
	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vulkan_const.StructureTypeGraphicsPipelineCreateInfo,
		Flags:               s.GraphicsPipeline.PipelineCreateFlags.toVulkan(),
		StageCount:          uint32(len(stages)),
		PStages:             &stages[0],
		PVertexInputState:   &vertexInputInfo,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisampling,
		PColorBlendState:    &colorBlending,
		PDynamicState:       &dynamicState,
		Layout:              vk.PipelineLayout(shader.RenderId.pipelineLayout.handle),
		RenderPass:          renderPass.Handle,
		BasePipelineHandle:  vk.Pipeline(vk.NullHandle),
		Subpass:             s.GraphicsPipeline.Subpass,
	}
	hasDepth := false
	for i := 0; i < len(renderPass.construction.SubpassDescriptions) && !hasDepth; i++ {
		hasDepth = len(renderPass.construction.SubpassDescriptions[i].DepthStencilAttachment) > 0
	}
	var depthStencil vk.PipelineDepthStencilStateCreateInfo
	if hasDepth {
		depthStencil = vk.PipelineDepthStencilStateCreateInfo{
			SType:                 vulkan_const.StructureTypePipelineDepthStencilStateCreateInfo,
			Flags:                 0, // PipelineDepthStencilStateCreateFlags
			DepthTestEnable:       boolToVkBool(s.DepthStencil.DepthTestEnable),
			DepthCompareOp:        s.DepthStencil.DepthCompareOp.toVulkan(),
			DepthBoundsTestEnable: boolToVkBool(s.DepthStencil.DepthBoundsTestEnable),
			StencilTestEnable:     boolToVkBool(s.DepthStencil.StencilTestEnable),
			MinDepthBounds:        s.DepthStencil.MinDepthBounds,
			MaxDepthBounds:        s.DepthStencil.MaxDepthBounds,
			DepthWriteEnable:      boolToVkBool(s.DepthStencil.DepthWriteEnable),
			Front:                 s.DepthStencil.Front.toVulkan(),
			Back:                  s.DepthStencil.Back.toVulkan(),
		}
		pipelineInfo.PDepthStencilState = &depthStencil
	}
	tess := vk.PipelineTessellationStateCreateInfo{}
	if len(shader.data.TessellationControl) > 0 ||
		len(shader.data.TessellationEvaluation) > 0 {
		tess.SType = vulkan_const.StructureTypePipelineTessellationStateCreateInfo
		tess.Flags = 0 // PipelineTessellationStateCreateFlags
		tess.PatchControlPoints = s.Tessellation.PatchControlPoints
		pipelineInfo.PTessellationState = &tess
	}
	success := true
	pipelines := [1]vk.Pipeline{}
	if vk.CreateGraphicsPipelines(vk.Device(device.LogicalDevice.handle), vk.PipelineCache(vk.NullHandle), 1, &pipelineInfo, nil, &pipelines[0]) != vulkan_const.Success {
		success = false
		slog.Error("Failed to create graphics pipeline")
	} else {
		device.LogicalDevice.dbg.track(unsafe.Pointer(pipelines[0]))
	}
	shader.RenderId.graphicsPipeline.handle = unsafe.Pointer(pipelines[0])
	return success
}

func (g GPUPrimitiveTopology) toVulkan() vulkan_const.PrimitiveTopology {
	return vulkan_const.PrimitiveTopology(g)
}

func (g GPUPolygonMode) toVulkan() vulkan_const.PolygonMode {
	return vulkan_const.PolygonMode(g)
}

func (g GPUCullModeFlags) toVulkan() vk.CullModeFlags {
	return vk.CullModeFlags(g)
}

func (g GPUFrontFace) toVulkan() vulkan_const.FrontFace {
	return vulkan_const.FrontFace(g)
}

func (g GPULogicOp) toVulkan() vulkan_const.LogicOp {
	return vulkan_const.LogicOp(g)
}

func (g GPUCompareOp) toVulkan() vulkan_const.CompareOp {
	return vulkan_const.CompareOp(g)
}

func (g GPUStencilOp) toVulkan() vulkan_const.StencilOp {
	return vulkan_const.StencilOp(g)
}

func (g GPUBlendFactor) toVulkan() vulkan_const.BlendFactor {
	return vulkan_const.BlendFactor(g)
}

func (g GPUBlendOp) toVulkan() vulkan_const.BlendOp {
	return vulkan_const.BlendOp(g)
}

func (g GPUColorComponentFlags) toVulkan() vk.ColorComponentFlags {
	return vk.ColorComponentFlags(g)
}

func (g GPUPipelineCreateFlags) toVulkan() vk.PipelineCreateFlags {
	return vk.PipelineCreateFlags(g)
}

func (g GPUShaderStageFlags) toVulkan() vk.ShaderStageFlags {
	return vk.ShaderStageFlags(g)
}

func (s GPUStencilOpState) toVulkan() vk.StencilOpState {
	return vk.StencilOpState{
		FailOp:      s.FailOp.toVulkan(),
		PassOp:      s.PassOp.toVulkan(),
		DepthFailOp: s.DepthFailOp.toVulkan(),
		CompareOp:   s.CompareOp.toVulkan(),
		CompareMask: s.CompareMask,
		WriteMask:   s.WriteMask,
		Reference:   s.Reference,
	}
}

func (s *ShaderPipelineData) PrimitiveRestartToVK() vk.Bool32 {
	return boolToVkBool(s.InputAssembly.PrimitiveRestart)
}

func (s *ShaderPipelineData) DepthClampEnableToVK() vk.Bool32 {
	return boolToVkBool(s.Rasterization.DepthClampEnable)
}

func (s *ShaderPipelineData) RasterizerDiscardEnableToVK() vk.Bool32 {
	return boolToVkBool(s.Rasterization.RasterizerDiscardEnable)
}

func (s *ShaderPipelineData) DepthBiasEnableToVK() vk.Bool32 {
	return boolToVkBool(s.Rasterization.DepthBiasEnable)
}

func (s *ShaderPipelineData) SampleShadingEnableToVK() vk.Bool32 {
	return boolToVkBool(s.Multisample.SampleShadingEnable)
}

func (s *ShaderPipelineData) AlphaToCoverageEnableToVK() vk.Bool32 {
	return boolToVkBool(s.Multisample.AlphaToCoverageEnable)
}

func (s *ShaderPipelineData) AlphaToOneEnableToVK() vk.Bool32 {
	return boolToVkBool(s.Multisample.AlphaToOneEnable)
}

func (s *ShaderPipelineData) LogicOpEnableToVK() vk.Bool32 {
	return boolToVkBool(s.ColorBlend.LogicOpEnable)
}

func (s *ShaderPipelineData) DepthTestEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthStencil.DepthTestEnable)
}

func (s *ShaderPipelineData) DepthWriteEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthStencil.DepthWriteEnable)
}

func (s *ShaderPipelineData) DepthBoundsTestEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthStencil.DepthBoundsTestEnable)
}

func (s *ShaderPipelineData) StencilTestEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthStencil.StencilTestEnable)
}

func (s *ShaderPipelineInputAssembly) TopologyToVK() vulkan_const.PrimitiveTopology {
	return s.TopologyToGPU().toVulkan()
}

func (s *ShaderPipelinePipelineRasterization) PolygonModeToVK() vulkan_const.PolygonMode {
	return s.PolygonModeToGPU().toVulkan()
}

func (s *ShaderPipelinePipelineRasterization) CullModeToVK() vulkan_const.CullModeFlagBits {
	return vulkan_const.CullModeFlagBits(s.CullModeToGPU())
}

func (s *ShaderPipelinePipelineRasterization) FrontFaceToVK() vulkan_const.FrontFace {
	return s.FrontFaceToGPU().toVulkan()
}

func (s *ShaderPipelinePipelineMultisample) RasterizationSamplesToVK(device *GPUPhysicalDevice) GPUSampleCountFlags {
	return s.RasterizationSamplesToGPU(device)
}

func (s *ShaderPipelineColorBlend) LogicOpToVK() vulkan_const.LogicOp {
	return s.LogicOpToGPU().toVulkan()
}

func (s *ShaderPipelineTessellation) PatchControlPointsToVK() uint32 {
	return s.PatchControlPointsToGPU()
}

func (s *ShaderPipelineData) FrontStencilOpStateToVK() vk.StencilOpState {
	return s.FrontStencilOpStateToGPU().toVulkan()
}

func (s *ShaderPipelineData) BackStencilOpStateToVK() vk.StencilOpState {
	return s.BackStencilOpStateToGPU().toVulkan()
}

func (s *ShaderPipelineGraphicsPipeline) PipelineCreateFlagsToVK() vk.PipelineCreateFlags {
	return s.PipelineCreateFlagsToGPU().toVulkan()
}

func (s *ShaderPipelinePushConstant) ShaderStageFlagsToVK() vk.ShaderStageFlags {
	return s.ShaderStageFlagsToGPU().toVulkan()
}

func (a *ShaderPipelineColorBlendAttachments) SrcColorBlendFactorToVK() vulkan_const.BlendFactor {
	return a.SrcColorBlendFactorToGPU().toVulkan()
}

func (a *ShaderPipelineColorBlendAttachments) DstColorBlendFactorToVK() vulkan_const.BlendFactor {
	return a.DstColorBlendFactorToGPU().toVulkan()
}

func (a *ShaderPipelineColorBlendAttachments) ColorBlendOpToVK() vulkan_const.BlendOp {
	return a.ColorBlendOpToGPU().toVulkan()
}

func (a *ShaderPipelineColorBlendAttachments) SrcAlphaBlendFactorToVK() vulkan_const.BlendFactor {
	return a.SrcAlphaBlendFactorToGPU().toVulkan()
}

func (a *ShaderPipelineColorBlendAttachments) DstAlphaBlendFactorToVK() vulkan_const.BlendFactor {
	return a.DstAlphaBlendFactorToGPU().toVulkan()
}

func (a *ShaderPipelineColorBlendAttachments) AlphaBlendOpToVK() vulkan_const.BlendOp {
	return a.AlphaBlendOpToGPU().toVulkan()
}

func (a *ShaderPipelineColorBlendAttachments) ColorWriteMaskToVK() vulkan_const.ColorComponentFlagBits {
	return vulkan_const.ColorComponentFlagBits(a.ColorWriteMaskToGPU())
}
