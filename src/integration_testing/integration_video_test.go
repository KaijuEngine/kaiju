/******************************************************************************/
/* integration_video_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestNormalizeVideoRecordingOptionsDefaults(t *testing.T) {
	options, err := normalizeVideoRecordingOptions(videoRecordingOptions{
		FFmpegPath: "ffmpeg",
	})
	if err != nil {
		t.Fatalf("normalizeVideoRecordingOptions returned error: %v", err)
	}
	if options.OutputPath != standardVideoOutput {
		t.Fatalf("OutputPath = %q, want %q", options.OutputPath, standardVideoOutput)
	}
	if options.Format != videoFormatMP4 {
		t.Fatalf("Format = %q, want %q", options.Format, videoFormatMP4)
	}
	if options.FPS != videoDefaultFPS {
		t.Fatalf("FPS = %d, want %d", options.FPS, videoDefaultFPS)
	}
	if options.FrameStride != videoDefaultStride {
		t.Fatalf("FrameStride = %d, want %d", options.FrameStride, videoDefaultStride)
	}
}

func TestNormalizeVideoRecordingOptionsInfersAndAppendsExtensions(t *testing.T) {
	tests := []struct {
		name       string
		options    videoRecordingOptions
		wantPath   string
		wantFormat videoFormat
	}{
		{
			name: "mp4 extension",
			options: videoRecordingOptions{
				OutputPath: "clip.mp4",
				FFmpegPath: "ffmpeg",
			},
			wantPath:   "clip.mp4",
			wantFormat: videoFormatMP4,
		},
		{
			name: "webm extension",
			options: videoRecordingOptions{
				OutputPath: "clip.webm",
				FFmpegPath: "ffmpeg",
			},
			wantPath:   "clip.webm",
			wantFormat: videoFormatWebM,
		},
		{
			name: "explicit webm without extension",
			options: videoRecordingOptions{
				OutputPath: "clip",
				Format:     "webm",
				FFmpegPath: "ffmpeg",
			},
			wantPath:   "clip.webm",
			wantFormat: videoFormatWebM,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeVideoRecordingOptions(test.options)
			if err != nil {
				t.Fatalf("normalizeVideoRecordingOptions returned error: %v", err)
			}
			if got.OutputPath != test.wantPath {
				t.Fatalf("OutputPath = %q, want %q", got.OutputPath, test.wantPath)
			}
			if got.Format != test.wantFormat {
				t.Fatalf("Format = %q, want %q", got.Format, test.wantFormat)
			}
		})
	}
}

func TestNormalizeVideoRecordingOptionsRejectsInvalidExtension(t *testing.T) {
	if _, err := normalizeVideoRecordingOptions(videoRecordingOptions{
		OutputPath: "clip.avi",
		FFmpegPath: "ffmpeg",
	}); err == nil {
		t.Fatal("normalizeVideoRecordingOptions accepted unsupported extension")
	}
	if _, err := normalizeVideoRecordingOptions(videoRecordingOptions{
		OutputPath: "clip.mp4",
		Format:     "webm",
		FFmpegPath: "ffmpeg",
	}); err == nil {
		t.Fatal("normalizeVideoRecordingOptions accepted mismatched format and extension")
	}
}

func TestFFmpegVideoArgsMP4(t *testing.T) {
	options := normalizedVideoRecordingOptions{
		OutputPath: "out.mp4",
		Format:     videoFormatMP4,
		FPS:        24,
	}
	got := ffmpegVideoArgs(options, 641, 479)
	want := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-s:v", "641x479",
		"-framerate", "24",
		"-i", "pipe:0",
		"-an",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "18",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"out.mp4",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ffmpegVideoArgs() = %#v, want %#v", got, want)
	}
}

func TestFFmpegVideoArgsWebM(t *testing.T) {
	options := normalizedVideoRecordingOptions{
		OutputPath: "out.webm",
		Format:     videoFormatWebM,
		FPS:        30,
	}
	got := ffmpegVideoArgs(options, 640, 480)
	want := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-s:v", "640x480",
		"-framerate", "30",
		"-i", "pipe:0",
		"-an",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-c:v", "libvpx-vp9",
		"-crf", "30",
		"-b:v", "0",
		"-pix_fmt", "yuv420p",
		"out.webm",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ffmpegVideoArgs() = %#v, want %#v", got, want)
	}
}

func TestVideoRecorderStartProcessMissingFFmpeg(t *testing.T) {
	recorder := &videoRecorder{
		options: normalizedVideoRecordingOptions{
			OutputPath:  filepath.Join(t.TempDir(), "out.mp4"),
			Format:      videoFormatMP4,
			FPS:         30,
			FrameStride: 1,
			FFmpegPath:  filepath.Join(t.TempDir(), "missing-ffmpeg"),
		},
		width:  2,
		height: 2,
	}
	if err := recorder.startProcess(); err == nil {
		t.Fatal("startProcess succeeded with a missing ffmpeg executable")
	}
}
