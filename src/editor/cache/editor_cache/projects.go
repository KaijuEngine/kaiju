/*****************************************************************************/
/* projects.go                                                               */
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

package editor_cache

import (
	"kaiju/filesystem"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	projectListFile = "projects.txt"
)

func projectCacheFolder() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cache = filepath.Join(cache, CacheFolder)
	if _, err := os.Stat(cache); os.IsNotExist(err) {
		os.Mkdir(cache, os.ModePerm)
	}
	return cache, nil
}

func AddProject(project string) error {
	cache, err := projectCacheFolder()
	if err != nil {
		return err
	}
	list, err := ListProjects()
	if err != nil {
		return err
	}
	if !slices.Contains(list, project) {
		list = append(list, project)
		filesystem.WriteTextFile(filepath.Join(cache, projectListFile), strings.Join(list, "\n"))
	}
	return nil
}

func ListProjects() ([]string, error) {
	cache, err := projectCacheFolder()
	if err != nil {
		return []string{}, err
	}
	projectsList := filepath.Join(cache, projectListFile)
	list, err := filesystem.ReadTextFile(projectsList)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		} else {
			return []string{}, err
		}
	}
	lines := strings.Split(list, "\n")
	projects := make([]string, 0, len(lines))
	for _, s := range lines {
		s = strings.TrimSpace(s)
		if s != "" {
			projects = append(projects, s)
		}
	}
	return projects, nil
}
