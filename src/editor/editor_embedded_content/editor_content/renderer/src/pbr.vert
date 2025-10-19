#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 vertColors;
layout(location = LOCATION_START+1) in float metallic;
layout(location = LOCATION_START+2) in float roughness;
layout(location = LOCATION_START+3) in float emissive;
layout(location = LOCATION_START+4) in float light0;
layout(location = LOCATION_START+5) in float light1;
layout(location = LOCATION_START+6) in float light2;
layout(location = LOCATION_START+7) in float light3;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;
layout(location = 2) out vec3 fragTangentViewPos;
layout(location = 3) out vec3 fragTangentFragPos;
layout(location = 4) out vec3 fragLightTPos[NR_LIGHTS];
layout(location = 8) out vec3 fragLightTDir[NR_LIGHTS];
layout(location = 12) out vec4 fragPosLightSpace[NR_LIGHTS];
layout(location = 16) out vec3 fragPos;
layout(location = 17) out vec3 fragNormal;
layout(location = 18) out float fragMetallic;
layout(location = 19) out float fragRoughness;
layout(location = 20) out float fragEmissive;
layout(location = 21) out flat int lightCount;
layout(location = 22) out flat int lightIndexes[NR_LIGHTS];

void main() {
	fragColor = vertColors * Color;
	fragTexCoords = UV0;
	fragPos = vec3(model * vec4(Position, 1.0));
	gl_Position = projection * view * model * vec4(Position, 1.0);

	mat3 nmlMat = transpose(inverse(mat3(model)));
	vec3 T = normalize(nmlMat * Tangent.xyz);
	vec3 N = normalize(nmlMat * Normal);
	// re-orthogonalize T with respect to N
	T = normalize(T - dot(T, N) * N);
	// then retrieve perpendicular vector B with the cross product of T and N
	vec3 B = cross(N, T);
	mat3 TBN = transpose(mat3(T, B, N));

#ifdef MULTI_LIGHT
	int indexes[NR_LIGHTS] = { light0, light1, light2, light3 };
	for (int i = 0; i < NR_LIGHTS; ++i) {
		int idx = clamp(indexes[i], 0, MAX_LIGHTS - 1);
		fragLightTPos[i] = TBN * vertLights[idx].position;
		fragLightTDir[i] = TBN * normalize(vertLights[idx].direction);
		fragPosLightSpace[i] = vertLights[idx].matrix[0] * vec4(fragPos, 1.0);
	}
#else
	fragLightTPos[0] = TBN * vertLights[0].position;
	fragLightTDir[0] = TBN * normalize(vertLights[0].direction);
	fragPosLightSpace[0] = vertLights[0].matrix[0] * vec4(fragPos, 1.0);
#endif

	fragTangentViewPos = TBN * cameraPosition.xyz;
	fragTangentFragPos = TBN * fragPos;
	fragNormal = N;
	fragMetallic = metallic;
	fragRoughness = roughness;
	fragEmissive = emissive;
#ifdef MULTI_LIGHT
	lightCount = min(light0 + 1, 1) + min(light1 + 1, 1) + min(light2 + 1, 1) + min(light3 + 1, 1);
	for (int i = 0; i < NR_LIGHTS; ++i)
		lightIndexes[i] = indexes[i];
#endif
}
