#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 color;

layout(location = 0) out vec4 fragColor;

void main() {
	fragColor = vec4(color.rgb, 1.0);
	float outlineWidth = color.a;
	float Sx = length(vec3(model[0])); // X-axis scaling
	float Sy = length(vec3(model[1])); // Y-axis scaling
	float Sz = length(vec3(model[2])); // Z-axis scaling
	vec4 pWorld = model * vec4(Position, 1.0);
	vec4 pView = view * pWorld;
	float D = -pView.z;
	vec3 scaledPosition = vec3(
		Position.x * (1.0 + outlineWidth * D / Sx),
		Position.y * (1.0 + outlineWidth * D / Sy),
		Position.z * (1.0 + outlineWidth * D / Sz)
	);
	vec4 pWorldScaled = model * vec4(scaledPosition, 1.0);
	// Project to clip space
	gl_Position = projection * view * pWorldScaled;

	// TODO:  This needs to be fixed for 2D view
}
