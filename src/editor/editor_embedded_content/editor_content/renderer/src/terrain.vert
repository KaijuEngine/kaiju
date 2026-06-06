#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_COLOR 0
#define LAYOUT_VERT_UVS 1
#define LAYOUT_VERT_TERRAIN_SLOPE_PARAMS 2
#define LAYOUT_VERT_TERRAIN_GRASS_TINT 3
#define LAYOUT_VERT_TERRAIN_ROCK_TINT 4
#define LAYOUT_VERT_TERRAIN_LIGHT_DIRECTION_AMBIENT 5
#define LAYOUT_VERT_TERRAIN_LIGHT_COLOR_DIFFUSE 6
#define LAYOUT_VERT_TERRAIN_MATERIAL_PARAMS 7
#define LAYOUT_VERT_BRUSH_CENTER_RADIUS 8
#define LAYOUT_VERT_BRUSH_PARAMS 9
#define LAYOUT_VERT_BRUSH_COLOR 10
#define LAYOUT_VERT_FLAGS 11

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_FLAGS 1
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_TEX_COORDS 3
#define LAYOUT_FRAG_NORMAL 4
#define LAYOUT_FRAG_VIEW_DIR 5
#define LAYOUT_FRAG_TERRAIN_SLOPE_PARAMS 6
#define LAYOUT_FRAG_TERRAIN_GRASS_TINT 7
#define LAYOUT_FRAG_TERRAIN_ROCK_TINT 8
#define LAYOUT_FRAG_TERRAIN_LIGHT_DIRECTION_AMBIENT 9
#define LAYOUT_FRAG_TERRAIN_LIGHT_COLOR_DIFFUSE 10
#define LAYOUT_FRAG_TERRAIN_MATERIAL_PARAMS 11
#define LAYOUT_FRAG_BRUSH_CENTER_RADIUS 12
#define LAYOUT_FRAG_BRUSH_PARAMS 13
#define LAYOUT_FRAG_BRUSH_COLOR 14

#include "kaiju.glsl"

void main() {
	fragColor = Color * color;
	fragFlags = flags;
	fragTexCoords = UV0 * max(uvs.zw, vec2(0.0001)) + uvs.xy;
	fragNormal = normalize(transpose(inverse(mat3(model))) * Normal);
	fragSlopeParams = slopeParams;
	fragGrassTint = grassTint;
	fragRockTint = rockTint;
	fragLightDirectionAmbient = lightDirectionAmbient;
	fragLightColorDiffuse = lightColorDiffuse;
	fragMaterialParams = materialParams;
	fragBrushCenterRadius = brushCenterRadius;
	fragBrushParams = brushParams;
	fragBrushColor = brushColor;
	writeStandardPosition();
	vec4 wp = worldPosition();
	fragViewDir = cameraPosition.xyz - wp.xyz;
}
