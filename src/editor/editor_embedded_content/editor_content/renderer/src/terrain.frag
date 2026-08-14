#version 460
#define FRAGMENT_SHADER
#define HAS_GBUFFER

#define TERRAIN_WEIGHT_MAP_COUNT 2
#define TERRAIN_LAYER_COUNT 8
#define TERRAIN_LAYER_PARAM_COUNT 24
#define TERRAIN_LAYER_PARAMS_PER_LAYER 3
#define TERRAIN_ATLAS_COLUMNS 4
#define TERRAIN_ATLAS_ROWS 2
#define TERRAIN_ATLAS_TILE_SIZE 1024.0
#define TERRAIN_ATLAS_GUTTER 2.0
#define TERRAIN_MIN_ROUGHNESS 0.50
#define SAMPLER_COUNT 4

#ifndef TERRAIN_UNLIT_DEBUG
	#define SHADOW_SAMPLERS
#endif

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
#include "pbr_lighting.glsl"

layout(location = 9) flat in ivec4 fragTerrainLightIds;
layout(location = 10) flat in int fragTerrainInstance;

layout(set = 0, binding = 4) readonly buffer TerrainLayerBuffer {
	vec4 terrainLayerParams[][TERRAIN_LAYER_PARAM_COUNT];
};

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

mat3 terrainFallbackTBN(vec3 normal) {
	vec3 up = abs(normal.z) < 0.999 ? vec3(0.0, 0.0, 1.0) : vec3(0.0, 1.0, 0.0);
	vec3 tangent = normalize(cross(up, normal));
	vec3 bitangent = cross(normal, tangent);
	return mat3(tangent, bitangent, normal);
}

mat3 terrainCotangentFrame(vec3 normal, vec3 position, vec2 uv) {
	vec3 dp1 = dFdx(position);
	vec3 dp2 = dFdy(position);
	vec2 duv1 = dFdx(uv);
	vec2 duv2 = dFdy(uv);
	vec3 dp2Perp = cross(dp2, normal);
	vec3 dp1Perp = cross(normal, dp1);
	vec3 tangent = dp2Perp * duv1.x + dp1Perp * duv2.x;
	vec3 bitangent = dp2Perp * duv1.y + dp1Perp * duv2.y;
	float maxLength = max(dot(tangent, tangent), dot(bitangent, bitangent));
	if (maxLength <= 0.00000001) {
		return terrainFallbackTBN(normal);
	}
	float inverseLength = inversesqrt(maxLength);
	return mat3(tangent * inverseLength, bitangent * inverseLength, normal);
}

vec4 terrainLayerParameter(int layer, int parameter) {
	return terrainLayerParams[fragTerrainInstance][layer * TERRAIN_LAYER_PARAMS_PER_LAYER + parameter];
}

vec2 terrainLayerUv(int layer) {
	vec4 scaleOffset = terrainLayerParameter(layer, 0);
	vec4 rotationTint = terrainLayerParameter(layer, 1);
	vec2 uv = fragTexCoords * max(scaleOffset.xy, vec2(0.0001)) + scaleOffset.zw;
	uv -= vec2(0.5);
	uv = mat2(rotationTint.x, -rotationTint.y, rotationTint.y, rotationTint.x) * uv;
	return uv + vec2(0.5);
}

vec2 terrainAtlasUv(int layer, vec2 layerUv) {
	vec2 tile = vec2(float(layer % TERRAIN_ATLAS_COLUMNS), float(layer / TERRAIN_ATLAS_COLUMNS));
	float innerSize = TERRAIN_ATLAS_TILE_SIZE - TERRAIN_ATLAS_GUTTER * 2.0;
	vec2 pixel = tile * TERRAIN_ATLAS_TILE_SIZE + vec2(TERRAIN_ATLAS_GUTTER) +
		fract(layerUv) * (innerSize - 1.0) + vec2(0.5);
	return pixel / vec2(TERRAIN_ATLAS_TILE_SIZE * float(TERRAIN_ATLAS_COLUMNS),
		TERRAIN_ATLAS_TILE_SIZE * float(TERRAIN_ATLAS_ROWS));
}

void terrainWeights(out vec4 weights0, out vec4 weights1) {
	weights0 = max(texture(textures[0], fragTexCoords), vec4(0.0));
	weights1 = max(texture(textures[1], fragTexCoords), vec4(0.0));
	float totalWeight = dot(weights0, vec4(1.0)) + dot(weights1, vec4(1.0));
	if (totalWeight <= 0.001) {
		weights0 = vec4(1.0, 0.0, 0.0, 0.0);
		weights1 = vec4(0.0);
		return;
	}
	weights0 /= totalWeight;
	weights1 /= totalWeight;
}

float terrainLayerWeight(int layer, vec4 weights0, vec4 weights1) {
	return layer < 4 ? weights0[layer] : weights1[layer - 4];
}

void terrainMaterial(vec3 geometricNormal, out vec3 albedo, out vec3 normal, out float roughness) {
	vec4 weights0;
	vec4 weights1;
	terrainWeights(weights0, weights1);
	albedo = vec3(0.0);
	normal = vec3(0.0);
	roughness = 0.0;

	for (int layer = 0; layer < TERRAIN_LAYER_COUNT; ++layer) {
		float weight = terrainLayerWeight(layer, weights0, weights1);
		if (weight <= 0.0001) {
			continue;
		}
		vec2 layerUv = terrainLayerUv(layer);
		vec2 atlasUv = terrainAtlasUv(layer, layerUv);
		vec4 materialSample = texture(textures[2], atlasUv);
		vec4 rotationTint = terrainLayerParameter(layer, 1);
		vec4 tintTail = terrainLayerParameter(layer, 2);
		vec3 tint = vec3(rotationTint.zw, tintTail.x);
		albedo += pbrSrgbToLinear(materialSample.rgb) * max(tint, vec3(0.0)) * weight;
		roughness += materialSample.a * weight;

#ifndef TERRAIN_UNLIT_DEBUG
		vec3 tangentNormal = texture(textures[3], atlasUv).rgb * 2.0 - 1.0;
		if (dot(tangentNormal, tangentNormal) <= 0.0001) {
			tangentNormal = vec3(0.0, 0.0, 1.0);
		}
		mat3 tbn = terrainCotangentFrame(geometricNormal, fragPos, layerUv);
		normal += normalize(tbn * normalize(tangentNormal)) * weight;
#endif
	}
	albedo *= max(fragColor.rgb, vec3(0.0));
	// Stock terrain represents broad natural surfaces. Very low source values
	// turn dense normal-map detail into wet, sparkling plastic under the sun.
	roughness = clamp(roughness, TERRAIN_MIN_ROUGHNESS, 1.0);
#ifdef TERRAIN_UNLIT_DEBUG
	normal = geometricNormal;
#else
	normal = pbrSafeNormalize(normal, geometricNormal);
#endif
}

void main() {
	vec3 geometricNormal = pbrSafeNormalize(fragNormal, vec3(0.0, 1.0, 0.0));
	vec3 albedo;
	vec3 normal;
	float roughness;
	terrainMaterial(geometricNormal, albedo, normal, roughness);
	processGBuffer(normal);

#ifdef TERRAIN_UNLIT_DEBUG
	vec3 color = pbrLinearToSrgb(albedo);
#else
	const float metallic = 0.0;
	const float occlusion = 1.0;
	vec3 viewDirection = pbrSafeNormalize(cameraPosition.xyz - fragPos, geometricNormal);
	vec3 directLighting = vec3(0.0);
	vec3 ambientLighting = vec3(PBR_DEFAULT_AMBIENT_STRENGTH) * albedo * occlusion;
	for (int slot = 0; slot < 4; ++slot) {
		int lightIndex = fragTerrainLightIds[slot];
		if (lightIndex < 0 || lightIndex >= MAX_LIGHTS) {
			continue;
		}
		vec4 lightSpace = vertLights[lightIndex].matrix[0] * vec4(fragPos, 1.0);
		pbrAccumulateLight(lightIndex, lightSpace, albedo, normal, viewDirection,
			metallic, roughness, occlusion, directLighting, ambientLighting);
	}
	vec3 color = pbrFinalColor(ambientLighting, directLighting, vec3(0.0));
#endif
	color = applyBrushOverlay(color);
	processFinalColor(vec4(color, fragColor.a));
}
