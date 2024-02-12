#version 460
//#version 300 es
//precision mediump float;

#ifdef VULKAN
	layout(location = 0) in vec4 fragColor;
	layout(location = 1) in vec2 fragTexCoords;
#else
	in vec4 fragColor;
	in vec2 fragTexCoords;
#endif

layout(location = 0) out vec4 outColor;

void main() {
	outColor = fragColor;
}
