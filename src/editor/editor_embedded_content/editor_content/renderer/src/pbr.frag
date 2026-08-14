#version 460
#define FRAGMENT_SHADER
#define HAS_GBUFFER

#define SAMPLER_COUNT   4 // color, normal, metallicRoughness, emissive
#define SHADOW_SAMPLERS

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_METALLIC 5
#define LAYOUT_FRAG_ROUGHNESS 6
#define LAYOUT_FRAG_EMISSIVE 7

#define LAYOUT_ALL_LIGHT_REQUIREMENTS 8

#include "kaiju.glsl"
#include "pbr_lighting.glsl"

mat3 fallbackTBN(vec3 n) {
	vec3 up = abs(n.z) < 0.999 ? vec3(0.0, 0.0, 1.0) : vec3(0.0, 1.0, 0.0);
	vec3 t = normalize(cross(up, n));
	vec3 b = cross(n, t);
	return mat3(t, b, n);
}

mat3 cotangentFrame(vec3 n, vec3 pos, vec2 uv) {
	vec3 dp1 = dFdx(pos);
	vec3 dp2 = dFdy(pos);
	vec2 duv1 = dFdx(uv);
	vec2 duv2 = dFdy(uv);
	vec3 dp2Perp = cross(dp2, n);
	vec3 dp1Perp = cross(n, dp1);
	vec3 t = dp2Perp * duv1.x + dp1Perp * duv2.x;
	vec3 b = dp2Perp * duv1.y + dp1Perp * duv2.y;
	float maxLen = max(dot(t, t), dot(b, b));
	if (maxLen <= 0.00000001) {
		return fallbackTBN(n);
	}
	float invMax = inversesqrt(maxLen);
	return mat3(t * invMax, b * invMax, n);
}

vec3 pbrNormal(vec3 geometricNormal) {
	vec3 normalSample = texture(textures[1], fragTexCoords).rgb;
	vec3 tangentNormal = normalSample * 2.0 - 1.0;
	bool whiteFallback = all(greaterThanEqual(normalSample, vec3(0.999)));
	if (whiteFallback || dot(tangentNormal, tangentNormal) <= 0.0001) {
		tangentNormal = vec3(0.0, 0.0, 1.0);
	}
	mat3 tbn = cotangentFrame(geometricNormal, fragPos, fragTexCoords);
	return normalize(tbn * normalize(tangentNormal));
}

void main() {
	vec4 baseSample = texture(textures[0], fragTexCoords);
	vec3 albedo = pbrSrgbToLinear(baseSample.rgb) * max(fragColor.rgb, vec3(0.0));
	float alpha = baseSample.a * fragColor.a;

	vec4 mrSample = texture(textures[2], fragTexCoords);
	float metallic = clamp(mrSample.b * max(fragMetallic, 0.0), 0.0, 1.0);
	float roughness = clamp(mrSample.g * max(fragRoughness, PBR_MIN_ROUGHNESS), PBR_MIN_ROUGHNESS, 1.0);
	float occlusion = clamp(mrSample.r, 0.0, 1.0);
	vec3 emission = pbrSrgbToLinear(texture(textures[3], fragTexCoords).rgb) * max(fragEmissive, 0.0);

	vec3 geometricNormal = pbrSafeNormalize(fragNormal, vec3(0.0, 1.0, 0.0));
	vec3 N = pbrNormal(geometricNormal);
	vec3 V = pbrSafeNormalize(cameraPosition.xyz - fragPos, geometricNormal);

	processGBuffer(N);

	vec3 Lo = vec3(0.0);
	vec3 ambient = vec3(PBR_DEFAULT_AMBIENT_STRENGTH) * albedo * occlusion;

	for (int i = 0; i < fragLightCount; ++i) {
		int lightIdx = fragLightIndexes[i];
		pbrAccumulateLight(lightIdx, fragPosLightSpace[i], albedo, N, V,
			metallic, roughness, occlusion, Lo, ambient);
	}

	vec3 color = pbrFinalColor(ambient, Lo, emission);
	processFinalColor(vec4(color, alpha));
}
