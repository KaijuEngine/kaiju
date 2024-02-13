/*****************************************************************************/
/* asset_database_info.go                                                    */
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

package asset_info

import (
	"encoding/json"
	"errors"
	"kaiju/editor/cache"
	"kaiju/filesystem"
	"os"
	"path/filepath"
	"strings"
)

const (
	InfoExtension = ".adi"
)

var (
	ErrNoInfo = errors.New("asset database does not have info for this file")
)

type AssetDatabaseInfo struct {
	ID       string
	Path     string
	Type     string
	ParentID string
	Children []AssetDatabaseInfo
}

func InitForCurrentProject() error {
	return os.MkdirAll(indexPath(), os.ModePerm)
}

func indexPath() string {
	return filepath.Join(cache.ProjectCacheFolder, "index")
}

func toIndexPath(id string) string {
	return filepath.Join(indexPath(), id)
}

func toADI(path string) string {
	return path + InfoExtension
}

func Exists(path string) bool {
	s, err := os.Stat(toADI(path))
	return err == nil && !s.IsDir()
}

func New(path string, id string) AssetDatabaseInfo {
	if Exists(path) {
		return AssetDatabaseInfo{}
	}
	return AssetDatabaseInfo{
		ID:   id,
		Path: path,
		Type: strings.TrimPrefix(filepath.Ext(path), "."),
	}
}

func (a *AssetDatabaseInfo) SpawnChild(id string) AssetDatabaseInfo {
	return AssetDatabaseInfo{
		ID:       id,
		Path:     a.Path,
		Type:     a.Type,
		ParentID: a.ID,
	}
}

func Read(path string) (AssetDatabaseInfo, error) {
	adi := AssetDatabaseInfo{}
	if !Exists(path) {
		return adi, ErrNoInfo
	}
	adiFile := toADI(path)
	src, err := filesystem.ReadTextFile(adiFile)
	if err != nil {
		return adi, err
	}
	if err := json.Unmarshal([]byte(src), &adi); err != nil {
		return adi, err
	}
	return adi, nil
}

func writeIndexes(info AssetDatabaseInfo) error {
	idx := filepath.Join(cache.ProjectCacheFolder, "index", info.ID)
	if err := filesystem.WriteTextFile(idx, info.Path); err != nil {
		return err
	}
	for _, child := range info.Children {
		if err := writeIndexes(child); err != nil {
			return err
		}
	}
	return nil
}

func Write(adi AssetDatabaseInfo) error {
	adiFile := toADI(adi.Path)
	src, err := json.Marshal(adi)
	if err != nil {
		return err
	}
	if err := filesystem.WriteTextFile(adiFile, string(src)); err != nil {
		return err
	}
	return writeIndexes(adi)
}

func Move(info AssetDatabaseInfo, newPath string) error {
	oldPath := info.Path
	oldAdiPath := toADI(oldPath)
	newAdiFile := toADI(newPath)
	info.Path = newPath
	if err := Write(info); err != nil {
		return err
	}
	return os.Rename(oldAdiPath, newAdiFile)
}

func ID(path string) (string, error) {
	aid, err := Read(path)
	if err != nil {
		return "", err
	}
	return aid.ID, nil
}
