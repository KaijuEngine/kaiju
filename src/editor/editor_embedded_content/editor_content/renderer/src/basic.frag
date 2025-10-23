#version 460

#define AMBIENT_LIGHT_COLOR vec3(0.05, 0.05, 0.05)

layout(location = 0) in vec4 fragColor;
layout(location = 1) in flat uint fragFlags;
layout(location = 2) in vec3 fragPos;
layout(location = 3) in vec2 fragTexCoords;
layout(location = 4) in vec3 fragNormal;
layout(location = 5) in vec3 viewDir;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#else
layout(location = 1) out vec4 outPosition;
layout(location = 2) out vec4 outNormal;
#endif

// Hardcoded sun light (directional light)
const vec3 sunLightDir = vec3(-0.5, -0.7, -0.5);
const vec3 sunLightColor = vec3(1.0, 0.95, 0.85);

// Hardcoded material properties
const vec3 materialSpecular = vec3(0.5, 0.5, 0.5);
const float materialShininess = 32.0;
const float ambientStrength = 0.5;

void main() {
	vec4 texColor = texture(texSampler, fragTexCoords) * fragColor;
	vec3 normal = normalize(fragNormal);
#ifndef OIT
    outPosition = vec4(fragPos, 0.0);
    outNormal = vec4(normal, 0.0);
    if ((fragFlags & 0x00000001) != 0) {
        uint packed = packHalf2x16(vec2(outPosition.w, 0.0));
        uint alphaBits = packed & uint(0xFFFF);
        alphaBits |= uint(1) << uint(1);
        uint newPacked = alphaBits;
        vec2 newHalves = unpackHalf2x16(newPacked);
        outPosition.w = newHalves.x;
    }
#endif
	// Ambient
    vec3 ambient = ambientStrength * sunLightColor * texColor.rgb;
    // Diffuse
    float diff = max(dot(normal, -sunLightDir), 0.0);
    vec3 diffuse = diff * sunLightColor * texColor.rgb;
    // Specular (Blinn-Phong)
	/*
    vec3 halfwayDir = normalize(sunLightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), materialShininess);
    vec3 specular = spec * sunLightColor * materialSpecular;
	*/
    // Combine
    vec3 result = ambient + diffuse/* + specular*/;
	vec4 unWeightedColor = vec4(result, texColor.a);
#include "inc_fragment_oit_block.inl"
}
