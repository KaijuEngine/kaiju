//go:build editor

/******************************************************************************/
/* integration_test_new_project_overlay.go                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"log/slog"
	"os"

	"kaijuengine.com/editor/editor_overlay/new_project"
	"kaijuengine.com/engine"
)

const newProjectOverlayScreenshotOutput = "integration_test_new_project_overlay.png"

func init() {
	tests["new-project-overlay"] = IntegrationTestNewProjectOverlay
}

func IntegrationTestNewProjectOverlay(host *engine.Host) {
	if _, err := new_project.Show(host, new_project.Config{
		RecentProjects: []string{
			`C:\KaijuProjects\FoxAnim`,
			`C:\KaijuProjects\TestCompile`,
			`C:\KaijuProjects\physics_mesh`,
			`C:\KaijuProjects\Sudoku`,
		},
		OnCreate: func(name, path, templatePath string) {},
		OnOpen:   func(path string) {},
	}); err != nil {
		slog.Error("failed to show new project overlay", "error", err)
		os.Exit(1)
	}

	host.RunAfterFrames(12, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("failed to capture new project overlay screenshot", "error", err)
			os.Exit(1)
		}
		if err = assertNewProjectOverlayScreenshot(img); err != nil {
			_ = writeScreenshotImage(img, newProjectOverlayScreenshotOutput)
			slog.Error("new project overlay screenshot failed smoke check", "path", newProjectOverlayScreenshotOutput, "error", err)
			os.Exit(1)
		}
		if err = writeScreenshotImage(img, newProjectOverlayScreenshotOutput); err != nil {
			slog.Error("failed to write new project overlay screenshot", "path", newProjectOverlayScreenshotOutput, "error", err)
			os.Exit(1)
		}
		slog.Info("Screenshot captured", "path", newProjectOverlayScreenshotOutput)
		os.Exit(0)
	})
}

func assertNewProjectOverlayScreenshot(img image.Image) error {
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("screenshot has invalid bounds %v", bounds)
	}
	redAccentPixels := 0
	brightTextPixels := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r := int(r16 >> 8)
			g := int(g16 >> 8)
			b := int(b16 >> 8)
			if r > 90 && g < 80 && b < 80 {
				redAccentPixels++
			}
			if r > 210 && g > 210 && b > 210 {
				brightTextPixels++
			}
		}
	}
	if redAccentPixels < 400 {
		return fmt.Errorf("expected visible red accent/button pixels, found %d", redAccentPixels)
	}
	if run := longestVerticalRedRun(img); run < bounds.Dy()/2 {
		return fmt.Errorf("expected continuous left accent border, longest vertical run was %d pixels", run)
	}
	if brightTextPixels < 400 {
		return fmt.Errorf("expected visible title/control text pixels, found %d", brightTextPixels)
	}
	titleRegion := image.Rect(
		bounds.Min.X+bounds.Dx()/5,
		bounds.Min.Y+bounds.Dy()/8,
		bounds.Min.X+bounds.Dx()/2,
		bounds.Min.Y+bounds.Dy()/5,
	)
	if countBrightPixelsInRect(img, titleRegion) < 150 {
		return fmt.Errorf("expected visible title text pixels in %v", titleRegion)
	}
	return nil
}

func countBrightPixelsInRect(img image.Image, rect image.Rectangle) int {
	bounds := img.Bounds()
	rect = rect.Intersect(bounds)
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r := int(r16 >> 8)
			g := int(g16 >> 8)
			b := int(b16 >> 8)
			if r > 210 && g > 210 && b > 210 {
				count++
			}
		}
	}
	return count
}

func longestVerticalRedRun(img image.Image) int {
	bounds := img.Bounds()
	best := 0
	for x := bounds.Min.X; x < bounds.Min.X+bounds.Dx()/2; x++ {
		run := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r16, g16, b16, _ := img.At(x, y).RGBA()
			r := int(r16 >> 8)
			g := int(g16 >> 8)
			b := int(b16 >> 8)
			if r > 90 && g < 70 && b < 70 {
				run++
				best = max(best, run)
			} else {
				run = 0
			}
		}
	}
	return best
}
