#version 450

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 uvs;
layout(location = LOCATION_START+1) in vec4 fgColor;
layout(location = LOCATION_START+2) in vec4 bgColor;
layout(location = LOCATION_START+3) in vec4 scissor;
layout(location = LOCATION_START+4) in vec2 pxRange;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec4 fragBGColor;
layout(location = 2) out vec2 fragTexCoord;
layout(location = 3) out vec2 fragPxRange;
layout(location = 4) out vec2 fragTexRange;

void main() {
	vec4 vPos = model * vec4(Position, 1.0);
	// vPos.x = round(vPos.x);
	// vPos.y = round(vPos.y);
	gl_Position = uiProjection * uiView * vPos;
	vec2 uv = UV0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragTexCoord = uv;
	fragColor = Color * fgColor;
	fragBGColor = bgColor;
	fragPxRange = pxRange;
	fragTexRange = uvs.zw;

	gl_ClipDistance[0] = vPos.x - scissor.x;
	gl_ClipDistance[1] = vPos.y - scissor.y;
	gl_ClipDistance[2] = scissor.z - vPos.x;
	gl_ClipDistance[3] = scissor.w - vPos.y;
}