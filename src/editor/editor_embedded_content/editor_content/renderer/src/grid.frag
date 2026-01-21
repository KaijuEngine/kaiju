#version 460
#define FRAGMENT_SHADER

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1
#define LAYOUT_FRAG_WORLD_POS 2

#include "kaiju.glsl"

void main() {
	vec4 color = fragColor;
	if (color.a + 0.00001 >= 1.0) {
		float fadeScale = 3.0;
		float yCheck = color.y + 0.0001;
		if (color.z > yCheck || color.x > yCheck) {
			fadeScale = 20;
		}
		color.a = fadeScale / length(cameraPosition.xyz - fragWorldPos.xyz);
	}
	outColor = color;
}
