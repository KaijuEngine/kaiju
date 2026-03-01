/******************************************************************************/
/* gpu_physical_device.go                                                     */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
	"log/slog"
	"strings"
	"unsafe"
)

type GPUPhysicalDevice struct {
	handle              unsafe.Pointer
	Features            GPUPhysicalDeviceFeatures
	Properties          GPUPhysicalDeviceProperties
	QueueFamilies       []GPUQueueFamily
	Extensions          []GPUPhysicalDeviceExtension
	SurfaceFormats      []GPUSurfaceFormat
	PresentModes        []GPUPresentMode
	SurfaceCapabilities GPUSurfaceCapabilities
}

type GPUPhysicalDeviceMemoryProperties struct {
	MemoryTypes []GPUMemoryType
	MemoryHeaps []GPUMemoryHeap
}

type GPUPhysicalDeviceExtension struct {
	Name    string
	Version uint32
}

type GPUPhysicalDeviceFeatures struct {
	RobustBufferAccess                      bool
	FullDrawIndexUint32                     bool
	ImageCubeArray                          bool
	IndependentBlend                        bool
	GeometryShader                          bool
	TessellationShader                      bool
	SampleRateShading                       bool
	DualSrcBlend                            bool
	LogicOp                                 bool
	MultiDrawIndirect                       bool
	DrawIndirectFirstInstance               bool
	DepthClamp                              bool
	DepthBiasClamp                          bool
	FillModeNonSolid                        bool
	DepthBounds                             bool
	WideLines                               bool
	LargePoints                             bool
	AlphaToOne                              bool
	MultiViewport                           bool
	SamplerAnisotropy                       bool
	TextureCompressionETC2                  bool
	TextureCompressionASTC_LDR              bool
	TextureCompressionBC                    bool
	OcclusionQueryPrecise                   bool
	PipelineStatisticsQuery                 bool
	VertexPipelineStoresAndAtomics          bool
	FragmentStoresAndAtomics                bool
	ShaderTessellationAndGeometryPointSize  bool
	ShaderImageGatherExtended               bool
	ShaderStorageImageExtendedFormats       bool
	ShaderStorageImageMultisample           bool
	ShaderStorageImageReadWithoutFormat     bool
	ShaderStorageImageWriteWithoutFormat    bool
	ShaderUniformBufferArrayDynamicIndexing bool
	ShaderSampledImageArrayDynamicIndexing  bool
	ShaderStorageBufferArrayDynamicIndexing bool
	ShaderStorageImageArrayDynamicIndexing  bool
	ShaderClipDistance                      bool
	ShaderCullDistance                      bool
	ShaderFloat64                           bool
	ShaderInt64                             bool
	ShaderInt16                             bool
	ShaderResourceResidency                 bool
	ShaderResourceMinLod                    bool
	SparseBinding                           bool
	SparseResidencyBuffer                   bool
	SparseResidencyImage2D                  bool
	SparseResidencyImage3D                  bool
	SparseResidency2Samples                 bool
	SparseResidency4Samples                 bool
	SparseResidency8Samples                 bool
	SparseResidency16Samples                bool
	SparseResidencyAliased                  bool
	VariableMultisampleRate                 bool
	InheritedQueries                        bool
}

type GPUPhysicalDeviceProperties struct {
	ApiVersion        uint32
	DriverVersion     uint32
	VendorID          uint32
	DeviceID          uint32
	DeviceType        GPUPhysicalDeviceType
	DeviceName        string
	PipelineCacheUUID string
	Limits            GPUPhysicalDeviceLimits
	SparseProperties  GPUPhysicalDeviceSparseProperties
}

type GPUPhysicalDeviceLimits struct {
	MaxImageDimension1D                             uint32
	MaxImageDimension2D                             uint32
	MaxImageDimension3D                             uint32
	MaxImageDimensionCube                           uint32
	MaxImageArrayLayers                             uint32
	MaxTexelBufferElements                          uint32
	MaxUniformBufferRange                           uint32
	MaxStorageBufferRange                           uint32
	MaxPushConstantsSize                            uint32
	MaxMemoryAllocationCount                        uint32
	MaxSamplerAllocationCount                       uint32
	BufferImageGranularity                          uintptr
	SparseAddressSpaceSize                          uintptr
	MaxBoundDescriptorSets                          uint32
	MaxPerStageDescriptorSamplers                   uint32
	MaxPerStageDescriptorUniformBuffers             uint32
	MaxPerStageDescriptorStorageBuffers             uint32
	MaxPerStageDescriptorSampledImages              uint32
	MaxPerStageDescriptorStorageImages              uint32
	MaxPerStageDescriptorInputAttachments           uint32
	MaxPerStageResources                            uint32
	MaxDescriptorSetSamplers                        uint32
	MaxDescriptorSetUniformBuffers                  uint32
	MaxDescriptorSetUniformBuffersDynamic           uint32
	MaxDescriptorSetStorageBuffers                  uint32
	MaxDescriptorSetStorageBuffersDynamic           uint32
	MaxDescriptorSetSampledImages                   uint32
	MaxDescriptorSetStorageImages                   uint32
	MaxDescriptorSetInputAttachments                uint32
	MaxVertexInputAttributes                        uint32
	MaxVertexInputBindings                          uint32
	MaxVertexInputAttributeOffset                   uint32
	MaxVertexInputBindingStride                     uint32
	MaxVertexOutputComponents                       uint32
	MaxTessellationGenerationLevel                  uint32
	MaxTessellationPatchSize                        uint32
	MaxTessellationControlPerVertexInputComponents  uint32
	MaxTessellationControlPerVertexOutputComponents uint32
	MaxTessellationControlPerPatchOutputComponents  uint32
	MaxTessellationControlTotalOutputComponents     uint32
	MaxTessellationEvaluationInputComponents        uint32
	MaxTessellationEvaluationOutputComponents       uint32
	MaxGeometryShaderInvocations                    uint32
	MaxGeometryInputComponents                      uint32
	MaxGeometryOutputComponents                     uint32
	MaxGeometryOutputVertices                       uint32
	MaxGeometryTotalOutputComponents                uint32
	MaxFragmentInputComponents                      uint32
	MaxFragmentOutputAttachments                    uint32
	MaxFragmentDualSrcAttachments                   uint32
	MaxFragmentCombinedOutputResources              uint32
	MaxComputeSharedMemorySize                      uint32
	MaxComputeWorkGroupCount                        [3]uint32
	MaxComputeWorkGroupInvocations                  uint32
	MaxComputeWorkGroupSize                         [3]uint32
	SubPixelPrecisionBits                           uint32
	SubTexelPrecisionBits                           uint32
	MipmapPrecisionBits                             uint32
	MaxDrawIndexedIndexValue                        uint32
	MaxDrawIndirectCount                            uint32
	MaxSamplerLodBias                               float32
	MaxSamplerAnisotropy                            float32
	MaxViewports                                    uint32
	MaxViewportDimensions                           [2]uint32
	ViewportBoundsRange                             [2]float32
	ViewportSubPixelBits                            uint32
	MinMemoryMapAlignment                           uint
	MinTexelBufferOffsetAlignment                   uintptr
	MinUniformBufferOffsetAlignment                 uintptr
	MinStorageBufferOffsetAlignment                 uintptr
	MinTexelOffset                                  int32
	MaxTexelOffset                                  uint32
	MinTexelGatherOffset                            int32
	MaxTexelGatherOffset                            uint32
	MinInterpolationOffset                          float32
	MaxInterpolationOffset                          float32
	SubPixelInterpolationOffsetBits                 uint32
	MaxFramebufferWidth                             uint32
	MaxFramebufferHeight                            uint32
	MaxFramebufferLayers                            uint32
	FramebufferColorSampleCounts                    GPUSampleCountFlags
	FramebufferDepthSampleCounts                    GPUSampleCountFlags
	FramebufferStencilSampleCounts                  GPUSampleCountFlags
	FramebufferNoAttachmentsSampleCounts            GPUSampleCountFlags
	MaxColorAttachments                             uint32
	SampledImageColorSampleCounts                   GPUSampleCountFlags
	SampledImageIntegerSampleCounts                 GPUSampleCountFlags
	SampledImageDepthSampleCounts                   GPUSampleCountFlags
	SampledImageStencilSampleCounts                 GPUSampleCountFlags
	StorageImageSampleCounts                        GPUSampleCountFlags
	MaxSampleMaskWords                              uint32
	TimestampComputeAndGraphics                     bool
	TimestampPeriod                                 float32
	MaxClipDistances                                uint32
	MaxCullDistances                                uint32
	MaxCombinedClipAndCullDistances                 uint32
	DiscreteQueuePriorities                         uint32
	PointSizeRange                                  [2]float32
	LineWidthRange                                  [2]float32
	PointSizeGranularity                            float32
	LineWidthGranularity                            float32
	StrictLines                                     bool
	StandardSampleLocations                         bool
	OptimalBufferCopyOffsetAlignment                uintptr
	OptimalBufferCopyRowPitchAlignment              uintptr
	NonCoherentAtomSize                             uintptr
}

type GPUPhysicalDeviceSparseProperties struct {
	ResidencyStandard2DBlockShape            bool
	ResidencyStandard2DMultisampleBlockShape bool
	ResidencyStandard3DBlockShape            bool
	ResidencyAlignedMipSize                  bool
	ResidencyNonResidentStrict               bool
}

func ListPhysicalGpuDevices(inst *GPUApplicationInstance) ([]GPUPhysicalDevice, error) {
	defer tracing.NewRegion("rendering.ListPhysicalGpuDevices").End()
	return listPhysicalGpuDevicesImpl(inst)
}

func (g *GPUPhysicalDevice) IsValid() bool { return g.handle != nil }

func (g *GPUPhysicalDevice) FindGraphicsFamiliy() GPUQueueFamily {
	defer tracing.NewRegion("GPUPhysicalDevice.FindGraphicsFamiliy").End()
	for i := range g.QueueFamilies {
		if g.QueueFamilies[i].IsGraphics {
			return g.QueueFamilies[i]
		}
	}
	return InvalidGPUQueueFamily()
}

func (g *GPUPhysicalDevice) FindComputeFamiliy() GPUQueueFamily {
	defer tracing.NewRegion("GPUPhysicalDevice.FindComputeFamiliy").End()
	for i := range g.QueueFamilies {
		if g.QueueFamilies[i].IsCompute {
			return g.QueueFamilies[i]
		}
	}
	return InvalidGPUQueueFamily()
}

func (g *GPUPhysicalDevice) FindPresentFamily() GPUQueueFamily {
	defer tracing.NewRegion("GPUPhysicalDevice.FindPresentFamily").End()
	for i := range g.QueueFamilies {
		if g.QueueFamilies[i].HasPresentSupport {
			return g.QueueFamilies[i]
		}
	}
	return InvalidGPUQueueFamily()
}

func (g *GPUPhysicalDevice) IsExtensionSupported(extension string) bool {
	defer tracing.NewRegion("GPUPhysicalDevice.IsExtensionSupported").End()
	for i := range g.Extensions {
		if strings.EqualFold(g.Extensions[i].Name, extension) {
			return true
		}
	}
	return false
}

func (g *GPUPhysicalDevice) MaxUsableSampleCount() GPUSampleCountFlags {
	defer tracing.NewRegion("GPUPhysicalDevice.MaxUsableSampleCount").End()
	counts := vk.SampleCountFlags(g.Properties.Limits.FramebufferColorSampleCounts & g.Properties.Limits.FramebufferDepthSampleCounts)
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount64Bit)) != 0 {
		return GPUSampleCount64Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount32Bit)) != 0 {
		return GPUSampleCount32Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount16Bit)) != 0 {
		return GPUSampleCount16Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount8Bit)) != 0 {
		return GPUSampleCount8Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount4Bit)) != 0 {
		return GPUSampleCount4Bit
	}
	if (counts & vk.SampleCountFlags(vulkan_const.SampleCount2Bit)) != 0 {
		return GPUSampleCount2Bit
	}
	return GPUSampleCount1Bit
}

func (g *GPUPhysicalDevice) FormatProperties(format GPUFormat) GPUFormatProperties {
	defer tracing.NewRegion("GPUPhysicalDevice.FormatProperties").End()
	return g.formatPropertiesImpl(format)
}

func (g *GPUPhysicalDevice) FindSupportedFormat(candidates []GPUFormat, tiling GPUImageTiling, features GPUFormatFeatureFlags) GPUFormat {
	for i := 0; i < len(candidates); i++ {
		format := candidates[i]
		props := g.FormatProperties(format)
		if tiling == GPUImageTilingLinear && (props.LinearTilingFeatures&features) == features {
			return format
		} else if tiling == GPUImageTilingOptimal && (props.OptimalTilingFeatures&features) == features {
			return format
		}
	}
	slog.Error("Failed to find supported format")
	// TODO:  Return an error too
	return candidates[0]
}

func (g *GPUPhysicalDevice) FindMemoryType(typeFilter uint32, properties GPUMemoryPropertyFlags) int {
	defer tracing.NewRegion("GPUPhysicalDevice.FindMemoryType").End()
	return g.findMemoryTypeImpl(typeFilter, properties)
}

func (g *GPUPhysicalDevice) PadBufferSize(size uintptr) uintptr {
	// Calculate required alignment based on minimum device offset alignment
	minUboAlignment := g.Properties.Limits.MinUniformBufferOffsetAlignment
	alignedSize := size
	if minUboAlignment > 0 {
		alignedSize = (alignedSize + minUboAlignment - 1) & ^(minUboAlignment - 1)
	}
	return alignedSize
}

func (g *GPUPhysicalDevice) isPhysicalDeviceSuitableForRendering() bool {
	defer tracing.NewRegion("GPUPhysicalDevice.isPhysicalDeviceSuitableForRendering").End()
	exts := requiredDeviceExtensions()
	hasExtensions := true
	for i := 0; i < len(exts) && hasExtensions; i++ {
		// TODO:  This is temp, the extensions are probably going to change
		// with the new GPU implementation
		exts[i] = strings.TrimRight(exts[i], "\x00")
		hasExtensions = g.IsExtensionSupported(exts[i])
	}
	swapChainAdequate := false
	if hasExtensions {
		swapChainAdequate = len(g.SurfaceFormats) > 0 && len(g.PresentModes) > 0
	}
	graphicsFam := g.FindGraphicsFamiliy()
	presentFam := g.FindPresentFamily()
	return graphicsFam.IsValid() && presentFam.IsValid() &&
		hasExtensions && swapChainAdequate &&
		g.Features.SamplerAnisotropy
}

func selectPhysicalDeviceDefaltMethod(options []GPUPhysicalDevice) int {
	defer tracing.NewRegion("rendering.selectPhysicalDeviceDefaltMethod").End()
	slog.Info("locating suitable physical graphics device")
	var currentPhysicalDevice GPUPhysicalDevice
	var physicalDevice GPUPhysicalDevice
	selectedIndex := -1
	for i := range options {
		g := options[i]
		if g.isPhysicalDeviceSuitableForRendering() {
			currentPhysicalDevice = g
		}
		pick := !physicalDevice.IsValid()
		if !pick {
			t := physicalDevice.Properties.DeviceType
			ct := currentPhysicalDevice.Properties.DeviceType
			m := physicalDevice.Properties.Limits.MaxComputeSharedMemorySize
			cm := currentPhysicalDevice.Properties.Limits.MaxComputeSharedMemorySize
			a := physicalDevice.Properties.ApiVersion
			ca := currentPhysicalDevice.Properties.ApiVersion
			d := physicalDevice.Properties.DriverVersion
			cd := currentPhysicalDevice.Properties.DriverVersion
			if isPhysicalDeviceBetterType(ct, t) {
				pick = true
			} else if t == ct {
				pick = m < cm
				pick = pick || (m == cm && a < ca)
				pick = pick || (m == cm && a == ca && d < cd)
			}
		}
		if pick {
			physicalDevice = currentPhysicalDevice
			selectedIndex = i
		}
	}
	if !physicalDevice.IsValid() {
		slog.Error("Failed to find a compatible physical device")
		return -1
	}
	return selectedIndex
}

func isPhysicalDeviceBetterType(a GPUPhysicalDeviceType, b GPUPhysicalDeviceType) bool {
	defer tracing.NewRegion("rendering.isPhysicalDeviceBetterType").End()
	type score struct {
		deviceType GPUPhysicalDeviceType
		score      int
	}
	scores := []score{
		{GPUPhysicalDeviceTypeCpu, 1},
		{GPUPhysicalDeviceTypeOther, 1},
		{GPUPhysicalDeviceTypeVirtualGpu, 1},
		{GPUPhysicalDeviceTypeIntegratedGpu, 2},
		{GPUPhysicalDeviceTypeDiscreteGpu, 3},
	}
	aScore, bScore := 0, 0
	for i := 0; i < len(scores); i++ {
		if scores[i].deviceType == a {
			aScore += scores[i].score
		}
		if scores[i].deviceType == b {
			bScore += scores[i].score
		}
	}
	return aScore > bScore
}

func (g *GPUPhysicalDevice) FormatIsTileable(format GPUFormat, tiling GPUImageTiling) bool {
	defer tracing.NewRegion("GPUPhysicalDevice.FormatIsTileable").End()
	props := g.FormatProperties(format)
	switch tiling {
	case GPUImageTilingOptimal:
		return (props.OptimalTilingFeatures & GPUFormatFeatureSampledImageFilterLinearBit) != 0
	case GPUImageTilingLinear:
		return (props.LinearTilingFeatures & GPUFormatFeatureSampledImageFilterLinearBit) != 0
	default:
		return false
	}
}
