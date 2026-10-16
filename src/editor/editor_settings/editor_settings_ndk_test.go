package editor_settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTryFindAndroidNDKPathFromAndroidHome(t *testing.T) {
	sdk := t.TempDir()
	for _, v := range []string{"25.2.9519653", "27.0.12077973"} {
		if err := os.MkdirAll(filepath.Join(sdk, "ndk", v), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("ANDROID_NDK_HOME", "")
	t.Setenv("ANDROID_HOME", sdk)
	s := Settings{}
	s.tryFindAndroidNDKPath()
	want := filepath.Join(sdk, "ndk", "27.0.12077973")
	if s.BuildTools.AndroidNDK != want {
		t.Errorf("AndroidNDK = %q, want %q", s.BuildTools.AndroidNDK, want)
	}
}

func TestLatestNDKInFolderMissing(t *testing.T) {
	if got := latestNDKInFolder(filepath.Join(t.TempDir(), "ndk")); got != "" {
		t.Errorf("latestNDKInFolder = %q, want empty", got)
	}
}
