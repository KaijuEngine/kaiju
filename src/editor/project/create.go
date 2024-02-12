/*****************************************************************************/
/* create.go                                                                 */
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

package project

import (
	"errors"
	"kaiju/filesystem"
	"os"
	"path/filepath"
	"strings"
)

const projectTemplateFolder = "project_template"

func createSource(projTemplateFolder string) error {
	sourceDir := filepath.Join(projTemplateFolder, "/source")
	err := os.Mkdir(sourceDir, 0755)
	if err != nil {
		return err
	}
	mainFile := filepath.Join(sourceDir, "/source.go")
	_, err = os.Stat(mainFile)
	if err == nil {
		return errors.New("source file already exists and should not")
	}
	f, err := os.Create(mainFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(`package source

import "kaiju/engine"

func Main(host *engine.Host) {
	
}
`)
	return err
}

func setupBuildScripts(projectName, projTemplateFolder string) error {
	buildDir := filepath.Join(projTemplateFolder, "/build")
	files, err := filesystem.ListFilesRecursive(buildDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		src, err := filesystem.ReadTextFile(file)
		if err != nil {
			return err
		}
		src = strings.ReplaceAll(src, "[PROJECT_NAME]", projectName)
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(src)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateNew(projectName, path string) error {
	if filepath.Base(path) != projectName {
		return errors.New("project name and path do not match")
	}
	stat, err := os.Stat(path)
	if err != nil {
		if err = os.MkdirAll(path, 0755); err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return os.ErrExist
	}
	if err = filesystem.CopyDirectory(projectTemplateFolder, path); err != nil {
		return err
	}
	if err = setupBuildScripts(projectName, projectTemplateFolder); err != nil {
		return err
	}
	return createSource(path)
}
