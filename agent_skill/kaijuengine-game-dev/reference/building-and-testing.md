# Building & Testing

## Prerequisites

- Go 1.25.0+
- C build tools
- Vulkan SDK

## Build

```bash
cd src
go build -tags="debug,editor,filedrop" -o ../ ./
```

### Build tags

- `debug` — include debug information
- `editor` — build with editor support
- `filedrop`, `rawsrc` — additional editor/source modes
- `ai_driver` — compile in the localhost AI-driver control server (see the
  `kaiju-aidriver` skill)
- Platform-specific source files: `*.windows.go`, `*.darwin.go`, `*.linux.go`,
  `*.android.go` (selected automatically by GOOS)

When using `//go:build` tags for platform-specific files, never duplicate struct
definitions/constructors/shared methods across tagged files — shared code lives in
a single untagged file; tagged files hold only what genuinely differs.

## Content

Game content lives in a `game_content/` directory at runtime (`assets.NewFileDatabase("game_content")`).
When building from the editor, content is placed in `database/content` by UUID (or
custom name). Content paths are relative to the working directory at runtime.

## Integration testing

For visuals or non-unit-testable behavior, use the integration-testing framework.
Tests live in `src/integration_testing`. In your test file's `init`, register the
launch function into the `tests` map. Review `integration_testing_helpers.go` for
available helpers, and add generic helpers there.

Build the executable from `src/` after adding a test:

```bash
go build -tags="debug,editor,filedrop,rawsrc" -o ../ ./
```

This produces `kaijuengine.com.exe` (or the platform equivalent) in the project
root. Run a test with the `integrationtest` argument:

```bash
kaijuengine.com.exe -integrationtest=screenshot
```

### Screenshot test example

```go
package integration_testing

import (
    "os"
    "kaijuengine.com/engine"
)

func init() {
    tests["screenshot"] = IntegrationTestScreenshot
}

// Generates "integration_test.png" in the working directory for visual review.
func IntegrationTestScreenshot(host *engine.Host) {
    createRedSphere(host)
    host.RunAfterFrames(3, func() {
        takeScreenshot(host)
        os.Exit(0)
    })
}
```

### Video recording

Integration tests can export video for vision models:

```go
rec := startVideoRecording(host, videoRecordingOptions{OutputPath: "integration_test.mp4"})
// ... drive the scene ...
rec.Stop() // call BEFORE any os.Exit so the encoder can finalize
```

- MP4 and WebM are supported; format is inferred from the extension or set via
  `Format`.
- Frames stream to an external `ffmpeg`, located via `KAIJU_FFMPEG` then `PATH` —
  ffmpeg must be installed.
- The recorder captures after successfully rendered frames and fails if the
  swapchain/window size changes during recording.

## Driving a running game

To interact with a *live* game (screenshot + inject mouse/keyboard, then loop),
build with `-tags ai_driver` and use the **`kaiju-aidriver`** skill, which talks to
the in-game localhost control server.
