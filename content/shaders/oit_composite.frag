#version 450

// shader outputs
layout (location = 0) out vec4 outColor;

layout(input_attachment_index = 0, binding = 0) uniform subpassInput texColor;
layout(input_attachment_index = 1, binding = 1) uniform subpassInput texWeights;

void main() {
	vec4 accum = subpassLoad(texColor);
	float reveal = subpassLoad(texWeights).r;
	outColor = vec4(accum.rgb / max(accum.a, 1e-5), reveal);
}
