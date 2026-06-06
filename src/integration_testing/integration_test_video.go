/******************************************************************************/
/* integration_test_video.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"fmt"
	"log/slog"
	"math"
	"os"

	"kaijuengine.com/engine"
	"kaijuengine.com/matrix"
)

func init() {
	tests["video"] = IntegrationTestVideo
}

func IntegrationTestVideo(host *engine.Host) {
	ball := createRedSphere(host)
	rec, err := startVideoRecording(host, videoRecordingOptions{
		OutputPath: standardVideoOutput,
		FPS:        30,
	})
	if err != nil {
		videoIntegrationFail("start video recording", err)
	}
	updateId := host.Updater.AddUpdate(func(float64) {
		x := math.Sin(host.Runtime()*2.0) * 1.5
		ball.Transform.SetPosition(matrix.NewVec3(matrix.Float(x), 0, 0))
	})
	host.RunAfterFrames(90, func() {
		host.Updater.RemoveUpdate(&updateId)
		if err := rec.Stop(); err != nil {
			videoIntegrationFail("stop video recording", err)
		}
		info, err := os.Stat(standardVideoOutput)
		if err != nil {
			videoIntegrationFail("stat video output", err)
		}
		if info.Size() == 0 {
			videoIntegrationFail("video output is empty", fmt.Errorf("path %s", standardVideoOutput))
		}
		slog.Info("Video captured", "path", standardVideoOutput, "bytes", info.Size())
		os.Exit(0)
	})
}

func videoIntegrationFail(message string, err error) {
	if err != nil {
		slog.Error("video integration test failed", "message", message, "error", err)
	} else {
		slog.Error("video integration test failed", "message", message)
	}
	os.Exit(1)
}
