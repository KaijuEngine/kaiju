/******************************************************************************/
/* file_database.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package assets

import (
	"os"

	"kaijuengine.com/platform/profiler/tracing"
)

type FileDatabase struct {
	cache map[string][]byte
	root  *os.Root
}

func NewFileDatabase(root string) (Database, error) {
	r, err := os.OpenRoot(root)
	return &FileDatabase{
		cache: make(map[string][]byte),
		root:  r,
	}, err
}

func (a *FileDatabase) Cache(key string, data []byte) { a.cache[key] = data }
func (a *FileDatabase) CacheRemove(key string)        { delete(a.cache, key) }
func (a *FileDatabase) CacheClear()                   { clear(a.cache) }

func (a *FileDatabase) ReadText(key string) (string, error) {
	defer tracing.NewRegion("FileDatabase.ReadText: " + key).End()
	data, err := a.Read(key)
	return string(data), err
}

func (a *FileDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("FileDatabase.Read: " + key).End()
	if data, ok := a.cache[key]; ok {
		return data, nil
	}
	return a.root.ReadFile(key)
}

func (a *FileDatabase) Exists(key string) bool {
	defer tracing.NewRegion("FileDatabase.Exists: " + key).End()
	if _, ok := a.cache[key]; ok {
		return true
	}
	_, err := a.root.Stat(key)
	return err == nil
}

func (a *FileDatabase) Close() {
	if a.root != nil {
		a.root.Close()
		a.root = nil
	}
}
func (a *FileDatabase) PostWindowCreate(PostWindowCreateHandle) error { return nil }
