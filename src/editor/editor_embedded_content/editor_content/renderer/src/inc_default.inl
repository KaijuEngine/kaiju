#define NR_LIGHTS			4
#define MAX_LIGHTS			20
#define CUBEMAP_SIDES		6
#define PI              	3.14159265359

struct Light {
	mat4 matrix[CUBEMAP_SIDES];
	vec3 position;
	vec3 direction;
};

struct PointShadow {
	vec2 point; // X,Z
	float radius;
	float strength;
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

// cameraPosition.w = [0=perspective, 1=orthographic]
layout(set = 0, binding = 0) readonly uniform UniformBufferObject {
	mat4 view;
	mat4 projection;
	mat4 uiView;
	mat4 uiProjection;
	vec4 cameraPosition;
	vec3 uiCameraPosition;
	vec2 screenSize;
	float time;
	Light vertLights[MAX_LIGHTS];
	LightInfo lightInfos[MAX_LIGHTS];
};
