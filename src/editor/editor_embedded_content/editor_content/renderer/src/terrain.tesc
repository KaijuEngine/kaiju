#version 460

layout(vertices = 3) out;

layout(location = 0) in vec4 tescColor[];
layout(location = 1) in vec2 tescTexCoord[];
layout(location = 2) in vec3 tescCamPos[];
layout(location = 3) in float tescScalar[];
layout(location = 4) in mat4 tescView[];
layout(location = 8) in mat4 tescProjection[];
layout(location = 12) in mat4 tescModel[];
layout(location = 16) in mat3 tescNmlModel[];

layout(location = 0) out vec4 teseColor[];
layout(location = 1) out vec2 teseTexCoord[];
layout(location = 2) out vec3 teseCamPos[];
layout(location = 3) out float teseScalar[];
layout(location = 4) out mat4 teseView[];
layout(location = 8) out mat4 teseProjection[];
layout(location = 12) out mat4 teseModel[];
layout(location = 16) out mat3 teseNmlModel[];

void main() {
	gl_out[gl_InvocationID].gl_Position = gl_in[gl_InvocationID].gl_Position;

	teseColor[gl_InvocationID] = tescColor[gl_InvocationID];
	teseTexCoord[gl_InvocationID] = tescTexCoord[gl_InvocationID];
	teseCamPos[gl_InvocationID] = tescCamPos[gl_InvocationID];
	teseScalar[gl_InvocationID] = tescScalar[gl_InvocationID];
	teseView[gl_InvocationID] = tescView[gl_InvocationID];
	teseProjection[gl_InvocationID] = tescProjection[gl_InvocationID];
	teseModel[gl_InvocationID] = tescModel[gl_InvocationID];
	teseNmlModel[gl_InvocationID] = tescNmlModel[gl_InvocationID];

	//float dist = distance(gl_in[gl_InvocationID].gl_Position.xyz, tescCamPos[gl_InvocationID]);
	float dist = 0.0;
	int divide = 500;
	int outerDivide = clamp(int(divide * (1.0 - (dist / 50.0))), 1, divide);

	gl_TessLevelOuter[0] = outerDivide; // left for triangles
	gl_TessLevelOuter[1] = outerDivide; // bot for triangles
	gl_TessLevelOuter[2] = outerDivide; // right for triangles
	
	gl_TessLevelInner[0] = outerDivide * 2; // all inner sides for triangles
}
