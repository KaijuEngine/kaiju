package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestRenderGraphDefaultVector3FieldValuePadsComponents(t *testing.T) {
	value := renderGraphDefaultFieldValue(renderGraphNodeFieldSpec{
		Type:          renderGraphNodeFieldVector3,
		DefaultValues: []string{"1", "2"},
	})

	want := []string{"1", "2", "0"}
	for i := range want {
		if value.Parts[i] != want[i] {
			t.Fatalf("component %d = %q, want %q", i, value.Parts[i], want[i])
		}
	}
}

func TestRenderGraphDefaultVector2FieldValuePadsComponents(t *testing.T) {
	value := renderGraphDefaultFieldValue(renderGraphNodeFieldSpec{
		Type:          renderGraphNodeFieldVector2,
		DefaultValues: []string{"1"},
	})

	want := []string{"1", "0"}
	for i := range want {
		if value.Parts[i] != want[i] {
			t.Fatalf("component %d = %q, want %q", i, value.Parts[i], want[i])
		}
	}
}

func TestRenderGraphDefaultVector4FieldValuePadsComponents(t *testing.T) {
	value := renderGraphDefaultFieldValue(renderGraphNodeFieldSpec{
		Type:          renderGraphNodeFieldVector4,
		DefaultValues: []string{"1", "2"},
	})

	want := []string{"1", "2", "0", "0"}
	for i := range want {
		if value.Parts[i] != want[i] {
			t.Fatalf("component %d = %q, want %q", i, value.Parts[i], want[i])
		}
	}
}

func TestRenderGraphDefaultTextureFieldValueUsesDefaultAsset(t *testing.T) {
	value := renderGraphDefaultFieldValue(renderGraphNodeFieldSpec{
		Type:    renderGraphNodeFieldTexture,
		Default: "albedo.png",
	})

	if value.Text != "albedo.png" {
		t.Fatalf("Text = %q, want albedo.png", value.Text)
	}
}

func TestRenderGraphDefaultFieldValueUsesFirstSelectOption(t *testing.T) {
	value := renderGraphDefaultFieldValue(renderGraphNodeFieldSpec{
		Type: renderGraphNodeFieldSelect,
		Options: []renderGraphNodeFieldOption{
			{Label: "Mix", Value: "mix"},
			{Label: "Add", Value: "add"},
		},
	})

	if value.Option != "mix" {
		t.Fatalf("Option = %q, want mix", value.Option)
	}
}

func TestRenderGraphDefaultColorFieldValueFallsBackToWhite(t *testing.T) {
	value := renderGraphDefaultFieldValue(renderGraphNodeFieldSpec{Type: renderGraphNodeFieldColor})

	if !matrix.Vec4Approx(matrix.Vec4(value.Color), matrix.Vec4(matrix.ColorWhite())) {
		t.Fatalf("Color = %v, want white", value.Color)
	}
}
