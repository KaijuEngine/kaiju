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

const float MIN_ROUGHNESS = 0.045;
const float DEFAULT_AMBIENT_STRENGTH = 0.03;

vec3 safeNormalize(vec3 v, vec3 fallback) {
	float len2 = dot(v, v);
	if (len2 <= 0.00000001) {
		return fallback;
	}
	return v * inversesqrt(len2);
}

vec3 srgbToLinear(vec3 color) {
	return pow(max(color, vec3(0.0)), vec3(2.2));
}

vec3 linearToSrgb(vec3 color) {
	return pow(max(color, vec3(0.0)), vec3(1.0 / 2.2));
}

vec3 acesTonemap(vec3 color) {
	const float a = 2.51;
	const float b = 0.03;
	const float c = 2.43;
	const float d = 0.59;
	const float e = 0.14;
	return clamp((color * (a * color + b)) / (color * (c * color + d) + e), 0.0, 1.0);
}

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

float distanceAttenuation(LightInfo light, float dist) {
	float denom = light.constant + light.linear * dist + light.quadratic * dist * dist;
	return max(light.intensity, 0.0) / max(denom, 0.0001);
}

float lightVisibility(int lightType, int lightIdx, vec3 n, vec3 l, vec4 lightSpace, LightInfo light) {
	#ifdef SHADOW_SAMPLERS
		if (lightType == 0) {
			return 1.0 - directShadowCalculation(n, l, lightIdx, light.farPlane);
		}
		if (lightType == 1) {
			return 1.0 - pointShadowCalculation(fragPos, light.position, light.farPlane, lightIdx, n);
		}
		if (lightType == 2) {
			return 1.0 - spotShadowCalculation(lightSpace, n, l, light.nearPlane, light.farPlane, lightIdx);
		}
	#endif
	return 1.0;
}

void main() {
	vec4 baseSample = texture(textures[0], fragTexCoords);
	vec3 albedo = srgbToLinear(baseSample.rgb) * max(fragColor.rgb, vec3(0.0));
	float alpha = baseSample.a * fragColor.a;

	vec4 mrSample = texture(textures[2], fragTexCoords);
	float metallic = clamp(mrSample.b * max(fragMetallic, 0.0), 0.0, 1.0);
	float roughness = clamp(mrSample.g * max(fragRoughness, MIN_ROUGHNESS), MIN_ROUGHNESS, 1.0);
	float occlusion = clamp(mrSample.r, 0.0, 1.0);
	vec3 emission = srgbToLinear(texture(textures[3], fragTexCoords).rgb) * max(fragEmissive, 0.0);

	vec3 geometricNormal = safeNormalize(fragNormal, vec3(0.0, 1.0, 0.0));
	vec3 N = pbrNormal(geometricNormal);
	vec3 V = safeNormalize(cameraPosition.xyz - fragPos, geometricNormal);
	float NdotV = max(dot(N, V), 0.0);

	processGBuffer(N);

	vec3 F0 = mix(vec3(0.04), albedo, metallic);
	vec3 Lo = vec3(0.0);
	vec3 ambient = vec3(DEFAULT_AMBIENT_STRENGTH) * albedo * occlusion;

	for (int i = 0; i < fragLightCount; ++i) {
		int lightIdx = fragLightIndexes[i];
		if (lightIdx < 0 || lightIdx >= MAX_LIGHTS) {
			continue;
		}
		LightInfo light = lightInfos[lightIdx];
		vec3 L = vec3(0.0);
		float attenuation = 0.0;
		if (light.type == 0) {
			L = safeNormalize(-light.direction, geometricNormal);
			attenuation = max(light.intensity, 0.0);
		} else if (light.type == 1) {
			vec3 toLight = light.position - fragPos;
			float dist = length(toLight);
			L = safeNormalize(toLight, geometricNormal);
			attenuation = distanceAttenuation(light, dist);
		} else if (light.type == 2) {
			vec3 toLight = light.position - fragPos;
			float dist = length(toLight);
			L = safeNormalize(toLight, geometricNormal);
			attenuation = distanceAttenuation(light, dist);
			vec3 lightToFrag = safeNormalize(fragPos - light.position, -L);
			float theta = dot(safeNormalize(light.direction, -L), lightToFrag);
			float epsilon = max(light.cutoff - light.outerCutoff, 0.0001);
			attenuation *= clamp((theta - light.outerCutoff) / epsilon, 0.0, 1.0);
		} else {
			continue;
		}

		float NdotL = max(dot(N, L), 0.0);
		if (attenuation <= 0.0 || NdotL <= 0.0) {
			continue;
		}

		vec3 H = safeNormalize(V + L, N);
		float NDF = distributionGGX(N, H, roughness);
		float G = geometrySmith(N, V, L, roughness);
		vec3 F = fresnelSchlick(max(dot(H, V), 0.0), F0);
		vec3 kD = (vec3(1.0) - F) * (1.0 - metallic);
		vec3 specular = (NDF * G * F) / max(4.0 * NdotV * NdotL, 0.001);
		vec3 radiance = max(light.diffuse, vec3(0.0)) * attenuation;
		float visibility = lightVisibility(light.type, lightIdx, N, L, fragPosLightSpace[i], light);
		Lo += (kD * albedo / PI + specular) * radiance * NdotL * visibility;
		ambient += max(light.ambient, vec3(0.0)) * albedo * occlusion;
	}

	vec3 color = ambient + Lo + emission;
	color = linearToSrgb(acesTonemap(color));
	processFinalColor(vec4(color, alpha));
}
