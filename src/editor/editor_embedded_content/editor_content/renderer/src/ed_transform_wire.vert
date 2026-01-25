#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1

#include "kaiju.glsl"

void main() {
	fragColor = color;
	writeTexCoords();
	writeStandardPosition();
}
