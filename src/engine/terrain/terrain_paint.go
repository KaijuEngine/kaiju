/******************************************************************************/
/* terrain_paint.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package terrain

import (
	"errors"
	"strings"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type TextureBrushMode int

const (
	TextureBrushPaint TextureBrushMode = iota
	TextureBrushErase
	TextureBrushSmoothWeights
	TextureBrushFill
	TextureBrushReplace
	TextureBrushSample
)

type TexturePaintConstraints struct {
	UseSlope     bool
	SlopeMin     matrix.Float
	SlopeMax     matrix.Float
	UseHeight    bool
	HeightMin    matrix.Float
	HeightMax    matrix.Float
	AngleCutoff  matrix.Float
	NormalFacing matrix.Vec3
}

type TextureBrushStamp struct {
	Resolution int
	Alpha      []matrix.Float
}

type TexturePaintStroke struct {
	Mode                TextureBrushMode
	Center              matrix.Vec2
	Radius              matrix.Float
	Strength            matrix.Float
	Falloff             BrushFalloff
	Opacity             matrix.Float
	TargetWeight        matrix.Float
	Spacing             matrix.Float
	ReplaceLayer        int
	PreserveOtherLayers bool
	Constraints         TexturePaintConstraints
	NoiseStrength       matrix.Float
	NoiseScale          matrix.Float
	NoiseSeed           int
	Jitter              matrix.Float
	Stamp               *TextureBrushStamp
	StampScale          matrix.Float
	StampRotation       matrix.Float
}

type TexturePaintResult struct {
	Dirty         DirtyRegion
	Sampled       bool
	SampledLayer  int
	SampledWeight matrix.Float
}

type TerrainLayer struct {
	Name               string
	TextureContentID   string
	NormalContentID    string
	RoughnessContentID string
	Filter             rendering.TextureFilter
	Tiling             matrix.Vec2
	Offset             matrix.Vec2
	Rotation           matrix.Float
	Tint               matrix.Color
	TextureWorldSize   matrix.Vec2
	Locked             bool
	Hidden             bool
	Solo               bool
	TriplanarCliffs    bool
	TriplanarSlope     matrix.Float
}

type TerrainLayerSet struct {
	Layers        []TerrainLayer
	WeightMap     *TextureWeightMap
	lockedScratch []bool
}

type TerrainLayerTextureDiagnostic struct {
	Layer            int
	Name             string
	TextureContentID string
}

type AutoMaterialRule struct {
	Name          string
	Layer         int
	TargetWeight  matrix.Float
	Constraints   TexturePaintConstraints
	NoiseStrength matrix.Float
	NoiseScale    matrix.Float
	NoiseSeed     int
}

type TerrainAutoMaterialPreset struct {
	GrassLayer    int
	RockLayer     int
	SnowLayer     int
	FlatSlopeMax  matrix.Float
	CliffSlopeMin matrix.Float
	SnowHeightMin matrix.Float
	NoiseStrength matrix.Float
	NoiseScale    matrix.Float
	NoiseSeed     int
}

type TextureWeightMap struct {
	Resolution int
	Layers     int
	Weights    []matrix.Float

	cellScratch   []matrix.Float
	smoothScratch []matrix.Float
}

func NewTerrainLayer(textureContentID string) TerrainLayer {
	return TerrainLayer{
		Name:             textureContentID,
		TextureContentID: textureContentID,
		Filter:           rendering.TextureFilterLinear,
		Tiling:           matrix.Vec2One(),
		Tint:             matrix.ColorWhite(),
	}
}

func MissingTerrainLayerTextures(layers []TerrainLayer, textureExists func(string) bool) []TerrainLayerTextureDiagnostic {
	if textureExists == nil {
		return nil
	}
	missing := make([]TerrainLayerTextureDiagnostic, 0)
	for i := range layers {
		id := strings.TrimSpace(layers[i].TextureContentID)
		if id == "" || textureExists(id) {
			continue
		}
		name := strings.TrimSpace(layers[i].Name)
		if name == "" {
			name = id
		}
		missing = append(missing, TerrainLayerTextureDiagnostic{
			Layer:            i,
			Name:             name,
			TextureContentID: id,
		})
	}
	return missing
}

func NewTerrainLayerSet(resolution int) (*TerrainLayerSet, error) {
	weights, err := NewTextureWeightMap(resolution, 0)
	if err != nil {
		return nil, err
	}
	return &TerrainLayerSet{
		Layers:    make([]TerrainLayer, 0),
		WeightMap: weights,
	}, nil
}

func NewTextureWeightMap(resolution, layers int) (*TextureWeightMap, error) {
	if resolution < 2 {
		return nil, errors.New("terrain texture weight map resolution must be at least 2")
	}
	if layers < 0 {
		return nil, errors.New("terrain texture weight map layer count cannot be negative")
	}
	m := &TextureWeightMap{
		Resolution: resolution,
		Layers:     layers,
		Weights:    make([]matrix.Float, resolution*resolution*layers),
	}
	if layers > 0 {
		m.FillLayer(0)
	}
	return m, nil
}

func (s *TerrainLayerSet) LayerCount() int {
	if s == nil {
		return 0
	}
	return len(s.Layers)
}

func (s *TerrainLayerSet) AddLayer(layer TerrainLayer) int {
	if s == nil || s.WeightMap == nil {
		return -1
	}
	layer = normalizeTerrainLayer(layer)
	s.Layers = append(s.Layers, layer)
	defaultWeight := matrix.Float(0)
	if len(s.Layers) == 1 {
		defaultWeight = 1
	}
	s.WeightMap.AddLayer(defaultWeight)
	return len(s.Layers) - 1
}

func (s *TerrainLayerSet) SetLayer(layer int, value TerrainLayer) bool {
	if s == nil || layer < 0 || layer >= len(s.Layers) {
		return false
	}
	s.Layers[layer] = normalizeTerrainLayer(value)
	return true
}

func (s *TerrainLayerSet) RemoveLayer(layer int) bool {
	if s == nil || s.WeightMap == nil || layer < 0 || layer >= len(s.Layers) {
		return false
	}
	copy(s.Layers[layer:], s.Layers[layer+1:])
	s.Layers = s.Layers[:len(s.Layers)-1]
	if !s.WeightMap.RemoveLayer(layer) {
		return false
	}
	s.WeightMap.NormalizeAll()
	return true
}

func (s *TerrainLayerSet) MoveLayer(from, to int) bool {
	if s == nil || s.WeightMap == nil ||
		from < 0 || to < 0 || from >= len(s.Layers) || to >= len(s.Layers) || from == to {
		return false
	}
	layer := s.Layers[from]
	if from < to {
		copy(s.Layers[from:to], s.Layers[from+1:to+1])
	} else {
		copy(s.Layers[to+1:from+1], s.Layers[to:from])
	}
	s.Layers[to] = layer
	return s.WeightMap.MoveLayer(from, to)
}

func (s *TerrainLayerSet) NormalizeWeightsAt(x, z int) bool {
	if s == nil || s.WeightMap == nil {
		return false
	}
	return s.WeightMap.NormalizeWeightsAt(x, z)
}

func (s *TerrainLayerSet) PaintLayer(layer int, stroke PaintStroke) DirtyRegion {
	if s == nil || s.WeightMap == nil {
		return DirtyRegion{}
	}
	return s.PaintTextureLayer(layer, TexturePaintStroke{
		Mode:     TextureBrushPaint,
		Center:   stroke.Center,
		Radius:   stroke.Radius,
		Strength: stroke.Strength,
		Falloff:  stroke.Falloff,
		Spacing:  stroke.Spacing,
	}).Dirty
}

func (s *TerrainLayerSet) EraseLayer(layer int, stroke PaintStroke) DirtyRegion {
	if s == nil || s.WeightMap == nil {
		return DirtyRegion{}
	}
	return s.PaintTextureLayer(layer, TexturePaintStroke{
		Mode:     TextureBrushErase,
		Center:   stroke.Center,
		Radius:   stroke.Radius,
		Strength: stroke.Strength,
		Falloff:  stroke.Falloff,
		Spacing:  stroke.Spacing,
	}).Dirty
}

func (s *TerrainLayerSet) FillLayer(layer int) DirtyRegion {
	if s == nil || s.WeightMap == nil {
		return DirtyRegion{}
	}
	return s.WeightMap.FillLayerWithLocks(layer, s.lockedLayers())
}

func (s *TerrainLayerSet) ClearLayer(layer int) DirtyRegion {
	if s == nil || s.WeightMap == nil {
		return DirtyRegion{}
	}
	return s.WeightMap.ClearLayerWithLocks(layer, s.lockedLayers())
}

func (s *TerrainLayerSet) LayerWeightAt(layer, x, z int) matrix.Float {
	if s == nil || s.WeightMap == nil {
		return 0
	}
	return s.WeightMap.WeightAt(layer, x, z)
}

func (s *TerrainLayerSet) SetLayerWeightAt(layer, x, z int, weight matrix.Float) bool {
	if s == nil || s.WeightMap == nil {
		return false
	}
	return s.WeightMap.SetWeightAt(layer, x, z, weight)
}

func (s *TerrainLayerSet) PaintTextureLayer(layer int, stroke TexturePaintStroke) TexturePaintResult {
	if s == nil || s.WeightMap == nil {
		return TexturePaintResult{}
	}
	return s.WeightMap.PaintTextureLayerWithLocks(layer, stroke, s.lockedLayers())
}

func (s *TerrainLayerSet) EffectiveWeightMapForPreview() *TextureWeightMap {
	if s == nil || s.WeightMap == nil {
		return nil
	}
	if !s.hasPreviewFilter() {
		return s.WeightMap
	}
	out, err := NewTextureWeightMap(s.WeightMap.Resolution, s.WeightMap.Layers)
	if err != nil {
		return s.WeightMap
	}
	copy(out.Weights, s.WeightMap.Weights)
	for z := 0; z < out.Resolution; z++ {
		for x := 0; x < out.Resolution; x++ {
			for layer := 0; layer < out.Layers; layer++ {
				if !s.layerPreviewVisible(layer) {
					out.Weights[out.index(layer, x, z)] = 0
				}
			}
			out.NormalizeWeightsAt(x, z)
		}
	}
	return out
}

func (s *TerrainLayerSet) WeightDebugRGBA(layer int) []byte {
	if s == nil || s.WeightMap == nil || layer < 0 || layer >= s.WeightMap.Layers {
		return nil
	}
	pixels := make([]byte, s.WeightMap.Resolution*s.WeightMap.Resolution*4)
	for z := 0; z < s.WeightMap.Resolution; z++ {
		for x := 0; x < s.WeightMap.Resolution; x++ {
			weight := weightToByte(s.WeightMap.WeightAt(layer, x, z))
			i := (x + z*s.WeightMap.Resolution) * 4
			pixels[i+0] = weight
			pixels[i+1] = weight
			pixels[i+2] = weight
			pixels[i+3] = 255
		}
	}
	return pixels
}

func (s *TerrainLayerSet) lockedLayers() []bool {
	if s == nil || len(s.Layers) == 0 {
		return nil
	}
	if cap(s.lockedScratch) < len(s.Layers) {
		s.lockedScratch = make([]bool, len(s.Layers))
	}
	locked := s.lockedScratch[:len(s.Layers)]
	for i := range s.Layers {
		locked[i] = s.Layers[i].Locked
	}
	return locked
}

func (s *TerrainLayerSet) hasPreviewFilter() bool {
	if s == nil {
		return false
	}
	for i := range s.Layers {
		if s.Layers[i].Hidden || s.Layers[i].Solo {
			return true
		}
	}
	return false
}

func (s *TerrainLayerSet) layerPreviewVisible(layer int) bool {
	if s == nil || layer < 0 || layer >= len(s.Layers) {
		return false
	}
	hasSolo := false
	for i := range s.Layers {
		hasSolo = hasSolo || s.Layers[i].Solo
	}
	if hasSolo {
		return s.Layers[layer].Solo
	}
	return !s.Layers[layer].Hidden
}

func (m *TextureWeightMap) AddLayer(defaultWeight matrix.Float) {
	if m == nil {
		return
	}
	defaultWeight = matrix.Clamp(defaultWeight, 0, 1)
	oldLayers := m.Layers
	newLayers := oldLayers + 1
	next := make([]matrix.Float, m.Resolution*m.Resolution*newLayers)
	for z := 0; z < m.Resolution; z++ {
		for x := 0; x < m.Resolution; x++ {
			cell := x + z*m.Resolution
			for layer := 0; layer < oldLayers; layer++ {
				next[cell*newLayers+layer] = m.Weights[cell*oldLayers+layer]
			}
			next[cell*newLayers+oldLayers] = defaultWeight
		}
	}
	m.Layers = newLayers
	m.Weights = next
	if defaultWeight != 0 || newLayers == 1 {
		m.NormalizeAll()
	}
}

func (m *TextureWeightMap) RemoveLayer(removeLayer int) bool {
	if m == nil || removeLayer < 0 || removeLayer >= m.Layers {
		return false
	}
	oldLayers := m.Layers
	newLayers := oldLayers - 1
	next := make([]matrix.Float, m.Resolution*m.Resolution*newLayers)
	for z := 0; z < m.Resolution; z++ {
		for x := 0; x < m.Resolution; x++ {
			cell := x + z*m.Resolution
			dstLayer := 0
			for layer := 0; layer < oldLayers; layer++ {
				if layer == removeLayer {
					continue
				}
				next[cell*newLayers+dstLayer] = m.Weights[cell*oldLayers+layer]
				dstLayer++
			}
		}
	}
	m.Layers = newLayers
	m.Weights = next
	return true
}

func (m *TextureWeightMap) MoveLayer(from, to int) bool {
	if m == nil || from < 0 || to < 0 || from >= m.Layers || to >= m.Layers || from == to {
		return false
	}
	oldWeights := m.Weights
	next := make([]matrix.Float, len(oldWeights))
	order := make([]int, m.Layers)
	for i := range order {
		order[i] = i
	}
	moved := order[from]
	if from < to {
		copy(order[from:to], order[from+1:to+1])
	} else {
		copy(order[to+1:from+1], order[to:from])
	}
	order[to] = moved
	for z := 0; z < m.Resolution; z++ {
		for x := 0; x < m.Resolution; x++ {
			cell := x + z*m.Resolution
			for dstLayer, srcLayer := range order {
				next[cell*m.Layers+dstLayer] = oldWeights[cell*m.Layers+srcLayer]
			}
		}
	}
	m.Weights = next
	return true
}

func (m *TextureWeightMap) WeightAt(layer, x, z int) matrix.Float {
	if m == nil || !m.inBounds(layer, x, z) {
		return 0
	}
	return m.Weights[m.index(layer, x, z)]
}

func (m *TextureWeightMap) SetWeightAt(layer, x, z int, weight matrix.Float) bool {
	if m == nil || !m.inBounds(layer, x, z) {
		return false
	}
	m.Weights[m.index(layer, x, z)] = matrix.Clamp(weight, 0, 1)
	return true
}

func (m *TextureWeightMap) CopyRegion(region DirtyRegion) []matrix.Float {
	if m == nil || !region.Valid {
		return nil
	}
	region = region.Expand(0, m.Resolution)
	width := region.MaxX - region.MinX + 1
	height := region.MaxZ - region.MinZ + 1
	out := make([]matrix.Float, width*height*m.Layers)
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			cell := (x - region.MinX) + (z-region.MinZ)*width
			for layer := 0; layer < m.Layers; layer++ {
				out[cell*m.Layers+layer] = m.WeightAt(layer, x, z)
			}
		}
	}
	return out
}

func (m *TextureWeightMap) SetRegion(region DirtyRegion, weights []matrix.Float) DirtyRegion {
	if m == nil || !region.Valid {
		return DirtyRegion{}
	}
	region = region.Expand(0, m.Resolution)
	width := region.MaxX - region.MinX + 1
	height := region.MaxZ - region.MinZ + 1
	if len(weights) != width*height*m.Layers {
		return DirtyRegion{}
	}
	var dirty DirtyRegion
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			cell := (x - region.MinX) + (z-region.MinZ)*width
			changed := false
			for layer := 0; layer < m.Layers; layer++ {
				next := matrix.Clamp(weights[cell*m.Layers+layer], 0, 1)
				idx := m.index(layer, x, z)
				if m.Weights[idx] != next {
					m.Weights[idx] = next
					changed = true
				}
			}
			if changed {
				dirty = mergeDirtyRegions(dirty, DirtyRegion{
					MinX: x, MinZ: z, MaxX: x, MaxZ: z, Valid: true,
				})
			}
		}
	}
	return dirty
}

func (m *TextureWeightMap) Sample(layer int, x, z matrix.Float) matrix.Float {
	if m == nil || layer < 0 || layer >= m.Layers {
		return 0
	}
	if x < 0 {
		x = 0
	} else if x > matrix.Float(m.Resolution-1) {
		x = matrix.Float(m.Resolution - 1)
	}
	if z < 0 {
		z = 0
	} else if z > matrix.Float(m.Resolution-1) {
		z = matrix.Float(m.Resolution - 1)
	}
	x0 := int(matrix.Floor(x))
	z0 := int(matrix.Floor(z))
	x1 := min(x0+1, m.Resolution-1)
	z1 := min(z0+1, m.Resolution-1)
	tx := x - matrix.Float(x0)
	tz := z - matrix.Float(z0)
	w00 := m.WeightAt(layer, x0, z0)
	w10 := m.WeightAt(layer, x1, z0)
	w01 := m.WeightAt(layer, x0, z1)
	w11 := m.WeightAt(layer, x1, z1)
	return matrix.Lerp(matrix.Lerp(w00, w10, tx), matrix.Lerp(w01, w11, tx), tz)
}

func (m *TextureWeightMap) NormalizeWeightsAt(x, z int) bool {
	if m == nil || x < 0 || z < 0 || x >= m.Resolution || z >= m.Resolution || m.Layers == 0 {
		return false
	}
	sum := matrix.Float(0)
	for layer := 0; layer < m.Layers; layer++ {
		idx := m.index(layer, x, z)
		m.Weights[idx] = matrix.Clamp(m.Weights[idx], 0, 1)
		sum += m.Weights[idx]
	}
	if sum <= matrix.Tiny {
		for layer := 0; layer < m.Layers; layer++ {
			m.Weights[m.index(layer, x, z)] = 0
		}
		m.Weights[m.index(0, x, z)] = 1
		return true
	}
	for layer := 0; layer < m.Layers; layer++ {
		idx := m.index(layer, x, z)
		m.Weights[idx] /= sum
	}
	return true
}

func (m *TextureWeightMap) NormalizeAll() {
	if m == nil {
		return
	}
	for z := 0; z < m.Resolution; z++ {
		for x := 0; x < m.Resolution; x++ {
			m.NormalizeWeightsAt(x, z)
		}
	}
}

func (m *TextureWeightMap) PaintLayer(layer int, stroke PaintStroke) DirtyRegion {
	return m.paintLayer(layer, stroke, false)
}

func (m *TextureWeightMap) EraseLayer(layer int, stroke PaintStroke) DirtyRegion {
	return m.paintLayer(layer, stroke, true)
}

func (m *TextureWeightMap) FillLayer(fillLayer int) DirtyRegion {
	return m.FillLayerWithLocks(fillLayer, nil)
}

func (m *TextureWeightMap) FillLayerWithLocks(fillLayer int, locked []bool) DirtyRegion {
	if m == nil || fillLayer < 0 || fillLayer >= m.Layers {
		return DirtyRegion{}
	}
	if layerLocked(locked, fillLayer) {
		return DirtyRegion{}
	}
	var dirty DirtyRegion
	for z := 0; z < m.Resolution; z++ {
		for x := 0; x < m.Resolution; x++ {
			m.cellScratch = m.weightsAtInto(x, z, m.cellScratch)
			before := m.cellScratch
			for layer := 0; layer < m.Layers; layer++ {
				weight := matrix.Float(0)
				if layer == fillLayer {
					weight = 1
				}
				m.Weights[m.index(layer, x, z)] = weight
			}
			m.applyLockedWeightsAt(x, z, before, locked)
			if m.weightsChangedAt(x, z, before) {
				dirty = mergeDirtyRegions(dirty, DirtyRegion{
					MinX: x, MinZ: z, MaxX: x, MaxZ: z, Valid: true,
				})
			}
		}
	}
	return dirty
}

func (m *TextureWeightMap) ClearLayer(clearLayer int) DirtyRegion {
	return m.ClearLayerWithLocks(clearLayer, nil)
}

func (m *TextureWeightMap) ClearLayerWithLocks(clearLayer int, locked []bool) DirtyRegion {
	if m == nil || clearLayer < 0 || clearLayer >= m.Layers {
		return DirtyRegion{}
	}
	if layerLocked(locked, clearLayer) {
		return DirtyRegion{}
	}
	var dirty DirtyRegion
	for z := 0; z < m.Resolution; z++ {
		for x := 0; x < m.Resolution; x++ {
			m.cellScratch = m.weightsAtInto(x, z, m.cellScratch)
			before := m.cellScratch
			lockedSum := matrix.Float(0)
			unlockedOtherSum := matrix.Float(0)
			fallback := -1
			for layer := 0; layer < m.Layers; layer++ {
				if layerLocked(locked, layer) {
					lockedSum += matrix.Clamp(before[layer], 0, 1)
				} else if layer != clearLayer {
					unlockedOtherSum += matrix.Clamp(before[layer], 0, 1)
					if fallback < 0 {
						fallback = layer
					}
				}
			}
			remaining := matrix.Clamp(1-lockedSum, 0, 1)
			for layer := 0; layer < m.Layers; layer++ {
				switch {
				case layerLocked(locked, layer):
					m.Weights[m.index(layer, x, z)] = before[layer]
				case layer == clearLayer:
					m.Weights[m.index(layer, x, z)] = 0
				case unlockedOtherSum > matrix.Tiny:
					m.Weights[m.index(layer, x, z)] = before[layer] * remaining / unlockedOtherSum
				default:
					m.Weights[m.index(layer, x, z)] = 0
				}
			}
			if fallback >= 0 && unlockedOtherSum <= matrix.Tiny {
				m.Weights[m.index(fallback, x, z)] = remaining
			}
			m.applyLockedWeightsAt(x, z, before, locked)
			if m.weightsChangedAt(x, z, before) {
				dirty = mergeDirtyRegions(dirty, DirtyRegion{
					MinX: x, MinZ: z, MaxX: x, MaxZ: z, Valid: true,
				})
			}
		}
	}
	return dirty
}

func (m *TextureWeightMap) PaintTextureLayer(layer int, stroke TexturePaintStroke) TexturePaintResult {
	return m.PaintTextureLayerWithLocks(layer, stroke, nil)
}

func (m *TextureWeightMap) PaintTextureLayerWithLocks(layer int, stroke TexturePaintStroke, locked []bool) TexturePaintResult {
	return m.paintTextureLayer(layer, stroke, nil, locked)
}

type texturePaintFilter func(x, z int) bool

func (m *TextureWeightMap) paintTextureLayer(layer int, stroke TexturePaintStroke, filter texturePaintFilter, locked []bool) TexturePaintResult {
	if m == nil || m.Layers == 0 {
		return TexturePaintResult{}
	}
	stroke = normalizeTexturePaintStroke(stroke)
	if stroke.Mode == TextureBrushSample {
		return m.sampleTextureLayer(stroke)
	}
	if layer < 0 || layer >= m.Layers {
		return TexturePaintResult{}
	}
	if layerLocked(locked, layer) {
		return TexturePaintResult{}
	}
	region := m.textureStrokeRegion(stroke)
	if !region.Valid {
		return TexturePaintResult{}
	}
	var original []matrix.Float
	if stroke.Mode == TextureBrushSmoothWeights {
		if cap(m.smoothScratch) < len(m.Weights) {
			m.smoothScratch = make([]matrix.Float, len(m.Weights))
		}
		original = m.smoothScratch[:len(m.Weights)]
		copy(original, m.Weights)
	}
	var dirty DirtyRegion
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			if filter != nil && !filter(x, z) {
				continue
			}
			amount, ok := m.textureStrokeAmount(stroke, x, z)
			if !ok {
				continue
			}
			m.NormalizeWeightsAt(x, z)
			m.cellScratch = m.weightsAtInto(x, z, m.cellScratch)
			before := m.cellScratch
			changed := false
			switch stroke.Mode {
			case TextureBrushErase:
				changed = m.setTargetWeightAt(layer, x, z, 0, amount, stroke, before)
			case TextureBrushSmoothWeights:
				changed = m.smoothWeightsAt(x, z, amount, original, before)
			case TextureBrushFill:
				changed = m.setTargetWeightAt(layer, x, z, stroke.TargetWeight, amount, stroke, before)
			case TextureBrushReplace:
				changed = m.replaceWeightAt(layer, stroke.ReplaceLayer, x, z, amount, before)
			case TextureBrushPaint:
				fallthrough
			default:
				changed = m.setTargetWeightAt(layer, x, z, stroke.TargetWeight, amount, stroke, before)
			}
			if !changed && !m.weightsChangedAt(x, z, before) {
				continue
			}
			m.applyLockedWeightsAt(x, z, before, locked)
			if !m.weightsChangedAt(x, z, before) {
				continue
			}
			dirty = mergeDirtyRegions(dirty, DirtyRegion{
				MinX: x, MinZ: z, MaxX: x, MaxZ: z, Valid: true,
			})
		}
	}
	return TexturePaintResult{Dirty: dirty}
}

func normalizeTexturePaintStroke(stroke TexturePaintStroke) TexturePaintStroke {
	if stroke.Opacity <= 0 {
		stroke.Opacity = 1
	}
	stroke.Opacity = matrix.Clamp(stroke.Opacity, 0, 1)
	if stroke.Mode == TextureBrushPaint || stroke.Mode == TextureBrushFill {
		if stroke.TargetWeight == 0 {
			stroke.TargetWeight = 1
		}
	}
	stroke.TargetWeight = matrix.Clamp(stroke.TargetWeight, 0, 1)
	stroke.NoiseStrength = matrix.Clamp(stroke.NoiseStrength, 0, 1)
	if stroke.NoiseScale < 0 {
		stroke.NoiseScale = 0
	}
	if stroke.Jitter < 0 {
		stroke.Jitter = 0
	}
	if stroke.StampScale <= 0 {
		stroke.StampScale = 1
	}
	return stroke
}

func (m *TextureWeightMap) textureStrokeRegion(stroke TexturePaintStroke) DirtyRegion {
	if m == nil {
		return DirtyRegion{}
	}
	if stroke.Mode == TextureBrushFill || (stroke.Mode == TextureBrushReplace && stroke.Radius <= 0) {
		return DirtyRegion{MinX: 0, MinZ: 0, MaxX: m.Resolution - 1, MaxZ: m.Resolution - 1, Valid: true}
	}
	if stroke.Radius <= 0 {
		return DirtyRegion{}
	}
	minX := max(0, int(matrix.Floor(stroke.Center.X()-stroke.Radius)))
	maxX := min(m.Resolution-1, int(matrix.Ceil(stroke.Center.X()+stroke.Radius)))
	minZ := max(0, int(matrix.Floor(stroke.Center.Y()-stroke.Radius)))
	maxZ := min(m.Resolution-1, int(matrix.Ceil(stroke.Center.Y()+stroke.Radius)))
	if minX > maxX || minZ > maxZ {
		return DirtyRegion{}
	}
	return DirtyRegion{MinX: minX, MinZ: minZ, MaxX: maxX, MaxZ: maxZ, Valid: true}
}

func (m *TextureWeightMap) textureStrokeAmount(stroke TexturePaintStroke, x, z int) (matrix.Float, bool) {
	strength := matrix.Abs(stroke.Strength)
	if stroke.Mode == TextureBrushFill && strength == 0 {
		strength = 1
	}
	if stroke.Mode == TextureBrushReplace && stroke.Radius <= 0 && strength == 0 {
		strength = 1
	}
	if strength == 0 {
		return 0, false
	}
	weight := matrix.Float(1)
	if stroke.Mode != TextureBrushFill && stroke.Radius > 0 {
		dx := matrix.Float(x) - stroke.Center.X()
		dz := matrix.Float(z) - stroke.Center.Y()
		distance := matrix.Sqrt(dx*dx + dz*dz)
		if stroke.Jitter > 0 {
			jitter := textureHashSigned(x, z, stroke.NoiseSeed+17) * stroke.Jitter
			distance = matrix.Max(0, distance+jitter)
		}
		if distance > stroke.Radius {
			return 0, false
		}
		weight = brushWeight(distance, stroke.Radius, stroke.Falloff)
	}
	if stroke.Stamp != nil && stroke.Mode != TextureBrushFill {
		weight *= stroke.Stamp.Sample(
			(matrix.Float(x)-stroke.Center.X())/matrix.Max(stroke.Radius, matrix.Tiny),
			(matrix.Float(z)-stroke.Center.Y())/matrix.Max(stroke.Radius, matrix.Tiny),
			stroke.StampScale,
			stroke.StampRotation,
		)
	}
	if stroke.NoiseStrength > 0 {
		noise := textureHashNoise(x, z, stroke.NoiseScale, stroke.NoiseSeed)
		weight *= matrix.Lerp(1, noise, stroke.NoiseStrength)
	}
	amount := matrix.Clamp(strength*stroke.Opacity*weight, 0, 1)
	return amount, amount > 0
}

func (m *TextureWeightMap) setTargetWeightAt(layer, x, z int, target, amount matrix.Float, stroke TexturePaintStroke, before []matrix.Float) bool {
	if !m.inBounds(layer, x, z) {
		return false
	}
	afterTarget := matrix.Lerp(before[layer], matrix.Clamp(target, 0, 1), matrix.Clamp(amount, 0, 1))
	if stroke.PreserveOtherLayers {
		m.applyPreservedTargetWeight(layer, x, z, before, afterTarget, stroke.ReplaceLayer)
	} else {
		m.redistributeTargetWeight(layer, x, z, before, afterTarget)
	}
	m.NormalizeWeightsAt(x, z)
	return m.weightsChangedAt(x, z, before)
}

func (m *TextureWeightMap) redistributeTargetWeight(layer, x, z int, before []matrix.Float, afterTarget matrix.Float) {
	beforeOtherSum := matrix.Float(1) - before[layer]
	afterOtherSum := matrix.Float(1) - afterTarget
	for i := 0; i < m.Layers; i++ {
		if i == layer {
			m.Weights[m.index(i, x, z)] = afterTarget
		} else if beforeOtherSum > matrix.Tiny {
			m.Weights[m.index(i, x, z)] = before[i] * afterOtherSum / beforeOtherSum
		} else {
			m.Weights[m.index(i, x, z)] = 0
		}
	}
	if m.Layers > 1 && beforeOtherSum <= matrix.Tiny && afterOtherSum > matrix.Tiny {
		fallback := firstOtherLayer(layer, m.Layers)
		m.Weights[m.index(fallback, x, z)] = afterOtherSum
	}
}

func (m *TextureWeightMap) applyPreservedTargetWeight(layer, x, z int, before []matrix.Float, afterTarget matrix.Float, preserveLayer int) {
	for i := 0; i < m.Layers; i++ {
		m.Weights[m.index(i, x, z)] = before[i]
	}
	compensator := preserveLayer
	if compensator < 0 || compensator >= m.Layers || compensator == layer {
		compensator = firstOtherLayer(layer, m.Layers)
	}
	m.Weights[m.index(layer, x, z)] = afterTarget
	if compensator >= 0 {
		delta := afterTarget - before[layer]
		m.Weights[m.index(compensator, x, z)] = matrix.Clamp(before[compensator]-delta, 0, 1)
	}
}

func (m *TextureWeightMap) replaceWeightAt(layer, replaceLayer, x, z int, amount matrix.Float, before []matrix.Float) bool {
	if !m.inBounds(layer, x, z) || !m.inBounds(replaceLayer, x, z) || layer == replaceLayer {
		return false
	}
	move := before[replaceLayer] * matrix.Clamp(amount, 0, 1)
	m.Weights[m.index(layer, x, z)] = before[layer] + move
	m.Weights[m.index(replaceLayer, x, z)] = before[replaceLayer] - move
	m.NormalizeWeightsAt(x, z)
	return m.weightsChangedAt(x, z, before)
}

func (m *TextureWeightMap) smoothWeightsAt(x, z int, amount matrix.Float, original []matrix.Float, before []matrix.Float) bool {
	if len(original) != len(m.Weights) {
		return false
	}
	for layer := 0; layer < m.Layers; layer++ {
		average := m.neighborWeightAverage(original, layer, x, z)
		m.Weights[m.index(layer, x, z)] = matrix.Lerp(before[layer], average, matrix.Clamp(amount, 0, 1))
	}
	m.NormalizeWeightsAt(x, z)
	return m.weightsChangedAt(x, z, before)
}

func (m *TextureWeightMap) neighborWeightAverage(weights []matrix.Float, layer, x, z int) matrix.Float {
	var sum matrix.Float
	count := matrix.Float(0)
	for nz := max(0, z-1); nz <= min(m.Resolution-1, z+1); nz++ {
		for nx := max(0, x-1); nx <= min(m.Resolution-1, x+1); nx++ {
			sum += weights[(nx+nz*m.Resolution)*m.Layers+layer]
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

func (m *TextureWeightMap) sampleTextureLayer(stroke TexturePaintStroke) TexturePaintResult {
	x := int(matrix.Round(stroke.Center.X()))
	z := int(matrix.Round(stroke.Center.Y()))
	x = max(0, min(m.Resolution-1, x))
	z = max(0, min(m.Resolution-1, z))
	bestLayer := 0
	bestWeight := m.WeightAt(0, x, z)
	for layer := 1; layer < m.Layers; layer++ {
		weight := m.WeightAt(layer, x, z)
		if weight > bestWeight {
			bestLayer = layer
			bestWeight = weight
		}
	}
	return TexturePaintResult{
		Sampled:       true,
		SampledLayer:  bestLayer,
		SampledWeight: bestWeight,
	}
}

func (m *TextureWeightMap) copyWeightsAt(x, z int) []matrix.Float {
	out := make([]matrix.Float, m.Layers)
	m.weightsAtInto(x, z, out)
	return out
}

func (m *TextureWeightMap) weightsAtInto(x, z int, out []matrix.Float) []matrix.Float {
	if cap(out) < m.Layers {
		out = make([]matrix.Float, m.Layers)
	}
	out = out[:m.Layers]
	for layer := 0; layer < m.Layers; layer++ {
		out[layer] = m.WeightAt(layer, x, z)
	}
	return out
}

func (m *TextureWeightMap) weightsChangedAt(x, z int, before []matrix.Float) bool {
	if len(before) != m.Layers {
		return true
	}
	for layer := 0; layer < m.Layers; layer++ {
		if !matrix.ApproxTo(before[layer], m.WeightAt(layer, x, z), matrix.Roughly) {
			return true
		}
	}
	return false
}

func (m *TextureWeightMap) applyLockedWeightsAt(x, z int, before []matrix.Float, locked []bool) {
	if m == nil || len(locked) == 0 || len(before) != m.Layers {
		m.NormalizeWeightsAt(x, z)
		return
	}
	lockedSum := matrix.Float(0)
	unlockedSum := matrix.Float(0)
	unlockedCount := 0
	for layer := 0; layer < m.Layers; layer++ {
		if layerLocked(locked, layer) {
			weight := matrix.Clamp(before[layer], 0, 1)
			m.Weights[m.index(layer, x, z)] = weight
			lockedSum += weight
		} else {
			weight := matrix.Clamp(m.WeightAt(layer, x, z), 0, 1)
			m.Weights[m.index(layer, x, z)] = weight
			unlockedSum += weight
			unlockedCount++
		}
	}
	remaining := matrix.Clamp(1-lockedSum, 0, 1)
	if unlockedCount == 0 {
		return
	}
	if unlockedSum <= matrix.Tiny {
		even := remaining / matrix.Float(unlockedCount)
		for layer := 0; layer < m.Layers; layer++ {
			if !layerLocked(locked, layer) {
				m.Weights[m.index(layer, x, z)] = even
			}
		}
		return
	}
	scale := remaining / unlockedSum
	for layer := 0; layer < m.Layers; layer++ {
		if !layerLocked(locked, layer) {
			m.Weights[m.index(layer, x, z)] *= scale
		}
	}
}

func layerLocked(locked []bool, layer int) bool {
	return layer >= 0 && layer < len(locked) && locked[layer]
}

func firstOtherLayer(layer, layers int) int {
	for i := 0; i < layers; i++ {
		if i != layer {
			return i
		}
	}
	return -1
}

func weightsChanged(before, after []matrix.Float) bool {
	if len(before) != len(after) {
		return true
	}
	for i := range before {
		if !matrix.ApproxTo(before[i], after[i], matrix.Roughly) {
			return true
		}
	}
	return false
}

func (s *TextureBrushStamp) Sample(x, z, scale, rotation matrix.Float) matrix.Float {
	if s == nil || s.Resolution < 2 || len(s.Alpha) != s.Resolution*s.Resolution {
		return 1
	}
	if scale <= 0 {
		scale = 1
	}
	x /= scale
	z /= scale
	if rotation != 0 {
		sin, cos := matrix.Sin(rotation), matrix.Cos(rotation)
		x, z = x*cos-z*sin, x*sin+z*cos
	}
	u := x*0.5 + 0.5
	v := z*0.5 + 0.5
	if u < 0 || u > 1 || v < 0 || v > 1 {
		return 0
	}
	gx := u * matrix.Float(s.Resolution-1)
	gz := v * matrix.Float(s.Resolution-1)
	x0 := int(matrix.Floor(gx))
	z0 := int(matrix.Floor(gz))
	x1 := min(x0+1, s.Resolution-1)
	z1 := min(z0+1, s.Resolution-1)
	tx := gx - matrix.Float(x0)
	tz := gz - matrix.Float(z0)
	a00 := matrix.Clamp(s.Alpha[x0+z0*s.Resolution], 0, 1)
	a10 := matrix.Clamp(s.Alpha[x1+z0*s.Resolution], 0, 1)
	a01 := matrix.Clamp(s.Alpha[x0+z1*s.Resolution], 0, 1)
	a11 := matrix.Clamp(s.Alpha[x1+z1*s.Resolution], 0, 1)
	return matrix.Lerp(matrix.Lerp(a00, a10, tx), matrix.Lerp(a01, a11, tx), tz)
}

func textureHashNoise(x, z int, scale matrix.Float, seed int) matrix.Float {
	if scale > matrix.Tiny {
		x = int(matrix.Floor(matrix.Float(x) / scale))
		z = int(matrix.Floor(matrix.Float(z) / scale))
	}
	return textureHash01(x, z, seed)
}

func textureHashSigned(x, z, seed int) matrix.Float {
	return textureHash01(x, z, seed)*2 - 1
}

func textureHash01(x, z, seed int) matrix.Float {
	n := uint32(x)*374761393 + uint32(z)*668265263 + uint32(seed)*1442695041
	n = (n ^ (n >> 13)) * 1274126177
	n ^= n >> 16
	return matrix.Float(n&0xffff) / matrix.Float(0xffff)
}

func normalizeAutoMaterialRule(rule AutoMaterialRule) AutoMaterialRule {
	if rule.TargetWeight <= 0 {
		rule.TargetWeight = 1
	}
	rule.TargetWeight = matrix.Clamp(rule.TargetWeight, 0, 1)
	rule.NoiseStrength = matrix.Clamp(rule.NoiseStrength, 0, 1)
	if rule.NoiseScale < 0 {
		rule.NoiseScale = 0
	}
	return rule
}

func (m *TextureWeightMap) paintLayer(layer int, stroke PaintStroke, erase bool) DirtyRegion {
	if m == nil || layer < 0 || layer >= m.Layers || stroke.Radius <= 0 || stroke.Strength == 0 {
		return DirtyRegion{}
	}
	region := m.strokeDirtyRegion(stroke)
	if !region.Valid {
		return DirtyRegion{}
	}
	strength := matrix.Abs(stroke.Strength)
	var dirty DirtyRegion
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			dx := matrix.Float(x) - stroke.Center.X()
			dz := matrix.Float(z) - stroke.Center.Y()
			distance := matrix.Sqrt(dx*dx + dz*dz)
			if distance > stroke.Radius {
				continue
			}
			if m.paintWeightAt(layer, x, z, strength*brushWeight(distance, stroke.Radius, stroke.Falloff), erase) {
				dirty = mergeDirtyRegions(dirty, DirtyRegion{
					MinX: x, MinZ: z, MaxX: x, MaxZ: z, Valid: true,
				})
			}
		}
	}
	return dirty
}

func (m *TextureWeightMap) paintWeightAt(layer, x, z int, amount matrix.Float, erase bool) bool {
	amount = matrix.Clamp(amount, 0, 1)
	m.NormalizeWeightsAt(x, z)
	m.cellScratch = m.weightsAtInto(x, z, m.cellScratch)
	before := m.cellScratch
	beforeTarget := before[layer]
	afterTarget := matrix.Lerp(beforeTarget, 1, amount)
	if erase {
		afterTarget = beforeTarget * (1 - amount)
	}
	beforeOtherSum := 1 - beforeTarget
	afterOtherSum := 1 - afterTarget
	for i := 0; i < m.Layers; i++ {
		if i == layer {
			m.Weights[m.index(i, x, z)] = afterTarget
		} else if beforeOtherSum > matrix.Tiny {
			m.Weights[m.index(i, x, z)] = before[i] * afterOtherSum / beforeOtherSum
		} else {
			m.Weights[m.index(i, x, z)] = 0
		}
	}
	if m.Layers > 1 && beforeOtherSum <= matrix.Tiny && afterOtherSum > matrix.Tiny {
		fallback := 0
		if fallback == layer {
			fallback = 1
		}
		m.Weights[m.index(fallback, x, z)] = afterOtherSum
	}
	m.NormalizeWeightsAt(x, z)
	return m.weightsChangedAt(x, z, before)
}

func (m *TextureWeightMap) strokeDirtyRegion(stroke PaintStroke) DirtyRegion {
	if m == nil || stroke.Radius <= 0 {
		return DirtyRegion{}
	}
	minX := max(0, int(matrix.Floor(stroke.Center.X()-stroke.Radius)))
	maxX := min(m.Resolution-1, int(matrix.Ceil(stroke.Center.X()+stroke.Radius)))
	minZ := max(0, int(matrix.Floor(stroke.Center.Y()-stroke.Radius)))
	maxZ := min(m.Resolution-1, int(matrix.Ceil(stroke.Center.Y()+stroke.Radius)))
	if minX > maxX || minZ > maxZ {
		return DirtyRegion{}
	}
	return DirtyRegion{MinX: minX, MinZ: minZ, MaxX: maxX, MaxZ: maxZ, Valid: true}
}

func (m *TextureWeightMap) inBounds(layer, x, z int) bool {
	return layer >= 0 && layer < m.Layers && x >= 0 && z >= 0 && x < m.Resolution && z < m.Resolution
}

func (m *TextureWeightMap) index(layer, x, z int) int {
	return (x+z*m.Resolution)*m.Layers + layer
}

func normalizeTerrainLayer(layer TerrainLayer) TerrainLayer {
	if layer.Name == "" {
		layer.Name = layer.TextureContentID
	}
	if layer.Filter < 0 || layer.Filter >= rendering.TextureFilterMax {
		layer.Filter = rendering.TextureFilterLinear
	}
	if layer.Tiling.X() == 0 && layer.Tiling.Y() == 0 {
		layer.Tiling = matrix.Vec2One()
	}
	if layer.Tint == matrix.ColorZero() {
		layer.Tint = matrix.ColorWhite()
	}
	if layer.TextureWorldSize.X() < 0 {
		layer.TextureWorldSize.SetX(0)
	}
	if layer.TextureWorldSize.Y() < 0 {
		layer.TextureWorldSize.SetY(0)
	}
	if layer.TriplanarSlope < 0 {
		layer.TriplanarSlope = 0
	}
	return layer
}
