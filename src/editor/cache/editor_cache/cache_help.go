package editor_cache

import (
	"os"
	"path/filepath"
)

func cacheFolder() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	cache = filepath.Join(cache, CacheFolder)
	if _, err := os.Stat(cache); os.IsNotExist(err) {
		os.Mkdir(cache, os.ModePerm)
	}
	return cache, nil
}
