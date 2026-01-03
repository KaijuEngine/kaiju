#version 460

#include "inc_vertex.inl"

#define MAX_JOINTS			50

#ifdef SKINNING
layout(set = 0, binding = 2) readonly buffer SkinnedSSBO {
	mat4 jointTransforms[][MAX_JOINTS];
};
#endif

layout(location = LOCATION_START) in vec4 vertColors;
layout(location = LOCATION_START+1) in float metallic;
layout(location = LOCATION_START+2) in float roughness;
layout(location = LOCATION_START+3) in float emissive;
layout(location = LOCATION_START+4) in int lightIds[NR_LIGHTS];
layout(location = LOCATION_START+4+NR_LIGHTS+0) in uint flags;

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
layout(location = 26) out uint fragFlags;

void main() {
	fragColor = vertColors * Color;
	fragFlags = flags;
	fragTexCoords = UV0;
	fragPos = vec3(model * vec4(Position, 1.0));
#ifdef SKINNING
	vec4 pos = vec4(Position, 1.0);
	mat4 skinMatrix = JointWeights.x * jointTransforms[gl_InstanceIndex][JointIds.x]
					+ JointWeights.y * jointTransforms[gl_InstanceIndex][JointIds.y]
					+ JointWeights.z * jointTransforms[gl_InstanceIndex][JointIds.z]
					+ JointWeights.w * jointTransforms[gl_InstanceIndex][JointIds.w];
	vec4 wp = skinMatrix * pos;
	gl_Position = projection * view * wp;
#else
	gl_Position = projection * view * model * vec4(Position, 1.0);
#endif
	mat3 nmlMat = transpose(inverse(mat3(model)));
	vec3 T = normalize(nmlMat * Tangent.xyz);
	vec3 N = normalize(nmlMat * Normal);
	// re-orthogonalize T with respect to N
	T = normalize(T - dot(T, N) * N);
	// then retrieve perpendicular vector B with the cross product of T and N
	vec3 B = cross(N, T);
	mat3 TBN = transpose(mat3(T, B, N));
	lightCount = 0;
	for (int i = 0; i < NR_LIGHTS; ++i) {
		int idx = lightIds[i];
		if (idx < 0) {
			continue;
		}
		idx = min(idx, MAX_LIGHTS - 1);
		fragLightTPos[lightCount] = TBN * vertLights[idx].position;
		fragLightTDir[lightCount] = TBN * normalize(vertLights[idx].direction);
		fragPosLightSpace[lightCount] = vertLights[idx].matrix[0] * vec4(fragPos, 1.0);
		lightIndexes[lightCount] = idx;
		lightCount++;
	}
	fragTangentViewPos = TBN * cameraPosition.xyz;
	fragTangentFragPos = TBN * fragPos;
	fragNormal = N;
	fragMetallic = metallic;
	fragRoughness = roughness;
	fragEmissive = emissive;
}
