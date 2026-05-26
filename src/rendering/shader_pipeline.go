/******************************************************************************/
/* shader_pipeline.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"

	"kaijuengine.com/klib"
)

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
	PushConstant          ShaderPipelinePushConstant
}

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
	Subpass             uint32
	PipelineCreateFlags []string `options:"StringVkPipelineCreateFlagBits"`
}

type ShaderPipelinePushConstant struct {
	Size       uint32
	StageFlags []string `options:"StringVkAccessFlagBits"`
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

type ShaderPipelineDataCompiled struct {
	Name                  string
	InputAssembly         ShaderPipelineInputAssemblyCompiled
	Rasterization         ShaderPipelinePipelineRasterizationCompiled
	Multisample           ShaderPipelinePipelineMultisampleCompiled
	ColorBlend            ShaderPipelineColorBlendCompiled
	ColorBlendAttachments []ShaderPipelineColorBlendAttachmentsCompiled
	DepthStencil          ShaderPipelineDepthStencilCompiled
	Tessellation          ShaderPipelineTessellationCompiled
	GraphicsPipeline      ShaderPipelineGraphicsPipelineCompiled
	PushConstant          ShaderPipelinePushConstantCompiled
}

type ShaderPipelineInputAssemblyCompiled struct {
	Topology         GPUPrimitiveTopology
	PrimitiveRestart bool
}

type ShaderPipelinePipelineRasterizationCompiled struct {
	DepthClampEnable        bool
	DiscardEnable           bool
	PolygonMode             GPUPolygonMode
	CullMode                GPUCullModeFlags
	FrontFace               GPUFrontFace
	DepthBiasEnable         bool
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
}

type ShaderPipelinePipelineMultisampleCompiled struct {
	RasterizationSamples  GPUSampleCountFlags
	SampleShadingEnable   bool
	MinSampleShading      float32
	AlphaToCoverageEnable bool
	AlphaToOneEnable      bool
}

type ShaderPipelineColorBlendCompiled struct {
	LogicOpEnable  bool
	LogicOp        GPULogicOp
	BlendConstants [4]float32
}

type ShaderPipelineDepthStencilCompiled struct {
	DepthTestEnable       bool
	DepthWriteEnable      bool
	DepthCompareOp        GPUCompareOp
	DepthBoundsTestEnable bool
	StencilTestEnable     bool
	Front                 GPUStencilOpState
	Back                  GPUStencilOpState
	MinDepthBounds        float32
	MaxDepthBounds        float32
}

type ShaderPipelineTessellationCompiled struct {
	PatchControlPoints uint32
}

type ShaderPipelineGraphicsPipelineCompiled struct {
	Subpass             uint32
	PipelineCreateFlags GPUPipelineCreateFlags
}

type ShaderPipelinePushConstantCompiled struct {
	Size       uint32
	StageFlags GPUShaderStageFlags
}

type ShaderPipelineColorBlendAttachmentsCompiled struct {
	BlendEnable         bool
	SrcColorBlendFactor GPUBlendFactor
	DstColorBlendFactor GPUBlendFactor
	ColorBlendOp        GPUBlendOp
	SrcAlphaBlendFactor GPUBlendFactor
	DstAlphaBlendFactor GPUBlendFactor
	AlphaBlendOp        GPUBlendOp
	ColorWriteMask      GPUColorComponentFlags
}

func (d *ShaderPipelineData) Compile(device *GPUPhysicalDevice) ShaderPipelineDataCompiled {
	c := ShaderPipelineDataCompiled{
		Name: d.Name,
		InputAssembly: ShaderPipelineInputAssemblyCompiled{
			Topology:         d.InputAssembly.TopologyToGPU(),
			PrimitiveRestart: d.InputAssembly.PrimitiveRestart,
		},
		Rasterization: ShaderPipelinePipelineRasterizationCompiled{
			DepthClampEnable:        d.Rasterization.DepthClampEnable,
			DiscardEnable:           d.Rasterization.RasterizerDiscardEnable,
			PolygonMode:             d.Rasterization.PolygonModeToGPU(),
			CullMode:                d.Rasterization.CullModeToGPU(),
			FrontFace:               d.Rasterization.FrontFaceToGPU(),
			DepthBiasEnable:         d.Rasterization.DepthBiasEnable,
			DepthBiasConstantFactor: d.Rasterization.DepthBiasConstantFactor,
			DepthBiasClamp:          d.Rasterization.DepthBiasClamp,
			DepthBiasSlopeFactor:    d.Rasterization.DepthBiasSlopeFactor,
			LineWidth:               d.Rasterization.LineWidth,
		},
		Multisample: ShaderPipelinePipelineMultisampleCompiled{
			RasterizationSamples:  d.Multisample.RasterizationSamplesToGPU(device),
			SampleShadingEnable:   d.Multisample.SampleShadingEnable,
			MinSampleShading:      d.Multisample.MinSampleShading,
			AlphaToCoverageEnable: d.Multisample.AlphaToCoverageEnable,
			AlphaToOneEnable:      d.Multisample.AlphaToOneEnable,
		},
		ColorBlend: ShaderPipelineColorBlendCompiled{
			LogicOpEnable: d.ColorBlend.LogicOpEnable,
			LogicOp:       d.ColorBlend.LogicOpToGPU(),
			BlendConstants: [4]float32{
				d.ColorBlend.BlendConstants0,
				d.ColorBlend.BlendConstants1,
				d.ColorBlend.BlendConstants2,
				d.ColorBlend.BlendConstants3,
			},
		},
		ColorBlendAttachments: make([]ShaderPipelineColorBlendAttachmentsCompiled, len(d.ColorBlendAttachments)),
		DepthStencil: ShaderPipelineDepthStencilCompiled{
			DepthTestEnable:       d.DepthStencil.DepthTestEnable,
			DepthWriteEnable:      d.DepthStencil.DepthWriteEnable,
			DepthCompareOp:        compareOpToGPU(d.DepthStencil.DepthCompareOp),
			DepthBoundsTestEnable: d.DepthStencil.DepthBoundsTestEnable,
			StencilTestEnable:     d.DepthStencil.StencilTestEnable,
			Front: GPUStencilOpState{
				FailOp:      stencilOpToGPU(d.DepthStencil.FrontFailOp),
				PassOp:      stencilOpToGPU(d.DepthStencil.FrontPassOp),
				DepthFailOp: stencilOpToGPU(d.DepthStencil.FrontDepthFailOp),
				CompareOp:   compareOpToGPU(d.DepthStencil.FrontCompareOp),
				CompareMask: d.DepthStencil.FrontCompareMask,
				WriteMask:   d.DepthStencil.FrontWriteMask,
				Reference:   d.DepthStencil.FrontReference,
			},
			Back: GPUStencilOpState{
				FailOp:      stencilOpToGPU(d.DepthStencil.BackFailOp),
				PassOp:      stencilOpToGPU(d.DepthStencil.BackPassOp),
				DepthFailOp: stencilOpToGPU(d.DepthStencil.BackDepthFailOp),
				CompareOp:   compareOpToGPU(d.DepthStencil.BackCompareOp),
				CompareMask: d.DepthStencil.BackCompareMask,
				WriteMask:   d.DepthStencil.BackWriteMask,
				Reference:   d.DepthStencil.BackReference,
			},
			MinDepthBounds: d.DepthStencil.MinDepthBounds,
			MaxDepthBounds: d.DepthStencil.MaxDepthBounds,
		},
		Tessellation: ShaderPipelineTessellationCompiled{
			PatchControlPoints: d.Tessellation.PatchControlPointsToGPU(),
		},
		GraphicsPipeline: ShaderPipelineGraphicsPipelineCompiled{
			Subpass:             d.GraphicsPipeline.Subpass,
			PipelineCreateFlags: d.GraphicsPipeline.PipelineCreateFlagsToGPU(),
		},
		PushConstant: ShaderPipelinePushConstantCompiled{
			Size:       d.PushConstant.Size,
			StageFlags: d.PushConstant.ShaderStageFlagsToGPU(),
		},
	}
	for i := range d.ColorBlendAttachments {
		from := &d.ColorBlendAttachments[i]
		c.ColorBlendAttachments[i] = ShaderPipelineColorBlendAttachmentsCompiled{
			BlendEnable:         from.BlendEnable,
			SrcColorBlendFactor: from.SrcColorBlendFactorToGPU(),
			DstColorBlendFactor: from.DstColorBlendFactorToGPU(),
			ColorBlendOp:        from.ColorBlendOpToGPU(),
			SrcAlphaBlendFactor: from.SrcAlphaBlendFactorToGPU(),
			DstAlphaBlendFactor: from.DstAlphaBlendFactorToGPU(),
			AlphaBlendOp:        from.AlphaBlendOpToGPU(),
			ColorWriteMask:      from.ColorWriteMaskToGPU(),
		}
	}
	return c
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

func (a *ShaderPipelineColorBlendAttachments) SrcColorBlendFactorToGPU() GPUBlendFactor {
	return blendFactorToGPU(a.SrcColorBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) DstColorBlendFactorToGPU() GPUBlendFactor {
	return blendFactorToGPU(a.DstColorBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) ColorBlendOpToGPU() GPUBlendOp {
	return blendOpToGPU(a.ColorBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) SrcAlphaBlendFactorToGPU() GPUBlendFactor {
	return blendFactorToGPU(a.SrcAlphaBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) DstAlphaBlendFactorToGPU() GPUBlendFactor {
	return blendFactorToGPU(a.DstAlphaBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) AlphaBlendOpToGPU() GPUBlendOp {
	return blendOpToGPU(a.AlphaBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) ColorWriteMaskToGPU() GPUColorComponentFlags {
	mask := GPUColorComponentFlags(0)
	for i := range a.ColorWriteMask {
		mask |= GPUColorComponentFlags(StringVkColorComponentFlagBits[a.ColorWriteMask[i]])
	}
	return mask
}

func blendFactorToGPU(val string) GPUBlendFactor {
	if res, ok := StringVkBlendFactor[val]; ok {
		return GPUBlendFactor(res)
	} else if val != "" {
		slog.Warn("invalid string for vkBlendFactor", "value", val)
	}
	return 0
}

func blendOpToGPU(val string) GPUBlendOp {
	if res, ok := StringVkBlendOp[val]; ok {
		return GPUBlendOp(res)
	} else if val != "" {
		slog.Warn("invalid string for vkBlendOp", "value", val)
	}
	return 0
}

func compareOpToGPU(val string) GPUCompareOp {
	if res, ok := StringVkCompareOp[val]; ok {
		return GPUCompareOp(res)
	} else if val != "" {
		slog.Warn("invalid string for vkCompareOp", "value", val)
	}
	return 0
}

func stencilOpToGPU(val string) GPUStencilOp {
	if res, ok := StringVkStencilOp[val]; ok {
		return GPUStencilOp(res)
	} else if val != "" {
		slog.Warn("invalid string for vkStencilOpKeep", "value", val)
	}
	return 0
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

func (s *ShaderPipelineInputAssembly) TopologyToGPU() GPUPrimitiveTopology {
	if res, ok := StringVkPrimitiveTopology[s.Topology]; ok {
		return GPUPrimitiveTopology(res)
	} else if s.Topology != "" {
		slog.Warn("invalid string for vkPrimitiveTopology", "value", s.Topology)
	}
	return GPUPrimitiveTopologyTriangleList
}

func (s *ShaderPipelinePipelineRasterization) PolygonModeToGPU() GPUPolygonMode {
	if res, ok := StringVkPolygonMode[s.PolygonMode]; ok {
		return GPUPolygonMode(res)
	} else if s.PolygonMode != "" {
		slog.Warn("invalid string for vkPolygonMode", "value", s.PolygonMode)
	}
	return GPUPolygonModeFill
}

func (s *ShaderPipelinePipelineRasterization) CullModeToGPU() GPUCullModeFlags {
	if res, ok := StringVkCullModeFlagBits[s.CullMode]; ok {
		return GPUCullModeFlags(res)
	} else if s.CullMode != "" {
		slog.Warn("invalid string for vkCullModeFlagBits", "value", s.CullMode)
	}
	return GPUCullModeFrontBit
}

func (s *ShaderPipelinePipelineRasterization) FrontFaceToGPU() GPUFrontFace {
	if res, ok := StringVkFrontFace[s.FrontFace]; ok {
		return GPUFrontFace(res)
	} else if s.FrontFace != "" {
		slog.Warn("invalid string for vkFrontFace", "value", s.FrontFace)
	}
	return GPUFrontFaceClockwise
}

func (s *ShaderPipelinePipelineMultisample) RasterizationSamplesToGPU(device *GPUPhysicalDevice) GPUSampleCountFlags {
	return sampleCountToGpu(s.RasterizationSamples, device)
}

func (s *ShaderPipelineColorBlend) LogicOpToGPU() GPULogicOp {
	if res, ok := StringVkLogicOp[s.LogicOp]; ok {
		return GPULogicOp(res)
	} else if s.LogicOp != "" {
		slog.Warn("invalid string for vkLogicOp", "value", s.LogicOp)
	}
	return GPULogicOpCopy
}

func (s *ShaderPipelineData) BlendConstants() [4]float32 {
	return [4]float32{
		s.ColorBlend.BlendConstants0,
		s.ColorBlend.BlendConstants1,
		s.ColorBlend.BlendConstants2,
		s.ColorBlend.BlendConstants3,
	}
}

func (s *ShaderPipelineTessellation) PatchControlPointsToGPU() uint32 {
	if res, ok := StringVkPatchControlPoints[s.PatchControlPoints]; ok {
		return res
	} else if s.PatchControlPoints != "" {
		slog.Warn("invalid string for PatchControlPoints", "value", s.PatchControlPoints)
	}
	return 3
}

// TODO:  This and the BackStencilOpStateToGPU are duplicates because of a bad
// structure setup, please fix later
func (s *ShaderPipelineData) FrontStencilOpStateToGPU() GPUStencilOpState {
	return GPUStencilOpState{
		FailOp:      stencilOpToGPU(s.DepthStencil.FrontFailOp),
		PassOp:      stencilOpToGPU(s.DepthStencil.FrontPassOp),
		DepthFailOp: stencilOpToGPU(s.DepthStencil.FrontDepthFailOp),
		CompareOp:   compareOpToGPU(s.DepthStencil.FrontCompareOp),
		CompareMask: s.DepthStencil.FrontCompareMask,
		WriteMask:   s.DepthStencil.FrontWriteMask,
		Reference:   s.DepthStencil.FrontReference,
	}
}

func (s *ShaderPipelineData) BackStencilOpStateToGPU() GPUStencilOpState {
	return GPUStencilOpState{
		FailOp:      stencilOpToGPU(s.DepthStencil.BackFailOp),
		PassOp:      stencilOpToGPU(s.DepthStencil.BackPassOp),
		DepthFailOp: stencilOpToGPU(s.DepthStencil.BackDepthFailOp),
		CompareOp:   compareOpToGPU(s.DepthStencil.BackCompareOp),
		CompareMask: s.DepthStencil.BackCompareMask,
		WriteMask:   s.DepthStencil.BackWriteMask,
		Reference:   s.DepthStencil.BackReference,
	}
}

func (s *ShaderPipelineGraphicsPipeline) PipelineCreateFlagsToGPU() GPUPipelineCreateFlags {
	mask := GPUPipelineCreateFlags(0)
	for i := range s.PipelineCreateFlags {
		mask |= GPUPipelineCreateFlags(StringVkPipelineCreateFlagBits[s.PipelineCreateFlags[i]])
	}
	return mask
}

func (s *ShaderPipelinePushConstant) ShaderStageFlagsToGPU() GPUShaderStageFlags {
	mask := GPUShaderStageFlags(0)
	for i := range s.StageFlags {
		mask |= GPUShaderStageFlags(StringVkShaderStageFlagBits[s.StageFlags[i]])
	}
	return mask
}
