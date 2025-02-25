package rendering

import (
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
	"log/slog"
)

type ShaderPipelineData struct {
	Name                    string
	Topology                string `options:"StringVkPrimitiveTopology"`
	PrimitiveRestart        bool
	DepthClampEnable        bool
	RasterizerDiscardEnable bool
	PolygonMode             string `options:"StringVkPolygonMode"`
	CullMode                string `options:"StringVkCullModeFlagBits"`
	FrontFace               string `options:"StringVkFrontFace"`
	DepthBiasEnable         bool
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
	RasterizationSamples    string `options:"StringVkSampleCountFlagBits"`
	SampleShadingEnable     bool
	MinSampleShading        float32
	AlphaToCoverageEnable   bool
	AlphaToOneEnable        bool
	ColorBlendAttachments   []ShaderPipelineColorBlendAttachments
	LogicOpEnable           bool
	LogicOp                 string  `options:"StringVkLogicOp"`
	BlendConstants0         float32 `tip:"BlendConstants"`
	BlendConstants1         float32 `tip:"BlendConstants"`
	BlendConstants2         float32 `tip:"BlendConstants"`
	BlendConstants3         float32 `tip:"BlendConstants"`
	DepthTestEnable         bool
	DepthWriteEnable        bool
	DepthCompareOp          string `options:"StringVkCompareOp"`
	DepthBoundsTestEnable   bool
	StencilTestEnable       bool
	FrontFailOp             string `options:"StringVkStencilOp" tip:"FailOp"`
	FrontPassOp             string `options:"StringVkStencilOp" tip:"PassOp"`
	FrontDepthFailOp        string `options:"StringVkStencilOp" tip:"DepthFailOp"`
	FrontCompareOp          string `options:"StringVkCompareOp" tip:"CompareOp"`
	FrontCompareMask        uint32 `tip:"CompareMask"`
	FrontWriteMask          uint32 `tip:"WriteMask"`
	FrontReference          uint32 `tip:"Reference"`
	BackFailOp              string `options:"StringVkStencilOp" tip:"FailOp"`
	BackPassOp              string `options:"StringVkStencilOp" tip:"PassOp"`
	BackDepthFailOp         string `options:"StringVkStencilOp" tip:"DepthFailOp"`
	BackCompareOp           string `options:"StringVkCompareOp" tip:"CompareOp"`
	BackCompareMask         uint32 `tip:"CompareMask"`
	BackWriteMask           uint32 `tip:"WriteMask"`
	BackReference           uint32 `tip:"Reference"`
	MinDepthBounds          float32
	MaxDepthBounds          float32
	PatchControlPoints      string `options:"StringVkPatchControlPoints"`
	SubpassCount            uint32
	PipelineCreateFlags     []string `options:"StringVkPipelineCreateFlagBits"`
}

type ShaderPipelineColorBlendAttachments struct {
	BlendEnable         bool
	SrcColorBlendFactor string   `options:"StringVkBlendFactor"`
	DstColorBlendFactor string   `options:"StringVkBlendFactor"`
	ColorBlendOp        string   `options:"StringVkBlendOp"`
	SrcAlphaBlendFactor string   `options:"StringVkBlendFactor"`
	DstAlphaBlendFactor string   `options:"StringVkBlendFactor"`
	AlphaBlendOp        string   `options:"StringVkBlendOp"`
	ColorWriteMask      []string `options:"StringVkColorComponentFlagBits"`
}

func (a *ShaderPipelineColorBlendAttachments) ListSrcColorBlendFactor() []string {
	return klib.MapKeysSorted(StringVkBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) ListDstColorBlendFactor() []string {
	return klib.MapKeysSorted(StringVkBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) ListColorBlendOp() []string {
	return klib.MapKeysSorted(StringVkBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) ListSrcAlphaBlendFactor() []string {
	return klib.MapKeysSorted(StringVkBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) ListDstAlphaBlendFactor() []string {
	return klib.MapKeysSorted(StringVkBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) ListAlphaBlendOp() []string {
	return klib.MapKeysSorted(StringVkBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) BlendEnableToVK() vk.Bool32 {
	return boolToVkBool(a.BlendEnable)
}

func (a *ShaderPipelineColorBlendAttachments) SrcColorBlendFactorToVK() vk.BlendFactor {
	return blendFactorToVK(a.SrcColorBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) DstColorBlendFactorToVK() vk.BlendFactor {
	return blendFactorToVK(a.DstColorBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) ColorBlendOpToVK() vk.BlendOp {
	return blendOpToVK(a.ColorBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) SrcAlphaBlendFactorToVK() vk.BlendFactor {
	return blendFactorToVK(a.SrcAlphaBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) DstAlphaBlendFactorToVK() vk.BlendFactor {
	return blendFactorToVK(a.DstAlphaBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) AlphaBlendOpToVK() vk.BlendOp {
	return blendOpToVK(a.AlphaBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) ColorWriteMaskToVK() vk.ColorComponentFlagBits {
	mask := vk.ColorComponentFlagBits(0)
	for i := range a.ColorWriteMask {
		mask |= StringVkColorComponentFlagBits[a.ColorWriteMask[i]]
	}
	return mask
}

func (s ShaderPipelineData) ListTopology() []string {
	return klib.MapKeysSorted(StringVkPrimitiveTopology)
}

func (s ShaderPipelineData) ListPolygonMode() []string {
	return klib.MapKeysSorted(StringVkPolygonMode)
}

func (s ShaderPipelineData) ListCullMode() []string {
	return klib.MapKeysSorted(StringVkCullModeFlagBits)
}

func (s ShaderPipelineData) ListFrontFace() []string {
	return klib.MapKeysSorted(StringVkFrontFace)
}

func (s ShaderPipelineData) ListRasterizationSamples() []string {
	return klib.MapKeysSorted(StringVkSampleCountFlagBits)
}

func (s ShaderPipelineData) ListBlendFactor() []string {
	return klib.MapKeysSorted(StringVkBlendFactor)
}

func (s ShaderPipelineData) ListBlendOp() []string {
	return klib.MapKeysSorted(StringVkBlendOp)
}

func (s ShaderPipelineData) ListLogicOp() []string {
	return klib.MapKeysSorted(StringVkLogicOp)
}

func (s ShaderPipelineData) ListDepthCompareOp() []string {
	return klib.MapKeysSorted(StringVkCompareOp)
}

func (s ShaderPipelineData) ListBackCompareOp() []string {
	return klib.MapKeysSorted(StringVkCompareOp)
}

func (s ShaderPipelineData) ListFrontFailOp() []string {
	return klib.MapKeysSorted(StringVkStencilOp)
}

func (s ShaderPipelineData) ListFrontPassOp() []string {
	return klib.MapKeysSorted(StringVkStencilOp)
}

func (s ShaderPipelineData) ListFrontDepthFailOp() []string {
	return klib.MapKeysSorted(StringVkStencilOp)
}

func (s ShaderPipelineData) ListFrontCompareOp() []string {
	return klib.MapKeysSorted(StringVkStencilOp)
}

func (s ShaderPipelineData) ListBackFailOp() []string {
	return klib.MapKeysSorted(StringVkStencilOp)
}

func (s ShaderPipelineData) ListBackPassOp() []string {
	return klib.MapKeysSorted(StringVkStencilOp)
}

func (s ShaderPipelineData) ListBackDepthFailOp() []string {
	return klib.MapKeysSorted(StringVkStencilOp)
}

func (s ShaderPipelineData) ListPatchControlPoints() []string {
	return klib.MapKeysSorted(StringVkPatchControlPoints)
}

func (s *ShaderPipelineData) PrimitiveRestartToVK() vk.Bool32 {
	return boolToVkBool(s.PrimitiveRestart)
}

func (s *ShaderPipelineData) DepthClampEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthClampEnable)
}

func (s *ShaderPipelineData) RasterizerDiscardEnableToVK() vk.Bool32 {
	return boolToVkBool(s.RasterizerDiscardEnable)
}

func (s *ShaderPipelineData) DepthBiasEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthBiasEnable)
}

func (s *ShaderPipelineData) SampleShadingEnableToVK() vk.Bool32 {
	return boolToVkBool(s.SampleShadingEnable)
}

func (s *ShaderPipelineData) AlphaToCoverageEnableToVK() vk.Bool32 {
	return boolToVkBool(s.AlphaToCoverageEnable)
}

func (s *ShaderPipelineData) AlphaToOneEnableToVK() vk.Bool32 {
	return boolToVkBool(s.AlphaToOneEnable)
}

func (s *ShaderPipelineData) LogicOpEnableToVK() vk.Bool32 {
	return boolToVkBool(s.LogicOpEnable)
}

func (s *ShaderPipelineData) DepthTestEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthTestEnable)
}

func (s *ShaderPipelineData) DepthWriteEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthWriteEnable)
}

func (s *ShaderPipelineData) DepthBoundsTestEnableToVK() vk.Bool32 {
	return boolToVkBool(s.DepthBoundsTestEnable)
}

func (s *ShaderPipelineData) StencilTestEnableToVK() vk.Bool32 {
	return boolToVkBool(s.StencilTestEnable)
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
	return sampleCountToVK(s.RasterizationSamples)
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

func (s *ShaderPipelineData) PipelineCreateFlagsToVK() vk.PipelineCreateFlags {
	mask := vk.PipelineCreateFlagBits(0)
	for i := range s.PipelineCreateFlags {
		mask |= StringVkPipelineCreateFlagBits[s.PipelineCreateFlags[i]]
	}
	return vk.PipelineCreateFlags(mask)
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
		Flags:                  0, // PipelineInputAssemblyStateCreateFlags
		Topology:               s.TopologyToVK(),
		PrimitiveRestartEnable: s.PrimitiveRestartToVK(),
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
		Flags:                   0, // PipelineRasterizationStateCreateFlags
		DepthClampEnable:        s.DepthClampEnableToVK(),
		RasterizerDiscardEnable: s.RasterizerDiscardEnableToVK(),
		PolygonMode:             s.PolygonModeToVK(),
		LineWidth:               s.LineWidth,
		CullMode:                vk.CullModeFlags(s.CullModeToVK()),
		FrontFace:               s.FrontFaceToVK(),
		DepthBiasEnable:         s.DepthBiasEnableToVK(),
		DepthBiasConstantFactor: s.DepthBiasConstantFactor,
		DepthBiasClamp:          s.DepthBiasClamp,
		DepthBiasSlopeFactor:    s.DepthBiasSlopeFactor,
	}
	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		Flags:                 0, // PipelineMultisampleStateCreateFlags
		SampleShadingEnable:   s.SampleShadingEnableToVK(),
		RasterizationSamples:  s.RasterizationSamplesToVK(),
		MinSampleShading:      s.MinSampleShading,
		PSampleMask:           nil,
		AlphaToCoverageEnable: s.AlphaToCoverageEnableToVK(),
		AlphaToOneEnable:      s.AlphaToOneEnableToVK(),
	}
	colorBlendAttachment := make([]vk.PipelineColorBlendAttachmentState, len(s.ColorBlendAttachments))
	for i := range s.ColorBlendAttachments {
		colorBlendAttachment[i].BlendEnable = s.ColorBlendAttachments[i].BlendEnableToVK()
		colorBlendAttachment[i].SrcColorBlendFactor = s.ColorBlendAttachments[i].SrcColorBlendFactorToVK()
		colorBlendAttachment[i].DstColorBlendFactor = s.ColorBlendAttachments[i].DstColorBlendFactorToVK()
		colorBlendAttachment[i].ColorBlendOp = s.ColorBlendAttachments[i].ColorBlendOpToVK()
		colorBlendAttachment[i].SrcAlphaBlendFactor = s.ColorBlendAttachments[i].SrcAlphaBlendFactorToVK()
		colorBlendAttachment[i].DstAlphaBlendFactor = s.ColorBlendAttachments[i].DstAlphaBlendFactorToVK()
		colorBlendAttachment[i].AlphaBlendOp = s.ColorBlendAttachments[i].AlphaBlendOpToVK()
		writeMask := s.ColorBlendAttachments[i].ColorWriteMaskToVK()
		colorBlendAttachment[i].ColorWriteMask = vk.ColorComponentFlags(writeMask)
	}
	colorBlendAttachmentCount := len(colorBlendAttachment)
	colorBlending := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		Flags:           0, // PipelineColorBlendStateCreateFlags
		LogicOpEnable:   s.LogicOpEnableToVK(),
		LogicOp:         s.LogicOpToVK(),
		AttachmentCount: uint32(colorBlendAttachmentCount),
		PAttachments:    &colorBlendAttachment[0],
		BlendConstants:  s.BlendConstants(),
	}
	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		Flags:                  0, // PipelineLayoutCreateFlags
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
		Flags:                 0, // PipelineDepthStencilStateCreateFlags
		DepthTestEnable:       s.DepthTestEnableToVK(),
		DepthCompareOp:        compareOpToVK(s.DepthCompareOp),
		DepthBoundsTestEnable: s.DepthBoundsTestEnableToVK(),
		StencilTestEnable:     s.StencilTestEnableToVK(),
		MinDepthBounds:        s.MinDepthBounds,
		MaxDepthBounds:        s.MaxDepthBounds,
		DepthWriteEnable:      s.DepthWriteEnableToVK(),
		Front:                 s.FrontStencilOpStateToVK(),
		Back:                  s.BackStencilOpStateToVK(),
	}
	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		Flags:               s.PipelineCreateFlagsToVK(),
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
		tess.Flags = 0 // PipelineTessellationStateCreateFlags
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
