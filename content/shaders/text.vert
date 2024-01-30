#version 450
//#version 300 es
//precision mediump float;

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec4 inTangent;
layout(location = 3) in vec2 inTexCoord0;
layout(location = 4) in vec4 inColor;
layout(location = 5) in ivec4 inJointIds;
layout(location = 6) in vec4 inJointWeights;
layout(location = 7) in vec3 inMorphTarget;

#ifdef VULKAN
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
#else
	out vec4 fragColor;
	out vec4 fragBGColor;
	out vec2 fragTexCoord;
	out vec2 fragPxRange;
	out vec2 fragTexRange;

	#define INSTANCE_VEC4_COUNT 9
	uniform sampler2D instanceSampler;
#endif

struct InstanceData {
	mat4 model;
	vec4 uvs;
	vec4 fgColor;
	vec4 bgColor;
	vec4 scissor;
	vec2 pxRange;
};

#ifdef VULKAN
	layout(set = 0, binding = 0) readonly uniform UniformBufferObject {
#else
	uniform struct GlobalData {
#endif
    mat4 view;
    mat4 projection;
    mat4 uiView;
    mat4 uiProjection;
    vec3 cameraPosition;
    vec3 uiCameraPosition;
    float time;
} globalData;

InstanceData pullInstanceData() {
	InstanceData data;
#ifdef VULKAN
	data.model = model;
	data.uvs = uvs;
	data.fgColor = fgColor;
	data.bgColor = bgColor;
	data.scissor = scissor;
	data.pxRange = pxRange;
#else
	int xOffset = gl_InstanceID*INSTANCE_VEC4_COUNT;
    data.model[0] = texelFetch(instanceSampler, ivec2(xOffset,0), 0);
    data.model[1] = texelFetch(instanceSampler, ivec2(xOffset+1,0), 0);
    data.model[2] = texelFetch(instanceSampler, ivec2(xOffset+2,0), 0);
    data.model[3] = texelFetch(instanceSampler, ivec2(xOffset+3,0), 0);
	data.uvs = texelFetch(instanceSampler, ivec2(xOffset+4,0), 0);
	data.fgColor = texelFetch(instanceSampler, ivec2(xOffset+5,0), 0);
	data.bgColor = texelFetch(instanceSampler, ivec2(xOffset+6,0), 0);
	data.scissor = texelFetch(instanceSampler, ivec2(xOffset+7,0), 0);
	data.pxRange = texelFetch(instanceSampler, ivec2(xOffset+8,0), 0).xy;
#endif
	return data;
}

void main() {
	InstanceData data = pullInstanceData();
	mat4 uiView = globalData.uiView;
	mat4 uiProjection = globalData.uiProjection;

	vec4 vPos = data.model * vec4(inPosition, 1.0);
	gl_Position = uiProjection * uiView * vPos;
	vec2 uv = inTexCoord0;
	uv *= data.uvs.zw;
	uv.y += (1.0 - data.uvs.w) - data.uvs.y;
	uv.x += data.uvs.x;
	fragTexCoord = uv;
	fragColor = inColor * data.fgColor;
	fragBGColor = data.bgColor;
	fragPxRange = data.pxRange;
	fragTexRange = data.uvs.zw;

#ifdef VULKAN
	gl_ClipDistance[0] = vPos.x - data.scissor.x;
	gl_ClipDistance[1] = vPos.y - data.scissor.y;
	gl_ClipDistance[2] = data.scissor.z - vPos.x;
	gl_ClipDistance[3] = data.scissor.w - vPos.y;
#endif
}