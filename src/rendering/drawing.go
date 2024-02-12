/*****************************************************************************/
/* drawing.go                                                                */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package rendering

import (
	"kaiju/matrix"
	"slices"
	"sync"
)

type Drawing struct {
	Renderer    Renderer
	Shader      *Shader
	Mesh        *Mesh
	Textures    []*Texture
	ShaderData  DrawInstance
	Transform   *matrix.Transform
	UseBlending bool
}

func (d *Drawing) IsValid() bool { return d.Shader != nil }

type Drawings struct {
	draws     []ShaderDraw
	backDraws []Drawing
	mutex     sync.RWMutex
}

func NewDrawings() Drawings {
	return Drawings{
		draws:     make([]ShaderDraw, 0),
		backDraws: make([]Drawing, 0),
		mutex:     sync.RWMutex{},
	}
}

func (d *Drawings) findShaderDraw(shader *Shader) (*ShaderDraw, bool) {
	for i := range d.draws {
		if d.draws[i].shader == shader {
			return &d.draws[i], true
		}
	}
	return nil, false
}

func texturesMatch(a []*Texture, b []*Texture) bool {
	if len(a) != len(b) {
		return false
	}
	for _, ta := range a {
		if !slices.Contains(b, ta) {
			return false
		}
	}
	return true
}

func (d *Drawings) matchGroup(sd *ShaderDraw, dg *Drawing) int {
	idx := -1
	for i := 0; i < len(sd.instanceGroups) && idx < 0; i++ {
		g := &sd.instanceGroups[i]
		if g.Mesh == dg.Mesh && texturesMatch(g.Textures, dg.Textures) && dg.UseBlending == g.useBlending {
			idx = i
		}
	}
	return idx
}

func (d *Drawings) PreparePending() {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for i := range d.backDraws {
		drawing := &d.backDraws[i]
		draw, ok := d.findShaderDraw(drawing.Shader)
		if !ok {
			newDraw := NewShaderDraw(drawing.Shader)
			d.draws = append(d.draws, newDraw)
			draw = &d.draws[len(d.draws)-1]
		}
		drawing.ShaderData.setTransform(drawing.Transform)
		idx := d.matchGroup(draw, drawing)
		if idx >= 0 && !draw.instanceGroups[idx].destroyed {
			draw.instanceGroups[idx].AddInstance(drawing.ShaderData, drawing.Renderer, drawing.Shader)
		} else {
			group := NewDrawInstanceGroup(drawing.Mesh, drawing.ShaderData.Size())
			group.AddInstance(drawing.ShaderData, drawing.Renderer, drawing.Shader)
			group.Textures = drawing.Textures
			group.useBlending = drawing.UseBlending
			if idx >= 0 {
				draw.instanceGroups[idx] = group
			} else {
				draw.AddInstanceGroup(group)
			}
		}
	}
	d.backDraws = d.backDraws[:0]
}

func (d *Drawings) AddDrawing(drawing Drawing) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.backDraws = append(d.backDraws, drawing)
}

func (d *Drawings) AddDrawings(drawings []Drawing) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.backDraws = append(d.backDraws, drawings...)
}

func (d *Drawings) Render(renderer Renderer) {
	renderer.Draw(d.draws)
	renderer.BlitTargets(RenderTargetDraw{
		Target: renderer.DefaultTarget(),
		Rect:   matrix.Vec4{0, 0, 1, 1},
	})
}

func (d *Drawings) RenderToTarget(renderer Renderer, target RenderTarget) {
	renderer.DrawToTarget(d.draws, target)
}

func (d *Drawings) Destroy(renderer Renderer) {
	for i := range d.draws {
		d.draws[i].Destroy(renderer)
	}
}
