package shader_data_registry

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestCreateTerrainShaderDataAliases(t *testing.T) {
	for _, name := range []string{"terrain", "terrain_lit", "terrain_unlit", "heightScalar"} {
		data, ok := Create(name).(*ShaderDataTerrain)
		if !ok {
			t.Fatalf("Create(%q) returned %T, want *ShaderDataTerrain", name, data)
		}
		if data.UVs != matrix.NewVec4(0, 0, 1, 1) {
			t.Fatalf("Create(%q) UVs = %v", name, data.UVs)
		}
		if data.BrushParams.X() <= 0 {
			t.Fatalf("Create(%q) should initialize brush params, got %v", name, data.BrushParams)
		}
	}
}
