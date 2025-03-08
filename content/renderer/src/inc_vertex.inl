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
	vec4 cameraPosition;	// w = [0=perspective, 1=orthographic]
	vec3 uiCameraPosition;
	vec2 screenSize;
	float time;
};

#define LOCATION_HEAD   8
#define LOCATION_START  LOCATION_HEAD + 4

layout(location = LOCATION_HEAD) in mat4 model;