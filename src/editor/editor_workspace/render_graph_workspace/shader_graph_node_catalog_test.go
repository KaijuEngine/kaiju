package render_graph_workspace

import (
	"slices"
	"testing"
)

func TestShaderGraphNodeCatalogHasCommonMathNodes(t *testing.T) {
	want := []string{
		"add",
		"subtract",
		"multiply",
		"divide",
		"minimum",
		"maximum",
		"clamp",
		"step",
		"smoothstep",
		"lerp",
		"dot-product",
		"cross-product",
		"normalize",
		"length",
	}
	for _, id := range want {
		if _, ok := shaderGraphNodeCatalogSpec(id); !ok {
			t.Fatalf("expected catalog node %q to be registered", id)
		}
	}
}

func TestShaderGraphNodeCatalogHasTextureNodes(t *testing.T) {
	want := []string{
		"texture-2d",
		"sample-texture-2d",
		"uv",
		"uv-transform",
		"split-rgba",
		"channel-mask",
		"texel-size",
	}
	for _, id := range want {
		if _, ok := shaderGraphNodeCatalogSpec(id); !ok {
			t.Fatalf("expected catalog node %q to be registered", id)
		}
	}
}

func TestShaderGraphNodeCatalogIDsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, entry := range shaderGraphNodeCatalog() {
		if seen[entry.ID] {
			t.Fatalf("duplicate shader graph node catalog id %q", entry.ID)
		}
		seen[entry.ID] = true
	}
}

func TestShaderGraphTextureNodePortTypes(t *testing.T) {
	texture, ok := shaderGraphNodeCatalogSpec("texture-2d")
	if !ok {
		t.Fatal("texture-2d node missing")
	}
	if len(texture.Outputs) != 1 || texture.Outputs[0].Type != "texture2D" {
		t.Fatalf("texture-2d outputs = %#v, want texture2D", texture.Outputs)
	}

	sample, ok := shaderGraphNodeCatalogSpec("sample-texture-2d")
	if !ok {
		t.Fatal("sample-texture-2d node missing")
	}
	if len(sample.Inputs) != 2 || sample.Inputs[0].Type != "texture2D" || sample.Inputs[1].Type != "vec2" {
		t.Fatalf("sample-texture-2d inputs = %#v, want texture2D, vec2", sample.Inputs)
	}
	if len(sample.Outputs) != 6 || sample.Outputs[0].Type != "color" || sample.Outputs[1].Type != "vec3" || sample.Outputs[5].Type != "float" {
		t.Fatalf("sample-texture-2d outputs = %#v, want color, vec3, float channels", sample.Outputs)
	}
}

func TestShaderGraphCommonMathNodePortTypes(t *testing.T) {
	minimum, ok := shaderGraphNodeCatalogSpec("minimum")
	if !ok {
		t.Fatal("minimum node missing")
	}
	if len(minimum.Inputs) != 2 || len(minimum.Outputs) != 1 {
		t.Fatalf("minimum ports = %d inputs/%d outputs, want 2/1", len(minimum.Inputs), len(minimum.Outputs))
	}
	for _, port := range append(minimum.Inputs, minimum.Outputs...) {
		if port.Type != "float" {
			t.Fatalf("minimum port %q type = %q, want float", port.Name, port.Type)
		}
	}

	dot, ok := shaderGraphNodeCatalogSpec("dot-product")
	if !ok {
		t.Fatal("dot-product node missing")
	}
	if dot.Inputs[0].Type != "vec3" || dot.Inputs[1].Type != "vec3" || dot.Outputs[0].Type != "float" {
		t.Fatalf("dot-product ports = %#v -> %#v, want vec3,vec3 -> float", dot.Inputs, dot.Outputs)
	}
}

func TestShaderGraphNodeCatalogCompatibleIDsForOutputVec2(t *testing.T) {
	ids := shaderGraphNodeCatalogCompatibleIDs(true, " Vec2 ")

	if !slices.Contains(ids, "sample-texture-2d") {
		t.Fatal("output vec2 should offer nodes with vec2 inputs")
	}
	if !slices.Contains(ids, "uv-transform") {
		t.Fatal("output vec2 should offer uv-transform")
	}
	if slices.Contains(ids, "uv") {
		t.Fatal("output vec2 should not offer nodes with only vec2 outputs")
	}
	if slices.Contains(ids, "texel-size") {
		t.Fatal("output vec2 should not offer nodes whose vec2 port is output-only")
	}
}

func TestShaderGraphNodeCatalogCompatibleIDsForInputVec3(t *testing.T) {
	ids := shaderGraphNodeCatalogCompatibleIDs(false, " VeC3 ")

	if !slices.Contains(ids, "vector") {
		t.Fatal("input vec3 should offer nodes with vec3 outputs")
	}
	if !slices.Contains(ids, "sample-texture-2d") {
		t.Fatal("input vec3 should offer sample-texture-2d RGB output")
	}
	if slices.Contains(ids, "dot-product") {
		t.Fatal("input vec3 should not offer nodes with only vec3 inputs")
	}
	if slices.Contains(ids, "material-output") {
		t.Fatal("input vec3 should not offer material-output")
	}
}

func TestShaderGraphNodeSpecCompatiblePortIndexUsesSpecOrder(t *testing.T) {
	sample, ok := shaderGraphNodeCatalogSpec("sample-texture-2d")
	if !ok {
		t.Fatal("sample-texture-2d node missing")
	}
	if index, ok := shaderGraphNodeSpecCompatiblePortIndex(sample, true, "vec2"); !ok || index != 1 {
		t.Fatalf("output vec2 compatible input index = %d, %v; want 1, true", index, ok)
	}
	if index, ok := shaderGraphNodeSpecCompatiblePortIndex(sample, false, "vec3"); !ok || index != 1 {
		t.Fatalf("input vec3 compatible output index = %d, %v; want 1, true", index, ok)
	}
}
