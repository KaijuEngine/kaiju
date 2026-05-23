/******************************************************************************/
/* gpu_enums.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

type GPUPrimitiveTopology int32
type GPUPolygonMode int32
type GPUCullModeFlags uint32
type GPUFrontFace int32
type GPULogicOp int32
type GPUCompareOp int32
type GPUStencilOp int32
type GPUBlendFactor int32
type GPUBlendOp int32
type GPUColorComponentFlags uint32
type GPUPipelineCreateFlags uint32
type GPUShaderStageFlags uint32
type GPUPipelineBindPoint int32
type GPUDependencyFlags uint32

type GPUStencilOpState struct {
	FailOp      GPUStencilOp
	PassOp      GPUStencilOp
	DepthFailOp GPUStencilOp
	CompareOp   GPUCompareOp
	CompareMask uint32
	WriteMask   uint32
	Reference   uint32
}

const (
	GPUPrimitiveTopologyPointList                  GPUPrimitiveTopology = 0
	GPUPrimitiveTopologyLineList                   GPUPrimitiveTopology = 1
	GPUPrimitiveTopologyLineStrip                  GPUPrimitiveTopology = 2
	GPUPrimitiveTopologyTriangleList               GPUPrimitiveTopology = 3
	GPUPrimitiveTopologyTriangleStrip              GPUPrimitiveTopology = 4
	GPUPrimitiveTopologyTriangleFan                GPUPrimitiveTopology = 5
	GPUPrimitiveTopologyLineListWithAdjacency      GPUPrimitiveTopology = 6
	GPUPrimitiveTopologyLineStripWithAdjacency     GPUPrimitiveTopology = 7
	GPUPrimitiveTopologyTriangleListWithAdjacency  GPUPrimitiveTopology = 8
	GPUPrimitiveTopologyTriangleStripWithAdjacency GPUPrimitiveTopology = 9
	GPUPrimitiveTopologyPatchList                  GPUPrimitiveTopology = 10
)

const (
	GPUPolygonModeFill  GPUPolygonMode = 0
	GPUPolygonModeLine  GPUPolygonMode = 1
	GPUPolygonModePoint GPUPolygonMode = 2
)

const (
	GPUCullModeNone     GPUCullModeFlags = 0
	GPUCullModeFrontBit GPUCullModeFlags = 1
	GPUCullModeBackBit  GPUCullModeFlags = 2
	GPUCullModeAll      GPUCullModeFlags = GPUCullModeFrontBit | GPUCullModeBackBit
)

const (
	GPUFrontFaceCounterClockwise GPUFrontFace = 0
	GPUFrontFaceClockwise        GPUFrontFace = 1
)

const (
	GPULogicOpClear        GPULogicOp = 0
	GPULogicOpAnd          GPULogicOp = 1
	GPULogicOpAndReverse   GPULogicOp = 2
	GPULogicOpCopy         GPULogicOp = 3
	GPULogicOpAndInverted  GPULogicOp = 4
	GPULogicOpNoOp         GPULogicOp = 5
	GPULogicOpXor          GPULogicOp = 6
	GPULogicOpOr           GPULogicOp = 7
	GPULogicOpNor          GPULogicOp = 8
	GPULogicOpEquivalent   GPULogicOp = 9
	GPULogicOpInvert       GPULogicOp = 10
	GPULogicOpOrReverse    GPULogicOp = 11
	GPULogicOpCopyInverted GPULogicOp = 12
	GPULogicOpOrInverted   GPULogicOp = 13
	GPULogicOpNand         GPULogicOp = 14
	GPULogicOpSet          GPULogicOp = 15
)

const (
	GPUCompareOpNever          GPUCompareOp = 0
	GPUCompareOpLess           GPUCompareOp = 1
	GPUCompareOpEqual          GPUCompareOp = 2
	GPUCompareOpLessOrEqual    GPUCompareOp = 3
	GPUCompareOpGreater        GPUCompareOp = 4
	GPUCompareOpNotEqual       GPUCompareOp = 5
	GPUCompareOpGreaterOrEqual GPUCompareOp = 6
	GPUCompareOpAlways         GPUCompareOp = 7
)

const (
	GPUStencilOpKeep              GPUStencilOp = 0
	GPUStencilOpZero              GPUStencilOp = 1
	GPUStencilOpReplace           GPUStencilOp = 2
	GPUStencilOpIncrementAndClamp GPUStencilOp = 3
	GPUStencilOpDecrementAndClamp GPUStencilOp = 4
	GPUStencilOpInvert            GPUStencilOp = 5
	GPUStencilOpIncrementAndWrap  GPUStencilOp = 6
	GPUStencilOpDecrementAndWrap  GPUStencilOp = 7
)

const (
	GPUBlendFactorZero                  GPUBlendFactor = 0
	GPUBlendFactorOne                   GPUBlendFactor = 1
	GPUBlendFactorSrcColor              GPUBlendFactor = 2
	GPUBlendFactorOneMinusSrcColor      GPUBlendFactor = 3
	GPUBlendFactorDstColor              GPUBlendFactor = 4
	GPUBlendFactorOneMinusDstColor      GPUBlendFactor = 5
	GPUBlendFactorSrcAlpha              GPUBlendFactor = 6
	GPUBlendFactorOneMinusSrcAlpha      GPUBlendFactor = 7
	GPUBlendFactorDstAlpha              GPUBlendFactor = 8
	GPUBlendFactorOneMinusDstAlpha      GPUBlendFactor = 9
	GPUBlendFactorConstantColor         GPUBlendFactor = 10
	GPUBlendFactorOneMinusConstantColor GPUBlendFactor = 11
	GPUBlendFactorConstantAlpha         GPUBlendFactor = 12
	GPUBlendFactorOneMinusConstantAlpha GPUBlendFactor = 13
	GPUBlendFactorSrcAlphaSaturate      GPUBlendFactor = 14
	GPUBlendFactorSrc1Color             GPUBlendFactor = 15
	GPUBlendFactorOneMinusSrc1Color     GPUBlendFactor = 16
	GPUBlendFactorSrc1Alpha             GPUBlendFactor = 17
	GPUBlendFactorOneMinusSrc1Alpha     GPUBlendFactor = 18
)

const (
	GPUBlendOpAdd              GPUBlendOp = 0
	GPUBlendOpSubtract         GPUBlendOp = 1
	GPUBlendOpReverseSubtract  GPUBlendOp = 2
	GPUBlendOpMin              GPUBlendOp = 3
	GPUBlendOpMax              GPUBlendOp = 4
	GPUBlendOpZero             GPUBlendOp = 1000148000
	GPUBlendOpSrc              GPUBlendOp = 1000148001
	GPUBlendOpDst              GPUBlendOp = 1000148002
	GPUBlendOpSrcOver          GPUBlendOp = 1000148003
	GPUBlendOpDstOver          GPUBlendOp = 1000148004
	GPUBlendOpSrcIn            GPUBlendOp = 1000148005
	GPUBlendOpDstIn            GPUBlendOp = 1000148006
	GPUBlendOpSrcOut           GPUBlendOp = 1000148007
	GPUBlendOpDstOut           GPUBlendOp = 1000148008
	GPUBlendOpSrcAtop          GPUBlendOp = 1000148009
	GPUBlendOpDstAtop          GPUBlendOp = 1000148010
	GPUBlendOpXor              GPUBlendOp = 1000148011
	GPUBlendOpMultiply         GPUBlendOp = 1000148012
	GPUBlendOpScreen           GPUBlendOp = 1000148013
	GPUBlendOpOverlay          GPUBlendOp = 1000148014
	GPUBlendOpDarken           GPUBlendOp = 1000148015
	GPUBlendOpLighten          GPUBlendOp = 1000148016
	GPUBlendOpColordodge       GPUBlendOp = 1000148017
	GPUBlendOpColorburn        GPUBlendOp = 1000148018
	GPUBlendOpHardlight        GPUBlendOp = 1000148019
	GPUBlendOpSoftlight        GPUBlendOp = 1000148020
	GPUBlendOpDifference       GPUBlendOp = 1000148021
	GPUBlendOpExclusion        GPUBlendOp = 1000148022
	GPUBlendOpInvert           GPUBlendOp = 1000148023
	GPUBlendOpInvertRgb        GPUBlendOp = 1000148024
	GPUBlendOpLineardodge      GPUBlendOp = 1000148025
	GPUBlendOpLinearburn       GPUBlendOp = 1000148026
	GPUBlendOpVividlight       GPUBlendOp = 1000148027
	GPUBlendOpLinearlight      GPUBlendOp = 1000148028
	GPUBlendOpPinlight         GPUBlendOp = 1000148029
	GPUBlendOpHardmix          GPUBlendOp = 1000148030
	GPUBlendOpHslHue           GPUBlendOp = 1000148031
	GPUBlendOpHslSaturation    GPUBlendOp = 1000148032
	GPUBlendOpHslColor         GPUBlendOp = 1000148033
	GPUBlendOpHslLuminosity    GPUBlendOp = 1000148034
	GPUBlendOpPlus             GPUBlendOp = 1000148035
	GPUBlendOpPlusClamped      GPUBlendOp = 1000148036
	GPUBlendOpPlusClampedAlpha GPUBlendOp = 1000148037
	GPUBlendOpPlusDarker       GPUBlendOp = 1000148038
	GPUBlendOpMinus            GPUBlendOp = 1000148039
	GPUBlendOpMinusClamped     GPUBlendOp = 1000148040
	GPUBlendOpContrast         GPUBlendOp = 1000148041
	GPUBlendOpInvertOvg        GPUBlendOp = 1000148042
	GPUBlendOpRed              GPUBlendOp = 1000148043
	GPUBlendOpGreen            GPUBlendOp = 1000148044
	GPUBlendOpBlue             GPUBlendOp = 1000148045
)

const (
	GPUColorComponentRBit GPUColorComponentFlags = 1
	GPUColorComponentGBit GPUColorComponentFlags = 2
	GPUColorComponentBBit GPUColorComponentFlags = 4
	GPUColorComponentABit GPUColorComponentFlags = 8
)

const (
	GPUPipelineCreateDisableOptimizationBit      GPUPipelineCreateFlags = 1
	GPUPipelineCreateAllowDerivativesBit         GPUPipelineCreateFlags = 2
	GPUPipelineCreateDerivativeBit               GPUPipelineCreateFlags = 4
	GPUPipelineCreateViewIndexFromDeviceIndexBit GPUPipelineCreateFlags = 8
	GPUPipelineCreateDispatchBase                GPUPipelineCreateFlags = 16
	GPUPipelineCreateDeferCompileBitNvx          GPUPipelineCreateFlags = 32
)

const (
	GPUShaderStageVertexBit                 GPUShaderStageFlags = 1
	GPUShaderStageTessellationControlBit    GPUShaderStageFlags = 2
	GPUShaderStageTessellationEvaluationBit GPUShaderStageFlags = 4
	GPUShaderStageGeometryBit               GPUShaderStageFlags = 8
	GPUShaderStageFragmentBit               GPUShaderStageFlags = 16
	GPUShaderStageComputeBit                GPUShaderStageFlags = 32
	GPUShaderStageAllGraphics               GPUShaderStageFlags = 31
	GPUShaderStageAll                       GPUShaderStageFlags = 2147483647
	GPUShaderStageTaskBitNv                 GPUShaderStageFlags = 64
	GPUShaderStageMeshBitNv                 GPUShaderStageFlags = 128
	GPUShaderStageRaygenBitNvx              GPUShaderStageFlags = 256
	GPUShaderStageAnyHitBitNvx              GPUShaderStageFlags = 512
	GPUShaderStageClosestHitBitNvx          GPUShaderStageFlags = 1024
	GPUShaderStageMissBitNvx                GPUShaderStageFlags = 2048
	GPUShaderStageIntersectionBitNvx        GPUShaderStageFlags = 4096
	GPUShaderStageCallableBitNvx            GPUShaderStageFlags = 8192
)

const (
	GPUPipelineBindPointGraphics      GPUPipelineBindPoint = 0
	GPUPipelineBindPointCompute       GPUPipelineBindPoint = 1
	GPUPipelineBindPointRaytracingNvx GPUPipelineBindPoint = 1000165000
)

const (
	GPUDependencyByRegionBit    GPUDependencyFlags = 1
	GPUDependencyViewLocalBit   GPUDependencyFlags = 2
	GPUDependencyDeviceGroupBit GPUDependencyFlags = 4
)
