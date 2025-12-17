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
layout(location = 18) in float fragMetallic;
layout(location = 19) in float fragRoughness;
layout(location = 20) in float fragEmissive;
layout(location = 21) in flat int lightCount;
layout(location = 22) in flat int lightIndexes[NR_LIGHTS];

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

layout(binding = 1) uniform sampler2D textures[4];
//colorMap;
//normalMap;
//metallicRoughnessMap;
//emissiveMap;
layout(binding = 2) uniform sampler2D shadowMap[MAX_LIGHTS];
layout(binding = 3) uniform samplerCube shadowCubeMap[MAX_LIGHTS];

vec3 fresnelSchlick(float cosTheta, vec3 F0);
float DistributionGGX(vec3 N, vec3 H, float fragRoughness);
float GeometrySchlickGGX(float NdotV, float fragRoughness);
float GeometrySmith(vec3 N, vec3 V, vec3 L, float fragRoughness);
float DirectShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir, int lightIdx);
float SpotShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir, float near, float far, int lightIdx);
float PointShadowCalculation(vec3 fragPos, vec3 lightPos, vec3 viewPos, float far, int lightIdx);

float LinearizeDepth(float depth, float near, float far) {
	float z = depth * 2.0 - 1.0; // Back to NDC 
	return (2.0 * near * far) / (far + near - z * (far - near));
}

void main() {
	//colorMap = textures[0];
	//normalMap = textures[1];
	//metallicRoughnessMap = textures[2];
	//emissiveMap = textures[3];

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

	vec3 F0 = vec3(0.04);
	F0 = mix(F0, albedo, mMetallic);

	// Reflectance equation
	vec3 Lo = vec3(0.0);
	float shadow = 0.0;
	float iShadow = 0.0;
	for (int i = 0; i < lightCount; ++i) {
		int lightIdx = lightIndexes[i];
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
			lightShadow = DirectShadowCalculation(fplSpace, N, fltDir, lightIdx);
		} else if (light.type == 1) {
			float d = length(fltPos - fragTangentFragPos);
			attenuation = light.intensity / (light.constant +
				light.linear * d + light.quadratic * (d * d));
			lightShadow = PointShadowCalculation(fragPos, light.position, V, light.farPlane, lightIdx);
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
			lightShadow = SpotShadowCalculation(fplSpace, N, fltDir, light.nearPlane, light.farPlane, lightIdx);
		}
		shadow += lightShadow;
		iShadow += 1.0 - lightShadow;
		vec3 radiance = light.diffuse * attenuation;
		// Cook-torrance brdf
		float NDF = DistributionGGX(N, H, mRoughness);
		float G = GeometrySmith(N, V, L, mRoughness);
		vec3 F = fresnelSchlick(max(dot(H, V), 0.0), F0);
		
		vec3 kS = F;
		vec3 kD = vec3(1.0) - kS;
		kD *= 1.0 - mMetallic;
		
		vec3 numerator = NDF * G * F;
		float denominator = 4.0 * max(dot(N, V), 0.0) * max(dot(N, L), 0.0);
		vec3 specular = numerator / max(denominator, 0.001);
			
		// Add to outgoing radiance Lo
		float NdotL = max(dot(N, L), 0.0);
		Lo += (kD * albedo / PI + specular) * radiance * NdotL;
	}

	vec3 ambient = vec3(0.03) * albedo * occlusion;
	vec3 color = ambient + Lo * (1.0 - shadow);
	//vec3 color = ambient + Lo;
	//shadow = (1.2 - clamp(shadow - iShadow * 0.65, 0.0, 1.0));

	color = color / (color + vec3(1.0));
	color = pow(color, vec3(1.0/2.2));

	outColor = (vec4(color, 1.0) * fragColor) + emission;
	//vec4 unWeightedColor = (vec4(color, 1.0) * fragColor) + emission;
	//float weight = clamp(pow(min(1.0, unWeightedColor.a * 10.0) + 0.01, 3.0) * 1e8 *
	//pow(1.0 - gl_FragCoord.z * 0.9, 3.0), 1e-2, 3e3);
	//outColor = vec4(unWeightedColor.rgb * unWeightedColor.a, unWeightedColor.a) * weight;
	//reveal = unWeightedColor.a;

	//outColor = (vec4(color * shadow, 1.0) * fragColor) + emission;
	//outColor = (vec4(texture(textures[0], fragTexCoords).rgb, 1.0) * fragColor);
	//outColor = vec4(1.0, 1.0, 1.0, 1.0);
	//outColor = texture(shadowMap, fragTexCoords);
	//outColor = vec4(vec3(LinearizeDepth(texture(shadowMap, fragTexCoords).r, shadowNear, shadowFar) / 20.0), 1.0);

	//vec3 delta = fragPos - lightInfos[0].position;
	//float closestDepth = texture(textures[5], delta).r;
	//closestDepth *= shadowFar;
	//outColor = vec4(vec3(closestDepth / shadowFar), 1.0);
}

vec3 fresnelSchlick(float cosTheta, vec3 F0) {
	return F0 + (1.0 - F0) * pow(max(1.0 - cosTheta, 0.0), 5.0);
}

float DistributionGGX(vec3 N, vec3 H, float fragRoughness) {
	float a      = fragRoughness*fragRoughness;
	float a2     = a*a;
	float NdotH  = max(dot(N, H), 0.0);
	float NdotH2 = NdotH*NdotH;
	float num   = a2;
	float denom = (NdotH2 * (a2 - 1.0) + 1.0);
	denom = PI * denom * denom;
	return num / denom;
}

float GeometrySchlickGGX(float NdotV, float fragRoughness) {
	float r = (fragRoughness + 1.0);
	float k = (r*r) / 8.0;
	float num   = NdotV;
	float denom = NdotV * (1.0 - k) + k;
	return num / denom;
}

float GeometrySmith(vec3 N, vec3 V, vec3 L, float fragRoughness) {
	float NdotV = max(dot(N, V), 0.0);
	float NdotL = max(dot(N, L), 0.0);
	float ggx2  = GeometrySchlickGGX(NdotV, fragRoughness);
	float ggx1  = GeometrySchlickGGX(NdotL, fragRoughness);
	return ggx1 * ggx2;
}

float DirectShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir, int lightIdx) {
	// Perform perspective divide
	vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
	// Transform to [0,1] range
	projCoords.xy = projCoords.xy * 0.5 + 0.5;

	// Get closest depth value from light's perspective
	// (using [0,1] range fragPosLight as coords)
	float closestDepth = texture(shadowMap[lightIdx], projCoords.xy).r;

	// Get depth of current fragment from light's perspective
	float currentDepth = projCoords.z;

	float bias = 0.005;//max(0.05 * (1.0 - dot(normal, lightDir)), 0.005);
	float shadow = 0.0;

	// Make the shadow smoother
	vec2 texelSize = 1.0 / vec2(textureSize(shadowMap[lightIdx], 0));
	for (int x = -1; x <= 1; ++x) {
		for (int y = -1; y <= 1; ++y) {
			float pcfDepth = texture(shadowMap[lightIdx], projCoords.xy + vec2(x, y) * texelSize).r;
			shadow += (currentDepth - bias) > pcfDepth ? 1.0 : 0.0;
		}
	}
	shadow /= 9.0;
	
	if (projCoords.z > 1.0)
		shadow = 0.0;
	return shadow;
}

float SpotShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir, float near, float far, int lightIdx)
{
	// Perform perspective divide
	vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
	// Transform to [0,1] range
	projCoords.xy = projCoords.xy * 0.5 + 0.5;

	// Get closest depth value from light's perspective
	// (using [0,1] range fragPosLight as coords)
	float closestDepth = texture(shadowMap[lightIdx], projCoords.xy).r;

	// Get depth of current fragment from light's perspective
	float currentDepth = projCoords.z;

	closestDepth = LinearizeDepth(closestDepth, near, far) / far;
	currentDepth = LinearizeDepth(currentDepth, near, far) / far;

	float bias = 0.005;//max(0.05 * (1.0 - dot(normal, lightDir)), 0.005);
	float shadow = 0.0;

	// Make the shadow smoother
	vec2 texelSize = 1.0 / vec2(textureSize(shadowMap[lightIdx], 0));
	for (int x = -1; x <= 1; ++x) {
		for (int y = -1; y <= 1; ++y) {
			float pcfDepth = texture(shadowMap[lightIdx], projCoords.xy + vec2(x, y) * texelSize).r;
			pcfDepth = LinearizeDepth(pcfDepth, near, far) / far;
			shadow += (currentDepth - bias) > pcfDepth ? 1.0 : 0.0;
		}
	}
	shadow /= 9.0;
	
	if (projCoords.z > 1.0)
		shadow = 0.0;

	return shadow;
}

// array of offset direction for sampling
const vec3 pointSamplingDiskGrid[20] = vec3[]
(
	vec3(1, 1,  1), vec3( 1, -1,  1), vec3(-1, -1,  1), vec3(-1, 1,  1),
	vec3(1, 1, -1), vec3( 1, -1, -1), vec3(-1, -1, -1), vec3(-1, 1, -1),
	vec3(1, 1,  0), vec3( 1, -1,  0), vec3(-1, -1,  0), vec3(-1, 1,  0),
	vec3(1, 0,  1), vec3(-1,  0,  1), vec3( 1,  0, -1), vec3(-1, 0, -1),
	vec3(0, 1,  1), vec3( 0, -1,  1), vec3( 0, -1, -1), vec3( 0, 1, -1)
);
float PointShadowCalculation(vec3 fragPos, vec3 lightPos, vec3 viewPos, float far, int lightIdx) {
	vec3 delta = fragPos - lightPos;
	float currentDepth = length(delta);
	float shadow = 0.0;
	float bias = 0.15;
	int samples = 20;
	float viewDistance = length(viewPos - fragPos);
	float diskRadius = (1.0 + (viewDistance / far)) / 25.0;
	for (int i = 0; i < samples; ++i) {
		float closestDepth = texture(shadowCubeMap[lightIdx], delta + pointSamplingDiskGrid[i] * diskRadius).r;
		closestDepth *= far;   // undo mapping [0;1]
		if ((currentDepth - bias) > closestDepth)
			shadow += 1.0;
	}
	shadow /= float(samples);
	return shadow;
}
