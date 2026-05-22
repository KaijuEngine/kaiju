/******************************************************************************/
/* doc.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

// Package terrain contains editable heightfield terrain, texture-layer painting,
// and the CPU-side weight maps that back terrain splat textures.
//
// Terrain height data and texture weights are intentionally separate. Height
// strokes dirty mesh vertices through HeightField, while texture strokes dirty
// TextureWeightMap regions that are repacked into TerrainSplatTexture RGBA
// channels. This lets the editor paint material weights without rebuilding
// terrain geometry.
//
// Layer weights are normalized per texel: the weights for all layers at a weight
// map coordinate should sum to one after mutation. Use TerrainLayerSet or
// TextureWeightMap helpers instead of editing WeightMap.Weights directly so
// locked layers, dirty regions, undo captures, and splat texture uploads stay
// coherent.
package terrain
