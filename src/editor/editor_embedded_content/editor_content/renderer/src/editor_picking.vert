#version 460
#define MAX_LIGHTS 20

struct Light {
	mat4 matrix[6];
	vec3 position;
	vec3 direction;
};

struct LightInfo {
	vec3 position;
	float intensity;
	vec3 direction;
	float cutoff;
	vec3 ambient;
	float outerCutoff;
	vec3 diffuse;
	float constant;
	vec3 specular;
	float linear;
	float quadratic;
	float nearPlane;
	float farPlane;
	int type;
};

layout(set = 0, binding = 0) readonly uniform UniformBufferObject {
	mat4 view;
	mat4 projection;
	mat4 uiView;
	mat4 uiProjection;
	vec4 cameraPosition;
	vec3 uiCameraPosition;
	float time;
	vec2 screenSize;
	int cascadeCount;
	vec4 cascadePlaneDistances;
	Light vertLights[MAX_LIGHTS];
	LightInfo lightInfos[MAX_LIGHTS];
};

layout(location = 0) in vec3 Position;
layout(location = 1) in vec3 Normal;
layout(location = 2) in vec4 Tangent;
layout(location = 3) in vec2 UV0;
layout(location = 4) in vec4 Color;
layout(location = 5) in ivec4 JointIds;
layout(location = 6) in vec4 JointWeights;
layout(location = 7) in vec3 MorphTarget;
layout(location = 8) in mat4 model;
layout(location = 12) in uint pickID;

layout(location = 0) flat out uint fragPickID;

void main() {
	fragPickID = pickID;
#ifdef BILLBOARD
	vec4 centerWorld = model * vec4(0.0, 0.0, 0.0, 1.0);
	vec3 camRight = vec3(view[0].x, view[1].x, view[2].x);
	vec3 camUp    = vec3(view[0].y, view[1].y, view[2].y);
	vec3 offset = camRight * Position.x + camUp * Position.y;
	vec4 worldPos = centerWorld + vec4(offset, 0.0);
	gl_Position = projection * view * worldPos;
#else
	gl_Position = projection * view * model * vec4(Position, 1.0);
#endif
}
