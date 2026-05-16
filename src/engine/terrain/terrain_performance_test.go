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
