/******************************************************************************/
/* file_database.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"kaijuengine.com/platform/profiler/tracing"
)

type EmbedDatabase struct {
	cache       map[string][]byte
	flattened   map[string]string
	isFlattened bool
	efs         embed.FS
}

// Database for //go:embed directive embed.FS
func NewEmbedDatabase(efs embed.FS, flatten bool) (Database, error) {
	flattened := map[string]string{}
	if flatten {
		fs.WalkDir(efs, ".", func(path string, d os.DirEntry, err error) error {
			if d.IsDir() || err != nil {
				return nil
			}
			flattened[filepath.Base(path)] = path
			return nil
		})
	}

	return &EmbedDatabase{
		efs:         efs,
		cache:       make(map[string][]byte),
		flattened:   flattened,
		isFlattened: flatten,
	}, nil
}

func (e *EmbedDatabase) Cache(key string, data []byte) { e.cache[key] = data }
func (e *EmbedDatabase) CacheRemove(key string)        { delete(e.cache, key) }
func (e *EmbedDatabase) CacheClear()                   { clear(e.cache) }

func (e *EmbedDatabase) Close() {
	e.CacheClear()
	e.flattened = make(map[string]string)
}

func (e *EmbedDatabase) Exists(key string) bool {
	defer tracing.NewRegion("EmbedDatabase.Exists: " + key).End()
	if _, ok := e.cache[key]; ok {
		return ok
	}
	if e.isFlattened {
		_, ok := e.flattened[key]
		return ok
	}
	file, err := e.efs.Open(key)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

func (EmbedDatabase) PostWindowCreate(PostWindowCreateHandle) error {
	return nil
}

func (e *EmbedDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("EmbedDatabase.Read: " + key).End()
	if data, ok := e.cache[key]; ok {
		return data, nil
	}
	if e.isFlattened {
		key = e.flattened[key]
	}
	return e.efs.ReadFile(key)
}

func (e *EmbedDatabase) ReadText(key string) (string, error) {
	defer tracing.NewRegion("EmbedDatabase.ReadText: " + key).End()
	file, err := e.Read(key)
	if err != nil {
		return "", err
	}
	return string(file), nil
}
