package content_database

type ContentConfig struct {
	Categories []string
	Tags       []string
	Css        CssConfig
	Font       FontConfig
	Html       HtmlConfig
	Material   MaterialConfig
	Mesh       MeshConfig
	Music      MusicConfig
	Sound      SoundConfig
	Spv        SpvConfig
	Texture    TextureConfig
}
