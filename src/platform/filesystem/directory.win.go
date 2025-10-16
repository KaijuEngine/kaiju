//go:build windows

/******************************************************************************/
/* directory.win.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
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
	"kaiju/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func openFileBrowserCommand(path string) *exec.Cmd {
	return exec.Command("cmd.exe", "/C", "start", path)
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
