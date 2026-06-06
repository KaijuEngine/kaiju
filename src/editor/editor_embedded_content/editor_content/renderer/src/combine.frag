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
const float selectedDepthEdgeThreshold = 0.02;
const float selectedNormalEdgeThreshold = 0.92;

uint gBufferFlags(vec4 pos) {
	return floatBitsToUint(pos.w);
}

bool hasFlag(vec4 pos, uint flag) {
	return (gBufferFlags(pos) & flag) != 0u;
}

float cameraDepth(vec4 pos) {
	return abs((view * vec4(pos.xyz, 1.0)).z);
}

bool shouldOutlineAgainst(vec4 pos, vec4 normal, vec4 neighbor, vec4 neighborNormal) {
	if (!hasFlag(neighbor, occupiedFlag)) {
		return true;
	}
	if (hasFlag(neighbor, outlineFlag)) {
		float depthDelta = abs(cameraDepth(pos) - cameraDepth(neighbor));
		float normalDot = dot(normalize(normal.xyz), normalize(neighborNormal.xyz));
		return depthDelta > selectedDepthEdgeThreshold && normalDot < selectedNormalEdgeThreshold;
	}
	return cameraDepth(pos) <= cameraDepth(neighbor) + 0.0005;
}

void main() {
	vec2 gBufferSize = vec2(textureSize(textures[1], 0));
	vec2 uv = gl_FragCoord.xy / gBufferSize;
	vec4 pos = texture(textures[1], uv);
	vec4 normal = texture(textures[2], uv);
	vec2 dx = vec2(1.0 / gBufferSize.x, 0.0);
	vec2 dy = vec2(0.0, 1.0 / gBufferSize.y);
	bool outline = false;
	vec2 offsets[16] = vec2[](
		-dx - dy,  // top-left
		-dy,       // top (0, -dy.y)
		 dx - dy,  // top-right
		-dx,       // left (-dx.x, 0)
		 dx,       // right (dx.x, 0)
		-dx + dy,  // bottom-left
		 dy,       // bottom (0, dy.y)
		 dx + dy,  // bottom-right
		(-dx - dy) * 2.0,
		-dy * 2.0,
		 (dx - dy) * 2.0,
		-dx * 2.0,
		 dx * 2.0,
		(-dx + dy) * 2.0,
		 dy * 2.0,
		 (dx + dy) * 2.0
	);
	if (hasFlag(pos, outlineFlag)) {
		for (int i = 0; i < 16; i++) {
			vec2 offsetUV = uv + offsets[i];
			if (shouldOutlineAgainst(pos, normal, texture(textures[1], offsetUV), texture(textures[2], offsetUV))) {
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
