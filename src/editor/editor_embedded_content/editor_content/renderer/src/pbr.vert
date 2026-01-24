#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_METALLIC_ROUGHNESS_EMISSIVE_ALBEDO 1
#define LAYOUT_VERT_FLAGS 2
#define LAYOUT_VERT_LIGHT_IDS 3

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_METALLIC 5
#define LAYOUT_FRAG_ROUGHNESS 6
#define LAYOUT_FRAG_EMISSIVE 7

#define LAYOUT_ALL_LIGHT_REQUIREMENTS 8

#include "kaiju.glsl"

void main() {
	fragMetallic = meRoEmAo.r;
	fragRoughness = meRoEmAo.g;
	fragEmissive = meRoEmAo.b;
	fragFlags = flags;
	fragTexCoords = UV0;
	fragColor = color * Color;
	fragPos = vec3(model * vec4(Position, 1.0));
	gl_Position = projection * view * worldPosition();
	calcVertexLightInformation();
}
