#version 460
#define FRAGMENT_SHADER
#define HAS_GBUFFER

#define SAMPLER_COUNT 1

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

vec3 applyBrushOverlay(vec3 terrainColor) {
	if (fragBrushCenterRadius.w <= 0.0) {
		return terrainColor;
	}
	float radius = max(fragBrushCenterRadius.z, 0.001);
	float ringWidth = max(fragBrushParams.x, 0.001);
	float fillAlpha = clamp(fragBrushParams.y, 0.0, 1.0);
	float ringAlpha = clamp(fragBrushParams.z, 0.0, 1.0);
	float dist = distance(fragPos.xz, fragBrushCenterRadius.xy);
	float edgeFeather = max(fwidth(dist), 0.001);
	float fill = (1.0 - smoothstep(radius - ringWidth, radius, dist)) * fillAlpha;
	float ring = (1.0 - smoothstep(ringWidth, ringWidth + edgeFeather, abs(dist - radius))) * ringAlpha;
	float alpha = clamp(max(fill, ring) * fragBrushColor.a, 0.0, 1.0);
	return mix(terrainColor, fragBrushColor.rgb, alpha);
}

vec3 terrainAlbedo(vec3 sampledColor, vec3 normal) {
	float slope = 1.0 - clamp(normal.y, 0.0, 1.0);
	float slopePower = max(fragMaterialParams.z, 0.001);
	float blend = smoothstep(fragSlopeParams.x, fragSlopeParams.y, pow(slope, slopePower));
	vec3 slopeTint = mix(fragGrassTint.rgb, fragRockTint.rgb, blend);
	vec3 textured = mix(vec3(1.0), sampledColor, clamp(fragMaterialParams.x, 0.0, 1.0));
	vec3 tinted = mix(vec3(1.0), slopeTint, clamp(fragMaterialParams.y, 0.0, 1.0));
	return textured * tinted;
}

void main() {
	vec3 normal = normalize(fragNormal);
	vec4 texColor = texture(textures[0], fragTexCoords) * fragColor;
	vec3 terrainColor = terrainAlbedo(texColor.rgb, normal);
	terrainColor = applyBrushOverlay(terrainColor);
	processGBuffer(normal);

#ifdef TERRAIN_UNLIT_DEBUG
	processFinalColor(vec4(terrainColor, texColor.a));
#else
	vec3 lightDir = fragLightDirectionAmbient.xyz;
	if (length(lightDir) <= 0.001) {
		lightDir = vec3(-0.5, -0.7, -0.5);
	}
	lightDir = normalize(lightDir);
	vec3 lightColor = fragLightColorDiffuse.rgb;
	float ambientStrength = clamp(fragLightDirectionAmbient.w, 0.0, 1.0);
	float diffuseStrength = max(fragLightColorDiffuse.a, 0.0);
	float diff = max(dot(normal, -lightDir), 0.0) * diffuseStrength;
	vec3 ambient = ambientStrength * lightColor * terrainColor;
	vec3 diffuse = diff * lightColor * terrainColor;
	processFinalColor(vec4(ambient + diffuse, texColor.a));
#endif
}
