#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in int lightIndex;
layout(location = LOCATION_START+1) in vec3 sourcePos;
layout(location = LOCATION_START+2) in float shadowFar;
layout(location = LOCATION_START+3) in mat4 lightProjection;

layout(location = 0) out VertexData {
	mat4 projection;
	int lightIndex;
	vec4 fragPos;
	vec3 fragSourcePos;
	float fragShadowFar;
} vert;

void main() {
	vert.fragPos = model * vec4(Position, 1.0);
	gl_Position = vert.fragPos;
	vert.fragSourcePos = sourcePos;
	vert.fragShadowFar = shadowFar;
	vert.lightIndex = lightIndex;
	vert.projection = lightProjection;
}
