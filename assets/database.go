package assets

import "kaiju/filesystem"

type Database struct {
}

func NewDatabase() Database {
	return Database{}
}

func (a *Database) ReadTextAsset(key string) (string, error) {
	return filesystem.ReadTextFile(key)
}

func (a *Database) ReadAsset(key string) ([]byte, error) {
	return filesystem.ReadFile(key)
}

func (a *Database) AssetExists(key string) bool {
	return filesystem.FileExists(key)
}
