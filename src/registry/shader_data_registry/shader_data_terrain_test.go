/******************************************************************************/
/* shader_data_terrain_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"testing"
	"unsafe"

	"kaijuengine.com/matrix"
)

func TestShaderDataTerrainUsesBoundLayerParameters(t *testing.T) {
	data := Create("terrain").(*ShaderDataTerrain)
	if data.InstanceBoundDataSize() != TerrainLayerParameterCount*int(unsafe.Sizeof(matrix.Vec4{})) {
		t.Fatalf("terrain layer bound size = %d", data.InstanceBoundDataSize())
	}
	if data.BoundDataPointer() == nil || !data.UpdateBoundData() {
		t.Fatal("terrain layer parameters should be available as bound instance data")
	}
	var parameters [TerrainLayerParameterCount]matrix.Vec4
	parameters[7] = matrix.NewVec4(1, 2, 3, 4)
	data.SetLayerParameters(parameters)
	if !matrix.Vec4Approx(data.LayerParameters[7], parameters[7]) {
		t.Fatalf("terrain layer parameters were not retained: %v", data.LayerParameters[7])
	}
}

func TestShaderDataTerrainDefaultsToFourUnselectedLights(t *testing.T) {
	data := Create("terrain").(*ShaderDataTerrain)
	if data.LightIds != [4]int32{-1, -1, -1, -1} {
		t.Fatalf("terrain light IDs = %v", data.LightIds)
	}
}
