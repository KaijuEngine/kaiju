/******************************************************************************/
/* project_file_system.go                                                     */
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

package project_file_system

import (
	"errors"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"
	"strings"
)

var (
	baseStructure = []string{
		DatabaseFolder,
		ContentFolder,
		ContentConfigFolder,
		SrcFolder,
		StockFolder,
		SrcFontFolder,
		SrcCharsetFolder,
		SrcPluginFolder,
		SrcRenderFolder,
		SrcShaderFolder,
	}
	contentStructure = []string{
		ContentAudioFolder,
		ContentMusicFolder,
		ContentSoundFolder,
		ContentFontFolder,
		ContentMeshFolder,
		ContentUiFolder,
		ContentHtmlFolder,
		ContentCssFolder,
		ContentRenderFolder,
		ContentMaterialFolder,
		ContentSpvFolder,
		ContentStageFolder,
		ContentTextureFolder,
	}
	coreRequiredFolders = []string{
		DatabaseFolder,
		ContentFolder,
		ContentConfigFolder,
		SrcFolder,
		StockFolder,
	}
)

// FileSystem is the rooted project folder that is responsible for accessing any
// files or folders within the project. The type is a composition of os.Root,
// so all functions availabe to that structure are available to this one. Helper
// functions specific to projects are extended to it's behavior.
//
// The FileSystem has no awareness of actual content/assets and simply
// understands the structure for the project and allows raw read/write access
// to the files within the project.
type FileSystem struct {
	*os.Root
}

// New creates a new FileSystem that is rooted to the given project path. This
// function does not care about the status of the given path and only expects
// that the path supplied is to a folder on the filesystem. If the supplied path
// does not exist, an attempt will be made to create the folder.
func New(rootPath string) (FileSystem, error) {
	defer tracing.NewRegion("project_file_system.New").End()
	fs := FileSystem{}
	var err error
	if s, err := os.Stat(rootPath); err != nil {
		if err = os.MkdirAll(rootPath, os.ModePerm); err != nil {
			return fs, PathError{Path: rootPath, Msg: "failed to create the path", Err: err}
		}
	} else if !s.IsDir() {
		return fs, PathError{Path: rootPath, Err: errors.New("the supplied path is not a folder")}
	}
	fs.Root, err = os.OpenRoot(rootPath)
	return fs, err
}

// IsValid will return true if this project file system has been setup already.
func (fs *FileSystem) IsValid() bool { return fs.Root != nil }

// SetupStructure goes through and ensure all the base folders are created for
// the project. This will only create the folders if they do not yet exist.
// Folders are often missing if pulling the project from version control, as
// empty folders are not typically submitted. For more information on folder
// structure layout, please review the high level editor design documentation
// in the
// [README](https://github.com/KaijuEngine/kaiju/blob/master/src/editor/README.md).
func (fs *FileSystem) SetupStructure() error {
	defer tracing.NewRegion("FileSystem.SetupStructure").End()
	for i := range baseStructure {
		if err := fs.Mkdir(baseStructure[i], os.ModePerm); err != nil {
			return err
		}
	}
	for i := range contentStructure {
		if err := fs.Mkdir(filepath.Join("database/content", contentStructure[i]), os.ModePerm); err != nil {
			return err
		}
		if err := fs.Mkdir(filepath.Join("database/config", contentStructure[i]), os.ModePerm); err != nil {
			return err
		}
	}
	if err := fs.WriteFile(ProjectConfigFile, []byte("{}"), os.ModePerm); err != nil {
		return err
	}
	if err := fs.createCodeProject(); err != nil {
		return err
	}
	return fs.copyStockContent()
}

// Used to review the loaded FileSystem to ensure that the primary folders
// required for this rooted FileSystem are present to be considered a project.
// This will return an error if the core files are missing. Please review the
// source code file for this function to review the required core files and
// folders used, they are set as local package variables.
func (fs *FileSystem) EnsureDatabaseExists() error {
	defer tracing.NewRegion("FileSystem.EnsureDatabaseExists").End()
	for i := range coreRequiredFolders {
		if s, err := fs.Stat(coreRequiredFolders[i]); err != nil {
			return err
		} else {
			if !s.IsDir() {
				return PathError{Path: coreRequiredFolders[i], Err: errors.New("could not locate the folder")}
			}
		}
	}
	return nil
}

// ReadDir is a wrapper around [os.ReadDir] since [os.Root] doesn't provide an
// interface to this function directly. This simply grabs the rooted directory
// path and joins the name argument to it before forwarding to [os.ReadDir].
func (fs *FileSystem) ReadDir(name string) ([]os.DirEntry, error) {
	defer tracing.NewRegion("FileSystem.ReadDir").End()
	return os.ReadDir(filepath.Join(fs.Name(), name))
}

// FullPath will return the a cleaned version of the rooted file system path
// with the supplied name joined onto it.
func (fs *FileSystem) FullPath(name string) string {
	return filepath.Clean(filepath.Join(fs.Name(), name))
}

// NormalizePath will attempt to determine if the input path is a location held
// within the project file system. If it is a path within this file systme, then
// it will return the path as a relative path to the system, otherwise it will
// return the path supplied.
func (fs *FileSystem) NormalizePath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(fs.Name()))
}

// FileExists will return true if the file exists in the rooted file system
func (fs *FileSystem) FileExists(name string) bool {
	s, err := fs.Stat(name)
	return err == nil && !s.IsDir()
}

// FolderExists will return true if the folder exists in the rooted file system
func (fs *FileSystem) FolderExists(name string) bool {
	s, err := fs.Stat(name)
	return err == nil && s.IsDir()
}

// Exists will return true if the target file or folder exists in the rooted
// file system
func (fs *FileSystem) Exists(name string) bool {
	_, err := fs.Stat(name)
	return err == nil
}
