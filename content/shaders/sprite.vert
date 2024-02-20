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

layout(location = 0) out vec4 fragColor;
layout(location = 4) out vec2 fragTexCoord;

layout(set = 0, binding = 0) readonly uniform UniformBufferObject {
	mat4 view;
	mat4 projection;
	mat4 uiView;
	mat4 uiProjection;
	vec3 cameraPosition;
	vec3 uiCameraPosition;
	vec2 screenSize;
	float time;
} globalData;

void main() {
	vec4 vPos = model * vec4(Position, 1.0);
	gl_Position = globalData.uiProjection * globalData.uiView * vPos;
	vec2 uv = UV0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragColor = Color * fgColor;
	fragTexCoord = uv;
}
