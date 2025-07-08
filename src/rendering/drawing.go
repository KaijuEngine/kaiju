/******************************************************************************/
/* drawing.go                                                                 */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"sort"
	"sync"
)

type Drawing struct {
	Renderer     Renderer
	Material     *Material
	Mesh         *Mesh
	ShaderData   DrawInstance
	Transform    *matrix.Transform
	Sort         int
	CastsShadows bool
}

func (d *Drawing) IsValid() bool {
	return d.Material != nil
}

type RenderPassGroup struct {
	renderPass *RenderPass
	draws      []ShaderDraw
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
	idx := -1
	for i := 0; i < len(sd.instanceGroups) && idx < 0; i++ {
		g := &sd.instanceGroups[i]
		if g.Mesh == dg.Mesh &&
			(g.MaterialInstance == dg.Material || g.MaterialInstance.Root.Value() == dg.Material) {
			idx = i
		}
	}
	return idx
}

func (d *RenderPassGroup) findShaderDraw(material *Material) (*ShaderDraw, bool) {
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
	for i := range d.renderPassGroups {
		if d.renderPassGroups[i].renderPass == renderPass {
			return &d.renderPassGroups[i], true
		}
	}
	return nil, false
}

func (d *Drawings) addToRenderPassGroup(drawing *Drawing, rpGroup *RenderPassGroup) {
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
		group := NewDrawInstanceGroup(drawing.Mesh, drawing.ShaderData.Size())
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

func (d *Drawings) PreparePending() {
	defer tracing.NewRegion("Drawings.PreparePending").End()
	d.mutex.RLock()
	defer d.mutex.RUnlock()
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
		if drawing.CastsShadows {
			d.backDraws = append(d.backDraws, lightTransformDrawingToDepth(drawing))
			//d.backDraws = append(d.backDraws, lightTransformDrawingToCubeDepth(drawing))
		}
	}
	d.backDraws = d.backDraws[:0]
}

func (d *Drawings) AddDrawing(drawing Drawing) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.backDraws = append(d.backDraws, drawing)
	if drawing.Material == nil {
		panic("no")
	}
}

func (d *Drawings) AddDrawings(drawings []Drawing) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.backDraws = append(d.backDraws, drawings...)
	for i := range drawings {
		if drawings[i].Material == nil {
			panic("no")
		}
	}
}

func (d *Drawings) Render(renderer Renderer) {
	defer tracing.NewRegion("Drawings.Render").End()
	if len(d.renderPassGroups) == 0 {
		return
	}
	passes := make([]*RenderPass, 0, len(d.renderPassGroups))
	for i := range d.renderPassGroups {
		rp := d.renderPassGroups[i].renderPass
		if renderer.Draw(rp, d.renderPassGroups[i].draws) {
			passes = append(passes, rp)
		}
	}
	if len(passes) > 0 {
		sort.Slice(passes, func(i, j int) bool {
			return passes[i].construction.Sort < passes[j].construction.Sort
		})
		renderer.BlitTargets(passes)
	}
}

func (d *Drawings) Destroy(renderer Renderer) {
	for i := range d.renderPassGroups {
		for j := range d.renderPassGroups[i].draws {
			d.renderPassGroups[i].draws[j].Destroy(renderer)
		}
	}
	d.renderPassGroups = d.renderPassGroups[:0]
}

func (d *Drawings) Clear(renderer Renderer) {
	for i := range d.renderPassGroups {
		for j := range d.renderPassGroups[i].draws {
			d.renderPassGroups[i].draws[j].Clear(renderer)
		}
	}
}
