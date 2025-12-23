#version 460

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

void main() {
	vec4 tex = texture(texSampler, fragTexCoords) * fragColor;
	// vec4 unWeightedColor = texture(texSampler, fragTexCoords) * fragColor;
	// float v = max(tex.r - 0.267, 0.0) * 1.364;
	float v = tex.r;
	vec4 unWeightedColor = vec4(v * fragColor.rgb, v * fragColor.a);
#include "inc_fragment_oit_block.inl"
}
