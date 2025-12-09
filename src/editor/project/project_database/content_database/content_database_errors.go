/******************************************************************************/
/* content_database_errors.go                                                 */
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

package content_database

import "fmt"

type CategoryNotFoundError struct {
	Path string
	Type string
}

func (e CategoryNotFoundError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("failed to find category for type '%s'", e.Type)
	} else {
		return fmt.Sprintf("failed to find category for file '%s'", e.Path)
	}
}

type FontCharsetFilesMissingError struct {
	Path string
}

func (e FontCharsetFilesMissingError) Error() string {
	return fmt.Sprintf("no charset .txt files were found in the database directory '%s'", e.Path)
}

type NoMeshesInFileError struct {
	Path string
}

func (e NoMeshesInFileError) Error() string {
	return fmt.Sprintf("no meshes were found within the file '%s'", e.Path)
}

type MeshInvalidTextureError struct {
	Mesh          string
	SrcTexture    string
	BackupTexture string
}

func (e MeshInvalidTextureError) Error() string {
	return fmt.Sprintf("the texture '%s' on the mesh '%s' could not be located; also tried '%s'",
		e.SrcTexture, e.Mesh, e.BackupTexture)
}

type ImageImportError struct {
	Err   error
	Stage string
}

func (e ImageImportError) Error() string {
	return fmt.Sprintf("image import failed on stage '%s' with error: %v", e.Stage, e.Err)
}

type ReimportSourceMissingError struct {
	Id string
}

func (e ReimportSourceMissingError) Error() string {
	return fmt.Sprintf("could not re-import the content with id '%s' due to not having a source path", e.Id)
}

type ImageReimportUnsupportedError struct{}

func (ImageReimportUnsupportedError) Error() string {
	return "reaching this error is supposedly not possible, likely due to manual config file manipulation, the re-import category for this texture is not supported"
}

type ReimportMeshMissingError struct {
	Path string
	Name string
}

func (e ReimportMeshMissingError) Error() string {
	return fmt.Sprintf("re-import failed on mesh, the file %s was missing the mesh %s", e.Path, e.Name)
}
