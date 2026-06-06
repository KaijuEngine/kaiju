//go:build linux && !android

/******************************************************************************/
/* directory_linux.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package filesystem

import (
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"

	"kaijuengine.com/build"
	"kaijuengine.com/klib"
)

func knownPaths() map[string]string {
	out := map[string]string{
		"Root": "/",
		"Home": "/home",
	}
	if userHome, err := os.UserHomeDir(); err == nil && userHome != "" {
		out["UserHome"] = userHome
		common := []string{"Desktop", "Documents", "Downloads", "Music", "Pictures", "Videos"}
		for i := range common {
			p := filepath.Join(userHome, common[i])
			if s, err := os.Stat(p); err == nil && s.IsDir() {
				out[common[i]] = p
			}
		}
	}
	return out
}

func imageDirectory() (string, error) {
	userFolder, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userFolder, "Pictures"), nil
}

func gameDirectory() (string, error) {
	appdata, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appdata, build.CompanyDirName, build.Title.String()), nil
}

func openFileInTextEditor(path string) *exec.Cmd {
	return openFileBrowserCommand(path)
}

func openFileBrowserCommand(path string) *exec.Cmd {
	return exec.Command("xdg-open", path)
}

func openFileBrowserSelectCommand(path string) *exec.Cmd {
	// Most Linux file managers don't provide a reliable cross-desktop "select file"
	// interface via xdg-open -> open the containing directory
	return exec.Command("xdg-open", filepath.Dir(path))
}

func openFileDialogWindow(startPath string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	// TODO:  Eventually we'll create our own fully working file browser, instead of using current temp one
	klib.NotYetImplemented(-1)
	return nil
}

func openSaveFileDialogWindow(startPath string, fileName string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	// TODO:  Eventually we'll create our own fully working file browser, instead of using current temp one
	klib.NotYetImplemented(-1)
	return nil
}
