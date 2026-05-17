/******************************************************************************/
/* terrain_workspace_test.go                                                  */
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
