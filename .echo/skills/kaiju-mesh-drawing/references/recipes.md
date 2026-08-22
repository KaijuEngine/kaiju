# Mesh Drawing Recipes

Read the sections relevant to the requested drawing. These recipes assume a valid `*engine.Host` whose renderer and asset database have been initialized.

## Built-in and procedural meshes

The `rendering.NewMesh*` helpers create or retrieve cached meshes. Common choices include:

```go
quad := rendering.NewMeshQuad(host.MeshCache())
plane := rendering.NewMeshPlane(host.MeshCache())
cube := rendering.NewMeshCube(host.MeshCache())
texturedCube := rendering.NewMeshTexturableCube(host.MeshCache())
sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
capsule := rendering.NewMeshCapsule(host.MeshCache(), 0.5, 1, 32, 8)
cylinder := rendering.NewMeshCylinder(host.MeshCache(), 1, 0.5, 32, true)
```

Prefer `NewMeshTexturableCube` when correct per-face UVs/normals matter. Inspect `src/rendering/mesh.go` for the exact current parameter order and other wire/debug primitives.

## Custom static geometry

`rendering.Vertex` supplies the fixed base vertex attributes: `Position`, `Normal`, `Tangent`, `UV0`, `Color`, `JointIds`, `JointWeights`, and `MorphTarget`. Populate every attribute the selected shader consumes. White vertex color is important for stock shaders because it multiplies the per-instance/material color.

```go
verts := []rendering.Vertex{
	{Position: matrix.Vec3{-0.5, -0.5, 0}, Normal: matrix.Vec3{0, 0, 1}, UV0: matrix.Vec2{0, 1}, Color: matrix.ColorWhite()},
	{Position: matrix.Vec3{0, 0.5, 0}, Normal: matrix.Vec3{0, 0, 1}, UV0: matrix.Vec2{0.5, 0}, Color: matrix.ColorWhite()},
	{Position: matrix.Vec3{0.5, -0.5, 0}, Normal: matrix.Vec3{0, 0, 1}, UV0: matrix.Vec2{1, 1}, Color: matrix.ColorWhite()},
}
mesh := host.MeshCache().Mesh("my-system/triangle-v1", verts, []uint32{0, 2, 1})
```

Winding and the material pipeline's cull/front-face settings must agree. Stock opaque pipelines cull back faces and use counter-clockwise front faces.

Do not call `AddMesh` after `MeshCache.Mesh`; `Mesh` already inserts and queues the mesh. Use `rendering.NewMesh(...)` plus `AddMesh(...)` only when constructing the object separately.

## Dynamic geometry

Use a dynamic mesh for frequent CPU-side vertex changes with a constant vertex count:

```go
mesh := host.MeshCache().DynamicMesh("water/debug-grid-v1", verts, indexes)

// Later; len(updated) must match the original vertex count.
host.MeshCache().UpdateMeshVertices(mesh.Key(), updated)
```

`UpdateMeshVertices` updates bounds and queues a GPU update. Changing topology/index data is not supported by this update path; create a different mesh/key when topology changes.

## One-texture unlit drawing

The framework helper builds the material instance and `ShaderDataUnlit`; the caller still supplies a transform and registers the drawing:

```go
texture, err := host.TextureCache().Texture("icons/marker.png", rendering.TextureFilterLinear)
if err != nil {
	return err
}
drawing, err := framework.CreateDrawingFromMeshUnlit(
	host, mesh, []*rendering.Texture{texture})
if err != nil {
	return err
}

entity := engine.NewEntity(host.WorkGroup())
drawing.Transform = &entity.Transform
entity.StoreShaderData(drawing.ShaderData)
entity.OnDestroy.Add(drawing.ShaderData.Destroy)
entity.OnDeactivate.Add(drawing.ShaderData.Deactivate)
entity.OnActivate.Add(drawing.ShaderData.Activate)
host.Drawings.AddDrawing(drawing)
```

Use `CreateDrawingFromMeshUnlitTransparent` for blended alpha.

## PBR drawing

PBR has a fixed four-texture contract:

```go
keys := []string{
	"crate_base_color.png",
	assets.TexturePBRDefaultNormal,
	assets.TexturePBRDefaultMetallicRough,
	assets.TextureBlankSquare,
}
textures := make([]*rendering.Texture, len(keys))
for i := range keys {
	var err error
	textures[i], err = host.TextureCache().Texture(keys[i], rendering.TextureFilterLinear)
	if err != nil {
		return err
	}
}

root, err := host.MaterialCache().Material(assets.MaterialDefinitionPBR)
if err != nil {
	return err
}
mat := root.CreateInstance(textures)
sd := shader_data_registry.Create(mat.Shader.DrawInstanceDataName())
pbr, ok := sd.(*shader_data_registry.ShaderDataPBR)
if !ok {
	return fmt.Errorf("material %q does not use PBR instance data", mat.Id)
}
pbr.VertColors = matrix.ColorWhite()
pbr.MeRoEmAo = matrix.NewVec4(1, 1, 0, 1)
pbr.LightIds = [...]int32{-1, -1, -1, -1}
```

The stock PBR root already sets `IsLit`, `CastsShadows`, and `ReceivesShadows`. Avoid mutating those flags on a shared cached root unless the change is intentionally global. `CreateInstance` shallow-copies the flags, so set per-binding behavior on the returned instance when needed.

## Imported OBJ, FBX, or glTF

Loaders return `load_result.Result`, which may contain multiple submeshes, texture maps, and node-local transforms:

```go
res, err := loaders.GLTF("models/robot.gltf", host.AssetDatabase())
if err != nil {
	return err
}
modelDrawings, err := framework.CreateDrawingsPBR(host, res)
if err != nil {
	return err
}
host.Drawings.AddDrawings(modelDrawings.AllDrawings())
```

Equivalent loaders are `loaders.OBJ` and `loaders.FBX`. Choose a framework creation function whose texture contract matches the content. `CreateDrawingsPBR` resolves slots by semantic name and supplies normal/metallic-roughness/emissive fallbacks. The unlit/basic helpers use their simpler contracts.

The returned drawings own standalone `matrix.Transform` values derived from loader nodes. If they must follow runtime entities or a hierarchy, deliberately re-parent/copy those transforms or rebuild drawings with entity transforms. Do not take the address of a loop variable that will be reused.

Retain `modelDrawings` for the model lifetime and destroy every drawing's shader data when unloading it. Framework creation does not invent an owning entity or automatic cleanup hook.

For serialized engine meshes, use `kaiju_mesh.ReadMesh(ref, host)`; a submesh reference has the form produced by `kaiju_mesh.MeshRefString(asset, key)`.

## Generated textures

Use `InsertImageTexture` for complete PNG/JPEG/etc. bytes:

```go
tex, err := host.TextureCache().InsertImageTexture(
	"generated/minimap-v7.png", encodedPNG, rendering.TextureFilterLinear)
```

Use `InsertRawTexture` for raw pixel memory with explicit dimensions:

```go
tex, err := host.TextureCache().InsertRawTexture(
	"generated/mask-v3", rgbaBytes, width, height, rendering.TextureFilterNearest)
```

Cache keys are identities. Reusing a key returns the existing texture. For a replacement workflow, use the cache reload/eviction API only after checking live drawing references and ownership. Normal drawing code should not call `DelayedCreate` directly; frame processing uploads pending textures. Explicit render-thread creation is reserved for code paths that demonstrably need the resource before normal pending processing.

Choose nearest filtering for hard-edged pixel data and linear filtering for interpolated imagery. A texture is cached separately per filter.

## Efficient repeated instances

Share the mesh, texture objects, and texture-bound material; create only transforms and shader data per object:

```go
drawings := make([]rendering.Drawing, count)
entities := make([]*engine.Entity, count)
for i := range count {
	entity := engine.NewEntity(host.WorkGroup())
	sd := shader_data_registry.Create(mat.Shader.DrawInstanceDataName())
	sd.(*shader_data_registry.ShaderDataStandard).Color = colors[i]
	entity.StoreShaderData(sd)
	entity.OnDestroy.Add(sd.Destroy)
	entity.OnDeactivate.Add(sd.Deactivate)
	entity.OnActivate.Add(sd.Activate)
	entities[i] = entity
	drawings[i] = rendering.Drawing{
		Material: mat, Mesh: mesh, ShaderData: sd,
		Transform: &entity.Transform,
		ViewCuller: &host.Cameras.Primary,
	}
}
host.Drawings.AddDrawings(drawings)
```

Compatible records become one `DrawInstanceGroup`. Creating duplicate `Texture` pointers outside `TextureCache`, using different material instances, changing layer, or changing mesh identity prevents some grouping.

## Direct model matrices

Omit `Drawing.Transform` only when another system owns the model matrix:

```go
sd := shader_data_registry.Create(mat.Shader.DrawInstanceDataName())
model := matrix.Mat4Identity()
model.Translate(position)
sd.SetModel(model)
```

With no transform, `SetModel` is authoritative and culling recomputes from that model each frame. Once a transform has been attached, `SetModel` changes `InitModel`, which is composed before the transform's world matrix.

## Layers and special passes

Set `Drawing.Layer` when the object is not ordinary world geometry:

```go
drawing.Layer = rendering.RenderLayerEditor
```

Zero means world. Picking uses the editor picking layers and a matching material/shader-data type. A layer alone does not create a pass; the material supplies the render pass and the render view must include the layer mask.

When one logical object has a visible draw and a picking/outline companion draw, give each its own shader data. Use `rendering.LinkDrawInstanceLifecycle(owner, follower)` when the companion must inherit destroy/activate/deactivate state.

## Troubleshooting map

- Panic in `AddDrawing`: mesh or material is nil; inspect the preceding cache error.
- Panic/fault or corrupted attributes during instance upload: shader-data type/`Size()` does not match the material shader's instance layout.
- White/default texture: wrong texture key, swallowed cache error, or material root used instead of the intended texture-bound instance.
- Textures shifted between roles: the slice order/count differs from the shader sampler contract.
- Correct alpha but wrong compositing/order: an opaque material/pipeline was used instead of a transparent variant.
- Object stuck at origin: `Drawing.Transform` is nil or points at the wrong transform; alternatively direct model data was never set.
- Object follows the wrong entity: one shader-data pointer was reused for multiple drawings and its stored transform was overwritten.
- Object survives entity deletion: no `OnDestroy -> DrawInstance.Destroy` wiring.
- Deactivated entity remains visible: no activation event wiring; entity state and draw-instance state are separate.
- Object disappears at camera angles: inspect mesh bounds, transform dirtiness, `ViewCuller`, and layer/view camera selection.
- New geometry does not appear: its cache key collided with an existing mesh.
- Per-object sort appears ignored: compatible instances share group-level sort; split their grouping identity or use a suitable material/pass design.
