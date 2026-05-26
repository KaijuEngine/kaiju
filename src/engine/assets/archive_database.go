/******************************************************************************/
/* archive_database.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package assets

import (
	"path/filepath"
	"runtime"

	"kaijuengine.com/engine/assets/content_archive"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

type ArchiveDatabase struct {
	archive     *content_archive.Archive
	archivePath string
	key         []byte
}

func NewArchiveDatabase(archive string, key []byte) (Database, error) {
	defer tracing.NewRegion("ArchiveDatabase.NewArchiveDatabase").End()
	switch runtime.GOOS {
	case "android":
		return &ArchiveDatabase{archivePath: archive, key: key}, nil
	default:
		ar, err := content_archive.OpenArchiveFile(archive, key)
		return &ArchiveDatabase{archive: ar}, err
	}
}

func (a *ArchiveDatabase) PostWindowCreate(windowHandle PostWindowCreateHandle) error {
	defer tracing.NewRegion("ArchiveDatabase.PostWindowCreate").End()
	switch runtime.GOOS {
	case "android":
		data, err := windowHandle.ReadApplicationAsset(a.archivePath)
		if err != nil {
			return err
		}
		a.archive, err = content_archive.OpenArchiveFromBytes(data, a.key)
		a.archivePath = ""
		a.key = []byte{}
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *ArchiveDatabase) Cache(key string, data []byte) {}
func (a *ArchiveDatabase) CacheRemove(key string)        {}
func (a *ArchiveDatabase) CacheClear()                   {}

func (a *ArchiveDatabase) ReadText(key string) (string, error) {
	defer tracing.NewRegion("ArchiveDatabase.ReadText: " + key).End()
	data, err := a.Read(key)
	return string(data), err
}

func (a *ArchiveDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("ArchiveDatabase.Read: " + key).End()
	if filepath.IsAbs(key) {
		return filesystem.ReadFile(key[1:])
	}
	return a.archive.Read(key)
}

func (a *ArchiveDatabase) Exists(key string) bool {
	defer tracing.NewRegion("ArchiveDatabase.Exists: " + key).End()
	if filepath.IsAbs(key) {
		return filesystem.FileExists(key[1:])
	}
	return a.archive.Exists(key)
}

func (a *ArchiveDatabase) Close() {}
