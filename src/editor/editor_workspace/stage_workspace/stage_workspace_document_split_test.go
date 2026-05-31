/******************************************************************************/
/* stage_workspace_document_split_test.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"os"
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
	for _, needle := range []string{`id="dimensionToggle"`, `id="ftdePrompt"`} {
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
