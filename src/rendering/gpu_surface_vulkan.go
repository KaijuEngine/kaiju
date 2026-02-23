package rendering

import (
	"kaiju/platform/profiler/tracing"
	vk "kaiju/rendering/vulkan"
)

func (g *GPUSurface) destroyImpl(app *GPUApplication) {
	defer tracing.NewRegion("GPUSurface.destroyImpl").End()
	vk.DestroySurface(vk.Instance(app.Instance.handle), vk.Surface(g.handle), nil)
	app.dbg.remove(g.handle)
}
