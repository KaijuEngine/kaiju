package gi

import (
	"errors"
	"strings"
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type managerTestProvider struct {
	id         string
	supported  bool
	initErr    error
	shutdown   *int
	configured Settings
}

func (p *managerTestProvider) ID() string                                             { return p.id }
func (p *managerTestProvider) Supports(Capabilities) bool                             { return p.supported }
func (p *managerTestProvider) Initialize(ProviderContext) error                       { return p.initErr }
func (p *managerTestProvider) Configure(s Settings) error                             { p.configured = s; return nil }
func (*managerTestProvider) SyncScene(SceneDelta) error                               { return nil }
func (*managerTestProvider) AddUpdatePasses(*rendering.FrameGraph, FrameInputs) error { return nil }
func (*managerTestProvider) AddResolvePasses(*rendering.FrameGraph, FrameInputs) (Outputs, error) {
	return Outputs{}, nil
}
func (*managerTestProvider) ProbeField(ViewID) ProbeFieldBinding { return ProbeFieldBinding{} }
func (*managerTestProvider) ShaderData(matrix.Vec3, float32) rendering.GlobalIlluminationForRender {
	return rendering.GlobalIlluminationForRender{}
}
func (*managerTestProvider) Invalidate(Invalidation)  {}
func (*managerTestProvider) ResetHistory(ViewID)      {}
func (*managerTestProvider) SetScenario(string) error { return nil }
func (p *managerTestProvider) Stats() Stats           { return Stats{Provider: p.id} }
func (*managerTestProvider) DebugViews() []DebugView  { return nil }
func (p *managerTestProvider) Shutdown() {
	if p.shutdown != nil {
		*p.shutdown = *p.shutdown + 1
	}
}

func TestManagerAutoFallsBackToBaked(t *testing.T) {
	manager := NewManager(Capabilities{
		VulkanMajor:            1,
		VulkanMinor:            2,
		BufferDeviceAddress:    true,
		AccelerationStructure:  true,
		DeferredHostOperations: true,
		RayQuery:               true,
	})
	manager.RegisterProvider(ProviderDDGI, func() Provider {
		return &managerTestProvider{id: ProviderDDGI, supported: true, initErr: errors.New("driver failure")}
	})
	manager.RegisterProvider(ProviderBakedProbe, func() Provider {
		return &managerTestProvider{id: ProviderBakedProbe, supported: true}
	})
	if err := manager.Configure(SettingsForPreset(QualityPresetHigh)); err != nil {
		t.Fatal(err)
	}
	if got := manager.ActiveProvider(); got != ProviderBakedProbe {
		t.Fatalf("active provider = %q", got)
	}
	if reason := manager.Stats().FallbackReason; !strings.Contains(reason, "driver failure") {
		t.Fatalf("fallback reason = %q", reason)
	}
}

func TestManagerRequireExactDoesNotFallback(t *testing.T) {
	manager := NewManager(Capabilities{VulkanMajor: 1, VulkanMinor: 2})
	manager.RegisterProvider(ProviderDDGI, func() Provider {
		return &managerTestProvider{id: ProviderDDGI, supported: true, initErr: errors.New("driver failure")}
	})
	manager.RegisterProvider(ProviderBakedProbe, func() Provider {
		return &managerTestProvider{id: ProviderBakedProbe, supported: true}
	})
	settings := SettingsForPreset(QualityPresetHigh)
	settings.Mode = ModeDynamicDDGI
	settings.Fallback = FallbackRequireExact
	if err := manager.Configure(settings); err == nil {
		t.Fatal("expected exact provider failure")
	}
	if got := manager.ActiveProvider(); got != ProviderNull {
		t.Fatalf("failed reconfiguration should preserve previous provider, got %q", got)
	}
}

func TestNullProviderAddsBlackIrradiancePass(t *testing.T) {
	manager := NewManager(Capabilities{})
	graph := rendering.NewFrameGraph()
	output, err := manager.AddFramePasses(graph, FrameInputs{View: 7})
	if err != nil {
		t.Fatal(err)
	}
	if !output.Valid || output.DiffuseIrradiance == 0 {
		t.Fatalf("invalid null output: %+v", output)
	}
	schedule, err := graph.Compile()
	if err != nil {
		t.Fatal(err)
	}
	context := &rendering.FrameGraphExecutionContext{}
	if err := schedule.Execute(context); err != nil {
		t.Fatal(err)
	}
	if got := context.Values["gi.null.diffuse.7"]; got != [4]float32{} {
		t.Fatalf("null irradiance = %#v", got)
	}
}

func TestManagerStageSettingsReturnToProjectDefaults(t *testing.T) {
	manager := NewManager(Capabilities{})
	project := SettingsForPreset(QualityPresetMedium)
	if err := manager.SetDefaultSettings(project); err != nil {
		t.Fatal(err)
	}
	override := SettingsForPreset(QualityPresetLow)
	if err := manager.ApplyStageSettings(&override, ""); err != nil {
		t.Fatal(err)
	}
	if got := manager.Settings().Preset; got != QualityPresetLow {
		t.Fatalf("stage preset = %v, want low", got)
	}
	if err := manager.ApplyStageSettings(nil, ""); err != nil {
		t.Fatal(err)
	}
	if got := manager.Settings().Preset; got != QualityPresetMedium {
		t.Fatalf("restored preset = %v, want medium", got)
	}
}

func TestManagerClearsBakedScenarioBetweenStages(t *testing.T) {
	asset := constantProbeAsset(matrix.NewVec3(1, 1, 1))
	data, err := asset.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	manager := NewManager(Capabilities{})
	manager.SetAssetReader(probeAssetReader{"day.kjgi": data})
	settings := SettingsForPreset(QualityPresetMedium)
	settings.Mode = ModeBaked
	if err := manager.SetDefaultSettings(settings); err != nil {
		t.Fatal(err)
	}
	if err := manager.ApplyStageSettings(nil, "day.kjgi"); err != nil {
		t.Fatal(err)
	}
	if manager.Stats().ActiveProbes == 0 {
		t.Fatal("expected loaded probes")
	}
	if err := manager.ApplyStageSettings(nil, ""); err != nil {
		t.Fatal(err)
	}
	if manager.Stats().ActiveProbes != 0 {
		t.Fatal("previous stage probes leaked after clearing the scenario")
	}
}
