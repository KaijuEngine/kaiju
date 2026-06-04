/******************************************************************************/
/* combined_target_cache.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type combinedTargetDrawCache struct {
	entries map[string]*combinedTargetDrawEntry
}

// Entries are retained across frames so multiple stage viewports/picking views
// can alternate combine signatures without destroying GPU instance buffers.
type combinedTargetDrawEntry struct {
	signature string
	drawings  Drawings
}

func (c *combinedTargetDrawCache) Prepare(device *GPUDevice, signature string, specs []combinedTargetSpec, combineMat *Material, culler ViewCuller) (*combinedTargetDrawEntry, error) {
	defer tracing.NewRegion("combinedTargetDrawCache.Prepare").End()
	if signature == "" || len(specs) == 0 {
		return nil, nil
	}
	if combineMat == nil {
		return nil, errors.New("cannot prepare combined targets without a combine material")
	}
	if c.entries == nil {
		c.entries = make(map[string]*combinedTargetDrawEntry)
	}
	if entry, ok := c.entries[signature]; ok && entry.HasDrawings() {
		return entry, nil
	}
	entry := &combinedTargetDrawEntry{
		signature: signature,
		drawings:  NewDrawings(),
	}
	entry.build(device, specs, combineMat, culler)
	c.entries[signature] = entry
	return entry, nil
}

func (c *combinedTargetDrawCache) Destroy(device *GPUDevice) {
	defer tracing.NewRegion("combinedTargetDrawCache.Destroy").End()
	for key, entry := range c.entries {
		if entry != nil {
			entry.Destroy(device)
		}
		delete(c.entries, key)
	}
}

func (c *combinedTargetDrawCache) EntryCount() int {
	return len(c.entries)
}

func (e *combinedTargetDrawEntry) HasDrawings() bool {
	return e != nil && e.drawings.HasDrawings() &&
		len(e.drawings.renderPassGroups) > 0 &&
		len(e.drawings.renderPassGroups[0].draws) > 0
}

func (e *combinedTargetDrawEntry) DrawsAndPass() ([]ShaderDraw, *RenderPass, bool) {
	if !e.HasDrawings() {
		return nil, nil, false
	}
	group := &e.drawings.renderPassGroups[0]
	return group.draws, group.renderPass, true
}

func (e *combinedTargetDrawEntry) InstanceGroups() []DrawInstanceGroup {
	draws, _, ok := e.DrawsAndPass()
	if !ok || len(draws) == 0 {
		return nil
	}
	return draws[0].instanceGroups
}

func (e *combinedTargetDrawEntry) Destroy(device *GPUDevice) {
	if e == nil {
		return
	}
	e.drawings.Destroy(device)
}

func (e *combinedTargetDrawEntry) build(device *GPUDevice, specs []combinedTargetSpec, combineMat *Material, culler ViewCuller) {
	defer tracing.NewRegion("combinedTargetDrawEntry.build").End()
	mesh := NewMeshQuad(device.Painter.caches.MeshCache())
	for i := range specs {
		sd := &ShaderDataCombine{NewShaderDataBase(), matrix.Color{1, 1, 1, 1}}
		m := matrix.Mat4Identity()
		m.Scale(matrix.Vec3{1, 1, 1})
		sd.SetModel(m)
		mat := combineMat.CreateInstance([]*Texture{
			specs[i].color,
			specs[i].position,
			specs[i].normal,
		})
		e.drawings.AddDrawing(Drawing{
			Material:   mat,
			Mesh:       mesh,
			ShaderData: sd,
			Sort:       specs[i].sort,
			Layer:      RenderLayerUI,
			ViewCuller: culler,
		})
	}
	e.drawings.PreparePending(0)
}
