package content_database

import (
	"os"
	"path/filepath"
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/lighting/gi"
	"kaijuengine.com/matrix"
)

func TestGIProbeImportValidatesAsset(t *testing.T) {
	asset := gi.ProbeAsset{
		Bounds:     graviton.AABBFromMinMax(matrix.Vec3Zero(), matrix.NewVec3(1, 1, 1)),
		Spacing:    1,
		Dimensions: [3]uint32{2, 2, 2},
		Probes:     make([]gi.Probe, 8),
	}
	data, err := asset.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	valid := filepath.Join(t.TempDir(), "valid.kjgi")
	if err := os.WriteFile(valid, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (GIProbe{}).Import(valid, nil); err != nil {
		t.Fatalf("valid GI probe rejected: %v", err)
	}
	invalid := filepath.Join(t.TempDir(), "invalid.kjgi")
	if err := os.WriteFile(invalid, []byte("not probes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (GIProbe{}).Import(invalid, nil); err == nil {
		t.Fatal("invalid GI probe was accepted")
	}
}
