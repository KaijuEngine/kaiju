#version 460
#define FRAGMENT_SHADER

#define SAMPLER_COUNT   4 // color, normal, metallicRoughness, emissive
#define SHADOW_SAMPLERS

#define LAYOUT_FRAG_COLOR
#define LAYOUT_FRAG_FLAGS
#define LAYOUT_FRAG_POS
#define LAYOUT_FRAG_TEX_COORDS
#define LAYOUT_FRAG_NORMAL
#define LAYOUT_FRAG_METALLIC
#define LAYOUT_FRAG_ROUGHNESS
#define LAYOUT_FRAG_EMISSIVE

#define LAYOUT_ALL_LIGHT_REQUIREMENTS

#include "kaiju.glsl"

void main() {
	vec3 V = normalize(fragTangentViewPos - fragTangentFragPos);

	// Convert albedo from sRGB to linear space
	//vec3 albedo = pow(texture(textures[0], fragTexCoords).rgb, 2.2);
	vec3 albedo = texture(textures[0], fragTexCoords).rgb;
	albedo.x = pow(albedo.x, 2.2);
	albedo.y = pow(albedo.y, 2.2);
	albedo.z = pow(albedo.z, 2.2);
	vec3 N = texture(textures[1], fragTexCoords).rgb;
	N = normalize(N * 2.0 - 1.0);
	vec4 mrMap = texture(textures[2], fragTexCoords).rgba;
	//float mMetallic = max(mrMap.b, fragMetallic);
	//float mRoughness = max(mrMap.g, fragRoughness);
	float mMetallic = mrMap.b;
	float mRoughness = mrMap.g;
	vec4 emission = vec4(texture(textures[3], fragTexCoords).rgb * fragEmissive, 0.0);
	//float occlusion = max(mrMap.r, occlusion);
	float occlusion = 1.0;

    processGBuffer(N);

	vec3 F0 = vec3(0.04);
	F0 = mix(F0, albedo, mMetallic);

	// Reflectance equation
	vec3 Lo = vec3(0.0);
	for (int i = 0; i < fragLightCount; ++i) {
		int lightIdx = fragLightIndexes[i];
		LightInfo light = lightInfos[lightIdx];
		vec3 fltPos = fragLightTPos[i];
		vec3 fltDir = fragLightTDir[i];
		vec4 fplSpace = fragPosLightSpace[i];
		// Calculate per-light radiance
		vec3 L = normalize(fltPos - fragTangentFragPos);
		vec3 H = normalize(V + L);
		float attenuation = 1.0;
		float lightShadow = 0.0;
		if (light.type == 0) {
			attenuation = light.intensity;
			lightShadow = directShadowCalculation(fplSpace, N, fltDir, lightIdx);
		} else if (light.type == 1) {
			float d = length(fltPos - fragTangentFragPos);
			attenuation = light.intensity / (light.constant +
				light.linear * d + light.quadratic * (d * d));
			lightShadow = pointShadowCalculation(fragPos, light.position, light.farPlane, lightIdx, fragNormal);
		} else if (light.type == 2) {
			float d = length(fltPos - fragTangentFragPos);
			attenuation = light.intensity / (light.constant +
				light.linear * d + light.quadratic * (d * d));
			// Spotlight (soft edges)
			vec3 fragDir = normalize(fragPos - light.position);
			float theta = dot(light.direction, fragDir);
			//float theta = dot(L, fltDir);
			float epsilon = (light.cutoff - light.outerCutoff);
			float intensity = clamp((theta - light.outerCutoff) / epsilon, 0.0, 1.0);
			attenuation *= intensity;
			lightShadow = spotShadowCalculation(fplSpace, N, fltDir, light.nearPlane, light.farPlane, lightIdx);
		}
		vec3 radiance = light.diffuse * attenuation;
		// Cook-torrance brdf
		float NDF = distributionGGX(N, H, mRoughness);
		float G = geometrySmith(N, V, L, mRoughness);
		vec3 F = fresnelSchlick(max(dot(H, V), 0.0), F0);
		
		vec3 kS = F;
		vec3 kD = vec3(1.0) - kS;
		kD *= 1.0 - mMetallic;
		
		vec3 numerator = NDF * G * F;
		float denominator = 4.0 * max(dot(N, V), 0.0) * max(dot(N, L), 0.0);
		vec3 specular = numerator / max(denominator, 0.001);
			
		// Add to outgoing radiance Lo
		float visibility = 1.0 - lightShadow;
		float NdotL = max(dot(N, L), 0.0);
		Lo += (kD * albedo / PI + specular) * radiance * NdotL * visibility;
	}
	vec3 ambient = vec3(0.03) * albedo * occlusion;
	vec3 color = ambient + Lo;
	color = color / (color + vec3(1.0));
	color = pow(color, vec3(1.0/2.2));
	outColor = (vec4(color, 1.0) * fragColor) + emission;
}
