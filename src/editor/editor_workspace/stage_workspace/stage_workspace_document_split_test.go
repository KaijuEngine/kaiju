/******************************************************************************/
/* stage_workspace_document_split_test.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const stageWorkspaceTemplateDir = "../../editor_embedded_content/editor_content/editor/ui/workspace/"

func TestStageWorkspaceShellDoesNotIncludePanels(t *testing.T) {
	data, err := os.ReadFile(stageWorkspaceTemplateDir + "stage_workspace.go.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, needle := range []string{
		"kaiju-include",
		"stage_workspace_hierarchy.go.html",
		"stage_workspace_details.go.html",
		"stage_workspace_content.go.html",
		"stage_workspace_hierarchy.css",
		"stage_workspace_details.css",
		"stage_workspace_content.css",
	} {
		if strings.Contains(html, needle) {
			t.Fatalf("stage shell should not contain %q", needle)
		}
	}
	for _, needle := range []string{`id="dimensionToggle"`, `id="giSettingsToggle"`, `onclick="toggleGISettings"`, `id="ftdePrompt"`} {
		if !strings.Contains(html, needle) {
			t.Fatalf("stage shell should still contain %q", needle)
		}
	}
}

func TestStagePanelTemplatesAreStandaloneDocuments(t *testing.T) {
	cases := []struct {
		name   string
		rootID string
		css    string
	}{
		{"stage_workspace_hierarchy.go.html", `id="hierarchyArea"`, "stage_workspace_hierarchy.css"},
		{"stage_workspace_details.go.html", `id="detailsArea"`, "stage_workspace_details.css"},
		{"stage_workspace_content.go.html", `id="contentArea"`, "stage_workspace_content.css"},
		{"stage_workspace_gi.go.html", `id="giArea"`, "stage_workspace_panels.css"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(stageWorkspaceTemplateDir + tc.name)
			if err != nil {
				t.Fatal(err)
			}
			html := string(data)
			for _, needle := range []string{
				"<!DOCTYPE html>",
				"<head>",
				"<body>",
				"stage_workspace_panels.css",
				tc.css,
				tc.rootID,
			} {
				if !strings.Contains(html, needle) {
					t.Fatalf("%s should contain %q", tc.name, needle)
				}
			}
		})
	}
}

func TestStageGINumericFieldsCommitWithoutRewritingActiveInput(t *testing.T) {
	data, err := os.ReadFile(stageWorkspaceTemplateDir + "stage_workspace_gi.go.html")
	if err != nil {
		t.Fatal(err)
	}
	inputs := regexp.MustCompile(`<input[^>]*type="number"[^>]*>`).FindAllString(string(data), -1)
	if len(inputs) == 0 {
		t.Fatal("stage GI panel should contain numeric inputs")
	}
	for _, input := range inputs {
		if !strings.Contains(input, `onblur="`) || !strings.Contains(input, `onsubmit="`) {
			t.Fatalf("GI numeric input must commit on blur and submit, not each keystroke: %s", input)
		}
		if strings.Contains(input, `onchange="`) {
			t.Fatalf("GI numeric input must not rewrite itself during typing: %s", input)
		}
	}
}

func TestStageGIPanelOwnsItsScrollableContent(t *testing.T) {
	data, err := os.ReadFile(stageWorkspaceTemplateDir + "stage_workspace_gi.go.html")
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	if strings.Contains(html, `class="giBody"`) {
		t.Fatal("Stage GI controls must be direct children of the scroll container")
	}
	for _, needle := range []string{`#giArea`, `overflow-y: scroll`, `id="giBakeBoundsMode"`, `id="giStageStatus"`} {
		if !strings.Contains(html, needle) {
			t.Fatalf("Stage GI panel should contain %q", needle)
		}
	}
}
