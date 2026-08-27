/******************************************************************************/
/* shader_pipeline.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"

	"kaijuengine.com/klib"
	"kaijuengine.com/rendering/gpu_enums"
	"kaijuengine.com/rendering/gpu_types"
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
	Topology         gpu_enums.PrimitiveTopology
	PrimitiveRestart bool
}

type ShaderPipelinePipelineRasterizationCompiled struct {
	DepthClampEnable        bool
	DiscardEnable           bool
	PolygonMode             gpu_enums.PolygonMode
	CullMode                gpu_enums.CullModeFlags
	FrontFace               gpu_enums.FrontFace
	DepthBiasEnable         bool
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
}

type ShaderPipelinePipelineMultisampleCompiled struct {
	RasterizationSamples  gpu_types.SampleCountFlags
	SampleShadingEnable   bool
	MinSampleShading      float32
	AlphaToCoverageEnable bool
	AlphaToOneEnable      bool
}

type ShaderPipelineColorBlendCompiled struct {
	LogicOpEnable  bool
	LogicOp        gpu_enums.LogicOp
	BlendConstants [4]float32
}

type ShaderPipelineDepthStencilCompiled struct {
	DepthTestEnable       bool
	DepthWriteEnable      bool
	DepthCompareOp        gpu_enums.CompareOp
	DepthBoundsTestEnable bool
	StencilTestEnable     bool
	Front                 gpu_enums.StencilOpState
	Back                  gpu_enums.StencilOpState
	MinDepthBounds        float32
	MaxDepthBounds        float32
}

type ShaderPipelineTessellationCompiled struct {
	PatchControlPoints uint32
}

type ShaderPipelineGraphicsPipelineCompiled struct {
	Subpass             uint32
	PipelineCreateFlags gpu_enums.PipelineCreateFlags
}

type ShaderPipelinePushConstantCompiled struct {
	Size       uint32
	StageFlags gpu_enums.ShaderStageFlags
}

type ShaderPipelineColorBlendAttachmentsCompiled struct {
	BlendEnable         bool
	SrcColorBlendFactor gpu_enums.BlendFactor
	DstColorBlendFactor gpu_enums.BlendFactor
	ColorBlendOp        gpu_enums.BlendOp
	SrcAlphaBlendFactor gpu_enums.BlendFactor
	DstAlphaBlendFactor gpu_enums.BlendFactor
	AlphaBlendOp        gpu_enums.BlendOp
	ColorWriteMask      gpu_enums.ColorComponentFlags
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
			Front: gpu_enums.StencilOpState{
				FailOp:      stencilOpToGPU(d.DepthStencil.FrontFailOp),
				PassOp:      stencilOpToGPU(d.DepthStencil.FrontPassOp),
				DepthFailOp: stencilOpToGPU(d.DepthStencil.FrontDepthFailOp),
				CompareOp:   compareOpToGPU(d.DepthStencil.FrontCompareOp),
				CompareMask: d.DepthStencil.FrontCompareMask,
				WriteMask:   d.DepthStencil.FrontWriteMask,
				Reference:   d.DepthStencil.FrontReference,
			},
			Back: gpu_enums.StencilOpState{
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

func (a *ShaderPipelineColorBlendAttachments) SrcColorBlendFactorToGPU() gpu_enums.BlendFactor {
	return blendFactorToGPU(a.SrcColorBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) DstColorBlendFactorToGPU() gpu_enums.BlendFactor {
	return blendFactorToGPU(a.DstColorBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) ColorBlendOpToGPU() gpu_enums.BlendOp {
	return blendOpToGPU(a.ColorBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) SrcAlphaBlendFactorToGPU() gpu_enums.BlendFactor {
	return blendFactorToGPU(a.SrcAlphaBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) DstAlphaBlendFactorToGPU() gpu_enums.BlendFactor {
	return blendFactorToGPU(a.DstAlphaBlendFactor)
}

func (a *ShaderPipelineColorBlendAttachments) AlphaBlendOpToGPU() gpu_enums.BlendOp {
	return blendOpToGPU(a.AlphaBlendOp)
}

func (a *ShaderPipelineColorBlendAttachments) ColorWriteMaskToGPU() gpu_enums.ColorComponentFlags {
	mask := gpu_enums.ColorComponentFlags(0)
	for i := range a.ColorWriteMask {
		mask |= gpu_enums.ColorComponentFlags(StringVkColorComponentFlagBits[a.ColorWriteMask[i]])
	}
	return mask
}

func blendFactorToGPU(val string) gpu_enums.BlendFactor {
	if res, ok := StringVkBlendFactor[val]; ok {
		return gpu_enums.BlendFactor(res)
	} else if val != "" {
		slog.Warn("invalid string for vkBlendFactor", "value", val)
	}
	return 0
}

func blendOpToGPU(val string) gpu_enums.BlendOp {
	if res, ok := StringVkBlendOp[val]; ok {
		return gpu_enums.BlendOp(res)
	} else if val != "" {
		slog.Warn("invalid string for vkBlendOp", "value", val)
	}
	return 0
}

func compareOpToGPU(val string) gpu_enums.CompareOp {
	if res, ok := StringVkCompareOp[val]; ok {
		return gpu_enums.CompareOp(res)
	} else if val != "" {
		slog.Warn("invalid string for vkCompareOp", "value", val)
	}
	return 0
}

func stencilOpToGPU(val string) gpu_enums.StencilOp {
	if res, ok := StringVkStencilOp[val]; ok {
		return gpu_enums.StencilOp(res)
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

func (s *ShaderPipelineInputAssembly) TopologyToGPU() gpu_enums.PrimitiveTopology {
	if res, ok := StringVkPrimitiveTopology[s.Topology]; ok {
		return gpu_enums.PrimitiveTopology(res)
	} else if s.Topology != "" {
		slog.Warn("invalid string for vkPrimitiveTopology", "value", s.Topology)
	}
	return gpu_enums.PrimitiveTopologyTriangleList
}

func (s *ShaderPipelinePipelineRasterization) PolygonModeToGPU() gpu_enums.PolygonMode {
	if res, ok := StringVkPolygonMode[s.PolygonMode]; ok {
		return gpu_enums.PolygonMode(res)
	} else if s.PolygonMode != "" {
		slog.Warn("invalid string for vkPolygonMode", "value", s.PolygonMode)
	}
	return gpu_enums.PolygonModeFill
}

func (s *ShaderPipelinePipelineRasterization) CullModeToGPU() gpu_enums.CullModeFlags {
	if res, ok := StringVkCullModeFlagBits[s.CullMode]; ok {
		return gpu_enums.CullModeFlags(res)
	} else if s.CullMode != "" {
		slog.Warn("invalid string for vkCullModeFlagBits", "value", s.CullMode)
	}
	return gpu_enums.CullModeFrontBit
}

func (s *ShaderPipelinePipelineRasterization) FrontFaceToGPU() gpu_enums.FrontFace {
	if res, ok := StringVkFrontFace[s.FrontFace]; ok {
		return gpu_enums.FrontFace(res)
	} else if s.FrontFace != "" {
		slog.Warn("invalid string for vkFrontFace", "value", s.FrontFace)
	}
	return gpu_enums.FrontFaceClockwise
}

func (s *ShaderPipelinePipelineMultisample) RasterizationSamplesToGPU(device *GPUPhysicalDevice) gpu_types.SampleCountFlags {
	return sampleCountToGpu(s.RasterizationSamples, device)
}

func (s *ShaderPipelineColorBlend) LogicOpToGPU() gpu_enums.LogicOp {
	if res, ok := StringVkLogicOp[s.LogicOp]; ok {
		return gpu_enums.LogicOp(res)
	} else if s.LogicOp != "" {
		slog.Warn("invalid string for vkLogicOp", "value", s.LogicOp)
	}
	return gpu_enums.LogicOpCopy
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
func (s *ShaderPipelineData) FrontStencilOpStateToGPU() gpu_enums.StencilOpState {
	return gpu_enums.StencilOpState{
		FailOp:      stencilOpToGPU(s.DepthStencil.FrontFailOp),
		PassOp:      stencilOpToGPU(s.DepthStencil.FrontPassOp),
		DepthFailOp: stencilOpToGPU(s.DepthStencil.FrontDepthFailOp),
		CompareOp:   compareOpToGPU(s.DepthStencil.FrontCompareOp),
		CompareMask: s.DepthStencil.FrontCompareMask,
		WriteMask:   s.DepthStencil.FrontWriteMask,
		Reference:   s.DepthStencil.FrontReference,
	}
}

func (s *ShaderPipelineData) BackStencilOpStateToGPU() gpu_enums.StencilOpState {
	return gpu_enums.StencilOpState{
		FailOp:      stencilOpToGPU(s.DepthStencil.BackFailOp),
		PassOp:      stencilOpToGPU(s.DepthStencil.BackPassOp),
		DepthFailOp: stencilOpToGPU(s.DepthStencil.BackDepthFailOp),
		CompareOp:   compareOpToGPU(s.DepthStencil.BackCompareOp),
		CompareMask: s.DepthStencil.BackCompareMask,
		WriteMask:   s.DepthStencil.BackWriteMask,
		Reference:   s.DepthStencil.BackReference,
	}
}

func (s *ShaderPipelineGraphicsPipeline) PipelineCreateFlagsToGPU() gpu_enums.PipelineCreateFlags {
	mask := gpu_enums.PipelineCreateFlags(0)
	for i := range s.PipelineCreateFlags {
		mask |= gpu_enums.PipelineCreateFlags(StringVkPipelineCreateFlagBits[s.PipelineCreateFlags[i]])
	}
	return mask
}

func (s *ShaderPipelinePushConstant) ShaderStageFlagsToGPU() gpu_enums.ShaderStageFlags {
	mask := gpu_enums.ShaderStageFlags(0)
	for i := range s.StageFlags {
		mask |= gpu_enums.ShaderStageFlags(StringVkShaderStageFlagBits[s.StageFlags[i]])
	}
	return mask
}
