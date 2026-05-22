/******************************************************************************/
/* glsl_reader_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package glsl

import (
	"testing"
)

const fragPath = "../../editor/editor_embedded_content/editor_content/renderer/src/pbr.frag"

func TestParse(t *testing.T) {
	src, err := Parse(fragPath, "")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if src.src == "" {
		t.Fatal("src.src is empty")
	}

	// Verify bare defines (no value) -- flags used for #ifdef guards
	bareDefines := []string{
		"FRAGMENT_SHADER",
		"HAS_GBUFFER",
		"SHADOW_SAMPLERS",
	}
	for i := range bareDefines {
		v, ok := src.defines[bareDefines[i]]
		if !ok {
			t.Fatalf("missing bare define: %s", bareDefines[i])
		}
		if v != nil {
			t.Fatalf("define %s expected nil, got %v", bareDefines[i], v)
		}
	}

	// Verify defines with numeric values
	numericDefines := map[string]any{
		"SAMPLER_COUNT":                 float64(4),
		"LAYOUT_FRAG_COLOR":             float64(0),
		"LAYOUT_FRAG_FLAGS":             float64(1),
		"LAYOUT_FRAG_POS":               float64(2),
		"LAYOUT_FRAG_TEX_COORDS":        float64(3),
		"LAYOUT_FRAG_NORMAL":            float64(4),
		"LAYOUT_FRAG_METALLIC":          float64(5),
		"LAYOUT_FRAG_ROUGHNESS":         float64(6),
		"LAYOUT_FRAG_EMISSIVE":          float64(7),
		"LAYOUT_ALL_LIGHT_REQUIREMENTS": float64(8),
		"LAYOUT_FRAG_TANGENT_FRAG_POS":  float64(9),
		"LAYOUT_FRAG_LIGHT_T_POS":       float64(10),
		"LAYOUT_FRAG_LIGHT_T_DIR":       float64(14),
		"LAYOUT_FRAG_POS_LIGHT_SPACE":   float64(18),
		"LAYOUT_FRAG_LIGHT_COUNT":       float64(22),
		"LAYOUT_FRAG_LIGHT_INDEXES":     float64(23),
		"LOCATION_HEAD":                 float64(8),
		"LOCATION_START":                float64(12),
		"CUBEMAP_SIDES":                 float64(6),
		"NR_LIGHTS":                     float64(4),
		"MAX_JOINTS":                    float64(50),
		"MAX_LIGHTS":                    float64(20),
	}
	for name, expected := range numericDefines {
		v, ok := src.defines[name]
		if !ok {
			t.Fatalf("missing numeric define: %s", name)
		}
		if f, ok := v.(float64); ok {
			if f != expected.(float64) {
				t.Fatalf("define %s expected %v, got %v", name, expected, f)
			}
		} else {
			t.Fatalf("define %s expected float64(%v), got %T(%v)", name, expected, v, v)
		}
	}

	// Verify PI parsed as a float (dot support in regex)
	piVal, ok := src.defines["PI"]
	if !ok {
		t.Fatal("missing define: PI")
	}
	pi, ok := piVal.(float64)
	if !ok {
		t.Fatalf("PI expected float64, got %T", piVal)
	}
	if pi < 3.14159 || pi > 3.1416 {
		t.Fatalf("PI expected ~3.14159265359, got %v", pi)
	}

	// Verify FRAG_INOUT resolves to "in" (FRAGMENT_SHADER is defined)
	fragInout, ok := src.defines["FRAG_INOUT"]
	if !ok {
		t.Fatal("missing define: FRAG_INOUT")
	}
	if fragStr, ok := fragInout.(string); !ok || fragStr != "in" {
		t.Fatalf("FRAG_INOUT expected 'in', got %v", fragInout)
	}

	// Verify layouts -- check key layouts by name+location
	type layoutCheck struct {
		name     string
		location int
	}
	expectedLayouts := []layoutCheck{
		{"fragColor", 0},
		{"fragFlags", 1},
		{"fragPos", 2},
		{"fragTexCoords", 3},
		{"fragNormal", 4},
		{"fragMetallic", 5},
		{"fragRoughness", 6},
		{"fragEmissive", 7},
		{"fragTangentViewPos", 8},
		{"fragTangentFragPos", 9},
		{"fragLightTPos", 10},
		{"fragLightTDir", 14},
		{"fragPosLightSpace", 18},
		{"fragLightCount", 22},
		{"fragLightIndexes", 23},
	}

	found := make(map[string]bool)
	for i := range src.Layouts {
		l := &src.Layouts[i]
		if l.Source == "in" && l.Location >= 0 {
			for _, exp := range expectedLayouts {
				if l.Name == exp.name && l.Location == exp.location {
					found[exp.name] = true
				}
			}
		}
	}
	for _, exp := range expectedLayouts {
		if !found[exp.name] {
			t.Fatalf("missing expected layout: name=%q location=%d", exp.name, exp.location)
		}
	}

	t.Logf("All checks passed. Layouts: %d, Defines: %d", len(src.Layouts), len(src.defines))
}
