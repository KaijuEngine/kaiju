package assets

import (
	"io/fs"
	"kaiju/platform/filesystem"
	"kaiju/platform/profiler/tracing"
	"os"
	"path/filepath"
	"strings"
)

type DebugContentDatabase struct{}

func (DebugContentDatabase) Cache(key string, data []byte) { /* No caching planned*/ }
func (DebugContentDatabase) CacheRemove(key string)        { /* No caching planned*/ }
func (DebugContentDatabase) CacheClear()                   { /* No caching planned*/ }
func (DebugContentDatabase) Close()                        {}

func findDebugDatabaseFile(key string) string {
	finalPath := ""
	key = filepath.ToSlash(key)
	for i := range replacePrefixes {
		if strings.HasPrefix(key, replacePrefixes[i]) {
			key = strings.Replace(key, replacePrefixes[i], "database/stock/", 1)
			break
		}
	}
	if strings.HasPrefix(key, "database/stock/") {
		return key
	}
	filepath.Walk("database", func(path string, info fs.FileInfo, err error) error {
		if finalPath != "" {
			return nil
		}
		if info.Name() == key {
			finalPath = path
		}
		return nil
	})
	return finalPath
}

var replacePrefixes = []string{
	"fonts/",
	"textures/",
	"renderer/materials/",
	"renderer/passes/",
	"renderer/pipelines/",
	"renderer/shaders/",
	"renderer/spv/",
	"ui/",
}

func (e DebugContentDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("DebugContentDatabase.Read: " + key).End()
	if key[0] == absoluteFilePrefix {
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
	if key[0] == absoluteFilePrefix {
		return filesystem.FileExists(key[1:])
	}
	_, err := os.Stat(findDebugDatabaseFile(key))
	return err == nil
}
