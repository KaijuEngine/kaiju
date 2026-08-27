/******************************************************************************/
/* shader_data_ocean_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"testing"
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestShaderDataOceanLayoutAndDefaults(t *testing.T) {
	data := Create("ocean").(*ShaderDataOcean)
	wantSize := int(rendering.ShaderBaseDataSize + 6*unsafe.Sizeof(matrix.Vec4{}) +
		unsafe.Sizeof(StandardShaderDataFlags(0)) + unsafe.Sizeof([4]int32{}))
	if data.Size() != wantSize {
		t.Fatalf("ocean shader data size = %d, want %d", data.Size(), wantSize)
	}
	if data.LightIds != [4]int32{-1, -1, -1, -1} {
		t.Fatalf("ocean light IDs = %v", data.LightIds)
	}
	if data.WaveParams.X() <= 0 || data.WaveParams.X() >= 0.5 {
		t.Fatalf("ocean default amplitude is not subtle: %v", data.WaveParams.X())
	}
	data.SetBrush(matrix.NewVec2(2, 3), 0.75, 0.04, matrix.ColorWhite())
	if data.BrushCenterRadius != matrix.NewVec4(2, 3, 0.75, 1) || data.BrushParams.X() != 0.04 {
		t.Fatalf("ocean brush preview was not packed correctly: %+v, %+v",
			data.BrushCenterRadius, data.BrushParams)
	}
	data.ClearBrush()
	if data.BrushCenterRadius.W() != 0 {
		t.Fatal("ocean brush preview did not clear")
	}
}
