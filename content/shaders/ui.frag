#version 450

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec4 fragBGColor;
layout(location = 2) in vec4 fragSize2D;
layout(location = 3) in vec4 fragBorderRadius;
layout(location = 4) in vec4 fragBorderSize;
layout(location = 5) in mat4 fragBorderColor;
layout(location = 9) in vec2 fragTexCoord;
layout(location = 10) in vec2 fragBorderLen;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
layout(location = 1) out float reveal;

void main(void) {
	vec4 texColor = texture(texSampler, fragTexCoord) * fragColor;
	vec4 unWeightedColor = texColor;
#include "inc_fragment_oit_block.inl"
}
