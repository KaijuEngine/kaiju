#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 uvs;
layout(location = LOCATION_START+1) in vec4 fgColor;
layout(location = LOCATION_START+2) in vec4 bgColor;
layout(location = LOCATION_START+3) in vec4 scissor;
layout(location = LOCATION_START+4) in vec4 size2D;
layout(location = LOCATION_START+5) in vec4 borderRadius;
layout(location = LOCATION_START+6) in vec4 borderSize;
layout(location = LOCATION_START+7) in mat4 borderColor;
layout(location = LOCATION_START+11) in vec2 borderLen;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec4 fragBGColor;
layout(location = 2) out vec4 fragSize2D;
layout(location = 3) out vec4 fragBorderRadius;
layout(location = 4) out vec4 fragBorderSize;
layout(location = 5) out mat4 fragBorderColor;
layout(location = 9) out vec2 fragTexCoord;
layout(location = 10) out vec2 fragBorderLen;
layout(location = 11) out vec4 fragUvs;

void main() {
	vec4 vPos = model * vec4(Position, 1.0);
	gl_Position = uiProjection * uiView * vPos;
	vec2 uv = UV0;
	float v = (1.0 - uvs.w) - uvs.y;
	uv *= uvs.zw;
	uv.y += v;
	uv.x += uvs.x;
	fragTexCoord = uv;
	fragColor = Color * fgColor;
	fragBGColor = bgColor;
	fragSize2D = size2D;
	fragBorderRadius = borderRadius;
	fragBorderSize = borderSize;
	fragBorderColor = borderColor;
	fragBorderLen = borderLen;
	fragUvs = uvs;
	fragUvs.y = v;

	gl_ClipDistance[0] = vPos.x - scissor.x;
	gl_ClipDistance[1] = vPos.y - scissor.y;
	gl_ClipDistance[2] = scissor.z - vPos.x;
	gl_ClipDistance[3] = scissor.w - vPos.y;
}
