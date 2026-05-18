# Terrain Texture Painting

Terrain texture painting is split between CPU-side authoring data and GPU-side
splat textures.

## Authoring Model

- `terrain.TerrainLayerSet` owns the ordered layer list and one
  `terrain.TextureWeightMap`.
- `terrain.TerrainLayer` stores the layer name, albedo content id, filter,
  tiling, tint, lock state, preview visibility, and reserved normal/roughness
  content ids.
- `terrain.TextureWeightMap` stores normalized blend weights in cell-major order:
  `((x + z*Resolution) * Layers) + layer`.
- Every weight-map texel should sum to `1` across all layers after edits. Use
  `PaintTextureLayer`, `FillLayer`, `ClearLayer`, `MoveLayer`,
  `ApplyTextureWeightRegion`, or `NormalizeWeightsAt` rather than editing the
  `Weights` slice directly.

Height and texture painting use different resolutions. `TerrainConfig.Resolution`
controls height vertices, while `TerrainConfig.PaintResolution` controls the
weight map. If `PaintResolution` is not set, it defaults to the height
resolution.

## Editor Behavior

The terrain workspace has two active paint families:

- Height sculpting: raise, lower, smooth.
- Texture painting: paint, erase, blend, fill, pick.

Height sculpt modifiers are temporary:

- Shift forces smooth.
- Ctrl or Cmd inverts raise/lower.
- Shift wins if both smooth and invert are held.

Texture strokes capture weight-map regions for undo/redo. Sample strokes do not
capture history because they only change the selected layer. Fill and sample
strokes are single-shot operations; paint, erase, and blend can be line-stamped
between pointer positions.

## Splat Textures

Runtime materials read terrain weights from RGBA splat textures. Four layers fit
in one splat texture:

- layer `0` -> texture `0`, channel `R`
- layer `1` -> texture `0`, channel `G`
- layer `2` -> texture `0`, channel `B`
- layer `3` -> texture `0`, channel `A`
- layer `4` starts the next splat texture

Painting texture weights must dirty splat texture regions without dirtying
heightfield mesh vertices. `Terrain.MarkTextureRegionDirty` and
`Terrain.ApplyTextureDirty` handle partial uploads when a GPU texture exists,
and retain CPU pixels for tests/model-only terrain.

## Shader And Material Regression Notes

The stock terrain material currently binds:

- sampler 0: `Weight Map 0`
- samplers 1-4: `Layer Albedo 0` through `Layer Albedo 3`

`terrain.frag` normalizes sampled RGBA weights before blending albedo layers and
falls back to full layer 0 weight if all sampled channels are empty. When
changing the material or shader:

- Keep sampler labels aligned with `terrainWeightMapSlots` and
  `terrainAlbedoLayerSlots` in `src/engine/terrain/terrain.go`.
- Keep the generated `terrain.shader` sampler count aligned with
  `terrain.frag`'s `TERRAIN_WEIGHT_MAP_COUNT`, `TERRAIN_LAYER_COUNT`, and
  `SAMPLER_COUNT` constants.
- Verify both lit and unlit terrain shaders if shared shader data or varyings
  change.
- Run `go test ./engine/terrain ./editor/editor_workspace/terrain_workspace
  ./registry/shader_data_registry` from `src`.
- In the editor, create or open a terrain with at least four layers, paint all
  channels, toggle layer lock/visibility/solo, and confirm texture painting does
  not rebuild height geometry.

The CPU tests cover weight normalization, locking, line-stamped strokes, splat
packing, dirty-region tracking, undo/redo weight-region restores, and editor
mode/modifier behavior. If a future material supports more than four visible
albedo layers at once, extend the shader/material bindings and add a regression
test that exercises the additional splat texture.
