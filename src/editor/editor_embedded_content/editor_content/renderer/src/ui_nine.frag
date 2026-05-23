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

float edgeProximity(float edgeDistance, float borderWidth) {
	if (borderWidth <= 0.0) {
		return 100000.0;
	}
	return edgeDistance / borderWidth;
}

float sdfInsideAlpha(float dist) {
	float aa = max(fwidth(dist), edgeSoftness);
	return 1.0 - smoothstep(0.0, aa, dist);
}

float sdfOutsideAlpha(float dist) {
	float aa = max(fwidth(dist), edgeSoftness);
	return smoothstep(-aa, aa, dist);
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

		int sideIdx = 0;
		float closestSide = edgeProximity(pixPos.x, fragBorderSize.x);
		float topSide = edgeProximity(pixPos.y, fragBorderSize.y);
		float rightSide = edgeProximity(dimensions.x - pixPos.x, fragBorderSize.z);
		float bottomSide = edgeProximity(dimensions.y - pixPos.y, fragBorderSize.w);
		if (topSide < closestSide) {
			sideIdx = 1;
			closestSide = topSide;
		}
		if (rightSide < closestSide) {
			sideIdx = 2;
			closestSide = rightSide;
		}
		if (bottomSide < closestSide) {
			sideIdx = 3;
		}
		vec4 borderColor = fragBorderColorsLTRB[sideIdx];

		float dist = roundedBoxSDF(centerPixPos, size, fragBorderRadius);
		float smoothedAlpha = sdfInsideAlpha(dist);

		vec2 innerSize = size - vec2(
			(fragBorderSize.x + fragBorderSize.z) * 0.5,
			(fragBorderSize.y + fragBorderSize.w) * 0.5
		);
		vec2 innerCenterPixPos = centerPixPos + vec2(
			(fragBorderSize.x - fragBorderSize.z) * 0.5,
			(fragBorderSize.y - fragBorderSize.w) * 0.5
		);
		vec4 innerBorderRadius = max(fragBorderRadius - vec4(
			max(fragBorderSize.x, fragBorderSize.w),
			max(fragBorderSize.z, fragBorderSize.w),
			max(fragBorderSize.z, fragBorderSize.y),
			max(fragBorderSize.x, fragBorderSize.y)
		), vec4(0.0));
		float innerDist = roundedBoxSDF(innerCenterPixPos, innerSize, innerBorderRadius);
		float smoothedBorderAlpha = smoothedAlpha * sdfOutsideAlpha(innerDist);

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
