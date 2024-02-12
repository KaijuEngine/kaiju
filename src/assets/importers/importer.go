package importers

type Importer interface {
	Handles(path string) bool
	Import(path string) error
}
