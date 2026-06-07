/******************************************************************************/
/* font_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "testing"

func TestMSDFAtlasPxRangeMatchesGeneratedFonts(t *testing.T) {
	got := msdfAtlasPxRange()
	if got.X() != distanceFieldRange || got.Y() != distanceFieldRange {
		t.Fatalf("msdfAtlasPxRange = %v, want %v on both axes", got, distanceFieldRange)
	}
}
