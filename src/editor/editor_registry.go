/******************************************************************************/
/* editor_registry.go                                                         */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package editor

import (
	"kaiju/engine/assets/asset_importer"
	"kaiju/editor/content/content_opener"
)

func registerAssetImporters(ed *Editor) {
	ed.assetImporters.Register(asset_importer.ObjImporter{})
	ed.assetImporters.Register(asset_importer.GlbImporter{})
	ed.assetImporters.Register(asset_importer.GltfImporter{})
	ed.assetImporters.Register(asset_importer.PngImporter{})
	ed.assetImporters.Register(asset_importer.StageImporter{})
	ed.assetImporters.Register(asset_importer.HtmlImporter{})
	ed.assetImporters.Register(asset_importer.ShaderImporter{})
	ed.assetImporters.Register(asset_importer.RenderPassImporter{})
	ed.assetImporters.Register(asset_importer.ShaderPipelineImporter{})
	ed.assetImporters.Register(asset_importer.MaterialImporter{})
}

func registerContentOpeners(ed *Editor) {
	ed.contentOpener.Register(content_opener.ObjOpener{})
	ed.contentOpener.Register(content_opener.GlbOpener{})
	ed.contentOpener.Register(content_opener.GltfOpener{})
	ed.contentOpener.Register(content_opener.StageOpener{})
	ed.contentOpener.Register(content_opener.HTMLOpener{})
	ed.contentOpener.Register(content_opener.ImageOpener{})
	ed.contentOpener.Register(content_opener.ShaderOpener{})
	ed.contentOpener.Register(content_opener.RenderPassOpener{})
	ed.contentOpener.Register(content_opener.ShaderPipelineOpener{})
	ed.contentOpener.Register(content_opener.MaterialOpener{})
}
