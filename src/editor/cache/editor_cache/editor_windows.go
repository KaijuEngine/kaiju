package editor_cache

import (
	"encoding/json"
	"errors"
	"kaiju/filesystem"
	"path/filepath"
)

const (
	windowsFile = "layout.json"
)

type WindowInfo struct {
	X      int
	Y      int
	Width  int
	Height int
}

var windows = map[string]WindowInfo{}

func saveCache() error {
	cache, err := cacheFolder()
	if err != nil {
		return err
	}
	str, err := json.Marshal(windows)
	if err != nil {
		return err
	}
	return filesystem.WriteTextFile(filepath.Join(cache, windowsFile), string(str))
}

func readCache() error {
	cache, err := cacheFolder()
	if err != nil {
		return err
	}
	str, err := filesystem.ReadTextFile(filepath.Join(cache, windowsFile))
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), &windows)
}

func SetWindow(key string, x, y, w, h int) error {
	windows[key] = WindowInfo{x, y, w, h}
	saveCache()
	return nil
}

func Window(key string) (WindowInfo, error) {
	if w, ok := windows[key]; ok {
		return w, nil
	}
	if err := readCache(); err != nil {
		return WindowInfo{}, err
	}
	if w, ok := windows[key]; ok {
		return w, nil
	}
	return WindowInfo{}, errors.New("window info not found")
}
