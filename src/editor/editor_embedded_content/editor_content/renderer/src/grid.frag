#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;
layout(location = 2) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

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
