#version 460
#define FRAGMENT_SHADER
#define HAS_GBUFFER

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_NORMAL 3
#define LAYOUT_FRAG_TEX_COORDS 4
#define LAYOUT_FRAG_FLAGS 11

#include "kaiju.glsl"
#include "pbr_lighting.glsl"

layout(location = 1) flat in vec4 fragDeepColor;
layout(location = 5) flat in ivec4 fragOceanLightIds;
layout(location = 6) flat in vec4 fragOceanBrushCenterRadius;
layout(location = 7) flat in vec4 fragOceanBrushParams;
layout(location = 8) flat in vec4 fragOceanBrushColor;
layout(location = 10) flat in vec4 fragOceanWaveParams;

const int OCEAN_WAVE_COUNT = 6;
const vec2 oceanWaveDirections[OCEAN_WAVE_COUNT] = vec2[](
	vec2(0.8192, 0.5735), vec2(-0.4229, 0.9062),
	vec2(0.1736, -0.9848), vec2(-0.9397, -0.3420),
	vec2(0.6428, -0.7660), vec2(-0.0872, 0.9962)
);
const float oceanWaveFrequencies[OCEAN_WAVE_COUNT] = float[](
	0.61, 0.89, 1.27, 1.83, 2.57, 3.41
);
const float oceanWaveSpeeds[OCEAN_WAVE_COUNT] = float[](
	0.43, -0.57, 0.71, -0.86, 1.03, -1.21
);
const float oceanWaveWeights[OCEAN_WAVE_COUNT] = float[](
	0.29, 0.23, 0.18, 0.13, 0.10, 0.07
);

vec3 applyOceanBrushOverlay(vec3 oceanColor) {
	if (fragOceanBrushCenterRadius.w <= 0.0) {
		return oceanColor;
	}
	float radius = max(fragOceanBrushCenterRadius.z, 0.001);
	float ringWidth = max(fragOceanBrushParams.x, 0.001);
	float dist = distance(fragPos.xz, fragOceanBrushCenterRadius.xy);
	float edgeFeather = max(fwidth(dist), 0.001);
	float ring = 1.0 - smoothstep(ringWidth, ringWidth + edgeFeather, abs(dist - radius));
	float fill = (1.0 - smoothstep(radius - ringWidth, radius, dist)) * 0.12;
	float alpha = clamp(max(ring * 0.88, fill) * fragOceanBrushColor.a, 0.0, 1.0);
	return mix(oceanColor, fragOceanBrushColor.rgb, alpha);
}

float oceanWaveHeight(vec2 position, float amplitude, float speed, float frequency) {
	float height = 0.0;
	for (int wave = 0; wave < OCEAN_WAVE_COUNT; ++wave) {
		float phase = dot(position, oceanWaveDirections[wave]) * frequency *
			oceanWaveFrequencies[wave] + time * speed * oceanWaveSpeeds[wave];
		height += sin(phase) * oceanWaveWeights[wave];
	}
	return amplitude * height;
}

vec3 oceanRippleNormal(vec3 geometricNormal) {
	float frequency = max(fragOceanWaveParams.z, 0.001);
	float speed = max(fragOceanWaveParams.y, 0.0);
	vec2 rippleSlope = vec2(0.0);
	const vec2 rippleDirections[5] = vec2[](
		vec2(0.7314, -0.6820), vec2(-0.9119, -0.4104),
		vec2(0.3420, 0.9397), vec2(-0.5878, 0.8090), vec2(0.9848, 0.1736)
	);
	const float rippleFrequencies[5] = float[](4.1, 5.9, 7.7, 10.3, 13.1);
	const float rippleSpeeds[5] = float[](1.37, -1.09, 1.73, -2.03, 2.29);
	const float rippleWeights[5] = float[](0.016, 0.012, 0.009, 0.006, 0.004);
	for (int ripple = 0; ripple < 5; ++ripple) {
		float phase = dot(fragPos.xz, rippleDirections[ripple]) * frequency *
			rippleFrequencies[ripple] + time * speed * rippleSpeeds[ripple];
		rippleSlope += cos(phase) * rippleDirections[ripple] * rippleWeights[ripple];
	}
	return normalize(geometricNormal + vec3(-rippleSlope.x, 0.0, -rippleSlope.y));
}

bool oceanLight(int lightIndex, vec3 normal, vec3 viewDirection, float roughness,
	out vec3 diffuse, out vec3 glint, out vec3 ambient) {
	if (lightIndex < 0 || lightIndex >= MAX_LIGHTS) {
		return false;
	}
	LightInfo light = lightInfos[lightIndex];
	vec3 lightDirection;
	float attenuation;
	if (light.type == 0) {
		lightDirection = pbrSafeNormalize(-light.direction, normal);
		attenuation = max(light.intensity, 0.0);
	} else if (light.type == 1 || light.type == 2) {
		vec3 toLight = light.position - fragPos;
		float distanceToLight = length(toLight);
		lightDirection = pbrSafeNormalize(toLight, normal);
		attenuation = pbrDistanceAttenuation(light, distanceToLight);
		if (light.type == 2) {
			vec3 lightToFragment = pbrSafeNormalize(fragPos - light.position, -lightDirection);
			float theta = dot(pbrSafeNormalize(light.direction, -lightDirection), lightToFragment);
			float epsilon = max(light.cutoff - light.outerCutoff, 0.0001);
			attenuation *= clamp((theta - light.outerCutoff) / epsilon, 0.0, 1.0);
		}
	} else {
		return false;
	}
	float nDotL = max(dot(normal, lightDirection), 0.0);
	vec3 radiance = max(light.diffuse, vec3(0.0)) * attenuation;
	// Water body color should dominate. Lighting only gives the surface a
	// gentle directional read instead of drawing continent-sized bright lobes.
	diffuse = radiance * (0.040 + nDotL * 0.045);
	vec3 halfway = pbrSafeNormalize(lightDirection + viewDirection, normal);
	float tightHighlight = pow(max(dot(normal, halfway), 0.0), mix(260.0, 90.0, roughness));
	float glintStrength = mix(0.055, 0.025, roughness);
	glint = min(max(light.specular, vec3(0.0)) * attenuation * tightHighlight * glintStrength,
		vec3(0.075));
	ambient = max(light.ambient, vec3(0.0));
	return true;
}

void main() {
	vec3 geometricNormal = pbrSafeNormalize(fragNormal, vec3(0.0, 1.0, 0.0));
	vec3 N = oceanRippleNormal(geometricNormal);
	vec3 V = pbrSafeNormalize(cameraPosition.xyz - fragPos, geometricNormal);
	float nDotV = max(dot(N, V), 0.0);
	float fresnel = 0.02 + 0.98 * pow(1.0 - nDotV, 5.0);

	vec3 shallow = pbrSrgbToLinear(max(fragColor.rgb, vec3(0.0)));
	vec3 deep = pbrSrgbToLinear(max(fragDeepColor.rgb, vec3(0.0)));
	float amplitude = max(fragOceanWaveParams.x, 0.001);
	float waveHeight = oceanWaveHeight(fragPos.xz, amplitude,
		max(fragOceanWaveParams.y, 0.0), max(fragOceanWaveParams.z, 0.001));
	float crest = smoothstep(-0.35, 0.65, waveHeight / amplitude);
	vec3 albedo = mix(deep, shallow, clamp(0.54 + crest * 0.10, 0.0, 1.0));
	float roughness = clamp(fragOceanWaveParams.w, 0.0, 1.0);

	processGBuffer(N);

	vec3 diffuseLighting = vec3(0.0);
	vec3 glintLighting = vec3(0.0);
	vec3 ambientLighting = vec3(0.08) * albedo;
	for (int i = 0; i < 4; ++i) {
		vec3 diffuse;
		vec3 glint;
		vec3 ambient;
		if (oceanLight(fragOceanLightIds[i], N, V, roughness, diffuse, glint, ambient)) {
			diffuseLighting += diffuse * albedo;
			glintLighting += glint;
			ambientLighting += ambient * albedo;
		}
	}

	vec3 skyReflection = pbrSrgbToLinear(mix(vec3(0.025, 0.13, 0.22),
		vec3(0.08, 0.26, 0.36), fresnel));
	vec3 color = pbrFinalColor(ambientLighting, diffuseLighting + glintLighting,
		albedo * 0.42 + skyReflection * (0.012 + fresnel * 0.035));
	color = applyOceanBrushOverlay(color);
	processFinalColor(vec4(color, 1.0));
}
