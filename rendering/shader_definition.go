package rendering

import (
	"encoding/json"
	"kaiju/matrix"
	"math"
	"unsafe"

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

func (f ShaderDefField) Vec4Size() uint32 {
	return uint32(math.Ceil(float64(defTypes[f.Type].size) / float64(vec4Size)))
}

func (f ShaderDefField) Format() vk.Format {
	return defTypes[f.Type].format
}

type ShaderDef struct {
	OpenGL ShaderDefDriver
	Vulkan ShaderDefDriver
	Fields []ShaderDefField
}

const floatSize = int(unsafe.Sizeof(matrix.Float(0.0)))
const vec4Size = int(unsafe.Sizeof(matrix.Vec4{}))

type defType struct {
	size   uint32
	format vk.Format
	repeat int
}

var defTypes = map[string]defType{
	"float": defType{uint32(floatSize), vk.FormatR32Sfloat, 1},
	"vec2":  defType{uint32(floatSize) * 2, vk.FormatR32g32Sfloat, 1},
	"vec3":  defType{uint32(floatSize) * 3, vk.FormatR32g32b32Sfloat, 1},
	"vec4":  defType{uint32(vec4Size), vk.FormatR32g32b32a32Sfloat, 1},
	"mat4":  defType{uint32(vec4Size) * 4, vk.FormatR32g32b32a32Sfloat, 4},
}

func (sd *ShaderDef) AddField(name, glslType string) {
	sd.Fields = append(sd.Fields, ShaderDefField{name, glslType})
}

func ShaderDefFromJson(jsonStr string) (ShaderDef, error) {
	var def ShaderDef
	err := json.Unmarshal([]byte(jsonStr), &def)
	return def, err
}

func (sd ShaderDef) Stride() uint32 {
	stride := uint32(0)
	for _, field := range sd.Fields {
		stride += defTypes[field.Type].size
	}
	return stride
}

func (sd ShaderDef) ToAttributeDescription(locationStart uint32) []vk.VertexInputAttributeDescription {
	attrs := make([]vk.VertexInputAttributeDescription, 0, len(sd.Fields))
	location := locationStart
	offset := uint32(0)
	for _, field := range sd.Fields {
		for j := 0; j < defTypes[field.Type].repeat; j++ {
			attrs = append(attrs, vk.VertexInputAttributeDescription{
				Location: location,
				Binding:  0,
				Format:   field.Format(),
				Offset:   offset,
			})
			location++
			offset += defTypes[field.Type].size
		}
	}
	return attrs
}
