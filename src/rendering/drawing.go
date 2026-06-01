/******************************************************************************/
/* drawing.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"log/slog"
	"sort"
	"sync"

	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

// Drawing represents a renderable entity in the engine. It bundles together
// the material, mesh, shader instance data, transform, sorting order, and an
// optional view culler used during rendering.
type Drawing struct {
	// Material defines the visual appearance and render pass for the drawing.
	Material *Material
	// Mesh contains the geometry to be rendered.
	Mesh *Mesh
	// ShaderData holds per‑instance data for the shader (e.g., uniforms).
	ShaderData DrawInstance
	// Transform specifies the transform this drawing follows.
	Transform *matrix.Transform
	// Sort determines the draw order within a render pass.
	Sort int
	// Layer controls which render views include this drawing. Zero maps to world.
	Layer RenderLayerMask
	// ViewCuller optionally culls the drawing based on the view frustum.
	ViewCuller ViewCuller
}

// IsValid reports whether the Drawing is properly configured for rendering.
// A Drawing is considered valid if it has a non-nil Material. This check is
// used before submitting the drawing to the render pipeline.
func (d *Drawing) IsValid() bool {
	return d.Material != nil
}

func (d *Drawing) EffectiveLayer() RenderLayerMask {
	return normalizeRenderLayerMask(d.Layer)
}

func (d *Drawing) MatchesLayer(mask RenderLayerMask) bool {
	return d.EffectiveLayer()&mask != 0
}

type RenderPassGroup struct {
	renderPass *RenderPass
	draws      []ShaderDraw
}

func (d *RenderPassGroup) MatchesLayer(mask RenderLayerMask) bool {
	for i := range d.draws {
		for j := range d.draws[i].instanceGroups {
			if d.draws[i].instanceGroups[j].MatchesLayer(mask) {
				return true
			}
		}
	}
	return false
}

type Drawings struct {
	renderPassGroups []RenderPassGroup
	backDraws        []Drawing
	mutex            sync.RWMutex
}

func NewDrawings() Drawings {
	return Drawings{
		renderPassGroups: make([]RenderPassGroup, 0),
		backDraws:        make([]Drawing, 0),
		mutex:            sync.RWMutex{},
	}
}

func (d *Drawings) HasDrawings() bool { return len(d.renderPassGroups) > 0 }

func (d *Drawings) matchGroup(sd *ShaderDraw, dg *Drawing) int {
	defer tracing.NewRegion("Drawings.matchGroup").End()
	idx := -1
	for i := 0; i < len(sd.instanceGroups) && idx < 0; i++ {
		g := &sd.instanceGroups[i]
		if g.Mesh == dg.Mesh &&
			g.EffectiveLayer() == dg.EffectiveLayer() &&
			(g.MaterialInstance == dg.Material || g.MaterialInstance.Root.Value() == dg.Material) {
			idx = i
		}
	}
	return idx
}

func (d *RenderPassGroup) findShaderDraw(material *Material) (*ShaderDraw, bool) {
	defer tracing.NewRegion("Drawings.RenderPassGroup").End()
	rootMat := material
	if rootMat.Root.Value() != nil {
		rootMat = rootMat.Root.Value()
	}
	for i := range d.draws {
		mat := d.draws[i].material
		if mat.Root.Value() != nil {
			mat = mat.Root.Value()
		}
		if mat == rootMat {
			return &d.draws[i], true
		}
	}
	return nil, false
}

func (d *Drawings) findRenderPassGroup(renderPass *RenderPass) (*RenderPassGroup, bool) {
	defer tracing.NewRegion("Drawings.findRenderPassGroup").End()
	for i := range d.renderPassGroups {
		if d.renderPassGroups[i].renderPass == renderPass {
			return &d.renderPassGroups[i], true
		}
	}
	return nil, false
}

func (d *Drawings) addToRenderPassGroup(drawing *Drawing, rpGroup *RenderPassGroup) {
	defer tracing.NewRegion("Drawings.addToRenderPassGroup").End()
	draw, ok := rpGroup.findShaderDraw(drawing.Material)
	if !ok {
		newDraw := NewShaderDraw(drawing.Material)
		rpGroup.draws = append(rpGroup.draws, newDraw)
		draw = &rpGroup.draws[len(rpGroup.draws)-1]
	}
	drawing.ShaderData.setTransform(drawing.Transform)
	idx := d.matchGroup(draw, drawing)
	if idx >= 0 && !draw.instanceGroups[idx].destroyed {
		draw.instanceGroups[idx].AddInstance(drawing.ShaderData)
	} else {
		group := NewDrawInstanceGroup(drawing.Mesh, drawing.ShaderData.Size(), drawing.ViewCuller)
		group.Layer = drawing.EffectiveLayer()
		group.MaterialInstance = drawing.Material
		group.AddInstance(drawing.ShaderData)
		group.MaterialInstance.Textures = drawing.Material.Textures
		group.sort = drawing.Sort
		if idx >= 0 {
			draw.instanceGroups[idx] = group
		} else {
			draw.AddInstanceGroup(group)
		}
	}
}

func (d *Drawings) PreparePending(shadowCascades uint8) {
	defer tracing.NewRegion("Drawings.PreparePending").End()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	for i := 0; i < len(d.backDraws); i++ {
		drawing := &d.backDraws[i]
		rpGroup, ok := d.findRenderPassGroup(drawing.Material.renderPass)
		if !ok {
			d.renderPassGroups = append(d.renderPassGroups, RenderPassGroup{
				renderPass: drawing.Material.renderPass,
			})
			rpGroup = &d.renderPassGroups[len(d.renderPassGroups)-1]
		}
		d.addToRenderPassGroup(drawing, rpGroup)
		if drawing.Material.CastsShadows {
			for i := range shadowCascades {
				d.backDraws = append(d.backDraws, lightTransformDrawingToDepth(drawing, i))
			}
		}
	}
	d.backDraws = klib.WipeSlice(d.backDraws)
}

func (d *Drawings) CaptureFrameData(lights LightsForRender, views []RenderViewFrame) {
	defer tracing.NewRegion("Drawings.CaptureFrameData").End()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	views = renderViewsForDraw(views)
	for i := range views {
		layerMask := views[i].LayerMask()
		for j := range d.renderPassGroups {
			for k := range d.renderPassGroups[j].draws {
				draw := &d.renderPassGroups[j].draws[k]
				for g := range draw.instanceGroups {
					if draw.instanceGroups[g].MatchesLayer(layerMask) {
						draw.instanceGroups[g].CaptureDataForView(lights, views[i])
					}
				}
			}
		}
	}
}

func (d *Drawings) AddDrawing(drawing Drawing) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	drawing.Layer = drawing.EffectiveLayer()
	if p := drawing.Material.PrepassMaterial.Value(); p != nil {
		cpy := drawing
		cpy.Material = p
		d.backDraws = append(d.backDraws, cpy)
	}
	d.backDraws = append(d.backDraws, drawing)
	if drawing.Mesh == nil || drawing.Material == nil {
		panic("no")
	}
}

func (d *Drawings) AddDrawings(drawings []Drawing) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	normalized := make([]Drawing, len(drawings))
	for i := range drawings {
		normalized[i] = drawings[i]
		normalized[i].Layer = normalized[i].EffectiveLayer()
		if p := normalized[i].Material.PrepassMaterial.Value(); p != nil {
			cpy := normalized[i]
			cpy.Material = p
			d.backDraws = append(d.backDraws, cpy)
		}
	}
	d.backDraws = append(d.backDraws, normalized...)
	for i := range normalized {
		if normalized[i].Mesh == nil || normalized[i].Material == nil {
			panic("no")
		}
	}
}

func (d *Drawings) Render(device *GPUDevice, lights LightsForRender, views []RenderViewFrame) {
	defer tracing.NewRegion("Drawings.Render").End()
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if len(d.renderPassGroups) == 0 {
		return
	}
	views = renderViewsForDraw(views)
	passes := make([]*RenderPass, 0, len(d.renderPassGroups))
	shadows := [MaxLocalLights]TextureId{}
	shadowIdx := 0
	for i := range d.renderPassGroups {
		rp := d.renderPassGroups[i].renderPass
		if !rp.Buffer.IsValid() {
			rp.Recontstruct(device)
		}
		passes = append(passes, rp)
		if rp.IsShadowPass() {
			if shadowIdx < len(shadows) {
				shadows[shadowIdx] = rp.textures[0].RenderId
				shadowIdx++
			}
		}
	}
	sort.Slice(passes, func(i, j int) bool {
		return passes[i].construction.Sort < passes[j].construction.Sort
	})
	for _, view := range views {
		target := view.Target()
		if target != nil {
			if err := device.PrepareRenderTarget(target); err != nil {
				slog.Error("failed to prepare render target", "target", target.Name(), "error", err)
				continue
			}
		}
		drawnPasses := make([]*RenderPass, 0, len(passes))
		drawnPassLookup := make(map[*RenderPass]struct{}, len(passes))
		layerMask := view.LayerMask()
		for i := range passes {
			rp := passes[i]
			rpGroup, ok := d.findRenderPassGroup(rp)
			if !ok || !rpGroup.MatchesLayer(layerMask) {
				continue
			}
			device.DrawView(rp, rpGroup.draws, lights, shadows[:], view, layerMask)
			if _, ok := drawnPassLookup[rp]; !ok {
				drawnPasses = append(drawnPasses, rp)
				drawnPassLookup[rp] = struct{}{}
			}
		}
		if len(drawnPasses) == 0 {
			continue
		}
		if target != nil {
			device.BlitTargetsToRenderTarget(drawnPasses, target)
			if !device.FlushQueuedCommands() {
				return
			}
		} else {
			device.BlitTargets(drawnPasses)
		}
	}
}

func renderViewsForDraw(views []RenderViewFrame) []RenderViewFrame {
	targetViews := make([]RenderViewFrame, 0, len(views))
	var defaultView RenderViewFrame
	var firstSwapchainView RenderViewFrame
	for i := range views {
		if views[i].IsDestroyed() {
			continue
		}
		if views[i].Target() != nil {
			targetViews = append(targetViews, views[i])
			continue
		}
		if views[i].Name() == DefaultRenderViewName {
			defaultView = views[i]
		} else if firstSwapchainView.View == nil {
			firstSwapchainView = views[i]
		}
	}
	swapchainView := defaultView
	if swapchainView.View == nil {
		swapchainView = firstSwapchainView
	}
	if swapchainView.View == nil {
		view := newRenderView(RenderViewOptions{
			Name:      DefaultRenderViewName,
			LayerMask: RenderLayerAll,
			Clear:     true,
		}, 0)
		swapchainView = newRenderViewFrame(view)
	}
	return append(targetViews, swapchainView)
}

func (d *Drawings) Destroy(device *GPUDevice) {
	defer tracing.NewRegion("Drawings.Destroy").End()
	for i := range d.renderPassGroups {
		for j := range d.renderPassGroups[i].draws {
			d.renderPassGroups[i].draws[j].Destroy(device)
		}
	}
	d.backDraws = klib.WipeSlice(d.backDraws)
	d.renderPassGroups = klib.WipeSlice(d.renderPassGroups)
}

func (d *Drawings) DestroyViewState(device *GPUDevice, view *RenderView) {
	defer tracing.NewRegion("Drawings.DestroyViewState").End()
	if view == nil {
		return
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	for i := range d.renderPassGroups {
		for j := range d.renderPassGroups[i].draws {
			draw := &d.renderPassGroups[i].draws[j]
			for k := range draw.instanceGroups {
				draw.instanceGroups[k].DestroyViewState(device, view)
			}
		}
	}
}

func (d *Drawings) Clear() {
	defer tracing.NewRegion("Drawings.Clear").End()
	for i := range d.renderPassGroups {
		for j := range d.renderPassGroups[i].draws {
			d.renderPassGroups[i].draws[j].Clear()
		}
	}
}
