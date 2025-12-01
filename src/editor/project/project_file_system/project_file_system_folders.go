/******************************************************************************/
/* project_file_system_folders.go                                             */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package project_file_system

import "path/filepath"

const (
	DatabaseFolder      = "database"
	ContentFolder       = "database/content"
	ContentConfigFolder = "database/config"
	SrcFolder           = "database/src"
	StockFolder         = "database/stock"
	ProjectConfigFile   = "database/project.json"
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
	ContentMaterialFolder        = ContentRenderFolder + "/material"
	ContentRenderPassFolder      = ContentRenderFolder + "/renderpass"
	ContentShaderFolder          = ContentRenderFolder + "/shader"
	ContentShaderPipelineFolder  = ContentRenderFolder + "/pipeline"
	ContentSpvFolder             = ContentRenderFolder + "/spv"
	ContentStageFolder           = "stage"
	ContentTemplateFolder        = "template"
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
	ProjectCodeGameTitle      = KaijuSrcFolder + "/build/title.go"
	EntityDataBindingInit     = ProjectCodeFolder + "/entity_data_binding_init.go"
)

// ContentFolderPath returns the full filesystem path to a child entry inside
// the project's content folder. It joins the base `ContentFolder` constant with
// the supplied relative `child` path using the OS‑specific separator.
//
// Parameters:
//   child - a relative path (file or sub‑directory) within the content folder.
//
// Returns:
//   A string containing the absolute path to the specified child.
func ContentFolderPath(child string) string {
	return filepath.Join(ContentFolder, child)
}

func HtmlPath(id string) string {
	return filepath.Join(ContentFolder, ContentHtmlFolder, id)
}
