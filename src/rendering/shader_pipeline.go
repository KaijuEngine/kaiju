package rendering

import (
	vk "kaiju/rendering/vulkan"
	"log/slog"
)

type ShaderPipelineColorBlendAttachments struct {
	BlendEnable         bool
	SrcColorBlendFactor string
	DstColorBlendFactor string
	ColorBlendOp        string
	SrcAlphaBlendFactor string
	DstAlphaBlendFactor string
	AlphaBlendOp        string
	ColorWriteMaskR     bool
	ColorWriteMaskG     bool
	ColorWriteMaskB     bool
	ColorWriteMaskA     bool
}

type ShaderPipelineData struct {
	Name                    string
	Topology                string
	PrimitiveRestart        bool
	DepthClampEnable        bool
	RasterizerDiscardEnable bool
	PolygonMode             string
	CullMode                string
	FrontFace               string
	DepthBiasEnable         bool
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
	RasterizationSamples    string
	SampleShadingEnable     bool
	MinSampleShading        float32
	AlphaToCoverageEnable   bool
	AlphaToOneEnable        bool
	ColorBlendAttachments   []ShaderPipelineColorBlendAttachments
	LogicOpEnable           bool
	LogicOp                 string
	BlendConstants0         float32
	BlendConstants1         float32
	BlendConstants2         float32
	BlendConstants3         float32
	DepthTestEnable         bool
	DepthWriteEnable        bool
	DepthCompareOp          string
	DepthBoundsTestEnable   bool
	StencilTestEnable       bool
	FrontFailOp             string
	FrontPassOp             string
	FrontDepthFailOp        string
	FrontCompareOp          string
	FrontCompareMask        uint32
	FrontWriteMask          uint32
	FrontReference          uint32
	BackFailOp              string
	BackPassOp              string
	BackDepthFailOp         string
	BackCompareOp           string
	BackCompareMask         uint32
	BackWriteMask           uint32
	BackReference           uint32
	MinDepthBounds          float32
	MaxDepthBounds          float32
	PatchControlPoints      string
	SubpassCount            uint32
}

func boolToVkBool(val bool) vk.Bool32 {
	if val {
		return vk.True
	} else {
		return vk.False
	}
}

func blendFactorToVK(val string) vk.BlendFactor {
	if res, ok := StringVkBlendFactor[val]; ok {
		return res
	}
	slog.Warn("invalid string for vkBlendFactor", "value", val)
	return vk.BlendFactorSrcAlpha
}

func blendOpToVK(val string) vk.BlendOp {
	if res, ok := StringVkBlendOp[val]; ok {
		return res
	}
	slog.Warn("invalid string for vkBlendOp", "value", val)
	return vk.BlendOpAdd
}

func (s *ShaderPipelineData) TopologyToVK() vk.PrimitiveTopology {
	if res, ok := StringVkPrimitiveTopology[s.Topology]; ok {
		return res
	}
	slog.Warn("invalid string for vkPrimitiveTopology", "value", s.Topology)
	return vk.PrimitiveTopologyTriangleList
}

func (s *ShaderPipelineData) PolygonModeToVK() vk.PolygonMode {
	if res, ok := StringVkPolygonMode[s.PolygonMode]; ok {
		return res
	}
	slog.Warn("invalid string for vkPolygonMode", "value", s.PolygonMode)
	return vk.PolygonModeFill
}

func (s *ShaderPipelineData) CullModeToVK() vk.CullModeFlagBits {
	if res, ok := StringVkCullModeFlagBits[s.CullMode]; ok {
		return res
	}
	slog.Warn("invalid string for vkCullModeFlagBits", "value", s.CullMode)
	return vk.CullModeFrontBit
}

func (s *ShaderPipelineData) FrontFaceToVK() vk.FrontFace {
	if res, ok := StringVkFrontFace[s.FrontFace]; ok {
		return res
	}
	slog.Warn("invalid string for vkFrontFace", "value", s.FrontFace)
	return vk.FrontFaceClockwise
}

func (s *ShaderPipelineData) RasterizationSamplesToVK() vk.SampleCountFlagBits {
	if res, ok := StringVkSampleCountFlagBits[s.RasterizationSamples]; ok {
		return res
	}
	slog.Warn("invalid string for vkRasterizationSamples", "value", s.RasterizationSamples)
	return vk.SampleCount1Bit
}

func (s *ShaderPipelineData) LogicOpToVK() vk.LogicOp {
	if res, ok := StringVkLogicOp[s.LogicOp]; ok {
		return res
	}
	slog.Warn("invalid string for vkLogicOp", "value", s.LogicOp)
	return vk.LogicOpCopy
}

func (s *ShaderPipelineData) BlendConstants() [4]float32 {
	return [4]float32{
		s.BlendConstants0,
		s.BlendConstants1,
		s.BlendConstants2,
		s.BlendConstants3,
	}
}

func compareOpToVK(val string) vk.CompareOp {
	if res, ok := StringVkCompareOp[val]; ok {
		return res
	}
	slog.Warn("invalid string for vkCompareOp", "value", val)
	return vk.CompareOpLess
}

func stencilOpToVK(val string) vk.StencilOp {
	if res, ok := StringVkStencilOp[val]; ok {
		return res
	}
	slog.Warn("invalid string for vkStencilOpKeep", "value", val)
	return vk.StencilOpKeep
}

func (s *ShaderPipelineData) PatchControlPointsToVK() uint32 {
	if res, ok := StringVkPatchControlPoints[s.PatchControlPoints]; ok {
		return res
	}
	slog.Warn("invalid string for PatchControlPoints", "value", s.PatchControlPoints)
	return 3
}

// TODO:  This and the BackStencilOpStateToVK are duplicates because of a bad
// structure setup, please fix later
func (s *ShaderPipelineData) FrontStencilOpStateToVK() vk.StencilOpState {
	return vk.StencilOpState{
		FailOp:      stencilOpToVK(s.FrontFailOp),
		PassOp:      stencilOpToVK(s.FrontPassOp),
		DepthFailOp: stencilOpToVK(s.FrontDepthFailOp),
		CompareOp:   compareOpToVK(s.FrontCompareOp),
		CompareMask: s.FrontCompareMask,
		WriteMask:   s.FrontWriteMask,
		Reference:   s.FrontReference,
	}
}

func (s *ShaderPipelineData) BackStencilOpStateToVK() vk.StencilOpState {
	return vk.StencilOpState{
		FailOp:      stencilOpToVK(s.BackFailOp),
		PassOp:      stencilOpToVK(s.BackPassOp),
		DepthFailOp: stencilOpToVK(s.BackDepthFailOp),
		CompareOp:   compareOpToVK(s.BackCompareOp),
		CompareMask: s.BackCompareMask,
		WriteMask:   s.BackWriteMask,
		Reference:   s.BackReference,
	}
}

func (s *ShaderPipelineData) ConstructPipeline(renderer Renderer, shader *Shader, shaderStages []vk.PipelineShaderStageCreateInfo) bool {
	vr := renderer.(*Vulkan)
	bDesc := vertexGetBindingDescription(shader)
	bDescCount := uint32(len(bDesc))
	for i := uint32(1); i < bDescCount; i++ {
		bDesc[i].Stride = uint32(vr.padUniformBufferSize(vk.DeviceSize(bDesc[i].Stride)))
	}
	aDesc := vertexGetAttributeDescription(shader)
	vertexInputInfo := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   bDescCount,
		VertexAttributeDescriptionCount: uint32(len(aDesc)),
		PVertexBindingDescriptions:      &bDesc[0],
		PVertexAttributeDescriptions:    &aDesc[0],
	}
	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               s.TopologyToVK(),
		PrimitiveRestartEnable: boolToVkBool(s.PrimitiveRestart),
	}
	viewport := vk.Viewport{
		X:        0.0,
		Y:        0.0,
		Width:    float32(vr.swapChainExtent.Width),
		Height:   float32(vr.swapChainExtent.Height),
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{X: 0, Y: 0},
		Extent: vr.swapChainExtent,
	}
	dynamicStates := []vk.DynamicState{
		vk.DynamicStateViewport,
		vk.DynamicStateScissor,
	}
	dynamicState := vk.PipelineDynamicStateCreateInfo{
		SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynamicStates)),
		PDynamicStates:    &dynamicStates[0],
	}
	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    &viewport,
		ScissorCount:  1,
		PScissors:     &scissor,
	}
	rasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        boolToVkBool(s.DepthClampEnable),
		RasterizerDiscardEnable: boolToVkBool(s.RasterizerDiscardEnable),
		PolygonMode:             s.PolygonModeToVK(),
		LineWidth:               s.LineWidth,
		CullMode:                vk.CullModeFlags(s.CullModeToVK()),
		FrontFace:               s.FrontFaceToVK(),
		DepthBiasEnable:         boolToVkBool(s.DepthBiasEnable),
		DepthBiasConstantFactor: s.DepthBiasConstantFactor,
		DepthBiasClamp:          s.DepthBiasClamp,
		DepthBiasSlopeFactor:    s.DepthBiasSlopeFactor,
	}
	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		SampleShadingEnable:   boolToVkBool(s.SampleShadingEnable),
		RasterizationSamples:  s.RasterizationSamplesToVK(),
		MinSampleShading:      s.MinSampleShading,
		PSampleMask:           nil,
		AlphaToCoverageEnable: boolToVkBool(s.AlphaToCoverageEnable),
		AlphaToOneEnable:      boolToVkBool(s.AlphaToOneEnable),
	}
	colorBlendAttachment := make([]vk.PipelineColorBlendAttachmentState, len(s.ColorBlendAttachments))
	for i := range s.ColorBlendAttachments {
		colorBlendAttachment[i].BlendEnable = boolToVkBool(s.ColorBlendAttachments[i].BlendEnable)
		colorBlendAttachment[i].SrcColorBlendFactor = blendFactorToVK(s.ColorBlendAttachments[i].SrcColorBlendFactor)
		colorBlendAttachment[i].DstColorBlendFactor = blendFactorToVK(s.ColorBlendAttachments[i].DstColorBlendFactor)
		colorBlendAttachment[i].ColorBlendOp = blendOpToVK(s.ColorBlendAttachments[i].ColorBlendOp)
		colorBlendAttachment[i].SrcAlphaBlendFactor = blendFactorToVK(s.ColorBlendAttachments[i].SrcAlphaBlendFactor)
		colorBlendAttachment[i].DstAlphaBlendFactor = blendFactorToVK(s.ColorBlendAttachments[i].DstAlphaBlendFactor)
		colorBlendAttachment[i].AlphaBlendOp = blendOpToVK(s.ColorBlendAttachments[i].AlphaBlendOp)
		var writeMask vk.ColorComponentFlagBits = 0
		if s.ColorBlendAttachments[i].ColorWriteMaskR {
			writeMask |= vk.ColorComponentRBit
		}
		if s.ColorBlendAttachments[i].ColorWriteMaskG {
			writeMask |= vk.ColorComponentGBit
		}
		if s.ColorBlendAttachments[i].ColorWriteMaskB {
			writeMask |= vk.ColorComponentBBit
		}
		if s.ColorBlendAttachments[i].ColorWriteMaskA {
			writeMask |= vk.ColorComponentABit
		}
		colorBlendAttachment[i].ColorWriteMask = vk.ColorComponentFlags(writeMask)
	}
	colorBlendAttachmentCount := len(colorBlendAttachment)
	colorBlending := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   boolToVkBool(s.LogicOpEnable),
		LogicOp:         s.LogicOpToVK(),
		AttachmentCount: uint32(colorBlendAttachmentCount),
		PAttachments:    &colorBlendAttachment[0],
		BlendConstants:  s.BlendConstants(),
	}
	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            &shader.RenderId.descriptorSetLayout,
		PushConstantRangeCount: 0,
		PPushConstantRanges:    nil,
	}
	var pLayout vk.PipelineLayout
	if vk.CreatePipelineLayout(vr.device, &pipelineLayoutInfo, nil, &pLayout) != vk.Success {
		slog.Error("Failed to create pipeline layout")
		return false
	} else {
		vr.dbg.add(vk.TypeToUintPtr(pLayout))
	}
	shader.RenderId.pipelineLayout = pLayout
	depthStencil := vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       boolToVkBool(s.DepthTestEnable),
		DepthCompareOp:        compareOpToVK(s.DepthCompareOp),
		DepthBoundsTestEnable: boolToVkBool(s.DepthBoundsTestEnable),
		StencilTestEnable:     boolToVkBool(s.StencilTestEnable),
		MinDepthBounds:        s.MinDepthBounds,
		MaxDepthBounds:        s.MaxDepthBounds,
		DepthWriteEnable:      boolToVkBool(s.DepthWriteEnable),
		Front:                 s.FrontStencilOpStateToVK(),
		Back:                  s.BackStencilOpStateToVK(),
	}
	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          uint32(len(shaderStages)),
		PStages:             &shaderStages[0],
		PVertexInputState:   &vertexInputInfo,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisampling,
		PColorBlendState:    &colorBlending,
		PDynamicState:       &dynamicState,
		Layout:              shader.RenderId.pipelineLayout,
		RenderPass:          shader.RenderPass.Handle,
		BasePipelineHandle:  vk.Pipeline(vk.NullHandle),
		PDepthStencilState:  &depthStencil,
		Subpass:             s.SubpassCount,
	}
	tess := vk.PipelineTessellationStateCreateInfo{}
	if len(shader.CtrlPath) > 0 || len(shader.EvalPath) > 0 {
		tess.SType = vk.StructureTypePipelineTessellationStateCreateInfo
		tess.PatchControlPoints = s.PatchControlPointsToVK()
		pipelineInfo.PTessellationState = &tess
	}
	success := true
	pipelines := [1]vk.Pipeline{}
	if vk.CreateGraphicsPipelines(vr.device, vk.PipelineCache(vk.NullHandle), 1, &pipelineInfo, nil, &pipelines[0]) != vk.Success {
		success = false
		slog.Error("Failed to create graphics pipeline")
	} else {
		vr.dbg.add(vk.TypeToUintPtr(pipelines[0]))
	}
	shader.RenderId.graphicsPipeline = pipelines[0]
	return success
}
