package editor_cache

import (
	"encoding/json"
	"kaiju/filesystem"
	"path/filepath"
)

var editorConfig = map[string]any{}

func init() {
	readEditorConfigCache()
}

func SetEditorConfigValue(key string, value any) {
	editorConfig[key] = value
	saveEditorConfigCache()
}

func EditorConfigValue(key string) (any, bool) {
	v, ok := editorConfig[key]
	return v, ok
}

func saveEditorConfigCache() error {
	cache, err := cacheFolder()
	if err != nil {
		return err
	}
	str, err := json.Marshal(editorConfig)
	if err != nil {
		return err
	}
	return filesystem.WriteTextFile(filepath.Join(cache, configFile), string(str))
}

func readEditorConfigCache() error {
	cache, err := cacheFolder()
	if err != nil {
		return err
	}
	str, err := filesystem.ReadTextFile(filepath.Join(cache, configFile))
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), &editorConfig)
}
