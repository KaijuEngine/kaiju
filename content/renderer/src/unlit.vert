#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START)   in vec4 color;
layout(location = LOCATION_START+1) in vec4 uvs;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;

void main() {
	fragColor = Color * color;
	vec2 uv = UV0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragTexCoords = uv;
	vec4 wp = model * vec4(Position, 1.0);
	gl_Position = projection * view * wp;
}


