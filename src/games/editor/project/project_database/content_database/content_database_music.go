package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Music{}) }

type Music struct{}
type MusicConfig struct{}

func (Music) Path() string       { return "audio/music" }
func (Music) TypeName() string   { return "music" }
func (Music) ExtNames() []string { return []string{".mp3", ".ogg"} }

func (Music) Import(src string) (data []byte, dependencies []string, err error) {
	data, err = filesystem.ReadFile(src)
	return data, dependencies, err
}
