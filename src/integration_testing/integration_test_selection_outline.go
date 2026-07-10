/******************************************************************************/
/* integration_test_selection_outline.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"math"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	selectionOutlineScreenshotOutput    = "integration_test_selection_outline_occluder.png"
	selectedPairOutlineScreenshotOutput = "integration_test_selection_outline_selected_pair.png"
)

func init() {
	tests["selection-outline-occluder"] = IntegrationTestSelectionOutlineOccluder
	tests["selection-outline-selected-pair"] = IntegrationTestSelectionOutlineSelectedPair
}

func IntegrationTestSelectionOutlineOccluder(host *engine.Host) {
	sphereCenter, sphereRadius := createSelectionOutlineScene(host, false)
	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("selection outline integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertSelectionOutlineOccluder(host, img, sphereCenter, sphereRadius); err != nil {
			if writeErr := writeScreenshotImage(img, selectionOutlineScreenshotOutput); writeErr != nil {
				slog.Error("failed to write selection outline screenshot", "error", writeErr)
			}
			slog.Error("selection outline integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, selectionOutlineScreenshotOutput); err != nil {
			slog.Error("failed to write selection outline screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func IntegrationTestSelectionOutlineSelectedPair(host *engine.Host) {
	sphereCenter, sphereRadius := createSelectionOutlineScene(host, true)
	host.RunAfterFrames(8, func() {
		img, err := captureScreenshotImage(host)
		if err != nil {
			slog.Error("selected pair outline integration test failed", "error", err)
			os.Exit(1)
		}
		if err := assertSelectionOutlineSelectedPair(host, img, sphereCenter, sphereRadius); err != nil {
			if writeErr := writeScreenshotImage(img, selectedPairOutlineScreenshotOutput); writeErr != nil {
				slog.Error("failed to write selected pair outline screenshot", "error", writeErr)
			}
			slog.Error("selected pair outline integration test failed", "error", err)
			os.Exit(1)
		}
		if err := writeScreenshotImage(img, selectedPairOutlineScreenshotOutput); err != nil {
			slog.Error("failed to write selected pair outline screenshot", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func createSelectionOutlineScene(host *engine.Host, selectSphere bool) (matrix.Vec3, matrix.Float) {
	host.PrimaryCamera().SetPositionAndLookAt(
		matrix.NewVec3(0, 4.0, 5.2),
		matrix.NewVec3(0, 0, 0),
	)

	floorData := shader_data_registry.Create("basic")
	floorData.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorWhite()
	shader_data_registry.StandardShaderDataFlagsSet(
		floorData, shader_data_registry.ShaderDataStandardFlagOutline)
	floor := engine.NewEntity(host.WorkGroup())
	floor.Transform.SetScale(matrix.NewVec3(5.0, 1.0, 5.0))
	addBasicDrawing(host, rendering.NewMeshPlane(host.MeshCache()), floorData, floor)

	const sphereRadius = 0.65
	sphereCenter := matrix.NewVec3(0.0, sphereRadius, 0.0)
	sphereData := shader_data_registry.Create("basic")
	sphereData.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorSky()
	if selectSphere {
		shader_data_registry.StandardShaderDataFlagsSet(
			sphereData, shader_data_registry.ShaderDataStandardFlagOutline)
	}
	sphere := engine.NewEntity(host.WorkGroup())
	sphere.Transform.SetPosition(sphereCenter)
	sphere.Transform.SetScale(matrix.NewVec3(sphereRadius, sphereRadius, sphereRadius))
	addBasicDrawing(host, rendering.NewMeshSphere(host.MeshCache(), 1, 48, 48), sphereData, sphere)

	return sphereCenter, sphereRadius
}

func addBasicDrawing(host *engine.Host, mesh *rendering.Mesh, sd rendering.DrawInstance, entity *engine.Entity) {
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	if err != nil {
		panic("you've probably got the wrong asset database path")
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   mat.CreateInstance([]*rendering.Texture{tex}),
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &entity.Transform,
		ViewCuller: &host.Cameras.Primary,
	})
}

func assertSelectionOutlineOccluder(host *engine.Host, img *image.RGBA, sphereCenter matrix.Vec3, sphereRadius matrix.Float) error {
	outlinePixels := countOutlinePixels(img, img.Bounds())
	if outlinePixels < 40 {
		return fmt.Errorf("expected visible outline around selected floor, found %d outline pixels", outlinePixels)
	}

	centerX, centerY, radius, ok := projectedSphereBounds(host, img, sphereCenter, sphereRadius)
	if !ok {
		return fmt.Errorf("failed to project sphere into screenshot")
	}
	haloBounds := annulusBounds(img.Bounds(), centerX, centerY, radius+8)
	haloPixels := countOutlinePixelsInAnnulus(img, centerX, centerY, radius-5, radius+8, haloBounds)
	if haloPixels > 4 {
		return fmt.Errorf("expected no selected outline around unselected sphere, found %d outline pixels", haloPixels)
	}
	return nil
}

func assertSelectionOutlineSelectedPair(host *engine.Host, img *image.RGBA, sphereCenter matrix.Vec3, sphereRadius matrix.Float) error {
	outlinePixels := countOutlinePixels(img, img.Bounds())
	if outlinePixels < 80 {
		return fmt.Errorf("expected visible outline around selected objects, found %d outline pixels", outlinePixels)
	}

	centerX, centerY, radius, ok := projectedSphereBounds(host, img, sphereCenter, sphereRadius)
	if !ok {
		return fmt.Errorf("failed to project sphere into screenshot")
	}
	haloBounds := annulusBounds(img.Bounds(), centerX, centerY, radius+8)
	haloPixels := countOutlinePixelsInAnnulus(img, centerX, centerY, radius-5, radius+8, haloBounds)
	if haloPixels < 20 {
		return fmt.Errorf("expected selected sphere boundary outline, found %d outline pixels", haloPixels)
	}
	return nil
}

func projectedSphereBounds(host *engine.Host, img *image.RGBA, center matrix.Vec3, radius matrix.Float) (float64, float64, float64, bool) {
	cam := host.PrimaryCamera()
	centerPx, ok := projectWorldToImage(host, img, center)
	if !ok {
		return 0, 0, 0, false
	}
	rightPx, ok := projectWorldToImage(host, img, center.Add(cam.Right().Scale(matrix.Float(radius))))
	if !ok {
		return 0, 0, 0, false
	}
	radiusPx := math.Hypot(
		float64(rightPx.X()-centerPx.X()),
		float64(rightPx.Y()-centerPx.Y()),
	)
	return float64(centerPx.X()), float64(centerPx.Y()), radiusPx, radiusPx > 1
}

func projectWorldToImage(host *engine.Host, img *image.RGBA, pos matrix.Vec3) (matrix.Vec2, bool) {
	cam := host.PrimaryCamera()
	screen, ok := matrix.Mat4ToScreenSpace(pos, cam.View(), cam.Projection(), cam.Viewport())
	if !ok {
		return matrix.Vec2{}, false
	}
	scaleX := matrix.Float(img.Bounds().Dx()) / matrix.Float(host.Window.Width())
	scaleY := matrix.Float(img.Bounds().Dy()) / matrix.Float(host.Window.Height())
	x := matrix.Float(screen.X()) * scaleX
	y := (matrix.Float(host.Window.Height()) - matrix.Float(screen.Y())) * scaleY
	return matrix.NewVec2(matrix.Float(x), matrix.Float(y)), true
}

func countOutlinePixels(img *image.RGBA, bounds image.Rectangle) int {
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isOutlinePixel(img.RGBAAt(x, y)) {
				count++
			}
		}
	}
	return count
}

func countOutlinePixelsInAnnulus(img *image.RGBA, cx, cy, innerRadius, outerRadius float64, bounds image.Rectangle) int {
	count := 0
	innerRadius = math.Max(0, innerRadius)
	inner2 := innerRadius * innerRadius
	outer2 := outerRadius * outerRadius
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			d2 := dx*dx + dy*dy
			if d2 >= inner2 && d2 <= outer2 && isOutlinePixel(img.RGBAAt(x, y)) {
				count++
			}
		}
	}
	return count
}

func annulusBounds(imgBounds image.Rectangle, cx, cy, radius float64) image.Rectangle {
	return image.Rect(
		clampPixel(matrix.Float(cx-radius), matrix.Float(imgBounds.Dx())),
		clampPixel(matrix.Float(cy-radius), matrix.Float(imgBounds.Dy())),
		clampPixel(matrix.Float(cx+radius), matrix.Float(imgBounds.Dx()))+1,
		clampPixel(matrix.Float(cy+radius), matrix.Float(imgBounds.Dy()))+1,
	)
}

func isOutlinePixel(c color.RGBA) bool {
	return isColorNear(c, color.RGBA{R: 251, G: 84, B: 43, A: 255}, 20)
}
