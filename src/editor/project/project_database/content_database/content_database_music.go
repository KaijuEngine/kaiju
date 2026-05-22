/******************************************************************************/
/* content_database_music.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() { addCategory(Music{}) }

// Music is a [ContentCategory] represented by a file with a ".mp3" or ".ogg"
// extension. Music is as it sounds.
type Music struct{}
type MusicConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (Music) Path() string       { return project_file_system.ContentMusicFolder }
func (Music) TypeName() string   { return "Music" }
func (Music) ExtNames() []string { return []string{".mp3", ".ogg"} }

func (Music) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Music.Import").End()
	return pathToBinaryData(src)
}

func (c Music) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("Music.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (Music) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
