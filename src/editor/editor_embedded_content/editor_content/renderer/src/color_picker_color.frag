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

const vec4 RED = vec4(1.0, 0.0, 0.0, 1.0);
const vec4 GREEN = vec4(0.0, 1.0, 0.0, 1.0);
const vec4 BLUE = vec4(0.0, 0.0, 1.0, 1.0);
const float full = 0.33333;
const float halfFull = full * 0.5;
const float iFull = 1.0 - full;

void main(void) {
	vec4 texColor = vec4(0.0, 0.0, 0.0, 1.0);
	float x = fragTexCoord.x;
	if (x <= full) {
		if (x < halfFull) {
			texColor.r = 1.0;
			texColor.g = x / halfFull;
		} else {
			texColor.r = 1.0 - (((x / full) - 0.5) * 2.0);
			texColor.g = 1.0;
		}
	} else if (x >= iFull) {
		x -= iFull;
		if (x < halfFull) {
			texColor.b = 1.0;
			texColor.r = x / halfFull;
		} else {
			texColor.b = 1.0 - (((x / full) - 0.5) * 2.0);
			texColor.r = 1.0;
		}
	} else {
		x -= full;
		if (x < halfFull) {
			texColor.g = 1.0;
			texColor.b = x / halfFull;
		} else {
			texColor.g = 1.0 - (((x / full) - 0.5) * 2.0);
			texColor.b = 1.0;
		}
	}
	outColor = texColor;
}
