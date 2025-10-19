#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in int lightIndex;

void main() {
#ifdef MULTI_LIGHT
	mat4 fragLightSpaceMatrix = vertLights[lightIndex].matrix[0];
#else
	mat4 fragLightSpaceMatrix = vertLights[0].matrix[0];
#endif
	gl_Position = fragLightSpaceMatrix * model * vec4(Position, 1.0);
}
