package content_database

import "fmt"

type CategoryNotFoundError struct {
	Path string
}

func (e CategoryNotFoundError) Error() string {
	return fmt.Sprintf("failed to find category for file '%s'", e.Path)
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
