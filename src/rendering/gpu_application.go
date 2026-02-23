package rendering

import "unsafe"

type GPUApplication struct {
	Name    string
	Version struct {
		Major int
		Minor int
		Patch int
	}
	Instance GPUInstance
	Surface  GPUSurface
	dbg      memoryDebugger
}

func TEMP(r Renderer) *GPUApplication {
	vr := r.(*Vulkan)
	g := &GPUApplication{}
	g.Instance.handle = unsafe.Pointer(vr.instance)
	g.Surface.handle = unsafe.Pointer(vr.surface)
	return g
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
	return nil
}

func (g *GPUApplication) Destroy() {
	g.Surface.Destroy(g)
	g.Instance.Destroy(g)
}
