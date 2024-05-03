#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 color;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;

void main() {
	fragColor = Color * color;
	fragTexCoords = UV0;
	gl_Position = projection * view * model * vec4(Position, 1.0);
}
