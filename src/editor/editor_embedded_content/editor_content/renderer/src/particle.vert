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
	vec4 centerWorld = model * vec4(0.0, 0.0, 0.0, 1.0);
    vec3 camRight = vec3(view[0].x, view[1].x, view[2].x);
    vec3 camUp    = vec3(view[0].y, view[1].y, view[2].y);
    vec3 offset = camRight * Position.x + camUp * Position.y;
    vec4 worldPos = centerWorld + vec4(offset, 0.0);
    vec4 camSpace = view * worldPos;
    gl_Position = projection * camSpace;
}
