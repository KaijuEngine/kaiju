float LinearizeDepth(float depth, float near, float far) {
	float z = depth * 2.0 - 1.0; // Back to NDC 
	return (2.0 * near * far) / (far + near - z * (far - near));
}

float DirectShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir) {
	// Perform perspective divide
	vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
	// Transform to [0,1] range
	projCoords.xy = projCoords.xy * 0.5 + 0.5;

	// Get closest depth value from light's perspective
	// (using [0,1] range fragPosLight as coords)
	float closestDepth = texture(shadowMap, projCoords.xy).r;

	// Get depth of current fragment from light's perspective
	float currentDepth = projCoords.z;

	float bias = 0.005;//max(0.05 * (1.0 - dot(normal, lightDir)), 0.005);
	float shadow = 0.0;

	// Make the shadow smoother
	vec2 texelSize = 1.0 / vec2(textureSize(shadowMap, 0));
	for (int x = -1; x <= 1; ++x) {
		for (int y = -1; y <= 1; ++y) {
			float pcfDepth = texture(shadowMap, projCoords.xy + vec2(x, y) * texelSize).r;
			shadow += (currentDepth - bias) > pcfDepth ? 1.0 : 0.0;
		}
	}
	shadow /= 9.0;
	
	if (projCoords.z > 1.0)
		shadow = 0.0;
	return shadow;
}

float SpotShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir, float near, float far)
{
	// Perform perspective divide
	vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
	// Transform to [0,1] range
	projCoords.xy = projCoords.xy * 0.5 + 0.5;

	// Get closest depth value from light's perspective
	// (using [0,1] range fragPosLight as coords)
	float closestDepth = texture(shadowMap, projCoords.xy).r;

	// Get depth of current fragment from light's perspective
	float currentDepth = projCoords.z;

	closestDepth = LinearizeDepth(closestDepth, near, far) / far;
	currentDepth = LinearizeDepth(currentDepth, near, far) / far;

	float bias = 0.005;//max(0.05 * (1.0 - dot(normal, lightDir)), 0.005);
	float shadow = 0.0;

	// Make the shadow smoother
	vec2 texelSize = 1.0 / vec2(textureSize(shadowMap, 0));
	for (int x = -1; x <= 1; ++x) {
		for (int y = -1; y <= 1; ++y) {
			float pcfDepth = texture(shadowMap, projCoords.xy + vec2(x, y) * texelSize).r;
			pcfDepth = LinearizeDepth(pcfDepth, near, far) / far;
			shadow += (currentDepth - bias) > pcfDepth ? 1.0 : 0.0;
		}
	}
	shadow /= 9.0;
	
	if (projCoords.z > 1.0)
		shadow = 0.0;

	return shadow;
}

// array of offset direction for sampling
const vec3 pointSamplingDiskGrid[20] = vec3[]
(
	vec3(1, 1,  1), vec3( 1, -1,  1), vec3(-1, -1,  1), vec3(-1, 1,  1),
	vec3(1, 1, -1), vec3( 1, -1, -1), vec3(-1, -1, -1), vec3(-1, 1, -1),
	vec3(1, 1,  0), vec3( 1, -1,  0), vec3(-1, -1,  0), vec3(-1, 1,  0),
	vec3(1, 0,  1), vec3(-1,  0,  1), vec3( 1,  0, -1), vec3(-1, 0, -1),
	vec3(0, 1,  1), vec3( 0, -1,  1), vec3( 0, -1, -1), vec3( 0, 1, -1)
);
float PointShadowCalculation(vec3 fragPos, vec3 lightPos, vec3 viewPos, float far) {
	vec3 delta = fragPos - lightPos;
	float currentDepth = length(delta);
	float shadow = 0.0;
	float bias = 0.15;
	int samples = 20;
	float viewDistance = length(viewPos - fragPos);
	float diskRadius = (1.0 + (viewDistance / far)) / 25.0;
	for (int i = 0; i < samples; ++i) {
		float closestDepth = texture(shadowCubeMap, delta + pointSamplingDiskGrid[i] * diskRadius).r;
		closestDepth *= far;   // undo mapping [0;1]
		if ((currentDepth - bias) > closestDepth)
			shadow += 1.0;
	}
	shadow /= float(samples);
	return shadow;
}
