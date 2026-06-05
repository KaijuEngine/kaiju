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

func TestShaderGraphNodeCatalogHasContextNodes(t *testing.T) {
	want := []string{
		"time",
		"world-position",
		"normal-vector",
		"tangent-vector",
		"bitangent-vector",
		"view-direction",
		"camera-position",
		"screen-position",
		"vertex-color",
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

func TestShaderGraphContextNodePortTypes(t *testing.T) {
	world, ok := shaderGraphNodeCatalogSpec("world-position")
	if !ok {
		t.Fatal("world-position node missing")
	}
	if len(world.Outputs) != 4 || world.Outputs[0].Type != "vec3" || world.Outputs[3].Type != "float" {
		t.Fatalf("world-position outputs = %#v, want vec3 plus float components", world.Outputs)
	}

	screen, ok := shaderGraphNodeCatalogSpec("screen-position")
	if !ok {
		t.Fatal("screen-position node missing")
	}
	if len(screen.Outputs) != 5 || screen.Outputs[0].Type != "vec2" || screen.Outputs[1].Type != "vec2" || screen.Outputs[4].Type != "float" {
		t.Fatalf("screen-position outputs = %#v, want vec2, vec2, float components", screen.Outputs)
	}

	vertexColor, ok := shaderGraphNodeCatalogSpec("vertex-color")
	if !ok {
		t.Fatal("vertex-color node missing")
	}
	if len(vertexColor.Outputs) != 3 || vertexColor.Outputs[0].Type != "color" ||
		vertexColor.Outputs[1].Type != "vec3" || vertexColor.Outputs[2].Type != "float" {
		t.Fatalf("vertex-color outputs = %#v, want color, vec3, float", vertexColor.Outputs)
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

func TestShaderGraphPrincipledBSDFExposesExpandedPBRInputs(t *testing.T) {
	bsdf, ok := shaderGraphNodeCatalogSpec("principled-bsdf")
	if !ok {
		t.Fatal("principled-bsdf node missing")
	}
	want := []shaderGraphPortSpec{
		{Name: "Base Color", Type: "color"},
		{Name: "Roughness", Type: "float"},
		{Name: "Normal", Type: "vec3"},
		{Name: "Metallic", Type: "float"},
		{Name: "Occlusion", Type: "float"},
		{Name: "Emission Color", Type: "color"},
		{Name: "Emission Strength", Type: "float"},
		{Name: "Alpha", Type: "float"},
		{Name: "Specular", Type: "float"},
	}
	if len(bsdf.Inputs) != len(want) {
		t.Fatalf("principled-bsdf inputs = %#v, want %#v", bsdf.Inputs, want)
	}
	for i := range want {
		if bsdf.Inputs[i] != want[i] {
			t.Fatalf("principled-bsdf input %d = %#v, want %#v", i, bsdf.Inputs[i], want[i])
		}
	}
}

func TestShaderGraphMaterialOutputOnlyExposesCompiledSurfaceInput(t *testing.T) {
	output, ok := shaderGraphNodeCatalogSpec("material-output")
	if !ok {
		t.Fatal("material-output node missing")
	}
	if len(output.Inputs) != 1 {
		t.Fatalf("material-output inputs = %#v, want only surface", output.Inputs)
	}
	if output.Inputs[0].Name != "Surface" || output.Inputs[0].Type != "surface" {
		t.Fatalf("material-output input = %#v, want Surface surface", output.Inputs[0])
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
