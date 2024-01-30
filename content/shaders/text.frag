#version 450
//#version 300 es
//precision mediump float;

#ifdef VULKAN
	layout(location = 0) in vec4 fragColor;
	layout(location = 1) in vec4 fragBGColor;
	layout(location = 2) in vec2 fragTexCoord;
	layout(location = 3) in vec2 fragPxRange;
	layout(location = 4) in vec2 fragTexRange;

	layout(binding = 1) uniform sampler2D texSampler;
#else
	in vec4 fragColor;
	in vec4 fragBGColor;
	in vec2 fragTexCoord;
	in vec2 fragPxRange;
	in vec2 fragTexRange;

	uniform sampler2D texSampler;
#endif

layout(location = 0) out vec4 outColor;
layout(location = 1) out float reveal;

float median(float r, float g, float b) {
	return max(min(r, g), min(max(r, g), b));
}

void main() {
	ivec2 sz = ivec2(fragTexRange * vec2(textureSize(texSampler, 0)));
	//ivec2 sz = textureSize(texSampler, 0).xy;
	vec3 msdfColor = texture(texSampler, fragTexCoord).rgb;

	//float dx = dFdx(fragTexCoord.x) * sz.x;
	//float dy = dFdy(fragTexCoord.y) * sz.y;
	//float sigDist = median(msdfColor.r, msdfColor.g, msdfColor.b) - 0.5;
	//float w = fwidth(sigDist);
	//float opacity = smoothstep(0.5 - w, 0.5 + w, sigDist);

	//vec2 msdfUnit = fragPxRange / sz;
	//float sigDist = median(msdfColor.r, msdfColor.g, msdfColor.b) - 0.5;
	//sigDist *= dot(msdfUnit, 0.5 / fwidth(fragTexCoord));
	//float opacity = clamp(sigDist + 0.5, 0.0, 1.0);

	vec2 dxdy = fwidth(fragTexCoord) * textureSize(texSampler, 0);
	float dist = median(msdfColor.r, msdfColor.g, msdfColor.b) - 0.5;
	float opacity = clamp(dist * 8.0 / length(dxdy) + 0.5, 0.0, 1.0);

	vec4 unWeightedColor = mix(fragBGColor, fragColor, opacity);
#ifdef OIT
	float distWeight = clamp(0.03 / (1e-5 + pow(gl_FragCoord.z / 200.0, 4.0)), 1e-2, 3e3);
	float alphaWeight = min(1.0, max(max(unWeightedColor.r, unWeightedColor.g),
	max(unWeightedColor.b, unWeightedColor.a)) * 40.0 + 0.01);
	alphaWeight *= alphaWeight;
	float weight = alphaWeight * distWeight;
	outColor = vec4(unWeightedColor.rgb * unWeightedColor.a, unWeightedColor.a) * weight;
	reveal = unWeightedColor.a;
#else
	if (unWeightedColor.a < (1.0 - 0.0001))
		discard;
	outColor = unWeightedColor;
#endif
}