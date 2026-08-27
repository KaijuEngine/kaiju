/******************************************************************************/
/* gpu_enums.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gpu_enums

type PrimitiveTopology int32
type PolygonMode int32
type CullModeFlags uint32
type FrontFace int32
type LogicOp int32
type CompareOp int32
type StencilOp int32
type BlendFactor int32
type BlendOp int32
type ColorComponentFlags uint32
type PipelineCreateFlags uint32
type ShaderStageFlags uint32
type PipelineBindPoint int32
type DependencyFlags uint32

type StencilOpState struct {
	FailOp      StencilOp
	PassOp      StencilOp
	DepthFailOp StencilOp
	CompareOp   CompareOp
	CompareMask uint32
	WriteMask   uint32
	Reference   uint32
}

const (
	PrimitiveTopologyPointList                  PrimitiveTopology = 0
	PrimitiveTopologyLineList                   PrimitiveTopology = 1
	PrimitiveTopologyLineStrip                  PrimitiveTopology = 2
	PrimitiveTopologyTriangleList               PrimitiveTopology = 3
	PrimitiveTopologyTriangleStrip              PrimitiveTopology = 4
	PrimitiveTopologyTriangleFan                PrimitiveTopology = 5
	PrimitiveTopologyLineListWithAdjacency      PrimitiveTopology = 6
	PrimitiveTopologyLineStripWithAdjacency     PrimitiveTopology = 7
	PrimitiveTopologyTriangleListWithAdjacency  PrimitiveTopology = 8
	PrimitiveTopologyTriangleStripWithAdjacency PrimitiveTopology = 9
	PrimitiveTopologyPatchList                  PrimitiveTopology = 10
)

const (
	PolygonModeFill  PolygonMode = 0
	PolygonModeLine  PolygonMode = 1
	PolygonModePoint PolygonMode = 2
)

const (
	CullModeNone     CullModeFlags = 0
	CullModeFrontBit CullModeFlags = 1
	CullModeBackBit  CullModeFlags = 2
	CullModeAll      CullModeFlags = CullModeFrontBit | CullModeBackBit
)

const (
	FrontFaceCounterClockwise FrontFace = 0
	FrontFaceClockwise        FrontFace = 1
)

const (
	LogicOpClear        LogicOp = 0
	LogicOpAnd          LogicOp = 1
	LogicOpAndReverse   LogicOp = 2
	LogicOpCopy         LogicOp = 3
	LogicOpAndInverted  LogicOp = 4
	LogicOpNoOp         LogicOp = 5
	LogicOpXor          LogicOp = 6
	LogicOpOr           LogicOp = 7
	LogicOpNor          LogicOp = 8
	LogicOpEquivalent   LogicOp = 9
	LogicOpInvert       LogicOp = 10
	LogicOpOrReverse    LogicOp = 11
	LogicOpCopyInverted LogicOp = 12
	LogicOpOrInverted   LogicOp = 13
	LogicOpNand         LogicOp = 14
	LogicOpSet          LogicOp = 15
)

const (
	CompareOpNever          CompareOp = 0
	CompareOpLess           CompareOp = 1
	CompareOpEqual          CompareOp = 2
	CompareOpLessOrEqual    CompareOp = 3
	CompareOpGreater        CompareOp = 4
	CompareOpNotEqual       CompareOp = 5
	CompareOpGreaterOrEqual CompareOp = 6
	CompareOpAlways         CompareOp = 7
)

const (
	StencilOpKeep              StencilOp = 0
	StencilOpZero              StencilOp = 1
	StencilOpReplace           StencilOp = 2
	StencilOpIncrementAndClamp StencilOp = 3
	StencilOpDecrementAndClamp StencilOp = 4
	StencilOpInvert            StencilOp = 5
	StencilOpIncrementAndWrap  StencilOp = 6
	StencilOpDecrementAndWrap  StencilOp = 7
)

const (
	BlendFactorZero                  BlendFactor = 0
	BlendFactorOne                   BlendFactor = 1
	BlendFactorSrcColor              BlendFactor = 2
	BlendFactorOneMinusSrcColor      BlendFactor = 3
	BlendFactorDstColor              BlendFactor = 4
	BlendFactorOneMinusDstColor      BlendFactor = 5
	BlendFactorSrcAlpha              BlendFactor = 6
	BlendFactorOneMinusSrcAlpha      BlendFactor = 7
	BlendFactorDstAlpha              BlendFactor = 8
	BlendFactorOneMinusDstAlpha      BlendFactor = 9
	BlendFactorConstantColor         BlendFactor = 10
	BlendFactorOneMinusConstantColor BlendFactor = 11
	BlendFactorConstantAlpha         BlendFactor = 12
	BlendFactorOneMinusConstantAlpha BlendFactor = 13
	BlendFactorSrcAlphaSaturate      BlendFactor = 14
	BlendFactorSrc1Color             BlendFactor = 15
	BlendFactorOneMinusSrc1Color     BlendFactor = 16
	BlendFactorSrc1Alpha             BlendFactor = 17
	BlendFactorOneMinusSrc1Alpha     BlendFactor = 18
)

const (
	BlendOpAdd              BlendOp = 0
	BlendOpSubtract         BlendOp = 1
	BlendOpReverseSubtract  BlendOp = 2
	BlendOpMin              BlendOp = 3
	BlendOpMax              BlendOp = 4
	BlendOpZero             BlendOp = 1000148000
	BlendOpSrc              BlendOp = 1000148001
	BlendOpDst              BlendOp = 1000148002
	BlendOpSrcOver          BlendOp = 1000148003
	BlendOpDstOver          BlendOp = 1000148004
	BlendOpSrcIn            BlendOp = 1000148005
	BlendOpDstIn            BlendOp = 1000148006
	BlendOpSrcOut           BlendOp = 1000148007
	BlendOpDstOut           BlendOp = 1000148008
	BlendOpSrcAtop          BlendOp = 1000148009
	BlendOpDstAtop          BlendOp = 1000148010
	BlendOpXor              BlendOp = 1000148011
	BlendOpMultiply         BlendOp = 1000148012
	BlendOpScreen           BlendOp = 1000148013
	BlendOpOverlay          BlendOp = 1000148014
	BlendOpDarken           BlendOp = 1000148015
	BlendOpLighten          BlendOp = 1000148016
	BlendOpColordodge       BlendOp = 1000148017
	BlendOpColorburn        BlendOp = 1000148018
	BlendOpHardlight        BlendOp = 1000148019
	BlendOpSoftlight        BlendOp = 1000148020
	BlendOpDifference       BlendOp = 1000148021
	BlendOpExclusion        BlendOp = 1000148022
	BlendOpInvert           BlendOp = 1000148023
	BlendOpInvertRgb        BlendOp = 1000148024
	BlendOpLineardodge      BlendOp = 1000148025
	BlendOpLinearburn       BlendOp = 1000148026
	BlendOpVividlight       BlendOp = 1000148027
	BlendOpLinearlight      BlendOp = 1000148028
	BlendOpPinlight         BlendOp = 1000148029
	BlendOpHardmix          BlendOp = 1000148030
	BlendOpHslHue           BlendOp = 1000148031
	BlendOpHslSaturation    BlendOp = 1000148032
	BlendOpHslColor         BlendOp = 1000148033
	BlendOpHslLuminosity    BlendOp = 1000148034
	BlendOpPlus             BlendOp = 1000148035
	BlendOpPlusClamped      BlendOp = 1000148036
	BlendOpPlusClampedAlpha BlendOp = 1000148037
	BlendOpPlusDarker       BlendOp = 1000148038
	BlendOpMinus            BlendOp = 1000148039
	BlendOpMinusClamped     BlendOp = 1000148040
	BlendOpContrast         BlendOp = 1000148041
	BlendOpInvertOvg        BlendOp = 1000148042
	BlendOpRed              BlendOp = 1000148043
	BlendOpGreen            BlendOp = 1000148044
	BlendOpBlue             BlendOp = 1000148045
)

const (
	ColorComponentRBit ColorComponentFlags = 1
	ColorComponentGBit ColorComponentFlags = 2
	ColorComponentBBit ColorComponentFlags = 4
	ColorComponentABit ColorComponentFlags = 8
)

const (
	PipelineCreateDisableOptimizationBit      PipelineCreateFlags = 1
	PipelineCreateAllowDerivativesBit         PipelineCreateFlags = 2
	PipelineCreateDerivativeBit               PipelineCreateFlags = 4
	PipelineCreateViewIndexFromDeviceIndexBit PipelineCreateFlags = 8
	PipelineCreateDispatchBase                PipelineCreateFlags = 16
	PipelineCreateDeferCompileBitNvx          PipelineCreateFlags = 32
)

const (
	ShaderStageVertexBit                 ShaderStageFlags = 1
	ShaderStageTessellationControlBit    ShaderStageFlags = 2
	ShaderStageTessellationEvaluationBit ShaderStageFlags = 4
	ShaderStageGeometryBit               ShaderStageFlags = 8
	ShaderStageFragmentBit               ShaderStageFlags = 16
	ShaderStageComputeBit                ShaderStageFlags = 32
	ShaderStageAllGraphics               ShaderStageFlags = 31
	ShaderStageAll                       ShaderStageFlags = 2147483647
	ShaderStageTaskBitNv                 ShaderStageFlags = 64
	ShaderStageMeshBitNv                 ShaderStageFlags = 128
	ShaderStageRaygenBitNvx              ShaderStageFlags = 256
	ShaderStageAnyHitBitNvx              ShaderStageFlags = 512
	ShaderStageClosestHitBitNvx          ShaderStageFlags = 1024
	ShaderStageMissBitNvx                ShaderStageFlags = 2048
	ShaderStageIntersectionBitNvx        ShaderStageFlags = 4096
	ShaderStageCallableBitNvx            ShaderStageFlags = 8192
)

const (
	PipelineBindPointGraphics      PipelineBindPoint = 0
	PipelineBindPointCompute       PipelineBindPoint = 1
	PipelineBindPointRaytracingNvx PipelineBindPoint = 1000165000
)

const (
	DependencyByRegionBit    DependencyFlags = 1
	DependencyViewLocalBit   DependencyFlags = 2
	DependencyDeviceGroupBit DependencyFlags = 4
)
