package rendering

import "slices"

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

func (d *Drawings) matchGroup(sd *ShaderDraw, dg *DrawInstanceGroup) (*DrawInstanceGroup, bool) {
	var dig *DrawInstanceGroup = nil
	for i := 0; i < len(sd.instanceGroups) && dig == nil; i++ {
		g := &sd.instanceGroups[i]
		if g.Mesh == dg.Mesh && texturesMatch(g.Textures, dg.Textures) {
			dig = g
		}
	}
	return dig, dig != nil
}

func (d *Drawings) AddDrawing(shader *Shader, drawGroup DrawInstanceGroup) {
	draw, ok := d.findShaderDraw(shader)
	if !ok {
		newDraw := NewShaderDraw(shader)
		d.draws = append(d.draws, newDraw)
		draw = &d.draws[len(d.draws)-1]
	}
	if dg, ok := d.matchGroup(draw, &drawGroup); ok {
		dg.Merge(&drawGroup)
	} else {
		draw.instanceGroups = append(draw.instanceGroups, drawGroup)
	}
}

func (d *Drawings) Render(renderer Renderer) {
	renderer.Draw(d.draws)
}
