package rendering

import (
	"kaiju/matrix"
	"slices"
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
	draws []ShaderDraw
}

func NewDrawings() Drawings {
	return Drawings{
		draws: make([]ShaderDraw, 0),
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

func (d *Drawings) matchGroup(sd *ShaderDraw, dg *Drawing) (*DrawInstanceGroup, bool) {
	var dig *DrawInstanceGroup = nil
	for i := 0; i < len(sd.instanceGroups) && dig == nil; i++ {
		g := &sd.instanceGroups[i]
		if g.Mesh == dg.Mesh && texturesMatch(g.Textures, dg.Textures) && dg.UseBlending == g.useBlending {
			dig = g
		}
	}
	return dig, dig != nil
}

func (d *Drawings) AddDrawing(drawing Drawing) {
	draw, ok := d.findShaderDraw(drawing.Shader)
	if !ok {
		newDraw := NewShaderDraw(drawing.Shader)
		d.draws = append(d.draws, newDraw)
		draw = &d.draws[len(d.draws)-1]
	}
	drawing.ShaderData.setTransform(drawing.Transform)
	if dg, ok := d.matchGroup(draw, &drawing); ok {
		dg.AddInstance(drawing.ShaderData, drawing.Renderer, drawing.Shader)
	} else {
		group := NewDrawInstanceGroup(drawing.Mesh, drawing.ShaderData.Size())
		group.AddInstance(drawing.ShaderData, drawing.Renderer, drawing.Shader)
		group.Textures = drawing.Textures
		group.useBlending = drawing.UseBlending
		draw.AddInstanceGroup(group)
	}
}

func (d *Drawings) AddDrawings(drawings []Drawing) {
	for _, drawing := range drawings {
		d.AddDrawing(drawing)
	}
}

func (d *Drawings) Render(renderer Renderer) {
	renderer.Draw(d.draws)
}
