/******************************************************************************/
/* material_shadow_bindings_test.go                                           */
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

package editor_embedded_content

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kaijuengine.com/rendering"
)

func TestShadowReceivingMaterialsDeclareShadowBindings(t *testing.T) {
	root := filepath.FromSlash("editor_content/renderer")
	materialsDir := filepath.Join(root, "materials")
	shadersDir := filepath.Join(root, "shaders")
	entries, err := os.ReadDir(materialsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".material" {
			continue
		}
		materialPath := filepath.Join(materialsDir, entry.Name())
		var material rendering.MaterialData
		readJSON(t, materialPath, &material)
		if !material.ReceivesShadows {
			continue
		}
		var shader rendering.ShaderData
		readJSON(t, filepath.Join(shadersDir, material.Shader), &shader)
		compiled := shader.Compile()
		if !hasDescriptorBinding(compiled, 2, "sampler2D", rendering.MaxLocalLights) {
			t.Fatalf("%s receives shadows but %s does not declare sampler2D shadowMap binding 2",
				entry.Name(), material.Shader)
		}
		if !hasDescriptorBinding(compiled, 3, "samplerCube", rendering.MaxLocalLights) {
			t.Fatalf("%s receives shadows but %s does not declare samplerCube shadowCubeMap binding 3",
				entry.Name(), material.Shader)
		}
	}
}

func TestMaterialPipelinesMatchRenderPassColorAttachments(t *testing.T) {
	root := filepath.FromSlash("editor_content/renderer")
	materialsDir := filepath.Join(root, "materials")
	passesDir := filepath.Join(root, "passes")
	pipelinesDir := filepath.Join(root, "pipelines")
	entries, err := os.ReadDir(materialsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".material" {
			continue
		}
		materialPath := filepath.Join(materialsDir, entry.Name())
		var material rendering.MaterialData
		readJSON(t, materialPath, &material)
		if material.RenderPass == "" || material.ShaderPipeline == "" {
			continue
		}
		var renderPass rendering.RenderPassData
		readJSON(t, filepath.Join(passesDir, material.RenderPass), &renderPass)
		var pipeline rendering.ShaderPipelineData
		readJSON(t, filepath.Join(pipelinesDir, material.ShaderPipeline), &pipeline)
		subpass := int(pipeline.GraphicsPipeline.Subpass)
		if subpass < 0 || subpass >= len(renderPass.SubpassDescriptions) {
			t.Fatalf("%s uses %s subpass %d, but %s has %d subpasses",
				entry.Name(), material.ShaderPipeline, subpass, material.RenderPass, len(renderPass.SubpassDescriptions))
		}
		want := len(renderPass.SubpassDescriptions[subpass].ColorAttachmentReferences)
		got := len(pipeline.ColorBlendAttachments)
		if got != want {
			t.Fatalf("%s uses %s with %d color blend attachments, but %s subpass %d has %d color attachments",
				entry.Name(), material.ShaderPipeline, got, material.RenderPass, subpass, want)
		}
	}
}

func readJSON(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

func hasDescriptorBinding(shader rendering.ShaderDataCompiled, binding int, shaderType string, count int) bool {
	for _, group := range shader.LayoutGroups {
		for _, layout := range group.Layouts {
			if layout.Binding == binding && layout.Type == shaderType && layout.Count >= count {
				return true
			}
		}
	}
	return false
}
