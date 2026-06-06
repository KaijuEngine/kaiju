/******************************************************************************/
/* content_database_sound.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(Sound{}) }

// Sound is a [ContentCategory] represented by a file with a ".wav" extension.
// Sound is as it sounds.
type Sound struct{}
type SoundConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Sound) Path() string       { return project_file_system.ContentSoundFolder }
func (Sound) TypeName() string   { return "Sound" }
func (Sound) ExtNames() []string { return []string{".wav"} }

func (Sound) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Sound.Import").End()
	return pathToBinaryData(src)
}

func (c Sound) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Sound.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Sound) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
