package content_database

type Content struct {
	Data     []byte
	Path     string
	Category ContentCategory
	Config   ContentConfig
}
