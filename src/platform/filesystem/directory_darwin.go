//go:build darwin && !ios

/******************************************************************************/
/* directory_darwin.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package filesystem

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"kaijuengine.com/build"
)

func knownPaths() map[string]string {
	out := map[string]string{
		"Root": "/",
		// mac doesn’t have a fixed multi-user /home layout; omit "Home" to avoid confusion
	}
	if userHome, err := os.UserHomeDir(); err == nil && userHome != "" {
		out["UserHome"] = userHome
		common := []string{"Desktop", "Documents", "Downloads", "Pictures", "Music", "Movies"}
		for _, name := range common {
			p := filepath.Join(userHome, name)
			if s, err := os.Stat(p); err == nil && s.IsDir() {
				out[name] = p
			}
		}
	}
	return out
}

func imageDirectory() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, "Pictures"), nil
}

func gameDirectory() (string, error) {
	// macOS convention: ~/Library/Application Support/<Company>/<Title>
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(userHome, "Library", "Application Support", build.CompanyDirName, build.Title.String())
	return base, nil
}

func openFileInTextEditor(path string) *exec.Cmd {
	return openFileBrowserCommand(path)
}

func openFileBrowserCommand(path string) *exec.Cmd {
	if path == "" {
		return exec.Command("open", ".")
	}
	return exec.Command("open", path)
}

func openFileBrowserSelectCommand(path string) *exec.Cmd {
	if path == "" {
		return exec.Command("open", ".")
	}
	return exec.Command("open", "-R", path)
}

func openFileDialogWindow(startPath string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	// Use osascript (AppleScript) to show native macOS file picker
	script := `POSIX path of (choose file`

	if startPath != "" {
		script += ` default location "` + startPath + `"`
	}

	script += `)`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()

	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil
	}

	path := strings.TrimSpace(string(output))
	if path != "" {
		ok(path)
	} else if cancel != nil {
		cancel()
	}
	return nil
}

func openSaveFileDialogWindow(startPath string, fileName string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	// Use osascript (AppleScript) to show native macOS save dialog
	script := `POSIX path of (choose file name`

	if fileName != "" {
		script += ` default name "` + fileName + `"`
	}

	if startPath != "" {
		script += ` default location "` + startPath + `"`
	}

	script += `)`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()

	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil
	}

	path := strings.TrimSpace(string(output))
	if path != "" {
		ok(path)
	} else if cancel != nil {
		cancel()
	}
	return nil
}

func openFolderDialogWindow(startPath string, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	script := `POSIX path of (choose folder`
	if startPath != "" {
		script += ` default location "` + startPath + `"`
	}
	script += `)`
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil
	}
	path := strings.TrimSpace(string(output))
	if path != "" {
		ok(path)
	} else if cancel != nil {
		cancel()
	}
	return nil
}
