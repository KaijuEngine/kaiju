package rendering

import (
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"unsafe"
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
