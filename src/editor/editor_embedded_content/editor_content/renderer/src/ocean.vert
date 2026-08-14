#version 460
#define VERTEX_SHADER

#define LAYOUT_VERT_FLAGS 3
#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_POS 2
#define LAYOUT_FRAG_NORMAL 3
#define LAYOUT_FRAG_TEX_COORDS 4
#define LAYOUT_FRAG_FLAGS 11

#include "kaiju.glsl"

layout(location = LOCATION_START + 0) in vec4 shallowColor;
layout(location = LOCATION_START + 1) in vec4 deepColor;
layout(location = LOCATION_START + 2) in vec4 waveParams;
layout(location = LOCATION_START + 4) in ivec4 oceanLightIds;
layout(location = LOCATION_START + 5) in vec4 oceanBrushCenterRadius;
layout(location = LOCATION_START + 6) in vec4 oceanBrushParams;
layout(location = LOCATION_START + 7) in vec4 oceanBrushColor;

layout(location = 1) flat out vec4 fragDeepColor;
layout(location = 5) flat out ivec4 fragOceanLightIds;
layout(location = 6) flat out vec4 fragOceanBrushCenterRadius;
layout(location = 7) flat out vec4 fragOceanBrushParams;
layout(location = 8) flat out vec4 fragOceanBrushColor;
layout(location = 10) flat out vec4 fragOceanWaveParams;

const int OCEAN_WAVE_COUNT = 6;
const vec2 oceanWaveDirections[OCEAN_WAVE_COUNT] = vec2[](
	vec2(0.8192, 0.5735), vec2(-0.4229, 0.9062),
	vec2(0.1736, -0.9848), vec2(-0.9397, -0.3420),
	vec2(0.6428, -0.7660), vec2(-0.0872, 0.9962)
);
const float oceanWaveFrequencies[OCEAN_WAVE_COUNT] = float[](
	0.61, 0.89, 1.27, 1.83, 2.57, 3.41
);
const float oceanWaveSpeeds[OCEAN_WAVE_COUNT] = float[](
	0.43, -0.57, 0.71, -0.86, 1.03, -1.21
);
const float oceanWaveWeights[OCEAN_WAVE_COUNT] = float[](
	0.29, 0.23, 0.18, 0.13, 0.10, 0.07
);

float oceanHeight(vec2 position, float amplitude, float speed, float frequency) {
	float height = 0.0;
	for (int wave = 0; wave < OCEAN_WAVE_COUNT; ++wave) {
		float phase = dot(position, oceanWaveDirections[wave]) * frequency *
			oceanWaveFrequencies[wave] + time * speed * oceanWaveSpeeds[wave];
		height += sin(phase) * oceanWaveWeights[wave];
	}
	return amplitude * height;
}

vec3 oceanNormal(vec2 position, float amplitude, float speed, float frequency) {
	vec2 derivative = vec2(0.0);
	for (int wave = 0; wave < OCEAN_WAVE_COUNT; ++wave) {
		float waveFrequency = frequency * oceanWaveFrequencies[wave];
		float phase = dot(position, oceanWaveDirections[wave]) * waveFrequency +
			time * speed * oceanWaveSpeeds[wave];
		derivative += cos(phase) * waveFrequency * oceanWaveDirections[wave] *
			oceanWaveWeights[wave];
	}
	derivative *= amplitude;
	return normalize(vec3(-derivative.x, 1.0, -derivative.y));
}

void main() {
	vec4 world = model * vec4(Position, 1.0);
	float amplitude = max(waveParams.x, 0.0);
	float speed = max(waveParams.y, 0.0);
	float frequency = max(waveParams.z, 0.001);
	world.y += oceanHeight(world.xz, amplitude, speed, frequency);

	fragColor = shallowColor;
	fragDeepColor = deepColor;
	fragPos = world.xyz;
	fragNormal = normalize(oceanNormal(world.xz, amplitude, speed, frequency));
	fragTexCoords = UV0;
	fragFlags = flags;
	fragOceanLightIds = oceanLightIds;
	fragOceanBrushCenterRadius = oceanBrushCenterRadius;
	fragOceanBrushParams = oceanBrushParams;
	fragOceanBrushColor = oceanBrushColor;
	fragOceanWaveParams = waveParams;
	gl_Position = projection * view * world;
}
