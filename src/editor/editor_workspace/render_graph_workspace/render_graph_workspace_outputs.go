/******************************************************************************/
/* render_graph_workspace_outputs.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/glsl"
)

const (
	renderGraphGeneratedSourcePrefix         = "render_graph_"
	renderGraphGeneratedVertexSourceSuffix   = ".vert"
	renderGraphGeneratedFragmentSourceSuffix = ".frag"
)

var runRenderGraphGLSLC = func(input, output, flags string) error {
	args := []string{input, "-o", output}
	flags = strings.TrimSpace(flags)
	if flags != "" {
		args = append(args, strings.Fields(flags)...)
	}
	cmd := exec.Command("glslc", args...)
	if errStr, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errors.New(string(errStr))
		}
		return err
	}
	return nil
}

func (w *RenderGraphWorkspace) generateRenderGraphOutputs() error {
	defer tracing.NewRegion("RenderGraphWorkspace.generateRenderGraphOutputs").End()
	if w.ed == nil || w.Host == nil {
		return fmt.Errorf("render graph workspace is not initialized")
	}
	if strings.TrimSpace(w.currentGraphID) == "" {
		return fmt.Errorf("render graph must be saved before generating outputs")
	}
	pfs := w.ed.ProjectFileSystem()
	cache := w.ed.Cache()
	if pfs == nil || cache == nil {
		return fmt.Errorf("project file system or content cache is unavailable")
	}
	document := w.GraphDocument()
	compiled, err := compileRenderGraphDocumentOutput(document)
	if err != nil {
		return err
	}
	if err = ensureRenderGraphShaderInclude(pfs); err != nil {
		return err
	}
	vertexSourcePath := w.generated.VertexSourcePath
	if strings.TrimSpace(vertexSourcePath) == "" {
		vertexSourcePath = renderGraphGeneratedVertexSourcePath(w.currentGraphID)
	}
	stagedVertex, err := stageRenderGraphSource(pfs, vertexSourcePath, compiled.VertexSource)
	if err != nil {
		return err
	}
	defer stagedVertex.cleanup(pfs)

	vertexLayout, err := parseRenderGraphShaderLayout(pfs, stagedVertex.sourcePath)
	if err != nil {
		return err
	}
	vertexSpvBytes, err := compileRenderGraphShaderToSPV(pfs, stagedVertex.sourcePath)
	if err != nil {
		return err
	}

	fragmentSourcePath := w.generated.FragmentSourcePath
	if strings.TrimSpace(fragmentSourcePath) == "" {
		fragmentSourcePath = renderGraphGeneratedFragmentSourcePath(w.currentGraphID)
	}
	stagedFragment, err := stageRenderGraphSource(pfs, fragmentSourcePath, compiled.FragmentSource)
	if err != nil {
		return err
	}
	defer stagedFragment.cleanup(pfs)

	fragmentLayout, err := parseRenderGraphShaderLayout(pfs, stagedFragment.sourcePath)
	if err != nil {
		return err
	}
	fragmentSpvBytes, err := compileRenderGraphShaderToSPV(pfs, stagedFragment.sourcePath)
	if err != nil {
		return err
	}
	shaderData, err := buildRenderGraphShaderData(pfs, renderGraphGeneratedShaderName(w.currentGraphID),
		vertexSourcePath, w.generated.VertexSpvID, vertexLayout,
		fragmentSourcePath, w.generated.FragmentSpvID, fragmentLayout, compiled.SamplerLabels)
	if err != nil {
		return err
	}
	materialData := renderGraphGeneratedMaterialData(w.generated.ShaderID, compiled.Textures)

	generated := w.generated
	generated.VertexSourcePath = vertexSourcePath
	if err = pfs.WriteFile(vertexSourcePath, []byte(compiled.VertexSource), os.ModePerm); err != nil {
		return err
	}
	generated.VertexSpvID, err = w.upsertRawContent(generated.VertexSpvID,
		w.currentName+" Vertex SPV", vertexSpvBytes, content_database.Spv{})
	if err != nil {
		return err
	}
	shaderData.VertexSpv = generated.VertexSpvID
	generated.FragmentSourcePath = fragmentSourcePath
	if err = pfs.WriteFile(fragmentSourcePath, []byte(compiled.FragmentSource), os.ModePerm); err != nil {
		return err
	}
	generated.FragmentSpvID, err = w.upsertRawContent(generated.FragmentSpvID,
		w.currentName+" Fragment SPV", fragmentSpvBytes, content_database.Spv{})
	if err != nil {
		return err
	}
	shaderData.FragmentSpv = generated.FragmentSpvID
	shaderBytes, err := json.Marshal(shaderData)
	if err != nil {
		return err
	}
	generated.ShaderID, err = w.upsertRawContent(generated.ShaderID,
		w.currentName+" Shader", shaderBytes, content_database.Shader{})
	if err != nil {
		return err
	}
	materialData.Shader = generated.ShaderID
	materialBytes, err := json.Marshal(materialData)
	if err != nil {
		return err
	}
	generated.MaterialID, err = w.upsertRawContent(generated.MaterialID,
		w.currentName+" Material", materialBytes, content_database.Material{})
	if err != nil {
		return err
	}

	w.generated = generated
	if err = w.persistRenderGraphDocument(); err != nil {
		return err
	}
	w.Host.ShaderCache().ReloadShader(shaderData.Compile())
	if _, found := w.Host.MaterialCache().FindMaterial(generated.MaterialID); found {
		if err = w.Host.MaterialCache().ReplaceMaterial(generated.MaterialID); err != nil {
			return err
		}
	}
	return nil
}

func (w *RenderGraphWorkspace) upsertRawContent(id, name string, data []byte, cat content_database.ContentCategory) (string, error) {
	pfs := w.ed.ProjectFileSystem()
	cache := w.ed.Cache()
	if strings.TrimSpace(id) != "" {
		cc, err := cache.Read(id)
		if err == nil && cc.Config.Type == cat.TypeName() {
			if cc.Config.Name != name {
				if _, err = cache.Rename(id, name, pfs); err != nil {
					return "", err
				}
				w.ed.Events().OnContentRenamed.Execute(id)
			}
			if err = pfs.WriteFile(cc.ContentPath(), data, os.ModePerm); err != nil {
				return "", err
			}
			w.ed.Events().OnContentChangesSaved.Execute(id)
			return id, nil
		}
	}
	ids := content_database.ImportRaw(name, data, cat, pfs, cache)
	if len(ids) == 0 {
		return "", fmt.Errorf("failed to import generated %s content", cat.TypeName())
	}
	w.ed.Events().OnContentAdded.Execute(ids)
	for i := range ids {
		w.ed.Events().OnContentChangesSaved.Execute(ids[i])
	}
	return ids[0], nil
}

func (w *RenderGraphWorkspace) persistRenderGraphDocument() error {
	if strings.TrimSpace(w.currentGraphID) == "" {
		return fmt.Errorf("render graph id is empty")
	}
	data, err := w.SerializeGraph()
	if err != nil {
		return err
	}
	pfs := w.ed.ProjectFileSystem()
	cache := w.ed.Cache()
	cc, err := cache.Read(w.currentGraphID)
	if err != nil {
		return err
	}
	if err = pfs.WriteFile(cc.ContentPath(), data, os.ModePerm); err != nil {
		return err
	}
	w.ed.Events().OnContentChangesSaved.Execute(w.currentGraphID)
	return nil
}

type stagedRenderGraphFragment struct {
	sourcePath string
}

func stageRenderGraphSource(pfs *project_file_system.FileSystem, finalPath, source string) (stagedRenderGraphFragment, error) {
	dir := filepath.ToSlash(filepath.Dir(finalPath))
	tempPath := filepath.ToSlash(filepath.Join(dir, "."+filepath.Base(finalPath)+".tmp"+filepath.Ext(finalPath)))
	if err := pfs.MkdirAll(dir, os.ModePerm); err != nil {
		return stagedRenderGraphFragment{}, err
	}
	if err := pfs.WriteFile(tempPath, []byte(source), os.ModePerm); err != nil {
		return stagedRenderGraphFragment{}, err
	}
	return stagedRenderGraphFragment{sourcePath: tempPath}, nil
}

func (s stagedRenderGraphFragment) cleanup(pfs *project_file_system.FileSystem) {
	if strings.TrimSpace(s.sourcePath) == "" {
		return
	}
	if err := pfs.Remove(s.sourcePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("failed to remove staged render graph fragment", "path", s.sourcePath, "error", err)
	}
}

func ensureRenderGraphShaderInclude(pfs *project_file_system.FileSystem) error {
	includePath := filepath.ToSlash(filepath.Join(project_file_system.SrcShaderFolder, "kaiju.glsl"))
	if pfs.FileExists(includePath) {
		return nil
	}
	data, err := project_file_system.EngineFS.ReadFile(
		"editor/editor_embedded_content/editor_content/renderer/src/kaiju.glsl")
	if err != nil {
		return err
	}
	if err = pfs.MkdirAll(project_file_system.SrcShaderFolder, os.ModePerm); err != nil {
		return err
	}
	return pfs.WriteFile(includePath, data, os.ModePerm)
}

func parseRenderGraphShaderLayout(pfs *project_file_system.FileSystem, sourcePath string) (rendering.ShaderLayoutGroup, error) {
	source, err := glsl.Parse(pfs.FullPath(sourcePath), "")
	if err != nil {
		return rendering.ShaderLayoutGroup{}, err
	}
	return rendering.ShaderLayoutGroup{
		Type:       source.Type(),
		WorkGroups: source.WorkGroups,
		Layouts:    source.Layouts,
	}, nil
}

func compileRenderGraphShaderToSPV(pfs *project_file_system.FileSystem, sourcePath string) ([]byte, error) {
	temp, err := os.CreateTemp(os.TempDir(), "kaiju-render-graph-*"+filepath.Ext(sourcePath)+".spv")
	if err != nil {
		return nil, err
	}
	tempPath := temp.Name()
	if err = temp.Close(); err != nil {
		return nil, err
	}
	defer os.Remove(tempPath)
	if err = runRenderGraphGLSLC(pfs.FullPath(sourcePath), tempPath, ""); err != nil {
		return nil, err
	}
	return os.ReadFile(tempPath)
}

func buildRenderGraphShaderData(pfs *project_file_system.FileSystem, shaderName, vertexSource, vertexSpv string,
	vertexLayout rendering.ShaderLayoutGroup, fragmentSource, fragmentSpv string,
	fragmentLayout rendering.ShaderLayoutGroup, samplerLabels []string) (rendering.ShaderData, error) {
	stock, err := stockPBRShaderData(pfs)
	if err != nil {
		return rendering.ShaderData{}, err
	}
	if len(samplerLabels) == 0 {
		samplerLabels = renderGraphSamplerLabels(renderGraphDefaultTextureSlots())
	}
	hasStockVertex := false
	for i := range stock.LayoutGroups {
		if stock.LayoutGroups[i].Type == "Vertex" {
			hasStockVertex = true
			break
		}
	}
	if !hasStockVertex {
		return rendering.ShaderData{}, fmt.Errorf("stock pbr shader is missing vertex layout")
	}
	layouts := []rendering.ShaderLayoutGroup{vertexLayout}
	layouts = append(layouts, fragmentLayout)
	return rendering.ShaderData{
		Name:             shaderName,
		DrawInstanceData: "pbr",
		Vertex:           filepath.ToSlash(vertexSource),
		Fragment:         filepath.ToSlash(fragmentSource),
		LayoutGroups:     layouts,
		SamplerLabels:    append([]string(nil), samplerLabels...),
		VertexSpv:        vertexSpv,
		FragmentSpv:      fragmentSpv,
	}, nil
}

func stockPBRShaderData(pfs *project_file_system.FileSystem) (rendering.ShaderData, error) {
	data, err := pfs.ReadFile(filepath.ToSlash(filepath.Join(project_file_system.StockFolder, "pbr.shader")))
	if err != nil {
		return rendering.ShaderData{}, err
	}
	shader := rendering.ShaderData{}
	if err = json.Unmarshal(data, &shader); err != nil {
		return rendering.ShaderData{}, err
	}
	return shader, nil
}

func renderGraphGeneratedMaterialData(shaderID string, textures []rendering.MaterialTextureData) rendering.MaterialData {
	if len(textures) == 0 {
		textures = renderGraphDefaultTextureSlots()
	}
	return rendering.MaterialData{
		Shader:          shaderID,
		RenderPass:      "opaque.renderpass",
		ShaderPipeline:  "basic.shaderpipeline",
		Textures:        append([]rendering.MaterialTextureData(nil), textures...),
		IsLit:           true,
		ReceivesShadows: true,
		CastsShadows:    true,
	}
}

func renderGraphGeneratedFragmentSourcePath(graphID string) string {
	return filepath.ToSlash(filepath.Join(project_file_system.SrcShaderFolder,
		renderGraphGeneratedSourcePrefix+renderGraphGeneratedAssetKey(graphID)+renderGraphGeneratedFragmentSourceSuffix))
}

func renderGraphGeneratedVertexSourcePath(graphID string) string {
	return filepath.ToSlash(filepath.Join(project_file_system.SrcShaderFolder,
		renderGraphGeneratedSourcePrefix+renderGraphGeneratedAssetKey(graphID)+renderGraphGeneratedVertexSourceSuffix))
}

func renderGraphGeneratedShaderName(graphID string) string {
	return renderGraphGeneratedSourcePrefix + renderGraphGeneratedAssetKey(graphID)
}

func renderGraphGeneratedAssetKey(graphID string) string {
	key := strings.TrimSpace(graphID)
	key = strings.TrimSuffix(key, ".rendergraph")
	key = strings.TrimSuffix(key, filepath.Ext(key))
	key = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, key)
	key = strings.Trim(key, "_")
	if key == "" {
		return "unsaved"
	}
	return key
}
