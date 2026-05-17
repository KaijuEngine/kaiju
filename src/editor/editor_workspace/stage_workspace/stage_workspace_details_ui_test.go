/******************************************************************************/
/* stage_workspace_details_ui_test.go                                         */
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

package stage_workspace

import (
	"testing"

	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
)

func TestEnumOptionNameTrimsFieldPrefix(t *testing.T) {
	if got := enumOptionName("Shape", "ShapeTerrain"); got != "Terrain" {
		t.Fatalf("expected Terrain, got %q", got)
	}
	if got := enumOptionName("Mode", "Model"); got != "Model" {
		t.Fatalf("expected Model to remain untrimmed, got %q", got)
	}
}

func TestRigidBodyTerrainFieldVisibility(t *testing.T) {
	hidden := []string{"AssetKey", "Extent", "Mass", "Radius", "Height", "IsStatic"}
	for _, field := range hidden {
		if rigidBodyFieldVisibleForShape(field, engine_entity_data_physics.ShapeTerrain) {
			t.Fatalf("expected %s to be hidden for terrain rigid bodies", field)
		}
	}
	if !rigidBodyFieldVisibleForShape("Shape", engine_entity_data_physics.ShapeTerrain) {
		t.Fatal("expected Shape to stay visible for terrain rigid bodies")
	}
	if !rigidBodyFieldVisibleForShape("Mass", engine_entity_data_physics.ShapeBox) {
		t.Fatal("expected non-terrain rigid bodies to keep generic field visibility")
	}
}

func TestRigidBodyTerrainWarningVisibility(t *testing.T) {
	if !rigidBodyTerrainWarningVisible(engine_entity_data_physics.ShapeTerrain, false) {
		t.Fatal("expected terrain shape without terrain data to show a warning")
	}
	if rigidBodyTerrainWarningVisible(engine_entity_data_physics.ShapeTerrain, true) {
		t.Fatal("expected terrain shape with terrain data to hide the warning")
	}
	if rigidBodyTerrainWarningVisible(engine_entity_data_physics.ShapeBox, false) {
		t.Fatal("expected non-terrain shapes to hide the terrain warning")
	}
}
