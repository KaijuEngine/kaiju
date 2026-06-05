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
