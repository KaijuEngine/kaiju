package rendering

import (
	"log/slog"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) createSpvModule(mem []byte) (vk.ShaderModule, bool) {
	info := vk.ShaderModuleCreateInfo{}
	info.SType = vk.StructureTypeShaderModuleCreateInfo
	info.CodeSize = uint(len(mem))
	info.PCode = (*uint32)(unsafe.Pointer(&mem[0]))
	var outModule vk.ShaderModule
	if vk.CreateShaderModule(vr.device, &info, nil, &outModule) != vk.Success {
		slog.Error("Failed to create shader module", slog.String("module", string(mem)))
		return outModule, false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(outModule)))
		return outModule, true
	}
}

func (vr *Vulkan) createPipeline(shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo,
	shaderStageCount int, descriptorSetLayout vk.DescriptorSetLayout,
	pipelineLayout *vk.PipelineLayout, graphicsPipeline *vk.Pipeline,
	renderPass RenderPass, isTransparentPipeline bool) bool {
	bDesc := vertexGetBindingDescription(shader)
	bDescCount := uint32(len(bDesc))
	if shader.IsComposite() {
		bDescCount = 1
	}
	for i := uint32(1); i < bDescCount; i++ {
		bDesc[i].Stride = uint32(vr.padUniformBufferSize(vk.DeviceSize(bDesc[i].Stride)))
	}
	aDesc := vertexGetAttributeDescription(shader)
	vertexInputInfo := vk.PipelineVertexInputStateCreateInfo{}
	vertexInputInfo.SType = vk.StructureTypePipelineVertexInputStateCreateInfo
	vertexInputInfo.VertexBindingDescriptionCount = bDescCount
	vertexInputInfo.VertexAttributeDescriptionCount = uint32(len(aDesc))
	vertexInputInfo.PVertexBindingDescriptions = &bDesc[0]   // Optional
	vertexInputInfo.PVertexAttributeDescriptions = &aDesc[0] // Optional

	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{}
	inputAssembly.SType = vk.StructureTypePipelineInputAssemblyStateCreateInfo
	switch shader.DriverData.DrawMode {
	case MeshDrawModePoints:
		inputAssembly.Topology = vk.PrimitiveTopologyPointList
	case MeshDrawModeLines:
		inputAssembly.Topology = vk.PrimitiveTopologyLineList
	case MeshDrawModeTriangles:
		inputAssembly.Topology = vk.PrimitiveTopologyTriangleList
	case MeshDrawModePatches:
		inputAssembly.Topology = vk.PrimitiveTopologyPatchList
	}
	inputAssembly.PrimitiveRestartEnable = vk.False

	viewport := vk.Viewport{}
	viewport.X = 0.0
	viewport.Y = 0.0
	viewport.Width = float32(vr.swapChainExtent.Width)
	viewport.Height = float32(vr.swapChainExtent.Height)
	viewport.MinDepth = 0.0
	viewport.MaxDepth = 1.0

	scissor := vk.Rect2D{}
	scissor.Offset = vk.Offset2D{X: 0, Y: 0}
	scissor.Extent = vr.swapChainExtent

	dynamicStates := []vk.DynamicState{
		vk.DynamicStateViewport,
		vk.DynamicStateScissor,
	}

	dynamicState := vk.PipelineDynamicStateCreateInfo{}
	dynamicState.SType = vk.StructureTypePipelineDynamicStateCreateInfo
	dynamicState.DynamicStateCount = uint32(len(dynamicStates))
	dynamicState.PDynamicStates = &dynamicStates[0]

	viewportState := vk.PipelineViewportStateCreateInfo{}
	viewportState.SType = vk.StructureTypePipelineViewportStateCreateInfo
	viewportState.ViewportCount = 1
	viewportState.PViewports = &viewport
	viewportState.ScissorCount = 1
	viewportState.PScissors = &scissor

	rasterizer := vk.PipelineRasterizationStateCreateInfo{}
	rasterizer.SType = vk.StructureTypePipelineRasterizationStateCreateInfo
	rasterizer.DepthClampEnable = vk.False
	rasterizer.RasterizerDiscardEnable = vk.False
	rasterizer.PolygonMode = vk.PolygonModeFill
	rasterizer.LineWidth = 1.0
	rasterizer.CullMode = vk.CullModeFlags(shader.DriverData.CullMode)
	rasterizer.FrontFace = vk.FrontFaceClockwise
	rasterizer.DepthBiasEnable = vk.False
	rasterizer.DepthBiasConstantFactor = 0.0 // Optional
	rasterizer.DepthBiasClamp = 0.0          // Optional
	rasterizer.DepthBiasSlopeFactor = 0.0    // Optional

	multisampling := vk.PipelineMultisampleStateCreateInfo{}
	multisampling.SType = vk.StructureTypePipelineMultisampleStateCreateInfo
	multisampling.SampleShadingEnable = vk.True // Optional
	// TODO:  This is a temp hack for testing
	multisampling.RasterizationSamples = vk.SampleCount1Bit //shader.uniformType == SHADER_UNIFORM_TYPE_DEPTH ? 1 : vr.msaaSamples;
	multisampling.MinSampleShading = 0.2                    // Optional
	multisampling.PSampleMask = nil                         // Optional
	multisampling.AlphaToCoverageEnable = vk.False          // Optional
	multisampling.AlphaToOneEnable = vk.False               // Optional

	allChannels := vk.ColorComponentFlags(vk.ColorComponentRBit | vk.ColorComponentGBit | vk.ColorComponentBBit | vk.ColorComponentABit)
	var colorBlendAttachment [2]vk.PipelineColorBlendAttachmentState
	colorBlendAttachment[0].ColorWriteMask = allChannels
	colorBlendAttachment[0].BlendEnable = vk.True
	colorBlendAttachment[0].SrcColorBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].DstColorBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].ColorBlendOp = vk.BlendOpAdd
	colorBlendAttachment[0].SrcAlphaBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].DstAlphaBlendFactor = vk.BlendFactorOne
	colorBlendAttachment[0].AlphaBlendOp = vk.BlendOpAdd

	colorBlendAttachment[1].ColorWriteMask = allChannels
	colorBlendAttachment[1].BlendEnable = vk.True
	colorBlendAttachment[1].SrcColorBlendFactor = vk.BlendFactorZero
	colorBlendAttachment[1].DstColorBlendFactor = vk.BlendFactorOneMinusSrcColor
	colorBlendAttachment[1].ColorBlendOp = vk.BlendOpAdd
	colorBlendAttachment[1].SrcAlphaBlendFactor = vk.BlendFactorZero
	colorBlendAttachment[1].DstAlphaBlendFactor = vk.BlendFactorOneMinusSrcAlpha
	colorBlendAttachment[1].AlphaBlendOp = vk.BlendOpAdd
	colorBlendAttachmentCount := len(colorBlendAttachment)

	if !isTransparentPipeline {
		if shader.IsComposite() {
			colorBlendAttachment[0].SrcColorBlendFactor = vk.BlendFactorOneMinusSrcAlpha
			colorBlendAttachment[0].DstColorBlendFactor = vk.BlendFactorSrcAlpha
			colorBlendAttachment[0].SrcAlphaBlendFactor = vk.BlendFactorOneMinusSrcAlpha
			colorBlendAttachment[0].DstAlphaBlendFactor = vk.BlendFactorSrcAlpha
		} else {
			colorBlendAttachment[0].SrcColorBlendFactor = vk.BlendFactorSrcAlpha
			colorBlendAttachment[0].DstColorBlendFactor = vk.BlendFactorOneMinusSrcAlpha
			colorBlendAttachment[0].SrcAlphaBlendFactor = vk.BlendFactorOne
			colorBlendAttachment[0].DstAlphaBlendFactor = vk.BlendFactorZero
		}
		colorBlendAttachmentCount = 1
	}

	colorBlending := vk.PipelineColorBlendStateCreateInfo{}
	colorBlending.SType = vk.StructureTypePipelineColorBlendStateCreateInfo
	colorBlending.LogicOpEnable = vk.False
	colorBlending.LogicOp = vk.LogicOpCopy // Optional
	colorBlending.AttachmentCount = uint32(colorBlendAttachmentCount)
	colorBlending.PAttachments = &colorBlendAttachment[0]
	colorBlending.BlendConstants[0] = 0.0 // Optional
	colorBlending.BlendConstants[1] = 0.0 // Optional
	colorBlending.BlendConstants[2] = 0.0 // Optional
	colorBlending.BlendConstants[3] = 0.0 // Optional

	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{}
	pipelineLayoutInfo.SType = vk.StructureTypePipelineLayoutCreateInfo
	pipelineLayoutInfo.SetLayoutCount = 1                 // Optional
	pipelineLayoutInfo.PSetLayouts = &descriptorSetLayout // Optional
	pipelineLayoutInfo.PushConstantRangeCount = 0         // Optional
	pipelineLayoutInfo.PPushConstantRanges = nil          // Optional

	var pLayout vk.PipelineLayout
	if vk.CreatePipelineLayout(vr.device, &pipelineLayoutInfo, nil, &pLayout) != vk.Success {
		slog.Error("Failed to create pipeline layout")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(pLayout)))
	}
	*pipelineLayout = pLayout

	depthStencil := vk.PipelineDepthStencilStateCreateInfo{}
	depthStencil.SType = vk.StructureTypePipelineDepthStencilStateCreateInfo
	depthStencil.DepthTestEnable = vk.True
	if isTransparentPipeline {
		depthStencil.DepthWriteEnable = vk.False
	} else {
		depthStencil.DepthWriteEnable = vk.True
	}
	depthStencil.DepthCompareOp = vk.CompareOpLess
	depthStencil.DepthBoundsTestEnable = vk.False
	//depthStencil.minDepthBounds = 0.0F; // Optional
	//depthStencil.maxDepthBounds = 1.0F; // Optional
	depthStencil.StencilTestEnable = vk.False

	pipelineInfo := vk.GraphicsPipelineCreateInfo{}
	pipelineInfo.SType = vk.StructureTypeGraphicsPipelineCreateInfo
	pipelineInfo.StageCount = uint32(shaderStageCount)
	pipelineInfo.PStages = &shaderStages[:shaderStageCount][0]
	pipelineInfo.PVertexInputState = &vertexInputInfo
	pipelineInfo.PInputAssemblyState = &inputAssembly
	pipelineInfo.PViewportState = &viewportState
	pipelineInfo.PRasterizationState = &rasterizer
	pipelineInfo.PMultisampleState = &multisampling
	pipelineInfo.PColorBlendState = &colorBlending
	pipelineInfo.PDynamicState = &dynamicState
	pipelineInfo.Layout = *pipelineLayout
	pipelineInfo.RenderPass = renderPass.Handle
	//pipelineInfo.Subpass = 0
	//s := shader.SubShader
	//for s != nil {
	//	s = s.SubShader
	//	pipelineInfo.Subpass++
	//}
	if shader.IsComposite() {
		pipelineInfo.Subpass = 1
	} else {
		pipelineInfo.Subpass = 0
	}
	pipelineInfo.BasePipelineHandle = vk.Pipeline(vk.NullHandle)
	pipelineInfo.PDepthStencilState = &depthStencil

	tess := vk.PipelineTessellationStateCreateInfo{}
	if len(shader.CtrlPath) > 0 || len(shader.EvalPath) > 0 {
		tess.SType = vk.StructureTypePipelineTessellationStateCreateInfo
		// Quad patches = 4
		// Triangle patches = 3
		// Line patches = 2
		tess.PatchControlPoints = 3
		pipelineInfo.PTessellationState = &tess
	}

	success := true
	pipelines := [1]vk.Pipeline{}
	if vk.CreateGraphicsPipelines(vr.device, vk.PipelineCache(vk.NullHandle), 1, &pipelineInfo, nil, &pipelines[0]) != vk.Success {
		success = false
		slog.Error("Failed to create graphics pipeline")
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(pipelines[0])))
	}
	*graphicsPipeline = pipelines[0]
	return success
}
