#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;

// Color, position, normal
layout(binding = 1) uniform sampler2D textures[3];

layout(location = 0) out vec4 outColor;

const uint outlineFlag = 0x00000001u;
const uint occupiedFlag = 0x40000000u;
const vec4 outlineColor = vec4(251.0/255.0, 84.0/255.0, 43.0/255.0, 1.0);

uint gBufferFlags(vec4 pos) {
	return floatBitsToUint(pos.w);
}

bool hasFlag(vec4 pos, uint flag) {
	return (gBufferFlags(pos) & flag) != 0u;
}

float cameraDepth(vec4 pos) {
	return abs((view * vec4(pos.xyz, 1.0)).z);
}

bool shouldOutlineAgainst(vec4 pos, vec4 neighbor) {
	if (hasFlag(neighbor, outlineFlag)) {
		return false;
	}
	if (!hasFlag(neighbor, occupiedFlag)) {
		return true;
	}
	return cameraDepth(pos) <= cameraDepth(neighbor) + 0.0005;
}

void main() {
	vec2 uv = gl_FragCoord.xy / screenSize;
	vec4 pos = texture(textures[1], uv);
	vec2 dx = vec2(1.0 / screenSize.x, 0.0);
	vec2 dy = vec2(0.0, 1.0 / screenSize.y);
	bool outline = false;
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
	if (hasFlag(pos, outlineFlag)) {
		for (int i = 0; i < 8; i++) {
			if (shouldOutlineAgainst(pos, texture(textures[1], uv + offsets[i]))) {
				outline = true;
				break;
			}
		}
	}
	if (outline) {
		outColor = outlineColor;
	} else {
		outColor = texture(textures[0], fragTexCoords) * fragColor;
	}
}
