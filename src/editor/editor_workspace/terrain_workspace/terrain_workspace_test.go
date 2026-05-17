package terrain_workspace

import (
	"testing"

	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/matrix"
)

func TestAdjustTerrainBrushValueScalesAndClamps(t *testing.T) {
	if got := adjustTerrainBrushValue(10, 1, 1, 20); !matrix.ApproxTo(got, 11, matrix.Roughly) {
		t.Fatalf("expected increased brush value to be 11, got %f", got)
	}
	if got := adjustTerrainBrushValue(11, -1, 1, 20); !matrix.ApproxTo(got, 10, matrix.Roughly) {
		t.Fatalf("expected decreased brush value to be 10, got %f", got)
	}
	if got := adjustTerrainBrushValue(1, -1, 1, 20); got != 1 {
		t.Fatalf("expected brush value to clamp to min 1, got %f", got)
	}
	if got := adjustTerrainBrushValue(20, 1, 1, 20); got != 20 {
		t.Fatalf("expected brush value to clamp to max 20, got %f", got)
	}
}

func TestEffectiveTerrainBrushModeModifiers(t *testing.T) {
	if got := effectiveTerrainBrushMode(terrain.BrushRaise, true, true); got != terrain.BrushSmooth {
		t.Fatalf("shift should temporarily smooth, got %d", got)
	}
	if got := effectiveTerrainBrushMode(terrain.BrushRaise, false, true); got != terrain.BrushLower {
		t.Fatalf("ctrl should invert raise to lower, got %d", got)
	}
	if got := effectiveTerrainBrushMode(terrain.BrushLower, false, true); got != terrain.BrushRaise {
		t.Fatalf("ctrl should invert lower to raise, got %d", got)
	}
	if got := effectiveTerrainBrushMode(terrain.BrushSmooth, false, true); got != terrain.BrushSmooth {
		t.Fatalf("ctrl should leave smooth unchanged, got %d", got)
	}
}
