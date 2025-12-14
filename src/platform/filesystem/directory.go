/******************************************************************************/
/* directory.go                                                               */
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
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unsafe"
)

type DialogExtension struct {
	Name      string
	Extension string
}

// CreateDirectory creates a directory at the specified path with full permissions.
func CreateDirectory(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// DirectoryExists returns true if the directory exists at the specified path.
func DirectoryExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// DeleteDirectory deletes the directory at the specified path.
func DeleteDirectory(path string) error { return os.RemoveAll(path) }

// KnownDirectories returns a list of known, common directories on the current
// computer. On windows this is things like Photos, Documents, etc.
func KnownDirectories() map[string]string { return knownPaths() }

// ImageDirectory will attempt to find the default user directory where
// images are stored. This function is OS specific.
func ImageDirectory() (string, error) { return imageDirectory() }

// GameDirectory will attempt to find the default directory for the
// application to store and load it's data to and from
func GameDirectory() (string, error) {
	dir, err := gameDirectory()
	if err == nil {
		if _, err := os.Stat(dir); err != nil {
			return dir, os.MkdirAll(dir, os.ModePerm)
		}
	}
	return dir, err
}

// ListRecursive returns a list of all files and directories in the specified,
// it walks through all of the subdirectories as well.
func ListRecursive(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	return files, err
}

// ListFoldersRecursive returns a list of all directories in the specified,
// it walks through all of the subdirectories as well.
func ListFoldersRecursive(path string) ([]string, error) {
	var folders []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			folders = append(folders, path)
		}
		return nil
	})
	return folders, err
}

// ListFilesRecursive returns a list of all files in the specified,
// it walks through all of the subdirectories as well.
func ListFilesRecursive(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// CopyDirectory copies the directory at the source path to the destination path.
func CopyDirectory(src, dst string) error {
	dirInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !dirInfo.IsDir() {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(dst, dirInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func OpenFileBrowserToFolder(path string) error {
	err := openFileBrowserCommand(filepath.ToSlash(path)).Run()
	if err != nil {
		slog.Error("failed to open the file browser", "error", err)
	}
	return err
}

func OpenFileDialogWindow(startPath string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	return openFileDialogWindow(startPath, extensions, ok, cancel, windowHandle)
}

func OpenSaveFileDialogWindow(startPath string, fileName string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	return openSaveFileDialogWindow(startPath, fileName, extensions, ok, cancel, windowHandle)
}

func Zip(srcDir, outFile string, skipFiles, skipFolders, skipExtensions []string) error {
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()
	z := zip.NewWriter(out)
	defer z.Close()
	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if d.IsDir() && slices.Contains(skipFolders, relPath) {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			ext := filepath.Ext(d.Name())
			if slices.Contains(skipExtensions, ext) {
				return nil
			}
		}
		if d.IsDir() {
			_, err := z.Create(relPath + "/")
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		writer, err := z.Create(relPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		return err
	})
	return err
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}
		inFile, err := f.Open()
		if err != nil {
			return err
		}
		defer inFile.Close()
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()
		if _, err := io.Copy(outFile, inFile); err != nil {
			return err
		}
	}
	return nil
}
