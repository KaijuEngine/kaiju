#version 460

#define AMBIENT_LIGHT_COLOR vec3(0.05, 0.05, 0.05)

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;
layout(location = 2) in vec3 fragNormal;
layout(location = 3) in vec3 fragLightDirection;

layout(binding = 1) uniform sampler2D texSampler;

layout(location = 0) out vec4 outColor;
#ifdef OIT
layout(location = 1) out float reveal;
#endif

void main() {
    vec4 baseColor = texture(texSampler, fragTexCoords) * fragColor;
    vec3 normal = normalize(fragNormal);
    float diff = max(dot(normal, fragLightDirection), 0.0);
    vec4 diffuseColor = baseColor * vec4(vec3(diff), 1.0);
    vec4 ambientComponent = vec4(AMBIENT_LIGHT_COLOR, 1.0) * baseColor;
    vec4 unWeightedColor = diffuseColor + ambientComponent;
#include "inc_fragment_oit_block.inl"
}
