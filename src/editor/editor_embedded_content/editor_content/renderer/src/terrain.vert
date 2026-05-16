#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_BRUSH_CENTER_RADIUS 1
#define LAYOUT_VERT_BRUSH_PARAMS 2
#define LAYOUT_VERT_BRUSH_COLOR 3
#define LAYOUT_VERT_FLAGS 4

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_VIEW_DIR 5
#define LAYOUT_FRAG_BRUSH_CENTER_RADIUS 6
#define LAYOUT_FRAG_BRUSH_PARAMS 7
#define LAYOUT_FRAG_BRUSH_COLOR 8

#include "kaiju.glsl"

void main() {
	fragColor = Color * color;
	fragFlags = flags;
	fragTexCoords = UV0;
	fragNormal = normalize(transpose(inverse(mat3(model))) * Normal);
	fragBrushCenterRadius = brushCenterRadius;
	fragBrushParams = brushParams;
	fragBrushColor = brushColor;
	writeStandardPosition();
	vec4 wp = worldPosition();
	fragViewDir = cameraPosition.xyz - wp.xyz;
}
