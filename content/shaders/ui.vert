#version 300 es
precision mediump float;

layout (location = 0) in vec3 Position;
layout (location = 1) in vec3 Normal;
layout (location = 2) in vec4 Tangent;
layout (location = 3) in vec2 UV0;
layout (location = 4) in vec4 Color;
layout (location = 5) in vec4 JointIds;
layout (location = 6) in vec4 JointWeights;
layout (location = 7) in vec3 MorphTarget;

out vec4 fragColor;
out vec4 fragBGColor;
out vec4 fragSize2D;
out vec4 fragScissor;
out vec2 fragTexCoord;
out vec2 fragBorderLen;

uniform struct GlobalData {
    mat4 view;
    mat4 projection;
    mat4 uiView;
    mat4 uiProjection;
    vec3 cameraPosition;
    vec3 uiCameraPosition;
    float time;
} globalData;

uniform sampler2D instanceSampler;
#define INSTANCE_VEC4_COUNT 10
struct InstanceData {
	mat4 model;
	vec4 uvs;
	vec4 fgColor;
	vec4 bgColor;
	vec4 scissor;
	vec4 size2D;
	vec2 borderLen;
};

InstanceData pullInstanceData() {
	InstanceData data;
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
}
