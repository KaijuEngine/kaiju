package content_database

import (
	"kaiju/games/editor/project/project_file_system"
)

func init() { contentCategories = append(contentCategories, Sound{}) }

// Sound is a [ContentCategory] represented by a file with a ".wav" extension.
// Sound is as it sounds.
type Sound struct{}
type SoundConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Sound) Path() string       { return project_file_system.ContentSoundFolder }
func (Sound) TypeName() string   { return "sound" }
func (Sound) ExtNames() []string { return []string{".wav"} }

func (Sound) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	return pathToBinaryData(src)
}
