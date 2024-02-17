package rendering

import (
	"kaiju/assets"
	"log/slog"
	"strings"
	"unsafe"

	vk "github.com/KaijuEngine/go-vulkan"
)

func (vr *Vulkan) CreateShader(shader *Shader, assetDB *assets.Database) {
	var vert, frag, geom, tesc, tese vk.ShaderModule
	var vMem, fMem, gMem, cMem, eMem []byte
	vertStage := vk.PipelineShaderStageCreateInfo{}
	vMem, err := assetDB.Read(shader.VertPath)
	if err != nil {
		panic("Failed to load vertex shader")
	}
	vert, ok := vr.createSpvModule(vMem)
	if !ok {
		panic("Failed to create vertex shader module")
	}
	vertStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
	vertStage.Stage = vk.ShaderStageVertexBit
	vertStage.Module = vert
	vertStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.vertModule = vert

	fragStage := vk.PipelineShaderStageCreateInfo{}
	fMem, err = assetDB.Read(shader.FragPath)
	if err != nil {
		panic("Failed to load fragment shader")
	}
	frag, ok = vr.createSpvModule(fMem)
	if !ok {
		panic("Failed to create fragment shader module")
	}
	fragStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
	fragStage.Stage = vk.ShaderStageFragmentBit
	fragStage.Module = frag
	fragStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	shader.RenderId.fragModule = frag

	geomStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.GeomPath) > 0 {
		gMem, err = assetDB.Read(shader.GeomPath)
		if err != nil {
			panic("Failed to load geometry shader")
		}
		geom, ok = vr.createSpvModule(gMem)
		if !ok {
			panic("Failed to create geometry shader module")
		}
		geomStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		geomStage.Stage = vk.ShaderStageGeometryBit
		geomStage.Module = geom
		geomStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
	}

	tescStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.CtrlPath) > 0 {
		cMem, err = assetDB.Read(shader.CtrlPath)
		if err != nil {
			panic("Failed to load tessellation control shader")
		}
		tesc, ok = vr.createSpvModule(cMem)
		if !ok {
			panic("Failed to create tessellation control shader module")
		}
		tescStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		tescStage.Stage = vk.ShaderStageTessellationControlBit
		tescStage.Module = tesc
		tescStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.tescModule = tesc
	}

	teseStage := vk.PipelineShaderStageCreateInfo{}
	if len(shader.EvalPath) > 0 {
		eMem, err = assetDB.Read(shader.EvalPath)
		if err != nil {
			panic("Failed to load tessellation evaluation shader")
		}
		tese, ok = vr.createSpvModule(eMem)
		if !ok {
			panic("Failed to create tessellation evaluation shader module")
		}
		teseStage.SType = vk.StructureTypePipelineShaderStageCreateInfo
		teseStage.Stage = vk.ShaderStageTessellationEvaluationBit
		teseStage.Module = tese
		teseStage.PName = (*vk.Char)(unsafe.Pointer(&([]byte("main\x00"))[0]))
		shader.RenderId.teseModule = tese
	}

	id := &shader.RenderId

	id.descriptorSetLayout, err = vr.createDescriptorSetLayout(vr.device,
		shader.DriverData.DescriptorSetLayoutStructure)
	if err != nil {
		// TODO:  Handle this error properly
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

	shader.DriverData.pipelineConstructor(vr, shader, stages)
	// TODO:  Setup subshader in the shader definition?
	subShaderCheck := strings.TrimSuffix(shader.FragPath, ".spv") + oitSuffix
	if assetDB.Exists(subShaderCheck) {
		subShader := NewShader(shader.VertPath, subShaderCheck,
			shader.GeomPath, shader.CtrlPath, shader.EvalPath,
			&vr.defaultTarget.transparentRenderPass)
		subShader.DriverData = shader.DriverData
		shader.SubShader = subShader
	}
}

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

func defaultCreateShaderPipeline(renderer Renderer, shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo) bool {
	vr := renderer.(*Vulkan)
	isTransparentPipeline := !shader.IsComposite() &&
		shader.RenderPass == &vr.defaultTarget.transparentRenderPass
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
	pipelineLayoutInfo.SetLayoutCount = 1                                 // Optional
	pipelineLayoutInfo.PSetLayouts = &shader.RenderId.descriptorSetLayout // Optional
	pipelineLayoutInfo.PushConstantRangeCount = 0                         // Optional
	pipelineLayoutInfo.PPushConstantRanges = nil                          // Optional

	var pLayout vk.PipelineLayout
	if vk.CreatePipelineLayout(vr.device, &pipelineLayoutInfo, nil, &pLayout) != vk.Success {
		slog.Error("Failed to create pipeline layout")
		return false
	} else {
		vr.dbg.add(uintptr(unsafe.Pointer(pLayout)))
	}
	shader.RenderId.pipelineLayout = pLayout

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
	pipelineInfo.StageCount = uint32(len(shaderStages))
	pipelineInfo.PStages = &shaderStages[0]
	pipelineInfo.PVertexInputState = &vertexInputInfo
	pipelineInfo.PInputAssemblyState = &inputAssembly
	pipelineInfo.PViewportState = &viewportState
	pipelineInfo.PRasterizationState = &rasterizer
	pipelineInfo.PMultisampleState = &multisampling
	pipelineInfo.PColorBlendState = &colorBlending
	pipelineInfo.PDynamicState = &dynamicState
	pipelineInfo.Layout = shader.RenderId.pipelineLayout
	pipelineInfo.RenderPass = shader.RenderPass.Handle
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
	shader.RenderId.graphicsPipeline = pipelines[0]
	return success
}
