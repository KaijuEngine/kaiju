package rendering

import "log/slog"

const (
	engineVersionMajor = int(1)
	engineVersionMinor = int(0)
	engineVersionPatch = int(0)
)

type GPUApplication struct {
	Name    string
	Version struct {
		Major int
		Minor int
		Patch int
	}
	Instances []*GPUApplicationInstance
}

func (g *GPUApplication) IsValid() bool { return len(g.Instances) > 0 }

func (g *GPUApplication) FirstInstance() *GPUApplicationInstance {
	return g.Instances[0]
}

// TODO:  This function will likely go away with other refactors
func (g *GPUApplication) Dbg() *memoryDebugger {
	return g.FirstInstance().PrimaryDevice().LogicalDevice.dbg
}

func (g *GPUApplication) Instance(index int) (*GPUApplicationInstance, bool) {
	if index < 0 || index > len(g.Instances) {
		slog.Error("index out of range for the instances", "has", len(g.Instances), "wants", index)
		return nil, false
	}
	return g.Instances[index], true
}

func (g *GPUApplication) ApplicationVersion() (major int, minor int, patch int) {
	return g.Version.Major, g.Version.Minor, g.Version.Patch
}

func (g *GPUApplication) EngineVersion() (major int, minor int, patch int) {
	return engineVersionMajor, engineVersionMinor, engineVersionPatch
}

func (g *GPUApplication) CreateInstance(window RenderingContainer) (*GPUApplicationInstance, error) {
	g.Instances = append(g.Instances, &GPUApplicationInstance{})
	if err := g.Instances[len(g.Instances)-1].Initialize(window, g); err != nil {
		return nil, err
	}
	return g.Instances[len(g.Instances)-1], nil
}

func (g *GPUApplication) Destroy() {
	for i := range g.Instances {
		g.Instances[i].Destroy()
	}
}
