package editor_embedded_content

import (
	"path/filepath"
	"slices"
	"testing"

	"kaijuengine.com/rendering"
)

func TestTerrainMaterialsExposePBRAtlasBindings(t *testing.T) {
	want := []string{
		"Weight Map 0", "Weight Map 1",
		"Layer Albedo Roughness Atlas", "Layer Normal Atlas",
	}
	root := filepath.FromSlash("editor_content/renderer")
	for _, name := range []string{"terrain.material", "terrain_unlit.material"} {
		var material rendering.MaterialData
		readJSON(t, filepath.Join(root, "materials", name), &material)
		labels := make([]string, len(material.Textures))
		for i := range material.Textures {
			labels[i] = material.Textures[i].Label
		}
		if !slices.Equal(labels, want) {
			t.Fatalf("%s texture labels = %#v, want %#v", name, labels, want)
		}
		if name == "terrain.material" && (!material.IsLit || !material.ReceivesShadows || !material.CastsShadows) {
			t.Fatalf("lit terrain material must be lit and cast/receive shadows: %+v", material)
		}
		if name == "terrain_unlit.material" && (material.IsLit || material.ReceivesShadows) {
			t.Fatalf("unlit terrain material should not enable lighting or shadows: %+v", material)
		}
	}
}

func TestTerrainShadersDeclareFourSamplers(t *testing.T) {
	root := filepath.FromSlash("editor_content/renderer")
	for _, name := range []string{"terrain.shader", "terrain_unlit.shader"} {
		var shader rendering.ShaderData
		readJSON(t, filepath.Join(root, "shaders", name), &shader)
		if len(shader.SamplerLabels) != 4 {
			t.Fatalf("%s sampler labels = %d, want 4", name, len(shader.SamplerLabels))
		}
		compiled := shader.Compile()
		if !hasDescriptorBinding(compiled, 1, "sampler2D", 4) {
			t.Fatalf("%s does not declare four terrain samplers", name)
		}
		if !hasDescriptorBinding(compiled, 4, "TerrainLayerBuffer", 1) {
			t.Fatalf("%s does not declare the terrain layer parameter buffer", name)
		}
	}
}
