#version 460
//#version 300 es
//precision mediump float;

layout (location = 0) in vec3 Position;
layout (location = 1) in vec3 Normal;
layout (location = 2) in vec4 Tangent;
layout (location = 3) in vec2 UV0;
layout (location = 4) in vec4 Color;
layout (location = 5) in ivec4 JointIds;
layout (location = 6) in vec4 JointWeights;
layout (location = 7) in vec3 MorphTarget;

#ifdef VULKAN
	layout(location = 8) in mat4 model;
	layout(location = 12) in vec4 uvs;
	layout(location = 13) in vec4 fgColor;
	layout(location = 14) in vec4 bgColor;
	layout(location = 15) in vec4 scissor;
	layout(location = 16) in vec4 size2D;
	layout(location = 17) in vec2 borderLen;
	//layout(location = 17) in vec4 borderRadius;
	//layout(location = 18) in vec4 borderSize;
	//layout(location = 19) in mat4 borderColor;
	//layout(location = 23) in vec2 borderLen;

	layout(location = 0) out vec4 fragColor;
	layout(location = 1) out vec4 fragBGColor;
	layout(location = 2) out vec4 fragSize2D;
	layout(location = 3) out vec4 fragScissor;
	layout(location = 4) out vec2 fragTexCoord;
	layout(location = 5) out vec2 fragBorderLen;
#else
	out vec4 fragColor;
	out vec4 fragBGColor;
	out vec4 fragSize2D;
	out vec4 fragScissor;
	out vec2 fragTexCoord;
	out vec2 fragBorderLen;

	uniform sampler2D instanceSampler;
	#define INSTANCE_VEC4_COUNT 10
#endif

struct InstanceData {
	mat4 model;
	vec4 uvs;
	vec4 fgColor;
	vec4 bgColor;
	vec4 scissor;
	vec4 size2D;
	vec2 borderLen;
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
	data.size2D = size2D;
	data.borderLen = borderLen;
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
	data.size2D = texelFetch(instanceSampler, ivec2(xOffset+8,0), 0);
	data.borderLen = texelFetch(instanceSampler, ivec2(xOffset+9,0), 0).xy;
#endif
	return data;
}

void main() {
	InstanceData data = pullInstanceData();
	vec4 vPos = data.model * vec4(Position, 1.0);
	gl_Position = globalData.uiProjection * globalData.uiView * vPos;
	vec2 uv = UV0;
	uv *= data.uvs.zw;
	uv.y += (1.0 - data.uvs.w) - data.uvs.y;
	uv.x += data.uvs.x;
	fragColor = Color * data.fgColor;
	fragBGColor = data.bgColor;
	fragSize2D = data.size2D;
	fragScissor = data.scissor;
	fragTexCoord = uv;
	fragBorderLen = data.borderLen;

#ifdef VULKAN
	gl_ClipDistance[0] = vPos.x - data.scissor.x;
	gl_ClipDistance[1] = vPos.y - data.scissor.y;
	gl_ClipDistance[2] = data.scissor.z - vPos.x;
	gl_ClipDistance[3] = data.scissor.w - vPos.y;
#endif
}
