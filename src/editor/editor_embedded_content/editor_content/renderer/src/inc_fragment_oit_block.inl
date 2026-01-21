#ifdef OIT
	reveal = unWeightedColor.a;
	if (reveal < 0.001) {
		discard;
	}
	float weight = clamp(pow(min(1.0, unWeightedColor.a * 10.0) + 0.01, 3.0) * 1e8 * pow(1.0 - gl_FragCoord.z * 0.9, 3.0), 1e-2, 3e3);
	outColor = vec4(unWeightedColor.rgb * unWeightedColor.a, unWeightedColor.a) * weight;
#else
	if (unWeightedColor.a < (1.0 - 0.001)) {
		discard;
	}
	outColor = unWeightedColor;
#endif