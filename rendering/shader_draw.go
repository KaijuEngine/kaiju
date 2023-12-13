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

func (s *ShaderDraw) AddInstanceGroup(group *DrawInstanceGroup) {
	s.instanceGroups = append(s.instanceGroups, *group)
}
