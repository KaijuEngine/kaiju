package rendering

import (
	"encoding/json"

	vk "github.com/BrentFarris/go-vulkan"
)

type ShaderDefDriver struct {
	Vert string
	Frag string
	Geom string
	Tesc string
	Tese string
}

type ShaderDefField struct {
	Name string
	Type string
}

type ShaderDef struct {
	OpenGL ShaderDefDriver
	Vulkan ShaderDefDriver
	Fields []ShaderDefField
}

func (sd *ShaderDef) AddField(name, glslType string) {
	sd.Fields = append(sd.Fields, ShaderDefField{name, glslType})
}

func ShaderDefFromJson(jsonStr string) (ShaderDef, error) {
	var def ShaderDef
	err := json.Unmarshal([]byte(jsonStr), &def)
	return def, err
}

func (sd *ShaderDef) ToAttributeDescription() []vk.VertexInputAttributeDescription {
	panic("not implemented")
}
