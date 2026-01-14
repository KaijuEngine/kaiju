#version 460
#define FRAGMENT_SHADER

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1

#include "kaiju.glsl"

void main() {
	outColor = fragColor;
}
