/*****************************************************************************/
/* main.go                                                                   */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
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
	_ "embed"
	"kaiju/filesystem"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed ignore.txt
var ignore string

func findRoot() (string, error) {
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
	root, err := findRoot()
	if err != nil {
		panic(err)
	}
	entries, err := filesystem.ListFoldersRecursive(root)
	if err != nil {
		panic(err)
	}
	ignoreEntries := strings.Split(ignore, "\n")
	for i := range ignoreEntries {
		ignoreEntries[i] = strings.TrimSpace(ignoreEntries[i])
	}
	os.Chdir(root + "/interpreter")
	for _, entry := range entries {
		entry = strings.Replace(entry, root, "", 1)
		entry = strings.TrimPrefix(strings.TrimPrefix(entry, "/"), "\\")
		skip := strings.HasPrefix(entry, ".") || len(strings.TrimSpace(entry)) == 0
		for i := 0; i < len(ignoreEntries) && !skip; i++ {
			skip = strings.HasPrefix(entry, ignoreEntries[i])
		}
		if !skip {
			pkg := "kaiju/" + entry
			println("Extracting " + pkg)
			err = exec.Command("yaegi", "extract", pkg).Run()
			if err != nil {
				panic(err)
			}
		}
	}
}
