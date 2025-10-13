package content_database

import (
	"kaiju/klib"
)

func init() { contentCategories = append(contentCategories, Font{}) }

type Font struct{}
type FontConfig struct{}

func (Font) Path() string       { return "font" }
func (Font) TypeName() string   { return "font" }
func (Font) ExtNames() []string { return []string{".ttf"} }

func (Font) Import(src string) (data []byte, dependencies []string, err error) {
	// TODO:  Call the msdf compile
	klib.NotYetImplemented(0)
	return []byte{}, dependencies, err
}
