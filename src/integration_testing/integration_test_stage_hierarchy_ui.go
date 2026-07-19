//go:build editor

/******************************************************************************/
/* integration_test_stage_hierarchy_ui.go                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/editor/editor_workspace/stage_workspace"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
)

const stageHierarchyUIScreenshotOutput = "integration_test_stage_hierarchy_ui.png"

var stageHierarchyTestUIMan *ui.Manager

type mockHierarchyEntity struct {
	id       string
	name     string
	selected bool
	locked   bool
	hidden   bool
	children []mockHierarchyEntity
}

func init() {
	tests["stage-hierarchy-ui"] = IntegrationTestStageHierarchyUI
}

func IntegrationTestStageHierarchyUI(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	stageHierarchyTestUIMan = &uiMan
	pageData := stage_workspace.WorkspaceUIData{
		Filters: map[string]int{
			"Mesh":     3,
			"Material": 2,
			"Texture":  4,
			"Stage":    1,
		},
		Tags: map[string]int{
			"environment": 2,
			"gameplay":    1,
		},
		CameraMode: "3D",
	}
	shellDoc, err := markup.DocumentFromHTMLAsset(&uiMan,
		"editor/ui/workspace/stage_workspace.go.html", pageData, stageHierarchyNoopFuncs())
	if err != nil {
		slog.Error("failed to load stage workspace UI", "error", err)
		os.Exit(1)
	}
	doc, err := markup.DocumentFromHTMLAsset(&uiMan,
		"editor/ui/workspace/stage_workspace_hierarchy.go.html", nil, stageHierarchyNoopFuncs())
	if err != nil {
		slog.Error("failed to load stage hierarchy UI", "error", err)
		os.Exit(1)
	}
	hideStageHierarchyElement(shellDoc, "ftdePrompt")
	hideStageHierarchyElement(shellDoc, "dimensionToggle")
	hideStageHierarchyElement(shellDoc, "giSettingsToggle")
	if err = populateStageHierarchyMock(doc); err != nil {
		slog.Error("failed to populate stage hierarchy mock", "error", err)
		os.Exit(1)
	}

	host.RunAfterFrames(12, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("stage hierarchy UI integration test failed", "error", err)
			os.Exit(1)
		}
		if err = assertStageHierarchyScreenshot(img); err != nil {
			_ = writeScreenshotImage(img, stageHierarchyUIScreenshotOutput)
			slog.Error("stage hierarchy screenshot failed smoke check",
				"path", stageHierarchyUIScreenshotOutput, "error", err)
			os.Exit(1)
		}
		if err = writeScreenshotImage(img, stageHierarchyUIScreenshotOutput); err != nil {
			slog.Error("failed to write stage hierarchy screenshot",
				"path", stageHierarchyUIScreenshotOutput, "error", err)
			os.Exit(1)
		}
		slog.Info("Screenshot captured", "path", stageHierarchyUIScreenshotOutput)
		os.Exit(0)
	})
}

func populateStageHierarchyMock(doc *document.Document) error {
	hierarchyArea, ok := doc.GetElementById("hierarchyArea")
	if !ok {
		return fmt.Errorf("missing #hierarchyArea")
	}
	doc.SetElementClasses(hierarchyArea, "edPanelBg", "sideBarTall")
	hideStageHierarchyElement(doc, "contentArea")
	hideStageHierarchyElement(doc, "detailsArea")
	hideStageHierarchyElement(doc, "ftdePrompt")
	hideStageHierarchyElement(doc, "dimensionToggle")
	hideStageHierarchyElement(doc, "dragPreview")
	hideStageHierarchyElement(doc, "hierarchyDragPreview")
	hideStageHierarchyElement(doc, "entityDataSelectorOverlay")
	hideStageHierarchyElement(doc, "tooltip")

	entityList, ok := doc.GetElementById("entityList")
	if !ok {
		return fmt.Errorf("missing #entityList")
	}
	template, ok := doc.GetElementById("entityTemplate")
	if !ok {
		return fmt.Errorf("missing #entityTemplate")
	}
	mockEntities := []mockHierarchyEntity{
		{
			id:       "player-rig",
			name:     "Player Rig",
			selected: true,
			children: []mockHierarchyEntity{
				{id: "camera-boom", name: "Camera Boom"},
				{
					id:   "weapon-mount",
					name: "Weapon Mount",
					children: []mockHierarchyEntity{
						{id: "muzzle-flash", name: "Muzzle Flash Socket"},
						{id: "reticle-anchor", name: "Reticle Anchor"},
					},
				},
				{id: "interaction-probe", name: "Interaction Probe", hidden: true},
			},
		},
		{
			id:   "environment",
			name: "Environment",
			children: []mockHierarchyEntity{
				{
					id:   "street-blockout",
					name: "Street Blockout",
					children: []mockHierarchyEntity{
						{id: "storefronts", name: "Storefront Modules"},
						{id: "alley-lights", name: "Alley Lights", locked: true},
					},
				},
				{id: "navmesh-volume", name: "Navmesh Volume"},
			},
		},
		{id: "director", name: "Mission Director", locked: true},
		{id: "debug-spawners", name: "Debug Spawn Points"},
	}
	for i := range mockEntities {
		addMockHierarchyRow(doc, template, entityList, mockEntities[i], 0)
	}
	template.UI.Hide()
	entityList.UI.SetDirty(ui.DirtyTypeLayout)
	return nil
}

func addMockHierarchyRow(doc *document.Document, template, parent *document.Element, entity mockHierarchyEntity, depth int) *document.Element {
	row := doc.DuplicateElement(template)
	doc.SetElementId(row, entity.id)
	if parent != template.Parent.Value() {
		doc.ChangeElementParent(row, parent)
	}
	header := stageHierarchyEntryHeader(row)
	rowClasses := []string{"hierarchyEntry"}
	if entity.selected {
		rowClasses = append(rowClasses, "hierarchyEntrySelected")
	}
	doc.SetElementClasses(row, rowClasses...)
	doc.SetElementClasses(header, "entryHeader")
	row.SetAttribute("data-collapsed", "false")
	stageHierarchyActionLabel(row, 0).SetText(stageHierarchyEyeIcon(entity.hidden))
	stageHierarchyActionLabel(row, 1).SetText(stageHierarchyLockIcon(entity.locked))
	stageHierarchyActionLabel(row, 2).SetText(stageHierarchyToggleIcon(len(entity.children) > 0, depth > 0))
	stageHierarchyNameLabel(row).SetText(entity.name)
	stageHierarchyNameLabel(row).Base().Layout().SetOffsetX(0)
	for i := range entity.children {
		addMockHierarchyRow(doc, template, row, entity.children[i], depth+1)
	}
	return row
}

func hideStageHierarchyElement(doc *document.Document, id string) {
	if elm, ok := doc.GetElementById(id); ok {
		elm.UI.Hide()
	}
}

func stageHierarchyEntryHeader(row *document.Element) *document.Element {
	return row.Children[0]
}

func stageHierarchyActionLabel(row *document.Element, idx int) *ui.Label {
	return stageHierarchyEntryHeader(row).Children[idx].InnerLabel()
}

func stageHierarchyNameSpan(row *document.Element) *document.Element {
	return stageHierarchyNameElement(row).Children[0]
}

func stageHierarchyNameElement(row *document.Element) *document.Element {
	return stageHierarchyEntryHeader(row).Children[3]
}

func stageHierarchyNameLabel(row *document.Element) *ui.Label {
	return stageHierarchyNameSpan(row).InnerLabel()
}

func stageHierarchyEyeIcon(hidden bool) string {
	if hidden {
		return "\ue8f5"
	}
	return "\ue8f4"
}

func stageHierarchyLockIcon(locked bool) string {
	if locked {
		return "\ue897"
	}
	return "\ue898"
}

func stageHierarchyToggleIcon(hasChildren, isChild bool) string {
	if hasChildren {
		return "\ue5c5"
	}
	if isChild {
		return "\ue5da"
	}
	return ""
}

func stageHierarchyNoopFuncs() map[string]func(*document.Element) {
	noop := func(*document.Element) {}
	names := []string{
		"toggleDimension", "toggleGISettings", "hierarchyDrop", "hierarchySearch", "selectEntity",
		"entityContextMenu", "entityDragStart", "entityDrop", "entityDragEnter",
		"entityDragExit", "entityToggleVisibility", "entityToggleLock",
		"entityToggleChildren", "submitDetailsName", "onRightClick", "setPosX",
		"setPosY", "setPosZ", "setRotX", "setRotY", "setRotZ", "setScaleX",
		"setScaleY", "setScaleZ", "clickSelectMesh", "meshIdDrop",
		"meshIdDragEnter", "meshIdDragExit", "clickSelectMaterial",
		"materialIdDrop", "materialIdDragEnter", "materialIdDragExit", "changeGIContribution",
		"changeShaderData", "showColorPicker", "pasteEntityData",
		"copyEntityData", "removeEntityData", "changeData", "clickSelectContentId",
		"contentIdDrop", "contentIdDragEnter", "contentIdDragExit",
		"clickSelectEntityId", "entityIdDrop", "entityIdDragEnter",
		"entityIdDragExit", "clearEntityId", "showEntityDataSelector",
		"pasteEntityDataAsNew", "closeEntityDataSelector", "searchEntityData",
		"addEntityData", "inputFilter", "tagFilter", "clickFilter",
		"dblClickEntry", "entryDragStart", "entryMouseEnter", "entryMouseLeave",
		"entryMouseMove", "rightClickContent",
	}
	funcs := make(map[string]func(*document.Element), len(names))
	for i := range names {
		funcs[names[i]] = noop
	}
	return funcs
}

func assertStageHierarchyScreenshot(img image.Image) error {
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("screenshot has invalid bounds %v", bounds)
	}
	headerPixels := 0
	accentPixels := 0
	textPixels := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r := int(r16 >> 8)
			g := int(g16 >> 8)
			b := int(b16 >> 8)
			if r >= 28 && r <= 52 && g >= 28 && g <= 52 && b >= 28 && b <= 58 {
				headerPixels++
			}
			if r > 95 && g < 80 && b < 80 {
				accentPixels++
			}
			if r > 175 && g > 175 && b > 175 {
				textPixels++
			}
		}
	}
	if headerPixels < 1200 {
		return fmt.Errorf("expected visible hierarchy row surfaces, found %d candidate pixels", headerPixels)
	}
	if accentPixels < 40 {
		return fmt.Errorf("expected selected/accent pixels in hierarchy, found %d", accentPixels)
	}
	if textPixels < 250 {
		return fmt.Errorf("expected visible hierarchy text/icons, found %d bright pixels", textPixels)
	}
	return nil
}
