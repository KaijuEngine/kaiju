#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 color;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;

void main() {
	fragTexCoords = UV0;
	vec4 wp = model * vec4(Position, 1.0);
	fragColor = color;
	gl_Position = projection * view * wp;
}
