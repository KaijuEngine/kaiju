/*****************************************************************************/
/* main.go                                                                   */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package main

import (
	"archive/zip"
	_ "embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

//go:embed ignore.txt
var ignore string

//go:embed launch.json.txt
var vsLaunch string

//go:embed settings.json.txt
var vsSettings string

//go:embed go.mod.txt
var goMod string

func findRootFolder() (string, error) {
	wd, err := os.Getwd()
	if _, goMain, _, ok := runtime.Caller(0); ok {
		if newWd, pathErr := filepath.Abs(filepath.Dir(goMain)); pathErr == nil {
			wd = filepath.Dir(newWd + "/../../")
		}
	} else if err != nil {
		return "", err
	}
	return wd, nil
}

func main() {
	root, err := findRootFolder()
	if err != nil {
		panic(err)
	}
	var srcEntries, contentEntries []fs.DirEntry
	if srcEntries, err = os.ReadDir(root); err != nil {
		panic(err)
	}
	if contentEntries, err = os.ReadDir(filepath.Join(root, "../content")); err != nil {
		panic(err)
	}

	ignoreEntries := strings.Split(ignore, "\n")
	for i := range ignoreEntries {
		ignoreEntries[i] = strings.TrimSpace(ignoreEntries[i])
	}
	addFiles := map[string]string{
		".vscode/launch.json":   vsLaunch,
		".vscode/settings.json": vsSettings,
		"src/go.mod":            goMod,
	}
	if err := os.Chdir(root); err != nil {
		panic(err)
	}
	zipTemplate("../project_template.zip", srcEntries, contentEntries, ignoreEntries, addFiles)
}

func zipTemplate(outPath string, srcEntries, contentEntries []fs.DirEntry, ignore []string, explicitFiles map[string]string) {
	file, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := zip.NewWriter(file)
	defer w.Close()
	var containingFolder string
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || slices.Contains(ignore, path) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		f, err := w.Create(filepath.Join(containingFolder, path))
		if err != nil {
			return err
		}
		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}
		return nil
	}
	containingFolder = "src"
	for _, entry := range srcEntries {
		err = filepath.Walk(entry.Name(), walker)
		if err != nil {
			panic(err)
		}
	}
	containingFolder = "content"
	if err := os.Chdir("../content"); err != nil {
		panic(err)
	}
	for _, entry := range contentEntries {
		err = filepath.Walk(entry.Name(), walker)
		if err != nil {
			panic(err)
		}
	}
	for to, text := range explicitFiles {
		f, err := w.Create(to)
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(text))
		if err != nil {
			panic(err)
		}
	}
}
