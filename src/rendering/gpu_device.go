package rendering

import "kaiju/platform/profiler/tracing"

type GPUDevice struct {
	PhysicalDevice GPUPhysicalDevice
	LogicalDevice  GPULogicalDevice
}

func (g *GPUDevice) CreateSwapChain(window RenderingContainer, inst *GPUApplicationInstance) error {
	defer tracing.NewRegion("GPUDevice.CreateSwapChain").End()
	return g.LogicalDevice.SwapChain.Setup(window, inst, g)
}
