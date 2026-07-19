# Global illumination

Kaiju's global illumination (GI) system is provider based. Game code configures
the `gi.Manager`; rendering code consumes provider-neutral irradiance. A new GI
technique can therefore replace the current implementation without changing
materials, scenes, cameras, or game code.

The production baseline is deterministic baked irradiance probes. It works on
every Vulkan device supported by Kaiju and has no per-frame ray-tracing cost.
The renderer keeps a hardware DDGI provider slot for machines that expose an
enabled Vulkan 1.2 ray-query stack. Kaiju's current Vulkan binding does not yet
enable that stack, so `Auto`, `High`, and `Ultra` safely fall back to baked
probes instead of advertising a feature the device was not created with.

## Runtime setup

GI defaults to off. Configure it after creating the host and load a baked
scenario before rendering the stage:

```go
import "kaijuengine.com/engine/lighting/gi"

settings := gi.SettingsForPreset(gi.QualityPresetMedium)
settings.Mode = gi.ModeAuto

if err := host.GlobalIllumination().Configure(settings); err != nil {
	return err
}
if err := host.GlobalIllumination().SetScenario("lighting/day.kjgi"); err != nil {
	return err
}
```

`ModeAuto` selects the best registered provider whose required GPU features
were actually enabled. `FallbackAllow` is the default. Use
`FallbackRequireExact` during validation when silently selecting a lower tier
would hide a configuration error.

To transition between lighting scenarios, load another asset. Compatible probe
grids cross-fade over `ScenarioTransitionSeconds`; incompatible grids switch
immediately:

```go
host.GlobalIllumination().SetScenario("lighting/night.kjgi")
```

Inspect `host.GlobalIllumination().Stats()` to show the active provider, probe
and memory counts, convergence state, and any fallback reason in developer
tooling. Providers that schedule GPU work also populate the timing fields.

## Quality controls

Presets are starting points. Every value is public and can be overridden for a
project or platform profile.

| Preset | Intended path | GI budget | Memory | Probe spacing | Resolve scale | Update rate |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Off | Null provider | 0 ms | 0 MB | -- | -- | -- |
| Low | Baked probes | 0.5 ms | 48 MB | 4 m | 50% | 15 Hz |
| Medium | Baked probes | 1.0 ms | 96 MB | 2 m | 50% | 30 Hz |
| High | DDGI when available, otherwise baked | 1.5 ms | 160 MB | 2 m | 50% | 60 Hz |
| Ultra | DDGI when available, otherwise baked | 3.0 ms | 256 MB | 1.5 m | 100% | 60 Hz |

The principal knobs are:

- `GPUTimeBudgetMS` and `MemoryBudgetMB` set hard product targets.
- `ProbeSpacing` and `CoverageDistance` trade detail for coverage and memory.
- `RaysPerProbe` and `MaxProbeUpdatesPerFrame` control dynamic-provider work.
- `ResolveScale`, `UpdateHz`, and `HistoryWeight` control resolve cost and
  temporal stability.
- `ContactDetail`, `DynamicGeometry`, and `EmissiveParticipation` control
  optional sources of detail.
- `AdaptiveBudget` asks a dynamic provider to use the included fast-reduction,
  hysteretic-recovery budget controller when measured GI time exceeds the
  target. It does not change a fully baked field.

The baked provider uploads a camera-local window of probes to the global shader
uniform. The window is capped so the global block remains inside Vulkan's
minimum guaranteed uniform-buffer range. Large levels can therefore use large
bakes without making every probe resident in every draw.

## Baking `.kjgi` assets

`kaiju-gi-bake` is a deterministic CPU reference baker. It uses Kaiju's BVH,
supports environment, emissive, directional, and point-light contribution,
and emits versioned `.kjgi` assets containing RGB L2 spherical harmonics plus
distance moments.

Build and invoke it from `src`:

```text
go run ./cmd/kaiju-gi-bake -input lighting/day.json -output game_content/lighting/day.kjgi
```

## Editor authoring and preview

Project-wide defaults are available under **Settings > Global Illumination**.
The Stage workspace's **GI** button opens the live stage panel without hiding
the viewport. A stage can inherit the project defaults or copy them into a
complete override, select/import a `.kjgi` scenario, and inspect the provider,
fallback reason, probe count, memory use, GPU time, and bake freshness.

The stage panel can bake automatic scene bounds or a manual volume. It captures
contributed static meshes, terrain, directional lights, point lights, spot
lights, supported material tint/emissive values, and the configured environment
radiance. Baking runs in the background and can be cancelled. The previously
assigned asset is not changed unless the new bake validates and is written
successfully. Geometry and lighting hashes identify stale scenarios after an
author changes the stage.

Each entity has a **Global illumination** contribution setting in Details:

- **Automatic** treats ordinary meshes and terrain as static contributors.
- **Excluded** omits the entity from GI.
- **Static** explicitly includes its undeformed mesh in a bake.
- **Rigid** and **Receives only** are excluded from baked geometry and retain
  their meaning for a future dynamic DDGI provider.

Project GI defaults and per-stage overrides are serialized into debug and
release builds. Loading a stage with no scenario clears the previous stage's
baked probes, so lighting never leaks between stages.

Minimal input:

```json
{
  "bounds": {"min": [-10, 0, -10], "max": [10, 6, 10]},
  "probeSpacing": 2,
  "raysPerProbe": 256,
  "maxRayDistance": 100,
  "scenario": "day",
  "environment": [0.04, 0.06, 0.1],
  "triangles": [],
  "directionalLights": [
    {"towardLight": [0.3, 0.8, 0.2], "radiance": [4, 3.8, 3.4]}
  ],
  "pointLights": []
}
```

Each triangle has `points`, `albedo`, and `emissive` fields. Optional
`geometryHash` and `lightingHash` values are 64 hexadecimal characters; the
baker computes stable SHA-256 values when they are omitted. Unknown JSON fields
are rejected so misspelled bake settings cannot silently enter a build.

Use the same bounds, spacing, and dimensions for scenarios that must cross-fade.
Keep probe bounds just outside playable space, and prefer increasing bake rays
before shrinking spacing when the problem is noise rather than spatial detail.

## Renderer contract

The renderer provides linear HDR opaque color, sampleable depth,
normal/roughness, albedo/metallic, and camera-motion targets. Tone mapping is a
single final-combine operation. PBR, basic, and terrain materials sample the
same provider-neutral diffuse irradiance data.

Each render view owns stable current/previous camera matrices and a history
reset flag. Call `RenderView.ResetHistory()` after camera cuts, teleports, or
other discontinuities. Replacing a view camera resets history automatically.

Providers implement `gi.Provider`. The manager owns provider lifecycle,
capability selection, scene invalidation, per-view history, frame-graph pass
registration, scenario loading, statistics, and shutdown. A replacement
provider should never modify material shaders; it should produce `gi.Outputs`
and a `ProbeFieldBinding` through this contract.

## Current hardware DDGI boundary

A KHR ray-query DDGI implementation requires all of the following to be enabled
during Vulkan device creation: buffer device address, deferred host operations,
acceleration structures, and ray query. Extension names alone are not enough.
Kaiju reports advertised and enabled support separately and currently keeps the
enabled flags false because its vendored Vulkan header exposes the older NVX
prototype rather than the required KHR API.

The safe sequence for adding the DDGI provider is to upgrade the Vulkan binding,
chain and enable the feature structs, build/update BLAS and TLAS data, add
ray-query probe trace/update shaders, register `ProviderDDGI`, and retain the
existing baked/null fallbacks. The provider API, capability gate, frame graph,
adaptive budget controller, G-buffer inputs, temporal state, and presets are
already in place for that work.
