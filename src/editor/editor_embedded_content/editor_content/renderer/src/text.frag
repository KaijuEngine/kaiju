#version 450

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec4 fragBGColor;
layout(location = 2) in vec2 fragTexCoord;
layout(location = 3) in vec2 fragPxRange;
layout(location = 4) in vec2 fragTexRange;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

float median(float r, float g, float b) {
	return max(min(r, g), min(max(r, g), b));
}

float screenPxRange() {
	vec2 texSize = vec2(textureSize(texSampler, 0));
	vec2 unitRange = fragPxRange / texSize;
	vec2 screenTexSize = vec2(1.0) / max(fwidth(fragTexCoord), vec2(0.000001));
	return max(0.5 * dot(unitRange, screenTexSize), 1.0);
}

void main() {
	vec3 msdfColor = texture(texSampler, fragTexCoord).rgb;
	float dist = median(msdfColor.r, msdfColor.g, msdfColor.b) - 0.5;
	float opacity = clamp(dist * screenPxRange() + 0.5, 0.0, 1.0);

	// A negative background alpha marks opaque cutout text. Discard the quad
	// outside the MSDF glyph boundary and write a fully opaque foreground pixel
	// inside it. This keeps text opaque even when its parent is transparent.
	if (fragBGColor.a < 0.0) {
		if (opacity < 0.5) {
			discard;
		}
		outColor = vec4(fragColor.rgb, 1.0);
		return;
	}

	vec4 unWeightedColor = mix(fragBGColor, fragColor, opacity);
#include "inc_fragment_oit_block.inl"
}
