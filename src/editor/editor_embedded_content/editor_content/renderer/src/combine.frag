#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;

// Color, position, normal
layout(binding = 1) uniform sampler2D textures[3];

layout(location = 0) out vec4 outColor;

bool hasOutlineFlag(float value) {
    uint alphaBits = floatBitsToUint(value);
    return (alphaBits & 0x0001) != 0;
}

void main() {
	vec2 uv = gl_FragCoord.xy / screenSize;
	vec4 pos = texture(textures[1], uv);
	float mask = hasOutlineFlag(pos.w) ? 1.0 : 0.0;
	vec2 dx = vec2(1.0 / screenSize.x, 0.0);
	vec2 dy = vec2(0.0, 1.0 / screenSize.y);
	float sobel = 0.0;
	vec2 offsets[8] = vec2[](
		-dx - dy,  // top-left
		-dy,       // top (0, -dy.y)
		 dx - dy,  // top-right
		-dx,       // left (-dx.x, 0)
		 dx,       // right (dx.x, 0)
		-dx + dy,  // bottom-left
		 dy,       // bottom (0, dy.y)
		 dx + dy  // bottom-right
	);
	for (int i = 0; i < 8; i++) {
		float nmask = hasOutlineFlag(texture(textures[1], uv + offsets[i]).w) ? 1.0 : 0.0;
		sobel += abs(mask - nmask);
	}
	if (sobel > 0.0) {
		outColor = vec4(251.0/255.0, 84.0/255.0, 43.0/255.0, 1.0);
	} else {
		outColor = texture(textures[0], fragTexCoords) * fragColor;
	}
}
