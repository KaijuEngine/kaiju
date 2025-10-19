#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;
layout(location = 2) in vec3 fragTangentViewPos;
layout(location = 3) in vec3 fragTangentFragPos;
layout(location = 4) in vec3 fragLightTPos[NR_LIGHTS];
layout(location = 8) in vec3 fragLightTDir[NR_LIGHTS];
layout(location = 12) in vec4 fragPosLightSpace[NR_LIGHTS];
layout(location = 16) in vec3 fragPos;
layout(location = 17) in vec3 fragNormal;
layout(location = 19) in flat int lightCount;
layout(location = 20) in flat int lightIndexes[NR_LIGHTS];

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

layout(binding = 1) uniform sampler2D textures;

void main() {
	vec4 unWeightedColor = texture(textures, fragTexCoords) * vec4(0.2313, 0.407, 0.08, 1.0);
#include "inc_fragment_oit_block.inl"
}
