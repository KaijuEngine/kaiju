#version 460
#define FRAGMENT_SHADER

#define SAMPLER_COUNT 1

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1

#include "kaiju.glsl"

void main() {
	vec4 tex = texture(textures[0], fragTexCoords);
	float v = tex.r;
	vec4 unWeightedColor = vec4(v * fragColor.rgb, v * fragColor.a);
    processFinalColor(unWeightedColor);
}
