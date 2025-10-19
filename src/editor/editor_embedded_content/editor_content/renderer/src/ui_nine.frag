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
#ifdef OIT
layout(location = 1) out float reveal;
#endif

float processAxis(float coord, float border, float ratio) {
	float len = border * ratio;
	float lScale = 1.0 - len * 2.0;
	float bScale = 1.0 - (border * 2.0);
	float res = (coord - len) * (bScale / lScale) + border;
	if (coord < len)
		res = coord / ratio;
	else if (coord > 1.0 - len)
		res = 1.0 - ((1.0 - coord) / ratio);
	return res;
}

//https://www.shadertoy.com/view/tltXDl
float roundedBoxSDF(vec2 CenterPosition, vec2 Size, vec4 Radius) {
	Radius.xy = (CenterPosition.x>0.0)?Radius.xw : Radius.yz;
    Radius.x  = (CenterPosition.y<0.0)?Radius.x  : Radius.y;  // <-- Fix: Change >0.0 to <0.0 to account for y-flip
    vec2 q = abs(CenterPosition)-Size+Radius.x;
    return min(max(q.x,q.y),0.0) + length(max(q,0.0)) - Radius.x;
}

void main(void) {
	vec2 newUV = vec2(
		processAxis(fragTexCoord.x, fragBorderLen.x / fragSize2D.z, fragSize2D.z / fragSize2D.x),
		processAxis(fragTexCoord.y, fragBorderLen.y / fragSize2D.w, fragSize2D.w / fragSize2D.y)
	);
	vec4 unWeightedColor = texture(texSampler, newUV) * fragColor;
	// Border
	{
		vec2 dimensions = fragSize2D.xy;
		vec2 size = dimensions/2.0;
		vec2 pixPos = size-(fragTexCoord.xy * dimensions);
		float edgeSoftness = 2.0;
		// Border radius
		float dist = roundedBoxSDF(pixPos, size, fragBorderRadius);
		float smoothedAlpha = 1.0-smoothstep(0.0, edgeSoftness, dist);
		// Border size
		pixPos.x += fragBorderSize.x/2.0-fragBorderSize.z/2.0;
		pixPos.y += fragBorderSize.w/2.0-fragBorderSize.y/2.0;  // <-- Fix: Swap to bottom/2 - top/2 for correct y-shift direction
		size.x -= (fragBorderSize.x+fragBorderSize.z)/2.0;
		size.y -= (fragBorderSize.y+fragBorderSize.w)/2.0;
		float borderDistance = roundedBoxSDF(pixPos, size, fragBorderRadius);
		float innerMask = 1.0-smoothstep(0.0, edgeSoftness, borderDistance);  // <-- Rename for clarity; this is 1 inside inner
		float smoothedBorderAlpha = 1.0 - innerMask;  // <-- Fix: Invert to get 1 in border region (outside inner), 0 inside
		// Border color
		vec4 bc = fragBorderColor[0].xyzw;
		unWeightedColor = mix(unWeightedColor, bc, smoothedBorderAlpha);
		unWeightedColor.a = smoothedAlpha * unWeightedColor.a;
	}
#include "inc_fragment_oit_block.inl"
}
