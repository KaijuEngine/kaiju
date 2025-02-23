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
	switch val {
	case "Zero":
		return vk.BlendFactorZero
	case "One":
		return vk.BlendFactorOne
	case "SrcColor":
		return vk.BlendFactorSrcColor
	case "OneMinusSrcColor":
		return vk.BlendFactorOneMinusSrcColor
	case "DstColor":
		return vk.BlendFactorDstColor
	case "OneMinusDstColor":
		return vk.BlendFactorOneMinusDstColor
	case "OneMinusSrcAlpha":
		return vk.BlendFactorOneMinusSrcAlpha
	case "DstAlpha":
		return vk.BlendFactorDstAlpha
	case "OneMinusDstAlpha":
		return vk.BlendFactorOneMinusDstAlpha
	case "ConstantColor":
		return vk.BlendFactorConstantColor
	case "OneMinusConstantColor":
		return vk.BlendFactorOneMinusConstantColor
	case "ConstantAlpha":
		return vk.BlendFactorConstantAlpha
	case "OneMinusConstantAlpha":
		return vk.BlendFactorOneMinusConstantAlpha
	case "SrcAlphaSaturate":
		return vk.BlendFactorSrcAlphaSaturate
	case "Src1Color":
		return vk.BlendFactorSrc1Color
	case "OneMinusSrc1Color":
		return vk.BlendFactorOneMinusSrc1Color
	case "Src1Alpha":
		return vk.BlendFactorSrc1Alpha
	case "OneMinusSrc1Alpha":
		return vk.BlendFactorOneMinusSrc1Alpha
	case "SrcAlpha":
		fallthrough
	default:
		return vk.BlendFactorSrcAlpha
	}
}

func blendOpToVK(val string) vk.BlendOp {
	switch val {
	case "Subtract":
		return vk.BlendOpSubtract
	case "ReverseSubtract":
		return vk.BlendOpReverseSubtract
	case "Min":
		return vk.BlendOpMin
	case "Max":
		return vk.BlendOpMax
	case "Zero":
		return vk.BlendOpZero
	case "Src":
		return vk.BlendOpSrc
	case "Dst":
		return vk.BlendOpDst
	case "SrcOver":
		return vk.BlendOpSrcOver
	case "DstOver":
		return vk.BlendOpDstOver
	case "SrcIn":
		return vk.BlendOpSrcIn
	case "DstIn":
		return vk.BlendOpDstIn
	case "SrcOut":
		return vk.BlendOpSrcOut
	case "DstOut":
		return vk.BlendOpDstOut
	case "SrcAtop":
		return vk.BlendOpSrcAtop
	case "DstAtop":
		return vk.BlendOpDstAtop
	case "Xor":
		return vk.BlendOpXor
	case "Multiply":
		return vk.BlendOpMultiply
	case "Screen":
		return vk.BlendOpScreen
	case "Overlay":
		return vk.BlendOpOverlay
	case "Darken":
		return vk.BlendOpDarken
	case "Lighten":
		return vk.BlendOpLighten
	case "Colordodge":
		return vk.BlendOpColordodge
	case "Colorburn":
		return vk.BlendOpColorburn
	case "Hardlight":
		return vk.BlendOpHardlight
	case "Softlight":
		return vk.BlendOpSoftlight
	case "Difference":
		return vk.BlendOpDifference
	case "Exclusion":
		return vk.BlendOpExclusion
	case "Invert":
		return vk.BlendOpInvert
	case "InvertRgb":
		return vk.BlendOpInvertRgb
	case "Lineardodge":
		return vk.BlendOpLineardodge
	case "Linearburn":
		return vk.BlendOpLinearburn
	case "Vividlight":
		return vk.BlendOpVividlight
	case "Linearlight":
		return vk.BlendOpLinearlight
	case "Pinlight":
		return vk.BlendOpPinlight
	case "Hardmix":
		return vk.BlendOpHardmix
	case "HslHue":
		return vk.BlendOpHslHue
	case "HslSaturation":
		return vk.BlendOpHslSaturation
	case "HslColor":
		return vk.BlendOpHslColor
	case "HslLuminosity":
		return vk.BlendOpHslLuminosity
	case "Plus":
		return vk.BlendOpPlus
	case "PlusClamped":
		return vk.BlendOpPlusClamped
	case "PlusClampedAlpha":
		return vk.BlendOpPlusClampedAlpha
	case "PlusDarker":
		return vk.BlendOpPlusDarker
	case "Minus":
		return vk.BlendOpMinus
	case "MinusClamped":
		return vk.BlendOpMinusClamped
	case "Contrast":
		return vk.BlendOpContrast
	case "InvertOvg":
		return vk.BlendOpInvertOvg
	case "Red":
		return vk.BlendOpRed
	case "Green":
		return vk.BlendOpBlue
	case "Blue":
		return vk.BlendOpBlue
	case "Add":
		fallthrough
	default:
		return vk.BlendOpAdd
	}
}

func (s *ShaderPipelineData) TopologyToVK() vk.PrimitiveTopology {
	switch s.Topology {
	case "Points":
		return vk.PrimitiveTopologyPointList
	case "Lines":
		return vk.PrimitiveTopologyLineList
	case "Patches":
		return vk.PrimitiveTopologyPatchList
	case "Triangles":
		fallthrough
	default:
		return vk.PrimitiveTopologyTriangleList
	}
}

func (s *ShaderPipelineData) PolygonModeToVK() vk.PolygonMode {
	switch s.PolygonMode {
	case "Line":
		return vk.PolygonModeLine
	case "Point":
		return vk.PolygonModePoint
	case "Fill":
		fallthrough
	default:
		return vk.PolygonModeFill
	}
}

func (s *ShaderPipelineData) CullModeToVK() vk.CullModeFlagBits {
	switch s.CullMode {
	case "Front":
		return vk.CullModeBackBit
	case "None":
		return vk.CullModeNone
	case "Back":
		fallthrough
	default:
		return vk.CullModeFrontBit
	}
}

func (s *ShaderPipelineData) FrontFaceToVK() vk.FrontFace {
	switch s.FrontFace {
	case "CounterClockwise":
		return vk.FrontFaceCounterClockwise
	case "Clockwise":
		fallthrough
	default:
		return vk.FrontFaceClockwise
	}
}

func (s *ShaderPipelineData) RasterizationSamplesToVK() vk.SampleCountFlagBits {
	switch s.RasterizationSamples {
	case "2Bit":
		return vk.SampleCount2Bit
	case "4Bit":
		return vk.SampleCount4Bit
	case "8Bit":
		return vk.SampleCount8Bit
	case "16Bit":
		return vk.SampleCount16Bit
	case "32Bit":
		return vk.SampleCount32Bit
	case "64Bit":
		return vk.SampleCount64Bit
	case "1Bit":
		fallthrough
	default:
		return vk.SampleCount1Bit
	}
}

func (s *ShaderPipelineData) LogicOpToVK() vk.LogicOp {
	switch s.LogicOp {
	case "Clear":
		return vk.LogicOpClear
	case "And":
		return vk.LogicOpAnd
	case "AndReverse":
		return vk.LogicOpAndReverse
	case "Copy":
		return vk.LogicOpCopy
	case "AndInverted":
		return vk.LogicOpAndInverted
	case "NoOp":
		return vk.LogicOpNoOp
	case "Xor":
		return vk.LogicOpXor
	case "Or":
		return vk.LogicOpOr
	case "Nor":
		return vk.LogicOpNor
	case "Equivalent":
		return vk.LogicOpEquivalent
	case "Invert":
		return vk.LogicOpInvert
	case "OrReverse":
		return vk.LogicOpOrReverse
	case "CopyInverted":
		return vk.LogicOpCopyInverted
	case "OrInverted":
		return vk.LogicOpOrInverted
	case "Nand":
		return vk.LogicOpNand
	case "Set":
		return vk.LogicOpSet
	default:
		return vk.LogicOpCopy
	}
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
	switch val {
	case "Never":
		return vk.CompareOpNever
	case "Equal":
		return vk.CompareOpEqual
	case "LessOrEqual":
		return vk.CompareOpLessOrEqual
	case "Greater":
		return vk.CompareOpGreater
	case "NotEqual":
		return vk.CompareOpNotEqual
	case "GreaterOrEqual":
		return vk.CompareOpGreaterOrEqual
	case "Always":
		return vk.CompareOpAlways
	case "Less":
		fallthrough
	default:
		return vk.CompareOpLess
	}
}

func stencilOpToVK(val string) vk.StencilOp {
	switch val {
	case "Zero":
		return vk.StencilOpZero
	case "Replace":
		return vk.StencilOpReplace
	case "IncrementAndClamp":
		return vk.StencilOpIncrementAndClamp
	case "DecrementAndClamp":
		return vk.StencilOpDecrementAndClamp
	case "Invert":
		return vk.StencilOpInvert
	case "IncrementAndWrap":
		return vk.StencilOpIncrementAndWrap
	case "DecrementAndWrap":
		return vk.StencilOpDecrementAndWrap
	case "Keep":
		fallthrough
	default:
		return vk.StencilOpKeep
	}
}

func (s *ShaderPipelineData) PatchControlPointsToVK() uint32 {
	switch s.PatchControlPoints {
	case "Quads":
		return 4
	case "Lines":
		return 2
	case "Triangles":
		fallthrough
	default:
		return 3
	}
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
