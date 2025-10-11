#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoord;
layout(location = 2) in vec3 fragTangentViewPos;
layout(location = 3) in vec3 fragTangentFragPos;
layout(location = 4) in vec3 fragLightTPos;
layout(location = 5) in vec4 fragPosLightSpace;
layout(location = 6) in vec3 fragPos;
layout(location = 7) in vec3 fragNormal;

layout(location = 0) out vec4 outColor;

// Height/Cavity/Roughness map
// Terrain normal map
// Rock color, normal
// Ground color, normal
layout(binding = 1) uniform sampler2D textures[6];

#define TEX_HEIGHT		0
#define TEX_HEIGHT_NML	1
#define TEX_ROCK		2
#define TEX_ROCK_NML	3
#define TEX_GROUND		4
#define TEX_GROUND_NML	5

vec3 specularity = vec3(0.05);
float tileScale = 200.0;

const float baseSoftness = 0.2;

void main() {
	vec3 hsr = texture(textures[TEX_HEIGHT], fragTexCoord).rgb;
	vec3 baseNormal = texture(textures[TEX_HEIGHT_NML], fragTexCoord).rgb * 2.0 - 1.0;
	vec2 tiledCoord = fragTexCoord * tileScale;
	vec3 groundColor = texture(textures[TEX_GROUND], tiledCoord).rgb;
	vec3 rockColor = texture(textures[TEX_ROCK], tiledCoord).rgb;
	vec3 groundNormal = texture(textures[TEX_GROUND_NML], tiledCoord).rgb * 2.0 - 1.0;
	vec3 rockNormal = texture(textures[TEX_ROCK_NML], tiledCoord).rgb * 2.0 - 1.0;

	float height = hsr.r;
	//float slope = hsr.g;
	float roughness = hsr.b;

	float slope = 1.0 - baseNormal.z;
	float rockFactor = smoothstep(0.2, 0.4, slope);

	// Debug: Uncomment to visualize slope (black = steep, white = flat)
	//outColor = vec4(vec3(slope), 1.0); return;

	vec3 albedo = mix(groundColor, rockColor, rockFactor);
	vec3 detailNormal = mix(groundNormal, rockNormal, rockFactor);

	// Blend normals: Add XY, preserve Z dominance from baseNormal
	vec3 tangentNormal = vec3(baseNormal.xy + detailNormal.xy, baseNormal.z * detailNormal.z);
    tangentNormal = normalize(tangentNormal);

    float shininess = 64.0 * (1.0 - roughness);
    shininess = max(shininess, 0.1);

    vec3 viewDir = normalize(fragTangentViewPos - fragTangentFragPos);
    vec3 lightDir = normalize(fragLightTPos);

    vec3 halfwayDir = normalize(lightDir + viewDir);

    float diff = max(dot(tangentNormal, lightDir), 0.0);
    float spec = pow(max(dot(tangentNormal, halfwayDir), 0.0), shininess);

    LightInfo light = lightInfos[0];
    vec3 ambient = light.ambient * albedo * 1.5; // Boost ambient to reduce dark areas
    // vec3 diffuse = light.diffuse * diff * albedo;
    vec3 diffuse = light.diffuse * albedo;
    vec3 specular = light.specular * spec * specularity;
    vec3 lighting = ambient + diffuse + specular;
	// Disc shadow calculations
	float maxShadow = 0.0;
	for (int i = 0; i < MAX_POINT_SHADOWS; i++) {
		PointShadow s = staticShadows[i];
		vec2 toShadowXZ = fragPos.xz - s.point.xy;
		float dist = length(toShadowXZ);
		float effectiveSoftness = s.strength * baseSoftness;
		float falloff = 1.0 - smoothstep(s.radius - effectiveSoftness, s.radius, dist);
		float shadow = falloff * s.strength;
		maxShadow = max(maxShadow, shadow);
	}
	for (int i = 0; i < MAX_POINT_SHADOWS; i++) {
		PointShadow s = dynamicShadows[i];
		vec2 toShadowXZ = fragPos.xz - s.point.xy;
		float dist = length(toShadowXZ);
		float effectiveSoftness = s.strength * baseSoftness;
		float falloff = 1.0 - smoothstep(s.radius - effectiveSoftness, s.radius, dist);
		float shadow = falloff * s.strength;
		maxShadow = max(maxShadow, shadow);
	}
	outColor = vec4(lighting * (1.0 - maxShadow), 1.0);
}
