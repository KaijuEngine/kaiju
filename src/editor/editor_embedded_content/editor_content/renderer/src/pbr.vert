#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR
#define LAYOUT_VERT_METALLIC
#define LAYOUT_VERT_ROUGHNESS
#define LAYOUT_VERT_EMISSIVE
#define LAYOUT_VERT_FLAGS

#define LAYOUT_FRAG_COLOR
#define LAYOUT_FRAG_FLAGS
#define LAYOUT_FRAG_POS
#define LAYOUT_FRAG_TEX_COORDS
#define LAYOUT_FRAG_NORMAL
#define LAYOUT_FRAG_METALLIC
#define LAYOUT_FRAG_ROUGHNESS
#define LAYOUT_FRAG_EMISSIVE

#define LAYOUT_ALL_LIGHT_REQUIREMENTS

#include "kaiju.glsl"

void main() {
	fragMetallic = metallic;
	fragRoughness = roughness;
	fragEmissive = emissive;
	fragFlags = flags;
	fragTexCoords = UV0;
	fragColor = color * Color;
	fragPos = vec3(model * vec4(Position, 1.0));
	gl_Position = projection * view * worldPosition();
	calcVertexLightInformation();
}
