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
