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
#define LAYOUT_FRAG_BRUSH_CENTER_RADIUS 6
#define LAYOUT_FRAG_BRUSH_PARAMS 7
#define LAYOUT_FRAG_BRUSH_COLOR 8

#include "kaiju.glsl"

const vec3 sunLightDir = vec3(-0.5, -0.7, -0.5);
const vec3 sunLightColor = vec3(1.0, 0.95, 0.85);
const float ambientStrength = 0.45;

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

void main() {
	vec3 normal = normalize(fragNormal);
	vec4 texColor = texture(textures[0], fragTexCoords) * fragColor;
	float slope = 1.0 - clamp(normal.y, 0.0, 1.0);
	vec3 grassTint = vec3(0.55, 0.72, 0.42);
	vec3 rockTint = vec3(0.55, 0.52, 0.48);
	vec3 terrainColor = texColor.rgb * mix(grassTint, rockTint, smoothstep(0.25, 0.7, slope));
	terrainColor = applyBrushOverlay(terrainColor);
	float diff = max(dot(normal, -sunLightDir), 0.0);
	vec3 ambient = ambientStrength * sunLightColor * terrainColor;
	vec3 diffuse = diff * sunLightColor * terrainColor;
	processGBuffer(normal);
	processFinalColor(vec4(ambient + diffuse, texColor.a));
}
