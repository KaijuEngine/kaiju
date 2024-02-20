#version 460

layout (location = 0) in vec3 Position;
layout (location = 1) in vec3 Normal;
layout (location = 2) in vec4 Tangent;
layout (location = 3) in vec2 UV0;
layout (location = 4) in vec4 Color;
layout (location = 5) in ivec4 JointIds;
layout (location = 6) in vec4 JointWeights;
layout (location = 7) in vec3 MorphTarget;

layout(set = 0, binding = 0) readonly uniform UniformBufferObject {
	mat4 view;
	mat4 projection;
	mat4 uiView;
	mat4 uiProjection;
	vec3 cameraPosition;
	vec3 uiCameraPosition;
	vec2 screenSize;
	float time;
} globalData;

layout(location = 8) in mat4 model;
layout(location = 12) in vec4 color;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;

void main() {
	fragColor = Color * color;
	fragTexCoords = UV0;
	vec3 pos = vec3(Position.x, -Position.y, Position.z) * 2.0;
	gl_Position = vec4(pos, 1.0);
}
