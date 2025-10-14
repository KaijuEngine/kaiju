package content_database

import (
	"kaiju/games/editor/project/project_file_system"
)

func init() { contentCategories = append(contentCategories, Music{}) }

// Music is a [ContentCategory] represented by a file with a ".mp3" or ".ogg"
// extension. Music is as it sounds.
type Music struct{}
type MusicConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Music) Path() string       { return project_file_system.ContentMusicFolder }
func (Music) TypeName() string   { return "music" }
func (Music) ExtNames() []string { return []string{".mp3", ".ogg"} }

func (Music) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	return pathToBinaryData(src)
}
