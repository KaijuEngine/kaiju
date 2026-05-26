/******************************************************************************/
/* stage_picking_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"reflect"
	"testing"

	"kaijuengine.com/matrix"
)

func TestPickingPointReadRegionScalesViewportLocalBottomToTextureTop(t *testing.T) {
	region, ok := pickingPointReadRegion(
		matrix.NewVec2(25, 10),
		matrix.NewVec2(200, 100),
		400,
		200,
	)
	if !ok {
		t.Fatalf("expected point region to be valid")
	}
	want := matrix.Vec4i{50, 179, 1, 1}
	if region != want {
		t.Fatalf("region = %v, want %v", region, want)
	}
}

func TestPickingBoxReadRegionNormalizesInvertedDragAndClamps(t *testing.T) {
	region, ok := pickingBoxReadRegion(
		matrix.NewVec4(180, 90, -20, -10),
		matrix.NewVec2(100, 50),
		200,
		100,
	)
	if !ok {
		t.Fatalf("expected clamped region to be valid")
	}
	want := matrix.Vec4i{0, 0, 200, 100}
	if region != want {
		t.Fatalf("region = %v, want %v", region, want)
	}
}

func TestPickingBoxReadRegionConvertsPartialInvertedDrag(t *testing.T) {
	region, ok := pickingBoxReadRegion(
		matrix.NewVec4(80, 45, 20, 5),
		matrix.NewVec2(100, 50),
		200,
		100,
	)
	if !ok {
		t.Fatalf("expected box region to be valid")
	}
	want := matrix.Vec4i{40, 10, 120, 80}
	if region != want {
		t.Fatalf("region = %v, want %v", region, want)
	}
}

func TestDecodePickIDsIgnoresZeroAndDeduplicates(t *testing.T) {
	data := []byte{
		0, 0, 0, 0,
		7, 0, 0, 0,
		9, 0, 0, 0,
		7, 0, 0, 0,
	}
	got := decodePickIDs(data)
	want := []uint32{7, 9}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
}
