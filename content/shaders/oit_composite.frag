#version 450
//#version 300 es
//precision mediump float;

#ifdef VULKAN
	// shader outputs
	layout (location = 0) out vec4 outColor;

	layout(input_attachment_index = 0, binding = 0) uniform subpassInput texColor;
	layout(input_attachment_index = 1, binding = 1) uniform subpassInput texWeights;

	void main() {
		vec4 accum = subpassLoad(texColor);
		float reveal = subpassLoad(texWeights).r;
		outColor = vec4(accum.rgb / max(accum.a, 1e-5), reveal);
	}
#else
	out vec4 outColor;

	uniform sampler2D accum;
	uniform sampler2D reveal;

	const float EPSILON = 0.00001;

	bool isApproximatelyEqual(float a, float b) {
		return abs(a - b) <= (abs(a) < abs(b) ? abs(b) : abs(a)) * EPSILON;
	}

	float max3(vec3 v) {
		return max(max(v.x, v.y), v.z);
	}

	void main() {
		ivec2 coords = ivec2(gl_FragCoord.xy);
		float revealage = texelFetch(reveal, coords, 0).r;
		if (isApproximatelyEqual(revealage, 1.0)) 
			discard;
		vec4 accumulation = texelFetch(accum, coords, 0);
		if (isinf(max3(abs(accumulation.rgb)))) 
			accumulation.rgb = vec3(accumulation.a);
		vec3 averageColor = accumulation.rgb / max(accumulation.a, EPSILON);
		outColor = vec4(averageColor, 1.0 - revealage);
	}
#endif
