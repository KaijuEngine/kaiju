#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;

// Color, position, normal
layout(binding = 1) uniform sampler2D textures[3];

layout(location = 0) out vec4 outColor;

const uint outlineFlag = 0x00000001u;
// Matches shader_data_registry.ShaderDataStandardOutlineGroupMask
const uint outlineGroupMask = 0x007FFF00u;
const uint occupiedFlag = 0x40000000u;
const uint selectedFlags = outlineFlag | occupiedFlag;
const vec3 outlineColor = vec3(251.0/255.0, 84.0/255.0, 43.0/255.0);
// Pixels of solid outline drawn outside of the selected silhouette, the pixel
// after that is faded to anti-alias the outer edge of the line
const float outlineWidth = 2.0;
const int outlineReach = 2;
// Fraction of camera depth that two surfaces can differ by and still be
// considered to be at the same depth (contact points, coplanar surfaces)
const float depthTolerance = 0.001;

uint gBufferFlags(vec4 pos) {
	return floatBitsToUint(pos.w);
}

bool isOccupied(uint flags) {
	return (flags & occupiedFlag) != 0u;
}

bool isSelected(uint flags) {
	return (flags & selectedFlags) == selectedFlags;
}

// Selection identity, unselected surfaces return 0 which no selection uses
uint selectionKey(uint flags) {
	return isSelected(flags) ? (flags & outlineGroupMask) | outlineFlag : 0u;
}

float cameraDepth(vec4 pos) {
	return abs((view * vec4(pos.xyz, 1.0)).z);
}

bool isCoplanar(vec4 aPos, ivec2 aTexel, vec4 bPos, ivec2 bTexel) {
	vec3 aNml = texelFetch(textures[2], aTexel, 0).xyz;
	vec3 bNml = texelFetch(textures[2], bTexel, 0).xyz;
	float tolerance = max(cameraDepth(aPos), cameraDepth(bPos)) * depthTolerance;
	vec3 delta = bPos.xyz - aPos.xyz;
	return abs(dot(aNml, delta)) <= tolerance && abs(dot(bNml, delta)) <= tolerance;
}

// True when the surface in front is not hidden behind the surface in back, so
// the silhouette of front should be drawn over back
bool isVisibleOver(vec4 frontPos, ivec2 frontTexel, vec4 backPos, ivec2 backTexel) {
	if (!isOccupied(gBufferFlags(backPos))) {
		return true;
	}
	if (cameraDepth(frontPos) <= cameraDepth(backPos) * (1.0 + depthTolerance)) {
		return true;
	}
	return isCoplanar(frontPos, frontTexel, backPos, backTexel);
}

void main() {
	ivec2 gBufferMax = textureSize(textures[1], 0) - 1;
	ivec2 texel = clamp(ivec2(gl_FragCoord.xy), ivec2(0), gBufferMax);
	vec4 centerPos = texelFetch(textures[1], texel, 0);
	uint centerKey = selectionKey(gBufferFlags(centerPos));
	// Squared distance to the closest selected pixel whose silhouette this
	// pixel sits outside of
	float nearest = 1e9;
	// How much of the immediate neighborhood is outside of this pixel's
	// selection, used to soften the inside edge of the line
	float innerCoverage = 0.0;
	for (int y = -outlineReach; y <= outlineReach; y++) {
		for (int x = -outlineReach; x <= outlineReach; x++) {
			float dist2 = float(x*x + y*y);
			if (dist2 == 0.0 || dist2 >= (outlineWidth + 1.0) * (outlineWidth + 1.0)) {
				continue;
			}
			ivec2 neighborTexel = clamp(texel + ivec2(x, y), ivec2(0), gBufferMax);
			vec4 neighborPos = texelFetch(textures[1], neighborTexel, 0);
			uint neighborKey = selectionKey(gBufferFlags(neighborPos));
			if (neighborKey == centerKey) {
				continue;
			}
			if (neighborKey != 0u && dist2 < nearest
				&& isVisibleOver(neighborPos, neighborTexel, centerPos, texel)) {
				nearest = dist2;
			}
			if (centerKey != 0u && abs(x) <= 1 && abs(y) <= 1
				&& isVisibleOver(centerPos, texel, neighborPos, neighborTexel)) {
				innerCoverage += 1.0 / 8.0;
			}
		}
	}
	float outlineAlpha = max(
		clamp(outlineWidth + 1.0 - sqrt(nearest), 0.0, 1.0), innerCoverage);
	vec4 sceneColor = texture(textures[0], fragTexCoords) * fragColor;
	// Composite the outline over the scene color, output is blended with
	// straight alpha so un-premultiply the result
	float alpha = outlineAlpha + sceneColor.a * (1.0 - outlineAlpha);
	vec3 color = outlineColor * outlineAlpha
		+ sceneColor.rgb * sceneColor.a * (1.0 - outlineAlpha);
	outColor = alpha > 0.0 ? vec4(color / alpha, alpha) : vec4(0.0);
}
