/******************************************************************************/
/* terrain_workspace_test.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package terrain_workspace

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/matrix"
)

func TestAdjustTerrainBrushValueScalesAndClamps(t *testing.T) {
	if got := adjustTerrainBrushValue(10, 1, 1, 20); !matrix.ApproxTo(got, 11, matrix.Roughly) {
		t.Fatalf("expected increased brush value to be 11, got %f", got)
	}
	if got := adjustTerrainBrushValue(11, -1, 1, 20); !matrix.ApproxTo(got, 10, matrix.Roughly) {
		t.Fatalf("expected decreased brush value to be 10, got %f", got)
	}
	if got := adjustTerrainBrushValue(1, -1, 1, 20); got != 1 {
		t.Fatalf("expected brush value to clamp to min 1, got %f", got)
	}
	if got := adjustTerrainBrushValue(20, 1, 1, 20); got != 20 {
		t.Fatalf("expected brush value to clamp to max 20, got %f", got)
	}
}

func TestEffectiveTerrainBrushModeModifiers(t *testing.T) {
	if got := effectiveTerrainBrushMode(terrain.BrushRaise, false, false); got != terrain.BrushRaise {
		t.Fatalf("unmodified raise should stay raise, got %d", got)
	}
	if got := effectiveTerrainBrushMode(terrain.BrushRaise, true, true); got != terrain.BrushSmooth {
		t.Fatalf("shift should temporarily smooth, got %d", got)
	}
	if got := effectiveTerrainBrushMode(terrain.BrushRaise, false, true); got != terrain.BrushLower {
		t.Fatalf("ctrl should invert raise to lower, got %d", got)
	}
	if got := effectiveTerrainBrushMode(terrain.BrushLower, false, true); got != terrain.BrushRaise {
		t.Fatalf("ctrl should invert lower to raise, got %d", got)
	}
	if got := effectiveTerrainBrushMode(terrain.BrushSmooth, false, true); got != terrain.BrushSmooth {
		t.Fatalf("ctrl should leave smooth unchanged, got %d", got)
	}
}

func TestTerrainWorkspaceToolModeSwitching(t *testing.T) {
	var w TerrainWorkspace
	w.toolMode = TerrainToolHeightSculpt
	w.mode = terrain.BrushRaise
	w.textureMode = terrain.TextureBrushPaint

	w.clickModeTexture(nil)
	if w.toolMode != TerrainToolTexturePaint {
		t.Fatalf("expected texture mode after texture mode click, got %d", w.toolMode)
	}
	if w.mode != terrain.BrushRaise {
		t.Fatalf("texture mode switching should not change height brush, got %d", w.mode)
	}

	w.clickTextureErase(nil)
	if w.textureMode != terrain.TextureBrushErase {
		t.Fatalf("expected texture erase tool, got %d", w.textureMode)
	}
	if w.toolMode != TerrainToolTexturePaint {
		t.Fatalf("texture tool click should leave workspace in texture mode, got %d", w.toolMode)
	}

	w.clickToolSmooth(nil)
	if w.toolMode != TerrainToolHeightSculpt {
		t.Fatalf("height tool click should return to height sculpt mode, got %d", w.toolMode)
	}
	if w.mode != terrain.BrushSmooth {
		t.Fatalf("expected smooth height brush, got %d", w.mode)
	}
	if w.textureMode != terrain.TextureBrushErase {
		t.Fatalf("height tool switching should not change selected texture brush, got %d", w.textureMode)
	}

	w.clickToolLower(nil)
	if w.toolMode != TerrainToolHeightSculpt || w.mode != terrain.BrushLower {
		t.Fatalf("expected lower height brush in height mode, got mode %d brush %d", w.toolMode, w.mode)
	}
}

func TestTextureToolNames(t *testing.T) {
	tests := map[terrain.TextureBrushMode]string{
		terrain.TextureBrushPaint:         "Paint",
		terrain.TextureBrushErase:         "Erase",
		terrain.TextureBrushSmoothWeights: "Blend",
		terrain.TextureBrushFill:          "Fill",
		terrain.TextureBrushSample:        "Pick",
	}
	for mode, want := range tests {
		if got := textureToolName(mode); got != want {
			t.Fatalf("expected texture mode %d to read %q, got %q", mode, want, got)
		}
	}
}

func TestTerrainLayerTextureDiagnosticStatus(t *testing.T) {
	status := terrainLayerTextureDiagnosticStatus([]terrain.TerrainLayerTextureDiagnostic{{
		Layer:            2,
		Name:             "Rock",
		TextureContentID: "missing-rock",
	}})
	if !strings.Contains(status, "L3") || !strings.Contains(status, "missing-rock") {
		t.Fatalf("expected missing texture status to include layer and texture id, got %q", status)
	}
	status = terrainLayerTextureDiagnosticStatus([]terrain.TerrainLayerTextureDiagnostic{
		{Layer: 0, TextureContentID: "a"},
		{Layer: 1, TextureContentID: "b"},
	})
	if !strings.Contains(status, "2 missing") {
		t.Fatalf("expected aggregate missing texture status, got %q", status)
	}
}

func TestTerrainWorkspaceMarkupSplitsToolRows(t *testing.T) {
	data, err := os.ReadFile("../../editor_embedded_content/editor_content/editor/ui/workspace/terrain_workspace.go.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, id := range []string{
		`id="terrainModeRow"`,
		`id="heightToolRow"`,
		`id="heightBrushControls"`,
		`id="texturePaintRow"`,
		`onclick="clickModeTexture"`,
		`onclick="clickTexturePaint"`,
		`onclick="clickReplaceLayer"`,
		`id="textureLayerSelect"`,
		`id="textureLayerPalette"`,
		`id="textureLayerName"`,
		`id="textureFilter"`,
		`id="textureTintR"`,
	} {
		if !strings.Contains(html, id) {
			t.Fatalf("expected terrain workspace markup to contain %s", id)
		}
	}
	onclick := regexp.MustCompile(`onclick="([^"]+)"`)
	for _, match := range onclick.FindAllStringSubmatch(html, -1) {
		if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(match[1]) {
			t.Fatalf("onclick must name a single Go function, got %q", match[1])
		}
	}
}
