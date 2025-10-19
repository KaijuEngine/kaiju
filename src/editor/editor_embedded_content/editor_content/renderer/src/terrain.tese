#version 460

#include "inc_default.inl"

layout (triangles, equal_spacing, cw) in;

layout(location = 0) in vec4 teseColor[];
layout(location = 1) in vec2 teseTexCoord[];
layout(location = 2) in vec3 teseCamPos[];
layout(location = 3) in float teseScalar[];
layout(location = 4) in mat4 teseView[];
layout(location = 8) in mat4 teseProjection[];
layout(location = 12) in mat4 teseModel[];
layout(location = 16) in mat3 teseNmlModel[];

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoord;
layout(location = 2) out vec3 fragTangentViewPos;
layout(location = 3) out vec3 fragTangentFragPos;
layout(location = 4) out vec3 fragLightTPos;
layout(location = 5) out vec4 fragPosLightSpace;
layout(location = 6) out vec3 fragPos;
layout(location = 7) out vec3 fragNormal;

// Height/Cavity/Roughness map
// Terrain normal map
// Rock color, normal
// Ground color, normal
layout(binding = 1) uniform sampler2D textures[6];

float height_map(vec2 position) {
	vec3 rgb = texture(textures[0], position).rgb;
	return rgb.r * teseScalar[0];
}

void main() {
    // barycentric coordinates
    float u = gl_TessCoord.x;
    float v = gl_TessCoord.y;
    float w = gl_TessCoord.z;

    fragTexCoord = u * teseTexCoord[0] + v * teseTexCoord[1] + w * teseTexCoord[2];

    vec4 p0 = u * gl_in[0].gl_Position;
    vec4 p1 = v * gl_in[1].gl_Position;
    vec4 p2 = w * gl_in[2].gl_Position;
    vec4 pos = p0 + p1 + p2;

    vec4 prePos = pos;

	prePos.y = height_map(fragTexCoord);
	//fragHeight = prePos.y - pos.y;

    // Generate 2 points near this point to create a triangle
    vec3 fp0 = prePos.xyz + vec3(0.1, 0.0, 0.0);
    vec3 fp1 = prePos.xyz + vec3(0.0, 0.0, 0.1);
	fp0.y = height_map(fragTexCoord + vec2(0.01, 0.0));
	fp1.y = height_map(fragTexCoord + vec2(0.0, 0.01));
    
    vec3 e0 = fp1 - fp0;
    vec3 e1 = prePos.xyz - fp1;
    vec3 triNml = normalize(cross(e1, e0));
    vec3 triTangent = normalize(cross(prePos.xyz, triNml));

    pos = teseModel[0] * prePos;

    vec3 T = normalize(teseNmlModel[0] * triTangent);
    vec3 N = normalize(teseNmlModel[0] * triNml);
    T = normalize(T - dot(T, N) * N);
    vec3 B = cross(N, T);
    mat3 TBN = transpose(mat3(T, B, N));
	
	fragPos = pos.xyz;
	fragColor = teseColor[0];
    fragLightTPos = TBN * vertLights[0].position;
	fragPosLightSpace = vertLights[0].matrix[0] * vec4(fragPos, 1.0);
	fragTangentViewPos = TBN * cameraPosition.xyz;
	fragTangentFragPos = TBN * fragPos;
	fragNormal = triNml;

    gl_Position = teseProjection[0] * teseView[0] * pos;
}
