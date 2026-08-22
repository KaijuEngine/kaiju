# Custom Materials, Shaders, and Draw-Instance Data

Read this reference when the requested mesh cannot use an existing basic, unlit, PBR, terrain, sprite, text, or editor material unchanged.

## The asset chain

A material definition connects four contracts:

```text
*.material
  Shader         -> *.shader (stages, reflected layouts, sampler labels)
  RenderPass     -> *.renderpass (attachments and pass ordering)
  ShaderPipeline -> *.shaderpipeline (topology, culling, depth, blending)
  Textures[]     -> ordered default sampler bindings
```

Stock examples live under `src/editor/editor_embedded_content/editor_content/renderer/` in `materials`, `shaders`, `renderpasses`, `pipelines`, `src`, and `spv`. Start by copying the closest semantic example, then change only the required contract.

The material cache loads the asset by key and compiles/caches its shader, render pass, pipeline, and default textures. Transparent behavior comes from the transparent render pass and pipeline; naming a material `_transparent` is a convention/helper signal, not what configures blending.

## DrawInstanceData selects the Go factory

`Shader.DrawInstanceDataName()` returns the `.shader` file's `DrawInstanceData` value, or falls back to the shader `Name` when the field is blank. Therefore:

- Reusing an existing Go layout under a differently named shader: set `"DrawInstanceData":"pbr"` (or another registered key).
- Defining a new layout: add a matching factory in `src/registry/shader_data_registry` and set `DrawInstanceData` to that registered key.
- Runtime creation should use `shader_data_registry.Create(mat.Shader.DrawInstanceDataName())` so material overrides and generated shaders select the right layout.

The registry's `register` function is package-private. New registry entries belong in a Go file inside `src/registry/shader_data_registry`, normally registered from `init`. `Create` warns and falls back to basic data when a name is absent; type-assert or otherwise verify the returned layout. Some legacy shader assets rely on aliases or family-specific callers, so do not assume a blank `DrawInstanceData` is correct merely because drawing creation did not return an error.

## Implementing an instance layout

The ordinary pattern embeds an initialized `rendering.ShaderDataBase`, then places GPU instance fields in the exact order expected by the vertex shader:

```go
type ShaderDataHeat struct {
	rendering.ShaderDataBase `visible:"false"`

	Tint      matrix.Color
	HeatRange matrix.Vec2
	Flags     uint32
}

func newShaderDataHeat() rendering.DrawInstance {
	return &ShaderDataHeat{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Tint:           matrix.ColorWhite(),
	}
}

func (ShaderDataHeat) Size() int {
	return int(rendering.ShaderBaseDataSize +
		unsafe.Sizeof(ShaderDataHeat{}.Tint) +
		unsafe.Sizeof(ShaderDataHeat{}.HeatRange) +
		unsafe.Sizeof(ShaderDataHeat{}.Flags))
}

func init() {
	register(newShaderDataHeat, "heat")
}
```

`ShaderDataBase.DataPointer()` points at its model matrix. The renderer copies `Size()` bytes starting there, so the bytes following the embedded base must be the shader's per-instance attributes. `ShaderDataBase` also supplies lifecycle, transform, culling, default no-op lighting, and default no-bound-buffer behavior.

Critical layout rules:

- Construct the base with `rendering.NewShaderDataBase()`; the zero value does not initialize identity model matrices.
- Keep the Go field order, scalar widths, array lengths, and shader instance attribute order identical.
- Include `rendering.ShaderBaseDataSize` plus only bytes copied from the model onward. Follow existing registry types and use `unsafe.Sizeof` rather than hand-written byte constants.
- Account for Go alignment/padding and Vulkan attribute alignment. After generating the shader metadata, compare its reflected instance stride to `Size()` and add a focused regression test when the layout is new or unusual.
- Keep pointer, slice, map, string, and interface fields out of the copied instance-data region. Put CPU-only headers before `ShaderDataBase` as the skinned layouts do, or elsewhere outside the copied range.
- Do not mix different instance layouts inside drawings that resolve to the same material/mesh/layer group.

If the layout uses the stock outline flags, use `shader_data_registry.StandardShaderDataFlags` and its set/clear helpers so the enable bit remains correct.

## GLSL and generated shader metadata

Kaiju's shader source uses `kaiju.glsl` macros and reflected vertex-instance inputs. Inspect the closest stock pair, such as `basic.vert`/`basic.frag`, `unlit.vert`/`unlit.frag`, or `pbr.vert`/`pbr.frag`.

The checked-in `*.shader` files contain generated `LayoutGroups` and SPIR-V output names. The stock generator is `src/generators/spirv`; its `prebuilt.json` determines which shaders are regenerated. Run it from `src` with `go run ./generators/spirv` when intentionally rebuilding that stock list, with Vulkan SDK `glslc` available. Review all rewritten shader metadata and SPIR-V artifacts rather than assuming only one file changed.

For a new/custom build path, follow the owning subsystem's compiler workflow instead of adding the shader to the stock prebuilt list automatically.

## Texture bindings

Three lists must stay synchronized:

1. sampler declarations/count and indexes used in GLSL;
2. `.shader` `SamplerLabels` order;
3. `.material` `Textures` order, and every `CreateInstance(textures)` call.

Material labels are descriptive metadata; the slice position controls the binding. Add semantic fallback textures, not one generic white texture for every PBR slot. Check `framework.textureFallbackForSlot` for the stock PBR choices.

`Material.CreateInstance` caches by ordered texture keys and object addresses, clones the slice, shallow-copies material flags/state, and points `Root` at the cached root. Call it on the cached root or `SelectRoot()`; an instance does not own an instance-cache map. It does not compile a new shader or validate sampler compatibility.

## Render pass and pipeline choices

- Opaque: depth test/write and opaque render pass/pipeline.
- Transparent: transparent pass plus blending and normally no depth write. Use the closest stock transparent definition.
- Lines/points/patches: select matching input topology in the shader pipeline; a mesh's data alone does not change topology.
- Double-sided: use a pipeline with culling disabled rather than trying to solve it in shader data.
- Picking/depth/prepass: use a dedicated material and distinct compatible shader data. `Drawing.Material.PrepassMaterial` causes `AddDrawing` to queue a companion prepass automatically.
- Shadows: setting `CastsShadows` causes `PreparePending` to create depth-pass shadow instances. Their lifecycle follows the source instance. `ReceivesShadows`/`IsLit` affect material/shader behavior and should agree with the selected shader.

Avoid mutating a cached root material's render flags or texture slice for one object. Use a material instance for texture-specific/per-object binding state, or define/cache a separate material when pipeline/pass behavior differs.

## Skinning and bound instance data

Skinned instance types embed `rendering.SkinnedShaderDataHeader` before `ShaderDataBase`. They override:

- `SkinningHeader()` to expose bones;
- `InstanceBoundDataSize()` for the joint-matrix block;
- `BoundDataPointer()` for its memory;
- `UpdateBoundData()` to refresh matrices.

Study `shader_data_standard_skinned.go`, `shader_data_pbr_skinned.go`, and `skinned_shader_data_header.go` before changing skinning. The material shader layout must declare the matching buffer binding/capacity. Do not put the joint matrix block in ordinary copied instance attributes.

## Custom culling and companion draws

`Drawing.ViewCuller` implements:

```go
type ViewCuller interface {
	IsInView(graviton.AABB) bool
	ViewChanged() bool
}
```

The default path transforms `Mesh.Bounds()` by the draw instance's model. A culler must return `ViewChanged` when its view state changes or culling may remain cached. Render-view cameras can override the fallback on a per-view basis; UI-layer groups deliberately retain their fallback culler.

For a companion draw such as picking or an outline, create independent shader data and call:

```go
rendering.LinkDrawInstanceLifecycle(visibleData, companionData)
```

This propagates activation, deactivation, and destruction. It does not synchronize arbitrary shader fields.

## Validation checklist

- Material asset resolves through the active `assets.Database`.
- Shader, render pass, and pipeline assets compile without swallowed errors.
- Shader `DrawInstanceDataName()` resolves to the intended registered factory.
- Go `DrawInstance.Size()` and field order match the generated instance layout.
- Sampler count and order match across GLSL, `.shader`, `.material`, and call sites.
- Opaque/transparent, topology, cull, and depth states match the intended geometry.
- One shader-data instance is allocated per logical drawing and wired for cleanup.
- Focused Go tests cover layout/factory behavior; a visual integration test covers the actual Vulkan draw.
