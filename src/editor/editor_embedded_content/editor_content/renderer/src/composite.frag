#version 450

// shader outputs
layout (location = 0) out vec4 outColor;

layout(input_attachment_index = 0, binding = 0) uniform subpassInput texColor;
layout(input_attachment_index = 1, binding = 1) uniform subpassInput texWeights;

const float EPSILON = 1e-5;

bool isApproximatelyEqual(float a, float b) {
    return abs(a - b) <= (abs(a) < abs(b) ? abs(b) : abs(a)) * EPSILON;
}

float max3(vec3 v) {
    return max(max(v.x, v.y), v.z);
}

void main() {
	float reveal = subpassLoad(texWeights).r;
	if (isApproximatelyEqual(reveal, 1.0)) {
        discard;
	}
	vec4 accum = subpassLoad(texColor);
	if (isinf(max3(abs(accum.rgb)))) {
        accum.rgb = vec3(accum.a);
	}
	outColor = vec4(accum.rgb / max(accum.a, EPSILON), 1.0 - reveal);
}
