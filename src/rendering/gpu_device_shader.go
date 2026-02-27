package rendering

import "kaijuengine.com/platform/profiler/tracing"

func (g *GPUDevice) DestroyShaderHandle(id ShaderId) {
	defer tracing.NewRegion("GPUDevice.DestroyShaderHandle").End()
	g.destroyShaderHandleImpl(id)
}
