/******************************************************************************/
/* layout_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
)

func testLayoutUI(width, height matrix.Float) *UI {
	entity := engine.NewEntity(nil)
	target := &UI{
		entity:  *entity,
		elmType: ElementTypePanel,
		elmData: &panelData{
			minSize: matrix.NewVec2(-1, -1),
			maxSize: matrix.NewVec2(-1, -1),
		},
	}
	target.layout.initialize(target)
	target.layout.Scale(width, height)
	return target
}

func TestLayoutClearStylesRemovesBoxModelSize(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	layout := target.Layout()

	layout.SetPadding(3, 2, 1, 4)
	layout.SetBorder(5, 7, 11, 13)

	if got := layout.PixelSize(); got.X() != 30 || got.Y() != 46 {
		t.Fatalf("padded and bordered size = %v, want [30 46]", got)
	}

	layout.ClearStyles()

	if got := layout.PixelSize(); got.X() != 10 || got.Y() != 20 {
		t.Fatalf("cleared size = %v, want [10 20]", got)
	}
	if got := layout.Padding(); got.X() != 0 || got.Y() != 0 || got.Z() != 0 || got.W() != 0 {
		t.Fatalf("padding after ClearStyles = %v, want zero", got)
	}
	if got := layout.Border(); got.X() != 0 || got.Y() != 0 || got.Z() != 0 || got.W() != 0 {
		t.Fatalf("border after ClearStyles = %v, want zero", got)
	}
}

func TestLayoutPaddingReapplyDoesNotAccumulateAfterClearStyles(t *testing.T) {
	t.Parallel()

	target := testLayoutUI(10, 20)
	layout := target.Layout()

	for range 3 {
		layout.ClearStyles()
		layout.SetPadding(3, 0, 0, 0)
	}

	if got := layout.PixelSize(); got.X() != 13 || got.Y() != 20 {
		t.Fatalf("reapplied padded size = %v, want [13 20]", got)
	}
}
