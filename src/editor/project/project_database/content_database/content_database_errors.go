/******************************************************************************/
/* content_database_errors.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"errors"
	"fmt"
)

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

var DeleteContentMissingIdError = errors.New("id was blank")
