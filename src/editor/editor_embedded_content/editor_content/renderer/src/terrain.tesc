#version 460
#define TESS_CONTROL_SHADER

#include "kaiju.glsl"

#define TERRAIN_TESS_MIN_LEVEL 1.0
#define TERRAIN_TESS_MAX_LEVEL 8.0
#define TERRAIN_TESS_NEAR_DISTANCE 12.0
#define TERRAIN_TESS_FAR_DISTANCE 120.0

layout(vertices = 3) out;

layout(location = 0) in vec4 inColor[];
layout(location = 1) flat in uint inFlags[];
layout(location = 2) in vec3 inPos[];
layout(location = 3) in vec2 inTexCoords[];
layout(location = 4) in vec3 inNormal[];
layout(location = 5) in vec3 inViewDir[];
layout(location = 6) in vec4 inBrushCenterRadius[];
layout(location = 7) in vec4 inBrushParams[];
layout(location = 8) in vec4 inBrushColor[];
layout(location = 9) flat in ivec4 inTerrainLightIds[];
layout(location = 10) flat in int inTerrainInstance[];

layout(location = 0) out vec4 tcColor[];
layout(location = 1) flat out uint tcFlags[];
layout(location = 2) out vec3 tcPos[];
layout(location = 3) out vec2 tcTexCoords[];
layout(location = 4) out vec3 tcNormal[];
layout(location = 5) out vec3 tcViewDir[];
layout(location = 6) out vec4 tcBrushCenterRadius[];
layout(location = 7) out vec4 tcBrushParams[];
layout(location = 8) out vec4 tcBrushColor[];
layout(location = 9) flat out ivec4 tcTerrainLightIds[];
layout(location = 10) flat out int tcTerrainInstance[];

float terrainTessFactor(vec3 a, vec3 b) {
	vec3 edgeCenter = (a + b) * 0.5;
	float distanceToCamera = distance(edgeCenter, cameraPosition.xyz);
	float t = clamp((TERRAIN_TESS_FAR_DISTANCE - distanceToCamera) /
		(TERRAIN_TESS_FAR_DISTANCE - TERRAIN_TESS_NEAR_DISTANCE), 0.0, 1.0);
	return mix(TERRAIN_TESS_MIN_LEVEL, TERRAIN_TESS_MAX_LEVEL, t * t);
}

void main() {
	gl_out[gl_InvocationID].gl_Position = gl_in[gl_InvocationID].gl_Position;
	tcColor[gl_InvocationID] = inColor[gl_InvocationID];
	tcFlags[gl_InvocationID] = inFlags[gl_InvocationID];
	tcPos[gl_InvocationID] = inPos[gl_InvocationID];
	tcTexCoords[gl_InvocationID] = inTexCoords[gl_InvocationID];
	tcNormal[gl_InvocationID] = inNormal[gl_InvocationID];
	tcViewDir[gl_InvocationID] = inViewDir[gl_InvocationID];
	tcBrushCenterRadius[gl_InvocationID] = inBrushCenterRadius[gl_InvocationID];
	tcBrushParams[gl_InvocationID] = inBrushParams[gl_InvocationID];
	tcBrushColor[gl_InvocationID] = inBrushColor[gl_InvocationID];
	tcTerrainLightIds[gl_InvocationID] = inTerrainLightIds[gl_InvocationID];
	tcTerrainInstance[gl_InvocationID] = inTerrainInstance[gl_InvocationID];

	if (gl_InvocationID == 0) {
		float e0 = terrainTessFactor(inPos[1], inPos[2]);
		float e1 = terrainTessFactor(inPos[2], inPos[0]);
		float e2 = terrainTessFactor(inPos[0], inPos[1]);
		gl_TessLevelOuter[0] = e0;
		gl_TessLevelOuter[1] = e1;
		gl_TessLevelOuter[2] = e2;
		gl_TessLevelInner[0] = (e0 + e1 + e2) / 3.0;
	}
}
