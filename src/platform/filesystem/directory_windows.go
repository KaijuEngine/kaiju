//go:build windows

/******************************************************************************/
/* directory_windows.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package filesystem

/*
#cgo windows LDFLAGS: -lcomdlg32
#cgo noescape   open_file_dialog
#cgo noescape   save_file_dialog
#include "directory.win32.h"
*/
import "C"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"kaijuengine.com/build"

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
		// TODO:  Get the drive name
		paths[str] = str
	}
	for name, id := range folders {
		path, err := windows.KnownFolderPath(id, 0)
		if err != nil {
			fmt.Printf("%s: %v\n", name, err)
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
	appdata, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appdata, "../Local", build.CompanyDirName, build.Title.AsFilePathString()), nil
}

func openFileInTextEditor(path string) *exec.Cmd {
	return exec.Command("cmd.exe", "/C", "notepad", path)
}

func openFileBrowserCommand(path string) *exec.Cmd {
	return exec.Command("cmd.exe", "/C", "start", path)
}

func openFileBrowserSelectCommand(path string) *exec.Cmd {
	return exec.Command("explorer.exe", "/select,", path)
}

func openFileDialogWindow(startPath string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	ext := strings.Builder{}
	for i := range extensions {
		e := &extensions[i]
		ext.WriteString(fmt.Sprintf("%s\n*%s\n", e.Name, e.Extension))
	}
	if len(extensions) == 0 {
		ext.WriteString("All Files\x00*.*\x00\x00")
	}
	cStartPath := C.CString(startPath)
	defer C.free(unsafe.Pointer(cStartPath))
	cExt := C.CString(ext.String())
	defer C.free(unsafe.Pointer(cExt))
	savePath := C.GoString(C.open_file_dialog(cStartPath, cExt, windowHandle))
	if savePath != "" {
		ok(savePath)
	} else if cancel != nil {
		cancel()
	}
	return nil
}

func openSaveFileDialogWindow(startPath string, fileName string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	ext := strings.Builder{}
	for i := range extensions {
		e := &extensions[i]
		ext.WriteString(fmt.Sprintf("%s\n*%s\n", e.Name, e.Extension))
	}
	if len(extensions) == 0 {
		ext.WriteString("All Files\x00*.*\x00\x00")
	}
	cStartPath := C.CString(startPath)
	defer C.free(unsafe.Pointer(cStartPath))
	cFileName := C.CString(fileName)
	defer C.free(unsafe.Pointer(cFileName))
	cExt := C.CString(ext.String())
	defer C.free(unsafe.Pointer(cExt))
	savePath := C.GoString(C.save_file_dialog(cStartPath, cFileName, cExt, windowHandle))
	if savePath != "" {
		ok(savePath)
	} else if cancel != nil {
		cancel()
	}
	return nil
}
