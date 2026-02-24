package rendering

import (
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
)

func (g *GPUSurface) destroyImpl(inst *GPUApplicationInstance) {
	defer tracing.NewRegion("GPUSurface.destroyImpl").End()
	vk.DestroySurface(vk.Instance(inst.handle), vk.Surface(g.handle), nil)
	inst.dbg.remove(g.handle)
}
