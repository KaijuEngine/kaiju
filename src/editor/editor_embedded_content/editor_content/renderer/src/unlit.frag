#version 460
#define FRAGMENT_SHADER

#define SAMPLER_COUNT 1

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1
#define LAYOUT_FRAG_FLAGS 2
#define LAYOUT_FRAG_POS 3
#define LAYOUT_FRAG_NORMAL 4

#include "kaiju.glsl"

void main() {
    vec4 unWeightedColor = texture(textures[0], fragTexCoords) * fragColor;
    processFinalColor(unWeightedColor);
}
