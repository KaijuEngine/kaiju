---
name: kaiju-mesh-drawing
description: 'Create, modify, debug, or explain Kaiju mesh rendering through rendering.Drawing, DrawInstance shader data, mesh/material/texture caches, instancing, culling, layers, transparency, PBR, imported models, and custom shaders.'
triggers:
    - draw a mesh in Kaiju
    - Kaiju Drawing
    - Kaiju DrawInstance
    - mesh material texture
    - mesh not rendering
    - material instance
    - shader data registry
    - render mesh
    - PBR drawing
    - transparent mesh
---

# Kaiju Mesh Drawing

Use this skill for any task that draws mesh geometry in Kaiju or diagnoses why a mesh is missing, mis-textured, incorrectly transformed, unlit, unbatched, or not cleaned up.

Kaiju's draw-instance system is not an entity renderer. A visible object is assembled explicitly from four independently owned pieces:

```text
rendering.Drawing
  Mesh        -> shared geometry from MeshCache
  Material    -> shared render state plus an ordered texture binding set
  ShaderData  -> one per logical drawing; implements rendering.DrawInstance
  Transform   -> optional entity/standalone transform followed by ShaderDataBase
```

`host.Drawings.AddDrawing` queues that record. Before rendering, `Drawings` groups it by render pass, root material/shader, material instance, mesh, and layer. Each group uploads the visible `DrawInstance` records to an instance buffer and issues an instanced mesh draw per render view.

## Start from the current implementation

Read the nearest existing caller and only the core files relevant to the task. The authoritative entry points are:

- `src/rendering/drawing.go`: `Drawing`, queueing, grouping, prepasses, render layers.
- `src/rendering/draw_instance.go`: `DrawInstance`, `ShaderDataBase`, culling, activation, destruction, instance-buffer copying.
- `src/rendering/material.go` and `material_cache.go`: material definitions, cached roots, texture-bound material instances.
- `src/rendering/mesh.go`, `mesh_cache.go`, and `vertex.go`: mesh construction, caching, bounds, deferred GPU creation, dynamic updates.
- `src/rendering/texture.go` and `texture_cache.go`: texture formats, filtering, generated textures, deferred GPU upload.
- `src/registry/shader_data_registry`: stock per-instance layouts and factories.
- `src/framework/mesh_drawing_maker.go` and `drawing_reader.go`: higher-level unlit/basic/PBR model helpers.

For concrete creation patterns, read [references/recipes.md](references/recipes.md). For a new material, shader, instance layout, skinning path, or bound buffer, read [references/custom-shaders-and-data.md](references/custom-shaders-and-data.md).

## Required invariants

Preserve these rules in every implementation:

1. Use `host.MeshCache()`, `host.MaterialCache()`, and `host.TextureCache()` for render resources. Their returned objects are shared and their GPU creation/upload is deferred to frame processing.
2. A `Drawing` must have non-nil `Material`, `Mesh`, and `ShaderData`. `AddDrawing` panics for a missing mesh/material; nil or incompatible shader data fails later.
3. Use one distinct `DrawInstance` per logical drawing. Do not attach the same shader-data pointer to multiple transforms: `AddDrawing` stores the drawing's transform inside the instance, so later submissions overwrite it.
4. Match shader data to the material's shader. Prefer `shader_data_registry.Create(mat.Shader.DrawInstanceDataName())` and verify the concrete type when the caller depends on one. A missing registry key logs a warning and silently returns basic shader data, so a non-nil result is not proof of compatibility. Only hardcode `"basic"`, `"unlit"`, or `"pbr"` when the material is known to use that layout.
5. Bind textures in the exact sampler order and count expected by the material/shader. `Material.CreateInstance(textures)` does not validate either. Call it on the cached root (or `mat.SelectRoot()`), not on an existing instance. The same texture objects in the same order reuse the cached material instance and batch together.
6. Attach `Transform: &entity.Transform` when the drawing should follow an entity. The draw system calls the unexported `setTransform` hook and derives `model = InitModel * transform.WorldMatrix()`.
7. Treat `DrawInstance` as the drawing's lifetime/visibility handle. Call `Destroy()` to remove it lazily, `Deactivate()` to hide it without removal, and `Activate()` to show it again. Entity activation/destruction does not automatically do this for arbitrary drawings; wire the events explicitly.
8. Keep the draw instance alive while the drawing is live. Store it on the owning object/entity when later mutation or cleanup needs access to it.
9. Give custom meshes stable, unique cache keys. `MeshCache.Mesh(key, ...)` returns the already cached mesh for that key and ignores replacement geometry.
10. Use Kaiju's `matrix` package and `matrix.Float`; do not introduce an external math library.

## Canonical opaque drawing

Use this as the default low-level pattern. Adapt the concrete mesh, material, texture list, and shader-data fields rather than rebuilding the renderer abstractions.

```go
mesh := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)

rootMat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
if err != nil {
	return err
}
diffuse, err := host.TextureCache().Texture(
	assets.TextureSquare, rendering.TextureFilterLinear)
if err != nil {
	return err
}
mat := rootMat.CreateInstance([]*rendering.Texture{diffuse})

sd := shader_data_registry.Create(mat.Shader.DrawInstanceDataName())
standard, ok := sd.(*shader_data_registry.ShaderDataStandard)
if !ok {
	return fmt.Errorf("material %q does not use standard shader data", mat.Id)
}
standard.Color = matrix.ColorWhite()

entity := engine.NewEntity(host.WorkGroup())
entity.StoreShaderData(sd)
entity.OnDestroy.Add(sd.Destroy)
entity.OnDeactivate.Add(sd.Deactivate)
entity.OnActivate.Add(sd.Activate)

host.Drawings.AddDrawing(rendering.Drawing{
	Material:   mat,
	Mesh:       mesh,
	ShaderData: sd,
	Transform:  &entity.Transform,
	ViewCuller: &host.Cameras.Primary,
})
```

If the surrounding function cannot return an error, handle cache errors consistently with that subsystem; do not silently keep nil resources.

## Choosing the path

- Built-in or procedural geometry: use a `rendering.NewMesh*` helper. These helpers cache the mesh.
- Raw vertices/indices: use `host.MeshCache().Mesh(uniqueKey, verts, indexes)`; use `DynamicMesh` only for frequent same-sized CPU vertex updates.
- OBJ/FBX/glTF `load_result.Result`: prefer `framework.CreateDrawingsUnlit`, `CreateDrawingsBasic`, or `CreateDrawingsPBR` when their material contract fits. Preserve each submesh/node transform and use `AddDrawings(result.AllDrawings())`.
- Serialized Kaiju mesh assets: read through `rendering/loaders/kaiju_mesh` and put the resulting vertices/indices in `MeshCache`.
- Stock material with defaults: the cached root material already contains its default textures. Use it directly if no override is needed, or call `CreateInstance` with the exact desired bindings.
- Custom textures: load asset-backed textures with `TextureCache.Texture`; insert encoded image bytes with `InsertImageTexture`; insert raw pixel bytes with `InsertRawTexture`.
- Transparent object: select the matching transparent material/render pass. Reducing alpha on an opaque material does not switch the render pass or depth/blend state.
- Many copies of one mesh: share `Mesh` and the same texture-bound `Material`; allocate one shader-data value and transform per object. The engine will batch compatible instances.
- Custom material/shader/data layout: follow the custom reference and keep the Go memory layout, generated shader layout, sampler list, and material definition synchronized.

## Stock contracts

| Material | Instance type | Texture slots |
| --- | --- | --- |
| `assets.MaterialDefinitionBasic` | `ShaderDataStandard` | diffuse |
| `assets.MaterialDefinitionUnlit` | `ShaderDataUnlit` | diffuse |
| `assets.MaterialDefinitionPBR` | `ShaderDataPBR` | base color, normal, metallic-roughness, emissive |
| transparent basic/unlit/PBR variants | corresponding basic/unlit/PBR data | same order as opaque counterpart |
| basic/PBR skinned variants | corresponding skinned data | same material-specific order; joint matrices use bound instance data |

PBR fallbacks are `TextureSquare`, `TexturePBRDefaultNormal`, `TexturePBRDefaultMetallicRough`, and `TextureBlankSquare` in that order. `ShaderDataPBR.MeRoEmAo` is the per-instance metallic/roughness/emissive/AO parameter vector; the stock shader currently consumes its RGB components. Initialize it to `(1, 1, 0, 1)` unless the requested look calls for other values. Initialize `LightIds` to `-1`; `SelectLights` populates them.

## Transform, culling, layers, and ordering

- With a transform, changing it marks it dirty; the draw instance refreshes its model matrix and transforms `Mesh.Bounds()` for culling.
- Without a transform, `ShaderDataBase.SetModel` or `ModelPtr` owns the model matrix. This is appropriate for direct-model systems such as particles.
- `ViewCuller: &host.Cameras.Primary` is the normal fallback. A render view's camera supplies per-view culling where available. Nil disables fallback culling.
- Layer `0` normalizes to `RenderLayerWorld`. Other standard masks are UI, editor, editor picking, and editor gizmo picking. The render view must include the layer.
- `Sort` orders instance groups, not individual instances. Compatible drawings with the same mesh/material/layer join one group; the first-created group's sort/culler applies. Do not rely on different `Sort` values to order otherwise compatible instances independently.

## Lifetime and mutation

Material and mesh objects are shared cache resources; normally do not destroy them per entity. The per-drawing shader data is what gets destroyed.

```go
entity.OnDestroy.Add(sd.Destroy)
entity.OnDeactivate.Add(sd.Deactivate)
entity.OnActivate.Add(sd.Activate)
```

Destroying marks an instance. The renderer compacts it out during a later instance update and destroys an empty group. Mutating exported shader-data fields is observed on the next frame because visible instance bytes are recopied each frame; there is no separate dirty call for ordinary per-instance fields.

When replacing a material or mesh on an existing object, create a compatible new shader-data instance and submit a new `Drawing`, then destroy the old shader data. A submitted `Drawing` is a value copied into the draw system; mutating that old local `Drawing` struct does not rebind the queued group.

## Verification

For ordinary code changes, run focused package tests plus `go test ./rendering ./framework` from `src` when those packages are affected. For visual behavior, follow the repository integration-test instructions in `AGENTS.md`: register a test in `src/integration_testing`, build with `debug,editor,filedrop,rawsrc`, render after several frames, capture a screenshot or short video, and destroy/stop resources before exit.

When a mesh is invisible, check in this order: cache errors/nil pointers, material-to-shader-data match, texture count/order, camera-facing transform, layer/view mask, `Deactivate`/`Destroy` state, mesh bounds/culling, transparent-vs-opaque material, then shader/instance byte-layout agreement.
