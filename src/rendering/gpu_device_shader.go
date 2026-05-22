/******************************************************************************/
/* gpu_device_shader.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "kaijuengine.com/platform/profiler/tracing"

func (g *GPUDevice) DestroyShaderHandle(id ShaderId) {
	defer tracing.NewRegion("GPUDevice.DestroyShaderHandle").End()
	g.destroyShaderHandleImpl(id)
}
