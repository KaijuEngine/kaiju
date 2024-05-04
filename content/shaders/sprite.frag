#version 450

layout(location = 0) in vec4 fragColor;
layout(location = 4) in vec2 fragTexCoord;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
layout(location = 1) out float reveal;

void main(void) {
	vec4 texColor = texture(texSampler, fragTexCoord) * fragColor;
	vec4 unWeightedColor = texColor;
#include "inc_fragment_oit_block.inl"
}
