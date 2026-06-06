/******************************************************************************/
/* project_file_system_folders.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project_file_system

import (
	"path/filepath"
	"strings"
)

type ContentPath string
type ConfigPath string

const (
	DatabaseFolder      = "database"
	ContentFolder       = "database/content"
	ContentConfigFolder = "database/config"
	SrcFolder           = "database/src"
	StockFolder         = "database/stock"
	DebugFolder         = "database/debug"
	ProjectConfigFile   = "database/project.json"
	EditorCache         = "database/.edcache"
)

const (
	EditorCacheContentPreviews = EditorCache + "/previews"
)

const (
	SrcFontFolder    = SrcFolder + "/font"
	SrcCharsetFolder = SrcFolder + "/font/charset"
	SrcPluginFolder  = SrcFolder + "/plugin"
	SrcRenderFolder  = SrcFolder + "/render"
	SrcShaderFolder  = SrcFolder + "/render/shader"
)

const (
	ContentAudioFolder           = "audio"
	ContentMusicFolder           = ContentAudioFolder + "/music"
	ContentSoundFolder           = ContentAudioFolder + "/sound"
	ContentFontFolder            = "font"
	ContentMeshFolder            = "mesh"
	ContentRenderFolder          = "render"
	ContentRenderGraphFolder     = ContentRenderFolder + "/graph"
	ContentMaterialFolder        = ContentRenderFolder + "/material"
	ContentRenderPassFolder      = ContentRenderFolder + "/renderpass"
	ContentShaderFolder          = ContentRenderFolder + "/shader"
	ContentParticlesFolder       = ContentRenderFolder + "/particles"
	ContentShaderPipelineFolder  = ContentRenderFolder + "/pipeline"
	ContentSpvFolder             = ContentRenderFolder + "/spv"
	ContentStageFolder           = "stage"
	ContentTemplateFolder        = "template"
	ContentTerrainFolder         = "terrain"
	ContentTextureFolder         = "texture"
	ContentTableFolder           = "table"
	ContentTableOfContentsFolder = ContentTableFolder + "/content"
	ContentUiFolder              = "ui"
	ContentHtmlFolder            = ContentUiFolder + "/html"
	ContentCssFolder             = ContentUiFolder + "/css"
)

const (
	KaijuSrcFolder            = "kaiju"
	ProjectCodeFolder         = "src"
	ProjectFileTemplates      = KaijuSrcFolder + "/file_templates"
	ProjectCodeGameHostFolder = ProjectCodeFolder + "/game_host"
	ProjectBuildFolder        = "build"
	ProjectBuildAndroidFolder = ProjectBuildFolder + "/android"
	ProjectVSCodeFolder       = ".vscode"
	ProjectLaunchJsonFile     = ".vscode/launch.json"
	ProjectCodeMain           = ProjectCodeFolder + "/main.go"
	ProjectCodeGame           = ProjectCodeFolder + "/game.go"
	ProjectModFile            = ProjectCodeFolder + "/go.mod"
	ProjectCodeGameHost       = ProjectCodeGameHostFolder + "/game_host.go"
	ProjectWorkFile           = "go.work"
	ProjectGitignoreFile      = ".gitignore"
	ProjectCodeGameTitle      = KaijuSrcFolder + "/build/title.go"
	EntityDataBindingInit     = ProjectCodeFolder + "/entity_data_binding_init.go"
)

func AsContentPath(path string) ContentPath {
	return ContentPath(filepath.ToSlash(path))
}

func HtmlPath(id string) ContentPath {
	return AsContentPath(filepath.Join(ContentFolder, ContentHtmlFolder, id))
}

func StagePath(id string) ContentPath {
	return AsContentPath(filepath.Join(ContentFolder, ContentStageFolder, id))
}

func ShaderPath(id string) ContentPath {
	return AsContentPath(filepath.Join(ContentFolder, ContentShaderFolder, id))
}

func SpvPath(id string) ContentPath {
	return AsContentPath(filepath.Join(ContentFolder, ContentSpvFolder, id))
}

func AsConfigPath(path string) ConfigPath {
	return ConfigPath(filepath.ToSlash(path))
}

func (p ContentPath) String() string { return string(p) }
func (p ConfigPath) String() string  { return string(p) }

func (p ContentPath) ToConfigPath() ConfigPath {
	return ConfigPath(strings.Replace(string(p), ContentFolder, ContentConfigFolder, 1) + configFileExt)
}

func (p ConfigPath) ToContentPath() ContentPath {
	s := strings.Replace(string(p), ContentConfigFolder, ContentFolder, 1)
	s = strings.TrimSuffix(s, configFileExt)
	return ContentPath(s)
}
