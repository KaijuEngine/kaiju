#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 color;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;
layout(location = 2) out vec3 fragNormal;
layout(location = 3) out vec3 fragLightDirection;

void main() {
	fragColor = Color * color;
	fragTexCoords = UV0;
	fragNormal = Normal;
	vec4 wp = model * vec4(Position, 1.0);
	fragLightDirection = normalize(cameraPosition - wp.xyz);	
	gl_Position = projection * view * wp;
}
