#version 460
#define FRAGMENT_SHADER

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_TEX_COORDS 2
#define LAYOUT_FRAG_NORMAL 3

#include "kaiju.glsl"

// Hardcoded sun light (directional light)
const vec3 sunLightDir = vec3(-0.5, -0.7, -0.5);
const vec3 sunLightColor = vec3(1.0, 0.95, 0.85);

// Hardcoded material properties
const float ambientStrength = 0.5;

void main() {
	vec4 texColor = fragColor;
	vec3 normal = normalize(fragNormal);
    vec3 ambient = ambientStrength * sunLightColor * texColor.rgb;
    float diff = max(dot(normal, -sunLightDir), 0.0);
    vec3 diffuse = diff * sunLightColor * texColor.rgb;
    vec3 result = ambient + diffuse;
    processFinalColor(vec4(result, texColor.a));
}
