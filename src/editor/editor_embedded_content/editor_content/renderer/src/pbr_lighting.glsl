const float PBR_MIN_ROUGHNESS = 0.045;
const float PBR_DEFAULT_AMBIENT_STRENGTH = 0.03;

vec3 pbrSafeNormalize(vec3 value, vec3 fallback) {
	float len2 = dot(value, value);
	if (len2 <= 0.00000001) {
		return fallback;
	}
	return value * inversesqrt(len2);
}

vec3 pbrSrgbToLinear(vec3 color) {
	return pow(max(color, vec3(0.0)), vec3(2.2));
}

vec3 pbrLinearToSrgb(vec3 color) {
	return pow(max(color, vec3(0.0)), vec3(1.0 / 2.2));
}

vec3 pbrAcesTonemap(vec3 color) {
	const float a = 2.51;
	const float b = 0.03;
	const float c = 2.43;
	const float d = 0.59;
	const float e = 0.14;
	return clamp((color * (a * color + b)) / (color * (c * color + d) + e), 0.0, 1.0);
}

float pbrDistanceAttenuation(LightInfo light, float distanceToLight) {
	float denominator = light.constant + light.linear * distanceToLight +
		light.quadratic * distanceToLight * distanceToLight;
	return max(light.intensity, 0.0) / max(denominator, 0.0001);
}

float pbrLightVisibility(int lightType, int lightIndex, vec3 normal, vec3 lightDirection,
	vec4 lightSpace, LightInfo light) {
#ifdef SHADOW_SAMPLERS
	if (light.shadowIndex < 0) {
		return 1.0;
	}
	if (lightType == 0) {
		return 1.0 - directShadowCalculation(normal, lightDirection, lightIndex,
			light.shadowIndex, light.farPlane);
	}
	if (lightType == 1) {
		return 1.0 - pointShadowCalculation(fragPos, light.position, light.farPlane,
			light.shadowIndex, normal);
	}
	if (lightType == 2) {
		return 1.0 - spotShadowCalculation(lightSpace, normal, lightDirection,
			light.nearPlane, light.farPlane, light.shadowIndex);
	}
#endif
	return 1.0;
}

void pbrAccumulateLight(int lightIndex, vec4 lightSpace, vec3 albedo, vec3 normal,
	vec3 viewDirection, float metallic, float roughness, float occlusion,
	inout vec3 directLighting, inout vec3 ambientLighting) {
	if (lightIndex < 0 || lightIndex >= MAX_LIGHTS) {
		return;
	}
	LightInfo light = lightInfos[lightIndex];
	// Ambient is a fill term and remains visible on surfaces facing away from
	// the light's direct contribution.
	ambientLighting += max(light.ambient, vec3(0.0)) * albedo * occlusion;
	vec3 lightDirection = vec3(0.0);
	float attenuation = 0.0;
	if (light.type == 0) {
		lightDirection = pbrSafeNormalize(-light.direction, normal);
		attenuation = max(light.intensity, 0.0);
	} else if (light.type == 1) {
		vec3 toLight = light.position - fragPos;
		float distanceToLight = length(toLight);
		lightDirection = pbrSafeNormalize(toLight, normal);
		attenuation = pbrDistanceAttenuation(light, distanceToLight);
	} else if (light.type == 2) {
		vec3 toLight = light.position - fragPos;
		float distanceToLight = length(toLight);
		lightDirection = pbrSafeNormalize(toLight, normal);
		attenuation = pbrDistanceAttenuation(light, distanceToLight);
		vec3 lightToFragment = pbrSafeNormalize(fragPos - light.position, -lightDirection);
		float theta = dot(pbrSafeNormalize(light.direction, -lightDirection), lightToFragment);
		float epsilon = max(light.cutoff - light.outerCutoff, 0.0001);
		attenuation *= clamp((theta - light.outerCutoff) / epsilon, 0.0, 1.0);
	} else {
		return;
	}

	float nDotL = max(dot(normal, lightDirection), 0.0);
	if (attenuation <= 0.0 || nDotL <= 0.0) {
		return;
	}
	float nDotV = max(dot(normal, viewDirection), 0.0);
	vec3 halfway = pbrSafeNormalize(viewDirection + lightDirection, normal);
	vec3 f0 = mix(vec3(0.04), albedo, metallic);
	float distribution = distributionGGX(normal, halfway, roughness);
	float geometry = geometrySmith(normal, viewDirection, lightDirection, roughness);
	vec3 fresnel = fresnelSchlick(max(dot(halfway, viewDirection), 0.0), f0);
	vec3 diffuseFactor = (vec3(1.0) - fresnel) * (1.0 - metallic);
	vec3 specular = (distribution * geometry * fresnel) /
		max(4.0 * nDotV * nDotL, 0.001);
	vec3 radiance = max(light.diffuse, vec3(0.0)) * attenuation;
	float visibility = pbrLightVisibility(light.type, lightIndex, normal,
		lightDirection, lightSpace, light);
	directLighting += (diffuseFactor * albedo / PI + specular) * radiance * nDotL * visibility;
}

vec3 pbrFinalColor(vec3 ambientLighting, vec3 directLighting, vec3 emission) {
	return pbrLinearToSrgb(pbrAcesTonemap(ambientLighting + directLighting + emission));
}
