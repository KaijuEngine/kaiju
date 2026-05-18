/******************************************************************************/
/* terrain_content_preview_test.go                                            */
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
	"image"
	"testing"

	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/matrix"
)

func TestTerrainPreviewImageUsesPaintedLayerWeights(t *testing.T) {
	cfg := terrain.TerrainConfig{
		Resolution:      4,
		PaintResolution: 4,
		WorldSize:       matrix.NewVec2(4, 4),
		MinHeight:       0,
		MaxHeight:       1,
	}
	set, err := terrain.NewTerrainLayerSet(cfg.PaintResolution)
	if err != nil {
		t.Fatal(err)
	}
	red := set.AddLayer(terrain.TerrainLayer{
		Name:             "Red",
		TextureContentID: "red",
		Tint:             matrix.ColorRed(),
	})
	blue := set.AddLayer(terrain.TerrainLayer{
		Name:             "Blue",
		TextureContentID: "blue",
		Tint:             matrix.ColorBlue(),
	})
	for z := 0; z < set.WeightMap.Resolution; z++ {
		for x := 0; x < set.WeightMap.Resolution; x++ {
			if x < set.WeightMap.Resolution/2 {
				set.SetLayerWeightAt(red, x, z, 1)
				set.SetLayerWeightAt(blue, x, z, 0)
			} else {
				set.SetLayerWeightAt(red, x, z, 0)
				set.SetLayerWeightAt(blue, x, z, 1)
			}
		}
	}
	asset, err := terrain.NewAssetWithLayerSet(cfg, nil, set)
	if err != nil {
		t.Fatal(err)
	}
	img := terrainPreviewImage(asset).(*image.RGBA)
	left := img.RGBAAt(0, 1)
	right := img.RGBAAt(img.Bounds().Dx()-1, 1)
	if left.R <= left.B {
		t.Fatalf("expected left painted preview pixel to be red-dominant, got %+v", left)
	}
	if right.B <= right.R {
		t.Fatalf("expected right painted preview pixel to be blue-dominant, got %+v", right)
	}
}
