/******************************************************************************/
/* render_view_mode_gpu.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "log/slog"

const renderViewWireframeShaderKey = "render-view-wireframe"

func (g *GPUDevice) materialForRenderView(view RenderViewFrame, material *Material) *Material {
	if material == nil {
		return nil
	}
	selection := ResolveRenderViewModeSelectionForMode(view.ViewMode(), material, g.PhysicalDevice.Features)
	selected := selection.Material
	if selected == nil {
		selected = material
	}
	switch selection.PipelineOverride {
	case RenderViewPipelineOverrideWireframe:
		shader := g.shaderForRenderViewPipelineOverride(selected.Shader,
			selected.pipelineInfo, selection.PipelineOverride)
		if shader != nil {
			return selected.materialWithViewModeShader(selection.Mode, shader)
		}
	}
	return selected
}

func (g *GPUDevice) shaderForRenderViewPipelineOverride(base *Shader, pipelineInfo ShaderPipelineDataCompiled, override RenderViewPipelineOverride) *Shader {
	if base == nil || !base.RenderId.IsValid() || g == nil || g.Painter.caches == nil {
		return nil
	}
	key := renderViewPipelineOverrideKey(override)
	if key == "" {
		return nil
	}
	if base.subShaders == nil {
		base.subShaders = make(map[string]*Shader)
	}
	if shader := base.SubShader(key); shader != nil {
		if shader.RenderId.IsValid() {
			return shader
		}
		base.RemoveSubShader(key)
	}
	pipelineInfo = applyRenderViewPipelineOverride(pipelineInfo, override)
	shaderData := base.data
	shaderData.Name = base.ShaderDataName() + "." + key
	shader := NewShader(shaderData)
	shader.pipelineInfo = &pipelineInfo
	shader.renderPass = base.renderPass
	if err := g.CreateShader(shader, g.Painter.caches.AssetDatabase()); err != nil {
		slog.Error("failed to create render view pipeline shader override",
			"shader", base.ShaderDataName(), "override", key, "error", err)
		return nil
	}
	base.subShaders[key] = shader
	return shader
}

func renderViewPipelineOverrideKey(override RenderViewPipelineOverride) string {
	switch override {
	case RenderViewPipelineOverrideWireframe:
		return renderViewWireframeShaderKey
	default:
		return ""
	}
}

func applyRenderViewPipelineOverride(pipeline ShaderPipelineDataCompiled, override RenderViewPipelineOverride) ShaderPipelineDataCompiled {
	switch override {
	case RenderViewPipelineOverrideWireframe:
		pipeline.Name += ".wireframe"
		pipeline.Rasterization.PolygonMode = GPUPolygonModeLine
		pipeline.Rasterization.CullMode = GPUCullModeNone
		if pipeline.Rasterization.LineWidth <= 0 {
			pipeline.Rasterization.LineWidth = 1
		}
	}
	return pipeline
}
