/******************************************************************************/
/* render_view_mode_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "testing"

func TestRenderViewModeSelectionUsesWireframePipelineWhenSupported(t *testing.T) {
	view := newRenderView(RenderViewOptions{
		Name:     "wire",
		ViewMode: RenderViewModeWireframe,
	}, 0)
	material := &Material{}
	selection := ResolveRenderViewModeSelection(view, material, GPUPhysicalDeviceFeatures{
		FillModeNonSolid: true,
	})
	if selection.Material != material {
		t.Fatalf("selection material = %#v, want base material", selection.Material)
	}
	if selection.PipelineOverride != RenderViewPipelineOverrideWireframe {
		t.Fatalf("pipeline override = %v, want wireframe", selection.PipelineOverride)
	}
}

func TestRenderViewModeSelectionFallsBackWhenWireframeUnsupported(t *testing.T) {
	view := newRenderView(RenderViewOptions{
		Name:     "wire",
		ViewMode: RenderViewModeWireframe,
	}, 0)
	material := &Material{}
	selection := ResolveRenderViewModeSelection(view, material, GPUPhysicalDeviceFeatures{})
	if selection.Material != material {
		t.Fatalf("selection material = %#v, want base material", selection.Material)
	}
	if selection.PipelineOverride != RenderViewPipelineOverrideNone {
		t.Fatalf("pipeline override = %v, want none", selection.PipelineOverride)
	}
}

func TestRenderViewModeSelectionUsesCompatibleMaterialOverride(t *testing.T) {
	view := newRenderView(RenderViewOptions{
		Name:     "unlit",
		ViewMode: RenderViewModeUnlit,
	}, 0)
	base := &Material{
		shaderInfo: compatibleViewModeShaderInfo(),
	}
	override := &Material{
		shaderInfo: compatibleViewModeShaderInfo(),
	}
	base.SetRenderViewModeOverride(RenderViewModeUnlit, override)
	selection := ResolveRenderViewModeSelection(view, base, GPUPhysicalDeviceFeatures{})
	if selection.Material != override {
		t.Fatalf("selection material = %#v, want override", selection.Material)
	}
	if selection.PipelineOverride != RenderViewPipelineOverrideNone {
		t.Fatalf("pipeline override = %v, want none", selection.PipelineOverride)
	}
}

func TestRenderViewModeSelectionRejectsIncompatibleMaterialOverride(t *testing.T) {
	view := newRenderView(RenderViewOptions{
		Name:     "unlit",
		ViewMode: RenderViewModeUnlit,
	}, 0)
	base := &Material{
		shaderInfo: compatibleViewModeShaderInfo(),
	}
	override := &Material{
		shaderInfo: ShaderDataCompiled{},
	}
	base.SetRenderViewModeOverride(RenderViewModeUnlit, override)
	selection := ResolveRenderViewModeSelection(view, base, GPUPhysicalDeviceFeatures{})
	if selection.Material != base {
		t.Fatalf("selection material = %#v, want base fallback", selection.Material)
	}
}

func TestRenderViewModeSelectionProfileFallsBackToUnlitOverride(t *testing.T) {
	view := newRenderView(RenderViewOptions{
		Name:     "profile",
		ViewMode: RenderViewModeProfile,
	}, 0)
	base := &Material{
		shaderInfo: compatibleViewModeShaderInfo(),
	}
	override := &Material{
		shaderInfo: compatibleViewModeShaderInfo(),
	}
	base.SetRenderViewModeOverride(RenderViewModeUnlit, override)
	selection := ResolveRenderViewModeSelection(view, base, GPUPhysicalDeviceFeatures{})
	if selection.Material != override {
		t.Fatalf("selection material = %#v, want unlit override", selection.Material)
	}
}

func TestParseRenderViewMode(t *testing.T) {
	tests := map[string]RenderViewMode{
		"":              RenderViewModeNormal,
		"default":       RenderViewModeNormal,
		"wireframe":     RenderViewModeWireframe,
		"Unlit":         RenderViewModeUnlit,
		"profile-style": RenderViewModeProfile,
	}
	for value, want := range tests {
		got, ok := ParseRenderViewMode(value)
		if !ok || got != want {
			t.Fatalf("ParseRenderViewMode(%q) = %v/%v, want %v/true", value, got, ok, want)
		}
	}
	if _, ok := ParseRenderViewMode("bad"); ok {
		t.Fatalf("ParseRenderViewMode accepted invalid mode")
	}
}

func compatibleViewModeShaderInfo() ShaderDataCompiled {
	return ShaderDataCompiled{
		LayoutGroups: []ShaderLayoutGroup{{
			Type: "Vertex",
			Layouts: []ShaderLayout{
				{Source: "in", Location: 8, Type: "mat4", Name: "model"},
				{Binding: 0, Set: 0, Source: "uniform", Count: 1, Type: "UniformBufferObject"},
			},
		}},
	}
}
