package gi

import "testing"

func TestBudgetControllerShedsLoadAndRecoversSlowly(t *testing.T) {
	settings := SettingsForPreset(QualityPresetHigh)
	controller := NewBudgetController(settings)
	decision := controller.Update(settings.GPUTimeBudgetMS * 3)
	if !decision.OverBudget || decision.MaxProbeUpdates >= settings.MaxProbeUpdatesPerFrame {
		t.Fatalf("over-budget decision = %+v", decision)
	}
	reduced := decision.MaxProbeUpdates
	for range 29 {
		decision = controller.Update(0)
		if decision.MaxProbeUpdates > reduced {
			t.Fatalf("controller recovered before hysteresis window: %+v", decision)
		}
		reduced = min(reduced, decision.MaxProbeUpdates)
	}
	for range 60 {
		decision = controller.Update(0)
	}
	if decision.MaxProbeUpdates <= reduced {
		t.Fatalf("controller did not recover after sustained headroom: %+v", decision)
	}
}

func TestBudgetControllerCanBeDisabled(t *testing.T) {
	settings := SettingsForPreset(QualityPresetHigh)
	settings.AdaptiveBudget = false
	controller := NewBudgetController(settings)
	decision := controller.Update(settings.GPUTimeBudgetMS * 100)
	if decision.OverBudget || decision.MaxProbeUpdates != settings.MaxProbeUpdatesPerFrame || decision.ResolveScale != settings.ResolveScale {
		t.Fatalf("disabled controller changed quality: %+v", decision)
	}
}
