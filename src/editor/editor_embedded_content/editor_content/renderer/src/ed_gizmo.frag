#version 460

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

void main() {
    vec4 unWeightedColor = texture(texSampler, fragTexCoords) * fragColor;
#include "inc_fragment_oit_block.inl"
}
