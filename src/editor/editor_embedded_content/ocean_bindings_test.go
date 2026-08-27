/******************************************************************************/
/* ocean_bindings_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_embedded_content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kaijuengine.com/rendering"
)

func TestOceanMaterialAndGeneratedShaderContract(t *testing.T) {
	root := filepath.FromSlash("editor_content/renderer")
	var material rendering.MaterialData
	readJSON(t, filepath.Join(root, "materials", "ocean.material"), &material)
	if !material.IsLit || material.ReceivesShadows || material.CastsShadows {
		t.Fatalf("unexpected ocean lighting contract: %+v", material)
	}
	if len(material.Textures) != 0 {
		t.Fatalf("procedural ocean material has %d textures, want 0", len(material.Textures))
	}

	var shader rendering.ShaderData
	readJSON(t, filepath.Join(root, "shaders", "ocean.shader"), &shader)
	compiled := shader.Compile()
	if shader.DrawInstanceDataName() != "ocean" || len(shader.SamplerLabels) != 0 {
		t.Fatalf("unexpected ocean shader identity/samplers: %q, %v", shader.DrawInstanceDataName(), shader.SamplerLabels)
	}
	if hasDescriptorBinding(compiled, 2, "sampler2D", rendering.MaxLocalLights) ||
		hasDescriptorBinding(compiled, 3, "samplerCube", rendering.MaxLocalLights) {
		t.Fatal("ocean shader should not sample terrain shadows outside their battlefield projection")
	}
	if hasDescriptorBinding(compiled, 1, "sampler2D", 1) {
		t.Fatal("procedural ocean unexpectedly declares a material texture sampler")
	}

	fragment, err := os.ReadFile(filepath.Join(root, "src", "ocean.frag"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(fragment)
	for _, required := range []string{"OCEAN_WAVE_COUNT = 6", "oceanRippleNormal", "oceanLight", "applyOceanBrushOverlay", "fresnel", "time"} {
		if !strings.Contains(source, required) {
			t.Fatalf("ocean fragment shader is missing %q", required)
		}
	}
	if strings.Contains(source, "pbrAccumulateLight") {
		t.Fatal("ocean should use its restrained water-lighting model, not broad GGX highlights")
	}

	terrain, err := os.ReadFile(filepath.Join(root, "src", "terrain.frag"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(terrain), "TERRAIN_MIN_ROUGHNESS 0.50") {
		t.Fatal("terrain shader should keep natural surfaces from becoming mirror-like")
	}
}
