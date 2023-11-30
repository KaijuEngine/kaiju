package assets

import "kaiju/filesystem"

type Database struct {
}

func NewDatabase() Database {
	return Database{}
}

func (a *Database) ReadAsset(key string) (string, error) {
	return filesystem.ReadTextFile(key)
}
