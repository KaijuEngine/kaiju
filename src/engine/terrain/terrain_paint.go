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

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type TerrainLayer struct {
	TextureContentID   string
	NormalContentID    string
	RoughnessContentID string
	Filter             rendering.TextureFilter
	Tiling             matrix.Vec2
	Offset             matrix.Vec2
	Rotation           matrix.Float
	Tint               matrix.Color
}

type TerrainLayerSet struct {
	Layers    []TerrainLayer
	WeightMap *TextureWeightMap
}

type TextureWeightMap struct {
	Resolution int
	Layers     int
	Weights    []matrix.Float
}

func NewTerrainLayer(textureContentID string) TerrainLayer {
	return TerrainLayer{
		TextureContentID: textureContentID,
		Filter:           rendering.TextureFilterLinear,
		Tiling:           matrix.Vec2One(),
		Tint:             matrix.ColorWhite(),
	}
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
	return s.WeightMap.PaintLayer(layer, stroke)
}

func (s *TerrainLayerSet) EraseLayer(layer int, stroke PaintStroke) DirtyRegion {
	if s == nil || s.WeightMap == nil {
		return DirtyRegion{}
	}
	return s.WeightMap.EraseLayer(layer, stroke)
}

func (s *TerrainLayerSet) FillLayer(layer int) DirtyRegion {
	if s == nil || s.WeightMap == nil {
		return DirtyRegion{}
	}
	return s.WeightMap.FillLayer(layer)
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
	if m == nil || fillLayer < 0 || fillLayer >= m.Layers {
		return DirtyRegion{}
	}
	for z := 0; z < m.Resolution; z++ {
		for x := 0; x < m.Resolution; x++ {
			for layer := 0; layer < m.Layers; layer++ {
				weight := matrix.Float(0)
				if layer == fillLayer {
					weight = 1
				}
				m.Weights[m.index(layer, x, z)] = weight
			}
		}
	}
	return DirtyRegion{MinX: 0, MinZ: 0, MaxX: m.Resolution - 1, MaxZ: m.Resolution - 1, Valid: true}
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
	before := make([]matrix.Float, m.Layers)
	for i := 0; i < m.Layers; i++ {
		before[i] = m.WeightAt(i, x, z)
	}
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
	for i := 0; i < m.Layers; i++ {
		if !matrix.ApproxTo(before[i], m.WeightAt(i, x, z), matrix.Roughly) {
			return true
		}
	}
	return false
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
	if layer.Filter < 0 || layer.Filter >= rendering.TextureFilterMax {
		layer.Filter = rendering.TextureFilterLinear
	}
	if layer.Tiling.X() == 0 && layer.Tiling.Y() == 0 {
		layer.Tiling = matrix.Vec2One()
	}
	if layer.Tint == matrix.ColorZero() {
		layer.Tint = matrix.ColorWhite()
	}
	return layer
}
