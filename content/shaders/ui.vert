#version 460

layout (location = 0) in vec3 Position;
layout (location = 1) in vec3 Normal;
layout (location = 2) in vec4 Tangent;
layout (location = 3) in vec2 UV0;
layout (location = 4) in vec4 Color;
layout (location = 5) in ivec4 JointIds;
layout (location = 6) in vec4 JointWeights;
layout (location = 7) in vec3 MorphTarget;

layout(location = 8) in mat4 model;
layout(location = 12) in vec4 uvs;
layout(location = 13) in vec4 fgColor;
layout(location = 14) in vec4 bgColor;
layout(location = 15) in vec4 scissor;
layout(location = 16) in vec4 size2D;
layout(location = 17) in vec4 borderRadius;
layout(location = 18) in vec4 borderSize;
layout(location = 19) in mat4 borderColor;
layout(location = 23) in vec2 borderLen;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec4 fragBGColor;
layout(location = 2) out vec4 fragSize2D;
layout(location = 3) out vec4 fragBorderRadius;
layout(location = 4) out vec4 fragBorderSize;
layout(location = 5) out mat4 fragBorderColor;
layout(location = 9) out vec2 fragTexCoord;
layout(location = 10) out vec2 fragBorderLen;

struct InstanceData {
	mat4 model;
	vec4 uvs;
	vec4 fgColor;
	vec4 bgColor;
	vec4 scissor;
	vec4 size2D;
	vec4 borderRadius;
	vec4 borderSize;
	mat4 borderColor;
	vec2 borderLen;
};

layout(set = 0, binding = 0) readonly uniform UniformBufferObject {
	mat4 view;
	mat4 projection;
	mat4 uiView;
	mat4 uiProjection;
	vec3 cameraPosition;
	vec3 uiCameraPosition;
	vec2 screenSize;
	float time;
};

void main() {
	vec4 vPos = model * vec4(Position, 1.0);
	gl_Position = uiProjection * uiView * vPos;
	vec2 uv = UV0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragColor = Color * fgColor;
	fragBGColor = bgColor;
	fragSize2D = size2D;
	fragTexCoord = uv;
	fragBorderRadius = borderRadius;
	fragBorderSize = borderSize;
	fragBorderColor = borderColor;
	fragBorderLen = borderLen;

	gl_ClipDistance[0] = vPos.x - scissor.x;
	gl_ClipDistance[1] = vPos.y - scissor.y;
	gl_ClipDistance[2] = scissor.z - vPos.x;
	gl_ClipDistance[3] = scissor.w - vPos.y;
}
