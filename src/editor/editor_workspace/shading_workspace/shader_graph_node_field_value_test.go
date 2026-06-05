package shading_workspace

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestShaderGraphDefaultVector3FieldValuePadsComponents(t *testing.T) {
	value := shaderGraphDefaultFieldValue(shaderGraphNodeFieldSpec{
		Type:          shaderGraphNodeFieldVector3,
		DefaultValues: []string{"1", "2"},
	})

	want := []string{"1", "2", "0"}
	for i := range want {
		if value.Parts[i] != want[i] {
			t.Fatalf("component %d = %q, want %q", i, value.Parts[i], want[i])
		}
	}
}

func TestShaderGraphDefaultFieldValueUsesFirstSelectOption(t *testing.T) {
	value := shaderGraphDefaultFieldValue(shaderGraphNodeFieldSpec{
		Type: shaderGraphNodeFieldSelect,
		Options: []shaderGraphNodeFieldOption{
			{Label: "Mix", Value: "mix"},
			{Label: "Add", Value: "add"},
		},
	})

	if value.Option != "mix" {
		t.Fatalf("Option = %q, want mix", value.Option)
	}
}

func TestShaderGraphDefaultColorFieldValueFallsBackToWhite(t *testing.T) {
	value := shaderGraphDefaultFieldValue(shaderGraphNodeFieldSpec{Type: shaderGraphNodeFieldColor})

	if !matrix.Vec4Approx(matrix.Vec4(value.Color), matrix.Vec4(matrix.ColorWhite())) {
		t.Fatalf("Color = %v, want white", value.Color)
	}
}
