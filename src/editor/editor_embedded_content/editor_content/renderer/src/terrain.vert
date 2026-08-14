#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_UVS 1
#define LAYOUT_VERT_BRUSH_CENTER_RADIUS 2
#define LAYOUT_VERT_BRUSH_PARAMS 3
#define LAYOUT_VERT_BRUSH_COLOR 4
#define LAYOUT_VERT_FLAGS 5

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

layout(location = LOCATION_START + 6) in ivec4 terrainLightIds;
layout(location = 9) flat out ivec4 fragTerrainLightIds;
layout(location = 10) flat out int fragTerrainInstance;

void main() {
	fragColor = Color * color;
	fragFlags = flags;
	fragTexCoords = UV0 * max(uvs.zw, vec2(0.0001)) + uvs.xy;
	fragNormal = normalize(transpose(inverse(mat3(model))) * Normal);
	fragBrushCenterRadius = brushCenterRadius;
	fragBrushParams = brushParams;
	fragBrushColor = brushColor;
	fragTerrainLightIds = terrainLightIds;
	fragTerrainInstance = gl_InstanceIndex;
	writeStandardPosition();
	fragViewDir = cameraPosition.xyz - fragPos;
}
