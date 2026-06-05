//go:build editor

/******************************************************************************/
/* integration_test_render_graph_workspace.go                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/render_graph_workspace"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

const (
	renderGraphSplineScreenshotOutput     = "integration_test_render_graph_spline.png"
	renderGraphCreateMenuScreenshotOutput = "integration_test_render_graph_create_menu.png"
	renderGraphNodeFieldsScreenshotOutput = "integration_test_render_graph_node_fields.png"
)

func init() {
	tests["render-graph-spline"] = IntegrationTestRenderGraphSpline
	tests["render-graph-create-menu"] = IntegrationTestRenderGraphCreateMenu
	tests["render-graph-node-fields"] = IntegrationTestRenderGraphNodeFields
}

func IntegrationTestRenderGraphSpline(host *engine.Host) {
	ed, err := newRenderGraphWorkspaceTestEditor(host)
	if err != nil {
		failRenderGraphSplineIntegration("create test editor", err)
	}
	createStageViewportSelectedSphere(host)
	workspace := &render_graph_workspace.RenderGraphWorkspace{}
	if err = workspace.Initialize(ed); err != nil {
		failRenderGraphSplineIntegration("initialize render graph workspace", err)
	}
	workspace.Open()
	updateId := host.Updater.AddUpdate(workspace.Update)

	host.RunAfterFrames(24, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			failRenderGraphSplineIntegration("capture screenshot", err)
		}
		if err = assertRenderGraphSplineScreenshot(host, workspace, img); err != nil {
			_ = writeScreenshotImage(img, renderGraphSplineScreenshotOutput)
			failRenderGraphSplineIntegration("screenshot smoke check", err)
		}
		if err = writeScreenshotImage(img, renderGraphSplineScreenshotOutput); err != nil {
			failRenderGraphSplineIntegration("write screenshot", err)
		}
		host.Updater.RemoveUpdate(&updateId)
		ed.cleanup()
		slog.Info("Screenshot captured", "path", renderGraphSplineScreenshotOutput)
		os.Exit(0)
	})
}

func IntegrationTestRenderGraphCreateMenu(host *engine.Host) {
	ed, err := newRenderGraphWorkspaceTestEditor(host)
	if err != nil {
		failRenderGraphCreateMenuIntegration("create test editor", err)
	}
	workspace := &render_graph_workspace.RenderGraphWorkspace{}
	if err = workspace.Initialize(ed); err != nil {
		failRenderGraphCreateMenuIntegration("initialize render graph workspace", err)
	}
	workspace.Open()
	updateId := host.Updater.AddUpdate(workspace.Update)

	host.RunAfterFrames(8, func() {
		workspace.ShowCreateNodeMenu()
	})
	host.RunAfterFrames(24, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			failRenderGraphCreateMenuIntegration("capture screenshot", err)
		}
		if err = assertRenderGraphCreateMenuScreenshot(host, workspace, img); err != nil {
			_ = writeScreenshotImage(img, renderGraphCreateMenuScreenshotOutput)
			failRenderGraphCreateMenuIntegration("screenshot smoke check", err)
		}
		if err = writeScreenshotImage(img, renderGraphCreateMenuScreenshotOutput); err != nil {
			failRenderGraphCreateMenuIntegration("write screenshot", err)
		}
		host.Updater.RemoveUpdate(&updateId)
		ed.cleanup()
		slog.Info("Screenshot captured", "path", renderGraphCreateMenuScreenshotOutput)
		os.Exit(0)
	})
}

func IntegrationTestRenderGraphNodeFields(host *engine.Host) {
	ed, err := newRenderGraphWorkspaceTestEditor(host)
	if err != nil {
		failRenderGraphNodeFieldsIntegration("create test editor", err)
	}
	workspace := &render_graph_workspace.RenderGraphWorkspace{}
	if err = workspace.Initialize(ed); err != nil {
		failRenderGraphNodeFieldsIntegration("initialize render graph workspace", err)
	}
	workspace.Open()
	updateId := host.Updater.AddUpdate(workspace.Update)

	host.RunAfterFrames(8, func() {
		workspace.CreateNodeFromAction(render_graph_workspace.CreateNodeActionArgs{
			NodeID: "value", X: 42, Y: 190, UsePosition: true,
		})
		workspace.CreateNodeFromAction(render_graph_workspace.CreateNodeActionArgs{
			NodeID: "color", X: 280, Y: 190, UsePosition: true,
		})
		workspace.CreateNodeFromAction(render_graph_workspace.CreateNodeActionArgs{
			NodeID: "vector", X: 518, Y: 190, UsePosition: true,
		})
		workspace.CreateNodeFromAction(render_graph_workspace.CreateNodeActionArgs{
			NodeID: "mix-color", X: 756, Y: 155, UsePosition: true,
		})
	})
	host.RunAfterFrames(32, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			failRenderGraphNodeFieldsIntegration("capture screenshot", err)
		}
		if err = assertRenderGraphNodeFieldsScreenshot(host, workspace, img); err != nil {
			_ = writeScreenshotImage(img, renderGraphNodeFieldsScreenshotOutput)
			failRenderGraphNodeFieldsIntegration("screenshot smoke check", err)
		}
		if err = writeScreenshotImage(img, renderGraphNodeFieldsScreenshotOutput); err != nil {
			failRenderGraphNodeFieldsIntegration("write screenshot", err)
		}
		host.Updater.RemoveUpdate(&updateId)
		ed.cleanup()
		slog.Info("Screenshot captured", "path", renderGraphNodeFieldsScreenshotOutput)
		os.Exit(0)
	})
}

func assertRenderGraphSplineScreenshot(host *engine.Host, workspace *render_graph_workspace.RenderGraphWorkspace, img *image.RGBA) error {
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("screenshot has invalid bounds %v", bounds)
	}
	graphRect := renderGraphWorkspaceGraphRect(host, workspace, bounds)
	graphTop := graphRect.Min.Y
	stageRect := image.Rect(graphRect.Min.X, bounds.Min.Y+24, graphRect.Max.X, graphTop)
	if pixels := countSaturatedPixels(img, stageRect); pixels < 150 {
		return fmt.Errorf("expected rendered scene content in top stage viewport, found %d saturated pixels", pixels)
	}
	greenSplinePixels := 0
	greenWirePixels := 0
	redAccentPixels := 0
	wireMinX := graphRect.Min.X + 230
	wireMaxX := graphRect.Min.X + 390
	wireMinY := graphTop + 105
	wireMaxY := graphTop + 245
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r := int(r16 >> 8)
			g := int(g16 >> 8)
			b := int(b16 >> 8)
			if g > 120 && g-r > 45 && g-b > 45 && r > 35 && b < 140 {
				greenSplinePixels++
				if x >= wireMinX && x <= wireMaxX && y >= wireMinY && y <= wireMaxY {
					greenWirePixels++
				}
			}
			if r > 80 && r < 150 && g > 20 && g < 95 && b > 20 && b < 105 && r-g > 35 {
				redAccentPixels++
			}
		}
	}
	if greenSplinePixels < 180 {
		return fmt.Errorf("expected visible green spline pixels, found %d", greenSplinePixels)
	}
	if greenWirePixels < 30 {
		return fmt.Errorf("expected visible green wire pixels between nodes, found %d", greenWirePixels)
	}
	if redAccentPixels < 1200 {
		return fmt.Errorf("expected visible node accent pixels, found %d", redAccentPixels)
	}
	return nil
}

func assertRenderGraphCreateMenuScreenshot(host *engine.Host, workspace *render_graph_workspace.RenderGraphWorkspace, img *image.RGBA) error {
	menu, ok := workspace.Doc.GetElementById("createNodeMenu")
	if !ok || menu == nil || menu.UI == nil || !menu.UI.IsActive() {
		return fmt.Errorf("create node menu is not active")
	}
	rect := elementBoundsRectangle(host, img.Bounds(), menu.UI)
	if rect.Dx() < 250 || rect.Dy() < 250 {
		return fmt.Errorf("create node menu has invalid screenshot bounds %v", rect)
	}
	redPixels := 0
	darkPanelPixels := 0
	lightTextPixels := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			i := img.PixOffset(x, y)
			r := int(img.Pix[i])
			g := int(img.Pix[i+1])
			b := int(img.Pix[i+2])
			if r > 80 && r < 150 && g > 20 && g < 95 && b > 20 && b < 105 && r-g > 35 {
				redPixels++
			}
			if r >= 15 && r <= 45 && g >= 15 && g <= 48 && b >= 18 && b <= 58 {
				darkPanelPixels++
			}
			if r > 180 && g > 180 && b > 180 {
				lightTextPixels++
			}
		}
	}
	if redPixels < 4 {
		return fmt.Errorf("expected visible create menu accent pixels, found %d", redPixels)
	}
	if darkPanelPixels < 10000 {
		return fmt.Errorf("expected visible create menu panel pixels, found %d", darkPanelPixels)
	}
	if lightTextPixels < 60 {
		return fmt.Errorf("expected visible create menu text pixels, found %d", lightTextPixels)
	}
	return nil
}

func assertRenderGraphNodeFieldsScreenshot(host *engine.Host, workspace *render_graph_workspace.RenderGraphWorkspace, img *image.RGBA) error {
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("screenshot has invalid bounds %v", bounds)
	}
	graphRect := renderGraphWorkspaceGraphRect(host, workspace, bounds)
	whiteSwatchPixels := 0
	fieldBorderPixels := 0
	for y := graphRect.Min.Y; y < graphRect.Max.Y; y++ {
		for x := graphRect.Min.X; x < graphRect.Max.X; x++ {
			i := img.PixOffset(x, y)
			r := int(img.Pix[i])
			g := int(img.Pix[i+1])
			b := int(img.Pix[i+2])
			if r > 220 && g > 220 && b > 220 {
				whiteSwatchPixels++
			}
			if r >= 45 && r <= 75 && g >= 48 && g <= 78 && b >= 58 && b <= 92 {
				fieldBorderPixels++
			}
		}
	}
	if whiteSwatchPixels < 80 {
		return fmt.Errorf("expected visible color field swatch pixels, found %d", whiteSwatchPixels)
	}
	if fieldBorderPixels < 180 {
		return fmt.Errorf("expected visible field control border pixels, found %d", fieldBorderPixels)
	}
	return nil
}

func renderGraphWorkspaceGraphRect(host *engine.Host, workspace *render_graph_workspace.RenderGraphWorkspace, bounds image.Rectangle) image.Rectangle {
	if workspace != nil && workspace.Doc != nil {
		if graphArea, ok := workspace.Doc.GetElementById("shaderGraphArea"); ok && graphArea != nil && graphArea.UI != nil {
			rect := elementBoundsRectangle(host, bounds, graphArea.UI)
			if rect.Dx() > 0 && rect.Dy() > 0 {
				return rect
			}
		}
	}
	graphTop := bounds.Min.Y + int(24+float32(bounds.Dy()-45)*0.5)
	return image.Rect(bounds.Min.X, graphTop, bounds.Max.X, bounds.Max.Y)
}

type renderGraphWorkspaceTestEditor struct {
	host           *engine.Host
	settings       editor_settings.Settings
	events         editor_events.EditorEvents
	history        memento.History
	project        project.Project
	projectDir     string
	projectFS      project_file_system.FileSystem
	cache          content_database.Cache
	contentPreview content_previews.ContentPreviewer
	stageView      editor_stage_view.StageView
}

func newRenderGraphWorkspaceTestEditor(host *engine.Host) (*renderGraphWorkspaceTestEditor, error) {
	projectDir, err := os.MkdirTemp("", "kaiju-render-graph-spline-")
	if err != nil {
		return nil, err
	}
	fs, err := project_file_system.New(projectDir)
	if err != nil {
		return nil, err
	}
	if err = fs.Mkdir(project_file_system.DatabaseFolder, os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, err
	}
	ed := &renderGraphWorkspaceTestEditor{
		host:       host,
		projectDir: projectDir,
		projectFS:  fs,
		cache:      content_database.New(),
		settings: editor_settings.Settings{
			RefreshRate:           60,
			BatteryRefreshRate:    60,
			UIScrollSpeed:         20,
			ShowGrid:              false,
			UseBatteryRefreshRate: false,
			EditorCamera: editor_settings.EditorCameraSettings{
				ZoomSpeed:          120,
				FlySpeed:           10,
				FlyBoostMultiplier: 4,
				FlyXSensitivity:    0.2,
				FlyYSensitivity:    0.2,
			},
		},
	}
	ed.history.Initialize(512)
	ed.project.Settings.Name = "Render Graph Workspace Integration"
	ed.project.Settings.EditorSettings.CameraMode = editor_controls.EditorCameraMode3d
	ed.stageView.Initialize(host, ed)
	return ed, nil
}

func (e *renderGraphWorkspaceTestEditor) cleanup() {
	if e.projectFS.Root != nil {
		_ = e.projectFS.Close()
	}
	if e.projectDir != "" {
		_ = os.RemoveAll(e.projectDir)
	}
}

func (e *renderGraphWorkspaceTestEditor) Host() *engine.Host { return e.host }

func (e *renderGraphWorkspaceTestEditor) Cache() *content_database.Cache { return &e.cache }

func (e *renderGraphWorkspaceTestEditor) ContentPreviewer() *content_previews.ContentPreviewer {
	return &e.contentPreview
}

func (e *renderGraphWorkspaceTestEditor) Actions() *editor_action.Service { return nil }

func (e *renderGraphWorkspaceTestEditor) Settings() *editor_settings.Settings { return &e.settings }

func (e *renderGraphWorkspaceTestEditor) Events() *editor_events.EditorEvents { return &e.events }

func (e *renderGraphWorkspaceTestEditor) History() *memento.History { return &e.history }

func (e *renderGraphWorkspaceTestEditor) Project() *project.Project { return &e.project }

func (e *renderGraphWorkspaceTestEditor) ProjectFileSystem() *project_file_system.FileSystem {
	return &e.projectFS
}

func (e *renderGraphWorkspaceTestEditor) StageView() *editor_stage_view.StageView {
	return &e.stageView
}

func (e *renderGraphWorkspaceTestEditor) BlurInterface() {}

func (e *renderGraphWorkspaceTestEditor) FocusInterface() {}

func (e *renderGraphWorkspaceTestEditor) IsInputFocused() bool { return false }

func (e *renderGraphWorkspaceTestEditor) SelectWorkspace(string) error { return nil }

func (e *renderGraphWorkspaceTestEditor) Workspace(string) (editor_workspace.Workspace, bool) {
	return nil, false
}

func (e *renderGraphWorkspaceTestEditor) Workspaces() []editor_workspace.Workspace { return nil }

func (e *renderGraphWorkspaceTestEditor) UpdateSettings() {}

func (e *renderGraphWorkspaceTestEditor) ShowReferences(string) {}

func failRenderGraphSplineIntegration(message string, err error) {
	if err != nil {
		slog.Error("render graph spline integration test failed",
			"path", renderGraphSplineScreenshotOutput,
			"message", message, "error", err)
	} else {
		slog.Error("render graph spline integration test failed",
			"path", renderGraphSplineScreenshotOutput,
			"message", message)
	}
	os.Exit(1)
}

func failRenderGraphCreateMenuIntegration(message string, err error) {
	if err != nil {
		slog.Error("render graph create menu integration test failed",
			"path", renderGraphCreateMenuScreenshotOutput,
			"message", message, "error", err)
	} else {
		slog.Error("render graph create menu integration test failed",
			"path", renderGraphCreateMenuScreenshotOutput,
			"message", message)
	}
	os.Exit(1)
}

func failRenderGraphNodeFieldsIntegration(message string, err error) {
	if err != nil {
		slog.Error("render graph node fields integration test failed",
			"path", renderGraphNodeFieldsScreenshotOutput,
			"message", message, "error", err)
	} else {
		slog.Error("render graph node fields integration test failed",
			"path", renderGraphNodeFieldsScreenshotOutput,
			"message", message)
	}
	os.Exit(1)
}
