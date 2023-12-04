#version 330 core
layout (location = 0) in vec3 Position;
layout (location = 1) in vec3 Normal;
layout (location = 2) in vec4 Tangent;
layout (location = 3) in vec2 UV0;
layout (location = 4) in vec4 Color;
layout (location = 5) in vec4 JointIds;
layout (location = 6) in vec4 JointWeights;
layout (location = 7) in vec3 MorphTarget;

uniform struct GlobalData {
    mat4 view;
    mat4 projection;
    vec3 cameraPosition;
    float time;
} globalData;

out vec4 fragColor;

void main() {
    fragColor = Color;
    mat4 model = mat4(1.0);
    gl_Position = globalData.projection * globalData.view * model * vec4(Position, 1.0);
}
