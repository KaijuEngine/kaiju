/******************************************************************************/
/* shader_vulkan.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/platform/profiler/tracing"
	vk "kaijuengine.com/rendering/vulkan"
)

func (sd *ShaderDataCompiled) ToAttributeDescription(locationStart uint32) []vk.VertexInputAttributeDescription {
	defer tracing.NewRegion("Shader.ToAttributeDescription").End()
	attrs := make([]vk.VertexInputAttributeDescription, 0)
	offset := uint32(0)
	if !sd.IsCompute() {
		g := sd.SelectLayout("Vertex")
		for i := range g.Layouts {
			l := &g.Layouts[i]
			if l.Source == "in" && uint32(l.Location) >= locationStart {
				dt := defTypes[l.Type]
				dt.repeat *= max(1, l.Count)
				for r := range dt.repeat {
					attrs = append(attrs, vk.VertexInputAttributeDescription{
						Location: uint32(l.Location + r),
						Binding:  1,
						Format:   dt.format,
						Offset:   offset,
					})
					offset += dt.size
				}
			}
		}
	}
	return attrs
}
