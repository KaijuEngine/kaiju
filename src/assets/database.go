package assets

import (
	"kaiju/filesystem"
	"path/filepath"
)

type Database struct {
}

func NewDatabase() Database {
	return Database{}
}

func (a *Database) ReadText(key string) (string, error) {
	key = filepath.Join("content", key)
	return filesystem.ReadTextFile(key)
}

func (a *Database) Read(key string) ([]byte, error) {
	key = filepath.Join("content", key)
	return filesystem.ReadFile(key)
}

func (a *Database) Exists(key string) bool {
	key = filepath.Join("content", key)
	return filesystem.FileExists(key)
}

func (a *Database) Destroy() {

}
