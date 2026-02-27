package rendering

import (
	"errors"
	"kaijuengine.com/build"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
	"runtime"
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

func (g *GPUApplicationInstance) Initialize(window RenderingContainer, app *GPUApplication, assets assets.Database) error {
	defer tracing.NewRegion("GPUApplicationInstance.Initialize").End()
	g.dbg = &memoryDebugger{}
	if err := g.Setup(window, app); err != nil {
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
	if err := g.SetupLogicalDevice(0); err != nil {
		return err
	}
	g.SetupDebug()
	device := g.PrimaryDevice()
	if err := device.CreateSwapChain(window, g); err != nil {
		return err
	}
	swapChain := &device.LogicalDevice.SwapChain
	if err := swapChain.SetupImageViews(device); err != nil {
		return err
	}
	if err := swapChain.SetupRenderPass(device, assets); err != nil {
		return err
	}
	if err := swapChain.CreateColor(device); err != nil {
		return err
	}
	if err := swapChain.CreateDepth(device); err != nil {
		return err
	}
	if err := swapChain.CreateFrameBuffer(device); err != nil {
		return err
	}
	if err := device.createGlobalUniforms(); err != nil {
		return err
	}
	if err := device.createDescriptorPool(1000); err != nil {
		return err
	}
	if err := swapChain.SetupSyncObjects(device); err != nil {
		return err
	}
	var err error
	for i := range len(device.Painter.combineCmds) {
		if device.Painter.combineCmds[i], err = NewCommandRecorder(device); err != nil {
			return err
		}
	}
	for i := range len(device.Painter.blitCmds) {
		if device.Painter.blitCmds[i], err = NewCommandRecorder(device); err != nil {
			return err
		}
	}
	return nil
}

func (g *GPUApplicationInstance) SetupCaches(caches RenderCaches, width, height int32) error {
	defer tracing.NewRegion("GPUApplicationInstance.SetupCaches").End()
	device := g.PrimaryDevice()
	device.Painter.caches = caches
	var err error
	device.Painter.fallbackShadowMap, err = caches.TextureCache().Texture(assets.TextureSquare, TextureFilterLinear)
	if err != nil {
		return err
	}
	device.Painter.fallbackCubeShadowMap, err = caches.TextureCache().Texture(assets.TextureCube, TextureFilterLinear)
	if err != nil {
		return err
	}
	device.Painter.fallbackCubeShadowMap.SetPendingDataDimensions(TextureDimensionsCube)
	caches.TextureCache().CreatePending()
	return err
}

func (g *GPUApplicationInstance) Destroy() {
	defer tracing.NewRegion("GPUApplicationInstance.Destroy").End()
	g.dbg.remove(g.handle)
	for i := range g.Devices {
		device := &g.Devices[i]
		if !device.LogicalDevice.IsValid() {
			continue
		}
		device.LogicalDevice.WaitForRender(device)
		device.Painter.combinedDrawings.Destroy(device)
		device.LogicalDevice.bufferTrash.Purge()
		for k := range device.LogicalDevice.renderPassCache {
			device.LogicalDevice.renderPassCache[k].Destroy(device)
		}
		device.LogicalDevice.renderPassCache = make(map[string]*RenderPass)
		runtime.GC()
		for i := range device.Painter.preRuns {
			if device.Painter.preRuns[i] != nil {
				device.Painter.preRuns[i]()
			}
		}
		device.Painter.preRuns = make([]func(), 0)
		device.Painter.caches = nil
		for i := range device.Painter.combineCmds {
			device.Painter.combineCmds[i].Destroy(device)
		}
		for i := range device.Painter.blitCmds {
			device.Painter.blitCmds[i].Destroy(device)
		}
		device.singleTimeCommandPool.All(func(elm *CommandRecorder) {
			if elm.buffer != vk.NullCommandBuffer {
				elm.Destroy(device)
			}
		})
		device.destroyGlobalUniforms()
		device.Painter.DestroyDescriptorPools(device)
		device.LogicalDevice.SwapChain.renderPass.Destroy(device)
		device.LogicalDevice.SwapChain.Destroy(device)
		device.LogicalDevice.Destroy()
	}
	g.Surface.Destroy(g)
	g.GPUInstance.Destroy()
	g.dbg.print()
}

func (g *GPUApplicationInstance) SelectPhysicalDevice(method func(options []GPUPhysicalDevice) int) error {
	defer tracing.NewRegion("GPUApplicationInstance.SelectPhysicalDevice").End()
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
	} else if g.primaryDeviceIndex >= len(devices) {
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
				g.Devices[i].LogicalDevice.SetupDebug(g.PrimaryDevice())
			}
		}
	}
}

func (g *GPUApplicationInstance) Resize(window RenderingContainer, width, height int) {
	defer tracing.NewRegion("GPUApplicationInstance.Resize").End()
	for i := range g.Devices {
		if g.Devices[i].LogicalDevice.IsValid() {
			g.Devices[i].LogicalDevice.RemakeSwapChain(window, g, &g.Devices[i])
		}
	}
}
