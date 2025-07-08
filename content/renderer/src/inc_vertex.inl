#include "inc_default.inl"

layout (location = 0) in vec3 Position;
layout (location = 1) in vec3 Normal;
layout (location = 2) in vec4 Tangent;
layout (location = 3) in vec2 UV0;
layout (location = 4) in vec4 Color;
layout (location = 5) in ivec4 JointIds;
layout (location = 6) in vec4 JointWeights;
layout (location = 7) in vec3 MorphTarget;

#define LOCATION_HEAD   8
#define LOCATION_START  LOCATION_HEAD + 4

layout(location = LOCATION_HEAD) in mat4 model;