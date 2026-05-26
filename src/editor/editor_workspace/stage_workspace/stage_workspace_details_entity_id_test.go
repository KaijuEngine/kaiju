/******************************************************************************/
/* stage_workspace_details_entity_id_test.go                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"os"
	"strings"
	"testing"

	"kaijuengine.com/engine"
)

func TestEntityIdDisplayNameIncludesFriendlyNameAndId(t *testing.T) {
	got := entityIdDisplayName("Target Entity", engine.EntityId("target-id"))
	want := "Target Entity (target-id)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestStageDetailsTemplateIncludesEntityIdControl(t *testing.T) {
	const templatePath = "../../editor_embedded_content/editor_content/editor/ui/workspace/stage_workspace_details.go.html"
	bin, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatal(err)
	}
	html := string(bin)
	for _, needle := range []string{
		`class="dataEntityId"`,
		`onclick="clickSelectEntityId"`,
		`ondrop="entityIdDrop"`,
		`onclick="clearEntityId"`,
		`class="dataContentId"`,
		`onclick="clickSelectContentId"`,
		`id="detailsMaterialBlock"`,
		`onclick="clickSelectMaterial"`,
		`ondrop="materialIdDrop"`,
	} {
		if !strings.Contains(html, needle) {
			t.Fatalf("expected details template to contain %s", needle)
		}
	}
}
