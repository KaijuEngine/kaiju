# Rendering (`src/rendering/`)

Vulkan-based, with a comprehensive caching layer.

## Architecture

```
GPUApplication
    └── GPUInstance
        └── GPUDevice
            ├── GPULogicalDevice
            │   └── SwapChain
            └── GPUPainter
                └── CommandBuffers
```

GPU abstraction files: `gpu_application.go` (Vulkan app instance), `gpu_device.go`
(logical device), `gpu_swap_chain.go` (framebuffer swapping), `gpu_physical_device.go`
(hardware detection), with platform-specific `gpu_*_vulkan.go` files.

## Caches

- **ShaderCache** — compiles and caches GLSL shaders
- **TextureCache** — loads and caches textures
- **MeshCache** — loads and caches meshes
- **MaterialCache** — manages materials
- **FontCache** — font rendering

## Drawing

- **Drawing** — one renderable: Material + Mesh + Transform + ShaderData
- **Drawings** — the queue of all drawings (`host.Drawings`)
- **RenderPass** — a single pass through the pipeline

### Creating a drawing

```go
// 1. Mesh from cache
mesh := host.MeshCache().Mesh("path/to/mesh.obj") // or a UUID key

// 2. Material
mat, _ := host.MaterialCache().Material(assets.MaterialDefinitionBasic)

// 3. Material instance (with textures)
matInstance := mat.CreateInstance([]*rendering.Texture{texture})

// 4. Shader data matching the shader
sd := shader_data_registry.Create("basic")

// 5. Build and enqueue
draw := rendering.Drawing{
    Material:   matInstance,
    Mesh:       mesh,
    ShaderData: sd,
    Transform:  &entity.Transform, // links the drawing to the entity
    ViewCuller: &host.Cameras.Primary,
}
host.Drawings.AddDrawing(draw)
```

### Pre-built meshes (`src/rendering/mesh.go`)

```go
rendering.NewMeshSphere(cache, radius, widthSegments, heightSegments)
rendering.NewMeshCube(cache, size)
rendering.NewMeshPlane(cache, width, depth)
```

## Terrain texture painting

(`src/engine/terrain`, `src/editor/editor_workspace/terrain_workspace`)

Texture painting uses a separate layer/weight-map path from height sculpting.

- `TerrainConfig.Resolution` controls height vertices; `TerrainConfig.PaintResolution`
  controls texture weights and defaults to the height resolution when unset.
- `TerrainLayerSet` owns the ordered `TerrainLayer` list plus one
  `TextureWeightMap`. Layer indexes and weight-map channels must stay aligned.
- `TextureWeightMap.Weights` is cell-major:
  `((x + z*Resolution) * Layers) + layer`. Each texel should normalize to `1`
  across all layers after edits.
- Prefer `TerrainLayerSet`/`Terrain` helpers (`AddLayer`, `MoveLayer`,
  `PaintTextureLayer`, `FillLayer`, `ClearLayer`, `ApplyTextureWeightRegion`) over
  direct slice edits so locks, undo, dirty regions, and splat uploads stay correct.
- Locked layers preserve their weights during painting. Hidden and Solo affect
  preview/effective splat packing, not the stored weights.
- Four terrain layers pack into one RGBA splat texture: layer 0→R, 1→G, 2→B, 3→A;
  layer 4 starts the next splat texture.
- Texture painting dirties splat texture regions only; it must not dirty/rebuild
  heightfield mesh vertices. Height sculpting owns mesh vertex dirty regions.
- Editor height brush modifiers are temporary: Shift forces smooth, Ctrl/Cmd
  inverts raise/lower, Shift wins when both are held.
- Terrain UI markup maps `onclick` to a single Go function name; no JS-style
  chained calls.

### Shader / material contract

- The stock terrain material binds sampler 0 as `Weight Map 0` and samplers 1–4 as
  `Layer Albedo 0`–`Layer Albedo 3`.
- Keep `terrainWeightMapSlots`, `terrainAlbedoLayerSlots`, terrain material texture
  labels, generated `terrain.shader` sampler count, and `terrain.frag`'s terrain
  layer constants in sync.
- See `docs/engine/terrain_texture_painting.md` for regression notes and the editor
  smoke-test checklist.
