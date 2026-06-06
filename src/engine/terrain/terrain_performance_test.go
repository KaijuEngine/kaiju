/******************************************************************************/
/* terrain_performance_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package terrain

import (
	"fmt"
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

var productionTerrainResolutions = []int{257, 513, 1025}

func productionTerrainConfig(resolution int) TerrainConfig {
	return TerrainConfig{
		Resolution:    resolution,
		WorldSize:     matrix.NewVec2(512, 512),
		MinHeight:     -128,
		MaxHeight:     128,
		InitialHeight: 0,
		ChunkSize:     32,
	}
}

func BenchmarkProductionTerrainCreateModel(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			cfg := productionTerrainConfig(resolution)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				model, err := NewModel(cfg)
				if err != nil {
					b.Fatal(err)
				}
				if model.HeightField.Resolution != resolution {
					b.Fatalf("expected resolution %d, got %d", resolution, model.HeightField.Resolution)
				}
			}
		})
	}
}

func BenchmarkProductionTerrainPaintStroke(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrain(b, resolution)
			stroke := PaintStroke{
				Mode:     BrushRaise,
				Center:   matrix.NewVec2(0, 0),
				Radius:   12,
				Strength: 0.5,
				Falloff:  FalloffSmooth,
				Spacing:  3,
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%2 == 0 {
					stroke.Mode = BrushRaise
				} else {
					stroke.Mode = BrushLower
				}
				stroke.Center = matrix.NewVec2(matrix.Float(i%16)-8, matrix.Float((i/16)%16)-8)
				model.Paint(stroke)
			}
		})
	}
}

func BenchmarkProductionTerrainTexturePaintStroke(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrainWithLayers(b, resolution, 8)
			stroke := TexturePaintStroke{
				Mode:     TextureBrushPaint,
				Center:   matrix.NewVec2(0, 0),
				Radius:   18,
				Strength: 0.35,
				Falloff:  FalloffSmooth,
				Spacing:  4,
			}
			erase := stroke
			erase.Mode = TextureBrushErase
			model.PaintTextureLayer(3, stroke)
			model.PaintTextureLayer(3, erase)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stroke.Center = matrix.NewVec2(matrix.Float(i%32)-16, matrix.Float((i/32)%32)-16)
				erase.Center = stroke.Center
				model.PaintTextureLayer(3, stroke)
				model.PaintTextureLayer(3, erase)
			}
		})
	}
}

func BenchmarkProductionTerrainTextureDirtyRegionUpload(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrainWithLayers(b, resolution, 8)
			region := DirtyRegion{
				MinX:  resolution/2 - 16,
				MinZ:  resolution/2 - 16,
				MaxX:  resolution/2 + 16,
				MaxZ:  resolution/2 + 16,
				Valid: true,
			}
			bytesPerUpload := (region.MaxX - region.MinX + 1) *
				(region.MaxZ - region.MinZ + 1) * splatTextureChannels *
				splatTextureCount(model.LayerSet.WeightMap.Layers)
			model.MarkTextureRegionDirty(region)
			model.ApplyTextureDirty(region)
			for i := range model.SplatTextures {
				model.SplatTextures[i].Dirty = DirtyRegion{}
			}
			b.SetBytes(int64(bytesPerUpload))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				model.MarkTextureRegionDirty(region)
				model.ApplyTextureDirty(region)
			}
		})
	}
}

func BenchmarkProductionTerrainTextureLongStrokeManyLayers(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrainWithLayers(b, resolution, 16)
			stroke := TexturePaintStroke{
				Mode:     TextureBrushPaint,
				Radius:   20,
				Strength: 0.25,
				Falloff:  FalloffSmooth,
				Spacing:  5,
			}
			from := matrix.NewVec2(-192, -192)
			to := matrix.NewVec2(192, 192)
			model.PaintTextureLine(12, from, to, stroke)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				offset := matrix.Float(i%16) * 2
				model.PaintTextureLine(12,
					from.Add(matrix.NewVec2(offset, 0)),
					to.Add(matrix.NewVec2(0, offset)),
					stroke,
				)
			}
		})
	}
}

func BenchmarkProductionTerrainSmoothPaintLine(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrain(b, resolution)
			stroke := PaintStroke{
				Mode:     BrushSmooth,
				Radius:   10,
				Strength: 0.35,
				Falloff:  FalloffLinear,
				Spacing:  2.5,
			}
			from := matrix.NewVec2(-64, -64)
			to := matrix.NewVec2(64, 64)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				offset := matrix.Float(i%8) * 2
				model.PaintLine(from.Add(matrix.NewVec2(offset, 0)), to.Add(matrix.NewVec2(0, offset)), stroke)
			}
		})
	}
}

func BenchmarkProductionTerrainRayHitLocal(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrain(b, resolution)
			model.Paint(PaintStroke{
				Mode:     BrushRaise,
				Center:   matrix.NewVec2(0, 0),
				Radius:   40,
				Strength: 24,
				Falloff:  FalloffSmooth,
			})
			ray := graviton.Ray{
				Origin:    matrix.NewVec3(0, 96, 0),
				Direction: matrix.Vec3Down(),
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := model.RayHitLocal(ray); !ok {
					b.Fatal("expected terrain ray hit")
				}
			}
		})
	}
}

func BenchmarkProductionTerrainAssetSerialize(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrain(b, resolution)
			model.Paint(PaintStroke{
				Mode:     BrushRaise,
				Center:   matrix.NewVec2(0, 0),
				Radius:   32,
				Strength: 8,
				Falloff:  FalloffSmooth,
			})
			asset, err := NewAssetFromHeightField(model.Config, model.HeightField)
			if err != nil {
				b.Fatal(err)
			}
			b.SetBytes(int64(len(asset.Heights) * 2))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := asset.Serialize(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkProductionTerrainAssetDeserialize(b *testing.B) {
	for _, resolution := range productionTerrainResolutions {
		b.Run(fmt.Sprintf("resolution_%d", resolution), func(b *testing.B) {
			model := mustBenchmarkTerrain(b, resolution)
			asset, err := NewAssetFromHeightField(model.Config, model.HeightField)
			if err != nil {
				b.Fatal(err)
			}
			data, err := asset.Serialize()
			if err != nil {
				b.Fatal(err)
			}
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := DeserializeAsset(data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func mustBenchmarkTerrain(b *testing.B, resolution int) *Terrain {
	b.Helper()
	model, err := NewModel(productionTerrainConfig(resolution))
	if err != nil {
		b.Fatal(err)
	}
	return model
}

func mustBenchmarkTerrainWithLayers(b *testing.B, resolution, layers int) *Terrain {
	b.Helper()
	cfg := productionTerrainConfig(resolution)
	cfg.PaintResolution = resolution
	model, err := NewModel(cfg)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < layers; i++ {
		if got := model.AddLayer(NewTerrainLayer(fmt.Sprintf("layer_%d", i))); got != i {
			b.Fatalf("expected layer %d, got %d", i, got)
		}
	}
	return model
}
