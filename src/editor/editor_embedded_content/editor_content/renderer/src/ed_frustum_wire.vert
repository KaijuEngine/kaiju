#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 color;
layout(location = LOCATION_START+1) in mat4 frustumProjection;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;

vec3 transformPoint(vec3 point) {
	vec4 pt0 = vec4(point.xyz, 1.0);
	vec4 res = frustumProjection * pt0;
	vec3 v3 = res.xyz;
	return v3 / res.w;
}

void main() {
	fragTexCoords = UV0;
	vec3 p = transformPoint(Position);
	vec4 wp = model * vec4(p, 1.0);
	fragColor = color;
	gl_Position = projection * view * wp;
}
