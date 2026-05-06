//go:build windows && !filedialog

package filesystem

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

func knownPaths() map[string]string {
	folders := map[string]*windows.KNOWNFOLDERID{
		"Desktop":   windows.FOLDERID_Desktop,
		"Documents": windows.FOLDERID_Documents,
		"Downloads": windows.FOLDERID_Downloads,
		"Music":     windows.FOLDERID_Music,
		"Pictures":  windows.FOLDERID_Pictures,
		"Videos":    windows.FOLDERID_Videos,
	}
	paths := make(map[string]string, len(folders))
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		str := fmt.Sprintf("%c:\\", r)
		s, err := os.Stat(str)
		if err != nil || !s.IsDir() {
			continue
		}
		paths[str] = str
	}
	for name, id := range folders {
		path, err := windows.KnownFolderPath(id, 0)
		if err != nil {
			continue
		}
		paths[name] = path
	}
	return paths
}

func imageDirectory() (string, error) {
	userFolder, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userFolder, "Pictures"), nil
}

func gameDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Saved Games"), nil
}

func openFileInTextEditor(path string) *exec.Cmd {
	return exec.Command("code", path)
}

func openFileBrowserCommand(path string) *exec.Cmd {
	return exec.Command("explorer", path)
}

func openFileBrowserSelectCommand(path string) *exec.Cmd {
	return exec.Command("explorer", "/select,", path)
}

func openFileDialogWindow(_ string, _ []DialogExtension, _ func(string), _ func(), _ unsafe.Pointer) error {
	return errors.New("file dialog is disabled; rebuild with 'filedialog' tag")
}

func openSaveFileDialogWindow(_ string, _ string, _ []DialogExtension, _ func(string), _ func(), _ unsafe.Pointer) error {
	return errors.New("file dialog is disabled; rebuild with 'filedialog' tag")
}

func openFolderDialogWindow(_ string, _ func(string), _ func(), _ unsafe.Pointer) error {
	return errors.New("file dialog is disabled; rebuild with 'filedialog' tag")
}

func openNativeDialogWindow(_ NativeDialogRequest, _ func(NativeDialogResult)) error {
	return errors.New("native dialog request API requires 'filedialog' build tag")
}

func processDialogCallbacks() {}

func shutdownNativeDialogs() {}
