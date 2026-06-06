/******************************************************************************/
/* integration_video.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"kaijuengine.com/engine"
	"kaijuengine.com/rendering"
)

const (
	standardVideoOutput = "integration_test.mp4"
	videoFFmpegEnv      = "KAIJU_FFMPEG"
	videoDefaultFPS     = 30
	videoDefaultStride  = 1
)

type videoFormat string

const (
	videoFormatMP4  videoFormat = "mp4"
	videoFormatWebM videoFormat = "webm"
)

type videoRecordingOptions struct {
	OutputPath  string
	Format      string
	FPS         int
	FrameStride int
	FFmpegPath  string
}

type normalizedVideoRecordingOptions struct {
	OutputPath  string
	Format      videoFormat
	FPS         int
	FrameStride int
	FFmpegPath  string
}

type videoRecorder struct {
	host          *engine.Host
	options       normalizedVideoRecordingOptions
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stderr        bytes.Buffer
	width         int
	height        int
	frameWidth    int
	frameHeight   int
	framesSeen    int
	framesWritten int
	stopped       bool
	err           error
	stopErr       error
	mu            sync.Mutex
}

func startVideoRecording(host *engine.Host, options videoRecordingOptions) (*videoRecorder, error) {
	if host == nil {
		return nil, fmt.Errorf("cannot start video recording with a nil host")
	}
	normalized, err := normalizeVideoRecordingOptions(options)
	if err != nil {
		return nil, err
	}
	var width, height int
	err = fmt.Errorf("cannot start video recording without a valid renderer")
	host.RunOnRenderThread(func(device *rendering.GPUDevice) {
		width, height, err = videoRecordingSize(device)
	})
	if err != nil {
		return nil, err
	}
	recorder := &videoRecorder{
		host:    host,
		options: normalized,
		width:   width,
		height:  height,
	}
	if err = recorder.startProcess(); err != nil {
		return nil, err
	}
	recorder.scheduleNextFrame()
	slog.Info("Video recording started", "path", normalized.OutputPath,
		"format", normalized.Format, "fps", normalized.FPS)
	return recorder, nil
}

func normalizeVideoRecordingOptions(options videoRecordingOptions) (normalizedVideoRecordingOptions, error) {
	format, outputPath, err := normalizeVideoFormatAndPath(options.Format, options.OutputPath)
	if err != nil {
		return normalizedVideoRecordingOptions{}, err
	}
	fps := options.FPS
	if fps <= 0 {
		fps = videoDefaultFPS
	}
	stride := options.FrameStride
	if stride <= 0 {
		stride = videoDefaultStride
	}
	ffmpegPath, err := resolveVideoFFmpegPath(options.FFmpegPath)
	if err != nil {
		return normalizedVideoRecordingOptions{}, err
	}
	return normalizedVideoRecordingOptions{
		OutputPath:  outputPath,
		Format:      format,
		FPS:         fps,
		FrameStride: stride,
		FFmpegPath:  ffmpegPath,
	}, nil
}

func normalizeVideoFormatAndPath(formatValue, outputPath string) (videoFormat, string, error) {
	format, err := normalizeVideoFormat(formatValue)
	if err != nil {
		return "", "", err
	}
	ext := strings.ToLower(filepath.Ext(outputPath))
	if format == "" {
		switch ext {
		case ".mp4":
			format = videoFormatMP4
		case ".webm":
			format = videoFormatWebM
		case "":
			format = videoFormatMP4
		default:
			return "", "", fmt.Errorf("unsupported video output extension %q", ext)
		}
	} else if ext != "" && ext != "."+string(format) {
		return "", "", fmt.Errorf("video format %q does not match output extension %q", format, ext)
	}
	if outputPath == "" {
		outputPath = "integration_test." + string(format)
	} else if ext == "" {
		outputPath += "." + string(format)
	}
	return format, outputPath, nil
}

func normalizeVideoFormat(format string) (videoFormat, error) {
	format = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(format, ".")))
	switch format {
	case "":
		return "", nil
	case string(videoFormatMP4):
		return videoFormatMP4, nil
	case string(videoFormatWebM):
		return videoFormatWebM, nil
	default:
		return "", fmt.Errorf("unsupported video format %q", format)
	}
}

func resolveVideoFFmpegPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	if path = os.Getenv(videoFFmpegEnv); path != "" {
		return path, nil
	}
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("could not find ffmpeg; install ffmpeg, add it to PATH, or set %s", videoFFmpegEnv)
	}
	return path, nil
}

func videoRecordingSize(device *rendering.GPUDevice) (int, int, error) {
	if device == nil || !device.LogicalDevice.SwapChain.IsValid() {
		return 0, 0, fmt.Errorf("cannot record video without a valid swap chain")
	}
	size := device.LogicalDevice.SwapChain.Extent
	width := int(size.X())
	height := int(size.Y())
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("cannot record video with invalid swap chain size %dx%d", width, height)
	}
	return width, height, nil
}

func (r *videoRecorder) startProcess() error {
	args := ffmpegVideoArgs(r.options, r.width, r.height)
	cmd := exec.Command(r.options.FFmpegPath, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = &r.stderr
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	r.cmd = cmd
	r.stdin = stdin
	return nil
}

func ffmpegVideoArgs(options normalizedVideoRecordingOptions, width, height int) []string {
	args := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-s:v", fmt.Sprintf("%dx%d", width, height),
		"-framerate", strconv.Itoa(options.FPS),
		"-i", "pipe:0",
		"-an",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
	}
	switch options.Format {
	case videoFormatWebM:
		args = append(args,
			"-c:v", "libvpx-vp9",
			"-crf", "30",
			"-b:v", "0",
			"-pix_fmt", "yuv420p",
			options.OutputPath)
	default:
		args = append(args,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-crf", "18",
			"-pix_fmt", "yuv420p",
			"-movflags", "+faststart",
			options.OutputPath)
	}
	return args
}

func (r *videoRecorder) scheduleNextFrame() {
	r.host.RunAfterRender(r.captureFrame)
}

func (r *videoRecorder) captureFrame(device *rendering.GPUDevice, frame engine.RenderFrame) {
	scheduleNext := false
	r.mu.Lock()
	if r.stopped || r.err != nil {
		r.mu.Unlock()
		return
	}
	r.framesSeen++
	if (r.framesSeen-1)%r.options.FrameStride != 0 {
		scheduleNext = true
		r.mu.Unlock()
		r.scheduleNextFrame()
		return
	}
	pixels, width, height, err := captureScreenshotPixels(device)
	if err == nil {
		err = r.validateFrameSize(width, height, frame)
	}
	if err == nil {
		err = writeAll(r.stdin, pixels)
	}
	if err != nil {
		r.err = err
	} else {
		r.framesWritten++
		scheduleNext = true
	}
	r.mu.Unlock()
	if scheduleNext {
		r.scheduleNextFrame()
	}
}

func (r *videoRecorder) validateFrameSize(width, height int, frame engine.RenderFrame) error {
	if width != r.width || height != r.height {
		return fmt.Errorf("video frame size changed from %dx%d to %dx%d", r.width, r.height, width, height)
	}
	frameWidth := int(frame.Width)
	frameHeight := int(frame.Height)
	if r.frameWidth == 0 && r.frameHeight == 0 {
		r.frameWidth = frameWidth
		r.frameHeight = frameHeight
		return nil
	}
	if frameWidth != r.frameWidth || frameHeight != r.frameHeight {
		return fmt.Errorf("video render frame size changed from %dx%d to %dx%d",
			r.frameWidth, r.frameHeight, frameWidth, frameHeight)
	}
	return nil
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

func (r *videoRecorder) Stop() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	if r.stopped {
		err := errors.Join(r.err, r.stopErr)
		r.mu.Unlock()
		return err
	}
	r.stopped = true
	stdin := r.stdin
	cmd := r.cmd
	captureErr := r.err
	framesWritten := r.framesWritten
	r.mu.Unlock()

	var closeErr error
	if stdin != nil {
		closeErr = stdin.Close()
	}
	var waitErr error
	if cmd != nil {
		waitErr = cmd.Wait()
		if waitErr != nil {
			waitErr = fmt.Errorf("ffmpeg failed: %w%s", waitErr, r.ffmpegStderr())
		}
	}
	stopErr := errors.Join(captureErr, closeErr, waitErr)
	r.mu.Lock()
	r.stopErr = stopErr
	r.mu.Unlock()
	if stopErr == nil {
		slog.Info("Video recording stopped", "path", r.options.OutputPath,
			"frames", framesWritten)
	}
	return stopErr
}

func (r *videoRecorder) ffmpegStderr() string {
	text := strings.TrimSpace(r.stderr.String())
	if text == "" {
		return ""
	}
	return ": " + text
}
