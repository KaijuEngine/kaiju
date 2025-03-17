/******************************************************************************/
/* asset_database_info.go                                                     */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package asset_info

import (
	"encoding/json"
	"errors"
	"kaiju/platform/filesystem"
	"os"
	"path/filepath"
	"strings"
)

const (
	InfoExtension = ".adi"
	ProjectCache  = ".cache"
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
	Metadata any
}

func InitForCurrentProject() error {
	return os.MkdirAll(indexPath(), os.ModePerm)
}

func indexPath() string {
	return filepath.Join(ProjectCache, "index")
}

func toIndexPath(id string) string {
	return filepath.Join(indexPath(), id)
}

func toADI(path string) string {
	return path + InfoExtension
}

// Exists checks to see if a given path has a generated ADI file
// the file it searches for will be path/to/file.ext.adi
func Exists(path string) bool {
	s, err := os.Stat(toADI(path))
	return err == nil && !s.IsDir()
}

func New(path string, id string) AssetDatabaseInfo {
	if Exists(path) {
		return AssetDatabaseInfo{
			Children: make([]AssetDatabaseInfo, 0),
			Metadata: make(map[string]string),
		}
	}
	return AssetDatabaseInfo{
		ID:       id,
		Path:     path,
		Type:     strings.TrimPrefix(filepath.Ext(path), "."),
		Children: make([]AssetDatabaseInfo, 0),
		Metadata: make(map[string]string),
	}
}

func (a *AssetDatabaseInfo) SpawnChild(id string) AssetDatabaseInfo {
	return AssetDatabaseInfo{
		ID:       id,
		Path:     a.Path,
		Type:     a.Type,
		ParentID: a.ID,
		Children: make([]AssetDatabaseInfo, 0),
		Metadata: make(map[string]string),
	}
}

// Read will read the ADI file for the given path and return the
// AssetDatabaseInfo struct. Possible errors are:
//
// [-] ErrNoInfo: if the file does not exist
// [-] json.Unmarshal error: if the file is corrupted
// [-] filesystem.ReadTextFile error: if the file cannot be read
func Read(path string, metadata any) (AssetDatabaseInfo, error) {
	adi := AssetDatabaseInfo{
		Metadata: metadata,
	}
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
	if adi.Children == nil {
		adi.Children = make([]AssetDatabaseInfo, 0)
	}
	if adi.Metadata == nil {
		adi.Metadata = make(map[string]string)
	}
	return adi, nil
}

// Lookup will find any asset given it's id (path). This has no information for
// importers, so it will populate the metadata with a non-castable interface
func Lookup(id string) (AssetDatabaseInfo, error) {
	adiFile := toIndexPath(id)
	src, err := filesystem.ReadTextFile(adiFile)
	if err != nil {
		return AssetDatabaseInfo{}, err
	}
	adi, err := Read(src, nil)
	if err == nil {
		if adi.ID != id {
			for i := range adi.Children {
				if adi.Children[i].ID == id {
					adi = adi.Children[i]
					break
				}
			}
		}
	}
	return adi, err
}

func writeIndexes(info AssetDatabaseInfo) error {
	idx := filepath.Join(ProjectCache, "index", info.ID)
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

// ID returns the ID of the asset within it's ADI file, if
// the ADI file is not found, the read error is returned
func ID(path string) (string, error) {
	aid, err := Read(path, nil)
	if err != nil {
		return "", err
	}
	return aid.ID, nil
}
