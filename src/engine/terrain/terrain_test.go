package terrain

import (
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

func TestHeightFieldBilinearSampling(t *testing.T) {
	field, err := NewHeightField(2, 0, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	field.SetHeight(0, 0, 0)
	field.SetHeight(1, 0, 10)
	field.SetHeight(0, 1, 10)
	field.SetHeight(1, 1, 0)
	if got := field.Sample(0.5, 0.5); !matrix.ApproxTo(got, 5, matrix.Roughly) {
		t.Fatalf("expected bilinear center sample to be 5, got %f", got)
	}
	if got := field.Sample(-10, 10); !matrix.ApproxTo(got, 10, matrix.Roughly) {
		t.Fatalf("expected clamped edge sample to be 10, got %f", got)
	}
}

func TestHeightFieldClampsWritesAndTracksDirtyRegion(t *testing.T) {
	field, err := NewHeightField(4, -1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	field.ClearDirty()
	if !field.SetHeight(1, 2, 4) {
		t.Fatal("expected write to change height")
	}
	if got := field.Height(1, 2); got != 1 {
		t.Fatalf("expected height to clamp to 1, got %f", got)
	}
	if field.SetHeight(-1, 2, 0.5) {
		t.Fatal("out of bounds write should not dirty the heightfield")
	}
	dirty := field.DirtyRegion()
	expected := DirtyRegion{MinX: 1, MinZ: 2, MaxX: 1, MaxZ: 2, Valid: true}
	if dirty != expected {
		t.Fatalf("expected dirty region %+v, got %+v", expected, dirty)
	}
}

func TestPaintRaiseFalloff(t *testing.T) {
	field, err := NewHeightField(5, 0, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	field.ClearDirty()
	dirty := field.Paint(PaintStroke{
		Mode:     BrushRaise,
		Center:   matrix.NewVec2(2, 2),
		Radius:   2,
		Strength: 4,
		Falloff:  FalloffLinear,
	})
	if got := field.Height(2, 2); !matrix.ApproxTo(got, 4, matrix.Roughly) {
		t.Fatalf("expected center raise of 4, got %f", got)
	}
	if got := field.Height(3, 2); !matrix.ApproxTo(got, 2, matrix.Roughly) {
		t.Fatalf("expected halfway falloff raise of 2, got %f", got)
	}
	if got := field.Height(4, 2); !matrix.ApproxTo(got, 0, matrix.Roughly) {
		t.Fatalf("expected edge falloff to reach 0, got %f", got)
	}
	if !dirty.Valid {
		t.Fatal("expected paint to return dirty region")
	}
}

func TestPaintSmoothUsesNeighborAverage(t *testing.T) {
	field, err := NewHeightField(3, 0, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	field.SetHeight(1, 1, 9)
	field.ClearDirty()
	field.Paint(PaintStroke{
		Mode:     BrushSmooth,
		Center:   matrix.NewVec2(1, 1),
		Radius:   0.5,
		Strength: 1,
		Falloff:  FalloffConstant,
	})
	if got := field.Height(1, 1); !matrix.ApproxTo(got, 1, matrix.Roughly) {
		t.Fatalf("expected center to smooth to 3x3 average of 1, got %f", got)
	}
}

func TestPaintLineDefaultSpacingUsesQuarterRadius(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:    5,
		WorldSize:     matrix.NewVec2(4, 4),
		MinHeight:     0,
		MaxHeight:     10,
		InitialHeight: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	model.VisitPaintLineStamps(matrix.NewVec2(0, 0), matrix.NewVec2(1, 0), PaintStroke{
		Radius:   1,
		Strength: 1,
	}, func(PaintStroke) bool {
		count++
		return true
	})
	if count != 5 {
		t.Fatalf("expected 5 stamps for distance 1 at radius*0.25 spacing, got %d", count)
	}
}

func TestHeightRegionCopyAndApplyRestoresHeights(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:    4,
		WorldSize:     matrix.NewVec2(4, 4),
		MinHeight:     0,
		MaxHeight:     10,
		InitialHeight: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	region := DirtyRegion{MinX: 1, MinZ: 1, MaxX: 2, MaxZ: 2, Valid: true}
	before := model.HeightField.CopyRegion(region)
	model.HeightField.SetHeight(1, 1, 5)
	model.HeightField.SetHeight(2, 2, 7)
	model.ApplyHeightRegion(region, before)
	for z := region.MinZ; z <= region.MaxZ; z++ {
		for x := region.MinX; x <= region.MaxX; x++ {
			if got := model.HeightField.Height(x, z); got != 0 {
				t.Fatalf("expected restored height at %d,%d to be 0, got %f", x, z, got)
			}
		}
	}
}

func TestTerrainRayHitUsesHeightField(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:    3,
		WorldSize:     matrix.NewVec2(2, 2),
		MinHeight:     0,
		MaxHeight:     10,
		InitialHeight: 0,
		ChunkSize:     2,
	})
	if err != nil {
		t.Fatal(err)
	}
	model.HeightField.SetHeight(1, 1, 4)
	hit, ok := model.RayHitLocal(graviton.Ray{
		Origin:    matrix.NewVec3(0, 8, 0),
		Direction: matrix.Vec3Down(),
	})
	if !ok {
		t.Fatal("expected downward ray to hit raised terrain")
	}
	if !matrix.ApproxTo(hit.LocalPoint.Y(), 4, 0.02) {
		t.Fatalf("expected ray to hit height 4, got %f", hit.LocalPoint.Y())
	}
	if !matrix.ApproxTo(hit.Distance, 4, 0.02) {
		t.Fatalf("expected hit distance 4, got %f", hit.Distance)
	}
}

func TestTerrainChunkMeshDataComesFromHeightField(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:    3,
		WorldSize:     matrix.NewVec2(2, 2),
		MinHeight:     0,
		MaxHeight:     10,
		InitialHeight: 0,
		ChunkSize:     2,
	})
	if err != nil {
		t.Fatal(err)
	}
	model.HeightField.SetHeight(1, 1, 3)
	chunk := TerrainChunk{StartX: 0, StartZ: 0, EndX: 2, EndZ: 2}
	verts, indexes := model.buildChunkMeshData(&chunk)
	if len(verts) != 9 {
		t.Fatalf("expected 9 grid vertices, got %d", len(verts))
	}
	if len(indexes) != 24 {
		t.Fatalf("expected 24 triangle indexes, got %d", len(indexes))
	}
	if !matrix.ApproxTo(verts[4].Position.Y(), 3, matrix.Roughly) {
		t.Fatalf("expected center vertex height 3, got %f", verts[4].Position.Y())
	}
	if verts[3].Normal.Equals(matrix.Vec3Up()) {
		t.Fatal("expected adjacent vertex to produce an edited normal")
	}
}

func TestTerrainChunkIndexesFaceUp(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:    2,
		WorldSize:     matrix.NewVec2(2, 2),
		MinHeight:     0,
		MaxHeight:     10,
		InitialHeight: 0,
		ChunkSize:     1,
	})
	if err != nil {
		t.Fatal(err)
	}
	chunk := TerrainChunk{StartX: 0, StartZ: 0, EndX: 1, EndZ: 1}
	verts, indexes := model.buildChunkMeshData(&chunk)
	if len(indexes) != 6 {
		t.Fatalf("expected 6 indexes, got %d", len(indexes))
	}
	for i := 0; i < len(indexes); i += 3 {
		a := verts[indexes[i]].Position
		b := verts[indexes[i+1]].Position
		c := verts[indexes[i+2]].Position
		normal := b.Subtract(a).Cross(c.Subtract(a)).Normal()
		if normal.Dot(matrix.Vec3Up()) <= 0 {
			t.Fatalf("expected triangle %d to face up, got normal %v from indexes %v", i/3, normal, indexes[i:i+3])
		}
	}
}

func TestDirtyRegionExpandPadsNormals(t *testing.T) {
	dirty := DirtyRegion{MinX: 1, MinZ: 2, MaxX: 3, MaxZ: 4, Valid: true}
	got := dirty.Expand(1, 5)
	expected := DirtyRegion{MinX: 0, MinZ: 1, MaxX: 4, MaxZ: 4, Valid: true}
	if got != expected {
		t.Fatalf("expected expanded region %+v, got %+v", expected, got)
	}
}

func TestTerrainAssetRoundTripsUint16Heights(t *testing.T) {
	config := TerrainConfig{
		Resolution:    2,
		WorldSize:     matrix.NewVec2(10, 10),
		MinHeight:     -10,
		MaxHeight:     30,
		InitialHeight: 0,
	}
	asset, err := NewAsset(config, []matrix.Float{-10, 0, 10, 30})
	if err != nil {
		t.Fatal(err)
	}
	if len(asset.Heights) != 4 {
		t.Fatalf("expected four normalized heights, got %d", len(asset.Heights))
	}
	if asset.Heights[0] != 0 || asset.Heights[3] != 65535 {
		t.Fatalf("expected min/max heights to use full uint16 range, got %d and %d", asset.Heights[0], asset.Heights[3])
	}
	data, err := asset.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := DeserializeAsset(data)
	if err != nil {
		t.Fatal(err)
	}
	heights := loaded.FloatHeights()
	want := []matrix.Float{-10, 0, 10, 30}
	for i := range want {
		if !matrix.ApproxTo(heights[i], want[i], 0.001) {
			t.Fatalf("expected height %d to be %f, got %f", i, want[i], heights[i])
		}
	}
}

func TestTerrainAssetBuildsModelFromStoredHeights(t *testing.T) {
	asset, err := NewAsset(TerrainConfig{
		Resolution:    3,
		WorldSize:     matrix.NewVec2(2, 2),
		MinHeight:     0,
		MaxHeight:     10,
		InitialHeight: 0,
		ChunkSize:     2,
	}, []matrix.Float{
		0, 0, 0,
		0, 6, 0,
		0, 0, 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	model, err := newTerrainWithHeights(asset.Config, asset.FloatHeights(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !matrix.ApproxTo(model.HeightField.Height(1, 1), 6, 0.001) {
		t.Fatalf("expected center height from asset to be 6, got %f", model.HeightField.Height(1, 1))
	}
}

func TestTerrainCollisionUsesConfiguredHeightBounds(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:    2,
		WorldSize:     matrix.NewVec2(8, 6),
		MinHeight:     -3,
		MaxHeight:     9,
		InitialHeight: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	collision, err := model.Collision()
	if err != nil {
		t.Fatal(err)
	}
	bounds := collision.LocalBounds()
	if !matrix.Vec3ApproxTo(bounds.Center, matrix.NewVec3(0, 3, 0), 0.0001) {
		t.Fatalf("expected configured bounds center 0,3,0, got %v", bounds.Center)
	}
	if !matrix.Vec3ApproxTo(bounds.Extent, matrix.NewVec3(4, 6, 3), 0.0001) {
		t.Fatalf("expected configured bounds extent 4,6,3, got %v", bounds.Extent)
	}
}

func TestTerrainNewCollisionSharesHeightStorage(t *testing.T) {
	model, err := NewModel(TerrainConfig{
		Resolution:    2,
		WorldSize:     matrix.NewVec2(8, 6),
		MinHeight:     -3,
		MaxHeight:     9,
		InitialHeight: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	collision := model.NewCollision()
	if collision == nil {
		t.Fatal("expected terrain collision")
	}
	if &collision.Heights[0] != &model.HeightField.Heights[0] {
		t.Fatal("expected terrain collision to share height storage")
	}
	model.HeightField.SetHeight(1, 1, 5)
	if got := collision.Height(1, 1); got != 5 {
		t.Fatalf("expected shared collision height to update to 5, got %f", got)
	}
	if got, want := model.CollisionBounds(), collision.LocalBounds(); got != want {
		t.Fatalf("expected collision bounds %v, got %v", want, got)
	}
}
