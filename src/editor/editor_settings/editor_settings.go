/******************************************************************************/
/* editor_settings.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_settings

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/KaijuEngine/uuid"
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	settingsFileName       = "settings.json"
	maxRecentProjectsCount = 5
)

type Settings struct {
	RecentProjects        []string `visible:"false"`
	RefreshRate           int32    `clamp:"60,0,320"`
	UseBatteryRefreshRate bool     `default:"false" label:"Use Battery Refresh Rate"`
	BatteryRefreshRate    int32    `default:"30" clamp:"30,0,320" label:"Battery Refresh Rate"`
	CodeEditor            string   `default:"code"`
	ImageEditor           string
	MeshEditor            string
	AudioEditor           string
	UIScrollSpeed         float32 `default:"20" label:"UI Scroll Speed"`
	ShowGrid              bool    `default:"true" label:"Show Viewport Grid"`
	EditorCamera          EditorCameraSettings
	Snapping              SnapSettings
	BuildTools            BuildToolSettings
	WebAPI                WebAPISettings                `visible:"false" label:"Web API"`
	ActionBindings        []editor_action.ActionBinding `visible:"false" label:"Action Bindings"`
	// Workspaces is the persisted enable / visible / order state for every
	// known workspace, keyed by Workspace.ID(). Slice order is the load /
	// tab order. The editor's reconcile step on startup adds defaults for
	// any registered workspace that is missing from this slice and drops
	// entries whose workspace is no longer registered. Hidden from the
	// reflection-rendered settings UI because the Workspaces panel renders
	// it with a bespoke drag-to-reorder + toggle layout.
	Workspaces []WorkspaceConfig `visible:"false"`
}

// WorkspaceConfig is a single workspace's persisted state.
//
// Enabled=false skips initialization entirely: the workspace is not added
// to the active set, its tab is not rendered, and event subscriptions are
// not wired up. Enabled=true means initialized and tabbed.
type WorkspaceConfig struct {
	ID      string `visible:"false"`
	Enabled bool
}

type EditorCameraSettings struct {
	ZoomSpeed          float32 `default:"120" label:"Zoom Speed"`
	FlySpeed           float32 `default:"10"`
	FlyBoostMultiplier float32 `default:"4" label:"Fly Boost Multiplier"`
	FlyXSensitivity    float32 `default:"0.2"`
	FlyYSensitivity    float32 `default:"0.2"`
}

type SnapSettings struct {
	TranslateIncrement float32
	RotateIncrement    float32
	ScaleIncrement     float32
}

type BuildToolSettings struct {
	AndroidNDK string `label:"Android NDK"`
	JavaHome   string
}

type WebAPISettings struct {
	Enabled bool
	Port    int32  `default:"1337"`
	APIKey  string `label:"API Key"`
}

// setDefaults explicitly sets default values for all settings.
// Struct tag defaults are informational for the Editor UI, we
// must still explicitly set them in code.
func (s *Settings) setDefaults() {
	s.RefreshRate = 60
	s.BatteryRefreshRate = 60
	s.CodeEditor = "code"
	s.UIScrollSpeed = 20
	s.ShowGrid = true
	s.EditorCamera.ZoomSpeed = 120
	s.EditorCamera.FlySpeed = 10
	s.EditorCamera.FlyBoostMultiplier = 4
	s.EditorCamera.FlyXSensitivity = 0.2
	s.EditorCamera.FlyYSensitivity = 0.2
	s.NormalizeWebAPI()
}

func (s *Settings) NormalizeWebAPI() {
	if s.WebAPI.Port <= 0 || s.WebAPI.Port > 65535 {
		s.WebAPI.Port = 1337
	}
	if strings.TrimSpace(s.WebAPI.APIKey) == "" {
		s.WebAPI.APIKey = GenerateWebAPIKey()
	}
}

func GenerateWebAPIKey() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err == nil {
		return base64.RawURLEncoding.EncodeToString(key)
	}
	return uuid.NewString() + uuid.NewString()
}

func (s *Settings) AddRecentProject(path string) {
	for i := len(s.RecentProjects) - 1; i >= 0; i-- {
		if strings.EqualFold(s.RecentProjects[i], path) {
			s.RecentProjects = slices.Delete(s.RecentProjects, i, i+1)
		}
	}
	s.RecentProjects = slices.Insert(s.RecentProjects, 0, path)
	if len(s.RecentProjects) > maxRecentProjectsCount {
		s.RecentProjects = s.RecentProjects[:maxRecentProjectsCount]
	}
	// goroutine
	go s.Save()
}

func (s *Settings) Save() error {
	defer tracing.NewRegion("Settings.Save").End()
	s.NormalizeWebAPI()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	f, err := os.Create(filepath.Join(appData, settingsFileName))
	if err != nil {
		return WriteError{err, false}
	}
	if err := json.NewEncoder(f).Encode(*s); err != nil {
		return WriteError{err, true}
	}
	return nil
}

func (s *Settings) Load() error {
	defer tracing.NewRegion("Settings.Load").End()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	// Set defaults before attempting to load.
	// An existing settings file overrides these values during decode,
	// but also populates previously untracked settings with non-zero
	// value defaults.
	s.setDefaults()
	path := filepath.Join(appData, settingsFileName)
	if _, err := os.Stat(path); err != nil {
		// If the settings file doesn't exist, then create it. It is returning
		// here as there is no need to continue with the load if we're saving
		return s.Save()
	}
	f, err := os.Open(path)
	if err != nil {
		return ReadError{err, false}
	}
	if err := json.NewDecoder(f).Decode(s); err != nil {
		return ReadError{err, true}
	}
	s.NormalizeWebAPI()
	if s.BuildTools.AndroidNDK == "" {
		s.tryFindAndroidNDKPath()
	}
	if s.BuildTools.JavaHome == "" {
		s.tryFindJavaHomePath()
	}
	return nil
}

func (s *Settings) tryFindAndroidNDKPath() {
	if env := os.Getenv("ANDROID_NDK_HOME"); env != "" {
		if info, err := os.Stat(env); err == nil && info.IsDir() {
			s.BuildTools.AndroidNDK = env
			return
		}
	}
	for _, ndk := range androidNDKSearchFolders() {
		if latest := latestNDKInFolder(ndk); latest != "" {
			s.BuildTools.AndroidNDK = latest
			return
		}
	}
}

// androidNDKSearchFolders lists the "ndk" folders of the Android SDK locations
// that the SDK manager and Android Studio use by default on each platform.
func androidNDKSearchFolders() []string {
	folders := []string{}
	for _, env := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if sdk := os.Getenv(env); sdk != "" {
			folders = append(folders, filepath.Join(sdk, "ndk"))
		}
	}
	if appdata, err := os.UserConfigDir(); err == nil && runtime.GOOS == "windows" {
		folders = append(folders, filepath.Join(appdata, "../Local/Android/Sdk/ndk"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		switch runtime.GOOS {
		case "darwin":
			folders = append(folders, filepath.Join(home, "Library/Android/sdk/ndk"))
		case "linux":
			folders = append(folders, filepath.Join(home, "Android/Sdk/ndk"))
		}
	}
	return folders
}

func latestNDKInFolder(ndk string) string {
	dir, err := os.ReadDir(ndk)
	if err != nil {
		return ""
	}
	dir = slices.DeleteFunc(dir, func(e os.DirEntry) bool { return !e.IsDir() })
	if len(dir) == 0 {
		return ""
	}
	slices.SortFunc(dir, func(a, b os.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})
	return filepath.Join(ndk, dir[len(dir)-1].Name())
}

func (s *Settings) tryFindJavaHomePath() {
	if env := os.Getenv("JAVA_HOME"); env != "" {
		if info, err := os.Stat(env); err == nil && info.IsDir() {
			s.BuildTools.JavaHome = env
			return
		}
	}
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		candidates = []string{
			`C:\Program Files\Android\Android Studio\jbr`,
			`C:\Program Files\Java`,
		}
	case "darwin":
		candidates = []string{
			"/Applications/Android Studio.app/Contents/jbr",
			"/Library/Java/JavaVirtualMachines",
		}
	default:
		candidates = []string{
			"/usr/lib/jvm",
			"/usr/java",
		}
	}
	for _, base := range candidates {
		if info, err := os.Stat(base); err != nil || !info.IsDir() {
			continue
		}
		s.BuildTools.JavaHome = base
		return
	}
}
