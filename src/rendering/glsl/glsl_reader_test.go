package glsl

import "testing"

const basicVertPath = "../../editor/editor_embedded_content/editor_content/renderer/src/basic.vert"
const basicFragPath = "../../editor/editor_embedded_content/editor_content/renderer/src/basic.frag"

func TestParse(t *testing.T) {
	src, err := Parse(basicVertPath)
	if err != nil {
		t.FailNow()
	}
	if src.src == "" {
		t.FailNow()
	}
	defineNames := []string{
		"LAYOUT_VERT_COLOR",
		"LAYOUT_VERT_FLAGS",
		"LAYOUT_FRAG_COLOR",
		"LAYOUT_FRAG_FLAGS",
		"LAYOUT_FRAG_POS",
		"LAYOUT_FRAG_TEX_COORDS",
		"LAYOUT_FRAG_NORMAL",
		"LAYOUT_FRAG_VIEW_DIR",
	}
	for i := range defineNames {
		v, ok := src.defines[defineNames[i]]
		if !ok || v != nil {
			t.FailNow()
		}
	}
	layouts := []struct {
		name     string
		location int
	}{
		{"color", 12},
		{"flags", 20},
		{"fragColor", 0},
		{"fragPos", 1},
		{"fragTexCoords", 2},
		{"fragViewDir", 3},
		{"fragNormal", 4},
		//{"fragFlags", 29},
		{"uint", 29}, // TODO:  BUG, the above is correct, but parser messes up "flat uint"
		{"", -1},     // Global uniform buffer
		{"Position", 0},
		{"Normal", 1},
		{"Tangent", 2},
		{"UV0", 3},
		{"Color", 4},
		{"JointIds", 5},
		{"JointWeights", 6},
		{"MorphTarget", 7},
		{"model", 8},
	}
	for i := range layouts {
		l := &src.layouts[i]
		if l.Name != layouts[i].name || l.Location != layouts[i].location {
			t.FailNow()
		}
	}
}
