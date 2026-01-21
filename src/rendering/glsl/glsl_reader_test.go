/******************************************************************************/
/* glsl_reader_test.go                                                        */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package glsl

import "testing"

const vertPath = "../../editor/editor_embedded_content/editor_content/renderer/src/pbr.frag"

func TestParse(t *testing.T) {
	src, err := Parse(vertPath, "")
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
		{"fragFlags", 29},
		{"", -1}, // Global uniform buffer
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
		l := &src.Layouts[i]
		if l.Name != layouts[i].name || l.Location != layouts[i].location {
			t.FailNow()
		}
	}
}
