package gpu_enums

import (
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"kaijuengine.com/rendering/vulkan_const"
)

var (
	gpuPipelineBindPointToVulkan = map[PipelineBindPoint]vulkan_const.PipelineBindPoint{
		PipelineBindPointGraphics:      vulkan_const.PipelineBindPointGraphics,
		PipelineBindPointCompute:       vulkan_const.PipelineBindPointCompute,
		PipelineBindPointRaytracingNvx: vulkan_const.PipelineBindPointRaytracingNvx,
	}
	gpuPipelineBindPointFromVulkan = map[vulkan_const.PipelineBindPoint]PipelineBindPoint{
		vulkan_const.PipelineBindPointGraphics:      PipelineBindPointGraphics,
		vulkan_const.PipelineBindPointCompute:       PipelineBindPointCompute,
		vulkan_const.PipelineBindPointRaytracingNvx: PipelineBindPointRaytracingNvx,
	}
	gpuDependencyFlagBits = []DependencyFlags{
		DependencyByRegionBit,
		DependencyViewLocalBit,
		DependencyDeviceGroupBit,
	}
	vkDependencyFlagBits = []vulkan_const.DependencyFlagBits{
		vulkan_const.DependencyByRegionBit,
		vulkan_const.DependencyViewLocalBit,
		vulkan_const.DependencyDeviceGroupBit,
	}
	_ = [unsafe.Sizeof(gpuDependencyFlagBits)/unsafe.Sizeof(gpuDependencyFlagBits[0]) - unsafe.Sizeof(vkDependencyFlagBits)/unsafe.Sizeof(vkDependencyFlagBits[0])]struct{}{}
)

func (g PipelineBindPoint) ToVulkan() vulkan_const.PipelineBindPoint {
	defer tracing.NewRegion("PipelineBindPoint.ToVulkan").End()
	out, ok := gpuPipelineBindPointToVulkan[g]
	if !ok {
		panic("invalid pipeline bind point supplied")
	}
	return out
}

func (g *PipelineBindPoint) FromVulkan(val vulkan_const.PipelineBindPoint) {
	defer tracing.NewRegion("PipelineBindPoint.FromVulkan").End()
	out, ok := gpuPipelineBindPointFromVulkan[val]
	if !ok {
		panic("invalid pipeline bind point supplied")
	}
	*g = out
}

func (g *DependencyFlags) FromVulkan(val vk.DependencyFlags) {
	defer tracing.NewRegion("DependencyFlags.FromVulkan").End()
	var flags DependencyFlags
	for i := range vkDependencyFlagBits {
		if val&vk.DependencyFlags(vkDependencyFlagBits[i]) != 0 {
			flags |= gpuDependencyFlagBits[i]
		}
	}
	*g = flags
}

func (g DependencyFlags) ToVulkan() vk.DependencyFlags {
	defer tracing.NewRegion("DependencyFlags.ToVulkan").End()
	val := g
	var flags vk.DependencyFlags
	for i := range gpuDependencyFlagBits {
		if val&gpuDependencyFlagBits[i] != 0 {
			flags |= vk.DependencyFlags(vkDependencyFlagBits[i])
		}
	}
	return flags
}

func (g PrimitiveTopology) ToVulkan() vulkan_const.PrimitiveTopology {
	return vulkan_const.PrimitiveTopology(g)
}

func (g PolygonMode) ToVulkan() vulkan_const.PolygonMode {
	return vulkan_const.PolygonMode(g)
}

func (g CullModeFlags) ToVulkan() vk.CullModeFlags {
	return vk.CullModeFlags(g)
}

func (g FrontFace) ToVulkan() vulkan_const.FrontFace {
	return vulkan_const.FrontFace(g)
}

func (g LogicOp) ToVulkan() vulkan_const.LogicOp {
	return vulkan_const.LogicOp(g)
}

func (g CompareOp) ToVulkan() vulkan_const.CompareOp {
	return vulkan_const.CompareOp(g)
}

func (g StencilOp) ToVulkan() vulkan_const.StencilOp {
	return vulkan_const.StencilOp(g)
}

func (g BlendFactor) ToVulkan() vulkan_const.BlendFactor {
	return vulkan_const.BlendFactor(g)
}

func (g BlendOp) ToVulkan() vulkan_const.BlendOp {
	return vulkan_const.BlendOp(g)
}

func (g ColorComponentFlags) ToVulkan() vk.ColorComponentFlags {
	return vk.ColorComponentFlags(g)
}

func (g PipelineCreateFlags) ToVulkan() vk.PipelineCreateFlags {
	return vk.PipelineCreateFlags(g)
}

func (g ShaderStageFlags) ToVulkan() vk.ShaderStageFlags {
	return vk.ShaderStageFlags(g)
}

func (s StencilOpState) ToVulkan() vk.StencilOpState {
	return vk.StencilOpState{
		FailOp:      s.FailOp.ToVulkan(),
		PassOp:      s.PassOp.ToVulkan(),
		DepthFailOp: s.DepthFailOp.ToVulkan(),
		CompareOp:   s.CompareOp.ToVulkan(),
		CompareMask: s.CompareMask,
		WriteMask:   s.WriteMask,
		Reference:   s.Reference,
	}
}
