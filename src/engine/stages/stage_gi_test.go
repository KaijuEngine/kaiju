package stages

import (
	"bytes"
	"testing"

	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/lighting/gi"
)

func TestStageGlobalIlluminationRoundTrips(t *testing.T) {
	settings := gi.SettingsForPreset(gi.QualityPresetHigh)
	stage := Stage{
		Id: "stage",
		GlobalIllumination: StageGlobalIllumination{
			OverrideProjectSettings: true,
			Settings:                settings,
			ProbeAsset:              "day.kjgi",
			BakeSettings:            DefaultStageGIBakeSettings(settings),
		},
		Entities: []EntityDescription{{Id: "entity", GIContribution: GIContributionRigid}},
	}
	jsonStage := stage.ToMinimized()
	var decoded Stage
	decoded.FromMinimized(jsonStage)
	if decoded.GlobalIllumination.ProbeAsset != "day.kjgi" || decoded.GlobalIllumination.Settings.Preset != gi.QualityPresetHigh {
		t.Fatalf("debug stage GI did not round trip: %+v", decoded.GlobalIllumination)
	}
	if decoded.Entities[0].GIContribution != GIContributionRigid {
		t.Fatalf("GI contribution = %v", decoded.Entities[0].GIContribution)
	}
	var buffer bytes.Buffer
	if err := pod.NewEncoder(&buffer).Encode(stage); err != nil {
		t.Fatal(err)
	}
	archive, err := ArchiveDeserializer(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if archive.GlobalIllumination.ProbeAsset != "day.kjgi" || archive.Entities[0].GIContribution != GIContributionRigid {
		t.Fatalf("archive stage GI did not round trip: %+v", archive)
	}
}

func TestOldStageDefaultsToProjectInheritance(t *testing.T) {
	var stage Stage
	stage.FromMinimized(StageJson{Id: "old"})
	if stage.GlobalIllumination.OverrideProjectSettings || stage.GlobalIllumination.ProbeAsset != "" {
		t.Fatalf("old stage unexpectedly overrides GI: %+v", stage.GlobalIllumination)
	}
	if got := (EntityDescription{}).GIContribution; got != GIContributionAutomatic {
		t.Fatalf("zero GI contribution = %v, want automatic", got)
	}
}
