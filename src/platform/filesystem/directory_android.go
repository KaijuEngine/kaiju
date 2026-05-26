//go:build android

/******************************************************************************/
/* directory_android.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package filesystem

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"

	"kaijuengine.com/klib"
)

func knownPaths() map[string]string {
	return map[string]string{
		"Root": "/",
		"Home": "/home",
	}
}

func imageDirectory() (string, error) {
	userFolder, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userFolder, "Pictures"), nil
}

func gameDirectory() (string, error) {
	klib.NotYetImplemented(318)
	return "", errors.New("not yet implemented")
	//appdata, err := os.UserConfigDir()
	//if err != nil {
	//	return "", err
	//}
	//return filepath.Join(appdata, "../Local", build.CompanyDirName, build.Title.String()), nil
}

func openFileInTextEditor(path string) *exec.Cmd {
	return openFileBrowserCommand(path)
}

func openFileBrowserCommand(path string) *exec.Cmd {
	klib.NotYetImplemented(-1)
	return nil
}

func openFileBrowserSelectCommand(path string) *exec.Cmd {
	return openFileBrowserCommand(path)
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

func openFolderDialogWindow(startPath string, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	// TODO:  Eventually we'll create our own fully working file browser, instead of using current temp one
	klib.NotYetImplemented(-1)
	return nil
}
