#version 460
#define FRAGMENT_SHADER
#define HAS_GBUFFER

#define SAMPLER_COUNT   1

#define SHADOW_SAMPLERS

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_VIEW_DIR 5

#include "kaiju.glsl"

// Hardcoded sun light (directional light)
const vec3 sunLightDir = vec3(-0.5, -0.7, -0.5);
const vec3 sunLightColor = vec3(1.0, 0.95, 0.85);

// Hardcoded material properties
const vec3 materialSpecular = vec3(0.5, 0.5, 0.5);
const float materialShininess = 32.0;
const float ambientStrength = 0.5;

void main() {
	vec4 texColor = texture(textures[0], fragTexCoords) * fragColor;
	vec3 normal = normalize(fragNormal);
    processGBuffer(normal);
	// Ambient
    vec3 ambient = ambientStrength * sunLightColor * texColor.rgb;
    // Diffuse
    float diff = max(dot(normal, -sunLightDir), 0.0);
    vec3 diffuse = diff * sunLightColor * texColor.rgb;
    // Specular (Blinn-Phong)
	/*
    vec3 halfwayDir = normalize(sunLightDir + fragViewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), materialShininess);
    vec3 specular = spec * sunLightColor * materialSpecular;
	*/
    // Combine
    vec3 result = ambient + diffuse/* + specular*/;
    processFinalColor(vec4(result, texColor.a));
}
