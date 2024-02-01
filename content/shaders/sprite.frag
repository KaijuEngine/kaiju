#version 450
//#version 300 es
//precision mediump float;

#ifdef VULKAN
	layout(location = 0) in vec4 fragColor;
	layout(location = 1) in vec4 fragBGColor;
	layout(location = 2) in vec4 fragSize2D;
	layout(location = 3) in vec4 fragScissor;
	layout(location = 4) in vec2 fragTexCoord;
	layout(location = 5) in vec2 fragBorderLen;

	layout(binding = 1) uniform sampler2D texSampler;
#else
	in vec4 fragColor;
	in vec4 fragBGColor;
	in vec4 fragSize2D;
	in vec4 fragScissor;
	in vec2 fragTexCoord;
	in vec2 fragBorderLen;

	uniform sampler2D texSampler;
#endif

layout(location = 0) out vec4 outColor;
layout(location = 1) out float reveal;

void main(void) {
#ifndef VULKAN
	if (gl_FragCoord.x < fragScissor.x || gl_FragCoord.x > fragScissor.z ||
		gl_FragCoord.y < fragScissor.y || gl_FragCoord.y > fragScissor.w)
	{
		discard;
	}
#endif
	vec4 texColor = texture(texSampler, fragTexCoord) * fragColor;
	vec4 unWeightedColor = texColor;
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
