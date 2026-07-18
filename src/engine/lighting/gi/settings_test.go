package gi

import "testing"

func TestPresetSettingsValidate(t *testing.T) {
	for preset := QualityPresetOff; preset <= QualityPresetCustom; preset++ {
		settings := SettingsForPreset(preset)
		if err := settings.Validate(); err != nil {
			t.Fatalf("preset %d failed validation: %v", preset, err)
		}
	}
}

func TestDynamicSettingsRequireProbeWork(t *testing.T) {
	settings := SettingsForPreset(QualityPresetHigh)
	settings.RaysPerProbe = 0
	if err := settings.Validate(); err == nil {
		t.Fatal("expected rays-per-probe validation error")
	}
}

func TestCapabilitiesSupportDynamicDDGI(t *testing.T) {
	if !(Capabilities{
		VulkanMajor:            1,
		VulkanMinor:            2,
		BufferDeviceAddress:    true,
		AccelerationStructure:  true,
		DeferredHostOperations: true,
		RayQuery:               true,
	}).SupportsDynamicDDGI() {
		t.Fatal("complete Vulkan 1.2 ray-query capability should be supported")
	}
	if (Capabilities{VulkanMajor: 1, VulkanMinor: 2}).SupportsDynamicDDGI() {
		t.Fatal("Vulkan 1.2 without ray-query features should not be supported")
	}
	if (Capabilities{VulkanMajor: 1, VulkanMinor: 1, BufferDeviceAddress: true, AccelerationStructure: true, DeferredHostOperations: true, RayQuery: true}).SupportsDynamicDDGI() {
		t.Fatal("Vulkan 1.1 should not satisfy the DDGI shader requirements")
	}
}
