#version 460
#define VERTEX_SHADER

#define BILLBOARD

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_UVS 1

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1

#include "kaiju.glsl"

void main() {
	fragColor = Color * color;
	writeTexCoords();
	writeStandardPosition();
}
