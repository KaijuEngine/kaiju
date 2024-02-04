package rendering

type ShaderDraw struct {
	shader         *Shader
	instanceGroups []DrawInstanceGroup
}

func NewShaderDraw(shader *Shader) ShaderDraw {
	return ShaderDraw{
		shader:         shader,
		instanceGroups: make([]DrawInstanceGroup, 0),
	}
}

func (s *ShaderDraw) AddInstanceGroup(group DrawInstanceGroup) {
	s.instanceGroups = append(s.instanceGroups, group)
}

func (s *ShaderDraw) Filter(filter func(*DrawInstanceGroup) bool) []*DrawInstanceGroup {
	selected := make([]*DrawInstanceGroup, 0, len(s.instanceGroups))
	for i := range s.instanceGroups {
		if filter(&s.instanceGroups[i]) {
			selected = append(selected, &s.instanceGroups[i])
		}
	}
	return selected
}

func (s *ShaderDraw) SolidGroups() []*DrawInstanceGroup {
	return s.Filter(func(g *DrawInstanceGroup) bool { return !g.useBlending })
}

func (s *ShaderDraw) TransparentGroups() []*DrawInstanceGroup {
	return s.Filter(func(g *DrawInstanceGroup) bool { return g.useBlending })
}

func (s *ShaderDraw) Destroy(renderer Renderer) {
	for i := range s.instanceGroups {
		s.instanceGroups[i].Destroy(renderer)
	}
}
