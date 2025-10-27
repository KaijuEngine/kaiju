#version 460

#include "inc_default.inl"

layout(location = 0) in vec4 fragColor;
layout(location = 1) in vec2 fragTexCoords;
layout(location = 2) in vec3 fragWorldPos;

layout(location = 0) out vec4 outColor;

void main() {
	float fadeScale = 3.0;
	float yCheck = fragColor.y + 0.0001;
	if (fragColor.z > yCheck || fragColor.x > yCheck) {
		fadeScale = 20;
	}
	float fade = fadeScale / length(cameraPosition.xyz - fragWorldPos.xyz);
	outColor = vec4(fragColor.xyz, fade);
}
