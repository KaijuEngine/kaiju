#version 450
//#version 300 es
//precision mediump float;

#ifdef VULKAN
	layout(location = 0) in vec4 fragColor;
	layout(location = 1) in vec4 fragBGColor;
	layout(location = 2) in vec4 fragSize2D;
	layout(location = 3) in vec4 fragBorderRadius;
	layout(location = 4) in vec4 fragBorderSize;
	layout(location = 5) in mat4 fragBorderColor;
	layout(location = 9) in vec2 fragTexCoord;
	layout(location = 10) in vec2 fragBorderLen;

	layout(binding = 1) uniform sampler2D texSampler;
#else
	in vec4 fragColor;
	in vec4 fragBGColor;
	in vec4 fragSize2D;
	in vec4 fragScissor;
	in vec4 fragBorderRadius;
	in vec4 fragBorderSize;
	in mat4 fragBorderColor;
	in vec2 fragTexCoord;
	in vec2 fragBorderLen;

	uniform sampler2D texSampler;
#endif

layout(location = 0) out vec4 outColor;
layout(location = 1) out float reveal;

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
    Radius.x  = (CenterPosition.y>0.0)?Radius.x  : Radius.y;
    vec2 q = abs(CenterPosition)-Size+Radius.x;
    return min(max(q.x,q.y),0.0) + length(max(q,0.0)) - Radius.x;
}

void main(void) {
#ifndef VULKAN
	if (gl_FragCoord.x < fragScissor.x || gl_FragCoord.x > fragScissor.z ||
		gl_FragCoord.y < fragScissor.y || gl_FragCoord.y > fragScissor.w)
	{
		discard;
	}
#endif
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
		pixPos.y += fragBorderSize.y/2.0-fragBorderSize.w/2.0;
		size.x -= (fragBorderSize.x+fragBorderSize.z)/2.0;
		size.y -= (fragBorderSize.y+fragBorderSize.w)/2.0;
		float borderDistance = roundedBoxSDF(pixPos, size, fragBorderRadius);
		float smoothedBorderAlpha = 1.0-smoothstep(0.0, edgeSoftness, borderDistance);
		smoothedBorderAlpha = 1.0 - smoothedBorderAlpha;
		// Border color
		vec4 bc = fragBorderColor[0].xyzw;
		unWeightedColor = mix(unWeightedColor, bc, smoothedBorderAlpha);
		unWeightedColor.a = smoothedAlpha * fragColor.a;
	}
#ifdef OIT
	float distWeight = clamp(0.03 / (1e-5 + pow(gl_FragCoord.z / 200.0, 4.0)), 1e-2, 3e3);
	float alphaWeight = min(1.0, max(max(unWeightedColor.r, unWeightedColor.g),
	max(unWeightedColor.b, unWeightedColor.a)) * 40.0 + 0.01);
	alphaWeight *= alphaWeight;
	float weight = alphaWeight * distWeight;
	outColor = vec4(unWeightedColor.rgb * unWeightedColor.a, unWeightedColor.a) * weight;
	reveal = unWeightedColor.a;
#else
	if (unWeightedColor.a < (1.0 - 0.0001))
		discard;
	outColor = unWeightedColor;
#endif
}
