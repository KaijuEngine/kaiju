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

func TestShaderGraphNodeCatalogHasMaterialTextureHelperNodes(t *testing.T) {
	want := []string{
		"normal-map",
		"normal-strength",
		"blend-normals",
		"orm-mra-unpack",
		"height-bump",
		"parallax",
		"triplanar",
		"detail-texture",
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

func TestShaderGraphNodeCatalogHasProceduralNodes(t *testing.T) {
	want := []string{
		"noise",
		"voronoi",
		"checker",
		"gradient",
		"remap",
		"posterize",
		"posterize-color",
		"fresnel",
		"rim-light",
		"fwidth",
		"ddx",
		"ddy",
	}
	for _, id := range want {
		if _, ok := shaderGraphNodeCatalogSpec(id); !ok {
			t.Fatalf("expected catalog node %q to be registered", id)
		}
	}
}

func TestShaderGraphNodeCatalogHasVectorCompositionNodes(t *testing.T) {
	want := []string{
		"vector2",
		"vector4",
		"combine-vec2",
		"combine-vec3",
		"combine-vec4",
		"split-vec2",
		"split-vec3",
		"split-vec4",
		"swizzle-vec2",
		"swizzle-vec3",
		"swizzle-vec4",
	}
	for _, id := range want {
		if _, ok := shaderGraphNodeCatalogSpec(id); !ok {
			t.Fatalf("expected catalog node %q to be registered", id)
		}
	}
}

func TestShaderGraphNodeCatalogHasVectorArithmeticNodes(t *testing.T) {
	want := []string{
		"add-vec2",
		"subtract-vec2",
		"multiply-vec2",
		"divide-vec2",
		"add-vec3",
		"subtract-vec3",
		"multiply-vec3",
		"divide-vec3",
		"add-vec4",
		"subtract-vec4",
		"multiply-vec4",
		"divide-vec4",
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
	if len(texture.Fields) == 0 || texture.Fields[0].Type != shaderGraphNodeFieldTexture ||
		!texture.Fields[0].Preview {
		t.Fatalf("texture-2d first field = %#v, want texture field with preview", texture.Fields)
	}
	if got, want := shaderGraphNodeFieldHeight(texture.Fields[0]), shaderGraphFieldHeight; got <= want {
		t.Fatalf("texture preview field height = %v, want greater than %v", got, want)
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

func TestShaderGraphMaterialTextureHelperNodePortTypes(t *testing.T) {
	normalMap, ok := shaderGraphNodeCatalogSpec("normal-map")
	if !ok {
		t.Fatal("normal-map node missing")
	}
	if len(normalMap.Inputs) != 3 || normalMap.Inputs[0].Type != "vec3" ||
		normalMap.Inputs[1].Type != "vec2" || normalMap.Inputs[2].Type != "float" ||
		len(normalMap.Outputs) != 2 || normalMap.Outputs[0].Type != "vec3" ||
		normalMap.Outputs[1].Type != "vec3" {
		t.Fatalf("normal-map ports = %#v -> %#v, want vec3,vec2,float -> vec3,vec3",
			normalMap.Inputs, normalMap.Outputs)
	}

	packed, ok := shaderGraphNodeCatalogSpec("orm-mra-unpack")
	if !ok {
		t.Fatal("orm-mra-unpack node missing")
	}
	if len(packed.Inputs) != 1 || packed.Inputs[0].Type != "color" ||
		len(packed.Outputs) != 3 || packed.Outputs[0].Type != "float" ||
		packed.Outputs[1].Name != "Roughness" || packed.Outputs[2].Name != "Metallic" {
		t.Fatalf("orm-mra-unpack ports = %#v -> %#v, want color -> occlusion/roughness/metallic floats",
			packed.Inputs, packed.Outputs)
	}

	parallax, ok := shaderGraphNodeCatalogSpec("parallax")
	if !ok {
		t.Fatal("parallax node missing")
	}
	if len(parallax.Inputs) != 3 || parallax.Inputs[0].Type != "vec2" ||
		parallax.Inputs[1].Type != "float" || parallax.Inputs[2].Type != "float" ||
		len(parallax.Outputs) != 2 || parallax.Outputs[0].Type != "vec2" ||
		parallax.Outputs[1].Type != "vec2" {
		t.Fatalf("parallax ports = %#v -> %#v, want vec2,float,float -> vec2,vec2",
			parallax.Inputs, parallax.Outputs)
	}

	triplanar, ok := shaderGraphNodeCatalogSpec("triplanar")
	if !ok {
		t.Fatal("triplanar node missing")
	}
	if len(triplanar.Inputs) != 5 || triplanar.Inputs[0].Type != "texture2D" ||
		triplanar.Inputs[1].Type != "vec3" || triplanar.Inputs[2].Type != "vec3" ||
		len(triplanar.Outputs) != 6 || triplanar.Outputs[0].Type != "color" ||
		triplanar.Outputs[1].Type != "vec3" || triplanar.Outputs[5].Type != "float" {
		t.Fatalf("triplanar ports = %#v -> %#v, want texture2D,vec3,vec3,float,float -> color/rgb/channels",
			triplanar.Inputs, triplanar.Outputs)
	}
}

func TestShaderGraphContextNodePortTypes(t *testing.T) {
	timeNode, ok := shaderGraphNodeCatalogSpec("time")
	if !ok {
		t.Fatal("time node missing")
	}
	if len(timeNode.Outputs) != 1 || timeNode.Outputs[0].Name != "Time" || timeNode.Outputs[0].Type != "float" {
		t.Fatalf("time outputs = %#v, want single Time float", timeNode.Outputs)
	}

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

func TestShaderGraphProceduralNodePortTypes(t *testing.T) {
	noise, ok := shaderGraphNodeCatalogSpec("noise")
	if !ok {
		t.Fatal("noise node missing")
	}
	if len(noise.Inputs) != 4 || noise.Inputs[0].Type != "vec2" ||
		len(noise.Outputs) != 2 || noise.Outputs[0].Type != "float" ||
		noise.Outputs[1].Type != "color" {
		t.Fatalf("noise ports = %#v -> %#v, want vec2,float,float,float -> float,color",
			noise.Inputs, noise.Outputs)
	}

	voronoi, ok := shaderGraphNodeCatalogSpec("voronoi")
	if !ok {
		t.Fatal("voronoi node missing")
	}
	if len(voronoi.Inputs) != 3 || voronoi.Inputs[0].Type != "vec2" ||
		len(voronoi.Outputs) != 4 || voronoi.Outputs[0].Name != "Distance" ||
		voronoi.Outputs[3].Type != "color" {
		t.Fatalf("voronoi ports = %#v -> %#v, want vec2,float,float -> distance/cell/edge/color",
			voronoi.Inputs, voronoi.Outputs)
	}

	checker, ok := shaderGraphNodeCatalogSpec("checker")
	if !ok {
		t.Fatal("checker node missing")
	}
	if len(checker.Inputs) != 2 || checker.Inputs[0].Type != "vec2" ||
		len(checker.Outputs) != 2 || checker.Outputs[0].Type != "color" ||
		checker.Outputs[1].Type != "float" {
		t.Fatalf("checker ports = %#v -> %#v, want vec2,float -> color,float",
			checker.Inputs, checker.Outputs)
	}

	remap, ok := shaderGraphNodeCatalogSpec("remap")
	if !ok {
		t.Fatal("remap node missing")
	}
	if len(remap.Inputs) != 5 || len(remap.Outputs) != 1 || remap.Outputs[0].Type != "float" {
		t.Fatalf("remap ports = %#v -> %#v, want five float inputs -> float", remap.Inputs, remap.Outputs)
	}

	rim, ok := shaderGraphNodeCatalogSpec("rim-light")
	if !ok {
		t.Fatal("rim-light node missing")
	}
	if len(rim.Inputs) != 5 || rim.Inputs[0].Type != "vec3" ||
		len(rim.Outputs) != 2 || rim.Outputs[0].Type != "float" ||
		rim.Outputs[1].Type != "color" {
		t.Fatalf("rim-light ports = %#v -> %#v, want vec3,vec3,float,float,color -> float,color",
			rim.Inputs, rim.Outputs)
	}
}

func TestShaderGraphVectorCompositionNodePortTypes(t *testing.T) {
	vector2, ok := shaderGraphNodeCatalogSpec("vector2")
	if !ok {
		t.Fatal("vector2 node missing")
	}
	if len(vector2.Outputs) != 1 || vector2.Outputs[0].Type != "vec2" {
		t.Fatalf("vector2 outputs = %#v, want vec2", vector2.Outputs)
	}

	vector4, ok := shaderGraphNodeCatalogSpec("vector4")
	if !ok {
		t.Fatal("vector4 node missing")
	}
	if len(vector4.Outputs) != 2 || vector4.Outputs[0].Type != "vec4" || vector4.Outputs[1].Type != "color" {
		t.Fatalf("vector4 outputs = %#v, want vec4 and color", vector4.Outputs)
	}

	combine2, ok := shaderGraphNodeCatalogSpec("combine-vec2")
	if !ok {
		t.Fatal("combine-vec2 node missing")
	}
	if len(combine2.Inputs) != 2 || combine2.Inputs[0].Type != "float" ||
		len(combine2.Outputs) != 1 || combine2.Outputs[0].Type != "vec2" {
		t.Fatalf("combine-vec2 ports = %#v -> %#v, want float,float -> vec2", combine2.Inputs, combine2.Outputs)
	}

	combine4, ok := shaderGraphNodeCatalogSpec("combine-vec4")
	if !ok {
		t.Fatal("combine-vec4 node missing")
	}
	if len(combine4.Inputs) != 4 || len(combine4.Outputs) != 2 ||
		combine4.Outputs[0].Type != "vec4" || combine4.Outputs[1].Type != "color" {
		t.Fatalf("combine-vec4 ports = %#v -> %#v, want four floats -> vec4 and color", combine4.Inputs, combine4.Outputs)
	}

	split3, ok := shaderGraphNodeCatalogSpec("split-vec3")
	if !ok {
		t.Fatal("split-vec3 node missing")
	}
	if len(split3.Inputs) != 1 || split3.Inputs[0].Type != "vec3" ||
		len(split3.Outputs) != 3 || split3.Outputs[2].Type != "float" {
		t.Fatalf("split-vec3 ports = %#v -> %#v, want vec3 -> three floats", split3.Inputs, split3.Outputs)
	}

	split4, ok := shaderGraphNodeCatalogSpec("split-vec4")
	if !ok {
		t.Fatal("split-vec4 node missing")
	}
	if len(split4.Inputs) != 1 || split4.Inputs[0].Type != "vec4" ||
		len(split4.Outputs) != 4 || split4.Outputs[3].Name != "W" || split4.Outputs[3].Type != "float" {
		t.Fatalf("split-vec4 ports = %#v -> %#v, want vec4 -> four floats", split4.Inputs, split4.Outputs)
	}

	swizzle3, ok := shaderGraphNodeCatalogSpec("swizzle-vec3")
	if !ok {
		t.Fatal("swizzle-vec3 node missing")
	}
	if len(swizzle3.Inputs) != 1 || swizzle3.Inputs[0].Type != "vec3" ||
		len(swizzle3.Outputs) != 1 || swizzle3.Outputs[0].Type != "vec3" {
		t.Fatalf("swizzle-vec3 ports = %#v -> %#v, want vec3 -> vec3", swizzle3.Inputs, swizzle3.Outputs)
	}

	swizzle4, ok := shaderGraphNodeCatalogSpec("swizzle-vec4")
	if !ok {
		t.Fatal("swizzle-vec4 node missing")
	}
	if len(swizzle4.Inputs) != 1 || swizzle4.Inputs[0].Type != "vec4" ||
		len(swizzle4.Outputs) != 2 || swizzle4.Outputs[0].Type != "vec4" || swizzle4.Outputs[1].Type != "color" {
		t.Fatalf("swizzle-vec4 ports = %#v -> %#v, want vec4 -> vec4 and color", swizzle4.Inputs, swizzle4.Outputs)
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

func TestShaderGraphVectorArithmeticNodePortTypes(t *testing.T) {
	tests := []struct {
		id         string
		vectorType string
		outputs    int
	}{
		{id: "add-vec2", vectorType: "vec2", outputs: 1},
		{id: "subtract-vec2", vectorType: "vec2", outputs: 1},
		{id: "multiply-vec2", vectorType: "vec2", outputs: 1},
		{id: "divide-vec2", vectorType: "vec2", outputs: 1},
		{id: "add-vec3", vectorType: "vec3", outputs: 1},
		{id: "subtract-vec3", vectorType: "vec3", outputs: 1},
		{id: "multiply-vec3", vectorType: "vec3", outputs: 1},
		{id: "divide-vec3", vectorType: "vec3", outputs: 1},
		{id: "add-vec4", vectorType: "vec4", outputs: 2},
		{id: "subtract-vec4", vectorType: "vec4", outputs: 2},
		{id: "multiply-vec4", vectorType: "vec4", outputs: 2},
		{id: "divide-vec4", vectorType: "vec4", outputs: 2},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			spec, ok := shaderGraphNodeCatalogSpec(tt.id)
			if !ok {
				t.Fatalf("%s node missing", tt.id)
			}
			if len(spec.Inputs) != 2 || spec.Inputs[0].Type != tt.vectorType || spec.Inputs[1].Type != tt.vectorType {
				t.Fatalf("%s inputs = %#v, want two %s inputs", tt.id, spec.Inputs, tt.vectorType)
			}
			if len(spec.Outputs) != tt.outputs || spec.Outputs[0].Type != tt.vectorType {
				t.Fatalf("%s outputs = %#v, want %d outputs starting with %s", tt.id, spec.Outputs, tt.outputs, tt.vectorType)
			}
			if tt.vectorType == "vec4" && spec.Outputs[1].Type != "color" {
				t.Fatalf("%s outputs = %#v, want second color output", tt.id, spec.Outputs)
			}
		})
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

func TestShaderGraphMaterialOutputExposesSurfaceAndDisplacementInputs(t *testing.T) {
	output, ok := shaderGraphNodeCatalogSpec("material-output")
	if !ok {
		t.Fatal("material-output node missing")
	}
	want := []shaderGraphPortSpec{
		{Name: "Surface", Type: "surface"},
		{Name: "Displacement", Type: "float"},
	}
	if len(output.Inputs) != len(want) {
		t.Fatalf("material-output inputs = %#v, want %#v", output.Inputs, want)
	}
	for i := range want {
		if output.Inputs[i] != want[i] {
			t.Fatalf("material-output input %d = %#v, want %#v", i, output.Inputs[i], want[i])
		}
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
