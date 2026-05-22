#version 450

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec4 fragBGColor;
layout(location = 2) in vec4 fragSize2D;
layout(location = 3) in vec4 fragBorderRadius;
layout(location = 4) in vec4 fragBorderSize;
layout(location = 5) in mat4 fragBorderColorsLTRB;
layout(location = 9) in vec2 fragTexCoord;
layout(location = 10) in vec2 fragNineSliceEdgeLen;
layout(location = 11) in vec4 fragUvs;
layout(location = 12) in vec4 fragOutlineColor;
layout(location = 13) in vec2 fragOutlineSize;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

const float edgeSoftness = 2.0;

float processAxis(float coord, float border, float ratio) {
	float len = border * ratio;
	float lScale = 1.0 - len * 2.0;
	float bScale = 1.0 - (border * 2.0);
	float res = (coord - len) * (bScale / lScale) + border;
	if (coord < len) {
		res = coord / ratio;
	} else if (coord > 1.0 - len) {
		res = 1.0 - ((1.0 - coord) / ratio);
	}
	return res;
}

//https://www.shadertoy.com/view/tltXDl
float roundedBoxSDF(vec2 centerPosition, vec2 size, vec4 radius) {
	radius.xy = (centerPosition.x > 0.0) ? radius.xw : radius.yz;
    radius.x  = (centerPosition.y < 0.0) ? radius.x  : radius.y;
    vec2 q = abs(centerPosition) - size + radius.x;
    return min(max(q.x,q.y),0.0) + length(max(q,0.0)) - radius.x;
}

void main(void) {
	vec2 normUV = (fragTexCoord - fragUvs.xy) / fragUvs.zw;
	float outlineWidth = max(0.0, fragOutlineSize.x);
	float outlineOffset = max(0.0, fragOutlineSize.y);
	float outlineOutset = outlineWidth + outlineOffset;
	vec2 dimensions = fragSize2D.xy;
	vec2 expandedDimensions = dimensions + vec2(outlineOutset * 2.0);
	vec2 expandedPixPos = normUV * expandedDimensions;
	vec2 pixPos = expandedPixPos - vec2(outlineOutset);
	vec4 unWeightedColor = vec4(0.0);
	bool insidePanel = pixPos.x >= 0.0 && pixPos.y >= 0.0 && pixPos.x <= dimensions.x && pixPos.y <= dimensions.y;
	if (insidePanel) {
		vec2 panelUV = pixPos / dimensions;
		vec2 scaledNormUV = vec2(
			processAxis(panelUV.x, fragNineSliceEdgeLen.x / fragSize2D.z, fragSize2D.z / fragSize2D.x),
			processAxis(panelUV.y, fragNineSliceEdgeLen.y / fragSize2D.w, fragSize2D.w / fragSize2D.y)
		);
		vec2 newUV = fragUvs.xy + scaledNormUV * fragUvs.zw;
		unWeightedColor = texture(texSampler, newUV) * fragColor;
		// Border
		vec2 size = dimensions / 2.0;
		vec2 centerPixPos = size-pixPos;
		// Pre-select what color in the color matrix should be used, the order
		// of colors are [0] = left, [1] = top, [2] = right, [3] = bottom
		int sideIdx = 0;
		if (pixPos.x < fragBorderSize.x) {
			sideIdx = 0;
		} else if (pixPos.y < fragBorderSize.y) {
			sideIdx = 1;
		} else if (pixPos.x > dimensions.x-fragBorderSize.z) {
			sideIdx = 2;
		} else if (pixPos.y > dimensions.y-fragBorderSize.w) {
			sideIdx = 3;
		}
        vec4 borderColor = fragBorderColorsLTRB[sideIdx];
		// Border radius
		float dist = roundedBoxSDF(centerPixPos, size, fragBorderRadius+fragBorderSize);
		float smoothedAlpha = 1.0-smoothstep(0.0, edgeSoftness, dist);
		// Border size
		centerPixPos.x += fragBorderSize.x/2.0-fragBorderSize.z/2.0;
		centerPixPos.y += fragBorderSize.y/2.0-fragBorderSize.w/2.0;
		size.x -= (fragBorderSize.x+fragBorderSize.z)/2.0;
		size.y -= (fragBorderSize.y+fragBorderSize.w)/2.0;
		float borderDistance = roundedBoxSDF(centerPixPos, size, fragBorderRadius);
		float innerMask = 1.0-smoothstep(0.0, edgeSoftness, borderDistance);
		float smoothedBorderAlpha = 1.0 - innerMask;
		// Border color
		unWeightedColor = mix(unWeightedColor, borderColor, smoothedBorderAlpha);
		unWeightedColor.a = smoothedAlpha * unWeightedColor.a;
	} else if (outlineWidth > 0.0 && fragOutlineColor.a > 0.0) {
		vec2 outside = max(max(-pixPos, pixPos - dimensions), vec2(0.0));
		float outsideDistance = max(outside.x, outside.y);
		float outerAlpha = 1.0 - smoothstep(outlineOffset + outlineWidth, outlineOffset + outlineWidth + edgeSoftness, outsideDistance);
		float innerAlpha = smoothstep(outlineOffset, outlineOffset + edgeSoftness, outsideDistance);
		unWeightedColor = fragOutlineColor;
		unWeightedColor.a *= outerAlpha * innerAlpha;
	}
#include "inc_fragment_oit_block.inl"
}
