#version 460
#define TESS_EVALUATION_SHADER

#include "kaiju.glsl"

#define TERRAIN_PHONG_STRENGTH 0.65

layout(triangles, fractional_odd_spacing, cw) in;

layout(location = 0) in vec4 tcColor[];
layout(location = 1) flat in uint tcFlags[];
layout(location = 2) in vec3 tcPos[];
layout(location = 3) in vec2 tcTexCoords[];
layout(location = 4) in vec3 tcNormal[];
layout(location = 5) in vec3 tcViewDir[];
layout(location = 6) in vec4 tcSlopeParams[];
layout(location = 7) in vec4 tcGrassTint[];
layout(location = 8) in vec4 tcRockTint[];
layout(location = 9) in vec4 tcLightDirectionAmbient[];
layout(location = 10) in vec4 tcLightColorDiffuse[];
layout(location = 11) in vec4 tcMaterialParams[];
layout(location = 12) in vec4 tcBrushCenterRadius[];
layout(location = 13) in vec4 tcBrushParams[];
layout(location = 14) in vec4 tcBrushColor[];

layout(location = 0) out vec4 fragColor;
layout(location = 1) flat out uint fragFlags;
layout(location = 2) out vec3 fragPos;
layout(location = 3) out vec2 fragTexCoords;
layout(location = 4) out vec3 fragNormal;
layout(location = 5) out vec3 fragViewDir;
layout(location = 6) out vec4 fragSlopeParams;
layout(location = 7) out vec4 fragGrassTint;
layout(location = 8) out vec4 fragRockTint;
layout(location = 9) out vec4 fragLightDirectionAmbient;
layout(location = 10) out vec4 fragLightColorDiffuse;
layout(location = 11) out vec4 fragMaterialParams;
layout(location = 12) out vec4 fragBrushCenterRadius;
layout(location = 13) out vec4 fragBrushParams;
layout(location = 14) out vec4 fragBrushColor;

vec3 interpolateVec3(vec3 a, vec3 b, vec3 c, vec3 weights) {
	return a * weights.x + b * weights.y + c * weights.z;
}

vec4 interpolateVec4(vec4 a, vec4 b, vec4 c, vec3 weights) {
	return a * weights.x + b * weights.y + c * weights.z;
}

vec2 interpolateVec2(vec2 a, vec2 b, vec2 c, vec3 weights) {
	return a * weights.x + b * weights.y + c * weights.z;
}

vec3 projectOntoTangentPlane(vec3 point, vec3 planePoint, vec3 normal) {
	return point - normal * dot(point - planePoint, normal);
}

vec3 phongTerrainPosition(vec3 linearPosition, vec3 weights) {
	vec3 n0 = normalize(tcNormal[0]);
	vec3 n1 = normalize(tcNormal[1]);
	vec3 n2 = normalize(tcNormal[2]);
	vec3 p0 = projectOntoTangentPlane(linearPosition, tcPos[0], n0);
	vec3 p1 = projectOntoTangentPlane(linearPosition, tcPos[1], n1);
	vec3 p2 = projectOntoTangentPlane(linearPosition, tcPos[2], n2);
	vec3 phongPosition = interpolateVec3(p0, p1, p2, weights);
	return mix(linearPosition, phongPosition, TERRAIN_PHONG_STRENGTH);
}

void main() {
	vec3 weights = gl_TessCoord;
	vec3 linearPosition = interpolateVec3(tcPos[0], tcPos[1], tcPos[2], weights);
	vec3 normal = normalize(interpolateVec3(tcNormal[0], tcNormal[1], tcNormal[2], weights));
	vec3 position = phongTerrainPosition(linearPosition, weights);

	fragColor = interpolateVec4(tcColor[0], tcColor[1], tcColor[2], weights);
	fragFlags = tcFlags[0];
	fragPos = position;
	fragTexCoords = interpolateVec2(tcTexCoords[0], tcTexCoords[1], tcTexCoords[2], weights);
	fragNormal = normal;
	fragViewDir = cameraPosition.xyz - position;
	fragSlopeParams = interpolateVec4(tcSlopeParams[0], tcSlopeParams[1], tcSlopeParams[2], weights);
	fragGrassTint = interpolateVec4(tcGrassTint[0], tcGrassTint[1], tcGrassTint[2], weights);
	fragRockTint = interpolateVec4(tcRockTint[0], tcRockTint[1], tcRockTint[2], weights);
	fragLightDirectionAmbient = interpolateVec4(tcLightDirectionAmbient[0], tcLightDirectionAmbient[1], tcLightDirectionAmbient[2], weights);
	fragLightColorDiffuse = interpolateVec4(tcLightColorDiffuse[0], tcLightColorDiffuse[1], tcLightColorDiffuse[2], weights);
	fragMaterialParams = interpolateVec4(tcMaterialParams[0], tcMaterialParams[1], tcMaterialParams[2], weights);
	fragBrushCenterRadius = interpolateVec4(tcBrushCenterRadius[0], tcBrushCenterRadius[1], tcBrushCenterRadius[2], weights);
	fragBrushParams = interpolateVec4(tcBrushParams[0], tcBrushParams[1], tcBrushParams[2], weights);
	fragBrushColor = interpolateVec4(tcBrushColor[0], tcBrushColor[1], tcBrushColor[2], weights);

	gl_Position = projection * view * vec4(position, 1.0);
}
