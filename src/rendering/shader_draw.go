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

func (s *ShaderDraw) Filter(filter func(DrawInstanceGroup) bool) []DrawInstanceGroup {
	selected := make([]DrawInstanceGroup, 0, len(s.instanceGroups))
	for _, g := range s.instanceGroups {
		if filter(g) {
			selected = append(selected, g)
		}
	}
	return selected
}

func (s *ShaderDraw) SolidGroups() []DrawInstanceGroup {
	return s.Filter(func(g DrawInstanceGroup) bool { return !g.useBlending })
}

func (s *ShaderDraw) TransparentGroups() []DrawInstanceGroup {
	return s.Filter(func(g DrawInstanceGroup) bool { return g.useBlending })
}