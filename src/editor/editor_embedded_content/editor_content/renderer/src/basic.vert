#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR
#define LAYOUT_VERT_FLAGS

#define LAYOUT_FRAG_COLOR
#define LAYOUT_FRAG_FLAGS
#define LAYOUT_FRAG_POS
#define LAYOUT_FRAG_TEX_COORDS
#define LAYOUT_FRAG_NORMAL
#define LAYOUT_FRAG_VIEW_DIR

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
