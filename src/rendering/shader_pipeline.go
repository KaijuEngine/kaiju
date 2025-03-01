package rendering

import (
	"kaiju/klib"
	vk "kaiju/rendering/vulkan"
	"log/slog"
)

type ShaderPipelineInputAssembly struct {
	Topology         string `options:"StringVkPrimitiveTopology"`
	PrimitiveRestart bool
}

type ShaderPipelinePipelineRasterization struct {
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
}

type ShaderPipelinePipelineMultisample struct {
	RasterizationSamples  string `options:"StringVkSampleCountFlagBits"`
	SampleShadingEnable   bool
	MinSampleShading      float32
	AlphaToCoverageEnable bool
	AlphaToOneEnable      bool
}

type ShaderPipelineColorBlend struct {
	LogicOpEnable   bool
	LogicOp         string  `options:"StringVkLogicOp"`
	BlendConstants0 float32 `tip:"BlendConstants"`
	BlendConstants1 float32 `tip:"BlendConstants"`
	BlendConstants2 float32 `tip:"BlendConstants"`
	BlendConstants3 float32 `tip:"BlendConstants"`
}

type ShaderPipelineDepthStencil struct {
	DepthTestEnable       bool
	DepthWriteEnable      bool
	DepthCompareOp        string `options:"StringVkCompareOp"`
	DepthBoundsTestEnable bool
	StencilTestEnable     bool
	FrontFailOp           string `options:"StringVkStencilOp" tip:"FailOp"`
	FrontPassOp           string `options:"StringVkStencilOp" tip:"PassOp"`
	FrontDepthFailOp      string `options:"StringVkStencilOp" tip:"DepthFailOp"`
	FrontCompareOp        string `options:"StringVkCompareOp" tip:"CompareOp"`
	FrontCompareMask      uint32 `tip:"CompareMask"`
	FrontWriteMask        uint32 `tip:"WriteMask"`
	FrontReference        uint32 `tip:"Reference"`
	BackFailOp            string `options:"StringVkStencilOp" tip:"FailOp"`
	BackPassOp            string `options:"StringVkStencilOp" tip:"PassOp"`
	BackDepthFailOp       string `options:"StringVkStencilOp" tip:"DepthFailOp"`
	BackCompareOp         string `options:"StringVkCompareOp" tip:"CompareOp"`
	BackCompareMask       uint32 `tip:"CompareMask"`
	BackWriteMask         uint32 `tip:"WriteMask"`
	BackReference         uint32 `tip:"Reference"`
	MinDepthBounds        float32
	MaxDepthBounds        float32
}

type ShaderPipelineTessellation struct {
	PatchControlPoints string `options:"StringVkPatchControlPoints"`
}

type ShaderPipelineGraphicsPipeline struct {
	SubpassCount        uint32
	PipelineCreateFlags []string `options:"StringVkPipelineCreateFlagBits"`
}

type ShaderPipelineData struct {
	Name                  string
	InputAssembly         ShaderPipelineInputAssembly
	Rasterization         ShaderPipelinePipelineRasterization
	Multisample           ShaderPipelinePipelineMultisample
	ColorBlend            ShaderPipelineColorBlend
	ColorBlendAttachments []ShaderPipelineColorBlendAttachments
	DepthStencil          ShaderPipelineDepthStencil
	Tessellation          ShaderPipelineTessellation
	GraphicsPipeline      ShaderPipelineGraphicsPipeline
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

func (s *ShaderPipelineData) TopologyToVK() vk.PrimitiveTopology {
	if res, ok := StringVkPrimitiveTopology[s.InputAssembly.Topology]; ok {
		return res
	}
	slog.Warn("invalid string for vkPrimitiveTopology", "value", s.InputAssembly.Topology)
	return vk.PrimitiveTopologyTriangleList
}

func (s *ShaderPipelineData) PolygonModeToVK() vk.PolygonMode {
	if res, ok := StringVkPolygonMode[s.Rasterization.PolygonMode]; ok {
		return res
	}
	slog.Warn("invalid string for vkPolygonMode", "value", s.Rasterization.PolygonMode)
	return vk.PolygonModeFill
}

func (s *ShaderPipelineData) CullModeToVK() vk.CullModeFlagBits {
	if res, ok := StringVkCullModeFlagBits[s.Rasterization.CullMode]; ok {
		return res
	}
	slog.Warn("invalid string for vkCullModeFlagBits", "value", s.Rasterization.CullMode)
	return vk.CullModeFrontBit
}

func (s *ShaderPipelineData) FrontFaceToVK() vk.FrontFace {
	if res, ok := StringVkFrontFace[s.Rasterization.FrontFace]; ok {
		return res
	}
	slog.Warn("invalid string for vkFrontFace", "value", s.Rasterization.FrontFace)
	return vk.FrontFaceClockwise
}

func (s *ShaderPipelineData) RasterizationSamplesToVK() vk.SampleCountFlagBits {
	return sampleCountToVK(s.Multisample.RasterizationSamples)
}

func (s *ShaderPipelineData) LogicOpToVK() vk.LogicOp {
	if res, ok := StringVkLogicOp[s.ColorBlend.LogicOp]; ok {
		return res
	}
	slog.Warn("invalid string for vkLogicOp", "value", s.ColorBlend.LogicOp)
	return vk.LogicOpCopy
}

func (s *ShaderPipelineData) BlendConstants() [4]float32 {
	return [4]float32{
		s.ColorBlend.BlendConstants0,
		s.ColorBlend.BlendConstants1,
		s.ColorBlend.BlendConstants2,
		s.ColorBlend.BlendConstants3,
	}
}

func (s *ShaderPipelineData) PatchControlPointsToVK() uint32 {
	if res, ok := StringVkPatchControlPoints[s.Tessellation.PatchControlPoints]; ok {
		return res
	}
	slog.Warn("invalid string for PatchControlPoints", "value", s.Tessellation.PatchControlPoints)
	return 3
}

// TODO:  This and the BackStencilOpStateToVK are duplicates because of a bad
// structure setup, please fix later
func (s *ShaderPipelineData) FrontStencilOpStateToVK() vk.StencilOpState {
	return vk.StencilOpState{
		FailOp:      stencilOpToVK(s.DepthStencil.FrontFailOp),
		PassOp:      stencilOpToVK(s.DepthStencil.FrontPassOp),
		DepthFailOp: stencilOpToVK(s.DepthStencil.FrontDepthFailOp),
		CompareOp:   compareOpToVK(s.DepthStencil.FrontCompareOp),
		CompareMask: s.DepthStencil.FrontCompareMask,
		WriteMask:   s.DepthStencil.FrontWriteMask,
		Reference:   s.DepthStencil.FrontReference,
	}
}

func (s *ShaderPipelineData) BackStencilOpStateToVK() vk.StencilOpState {
	return vk.StencilOpState{
		FailOp:      stencilOpToVK(s.DepthStencil.BackFailOp),
		PassOp:      stencilOpToVK(s.DepthStencil.BackPassOp),
		DepthFailOp: stencilOpToVK(s.DepthStencil.BackDepthFailOp),
		CompareOp:   compareOpToVK(s.DepthStencil.BackCompareOp),
		CompareMask: s.DepthStencil.BackCompareMask,
		WriteMask:   s.DepthStencil.BackWriteMask,
		Reference:   s.DepthStencil.BackReference,
	}
}

func (s *ShaderPipelineData) PipelineCreateFlagsToVK() vk.PipelineCreateFlags {
	mask := vk.PipelineCreateFlagBits(0)
	for i := range s.GraphicsPipeline.PipelineCreateFlags {
		mask |= StringVkPipelineCreateFlagBits[s.GraphicsPipeline.PipelineCreateFlags[i]]
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
		LineWidth:               s.Rasterization.LineWidth,
		CullMode:                vk.CullModeFlags(s.CullModeToVK()),
		FrontFace:               s.FrontFaceToVK(),
		DepthBiasEnable:         s.DepthBiasEnableToVK(),
		DepthBiasConstantFactor: s.Rasterization.DepthBiasConstantFactor,
		DepthBiasClamp:          s.Rasterization.DepthBiasClamp,
		DepthBiasSlopeFactor:    s.Rasterization.DepthBiasSlopeFactor,
	}
	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		Flags:                 0, // PipelineMultisampleStateCreateFlags
		SampleShadingEnable:   s.SampleShadingEnableToVK(),
		RasterizationSamples:  s.RasterizationSamplesToVK(),
		MinSampleShading:      s.Multisample.MinSampleShading,
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
		DepthCompareOp:        compareOpToVK(s.DepthStencil.DepthCompareOp),
		DepthBoundsTestEnable: s.DepthBoundsTestEnableToVK(),
		StencilTestEnable:     s.StencilTestEnableToVK(),
		MinDepthBounds:        s.DepthStencil.MinDepthBounds,
		MaxDepthBounds:        s.DepthStencil.MaxDepthBounds,
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
		Subpass:             s.GraphicsPipeline.SubpassCount,
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
