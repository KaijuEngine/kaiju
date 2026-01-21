#version 460
#define VERTEX_SHADER

#include "kaiju.glsl"

layout(push_constant) uniform Push {
    int CascadeIndex;
};

layout(location = LOCATION_START) in int lightIndex;

void main() {
    mat4 fragLightSpaceMatrix = vertLights[lightIndex].matrix[CascadeIndex];
    gl_Position = fragLightSpaceMatrix * model * vec4(Position, 1.0);
}
