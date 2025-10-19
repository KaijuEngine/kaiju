#version 460

#include "inc_default.inl"

layout (triangles) in;
layout (triangle_strip, max_vertices=18) out;

layout(location = 0) in VertexData {
	mat4 projection;
	int lightIndex;
	vec4 fragPos;
	vec3 fragSourcePos;
	float fragShadowFar;
} vert[];

layout(location = 0) out vec4 psFragPos;
layout(location = 1) out vec3 psFragSourcePos;
layout(location = 2) out float psFragShadowFar;

mat4 look_at(vec3 eye, vec3 center, vec3 up) {
	vec3 f = normalize(center - eye);
	vec3 s = normalize(cross(f, up));
	vec3 u = cross(s, f);
	return mat4(
		s.x, u.x, -f.x, 0.0,
		s.y, u.y, -f.y, 0.0,
		s.z, u.z, -f.z, 0.0,
		-dot(s, eye), -dot(u, eye), dot(f, eye), 1.0
	);
}

void main() {
	int lightIndex = vert[0].lightIndex;
	mat4 proj = vert[0].projection;
	vec3 lp = vertLights[lightIndex].position;
    vec3 lookSides[CUBEMAP_SIDES] = {
		vec3(lp.x + 1.0, lp.y, lp.z),
		vec3(lp.x - 1.0, lp.y, lp.z),
		vec3(lp.x, lp.y + 1.0, lp.z),
		vec3(lp.x, lp.y - 1.0, lp.z),
		vec3(lp.x, lp.y, lp.z + 1.0),
		vec3(lp.x, lp.y, lp.z - 1.0)
	};
	vec3 lookUps[CUBEMAP_SIDES] = {
		vec3(0.0, -1.0, 0.0),
		vec3(0.0, -1.0, 0.0),
		vec3(0.0, 0.0, 1.0),
		vec3(0.0, 0.0, -1.0),
		vec3(0.0, -1.0, 0.0),
		vec3(0.0, -1.0, 0.0)
	};
    for (int face = 0; face < CUBEMAP_SIDES; ++face) {
		// Built-in and specifies to which face we render
		gl_Layer = face;
		mat4 lightSpace = proj * look_at(lp, lookSides[face], lookUps[face]);
		// Each triangle vertex
		for (int i = 0; i < 3; ++i) {
			psFragPos = gl_in[i].gl_Position;
			gl_Position = lightSpace * psFragPos;
			psFragSourcePos = vert[0].fragSourcePos;
			psFragShadowFar = vert[0].fragShadowFar;
			EmitVertex();
		}
		EndPrimitive();
	}
}
