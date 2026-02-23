package rendering

import (
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"unsafe"
)

const (
	engineVersionMajor = int(1)
	engineVersionMinor = int(0)
	engineVersionPatch = int(0)
)

type GPUInstance struct {
	handle unsafe.Pointer
}

func (g *GPUInstance) IsValid() bool { return g.handle != nil }

func (g *GPUInstance) Create(window RenderingContainer, app *GPUApplication) error {
	defer tracing.NewRegion("GPUInstance.Create").End()
	slog.Info("creating kaiju gpu instance")
	return g.createImpl(window, app)
}

func (g *GPUInstance) Destroy(app *GPUApplication) {
	defer tracing.NewRegion("rendering.Destroy").End()
	g.destroyImpl(app)
}
