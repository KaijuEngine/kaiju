package assets

import (
	"kaiju/engine/assets/content_archive"
	"kaiju/platform/profiler/tracing"
)

type ArchiveDatabase struct {
	archive *content_archive.Archive
}

func NewArchiveDatabase(archive string, key []byte) (Database, error) {
	ar, err := content_archive.OpenArchiveFile(archive, key)
	return &ArchiveDatabase{archive: ar}, err
}

func (a *ArchiveDatabase) Cache(key string, data []byte) {}
func (a *ArchiveDatabase) CacheRemove(key string)        {}
func (a *ArchiveDatabase) CacheClear()                   {}

func (a *ArchiveDatabase) ReadText(key string) (string, error) {
	defer tracing.NewRegion("ArchiveDatabase.ReadText: " + key).End()
	b, err := a.archive.Read(key)
	return string(b), err
}

func (a *ArchiveDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("ArchiveDatabase.Read: " + key).End()
	return a.archive.Read(key)
}

func (a *ArchiveDatabase) Exists(key string) bool {
	defer tracing.NewRegion("ArchiveDatabase.Exists: " + key).End()
	return a.archive.Exists(key)
}

func (a *ArchiveDatabase) Close() {}
