#version 300 es
precision mediump float;

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec4 inTangent;
layout(location = 3) in vec2 inTexCoord0;
layout(location = 4) in vec4 inColor;
layout(location = 5) in ivec4 inJointIds;
layout(location = 6) in vec4 inJointWeights;
layout(location = 7) in vec3 inMorphTarget;

//uniform struct InstanceData {
//  mat4 model;
//	vec4 uvs;
//	vec4 fgColor;
//	vec4 bgColor;
//	vec4 scissor;
//	vec2 pxRange;
//}
#define INSTANCE_VEC4_COUNT 9
uniform sampler2D instanceSampler;

uniform struct GlobalData {
    mat4 view;
    mat4 projection;
    mat4 uiView;
    mat4 uiProjection;
    vec3 cameraPosition;
    vec3 uiCameraPosition;
    float time;
} globalData;

out vec4 fragColor;
out vec4 fragBGColor;
out vec2 fragTexCoord;
out vec2 fragPxRange;

mat4 pullModel(int xOffset) {
    mat4 model;
    model[0] = texelFetch(instanceSampler, ivec2(xOffset,0), 0);
    model[1] = texelFetch(instanceSampler, ivec2(xOffset+1,0), 0);
    model[2] = texelFetch(instanceSampler, ivec2(xOffset+2,0), 0);
    model[3] = texelFetch(instanceSampler, ivec2(xOffset+3,0), 0);
    return model;
}

void main() {
	mat4 view = globalData.view;
	mat4 uiView = globalData.uiView;
	mat4 projection = globalData.projection;
	mat4 uiProjection = globalData.uiProjection;
	vec3 cameraPosition = globalData.cameraPosition;
	vec3 uiCameraPosition = globalData.uiCameraPosition;
	float time = globalData.time;

    int xOffset = gl_InstanceID*INSTANCE_VEC4_COUNT;
    mat4 model = pullModel(xOffset);
	vec4 uvs = texelFetch(instanceSampler, ivec2(xOffset+4,0), 0);
	vec4 fgColor = texelFetch(instanceSampler, ivec2(xOffset+5,0), 0);
	vec4 bgColor = texelFetch(instanceSampler, ivec2(xOffset+6,0), 0);
	vec4 scissor = texelFetch(instanceSampler, ivec2(xOffset+7,0), 0);
	vec2 pxRange = texelFetch(instanceSampler, ivec2(xOffset+8,0), 0).xy;

    //vec4 vPos = model * vec4(inPosition, 1.0);
	gl_Position = uiProjection * uiView * model * vec4(inPosition, 1.0);
	vec2 uv = inTexCoord0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragTexCoord = uv;
	fragColor = inColor * fgColor;
	fragBGColor = bgColor;
	fragPxRange = pxRange;

    // TODO:  GLES clipping should be done through scissor
	//gl_ClipDistance[0] = vPos.x - scissor.x;
	//gl_ClipDistance[1] = vPos.y - scissor.y;
	//gl_ClipDistance[2] = scissor.z - vPos.x;
	//gl_ClipDistance[3] = scissor.w - vPos.y;
}