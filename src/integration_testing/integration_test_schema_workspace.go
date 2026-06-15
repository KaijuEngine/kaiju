//go:build editor

/******************************************************************************/
/* integration_test_schema_workspace.go                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/editor/editor_workspace/schema_workspace"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/platform/hid"
)

const (
	schemaWorkspaceScreenshotOutput = "integration_test_schema_workspace.png"
)

func init() {
	tests["schema-workspace"] = IntegrationTestSchemaWorkspace
}

func IntegrationTestSchemaWorkspace(host *engine.Host) {
	ed, err := newRenderGraphWorkspaceTestEditor(host)
	if err != nil {
		failSchemaWorkspaceIntegration("create test editor", err)
	}
	workspace := &schema_workspace.SchemaWorkspace{}
	if err = workspace.Initialize(ed); err != nil {
		failSchemaWorkspaceIntegration("initialize schema workspace", err)
	}
	workspace.Open()
	updateId := host.Updater.AddUpdate(workspace.Update)

	host.RunAfterFrames(8, func() {
		button, ok := workspace.Doc.GetElementById("schemaAddProperties")
		if !ok || button == nil || button.UI == nil {
			failSchemaWorkspaceIntegration("find properties button", nil)
		}
		if !button.UI.ExecuteEvent(ui.EventTypeClick) {
			failSchemaWorkspaceIntegration("click properties button", nil)
		}
	})
	host.RunAfterFrames(14, func() {
		host.Window.Mouse.SetPosition(162, 157,
			float32(host.Window.Width()), float32(host.Window.Height()))
		host.Window.Mouse.SetDown(hid.MouseButtonLeft)
	})
	host.RunAfterFrames(15, func() {
		host.Window.Mouse.SetUp(hid.MouseButtonLeft)
	})
	host.RunAfterFrames(36, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			failSchemaWorkspaceIntegration("capture screenshot", err)
		}
		if err = assertSchemaWorkspaceScreenshot(host, workspace, img); err != nil {
			_ = writeScreenshotImage(img, schemaWorkspaceScreenshotOutput)
			failSchemaWorkspaceIntegration("screenshot smoke check", err)
		}
		if err = writeScreenshotImage(img, schemaWorkspaceScreenshotOutput); err != nil {
			failSchemaWorkspaceIntegration("write screenshot", err)
		}
		host.Updater.RemoveUpdate(&updateId)
		workspace.Shutdown()
		ed.cleanup()
		slog.Info("Screenshot captured", "path", schemaWorkspaceScreenshotOutput)
		os.Exit(0)
	})
}

func assertSchemaWorkspaceScreenshot(host *engine.Host, workspace *schema_workspace.SchemaWorkspace, img *image.RGBA) error {
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("screenshot has invalid bounds %v", bounds)
	}
	if workspace == nil || workspace.Doc == nil {
		return fmt.Errorf("schema workspace document was not initialized")
	}
	bar, ok := workspace.Doc.GetElementById("schemaActionBar")
	if !ok || bar == nil || bar.UI == nil || !bar.UI.IsActive() {
		return fmt.Errorf("schema action bar is not active")
	}
	barRect := elementBoundsRectangle(host, bounds, bar.UI)
	if barRect.Dx() < 300 || barRect.Dy() < 36 {
		return fmt.Errorf("schema action bar has invalid screenshot bounds %v", barRect)
	}
	if barRect.Min.Y < bounds.Min.Y+bounds.Dy()/2 {
		return fmt.Errorf("schema action bar is not in the lower half of the workspace: %v", barRect)
	}
	if pixels := countSchemaActionBarPixels(img, barRect); pixels < 1000 {
		return fmt.Errorf("expected visible schema action bar pixels, found %d", pixels)
	}
	buttonIDs := []string{
		"schemaNewButton",
		"schemaLoadButton",
		"schemaSaveButton",
		"schemaAddProperties",
		"schemaAddPatternProperties",
		"schemaAddDefinitions",
		"schemaAddItems",
		"schemaAddItemsArray",
	}
	for _, id := range buttonIDs {
		button, ok := workspace.Doc.GetElementById(id)
		if !ok || button == nil || button.UI == nil || !button.UI.IsActive() {
			return fmt.Errorf("schema add button %q is not active", id)
		}
		rect := elementBoundsRectangle(host, bounds, button.UI)
		if rect.Dx() < 70 || rect.Dy() < 18 {
			return fmt.Errorf("schema add button %q has invalid screenshot bounds %v", id, rect)
		}
		if !rect.In(barRect) {
			return fmt.Errorf("schema add button %q is outside the action bar: button=%v bar=%v", id, rect, barRect)
		}
		if pixels := countSchemaButtonPixels(img, rect); pixels < 120 {
			return fmt.Errorf("expected visible schema add button %q pixels, found %d", id, pixels)
		}
	}
	if workspace.NodeCount() != 2 {
		return fmt.Errorf("expected card action to create 2 schema nodes, found %d", workspace.NodeCount())
	}
	nodeRect := image.Rect(bounds.Min.X+24, bounds.Min.Y+42, bounds.Min.X+330, bounds.Min.Y+190).Intersect(bounds)
	accentPixels, bodyPixels := countSchemaPropertiesNodePixels(img, nodeRect)
	if accentPixels < 300 {
		return fmt.Errorf("expected visible properties node accent pixels, found %d", accentPixels)
	}
	if bodyPixels < 1200 {
		return fmt.Errorf("expected visible properties node body pixels, found %d", bodyPixels)
	}
	propertyRect := image.Rect(bounds.Min.X+360, bounds.Min.Y+42, bounds.Min.X+670, bounds.Min.Y+220).Intersect(bounds)
	propertyAccent, propertyBody := countSchemaPropertiesNodePixels(img, propertyRect)
	if propertyAccent < 300 {
		return fmt.Errorf("expected visible child property node accent pixels, found %d", propertyAccent)
	}
	if propertyBody < 1600 {
		return fmt.Errorf("expected visible child property node body pixels, found %d", propertyBody)
	}
	return nil
}

func countSchemaActionBarPixels(img *image.RGBA, rect image.Rectangle) int {
	rect = rect.Intersect(img.Bounds())
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			i := img.PixOffset(x, y)
			r := int(img.Pix[i])
			g := int(img.Pix[i+1])
			b := int(img.Pix[i+2])
			if r >= 12 && r <= 28 && g >= 14 && g <= 32 && b >= 18 && b <= 38 {
				count++
			}
		}
	}
	return count
}

func countSchemaButtonPixels(img *image.RGBA, rect image.Rectangle) int {
	rect = rect.Intersect(img.Bounds())
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			i := img.PixOffset(x, y)
			r := int(img.Pix[i])
			g := int(img.Pix[i+1])
			b := int(img.Pix[i+2])
			if r >= 30 && r <= 55 && g >= 34 && g <= 60 && b >= 40 && b <= 72 {
				count++
			}
		}
	}
	return count
}

func countSchemaPropertiesNodePixels(img *image.RGBA, rect image.Rectangle) (int, int) {
	rect = rect.Intersect(img.Bounds())
	accent := 0
	body := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			i := img.PixOffset(x, y)
			r := int(img.Pix[i])
			g := int(img.Pix[i+1])
			b := int(img.Pix[i+2])
			if r > 80 && r < 150 && g > 20 && g < 95 && b > 20 && b < 105 && r-g > 35 {
				accent++
			}
			if r >= 18 && r <= 34 && g >= 20 && g <= 38 && b >= 24 && b <= 48 {
				body++
			}
		}
	}
	return accent, body
}

func failSchemaWorkspaceIntegration(message string, err error) {
	if err != nil {
		slog.Error("schema workspace integration test failed",
			"path", schemaWorkspaceScreenshotOutput,
			"message", message, "error", err)
	} else {
		slog.Error("schema workspace integration test failed",
			"path", schemaWorkspaceScreenshotOutput,
			"message", message)
	}
	os.Exit(1)
}
