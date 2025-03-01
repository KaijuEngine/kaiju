#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 color;

layout(location = 0) out vec4 fragColor;

void main() {
	float outlineWidth = color.a;
	mat4 vp = projection * view;
	vec4 clip = vp * model * vec4(Position, 1.0);
	vec3 clipNml = mat3(vp) * mat3(model) * Normal;
	vec2 offset = normalize(clipNml.xy) / screenSize.xy * outlineWidth * clip.w * 2;
	clip.xy += offset;
	fragColor = color;
	gl_Position = clip;
}
