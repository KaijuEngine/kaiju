package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"kaijuengine.com/engine/lighting/gi"
)

func TestRunBakesProbeAsset(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "scene.json")
	outputPath := filepath.Join(directory, "day.kjgi")
	input := `{
		"bounds":{"min":[0,0,0],"max":[1,1,1]},
		"probeSpacing":1,
		"raysPerProbe":32,
		"maxRayDistance":10,
		"scenario":"day",
		"environment":[0.1,0.2,0.3]
	}`
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), inputPath, outputPath); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	asset, err := gi.UnmarshalProbeAsset(data)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Scenario != "day" || len(asset.Probes) != 8 {
		t.Fatalf("baked asset = %+v", asset)
	}
}
