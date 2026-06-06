/******************************************************************************/
/* gpu_surface_vulkan.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
)

func (g *GPUSurface) destroyImpl(inst *GPUApplicationInstance) {
	defer tracing.NewRegion("GPUSurface.destroyImpl").End()
	vk.DestroySurface(vk.Instance(inst.handle), vk.Surface(g.handle), nil)
	inst.dbg.remove(g.handle)
}
