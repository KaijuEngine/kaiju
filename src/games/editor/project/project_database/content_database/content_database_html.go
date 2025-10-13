package content_database

import "kaiju/platform/filesystem"

func init() { contentCategories = append(contentCategories, Html{}) }

type Html struct{}
type HtmlConfig struct{}

func (Html) Path() string       { return "ui/html" }
func (Html) TypeName() string   { return "html" }
func (Html) ExtNames() []string { return []string{".html"} }

func (Html) Import(src string) (data []byte, dependencies []string, err error) {
	txt, err := filesystem.ReadTextFile(src)
	return []byte(txt), dependencies, err
}
