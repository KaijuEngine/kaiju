/******************************************************************************/
/* gpu_instance.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
)

type GPUInstance struct {
	handle unsafe.Pointer
}

func (g *GPUInstance) IsValid() bool { return g.handle != nil }

func (g *GPUInstance) Setup(window RenderingContainer, app *GPUApplication) error {
	defer tracing.NewRegion("GPUInstance.Create").End()
	return g.setupImpl(window, app)
}

func (g *GPUInstance) Destroy() {
	defer tracing.NewRegion("rendering.Destroy").End()
	g.destroyImpl()
}
