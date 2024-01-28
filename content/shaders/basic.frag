#version 460
//#version 300 es
//precision mediump float;

#ifdef VULKAN
	layout(location = 0) in vec4 fragColor;
	layout(location = 1) in vec2 fragTexCoords;

	layout(binding = 1) uniform sampler2D texSampler;
#else
	in vec4 fragColor;
	in vec2 fragTexCoords;

	uniform sampler2D texSampler;
#endif

layout(location = 0) out vec4 outColor;
layout(location = 1) out float reveal;

void main() {
    vec4 unWeightedColor = texture(texSampler, fragTexCoords) * fragColor;
#ifdef OIT
	float distWeight = clamp(0.03 / (1e-5 + pow(gl_FragCoord.z / 200.0, 4.0)), 1e-2, 3e3);
	float alphaWeight = min(1.0, max(max(unWeightedColor.r, unWeightedColor.g),
	max(unWeightedColor.b, unWeightedColor.a)) * 40.0 + 0.01);
	alphaWeight *= alphaWeight;
	float weight = alphaWeight * distWeight;
	outColor = vec4(unWeightedColor.rgb * unWeightedColor.a, unWeightedColor.a) * weight;
	reveal = unWeightedColor.a;
#else
	if (unWeightedColor.a < (1.0 - 0.0001))
		discard;
	outColor = unWeightedColor;
#endif
}