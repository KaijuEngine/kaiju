/******************************************************************************/
/* terrain_content_preview.go                                                 */
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

package content_previews

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log/slog"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

const terrainPreviewSize = 128

func (p *ContentPreviewer) renderTerrain(id string) {
	defer tracing.NewRegion("ContentPreviewer.renderTerrain").End()
	defer p.completeProc()
	asset, err := readTerrain(id, p.ed)
	if err != nil {
		slog.Error("failed to generate a preview for terrain", "id", id, "error", err)
		return
	}
	img := terrainPreviewImage(asset)
	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		slog.Error("failed to encode the terrain preview image", "id", id, "error", err)
		return
	}
	if err = p.writePreviewFile(id, buf.Bytes()); err != nil {
		slog.Error("failed to write the terrain preview image cache file", "id", id, "error", err)
		return
	}
	p.ed.Events().OnContentPreviewGenerated.Execute(id)
}

func readTerrain(id string, ed EditorInterface) (terrain.TerrainAsset, error) {
	defer tracing.NewRegion("content_previews.readTerrain").End()
	cc, err := ed.Cache().Read(id)
	if err != nil {
		return terrain.TerrainAsset{}, err
	}
	if cc.Config.Type != (content_database.Terrain{}).TypeName() {
		return terrain.TerrainAsset{},
			fmt.Errorf("can't generate a terrain preview image for content, the provided id '%s' is not terrain", id)
	}
	data, err := ed.ProjectFileSystem().ReadFile(cc.ContentPath())
	if err != nil {
		return terrain.TerrainAsset{}, err
	}
	return terrain.DeserializeAsset(data)
}

func terrainPreviewImage(asset terrain.TerrainAsset) image.Image {
	size := min(terrainPreviewSize, max(2, asset.Config.Resolution))
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	maxIdx := asset.Config.Resolution - 1
	layerSet, _ := asset.LayerSet()
	for y := 0; y < size; y++ {
		z := y * maxIdx / (size - 1)
		for x := 0; x < size; x++ {
			gx := x * maxIdx / (size - 1)
			h := asset.Heights[gx+z*asset.Config.Resolution]
			left := terrainPreviewHeight(asset, gx-1, z)
			right := terrainPreviewHeight(asset, gx+1, z)
			front := terrainPreviewHeight(asset, gx, z-1)
			back := terrainPreviewHeight(asset, gx, z+1)
			slope := uint8(min(48, int((absTerrainFloat(right-left)+absTerrainFloat(back-front))*6)))
			v := uint8(h >> 8)
			base := terrainPreviewLayerColor(layerSet, x, y, size)
			shade := matrix.Float(v)/255*0.45 + 0.55 + matrix.Float(slope)/255
			r := clampByte(int(base.R()*255*shade) - 18 + int(slope))
			g := clampByte(int(base.G()*255*shade) + 8 + int(slope/2))
			b := clampByte(int(base.B()*255*shade) - 20)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func terrainPreviewLayerColor(layerSet *terrain.TerrainLayerSet, x, y, previewSize int) matrix.Color {
	if layerSet == nil || layerSet.WeightMap == nil || layerSet.WeightMap.Layers == 0 || previewSize <= 1 {
		return matrix.ColorRGBInt(92, 118, 77)
	}
	weights := layerSet.WeightMap
	wx := x * (weights.Resolution - 1) / (previewSize - 1)
	wz := y * (weights.Resolution - 1) / (previewSize - 1)
	var out matrix.Color
	var sum matrix.Float
	for layer := 0; layer < weights.Layers && layer < len(layerSet.Layers); layer++ {
		weight := weights.WeightAt(layer, wx, wz)
		if weight <= matrix.Tiny {
			continue
		}
		tint := layerSet.Layers[layer].Tint
		if tint == matrix.ColorZero() || tint == matrix.ColorWhite() {
			tint = terrainPreviewFallbackLayerColor(layer)
		}
		out.SetR(out.R() + tint.R()*weight)
		out.SetG(out.G() + tint.G()*weight)
		out.SetB(out.B() + tint.B()*weight)
		sum += weight
	}
	if sum <= matrix.Tiny {
		return terrainPreviewFallbackLayerColor(0)
	}
	out.SetR(out.R() / sum)
	out.SetG(out.G() / sum)
	out.SetB(out.B() / sum)
	out.SetA(1)
	return out
}

func terrainPreviewFallbackLayerColor(layer int) matrix.Color {
	palette := []matrix.Color{
		matrix.ColorRGBInt(88, 126, 62),
		matrix.ColorRGBInt(116, 113, 106),
		matrix.ColorRGBInt(204, 207, 194),
		matrix.ColorRGBInt(137, 96, 54),
		matrix.ColorRGBInt(71, 106, 122),
		matrix.ColorRGBInt(113, 95, 137),
	}
	return palette[layer%len(palette)]
}

func terrainPreviewHeight(asset terrain.TerrainAsset, x, z int) matrix.Float {
	x = min(max(x, 0), asset.Config.Resolution-1)
	z = min(max(z, 0), asset.Config.Resolution-1)
	return asset.Height(x, z)
}

func absTerrainFloat(v matrix.Float) matrix.Float {
	if v < 0 {
		return -v
	}
	return v
}

func clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
