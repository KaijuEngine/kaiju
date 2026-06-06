package render_graph_workspace

import (
	"strings"
	"testing"

	"kaijuengine.com/matrix"
)

func TestRenderGraphDocumentJSONRoundTrip(t *testing.T) {
	clamp := true
	document := RenderGraphDocument{
		Generated: &RenderGraphGenerated{
			ShaderID:           "shader-id.shader",
			MaterialID:         "material-id.material",
			FragmentSpvID:      "fragment-id.spv",
			FragmentSourcePath: "database/src/render/shader/render_graph_test.frag",
		},
		Nodes: []RenderGraphNode{
			{
				ID:       "node-value",
				Type:     "value",
				Position: matrix.NewVec2(10, 20),
				Values: map[string]RenderGraphFieldValue{
					"value": {Text: "0.75"},
				},
			},
			{
				ID:       "node-mix",
				Type:     "mix-color",
				Position: matrix.NewVec2(140, 48),
				Values: map[string]RenderGraphFieldValue{
					"clamp": {Bool: &clamp},
					"mode":  {Option: "multiply"},
				},
			},
		},
		Comments: []RenderGraphComment{
			{
				ID:       "comment-lighting",
				Label:    "Lighting",
				Position: matrix.NewVec2(4, 8),
				Size:     matrix.NewVec2(320, 180),
			},
		},
		Connections: []RenderGraphConnection{
			{
				Output: RenderGraphPortRef{Node: "node-value", Port: 0},
				Input:  RenderGraphPortRef{Node: "node-mix", Port: 0},
			},
		},
	}

	data, err := SerializeRenderGraphDocument(document)
	if err != nil {
		t.Fatalf("SerializeRenderGraphDocument() error = %v", err)
	}
	if !strings.Contains(string(data), "\"version\": 1") {
		t.Fatalf("serialized render graph missing version: %s", string(data))
	}

	loaded, err := DeserializeRenderGraphDocument(data)
	if err != nil {
		t.Fatalf("DeserializeRenderGraphDocument() error = %v", err)
	}
	if loaded.Version != renderGraphDocumentVersion {
		t.Fatalf("Version = %d, want %d", loaded.Version, renderGraphDocumentVersion)
	}
	if got := loaded.Nodes[0].Values["value"].Text; got != "0.75" {
		t.Fatalf("loaded value text = %q, want %q", got, "0.75")
	}
	if got := loaded.Nodes[1].Values["clamp"].Bool; got == nil || !*got {
		t.Fatalf("loaded clamp bool = %v, want true", got)
	}
	if got := loaded.Connections[0].Input.Node; got != "node-mix" {
		t.Fatalf("loaded connection input node = %q, want node-mix", got)
	}
	if got := loaded.Comments[0].Label; got != "Lighting" {
		t.Fatalf("loaded comment label = %q, want Lighting", got)
	}
	if got := loaded.Comments[0].Size; !matrix.Vec2Approx(got, matrix.NewVec2(320, 180)) {
		t.Fatalf("loaded comment size = %v, want [320 180]", got)
	}
	if loaded.Generated == nil {
		t.Fatal("loaded generated output metadata is nil")
	}
	if got := loaded.Generated.MaterialID; got != "material-id.material" {
		t.Fatalf("loaded generated material id = %q, want material-id.material", got)
	}
}

func TestRenderGraphDocumentRejectsTinyComment(t *testing.T) {
	document := RenderGraphDocument{
		Version: renderGraphDocumentVersion,
		Nodes:   []RenderGraphNode{},
		Comments: []RenderGraphComment{
			{ID: "comment-a", Size: matrix.NewVec2(10, 10)},
		},
	}

	if _, err := SerializeRenderGraphDocument(document); err == nil {
		t.Fatal("SerializeRenderGraphDocument() should reject comment sizes below the minimum")
	}
}

func TestRenderGraphDocumentRejectsUnknownNodeType(t *testing.T) {
	document := RenderGraphDocument{
		Version: renderGraphDocumentVersion,
		Nodes: []RenderGraphNode{
			{ID: "node-a", Type: "missing-node-type"},
		},
	}

	if _, err := SerializeRenderGraphDocument(document); err == nil {
		t.Fatal("SerializeRenderGraphDocument() should reject unknown node types")
	}
}
