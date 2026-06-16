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
)

const schemaWorkspaceScreenshotOutput = "integration_test_schema_workspace.png"

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

	host.RunAfterFrames(24, func() {
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
	if workspace.Doc == nil {
		return fmt.Errorf("schema workspace document was not initialized")
	}
	requiredElements := []string{
		"schemaSidebar",
		"schemaMain",
		"schemaInspector",
		"schemaRootObjectPanel",
		"schemaDefinitionsPanel",
	}
	for _, id := range requiredElements {
		elm, ok := workspace.Doc.GetElementById(id)
		if !ok || elm == nil || elm.UI == nil {
			return fmt.Errorf("missing required schema workspace element #%s", id)
		}
		rect := elementBoundsRectangle(host, bounds, elm.UI)
		if rect.Dx() < 40 || rect.Dy() < 20 {
			return fmt.Errorf("schema workspace element #%s has invalid screenshot bounds %v", id, rect)
		}
	}

	mainElm, _ := workspace.Doc.GetElementById("schemaMain")
	mainRect := elementBoundsRectangle(host, bounds, mainElm.UI)
	darkPanelPixels := 0
	accentPixels := 0
	brightTextPixels := 0
	for y := mainRect.Min.Y; y < mainRect.Max.Y; y++ {
		for x := mainRect.Min.X; x < mainRect.Max.X; x++ {
			i := img.PixOffset(x, y)
			r := int(img.Pix[i])
			g := int(img.Pix[i+1])
			b := int(img.Pix[i+2])
			if r >= 24 && r <= 48 && g >= 24 && g <= 52 && b >= 26 && b <= 60 {
				darkPanelPixels++
			}
			if r >= 80 && r <= 150 && g >= 25 && g <= 80 && b >= 25 && b <= 90 && r-g > 35 {
				accentPixels++
			}
			if r > 170 && g > 170 && b > 170 {
				brightTextPixels++
			}
		}
	}
	if darkPanelPixels < 12000 {
		return fmt.Errorf("expected visible schema builder panels, found %d candidate pixels", darkPanelPixels)
	}
	if accentPixels < 80 {
		return fmt.Errorf("expected visible editor accent controls, found %d candidate pixels", accentPixels)
	}
	if brightTextPixels < 350 {
		return fmt.Errorf("expected visible schema builder text, found %d bright pixels", brightTextPixels)
	}

	inspectorElm, _ := workspace.Doc.GetElementById("schemaInspector")
	inspectorRect := elementBoundsRectangle(host, bounds, inspectorElm.UI)
	inspectorBrightTextPixels := 0
	for y := inspectorRect.Min.Y; y < inspectorRect.Max.Y; y++ {
		for x := inspectorRect.Min.X; x < inspectorRect.Max.X; x++ {
			i := img.PixOffset(x, y)
			if img.Pix[i] > 170 && img.Pix[i+1] > 170 && img.Pix[i+2] > 170 {
				inspectorBrightTextPixels++
			}
		}
	}
	if inspectorBrightTextPixels < 120 {
		return fmt.Errorf("expected visible metadata/json inspector text, found %d bright pixels", inspectorBrightTextPixels)
	}
	return nil
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
