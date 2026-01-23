#version 460
#define VERTEX_SHADER

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_TEX_COORDS 2
#define LAYOUT_FRAG_NORMAL 3

#include "kaiju.glsl"

layout(location = LOCATION_START+0) in mat4 previewView;
layout(location = LOCATION_START+4) in mat4 previewProjection;

void main() {
	fragColor = Color;
	fragTexCoords = UV0;
	fragNormal = mat3(model) * Normal;
	gl_Position = previewProjection * previewView * model * vec4(Position, 1.0);
}
