/******************************************************************************/
/* material_view_mode.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"reflect"
	"weak"
)

func (m *Material) SetRenderViewModeOverride(mode RenderViewMode, override *Material) {
	if m == nil || override == nil || mode == RenderViewModeNormal {
		return
	}
	root := m.SelectRoot()
	root.mutex.Lock()
	defer root.mutex.Unlock()
	if root.ViewModeOverrides == nil {
		root.ViewModeOverrides = make(map[RenderViewMode]*Material)
	}
	root.ViewModeOverrides[mode] = override.SelectRoot()
}

func (m *Material) RenderViewModeOverride(mode RenderViewMode) *Material {
	if m == nil {
		return nil
	}
	root := m.SelectRoot()
	root.mutex.Lock()
	defer root.mutex.Unlock()
	if root.ViewModeOverrides == nil {
		return nil
	}
	return root.ViewModeOverrides[mode]
}

func (m *Material) compatibleRenderViewModeOverride(mode RenderViewMode) *Material {
	override := m.RenderViewModeOverride(mode)
	if override == nil || !m.RenderViewModeOverrideCompatible(override) {
		return nil
	}
	return override
}

func (m *Material) RenderViewModeOverrideCompatible(override *Material) bool {
	if m == nil || override == nil {
		return false
	}
	return m.shaderInfo.RenderViewModeCompatible(override.shaderInfo)
}

func (m *Material) materialWithViewModeShader(mode RenderViewMode, shader *Shader) *Material {
	if m == nil || shader == nil {
		return m
	}
	root := m.SelectRoot()
	root.mutex.Lock()
	defer root.mutex.Unlock()
	if root.viewModePipelineMaterials == nil {
		root.viewModePipelineMaterials = make(map[RenderViewMode]*Material)
	}
	if material, ok := root.viewModePipelineMaterials[mode]; ok {
		return material
	}
	material := &Material{
		Id:                root.Id + "#" + mode.String(),
		shaderInfo:        root.shaderInfo,
		renderPass:        root.renderPass,
		pipelineInfo:      root.pipelineInfo,
		Shader:            shader,
		Textures:          root.Textures,
		ViewModeOverrides: root.ViewModeOverrides,
		Root:              weak.Pointer[Material]{},
		PrepassMaterial:   root.PrepassMaterial,
		IsLit:             root.IsLit,
		ReceivesShadows:   root.ReceivesShadows,
		CastsShadows:      root.CastsShadows,
	}
	root.viewModePipelineMaterials[mode] = material
	return material
}

func (s ShaderDataCompiled) RenderViewModeCompatible(other ShaderDataCompiled) bool {
	return s.Stride() == other.Stride() &&
		reflect.DeepEqual(s.ToDescriptorSetLayoutStructure(), other.ToDescriptorSetLayoutStructure())
}
