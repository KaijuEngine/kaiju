//go:build android

/******************************************************************************/
/* directory_android.go                                                       */
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

import (
	"errors"
	"kaiju/klib"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"
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

func openFileBrowserCommand(path string) *exec.Cmd {
	klib.NotYetImplemented(-1)
	return nil
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
