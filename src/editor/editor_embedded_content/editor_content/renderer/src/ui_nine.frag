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

float sideBorderAlpha(float edgeDistance, float borderWidth) {
	if (borderWidth <= 0.0) {
		return 0.0;
	}
	return 1.0 - smoothstep(borderWidth - 0.5, borderWidth + 0.5, edgeDistance);
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

		float leftBorder = sideBorderAlpha(pixPos.x, fragBorderSize.x);
		float topBorder = sideBorderAlpha(pixPos.y, fragBorderSize.y);
		float rightBorder = sideBorderAlpha(dimensions.x - pixPos.x, fragBorderSize.z);
		float bottomBorder = sideBorderAlpha(dimensions.y - pixPos.y, fragBorderSize.w);
		float smoothedBorderAlpha = max(max(leftBorder, topBorder), max(rightBorder, bottomBorder));

		int sideIdx = 0;
		float sideAlpha = leftBorder;
		if (topBorder > sideAlpha) {
			sideIdx = 1;
			sideAlpha = topBorder;
		}
		if (rightBorder > sideAlpha) {
			sideIdx = 2;
			sideAlpha = rightBorder;
		}
		if (bottomBorder > sideAlpha) {
			sideIdx = 3;
		}
		vec4 borderColor = fragBorderColorsLTRB[sideIdx];

		// Border radius
		float dist = roundedBoxSDF(centerPixPos, size, fragBorderRadius);
		float smoothedAlpha = 1.0-smoothstep(0.0, edgeSoftness, dist);

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
