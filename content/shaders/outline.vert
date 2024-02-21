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
};

layout(location = 8) in mat4 model;
layout(location = 12) in vec4 color;

layout(location = 0) out vec4 fragColor;

void main() {
	float outlineWidth = color.a;
	mat4 vp = projection * view;
	vec4 clip = vp * model * vec4(Position, 1.0);
	vec3 clipNml = mat3(vp) * mat3(model) * Normal;
	vec2 offset = normalize(clipNml.xy) / screenSize.xy * outlineWidth * clip.w * 2;
	clip.xy += offset;
	fragColor = color;
	gl_Position = clip;
}
