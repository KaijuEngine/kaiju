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
