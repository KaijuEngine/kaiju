---
title: Render targets and render views | Kaiju Engine
---

# Render targets and render views

Render targets let the engine render a view into a texture. Render views define
which camera, layer mask, target, sort order, and debug view mode should be used
for a render pass. The most common use is an editor or game viewport that renders
3D content into a texture and then displays that texture in the UI.

## Create a target and view

```go
target, err := host.RenderTargets.Create(rendering.RenderTargetOptions{
	Name:   "stage-main",
	Width:  640,
	Height: 360,
	Depth:  true,
})
if err != nil {
	panic(err)
}

view, err := host.RenderViews.Create(rendering.RenderViewOptions{
	Name:      "stage-main",
	Target:    target,
	Camera:    host.PrimaryCamera(),
	LayerMask: rendering.RenderLayerWorld | rendering.RenderLayerEditor,
	Clear:     true,
	Sort:      -100,
	ViewMode:  rendering.RenderViewModeNormal,
})
if err != nil {
	panic(err)
}
_ = view
```

Target views are rendered before the default swapchain view. The target's color
output becomes available as `RenderTargetOutputColor` after the render thread has
realized the target.

## Display the target in UI

```go
uiMan := ui.Manager{}
uiMan.Init(host)

preview := uiMan.Add().ToImage()
preview.Init(placeholderTexture)
preview.Base().Layout().SetPositioning(ui.PositioningAbsolute)
preview.Base().Layout().SetOffset(24, 24)
preview.Base().Layout().Scale(640, 360)

host.RunAfterFrames(2, func() {
	tex, err := target.Texture(rendering.RenderTargetOutputColor)
	if err != nil {
		panic(err)
	}
	preview.SetTexture(tex)
})
```

## View modes

`RenderViewOptions.ViewMode` supports:

- `RenderViewModeNormal`: normal material rendering.
- `RenderViewModeWireframe`: uses a compatible material override when one is
  configured; otherwise uses a same-shader wireframe pipeline variant when the
  Vulkan device supports non-solid fill.
- `RenderViewModeUnlit`: uses a compatible material override when configured.
- `RenderViewModeProfile`: uses a profile override when configured, then falls
  back to a compatible unlit override.

Material overrides must keep the same instance data and descriptor contract as
the original material. This keeps existing draw instances, textures, and per-view
buffers valid while allowing terrain or other special materials to swap to debug
or unlit shader variants.

```go
terrain, _ := host.MaterialCache().Material(assets.MaterialDefinitionTerrain)
terrainUnlit, _ := host.MaterialCache().Material(assets.MaterialDefinitionTerrainUnlit)
terrain.SetRenderViewModeOverride(rendering.RenderViewModeUnlit, terrainUnlit)
terrain.SetRenderViewModeOverride(rendering.RenderViewModeProfile, terrainUnlit)
```

## Cleanup

Destroy views and targets through their managers. Destruction is queued so GPU
resources are released safely on the render thread.

```go
_ = host.RenderViews.Destroy("stage-main")
_ = host.RenderTargets.Destroy("stage-main")
```

Destroying a render view releases per-view uniform buffers and draw-instance view
state. Destroying a render target releases the target textures after pending GPU
work is processed.

## Editor smoke-test checklist

- Open the Stage workspace and verify the default viewport still renders the
  stage through the normal view.
- Create a second viewport or run `kaijuengine.com.exe -integrationtest=render-view-modes`.
- Confirm the normal and wireframe previews show the same scene from the same
  camera, with the wireframe preview visibly rendered as lines.
- Resize the editor window and confirm target-backed viewports keep drawing after
  the next frame.
- Switch away from the Stage workspace and back, then confirm destroyed views or
  targets do not log repeated render-target realization errors.
- Verify UI overlays still render on the swapchain view and are not captured into
  world-only render target views unless the UI layer is included.
