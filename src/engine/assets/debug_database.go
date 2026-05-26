/******************************************************************************/
/* debug_database.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package assets

import (
	"io/fs"
	"os"
	"path/filepath"

	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

type DebugContentDatabase struct{}

func (DebugContentDatabase) Cache(key string, data []byte) { /* No caching planned*/ }
func (DebugContentDatabase) CacheRemove(key string)        { /* No caching planned*/ }
func (DebugContentDatabase) CacheClear()                   { /* No caching planned*/ }
func (DebugContentDatabase) Close()                        {}

var cachedKeys = map[string]string{}

func findDebugDatabaseFile(key string) string {
	if path, ok := cachedKeys[key]; ok {
		return path
	}
	finalPath := ""
	paths := []string{"database/stock", "database/content", "database/debug"}
	for i := 0; i < len(paths) && finalPath == ""; i++ {
		filepath.Walk(paths[i], func(path string, info fs.FileInfo, err error) error {
			name := info.Name()
			cachedKeys[name] = path
			if finalPath != "" {
				return err
			}
			if name == key {
				finalPath = path
			}
			return err
		})
	}
	return finalPath
}

func (e DebugContentDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("DebugContentDatabase.Read: " + key).End()
	if filepath.IsAbs(key) {
		return filesystem.ReadFile(key[1:])
	}
	return os.ReadFile(findDebugDatabaseFile(key))
}

func (e DebugContentDatabase) ReadText(key string) (string, error) {
	defer tracing.NewRegion("DebugContentDatabase.ReadText: " + key).End()
	b, err := e.Read(key)
	return string(b), err
}

func (e DebugContentDatabase) Exists(key string) bool {
	defer tracing.NewRegion("DebugContentDatabase.Exists: " + key).End()
	if filepath.IsAbs(key) {
		return filesystem.FileExists(key)
	}
	_, err := os.Stat(findDebugDatabaseFile(key))
	return err == nil
}

func (DebugContentDatabase) PostWindowCreate(PostWindowCreateHandle) error { return nil }
