#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_UVS 1
#define LAYOUT_VERT_FLAGS 2

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1
#define LAYOUT_FRAG_FLAGS 2
#define LAYOUT_FRAG_POS 3
#define LAYOUT_FRAG_NORMAL 4

#include "kaiju.glsl"

void main() {
	fragColor = Color * color;
	fragFlags = flags;
	vec2 uv = UV0;
	uv *= uvs.zw;
	uv.y += (1.0 - uvs.w) - uvs.y;
	uv.x += uvs.x;
	fragTexCoords = uv;
	fragNormal = mat3(model) * Normal;
	writeStandardPosition();
}


