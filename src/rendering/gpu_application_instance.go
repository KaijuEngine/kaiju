package rendering

import (
	"errors"
	"kaiju/build"
	"kaiju/platform/profiler/tracing"
)

type GPUApplicationInstance struct {
	GPUInstance
	Surface            GPUSurface
	Devices            []GPUDevice
	primaryDeviceIndex int
	dbg                *memoryDebugger
}

func (g *GPUApplicationInstance) SetupLogicalDevice(index int) error {
	defer tracing.NewRegion("GPUApplicationInstance.SetupLogicalDevice").End()
	if index < 0 || index >= len(g.Devices) {
		return errors.New("index out of bounds")
	}
	device := &g.Devices[index]
	return device.LogicalDevice.Setup(g, &device.PhysicalDevice)
}

func (g *GPUApplicationInstance) PrimaryDevice() *GPUDevice {
	return &g.Devices[g.primaryDeviceIndex]
}

func (g *GPUApplicationInstance) PhysicalDevice() *GPUPhysicalDevice {
	return &g.PrimaryDevice().PhysicalDevice
}

func (g *GPUApplicationInstance) Initialize(window RenderingContainer, app *GPUApplication) error {
	defer tracing.NewRegion("GPUApplicationInstance.Initialize").End()
	g.dbg = &memoryDebugger{}
	if err := g.Create(window, app); err != nil {
		return err
	}
	g.dbg.track(g.handle)
	if err := g.Surface.Create(&g.GPUInstance, window); err != nil {
		return err
	}
	// TODO:  Allow passing in the device selection method
	if err := g.SelectPhysicalDevice(nil); err != nil {
		return err
	}
	return nil
}

func (g *GPUApplicationInstance) Destroy() {
	g.Surface.Destroy(g)
	g.GPUInstance.Destroy()
	g.dbg.remove(g.handle)
}

func (g *GPUApplicationInstance) SelectPhysicalDevice(method func(options []GPUPhysicalDevice) int) error {
	devices, err := ListPhysicalGpuDevices(g)
	if err != nil {
		return err
	}
	if method == nil {
		method = selectPhysicalDeviceDefaltMethod
	}
	g.primaryDeviceIndex = method(devices)
	if g.primaryDeviceIndex < 0 {
		return errors.New("invalid primary physical device index: negative")
	} else if g.primaryDeviceIndex >= len(g.Devices) {
		return errors.New("invalid primary physical device index: out of range")
	}
	g.Devices = make([]GPUDevice, len(devices))
	for i := range devices {
		g.Devices[i].PhysicalDevice = devices[i]
	}
	return nil
}

func (g *GPUApplicationInstance) SetupDebug() {
	if build.Debug {
		for i := range g.Devices {
			if g.Devices[i].LogicalDevice.IsValid() {
				g.Devices[i].LogicalDevice.SetupDebug(g)
			}
		}
	}
}
