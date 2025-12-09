/******************************************************************************/
/* content_database_shader_pipeline.go                                        */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package content_database

import (
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
)

func init() { addCategory(ShaderPipeline{}) }

// ShaderPipeline is a [ContentCategory] represented by a file with a ".ShaderPipeline"
// extension. A ShaderPipeline is a conglomeration of a specific render pass, a
// specific ShaderPipeline pipeline, and a set of specific ShaderPipelines.
type ShaderPipeline struct{}
type ShaderPipelineConfig struct{}

// See the documentation for the interface [ContentCategory] to learn more about
// the following functions

func (ShaderPipeline) Path() string       { return project_file_system.ContentShaderPipelineFolder }
func (ShaderPipeline) TypeName() string   { return "ShaderPipeline" }
func (ShaderPipeline) ExtNames() []string { return []string{".shaderpipeline"} }

func (ShaderPipeline) Import(src string, _ *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("ShaderPipeline.Import").End()
	return pathToTextData(src)
}

func (c ShaderPipeline) Reimport(id string, cache *Cache, fs *project_file_system.FileSystem) (ProcessedImport, error) {
	defer tracing.NewRegion("ShaderPipeline.Reimport").End()
	return reimportByNameMatching(c, id, cache, fs)
}

func (ShaderPipeline) PostImportProcessing(proc ProcessedImport, res *ImportResult, fs *project_file_system.FileSystem, cache *Cache, linkedId string) error {
	return nil
}
