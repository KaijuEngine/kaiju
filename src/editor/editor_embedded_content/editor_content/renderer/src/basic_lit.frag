#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;
layout(location = 2) in vec3 fragTangentViewPos;
layout(location = 3) in vec3 fragTangentFragPos;
layout(location = 4) in vec3 fragLightTPos[NR_LIGHTS];
layout(location = 8) in vec3 fragLightTDir[NR_LIGHTS];
layout(location = 12) in vec4 fragPosLightSpace[NR_LIGHTS];
layout(location = 16) in vec3 fragPos;
layout(location = 17) in vec3 fragNormal;
layout(location = 19) in flat int lightCount;
layout(location = 20) in flat int lightIndexes[NR_LIGHTS];

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

layout(binding = 1) uniform sampler2D textures;

const float ambientStrength = 0.5;
const float baseSoftness = 0.2;

void main() {
	vec3 V = normalize(fragTangentViewPos - fragTangentFragPos);
	vec4 albedo = texture(textures, fragTexCoords);
	vec3 N = normalize(fragNormal);
	for (int i = 0; i < lightCount; ++i) {
		LightInfo light = lightInfos[lightIndexes[i]];
		vec3 fltPos = fragLightTPos[i];
		vec3 fltDir = fragLightTDir[i];
		vec4 fplSpace = fragPosLightSpace[i];
		// Calculate per-light radiance
		vec3 L = normalize(fltPos - fragTangentFragPos);
		vec3 H = normalize(V + L);
		float attenuation = 1.0;
		if (light.type == 0) {
			attenuation = light.intensity;
		} else if (light.type == 1) {
			float d = length(fltPos - fragTangentFragPos);
			attenuation = light.intensity / (light.constant +
				light.linear * d + light.quadratic * (d * d));
		} else if (light.type == 2) {
			float d = length(fltPos - fragTangentFragPos);
			attenuation = light.intensity
				/ (light.constant + light.linear * d + light.quadratic * (d * d));
			// Spotlight (soft edges)
			vec3 fragDir = normalize(fragPos - light.position);
			float theta = dot(light.direction, fragDir);
			//float theta = dot(L, fltDir);
			float epsilon = (light.cutoff - light.outerCutoff);
			float intensity = clamp((theta - light.outerCutoff) / epsilon, 0.0, 1.0);
			attenuation *= intensity;
		}
	}
	// TODO:  Select the strongest light?
	const LightInfo sun = lightInfos[0];
	// Ambient
    vec3 ambient = ambientStrength * albedo.rgb * sun.diffuse;
    // Diffuse
    float diff = max(dot(N, -sun.direction), 0.0);
    vec3 diffuse = diff * sun.diffuse * albedo.rgb;
	// Combine
    vec3 result = ambient + diffuse;
	// Shadows
	float maxShadow = 0.0;
#ifdef STATIC_SHADOWS
	for (int i = 0; i < MAX_POINT_SHADOWS; i++) {
		PointShadow s = staticShadows[i];
		vec2 toShadowXZ = fragPos.xz - s.point.xy;
		float dist = length(toShadowXZ);
		float effectiveSoftness = s.strength * baseSoftness;
		float falloff = 1.0 - smoothstep(s.radius - effectiveSoftness, s.radius, dist);
		float shadow = falloff * s.strength;
		maxShadow = max(maxShadow, shadow);
	}
	result *= (1.0 - maxShadow);
#endif
#ifdef DYNAMIC_SHADOWS
	for (int i = 0; i < MAX_POINT_SHADOWS; i++) {
		PointShadow s = dynamicShadows[i];
		vec2 toShadowXZ = fragPos.xz - s.point.xy;
		float dist = length(toShadowXZ);
		float effectiveSoftness = s.strength * baseSoftness;
		float falloff = 1.0 - smoothstep(s.radius - effectiveSoftness, s.radius, dist);
		float shadow = falloff * s.strength;
		maxShadow = max(maxShadow, shadow);
	}
	result *= (1.0 - maxShadow);
#endif
	vec4 unWeightedColor = vec4(result, albedo.a);
#include "inc_fragment_oit_block.inl"
}
