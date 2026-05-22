package integration_testing

import (
	"os"

	"kaijuengine.com/engine"
)

// This test will generate a screenshot file in the working directory named
// "integration_test.png". This image can be analyzed by external tools to
// ensure it presents a red circle.

func init() {
	tests["screenshot"] = IntegrationTestScreenshot
}

func IntegrationTestScreenshot(host *engine.Host) {
	createRedSphere(host)
	host.RunAfterFrames(3, func() {
		takeScreenshot(host)
		os.Exit(0)
	})
}
