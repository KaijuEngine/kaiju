/******************************************************************************/
/* material_shadow_bindings_test.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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
