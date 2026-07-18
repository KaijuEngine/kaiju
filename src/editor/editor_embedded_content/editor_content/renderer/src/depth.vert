#version 460
#define VERTEX_SHADER

#include "kaiju.glsl"

layout(push_constant) uniform Push {
    int CascadeIndex;
};

layout(location = LOCATION_START) in int lightIndex;

void main() {
	if (lightIndex < 0 || lightIndex >= MAX_LIGHTS) {
		gl_Position = vec4(2.0, 2.0, 2.0, 1.0);
		return;
	}
	mat4 fragLightSpaceMatrix = vertLights[lightIndex].matrix[CascadeIndex];
    gl_Position = fragLightSpaceMatrix * model * vec4(Position, 1.0);
}
