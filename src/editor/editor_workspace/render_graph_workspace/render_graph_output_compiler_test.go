package render_graph_workspace

import (
	"strings"
	"testing"

	"kaijuengine.com/matrix"
)

func TestRenderGraphCompilerDefaultGraphGeneratesPBRFragment(t *testing.T) {
	out, err := compileRenderGraphDocumentOutput(defaultRenderGraphCompilerDocument())
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	source := out.FragmentSource
	for _, want := range []string{
		"#define HAS_GBUFFER",
		"vec4 graphBaseColor = fragColor;",
		"float roughness = clamp(mrSample.g * max(fragRoughness, MIN_ROUGHNESS), MIN_ROUGHNESS, 1.0);",
		"vec3 N = pbrNormal(geometricNormal);",
		"processFinalColor(vec4(color, alpha));",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerMapsPrincipledInputs(t *testing.T) {
	color := matrix.NewColor(1, 0.25, 0.5, 1)
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "color",
			Type: "color",
			Values: map[string]RenderGraphFieldValue{
				"color": {Color: &color},
			},
		},
		RenderGraphNode{
			ID:   "roughness",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.72"},
			},
		},
		RenderGraphNode{
			ID:   "normal",
			Type: "vector",
			Values: map[string]RenderGraphFieldValue{
				"vector": {Parts: []string{"0", "1", "0"}},
			},
		},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "color", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 0},
		},
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "roughness", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 1},
		},
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "normal", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 2},
		},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	source := out.FragmentSource
	for _, want := range []string{
		"vec4 graphBaseColor = (vec4(1.0, 0.25, 0.5, 1.0) * fragColor);",
		"float roughness = clamp(0.72, MIN_ROUGHNESS, 1.0);",
		"vec3 N = safeNormalize(vec3(0.0, 1.0, 0.0), geometricNormal);",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerMapsExpandedPrincipledInputs(t *testing.T) {
	emission := matrix.NewColor(0.25, 0.5, 0.75, 1)
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "metallic",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.66"},
			},
		},
		RenderGraphNode{
			ID:   "occlusion",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.42"},
			},
		},
		RenderGraphNode{
			ID:   "emission-color",
			Type: "color",
			Values: map[string]RenderGraphFieldValue{
				"color": {Color: &emission},
			},
		},
		RenderGraphNode{
			ID:   "emission-strength",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "3.5"},
			},
		},
		RenderGraphNode{
			ID:   "alpha",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.7"},
			},
		},
		RenderGraphNode{
			ID:   "specular",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.25"},
			},
		},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "metallic", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 3},
		},
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "occlusion", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 4},
		},
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "emission-color", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 5},
		},
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "emission-strength", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 6},
		},
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "alpha", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 7},
		},
		RenderGraphConnection{
			Output: RenderGraphPortRef{Node: "specular", Port: 0},
			Input:  RenderGraphPortRef{Node: "bsdf", Port: 8},
		},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	source := out.FragmentSource
	for _, want := range []string{
		"float alpha = clamp(0.7, 0.0, 1.0);",
		"float metallic = clamp(0.66, 0.0, 1.0);",
		"float occlusion = clamp(0.42, 0.0, 1.0);",
		"vec3 emission = max((vec4(0.25, 0.5, 0.75, 1.0)).rgb, vec3(0.0)) * max(3.5, 0.0);",
		"vec3 F0 = mix(vec3(0.04 * clamp(0.25, 0.0, 1.0)), albedo, metallic);",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsShaderContextNodes(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{ID: "time", Type: "time"},
		RenderGraphNode{ID: "world", Type: "world-position"},
		RenderGraphNode{ID: "normal-context", Type: "normal-vector"},
		RenderGraphNode{ID: "tangent", Type: "tangent-vector"},
		RenderGraphNode{ID: "bitangent", Type: "bitangent-vector"},
		RenderGraphNode{ID: "view", Type: "view-direction"},
		RenderGraphNode{ID: "camera", Type: "camera-position"},
		RenderGraphNode{ID: "screen", Type: "screen-position"},
		RenderGraphNode{ID: "vertex-color", Type: "vertex-color"},
		RenderGraphNode{ID: "metallic", Type: "dot-product"},
		RenderGraphNode{ID: "roughness", Type: "dot-product"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "tangent", Port: 0}, Input: RenderGraphPortRef{Node: "metallic", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "bitangent", Port: 0}, Input: RenderGraphPortRef{Node: "metallic", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "metallic", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 3}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "view", Port: 0}, Input: RenderGraphPortRef{Node: "roughness", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "normal-context", Port: 0}, Input: RenderGraphPortRef{Node: "roughness", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "roughness", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "normal-context", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 2}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "world", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 4}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "vertex-color", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 5}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "time", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 6}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "screen", Port: 2}, Input: RenderGraphPortRef{Node: "bsdf", Port: 7}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "camera", Port: 3}, Input: RenderGraphPortRef{Node: "bsdf", Port: 8}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	source := out.FragmentSource
	for _, want := range []string{
		"float metallic = clamp(dot(safeNormalize(cotangentFrame(",
		"fragPos, fragTexCoords)[1], vec3(0.0, 0.0, 1.0))), 0.0, 1.0);",
		"float roughness = clamp(dot(safeNormalize(cameraPosition.xyz - fragPos, safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0))), safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0))), MIN_ROUGHNESS, 1.0);",
		"vec3 N = safeNormalize(safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0)), geometricNormal);",
		"float occlusion = clamp((fragPos).x, 0.0, 1.0);",
		"vec3 emission = max((fragColor).rgb, vec3(0.0)) * max(time, 0.0);",
		"float alpha = clamp((gl_FragCoord.xy / max(screenSize, vec2(1.0))).x, 0.0, 1.0);",
		"vec3 F0 = mix(vec3(0.04 * clamp((cameraPosition.xyz).z, 0.0, 1.0)), albedo, metallic);",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsVectorCompositionNodes(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "x",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.25"},
			},
		},
		RenderGraphNode{
			ID:   "y",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.5"},
			},
		},
		RenderGraphNode{
			ID:   "z",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.75"},
			},
		},
		RenderGraphNode{
			ID:   "w",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "1.0"},
			},
		},
		RenderGraphNode{ID: "combine2", Type: "combine-vec2"},
		RenderGraphNode{ID: "split2", Type: "split-vec2"},
		RenderGraphNode{ID: "combine3", Type: "combine-vec3"},
		RenderGraphNode{ID: "split3", Type: "split-vec3"},
		RenderGraphNode{ID: "combine4", Type: "combine-vec4"},
		RenderGraphNode{ID: "split4", Type: "split-vec4"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "x", Port: 0}, Input: RenderGraphPortRef{Node: "combine2", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "y", Port: 0}, Input: RenderGraphPortRef{Node: "combine2", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "combine2", Port: 0}, Input: RenderGraphPortRef{Node: "split2", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split2", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 3}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split2", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},

		RenderGraphConnection{Output: RenderGraphPortRef{Node: "x", Port: 0}, Input: RenderGraphPortRef{Node: "combine3", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "y", Port: 0}, Input: RenderGraphPortRef{Node: "combine3", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "z", Port: 0}, Input: RenderGraphPortRef{Node: "combine3", Port: 2}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "combine3", Port: 0}, Input: RenderGraphPortRef{Node: "split3", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "combine3", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 2}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split3", Port: 2}, Input: RenderGraphPortRef{Node: "bsdf", Port: 4}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split3", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 8}},

		RenderGraphConnection{Output: RenderGraphPortRef{Node: "x", Port: 0}, Input: RenderGraphPortRef{Node: "combine4", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "y", Port: 0}, Input: RenderGraphPortRef{Node: "combine4", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "z", Port: 0}, Input: RenderGraphPortRef{Node: "combine4", Port: 2}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "w", Port: 0}, Input: RenderGraphPortRef{Node: "combine4", Port: 3}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "combine4", Port: 0}, Input: RenderGraphPortRef{Node: "split4", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "combine4", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "combine4", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 5}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split4", Port: 3}, Input: RenderGraphPortRef{Node: "bsdf", Port: 7}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	source := out.FragmentSource
	for _, want := range []string{
		"vec4 graphBaseColor = (vec4(0.25, 0.5, 0.75, 1.0) * fragColor);",
		"float alpha = clamp((vec4(0.25, 0.5, 0.75, 1.0)).w, 0.0, 1.0);",
		"float metallic = clamp((vec2(0.25, 0.5)).x, 0.0, 1.0);",
		"float roughness = clamp((vec2(0.25, 0.5)).y, MIN_ROUGHNESS, 1.0);",
		"float occlusion = clamp((vec3(0.25, 0.5, 0.75)).z, 0.0, 1.0);",
		"vec3 emission = max((vec4(0.25, 0.5, 0.75, 1.0)).rgb, vec3(0.0)) * max(fragEmissive, 0.0);",
		"vec3 N = safeNormalize(vec3(0.25, 0.5, 0.75), geometricNormal);",
		"vec3 F0 = mix(vec3(0.04 * clamp((vec3(0.25, 0.5, 0.75)).x, 0.0, 1.0)), albedo, metallic);",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsVectorConstantNodes(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "v2",
			Type: "vector2",
			Values: map[string]RenderGraphFieldValue{
				"vector": {Parts: []string{"0.33", "0.77"}},
			},
		},
		RenderGraphNode{ID: "split2", Type: "split-vec2"},
		RenderGraphNode{
			ID:   "v4",
			Type: "vector4",
			Values: map[string]RenderGraphFieldValue{
				"vector": {Parts: []string{"0.2", "0.4", "0.6", "1"}},
			},
		},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "v2", Port: 0}, Input: RenderGraphPortRef{Node: "split2", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split2", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "v4", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	for _, want := range []string{
		"vec4 graphBaseColor = (vec4(0.2, 0.4, 0.6, 1.0) * fragColor);",
		"float roughness = clamp((vec2(0.33, 0.77)).y, MIN_ROUGHNESS, 1.0);",
	} {
		if !strings.Contains(out.FragmentSource, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsSwizzleVectorNodes(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "v2",
			Type: "vector2",
			Values: map[string]RenderGraphFieldValue{
				"vector": {Parts: []string{"0.33", "0.77"}},
			},
		},
		RenderGraphNode{
			ID:   "swizzle2",
			Type: "swizzle-vec2",
			Values: map[string]RenderGraphFieldValue{
				"x": {Option: "y"},
				"y": {Option: "1"},
			},
		},
		RenderGraphNode{ID: "split2", Type: "split-vec2"},
		RenderGraphNode{
			ID:   "v3",
			Type: "vector",
			Values: map[string]RenderGraphFieldValue{
				"vector": {Parts: []string{"0.1", "0.2", "0.3"}},
			},
		},
		RenderGraphNode{
			ID:   "swizzle3",
			Type: "swizzle-vec3",
			Values: map[string]RenderGraphFieldValue{
				"x": {Option: "z"},
				"y": {Option: "x"},
				"z": {Option: "0"},
			},
		},
		RenderGraphNode{
			ID:   "v4",
			Type: "vector4",
			Values: map[string]RenderGraphFieldValue{
				"vector": {Parts: []string{"0.2", "0.4", "0.6", "1"}},
			},
		},
		RenderGraphNode{
			ID:   "swizzle4",
			Type: "swizzle-vec4",
			Values: map[string]RenderGraphFieldValue{
				"x": {Option: "w"},
				"y": {Option: "z"},
				"z": {Option: "y"},
				"w": {Option: "x"},
			},
		},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "v2", Port: 0}, Input: RenderGraphPortRef{Node: "swizzle2", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "swizzle2", Port: 0}, Input: RenderGraphPortRef{Node: "split2", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split2", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "v3", Port: 0}, Input: RenderGraphPortRef{Node: "swizzle3", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "swizzle3", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 2}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "v4", Port: 0}, Input: RenderGraphPortRef{Node: "swizzle4", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "swizzle4", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	for _, want := range []string{
		"vec4 graphBaseColor = (vec4((vec4(0.2, 0.4, 0.6, 1.0)).w, (vec4(0.2, 0.4, 0.6, 1.0)).z, (vec4(0.2, 0.4, 0.6, 1.0)).y, (vec4(0.2, 0.4, 0.6, 1.0)).x) * fragColor);",
		"float roughness = clamp((vec2((vec2(0.33, 0.77)).y, 1.0)).x, MIN_ROUGHNESS, 1.0);",
		"vec3 N = safeNormalize(vec3((vec3(0.1, 0.2, 0.3)).z, (vec3(0.1, 0.2, 0.3)).x, 0.0), geometricNormal);",
	} {
		if !strings.Contains(out.FragmentSource, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsVectorArithmeticNodes(t *testing.T) {
	tests := []struct {
		nodeType   string
		vectorType string
		want       string
	}{
		{
			nodeType:   "add-vec2",
			vectorType: "vec2",
			want:       "float roughness = clamp(((vec2(0.25, 0.5) + vec2(0.5, 0.25))).x, MIN_ROUGHNESS, 1.0);",
		},
		{
			nodeType:   "subtract-vec2",
			vectorType: "vec2",
			want:       "float roughness = clamp(((vec2(0.25, 0.5) - vec2(0.5, 0.25))).x, MIN_ROUGHNESS, 1.0);",
		},
		{
			nodeType:   "multiply-vec2",
			vectorType: "vec2",
			want:       "float roughness = clamp(((vec2(0.25, 0.5) * vec2(0.5, 0.25))).x, MIN_ROUGHNESS, 1.0);",
		},
		{
			nodeType:   "divide-vec2",
			vectorType: "vec2",
			want:       "float roughness = clamp(((vec2(0.25, 0.5) / vec2(0.5, 0.25))).x, MIN_ROUGHNESS, 1.0);",
		},
		{
			nodeType:   "add-vec3",
			vectorType: "vec3",
			want:       "vec3 N = safeNormalize((vec3(0.25, 0.5, 0.75) + vec3(0.5, 0.25, 0.125)), geometricNormal);",
		},
		{
			nodeType:   "subtract-vec3",
			vectorType: "vec3",
			want:       "vec3 N = safeNormalize((vec3(0.25, 0.5, 0.75) - vec3(0.5, 0.25, 0.125)), geometricNormal);",
		},
		{
			nodeType:   "multiply-vec3",
			vectorType: "vec3",
			want:       "vec3 N = safeNormalize((vec3(0.25, 0.5, 0.75) * vec3(0.5, 0.25, 0.125)), geometricNormal);",
		},
		{
			nodeType:   "divide-vec3",
			vectorType: "vec3",
			want:       "vec3 N = safeNormalize((vec3(0.25, 0.5, 0.75) / vec3(0.5, 0.25, 0.125)), geometricNormal);",
		},
		{
			nodeType:   "add-vec4",
			vectorType: "vec4",
			want:       "vec4 graphBaseColor = ((vec4(0.25, 0.5, 0.75, 1.0) + vec4(0.5, 0.25, 0.125, 0.75)) * fragColor);",
		},
		{
			nodeType:   "subtract-vec4",
			vectorType: "vec4",
			want:       "vec4 graphBaseColor = ((vec4(0.25, 0.5, 0.75, 1.0) - vec4(0.5, 0.25, 0.125, 0.75)) * fragColor);",
		},
		{
			nodeType:   "multiply-vec4",
			vectorType: "vec4",
			want:       "vec4 graphBaseColor = ((vec4(0.25, 0.5, 0.75, 1.0) * vec4(0.5, 0.25, 0.125, 0.75)) * fragColor);",
		},
		{
			nodeType:   "divide-vec4",
			vectorType: "vec4",
			want:       "vec4 graphBaseColor = ((vec4(0.25, 0.5, 0.75, 1.0) / vec4(0.5, 0.25, 0.125, 0.75)) * fragColor);",
		},
	}
	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			doc := vectorArithmeticCompilerDocument(tt.nodeType, tt.vectorType)
			out, err := compileRenderGraphDocumentOutput(doc)
			if err != nil {
				t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
			}
			if !strings.Contains(out.FragmentSource, tt.want) {
				t.Fatalf("generated fragment missing %q", tt.want)
			}
		})
	}
}

func TestRenderGraphCompilerSupportsMathAndMixColorNodes(t *testing.T) {
	clamp := true
	a := matrix.NewColor(1, 0, 0, 1)
	b := matrix.NewColor(0, 0, 1, 1)
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "factor-a",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.25"},
			},
		},
		RenderGraphNode{
			ID:   "factor-b",
			Type: "value",
			Values: map[string]RenderGraphFieldValue{
				"value": {Text: "0.5"},
			},
		},
		RenderGraphNode{ID: "factor", Type: "add"},
		RenderGraphNode{
			ID:   "color-a",
			Type: "color",
			Values: map[string]RenderGraphFieldValue{
				"color": {Color: &a},
			},
		},
		RenderGraphNode{
			ID:   "color-b",
			Type: "color",
			Values: map[string]RenderGraphFieldValue{
				"color": {Color: &b},
			},
		},
		RenderGraphNode{
			ID:   "mix",
			Type: "mix-color",
			Values: map[string]RenderGraphFieldValue{
				"clamp": {Bool: &clamp},
				"mode":  {Option: "mix"},
			},
		},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "factor-a", Port: 0}, Input: RenderGraphPortRef{Node: "factor", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "factor-b", Port: 0}, Input: RenderGraphPortRef{Node: "factor", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "factor", Port: 0}, Input: RenderGraphPortRef{Node: "mix", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "color-a", Port: 0}, Input: RenderGraphPortRef{Node: "mix", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "color-b", Port: 0}, Input: RenderGraphPortRef{Node: "mix", Port: 2}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mix", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	want := "clamp(mix(vec4(1.0, 0.0, 0.0, 1.0), vec4(0.0, 0.0, 1.0, 1.0), clamp((0.25 + 0.5), 0.0, 1.0)), vec4(0.0), vec4(1.0))"
	if !strings.Contains(out.FragmentSource, want) {
		t.Fatalf("generated fragment missing mix expression %q", want)
	}
}

func TestRenderGraphCompilerSupportsTextureSamplingNodes(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "albedo",
			Type: "texture-2d",
			Values: map[string]RenderGraphFieldValue{
				"texture":     {Text: "brick-albedo.png"},
				"label":       {Text: "Brick Albedo"},
				"filter":      {Option: "Nearest"},
				"color-space": {Option: "srgb"},
			},
		},
		RenderGraphNode{
			ID:   "uv-scale",
			Type: "uv-transform",
			Values: map[string]RenderGraphFieldValue{
				"tiling": {Parts: []string{"2", "3"}},
				"offset": {Parts: []string{"0.25", "0.5"}},
			},
		},
		RenderGraphNode{ID: "sample", Type: "sample-texture-2d"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "albedo", Port: 0}, Input: RenderGraphPortRef{Node: "sample", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "uv-scale", Port: 0}, Input: RenderGraphPortRef{Node: "sample", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "sample", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	if len(out.Textures) != 5 {
		t.Fatalf("texture slot count = %d, want 5", len(out.Textures))
	}
	if got := out.SamplerLabels[4]; got != "Brick Albedo" {
		t.Fatalf("sampler label = %q, want Brick Albedo", got)
	}
	if got := out.Textures[4].Texture; got != "brick-albedo.png" {
		t.Fatalf("texture slot asset = %q, want brick-albedo.png", got)
	}
	if got := out.Textures[4].Filter; got != "Nearest" {
		t.Fatalf("texture slot filter = %q, want Nearest", got)
	}
	for _, want := range []string{
		"#define SAMPLER_COUNT   5",
		"graphSrgbToLinear(texture(textures[4], ((fragTexCoords) * vec2(2.0, 3.0) + vec2(0.25, 0.5))))",
	} {
		if !strings.Contains(out.FragmentSource, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerReusesTextureNodeSamplerSlot(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "mask",
			Type: "texture-2d",
			Values: map[string]RenderGraphFieldValue{
				"texture":     {Text: `textures\mask.png`},
				"label":       {Text: "Mask"},
				"color-space": {Option: "linear"},
			},
		},
		RenderGraphNode{ID: "sample-color", Type: "sample-texture-2d"},
		RenderGraphNode{ID: "sample-roughness", Type: "sample-texture-2d"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask", Port: 0}, Input: RenderGraphPortRef{Node: "sample-color", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask", Port: 0}, Input: RenderGraphPortRef{Node: "sample-roughness", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "sample-color", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "sample-roughness", Port: 2}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	if len(out.Textures) != 5 {
		t.Fatalf("texture slot count = %d, want 5", len(out.Textures))
	}
	if got := out.Textures[4].Texture; got != "textures/mask.png" {
		t.Fatalf("texture slot asset = %q, want textures/mask.png", got)
	}
	if strings.Contains(out.FragmentSource, "textures[5]") {
		t.Fatal("generated fragment allocated a duplicate sampler slot for the same texture node")
	}
	if strings.Contains(out.FragmentSource, "graphSrgbToLinear(texture(textures[4]") {
		t.Fatal("linear texture sample should not be wrapped in graphSrgbToLinear")
	}
}

func TestRenderGraphCompilerMakesDuplicateTextureLabelsUnique(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "mask-a",
			Type: "texture-2d",
			Values: map[string]RenderGraphFieldValue{
				"texture": {Text: "mask-a.png"},
				"label":   {Text: "Mask"},
			},
		},
		RenderGraphNode{
			ID:   "mask-b",
			Type: "texture-2d",
			Values: map[string]RenderGraphFieldValue{
				"texture": {Text: "mask-b.png"},
				"label":   {Text: "Mask"},
			},
		},
		RenderGraphNode{ID: "sample-a", Type: "sample-texture-2d"},
		RenderGraphNode{ID: "sample-b", Type: "sample-texture-2d"},
		RenderGraphNode{ID: "roughness", Type: "add"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask-a", Port: 0}, Input: RenderGraphPortRef{Node: "sample-a", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask-b", Port: 0}, Input: RenderGraphPortRef{Node: "sample-b", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "sample-a", Port: 2}, Input: RenderGraphPortRef{Node: "roughness", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "sample-b", Port: 2}, Input: RenderGraphPortRef{Node: "roughness", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "roughness", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	if len(out.SamplerLabels) != 6 {
		t.Fatalf("sampler label count = %d, want 6", len(out.SamplerLabels))
	}
	if out.SamplerLabels[4] != "Mask" || out.SamplerLabels[5] != "Mask 2" {
		t.Fatalf("custom sampler labels = %#v, want Mask and Mask 2", out.SamplerLabels[4:])
	}
	if !strings.Contains(out.FragmentSource, "#define SAMPLER_COUNT   6") {
		t.Fatal("generated fragment did not include dynamic sampler count 6")
	}
}

func TestRenderGraphCompilerSupportsTextureUtilityNodes(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "mask-color",
			Type: "color",
			Values: map[string]RenderGraphFieldValue{
				"color": {Color: ptrColor(matrix.NewColor(0.2, 0.4, 0.6, 1))},
			},
		},
		RenderGraphNode{
			ID:   "mask",
			Type: "channel-mask",
			Values: map[string]RenderGraphFieldValue{
				"channel": {Option: "luma"},
			},
		},
		RenderGraphNode{ID: "split", Type: "split-rgba"},
		RenderGraphNode{ID: "roughness", Type: "add"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask-color", Port: 0}, Input: RenderGraphPortRef{Node: "mask", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask-color", Port: 0}, Input: RenderGraphPortRef{Node: "split", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask", Port: 0}, Input: RenderGraphPortRef{Node: "roughness", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "split", Port: 2}, Input: RenderGraphPortRef{Node: "roughness", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "roughness", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	if !strings.Contains(out.FragmentSource, "dot((vec4(") ||
		!strings.Contains(out.FragmentSource, ")).rgb, vec3(0.2126, 0.7152, 0.0722))") {
		t.Fatal("generated fragment missing channel mask luma expression")
	}
	if !strings.Contains(out.FragmentSource, ")).b") {
		t.Fatal("generated fragment missing split rgba expression")
	}
}

func TestRenderGraphCompilerSupportsTexelSizeNode(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "mask",
			Type: "texture-2d",
			Values: map[string]RenderGraphFieldValue{
				"texture":     {Text: "mask.png"},
				"label":       {Text: "Mask"},
				"color-space": {Option: "linear"},
			},
		},
		RenderGraphNode{ID: "texel", Type: "texel-size"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask", Port: 0}, Input: RenderGraphPortRef{Node: "texel", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "texel", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	if !strings.Contains(out.FragmentSource, "(1.0 / vec2(textureSize(textures[4], 0))).x") {
		t.Fatal("generated fragment missing texel width expression")
	}
}

func TestRenderGraphCompilerSupportsNormalMapHelperNodes(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "rgb",
			Type: "vector",
			Values: map[string]RenderGraphFieldValue{
				"vector": {Parts: []string{"0.5", "0.5", "1"}},
			},
		},
		RenderGraphNode{
			ID:   "normal-map",
			Type: "normal-map",
			Values: map[string]RenderGraphFieldValue{
				"strength": {Text: "0.8"},
				"y":        {Option: "directx"},
			},
		},
		RenderGraphNode{
			ID:   "normal-strength",
			Type: "normal-strength",
			Values: map[string]RenderGraphFieldValue{
				"strength": {Text: "0.25"},
			},
		},
		RenderGraphNode{ID: "normal-context", Type: "normal-vector"},
		RenderGraphNode{ID: "blend", Type: "blend-normals"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "rgb", Port: 0}, Input: RenderGraphPortRef{Node: "normal-map", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "normal-map", Port: 0}, Input: RenderGraphPortRef{Node: "normal-strength", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "normal-strength", Port: 0}, Input: RenderGraphPortRef{Node: "blend", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "normal-context", Port: 0}, Input: RenderGraphPortRef{Node: "blend", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "blend", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 2}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	for _, want := range []string{
		"vec3 graphTangentNormalFromMap(vec3 sampleRGB, float strength, float flipY)",
		"graphTangentNormalFromMap(vec3(0.5, 0.5, 1.0), 0.8, -1.0)",
		"graphApplyNormalStrength(",
		"graphBlendNormals(",
		"vec3 N = safeNormalize(graphBlendNormals(",
	} {
		if !strings.Contains(out.FragmentSource, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsPackedPBRMapHelperNode(t *testing.T) {
	packed := matrix.NewColor(0.125, 0.25, 0.5, 1)
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "packed-color",
			Type: "color",
			Values: map[string]RenderGraphFieldValue{
				"color": {Color: &packed},
			},
		},
		RenderGraphNode{
			ID:   "packed",
			Type: "orm-mra-unpack",
			Values: map[string]RenderGraphFieldValue{
				"layout": {Option: "mra"},
			},
		},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "packed-color", Port: 0}, Input: RenderGraphPortRef{Node: "packed", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "packed", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "packed", Port: 2}, Input: RenderGraphPortRef{Node: "bsdf", Port: 3}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "packed", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 4}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	for _, want := range []string{
		"float metallic = clamp(clamp((vec4(0.125, 0.25, 0.5, 1.0)).r, 0.0, 1.0), 0.0, 1.0);",
		"float roughness = clamp(clamp((vec4(0.125, 0.25, 0.5, 1.0)).g, 0.0, 1.0), MIN_ROUGHNESS, 1.0);",
		"float occlusion = clamp(clamp((vec4(0.125, 0.25, 0.5, 1.0)).b, 0.0, 1.0), 0.0, 1.0);",
	} {
		if !strings.Contains(out.FragmentSource, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsHeightBumpAndParallaxHelpers(t *testing.T) {
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		renderGraphCompilerValueNode("height", "0.8"),
		RenderGraphNode{
			ID:   "bump",
			Type: "height-bump",
			Values: map[string]RenderGraphFieldValue{
				"strength": {Text: "0.02"},
			},
		},
		RenderGraphNode{
			ID:   "parallax",
			Type: "parallax",
			Values: map[string]RenderGraphFieldValue{
				"scale": {Text: "0.1"},
			},
		},
		RenderGraphNode{
			ID:   "mask",
			Type: "texture-2d",
			Values: map[string]RenderGraphFieldValue{
				"texture":     {Text: "height-mask.png"},
				"label":       {Text: "Height Mask"},
				"color-space": {Option: "linear"},
			},
		},
		RenderGraphNode{ID: "sample", Type: "sample-texture-2d"},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "height", Port: 0}, Input: RenderGraphPortRef{Node: "bump", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "bump", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 2}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "height", Port: 0}, Input: RenderGraphPortRef{Node: "parallax", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "mask", Port: 0}, Input: RenderGraphPortRef{Node: "sample", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "parallax", Port: 0}, Input: RenderGraphPortRef{Node: "sample", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "sample", Port: 2}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	for _, want := range []string{
		"vec3 graphBumpNormal(float height, float strength, vec3 geometricNormal)",
		"graphBumpNormal(0.8, 0.02, safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0)))",
		"vec2 graphParallaxUV(vec2 uv, float height, float scale, vec3 geometricNormal)",
		"texture(textures[4], graphParallaxUV(fragTexCoords, 0.8, 0.1, safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0))))",
	} {
		if !strings.Contains(out.FragmentSource, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerSupportsTriplanarAndDetailTextureHelpers(t *testing.T) {
	base := matrix.NewColor(0.25, 0.5, 0.75, 1)
	doc := defaultRenderGraphCompilerDocument()
	doc.Nodes = append(doc.Nodes,
		RenderGraphNode{
			ID:   "base",
			Type: "color",
			Values: map[string]RenderGraphFieldValue{
				"color": {Color: &base},
			},
		},
		RenderGraphNode{
			ID:   "detail-map",
			Type: "texture-2d",
			Values: map[string]RenderGraphFieldValue{
				"texture":     {Text: "detail.png"},
				"label":       {Text: "Detail"},
				"color-space": {Option: "linear"},
			},
		},
		RenderGraphNode{
			ID:   "triplanar",
			Type: "triplanar",
			Values: map[string]RenderGraphFieldValue{
				"scale": {Text: "2"},
				"blend": {Text: "5"},
			},
		},
		RenderGraphNode{
			ID:   "detail",
			Type: "detail-texture",
			Values: map[string]RenderGraphFieldValue{
				"mode":     {Option: "overlay"},
				"strength": {Text: "0.5"},
			},
		},
	)
	doc.Connections = append(doc.Connections,
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "base", Port: 0}, Input: RenderGraphPortRef{Node: "detail", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "detail-map", Port: 0}, Input: RenderGraphPortRef{Node: "triplanar", Port: 0}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "triplanar", Port: 0}, Input: RenderGraphPortRef{Node: "detail", Port: 1}},
		RenderGraphConnection{Output: RenderGraphPortRef{Node: "detail", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
	)

	out, err := compileRenderGraphDocumentOutput(doc)
	if err != nil {
		t.Fatalf("compileRenderGraphDocumentOutput() error = %v", err)
	}
	for _, want := range []string{
		"#define SAMPLER_COUNT   5",
		"vec4 graphOverlayColor(vec4 base, vec4 detail)",
		"graphTriplanarSample(textures[4], fragPos, safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0)), 2.0, 5.0)",
		"graphOverlayColor(vec4(0.25, 0.5, 0.75, 1.0), graphTriplanarSample(textures[4]",
		"vec4 graphBaseColor = (clamp(mix(vec4(0.25, 0.5, 0.75, 1.0), graphOverlayColor(",
	} {
		if !strings.Contains(out.FragmentSource, want) {
			t.Fatalf("generated fragment missing %q", want)
		}
	}
}

func TestRenderGraphCompilerValidationFailures(t *testing.T) {
	tests := []struct {
		name string
		doc  RenderGraphDocument
		want string
	}{
		{
			name: "missing output",
			doc: RenderGraphDocument{Nodes: []RenderGraphNode{
				{ID: "bsdf", Type: "principled-bsdf"},
			}},
			want: "missing a material output",
		},
		{
			name: "disconnected surface",
			doc: RenderGraphDocument{Nodes: []RenderGraphNode{
				{ID: "bsdf", Type: "principled-bsdf"},
				{ID: "output", Type: "material-output"},
			}},
			want: "surface is disconnected",
		},
		{
			name: "invalid port",
			doc: RenderGraphDocument{
				Nodes: []RenderGraphNode{
					{ID: "bsdf", Type: "principled-bsdf"},
					{ID: "output", Type: "material-output"},
				},
				Connections: []RenderGraphConnection{
					{Output: RenderGraphPortRef{Node: "bsdf", Port: 2}, Input: RenderGraphPortRef{Node: "output", Port: 0}},
				},
			},
			want: "invalid output port",
		},
		{
			name: "disconnected math input",
			doc: func() RenderGraphDocument {
				doc := defaultRenderGraphCompilerDocument()
				doc.Nodes = append(doc.Nodes, RenderGraphNode{ID: "add", Type: "add"})
				doc.Connections = append(doc.Connections, RenderGraphConnection{
					Output: RenderGraphPortRef{Node: "add", Port: 0},
					Input:  RenderGraphPortRef{Node: "bsdf", Port: 1},
				})
				return doc
			}(),
			want: "input 0 is disconnected",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compileRenderGraphDocumentOutput(tt.doc)
			if err == nil {
				t.Fatal("compileRenderGraphDocumentOutput() error = nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}

func ptrColor(color matrix.Color) *matrix.Color {
	return &color
}

func vectorArithmeticCompilerDocument(nodeType, vectorType string) RenderGraphDocument {
	doc := defaultRenderGraphCompilerDocument()
	switch vectorType {
	case "vec2":
		doc.Nodes = append(doc.Nodes,
			renderGraphCompilerValueNode("a-x", "0.25"),
			renderGraphCompilerValueNode("a-y", "0.5"),
			renderGraphCompilerValueNode("b-x", "0.5"),
			renderGraphCompilerValueNode("b-y", "0.25"),
			RenderGraphNode{ID: "a", Type: "combine-vec2"},
			RenderGraphNode{ID: "b", Type: "combine-vec2"},
			RenderGraphNode{ID: "math", Type: nodeType},
			RenderGraphNode{ID: "split", Type: "split-vec2"},
		)
		doc.Connections = append(doc.Connections,
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a-x", Port: 0}, Input: RenderGraphPortRef{Node: "a", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a-y", Port: 0}, Input: RenderGraphPortRef{Node: "a", Port: 1}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b-x", Port: 0}, Input: RenderGraphPortRef{Node: "b", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b-y", Port: 0}, Input: RenderGraphPortRef{Node: "b", Port: 1}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a", Port: 0}, Input: RenderGraphPortRef{Node: "math", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b", Port: 0}, Input: RenderGraphPortRef{Node: "math", Port: 1}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "math", Port: 0}, Input: RenderGraphPortRef{Node: "split", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "split", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 1}},
		)
	case "vec3":
		doc.Nodes = append(doc.Nodes,
			RenderGraphNode{
				ID:   "a",
				Type: "vector",
				Values: map[string]RenderGraphFieldValue{
					"vector": {Parts: []string{"0.25", "0.5", "0.75"}},
				},
			},
			RenderGraphNode{
				ID:   "b",
				Type: "vector",
				Values: map[string]RenderGraphFieldValue{
					"vector": {Parts: []string{"0.5", "0.25", "0.125"}},
				},
			},
			RenderGraphNode{ID: "math", Type: nodeType},
		)
		doc.Connections = append(doc.Connections,
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a", Port: 0}, Input: RenderGraphPortRef{Node: "math", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b", Port: 0}, Input: RenderGraphPortRef{Node: "math", Port: 1}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "math", Port: 0}, Input: RenderGraphPortRef{Node: "bsdf", Port: 2}},
		)
	case "vec4":
		doc.Nodes = append(doc.Nodes,
			renderGraphCompilerValueNode("a-x", "0.25"),
			renderGraphCompilerValueNode("a-y", "0.5"),
			renderGraphCompilerValueNode("a-z", "0.75"),
			renderGraphCompilerValueNode("a-w", "1"),
			renderGraphCompilerValueNode("b-x", "0.5"),
			renderGraphCompilerValueNode("b-y", "0.25"),
			renderGraphCompilerValueNode("b-z", "0.125"),
			renderGraphCompilerValueNode("b-w", "0.75"),
			RenderGraphNode{ID: "a", Type: "combine-vec4"},
			RenderGraphNode{ID: "b", Type: "combine-vec4"},
			RenderGraphNode{ID: "math", Type: nodeType},
		)
		doc.Connections = append(doc.Connections,
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a-x", Port: 0}, Input: RenderGraphPortRef{Node: "a", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a-y", Port: 0}, Input: RenderGraphPortRef{Node: "a", Port: 1}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a-z", Port: 0}, Input: RenderGraphPortRef{Node: "a", Port: 2}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a-w", Port: 0}, Input: RenderGraphPortRef{Node: "a", Port: 3}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b-x", Port: 0}, Input: RenderGraphPortRef{Node: "b", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b-y", Port: 0}, Input: RenderGraphPortRef{Node: "b", Port: 1}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b-z", Port: 0}, Input: RenderGraphPortRef{Node: "b", Port: 2}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b-w", Port: 0}, Input: RenderGraphPortRef{Node: "b", Port: 3}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "a", Port: 0}, Input: RenderGraphPortRef{Node: "math", Port: 0}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "b", Port: 0}, Input: RenderGraphPortRef{Node: "math", Port: 1}},
			RenderGraphConnection{Output: RenderGraphPortRef{Node: "math", Port: 1}, Input: RenderGraphPortRef{Node: "bsdf", Port: 0}},
		)
	}
	return doc
}

func renderGraphCompilerValueNode(id, value string) RenderGraphNode {
	return RenderGraphNode{
		ID:   id,
		Type: "value",
		Values: map[string]RenderGraphFieldValue{
			"value": {Text: value},
		},
	}
}

func defaultRenderGraphCompilerDocument() RenderGraphDocument {
	return RenderGraphDocument{
		Version: renderGraphDocumentVersion,
		Nodes: []RenderGraphNode{
			{ID: "bsdf", Type: "principled-bsdf"},
			{ID: "output", Type: "material-output"},
		},
		Connections: []RenderGraphConnection{
			{
				Output: RenderGraphPortRef{Node: "bsdf", Port: 0},
				Input:  RenderGraphPortRef{Node: "output", Port: 0},
			},
		},
	}
}
