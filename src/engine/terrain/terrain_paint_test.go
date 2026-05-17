/******************************************************************************/
/* terrain_paint_test.go                                                      */
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
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestTerrainLayerSetDefaultsToHeightResolution(t *testing.T) {
	model, err := NewModel(TerrainConfig{Resolution: 5})
	if err != nil {
		t.Fatal(err)
	}
	if model.LayerSet == nil || model.LayerSet.WeightMap == nil {
		t.Fatal("expected terrain to create a layer set")
	}
	if got := model.LayerSet.WeightMap.Resolution; got != model.HeightField.Resolution {
		t.Fatalf("expected paint resolution to default to height resolution, got %d", got)
	}
	if got := model.LayerCount(); got != 0 {
		t.Fatalf("expected no default paint layers, got %d", got)
	}
}

func TestTerrainLayersAddAndStorePaintFields(t *testing.T) {
	set, err := NewTerrainLayerSet(4)
	if err != nil {
		t.Fatal(err)
	}
	base := set.AddLayer(TerrainLayer{
		TextureContentID:   "grass",
		NormalContentID:    "grass_normal",
		RoughnessContentID: "grass_roughness",
		Filter:             rendering.TextureFilterNearest,
		Tiling:             matrix.NewVec2(8, 6),
		Offset:             matrix.NewVec2(0.25, 0.5),
		Rotation:           0.125,
		Tint:               matrix.ColorGreen(),
	})
	rock := set.AddLayer(NewTerrainLayer("rock"))
	if base != 0 || rock != 1 {
		t.Fatalf("expected layer indexes 0 and 1, got %d and %d", base, rock)
	}
	if got := set.LayerCount(); got != 2 {
		t.Fatalf("expected two layers, got %d", got)
	}
	if set.Layers[0].TextureContentID != "grass" || set.Layers[0].NormalContentID != "grass_normal" {
		t.Fatalf("expected layer content ids to be preserved, got %+v", set.Layers[0])
	}
	if set.Layers[0].Filter != rendering.TextureFilterNearest {
		t.Fatalf("expected nearest filter, got %d", set.Layers[0].Filter)
	}
	if set.Layers[0].Tiling != matrix.NewVec2(8, 6) || set.Layers[0].Offset != matrix.NewVec2(0.25, 0.5) {
		t.Fatalf("expected tiling and offset to be preserved, got %+v", set.Layers[0])
	}
	if got := set.LayerWeightAt(base, 2, 2); !matrix.ApproxTo(got, 1, matrix.Roughly) {
		t.Fatalf("expected first layer to initialize to full weight, got %f", got)
	}
	if got := set.LayerWeightAt(rock, 2, 2); !matrix.ApproxTo(got, 0, matrix.Roughly) {
		t.Fatalf("expected added layer to initialize to zero weight, got %f", got)
	}
}

func TestTextureWeightMapNormalizeWeightsAt(t *testing.T) {
	set, err := NewTerrainLayerSet(3)
	if err != nil {
		t.Fatal(err)
	}
	layerA := set.AddLayer(NewTerrainLayer("a"))
	layerB := set.AddLayer(NewTerrainLayer("b"))
	set.SetLayerWeightAt(layerA, 1, 1, 0.25)
	set.SetLayerWeightAt(layerB, 1, 1, 0.5)
	if !set.NormalizeWeightsAt(1, 1) {
		t.Fatal("expected coordinate normalization to succeed")
	}
	if got := set.LayerWeightAt(layerA, 1, 1); !matrix.ApproxTo(got, matrix.Float(1.0/3.0), 0.001) {
		t.Fatalf("expected normalized layer A weight, got %f", got)
	}
	if got := set.LayerWeightAt(layerB, 1, 1); !matrix.ApproxTo(got, matrix.Float(2.0/3.0), 0.001) {
		t.Fatalf("expected normalized layer B weight, got %f", got)
	}
}

func TestTerrainPaintAndEraseLayerWeights(t *testing.T) {
	set, err := NewTerrainLayerSet(5)
	if err != nil {
		t.Fatal(err)
	}
	base := set.AddLayer(NewTerrainLayer("base"))
	paint := set.AddLayer(NewTerrainLayer("paint"))
	dirty := set.PaintLayer(paint, PaintStroke{
		Center:   matrix.NewVec2(2, 2),
		Radius:   1,
		Strength: 1,
		Falloff:  FalloffConstant,
	})
	if !dirty.Valid {
		t.Fatal("expected texture painting to return a dirty region")
	}
	if got := set.LayerWeightAt(paint, 2, 2); !matrix.ApproxTo(got, 1, matrix.Roughly) {
		t.Fatalf("expected painted layer weight 1 after normalization, got %f", got)
	}
	if got := set.LayerWeightAt(base, 2, 2); !matrix.ApproxTo(got, 0, matrix.Roughly) {
		t.Fatalf("expected base layer weight 0 after normalization, got %f", got)
	}
	dirty = set.EraseLayer(paint, PaintStroke{
		Center:   matrix.NewVec2(2, 2),
		Radius:   1,
		Strength: 1,
		Falloff:  FalloffConstant,
	})
	if !dirty.Valid {
		t.Fatal("expected texture erasing to return a dirty region")
	}
	if got := set.LayerWeightAt(paint, 2, 2); !matrix.ApproxTo(got, 0, matrix.Roughly) {
		t.Fatalf("expected erased layer weight 0, got %f", got)
	}
	if got := set.LayerWeightAt(base, 2, 2); !matrix.ApproxTo(got, 1, matrix.Roughly) {
		t.Fatalf("expected base layer weight 1 after erase normalization, got %f", got)
	}
}

func TestTerrainFillAndRemoveLayerWeights(t *testing.T) {
	set, err := NewTerrainLayerSet(3)
	if err != nil {
		t.Fatal(err)
	}
	base := set.AddLayer(NewTerrainLayer("base"))
	paint := set.AddLayer(NewTerrainLayer("paint"))
	dirty := set.FillLayer(paint)
	if !dirty.Valid {
		t.Fatal("expected fill to return dirty region")
	}
	for z := 0; z < set.WeightMap.Resolution; z++ {
		for x := 0; x < set.WeightMap.Resolution; x++ {
			if got := set.LayerWeightAt(paint, x, z); !matrix.ApproxTo(got, 1, matrix.Roughly) {
				t.Fatalf("expected filled layer weight at %d,%d to be 1, got %f", x, z, got)
			}
			if got := set.LayerWeightAt(base, x, z); !matrix.ApproxTo(got, 0, matrix.Roughly) {
				t.Fatalf("expected base layer weight at %d,%d to be 0, got %f", x, z, got)
			}
		}
	}
	if !set.RemoveLayer(base) {
		t.Fatal("expected remove layer to succeed")
	}
	if got := set.LayerCount(); got != 1 {
		t.Fatalf("expected one remaining layer, got %d", got)
	}
	if got := set.LayerWeightAt(0, 1, 1); !matrix.ApproxTo(got, 1, matrix.Roughly) {
		t.Fatalf("expected remaining layer to stay normalized, got %f", got)
	}
}
