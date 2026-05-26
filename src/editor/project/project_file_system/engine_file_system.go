/******************************************************************************/
/* engine_file_system.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project_file_system

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type EngineFileSystemInterface interface {
	fs.FS
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

type EngineFileSystem struct{ EngineFileSystemInterface }

var EngineFS EngineFileSystem

func (efs EngineFileSystem) CopyFolder(pfs *FileSystem, from, to string, skipExt []string) error {
	var err error
	var copyFolder func(path string) error
	copyFolder = func(path string) error {
		if strings.EqualFold(path, "editor") {
			return nil
		}
		relPath, _ := filepath.Rel(from, path)
		folder := filepath.Join(to, relPath)
		if path != "." {
			if err := pfs.Mkdir(folder, os.ModePerm); err != nil {
				return err
			}
		}
		var dir []fs.DirEntry
		if dir, err = efs.ReadDir(path); err != nil {
			return err
		}
		for i := range dir {
			name := dir[i].Name()
			if slices.Contains(skipExt, filepath.Ext(name)) {
				continue
			}
			if strings.EqualFold(path, "build") && strings.EqualFold(name, "raw_game.go") {
				continue
			}
			entryPath := filepath.ToSlash(filepath.Join(path, name))
			if dir[i].IsDir() {
				if copyFolder(entryPath); err != nil {
					return err
				} else {
					continue
				}
			}
			if slices.Contains(skipFiles, entryPath) {
				continue
			}
			f, err := efs.Open(entryPath)
			if err != nil {
				return err
			}
			defer f.Close()
			t, err := pfs.Create(filepath.Join(folder, dir[i].Name()))
			if err != nil {
				return err
			}
			defer t.Close()
			if _, err := io.Copy(t, f); err != nil {
				return err
			}
		}
		return nil
	}
	copyFolder(from)
	return err
}
