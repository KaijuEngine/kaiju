/******************************************************************************/
/* stage_workspace_details_ui_test.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"os"
	"strings"
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

func TestStageDetailsTemplateIncludesTerrainTextureWarning(t *testing.T) {
	data, err := os.ReadFile("../../editor_embedded_content/editor_content/editor/ui/workspace/stage_workspace.go.html")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `data-validation="terrain-texture-missing"`) {
		t.Fatal("expected stage details template to include terrain texture missing validation row")
	}
}
