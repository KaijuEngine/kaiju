#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_FLAGS 1

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_VIEW_DIR 5

#include "kaiju.glsl"

void main() {
	fragColor = Color * color;
	fragFlags = flags;
	fragTexCoords = UV0;
	fragNormal = mat3(model) * Normal;
	vec4 wp = worldPosition();
    fragPos = wp.xyz;
	fragViewDir = cameraPosition.xyz - wp.xyz;
	gl_Position = projection * view * wp;
}
