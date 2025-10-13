package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Css{}) }

type Css struct{}
type CssConfig struct{}

func (Css) Path() string       { return "ui/css" }
func (Css) TypeName() string   { return "css" }
func (Css) ExtNames() []string { return []string{".css"} }

func (Css) Import(src string) (data []byte, dependencies []string, err error) {
	txt, err := filesystem.ReadTextFile(src)
	return []byte(txt), dependencies, err
}
