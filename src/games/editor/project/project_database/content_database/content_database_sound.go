package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Sound{}) }

type Sound struct{}
type SoundConfig struct{}

func (Sound) Path() string       { return "audio/sound" }
func (Sound) TypeName() string   { return "sound" }
func (Sound) ExtNames() []string { return []string{".wav"} }

func (Sound) Import(src string) (data []byte, dependencies []string, err error) {
	data, err = filesystem.ReadFile(src)
	return data, dependencies, err
}
