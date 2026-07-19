package common_workspace

import (
	"testing"

	"kaijuengine.com/engine/lighting/gi"
)

func TestApplyGISettingsPresetAndCustomField(t *testing.T) {
	settings, err := ApplyGISettingsField(gi.SettingsForPreset(gi.QualityPresetOff), "Preset", "2", false)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Preset != gi.QualityPresetMedium || settings.Mode != gi.ModeAuto {
		t.Fatalf("medium preset was not applied: %+v", settings)
	}
	settings, err = ApplyGISettingsField(settings, "GPUTimeBudgetMS", "1.25", false)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Preset != gi.QualityPresetCustom || settings.GPUTimeBudgetMS != 1.25 {
		t.Fatalf("custom edit was not applied: %+v", settings)
	}
}

func TestApplyGISettingsRejectsInvalidValue(t *testing.T) {
	current := gi.SettingsForPreset(gi.QualityPresetMedium)
	next, err := ApplyGISettingsField(current, "ProbeSpacing", "0", false)
	if err == nil {
		t.Fatal("expected invalid probe spacing")
	}
	if next != current {
		t.Fatal("invalid edit replaced the last valid settings")
	}
}
