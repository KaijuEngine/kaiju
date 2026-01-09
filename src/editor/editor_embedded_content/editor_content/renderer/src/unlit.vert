#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START)   in vec4 color;
layout(location = LOCATION_START+1) in vec4 uvs;
layout(location = LOCATION_START+2) in uint flags;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;
layout(location = 2) out uint fragFlags;
layout(location = 3) out vec3 fragPos;
layout(location = 4) out vec3 fragNormal;

void main() {
	fragColor = Color * color;
	fragFlags = flags;
	vec2 uv = UV0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragTexCoords = uv;
	fragNormal = mat3(model) * Normal;
	vec4 wp = model * vec4(Position, 1.0);
    fragPos = wp.xyz; 
	gl_Position = projection * view * wp;
}


