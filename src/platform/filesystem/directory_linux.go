//go:build linux && !android

/******************************************************************************/
/* directory_linux.go                                                         */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/build"
	"kaiju/klib"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"
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

func openFileBrowserCommand(path string) *exec.Cmd {
	return exec.Command("xdg-open", path)
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
