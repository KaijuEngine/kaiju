package assets

import (
	"kaiju/platform/profiler/tracing"
	"os"
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
	defer tracing.NewRegion("AssetDatabase.ReadText: " + key).End()
	if data, ok := a.cache[key]; ok {
		return string(data), nil
	}
	b, err := a.root.ReadFile(key)
	return string(b), err
}

func (a *FileDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("AssetDatabase.Read: " + key).End()
	if data, ok := a.cache[key]; ok {
		return data, nil
	}
	return a.root.ReadFile(key)
}

func (a *FileDatabase) Exists(key string) bool {
	defer tracing.NewRegion("AssetDatabase.Exists: " + key).End()
	if _, ok := a.cache[key]; ok {
		return true
	}
	_, err := a.root.Stat(key)
	return err == nil
}

func (a *FileDatabase) Close() {}
