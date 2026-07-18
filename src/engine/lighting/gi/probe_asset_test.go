package gi

import (
	"bytes"
	"context"
	"errors"
	"math"
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func constantProbeAsset(color matrix.Vec3) ProbeAsset {
	asset := ProbeAsset{
		Bounds:     graviton.AABBFromMinMax(matrix.Vec3Zero(), matrix.Vec3One()),
		Spacing:    1,
		Dimensions: [3]uint32{2, 2, 2},
		Scenario:   "day",
		Probes:     make([]Probe, 8),
	}
	constantCoefficient := matrix.Float(math.Sqrt(4 * math.Pi))
	for z := uint32(0); z < 2; z++ {
		for y := uint32(0); y < 2; y++ {
			for x := uint32(0); x < 2; x++ {
				probe := &asset.Probes[asset.ProbeIndex(x, y, z)]
				probe.Position = matrix.NewVec3(x, y, z)
				probe.RadianceSH[0] = color.Scale(constantCoefficient)
				probe.MeanDistance = 10
				probe.DistanceVariance = 1
				probe.Validity = 1
			}
		}
	}
	return asset
}

func TestProbeAssetRoundTripAndSample(t *testing.T) {
	asset := constantProbeAsset(matrix.NewVec3(0.25, 0.5, 1))
	asset.GeometryHash[0] = 7
	data, err := asset.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := UnmarshalProbeAsset(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Scenario != asset.Scenario || decoded.GeometryHash != asset.GeometryHash || len(decoded.Probes) != 8 {
		t.Fatalf("round trip mismatch: %+v", decoded)
	}
	got := decoded.SampleIrradiance(matrix.Vec3Half(), matrix.Vec3Up())
	want := matrix.NewVec3(0.25*math.Pi, 0.5*math.Pi, math.Pi)
	if !matrix.Vec3ApproxTo(got, want, 0.001) {
		t.Fatalf("irradiance = %v, want %v", got, want)
	}
	if _, err := UnmarshalProbeAsset(data[:len(data)-1]); err == nil {
		t.Fatal("expected truncated asset error")
	}
}

func TestBakeProbesIsDeterministicAndCancelable(t *testing.T) {
	input := BakeInput{
		Bounds:         graviton.AABBFromMinMax(matrix.Vec3Zero(), matrix.Vec3One()),
		ProbeSpacing:   1,
		RaysPerProbe:   64,
		MaxRayDistance: 10,
		Environment:    matrix.NewVec3(0.1, 0.2, 0.3),
	}
	first, err := BakeProbes(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BakeProbes(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	firstBytes, _ := first.MarshalBinary()
	secondBytes, _ := second.MarshalBinary()
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("deterministic bake emitted different assets")
	}
	got := first.SampleIrradiance(matrix.Vec3Half(), matrix.Vec3Up())
	want := input.Environment.Scale(matrix.Float(math.Pi))
	if !matrix.Vec3ApproxTo(got, want, 0.02) {
		t.Fatalf("environment irradiance = %v, want %v", got, want)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := BakeProbes(canceled, input); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled bake error = %v", err)
	}
}

type probeAssetReader map[string][]byte

func (r probeAssetReader) Read(key string) ([]byte, error) {
	data, ok := r[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return data, nil
}

func TestBakedProviderLoadsScenarioAndAddsResolve(t *testing.T) {
	data, err := constantProbeAsset(matrix.Vec3One()).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	manager := NewManager(Capabilities{})
	manager.SetAssetReader(probeAssetReader{"day.kjgi": data})
	settings := SettingsForPreset(QualityPresetMedium)
	settings.Mode = ModeBaked
	if err := manager.Configure(settings); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetScenario("day.kjgi"); err != nil {
		t.Fatal(err)
	}
	graph := rendering.NewFrameGraph()
	depth, _ := graph.AddResource(rendering.FrameGraphResourceDescription{Name: "depth", Imported: true})
	outputs, err := manager.AddFramePasses(graph, FrameInputs{View: 3, Depth: depth, RuntimeSeconds: 4})
	if err != nil {
		t.Fatal(err)
	}
	if !outputs.Valid || outputs.DiffuseIrradiance == 0 {
		t.Fatalf("invalid baked outputs: %+v", outputs)
	}
	schedule, err := graph.Compile()
	if err != nil {
		t.Fatal(err)
	}
	context := &rendering.FrameGraphExecutionContext{}
	if err := schedule.Execute(context); err != nil {
		t.Fatal(err)
	}
	state, ok := context.Values["gi.baked.resolve.3"].(BakedResolveState)
	if !ok || state.Current == nil || state.Current.Scenario != "day" || state.Blend != 1 {
		t.Fatalf("resolve state = %#v", state)
	}
	stats := manager.Stats()
	if stats.ActiveProbes != 8 || stats.MemoryUsedMB != 1 || !stats.Converged {
		t.Fatalf("baked provider stats = %+v", stats)
	}
	shaderData := manager.ShaderData(matrix.Vec3Half(), 0)
	if shaderData.DimensionsCount != [4]int32{2, 2, 2, 8} || shaderData.Probes[0].PositionValidity.W() != 1 {
		t.Fatalf("shader probe window = %+v", shaderData)
	}
}

func TestBakedProviderCrossFadesCompatibleScenarios(t *testing.T) {
	dayAsset := constantProbeAsset(matrix.NewVec3(1, 0, 0))
	dayAsset.Scenario = "day"
	nightAsset := constantProbeAsset(matrix.NewVec3(0, 0, 1))
	nightAsset.Scenario = "night"
	day, err := dayAsset.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	night, err := nightAsset.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	manager := NewManager(Capabilities{})
	manager.SetAssetReader(probeAssetReader{"day.kjgi": day, "night.kjgi": night})
	settings := SettingsForPreset(QualityPresetMedium)
	settings.Mode = ModeBaked
	settings.ScenarioTransitionSeconds = 2
	if err := manager.Configure(settings); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetScenario("day.kjgi"); err != nil {
		t.Fatal(err)
	}
	manager.ShaderData(matrix.Vec3Half(), 9)
	if err := manager.SetScenario("night.kjgi"); err != nil {
		t.Fatal(err)
	}
	start := manager.ShaderData(matrix.Vec3Half(), 10)
	middle := manager.ShaderData(matrix.Vec3Half(), 11)
	if manager.Stats().Converged {
		t.Fatal("scenario transition reported convergence before the cross-fade completed")
	}
	finished := manager.ShaderData(matrix.Vec3Half(), 12)
	if !manager.Stats().Converged {
		t.Fatal("scenario transition did not report convergence after the cross-fade completed")
	}
	coefficient := matrix.Float(math.Sqrt(4 * math.Pi))
	if got := start.Probes[0].RadianceSH[0]; !matrix.Vec4ApproxTo(got, matrix.NewVec4(coefficient, 0, 0, 0), 0.001) {
		t.Fatalf("transition start coefficient = %v", got)
	}
	if got := middle.Probes[0].RadianceSH[0]; !matrix.Vec4ApproxTo(got, matrix.NewVec4(coefficient*0.5, 0, coefficient*0.5, 0), 0.001) {
		t.Fatalf("transition midpoint coefficient = %v", got)
	}
	if got := finished.Probes[0].RadianceSH[0]; !matrix.Vec4ApproxTo(got, matrix.NewVec4(0, 0, coefficient, 0), 0.001) {
		t.Fatalf("transition end coefficient = %v", got)
	}
}

func TestShaderProbeWindowHonorsUniformBudget(t *testing.T) {
	dimensions := shaderProbeWindowDimensions([3]uint32{20, 15, 10})
	count := dimensions[0] * dimensions[1] * dimensions[2]
	if count == 0 || count > rendering.MaxGIShaderProbes {
		t.Fatalf("shader window dimensions = %v (%d probes)", dimensions, count)
	}
}
