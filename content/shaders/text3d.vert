#version 450

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec4 inTangent;
layout(location = 3) in vec2 inTexCoord0;
layout(location = 4) in vec4 inColor;
layout(location = 5) in ivec4 inJointIds;
layout(location = 6) in vec4 inJointWeights;
layout(location = 7) in vec3 inMorphTarget;

layout(location = 8) in mat4 model;
layout(location = 12) in vec4 uvs;
layout(location = 13) in vec4 fgColor;
layout(location = 14) in vec4 bgColor;
layout(location = 15) in vec4 scissor;
layout(location = 16) in vec2 pxRange;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec4 fragBGColor;
layout(location = 2) out vec2 fragTexCoord;
layout(location = 3) out vec2 fragPxRange;
layout(location = 4) out vec2 fragTexRange;

struct InstanceData {
	mat4 model;
	vec4 uvs;
	vec4 fgColor;
	vec4 bgColor;
	vec4 scissor;
	vec2 pxRange;
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
    vec2 uv = inTexCoord0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragTexCoord = uv;
	fragColor = inColor * fgColor;
	fragBGColor = bgColor;
	gl_Position = projection * view * model * vec4(inPosition, 1.0);
	fragPxRange = pxRange;
}