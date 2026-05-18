/******************************************************************************/
/* terrain_splat_texture_test.go                                              */
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
)

func TestSplatLayerChannelSelection(t *testing.T) {
	tests := []struct {
		layer       int
		wantTexture int
		wantChannel int
	}{
		{layer: 0, wantTexture: 0, wantChannel: 0},
		{layer: 3, wantTexture: 0, wantChannel: 3},
		{layer: 4, wantTexture: 1, wantChannel: 0},
		{layer: 7, wantTexture: 1, wantChannel: 3},
		{layer: 8, wantTexture: 2, wantChannel: 0},
	}
	for _, tt := range tests {
		got, ok := splatLayerChannel(tt.layer, 9)
		if !ok {
			t.Fatalf("expected layer %d to map to a splat channel", tt.layer)
		}
		if got.Texture != tt.wantTexture || got.Channel != tt.wantChannel {
			t.Fatalf("expected layer %d to map to texture %d channel %d, got %+v",
				tt.layer, tt.wantTexture, tt.wantChannel, got)
		}
	}
	if _, ok := splatLayerChannel(9, 9); ok {
		t.Fatal("expected out-of-range layer to fail channel lookup")
	}
}

func TestPackSplatTextureRegionUsesLayerChannels(t *testing.T) {
	weights, err := NewTextureWeightMap(4, 6)
	if err != nil {
		t.Fatal(err)
	}
	weights.SetWeightAt(4, 1, 2, 0.25)
	weights.SetWeightAt(5, 1, 2, 0.75)
	weights.SetWeightAt(4, 2, 2, 1)
	region := DirtyRegion{MinX: 1, MinZ: 2, MaxX: 2, MaxZ: 2, Valid: true}
	pixels := packSplatTextureRegion(weights, 1, region)
	if len(pixels) != 2*splatTextureChannels {
		t.Fatalf("expected two RGBA pixels, got %d bytes", len(pixels))
	}
	if pixels[0] != 64 || pixels[1] != 191 || pixels[2] != 0 || pixels[3] != 0 {
		t.Fatalf("expected first packed pixel to contain layer 4/5 weights, got %v", pixels[:4])
	}
	if pixels[4] != 255 || pixels[5] != 0 || pixels[6] != 0 || pixels[7] != 0 {
		t.Fatalf("expected second packed pixel to contain layer 4 weight, got %v", pixels[4:8])
	}
}

func TestTerrainTexturePaintingTracksSplatDirtyWithoutMeshDirty(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:      5,
		PaintResolution: 5,
		WorldSize:       matrix.NewVec2(100, 100),
		MinHeight:       0,
		MaxHeight:       1,
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		model.AddLayer(NewTerrainLayer("layer"))
	}
	for i := range model.SplatTextures {
		model.SplatTextures[i].Dirty = DirtyRegion{}
	}
	model.HeightField.ClearDirty()
	dirty := model.PaintLayer(5, PaintStroke{
		Center:   matrix.NewVec2(0, 0),
		Radius:   25,
		Strength: 1,
		Falloff:  FalloffConstant,
	})
	if !dirty.Valid {
		t.Fatal("expected texture paint to return a dirty region")
	}
	if model.HeightField.DirtyRegion().Valid {
		t.Fatal("texture-only painting should not dirty or rebuild terrain mesh vertices")
	}
	if got := model.SplatTextureCount(); got != 2 {
		t.Fatalf("expected six layers to pack into two splat textures, got %d", got)
	}
	for i := range model.SplatTextures {
		if model.SplatTextures[i].Dirty != dirty {
			t.Fatalf("expected splat texture %d to track dirty region %+v, got %+v",
				i, dirty, model.SplatTextures[i].Dirty)
		}
	}
	request := model.SplatTextureWriteRequest(1, DirtyRegion{MinX: 2, MinZ: 2, MaxX: 2, MaxZ: 2, Valid: true})
	if request.Region != (matrix.Vec4i{2, 2, 1, 1}) {
		t.Fatalf("expected one-pixel write at 2,2, got %v", request.Region)
	}
	if len(request.Pixels) != splatTextureChannels || request.Pixels[1] != 255 {
		t.Fatalf("expected layer 5 to pack into splat texture 1 channel 1, got %v", request.Pixels)
	}
}

func TestTerrainTexturePaintStrokeDoesNotAllocateAfterWarmup(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:      33,
		PaintResolution: 257,
		WorldSize:       matrix.NewVec2(256, 256),
	})
	if err != nil {
		t.Fatal(err)
	}
	model.AddLayer(NewTerrainLayer("base"))
	paint := model.AddLayer(NewTerrainLayer("paint"))
	for i := 0; i < 6; i++ {
		model.AddLayer(NewTerrainLayer("extra"))
	}
	stroke := TexturePaintStroke{
		Mode:     TextureBrushPaint,
		Center:   matrix.NewVec2(0, 0),
		Radius:   16,
		Strength: 0.25,
		Falloff:  FalloffSmooth,
		Spacing:  4,
	}
	erase := stroke
	erase.Mode = TextureBrushErase
	model.PaintTextureLayer(paint, stroke)
	model.PaintTextureLayer(paint, erase)
	allocs := testing.AllocsPerRun(100, func() {
		model.PaintTextureLayer(paint, stroke)
		model.PaintTextureLayer(paint, erase)
	})
	if allocs != 0 {
		t.Fatalf("expected warmed texture brush strokes to avoid allocations, got %.2f", allocs)
	}
}

func TestTerrainTexturePaintLineMergesSplatDirtyRegions(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:      33,
		PaintResolution: 257,
		WorldSize:       matrix.NewVec2(256, 256),
	})
	if err != nil {
		t.Fatal(err)
	}
	model.AddLayer(NewTerrainLayer("base"))
	paint := model.AddLayer(NewTerrainLayer("paint"))
	for i := range model.SplatTextures {
		model.SplatTextures[i].Dirty = DirtyRegion{}
	}
	model.HeightField.ClearDirty()
	result := model.PaintTextureLine(
		paint,
		matrix.NewVec2(-96, -96),
		matrix.NewVec2(96, 96),
		TexturePaintStroke{
			Mode:     TextureBrushPaint,
			Radius:   12,
			Strength: 0.5,
			Falloff:  FalloffSmooth,
			Spacing:  3,
		},
	)
	if !result.Dirty.Valid {
		t.Fatal("expected long texture stroke to dirty the weight map")
	}
	for i := range model.SplatTextures {
		if model.SplatTextures[i].Dirty != result.Dirty {
			t.Fatalf("expected splat texture %d to keep merged dirty region %+v, got %+v",
				i, result.Dirty, model.SplatTextures[i].Dirty)
		}
	}
	if model.HeightField.DirtyRegion().Valid {
		t.Fatal("texture paint line should not dirty or rebuild terrain mesh vertices")
	}
}

func TestTerrainTexturePaintLongStrokeManyLayersStress(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:      33,
		PaintResolution: 257,
		WorldSize:       matrix.NewVec2(256, 256),
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 16; i++ {
		model.AddLayer(NewTerrainLayer("layer"))
	}
	for i := range model.SplatTextures {
		model.SplatTextures[i].Dirty = DirtyRegion{}
	}
	model.HeightField.ClearDirty()
	result := model.PaintTextureLine(
		12,
		matrix.NewVec2(-120, -120),
		matrix.NewVec2(120, 120),
		TexturePaintStroke{
			Mode:     TextureBrushPaint,
			Radius:   18,
			Strength: 0.4,
			Falloff:  FalloffSmooth,
			Spacing:  4,
		},
	)
	if !result.Dirty.Valid {
		t.Fatal("expected long many-layer stroke to dirty weights")
	}
	if got := model.SplatTextureCount(); got != 4 {
		t.Fatalf("expected sixteen layers to use four splat textures, got %d", got)
	}
	if model.HeightField.DirtyRegion().Valid {
		t.Fatal("many-layer texture stroke should not dirty terrain mesh vertices")
	}
	assertTextureWeightsNormalized(t, model.LayerSet.WeightMap)
}
