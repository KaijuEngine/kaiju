/******************************************************************************/
/* baked_provider.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type AssetReader interface {
	Read(key string) ([]byte, error)
}

// BakedResolveState is the provider-neutral payload consumed by the renderer's
// baked probe resolve implementation.
type BakedResolveState struct {
	Previous *ProbeAsset
	Current  *ProbeAsset
	Blend    float32
}

type BakedProbeProvider struct {
	mutex             sync.RWMutex
	reader            AssetReader
	settings          Settings
	current           *ProbeAsset
	previous          *ProbeAsset
	transitionPending bool
	transitionStart   float32
	fields            map[ViewID]ProbeFieldBinding
	stats             Stats
}

func (*BakedProbeProvider) ID() string                 { return ProviderBakedProbe }
func (*BakedProbeProvider) Supports(Capabilities) bool { return true }

func (p *BakedProbeProvider) Initialize(context ProviderContext) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.reader = context.Assets
	p.fields = make(map[ViewID]ProbeFieldBinding)
	p.stats.Provider = ProviderBakedProbe
	return nil
}

func (p *BakedProbeProvider) Configure(settings Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	p.mutex.Lock()
	p.settings = settings
	p.mutex.Unlock()
	return nil
}

func (*BakedProbeProvider) SyncScene(SceneDelta) error { return nil }
func (*BakedProbeProvider) AddUpdatePasses(*rendering.FrameGraph, FrameInputs) error {
	return nil
}

func (p *BakedProbeProvider) AddResolvePasses(graph *rendering.FrameGraph, inputs FrameInputs) (Outputs, error) {
	if graph == nil {
		return Outputs{}, errors.New("baked GI resolve requires a frame graph")
	}
	p.mutex.Lock()
	if p.transitionPending {
		p.transitionStart = inputs.RuntimeSeconds
		p.transitionPending = false
	}
	blend := p.transitionBlendLocked(inputs.RuntimeSeconds)
	current, previous := p.current, p.previous
	p.mutex.Unlock()
	resource, err := graph.AddResource(rendering.FrameGraphResourceDescription{
		Name:    fmt.Sprintf("gi.baked.diffuse.%d", inputs.View),
		Kind:    rendering.FrameGraphResourceImage,
		PerView: true,
	})
	if err != nil {
		return Outputs{}, err
	}
	uses := frameInputReads(inputs)
	uses = append(uses, rendering.FrameGraphResourceUse{Resource: resource, Access: rendering.FrameGraphAccessWrite})
	state := BakedResolveState{Previous: previous, Current: current, Blend: blend}
	_, err = graph.AddPass(rendering.FrameGraphPassDescription{
		Name:  fmt.Sprintf("GI Baked Probe Resolve %d", inputs.View),
		Queue: rendering.FrameGraphQueueCompute,
		Uses:  uses,
		Execute: func(context *rendering.FrameGraphExecutionContext) error {
			context.Values[fmt.Sprintf("gi.baked.resolve.%d", inputs.View)] = state
			return nil
		},
	})
	if err != nil {
		return Outputs{}, err
	}
	binding := ProbeFieldBinding{Irradiance: resource, Valid: current != nil}
	p.mutex.Lock()
	p.fields[inputs.View] = binding
	p.stats.ActiveProbes = 0
	if current != nil {
		p.stats.ActiveProbes = uint32(len(current.Probes))
	}
	p.mutex.Unlock()
	return Outputs{DiffuseIrradiance: resource, Valid: true}, nil
}

func frameInputReads(inputs FrameInputs) []rendering.FrameGraphResourceUse {
	resources := [...]rendering.FrameGraphResource{
		inputs.DirectLighting, inputs.Depth, inputs.NormalRoughness,
		inputs.AlbedoMetallic, inputs.Motion,
	}
	seen := make(map[rendering.FrameGraphResource]struct{}, len(resources))
	uses := make([]rendering.FrameGraphResourceUse, 0, len(resources)+1)
	for i := range resources {
		if resources[i] == 0 {
			continue
		}
		if _, exists := seen[resources[i]]; exists {
			continue
		}
		seen[resources[i]] = struct{}{}
		uses = append(uses, rendering.FrameGraphResourceUse{Resource: resources[i], Access: rendering.FrameGraphAccessRead})
	}
	return uses
}

func (p *BakedProbeProvider) ProbeField(view ViewID) ProbeFieldBinding {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.fields[view]
}

func (p *BakedProbeProvider) ShaderData(position matrix.Vec3, runtimeSeconds float32) rendering.GlobalIlluminationForRender {
	p.mutex.Lock()
	if p.transitionPending {
		p.transitionStart = runtimeSeconds
		p.transitionPending = false
	}
	blend := p.transitionBlendLocked(runtimeSeconds)
	current, previous := p.current, p.previous
	p.mutex.Unlock()

	currentData := shaderDataForProbeAsset(current, position)
	if previous == nil || blend >= 1 {
		return currentData
	}
	previousData := shaderDataForProbeAsset(previous, position)
	if previousData.DimensionsCount != currentData.DimensionsCount ||
		previousData.BoundsMinSpacing != currentData.BoundsMinSpacing {
		return currentData
	}
	blendFactor := matrix.Float(blend)
	for probe := 0; probe < int(currentData.DimensionsCount[3]); probe++ {
		currentData.Probes[probe].PositionValidity = matrix.Vec4Lerp(
			previousData.Probes[probe].PositionValidity,
			currentData.Probes[probe].PositionValidity,
			blendFactor,
		)
		currentData.Probes[probe].DistanceMoments = matrix.Vec4Lerp(
			previousData.Probes[probe].DistanceMoments,
			currentData.Probes[probe].DistanceMoments,
			blendFactor,
		)
		for coefficient := range currentData.Probes[probe].RadianceSH {
			currentData.Probes[probe].RadianceSH[coefficient] = matrix.Vec4Lerp(
				previousData.Probes[probe].RadianceSH[coefficient],
				currentData.Probes[probe].RadianceSH[coefficient],
				blendFactor,
			)
		}
	}
	return currentData
}

func shaderDataForProbeAsset(asset *ProbeAsset, position matrix.Vec3) rendering.GlobalIlluminationForRender {
	if asset == nil || len(asset.Probes) == 0 {
		return rendering.GlobalIlluminationForRender{}
	}
	dimensions := shaderProbeWindowDimensions(asset.Dimensions)
	minPoint := asset.Bounds.Min()
	start := [3]uint32{}
	for axis := range 3 {
		grid := int(math.Floor(float64((position[axis] - minPoint[axis]) / asset.Spacing)))
		grid -= int(dimensions[axis]) / 2
		maxStart := max(0, int(asset.Dimensions[axis]-dimensions[axis]))
		start[axis] = uint32(max(0, min(grid, maxStart)))
	}
	windowMin := minPoint.Add(matrix.Vec3{
		matrix.Float(start[0]) * asset.Spacing,
		matrix.Float(start[1]) * asset.Spacing,
		matrix.Float(start[2]) * asset.Spacing,
	})
	result := rendering.GlobalIlluminationForRender{
		BoundsMinSpacing: windowMin.AsVec4WithW(asset.Spacing),
		DimensionsCount: [4]int32{
			int32(dimensions[0]), int32(dimensions[1]), int32(dimensions[2]),
			int32(dimensions[0] * dimensions[1] * dimensions[2]),
		},
	}
	destination := 0
	for z := uint32(0); z < dimensions[2]; z++ {
		for y := uint32(0); y < dimensions[1]; y++ {
			for x := uint32(0); x < dimensions[0]; x++ {
				probe := asset.Probes[asset.ProbeIndex(start[0]+x, start[1]+y, start[2]+z)]
				gpuProbe := &result.Probes[destination]
				gpuProbe.PositionValidity = probe.Position.AsVec4WithW(probe.Validity)
				gpuProbe.DistanceMoments = matrix.Vec4{probe.MeanDistance, probe.DistanceVariance, 0, 0}
				for coefficient := range gpuProbe.RadianceSH {
					gpuProbe.RadianceSH[coefficient] = probe.RadianceSH[coefficient].AsVec4WithW(0)
				}
				destination++
			}
		}
	}
	return result
}

func shaderProbeWindowDimensions(assetDimensions [3]uint32) [3]uint32 {
	dimensions := assetDimensions
	product := func() uint64 {
		return uint64(dimensions[0]) * uint64(dimensions[1]) * uint64(dimensions[2])
	}
	for product() > rendering.MaxGIShaderProbes {
		axis := 0
		for i := 1; i < 3; i++ {
			if dimensions[i] > dimensions[axis] {
				axis = i
			}
		}
		if dimensions[axis] <= 1 {
			break
		}
		dimensions[axis]--
	}
	return dimensions
}

func (*BakedProbeProvider) Invalidate(Invalidation) {}
func (*BakedProbeProvider) ResetHistory(ViewID)     {}

func (p *BakedProbeProvider) SetScenario(assetKey string) error {
	if assetKey == "" {
		p.mutex.Lock()
		p.current = nil
		p.previous = nil
		p.transitionPending = false
		clear(p.fields)
		p.stats.ActiveProbes = 0
		p.mutex.Unlock()
		return nil
	}
	p.mutex.RLock()
	reader := p.reader
	p.mutex.RUnlock()
	if reader == nil {
		return errors.New("baked GI has no asset reader")
	}
	data, err := reader.Read(assetKey)
	if err != nil {
		return fmt.Errorf("read baked GI scenario %q: %w", assetKey, err)
	}
	asset, err := UnmarshalProbeAsset(data)
	if err != nil {
		return fmt.Errorf("decode baked GI scenario %q: %w", assetKey, err)
	}
	p.mutex.Lock()
	p.previous = p.current
	p.current = &asset
	p.transitionPending = p.previous != nil && p.settings.ScenarioTransitionSeconds > 0
	if !p.transitionPending {
		p.previous = nil
	}
	p.mutex.Unlock()
	return nil
}

func (p *BakedProbeProvider) transitionBlendLocked(runtime float32) float32 {
	if p.previous == nil || p.settings.ScenarioTransitionSeconds <= 0 {
		return 1
	}
	blend := (runtime - p.transitionStart) / p.settings.ScenarioTransitionSeconds
	if blend >= 1 {
		p.previous = nil
		return 1
	}
	return max(0, blend)
}

func (p *BakedProbeProvider) Stats() Stats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	stats := p.stats
	if p.current == nil {
		stats.FallbackReason = "no baked probe scenario loaded"
		stats.Converged = true
		return stats
	}
	stats.ActiveProbes = uint32(len(p.current.Probes))
	memoryBytes := uint64(len(p.current.Probes)) * probeAssetFloatsPerProbe * 4
	if p.previous != nil {
		memoryBytes += uint64(len(p.previous.Probes)) * probeAssetFloatsPerProbe * 4
	}
	const bytesPerMegabyte = uint64(1024 * 1024)
	stats.MemoryUsedMB = uint32((memoryBytes + bytesPerMegabyte - 1) / bytesPerMegabyte)
	stats.Converged = p.previous == nil
	return stats
}

func (*BakedProbeProvider) DebugViews() []DebugView {
	return []DebugView{
		{ID: "probe_validity", Label: "Probe Validity"},
		{ID: "probe_irradiance", Label: "Probe Irradiance"},
		{ID: "probe_distance", Label: "Probe Distance"},
	}
}

func (p *BakedProbeProvider) Shutdown() {
	p.mutex.Lock()
	p.current = nil
	p.previous = nil
	p.fields = nil
	p.mutex.Unlock()
}
