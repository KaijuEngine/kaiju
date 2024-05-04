#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 uvs;
layout(location = 13) in vec4 fgColor;

layout(location = 0) out vec4 fragColor;
layout(location = 4) out vec2 fragTexCoord;

void main() {
	vec4 vPos = model * vec4(Position, 1.0);
	gl_Position = uiProjection * uiView * vPos;
	vec2 uv = UV0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragColor = Color * fgColor;
	fragTexCoord = uv;
}
