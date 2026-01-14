#version 460
#define FRAGMENT_SHADER

#define SAMPLER_COUNT 1

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1

#include "kaiju.glsl"

void main(void) {
	vec4 texColor = texture(textures[0], fragTexCoords) * fragColor;
    processFinalColor(texColor);
}
