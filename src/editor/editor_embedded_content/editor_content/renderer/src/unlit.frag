#version 460

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;
layout(location = 2) in flat uint fragFlags;
layout(location = 3) in vec3 fragPos;
layout(location = 4) in vec3 fragNormal;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#else
layout(location = 1) out vec4 outPosition;
layout(location = 2) out vec4 outNormal;
#endif

void main() {
    vec4 unWeightedColor = texture(texSampler, fragTexCoords) * fragColor;
#ifndef OIT
    outPosition = vec4(fragPos, uintBitsToFloat(fragFlags));
    outNormal = vec4(normalize(fragNormal), 0.0);
#endif
#include "inc_fragment_oit_block.inl"
}
