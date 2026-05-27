/******************************************************************************/
/* integration_test_fbx_monkey.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"log/slog"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/framework"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/loaders"
)

const fbxMonkeyScreenshotOutput = "integration_test_fbx_monkey.png"

func init() {
	tests["fbx_monkey"] = IntegrationTestFBXMonkey
}

func IntegrationTestFBXMonkey(host *engine.Host) {
	host.PrimaryCamera().SetFarPlane(1000)
	host.PrimaryCamera().SetPosition(matrix.NewVec3(0, 0, -320))
	host.PrimaryCamera().SetLookAtWithUp(matrix.Vec3Zero(), matrix.Vec3Up())
	res, err := loaders.FBX("monkey.fbx", host.AssetDatabase())
	if err != nil {
		slog.Error("Failed to load FBX monkey", "error", err)
		os.Exit(1)
	}
	drawings, err := framework.CreateDrawingsBasic(host, res)
	if err != nil {
		slog.Error("Failed to create FBX monkey drawings", "error", err)
		os.Exit(1)
	}
	for _, drawing := range drawings.AllDrawings() {
		host.Drawings.AddDrawing(drawing)
	}
	host.RunAfterFrames(3, func() {
		takeScreenshotToFile(host, fbxMonkeyScreenshotOutput)
		os.Exit(0)
	})
}
