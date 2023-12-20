#version 330 core

layout (location = 0) in vec3 Position;
layout (location = 1) in vec3 Normal;
layout (location = 2) in vec4 Tangent;
layout (location = 3) in vec2 UV0;
layout (location = 4) in vec4 Color;
layout (location = 5) in vec4 JointIds;
layout (location = 6) in vec4 JointWeights;
layout (location = 7) in vec3 MorphTarget;

#define INSTANCE_VEC4_COUNT 5
uniform sampler2D instanceSampler;

uniform struct GlobalData {
    mat4 view;
    mat4 projection;
    mat4 uiView;
    mat4 uiProjection;
    vec3 cameraPosition;
    vec3 uiCameraPosition;
    float time;
} globalData;

out vec4 fragColor;
out vec2 fragTexCoords;

mat4 pullModel(int xOffset) {
    mat4 model;
    model[0] = texelFetch(instanceSampler, ivec2(xOffset,0), 0);
    model[1] = texelFetch(instanceSampler, ivec2(xOffset+1,0), 0);
    model[2] = texelFetch(instanceSampler, ivec2(xOffset+2,0), 0);
    model[3] = texelFetch(instanceSampler, ivec2(xOffset+3,0), 0);
    return model;
}

void main() {
    int xOffset = gl_InstanceID*INSTANCE_VEC4_COUNT;
    mat4 model = pullModel(xOffset);
    vec4 color = texelFetch(instanceSampler, ivec2(xOffset+4,0), 0);
    fragColor = Color * color;
    fragTexCoords = UV0;
    gl_Position = globalData.projection * globalData.view * model * vec4(Position, 1.0);
}
