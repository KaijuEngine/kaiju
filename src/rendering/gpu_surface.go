/******************************************************************************/
/* gpu_surface.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"
	"unsafe"

	"kaijuengine.com/platform/profiler/tracing"
)

type GPUSurface struct {
	handle unsafe.Pointer
}

func (g *GPUSurface) Create(instance *GPUInstance, window RenderingContainer) error {
	defer tracing.NewRegion("GPUSurface.Create").End()
	slog.Info("creating gpu surface")
	return g.createImpl(instance, window)
}

func (g *GPUSurface) Destroy(inst *GPUApplicationInstance) {
	defer tracing.NewRegion("GPUSurface.Destroy").End()
	g.destroyImpl(inst)
}
