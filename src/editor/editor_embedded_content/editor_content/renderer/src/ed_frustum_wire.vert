#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_FRUSTUM_PROJECTION 1

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1

#include "kaiju.glsl"

void main() {
	fragColor = color;
	writeTexCoords();
	vec3 p = transformPoint(frustumProjection, Position);
	vec4 wp = model * vec4(p, 1.0);
	gl_Position = projection * view * wp;
}
