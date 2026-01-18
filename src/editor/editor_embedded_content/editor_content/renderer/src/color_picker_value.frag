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

const vec4 WHITE = vec4(1.0, 1.0, 1.0, 1.0);
const vec4 BLACK = vec4(0.0, 0.0, 0.0, 1.0);
const vec4 GRAY = vec4(0.5, 0.5, 0.5, 1.0);

void main(void) {
	outColor = mix(mix(fragColor, WHITE, fragTexCoord.x), BLACK, fragTexCoord.y);
}
