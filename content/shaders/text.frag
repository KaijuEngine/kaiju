#version 450
//#version 300 es
//precision mediump float;

#ifdef VULKAN
	layout(location = 0) in vec4 fragColor;
	layout(location = 1) in vec4 fragBGColor;
	layout(location = 2) in vec2 fragTexCoord;
	layout(location = 3) in vec2 fragPxRange;

	layout(binding = 1) uniform sampler2D texSampler;
#else
	in vec4 fragColor;
	in vec4 fragBGColor;
	in vec2 fragTexCoord;
	in vec2 fragPxRange;

	uniform sampler2D texSampler;
#endif

layout(location = 0) out vec4 outColor;
layout(location = 1) out float reveal;

float median(float r, float g, float b) {
	return max(min(r, g), min(max(r, g), b));
}

float screenPxRange() {
	vec2 unitRange = fragPxRange / vec2(textureSize(texSampler, 0));
	vec2 screenTexSize = vec2(1.0) / fwidth(fragTexCoord);
	return max(0.5 * dot(unitRange, screenTexSize), 1.0);
}

void main() {
	vec3 msd = texture(texSampler, fragTexCoord).rgb;
	float sd = median(msd.r, msd.g, msd.b);
	float screenPxDistance = screenPxRange() * (sd - 0.5);
	float opacity = clamp(screenPxDistance + 0.5, 0.0, 1.0);
	vec4 unWeightedColor = mix(fragBGColor, fragColor, opacity);
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