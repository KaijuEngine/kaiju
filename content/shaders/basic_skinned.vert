#version 460

#include "inc_vertex.inl"

#define MAX_JOINTS			50
#define MAX_SKIN_INSTANCES	50

layout(set = 0, binding = 2) readonly uniform SkinnedUBO {
	mat4 jointTransforms[MAX_SKIN_INSTANCES][MAX_JOINTS];
};

layout(location = LOCATION_START) in vec4 color;
layout(location = LOCATION_START+1) in int skinIndex;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;

void main() {
	vec4 pos = vec4(Position, 1.0);
	mat4 skinMatrix = JointWeights.x * jointTransforms[skinIndex][JointIds.x]
					+ JointWeights.y * jointTransforms[skinIndex][JointIds.y]
					+ JointWeights.z * jointTransforms[skinIndex][JointIds.z]
					+ JointWeights.w * jointTransforms[skinIndex][JointIds.w];
	fragColor = Color * color;
	fragTexCoords = UV0;
	gl_Position = projection * view * model * skinMatrix * pos;
}
