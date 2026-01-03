#version 460

#include "inc_vertex.inl"

#define MAX_JOINTS			50
#define MAX_SKIN_INSTANCES	50

#ifdef SKINNING
layout(set = 0, binding = 2) readonly buffer SkinnedSSBO {
	mat4 jointTransforms[MAX_SKIN_INSTANCES][MAX_JOINTS];
};
#endif

layout(location = LOCATION_START) in vec4 color;
#ifdef SKINNING
layout(location = LOCATION_START+1) in int skinIndex;
layout(location = LOCATION_START+2) in uint flags;
#else
layout(location = LOCATION_START+1) in uint flags;
#endif

layout(location = 0) out vec4 fragColor;
layout(location = 1) out uint fragFlags;
layout(location = 2) out vec3 fragPos;
layout(location = 3) out vec2 fragTexCoords;
layout(location = 4) out vec3 fragNormal;
layout(location = 5) out vec3 viewDir;

void main() {
	fragColor = Color * color;
	fragFlags = flags;
	fragTexCoords = UV0;
	fragNormal = mat3(model) * Normal;
#ifdef SKINNING
	vec4 pos = vec4(Position, 1.0);
	mat4 skinMatrix = JointWeights.x * jointTransforms[skinIndex][JointIds.x]
					+ JointWeights.y * jointTransforms[skinIndex][JointIds.y]
					+ JointWeights.z * jointTransforms[skinIndex][JointIds.z]
					+ JointWeights.w * jointTransforms[skinIndex][JointIds.w];
	vec4 wp = skinMatrix * pos;
#else
	vec4 wp = model * vec4(Position, 1.0);
#endif
    fragPos = wp.xyz; 
	viewDir = cameraPosition.xyz - wp.xyz;
	gl_Position = projection * view * wp;
}
