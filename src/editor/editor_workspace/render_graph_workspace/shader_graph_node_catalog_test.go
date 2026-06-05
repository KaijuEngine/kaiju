package render_graph_workspace

import "testing"

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

func TestShaderGraphNodeCatalogIDsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, entry := range shaderGraphNodeCatalog() {
		if seen[entry.ID] {
			t.Fatalf("duplicate shader graph node catalog id %q", entry.ID)
		}
		seen[entry.ID] = true
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
