package rendering

type GPUApplication struct {
	Name    string
	Version struct {
		Major int
		Minor int
		Patch int
	}
	Instance       GPUInstance
	Surface        GPUSurface
	PhysicalDevice GPUPhysicalDevice
	dbg            memoryDebugger
}

func (g *GPUApplication) ApplicationVersion() (major int, minor int, patch int) {
	return g.Version.Major, g.Version.Minor, g.Version.Patch
}

func (g *GPUApplication) EngineVersion() (major int, minor int, patch int) {
	return engineVersionMajor, engineVersionMinor, engineVersionPatch
}

func (g *GPUApplication) Create(window RenderingContainer) error {
	if err := g.Instance.Create(window, g); err != nil {
		return err
	}
	if err := g.Surface.Create(&g.Instance, window); err != nil {
		return err
	}
	// TODO:  Allow passing in the device selection method
	if err := g.SelectPhysicalDevice(nil); err != nil {
		return err
	}
	return nil
}

func (g *GPUApplication) SelectPhysicalDevice(method func(options []GPUPhysicalDevice) GPUPhysicalDevice) error {
	devices, err := ListPhysicalGpuDevices(g)
	if err != nil {
		return err
	}
	if method == nil {
		method = selectPhysicalDeviceDefaltMethod
	}
	g.PhysicalDevice = method(devices)
	return nil
}

func (g *GPUApplication) Destroy() {
	g.Surface.Destroy(g)
	g.Instance.Destroy(g)
}
