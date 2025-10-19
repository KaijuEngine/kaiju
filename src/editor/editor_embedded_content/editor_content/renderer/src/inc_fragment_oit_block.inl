#ifdef OIT
	float distWeight = clamp(0.03 / (1e-5 + pow(gl_FragCoord.z / 200.0, 4.0)), 1e-2, 3e3);
	float alphaWeight = max(min(1.0, max(max(unWeightedColor.r, unWeightedColor.g), unWeightedColor.b) * unWeightedColor.a), unWeightedColor.a);
	float weight = alphaWeight * distWeight;
	outColor = vec4(unWeightedColor.rgb * unWeightedColor.a, unWeightedColor.a) * weight;
	reveal = unWeightedColor.a;
#else
	if (unWeightedColor.a < (1.0 - 0.0001)) {
		discard;
	}
	outColor = unWeightedColor;
#endif