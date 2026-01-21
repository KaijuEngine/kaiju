#define LOCATION_HEAD  8
#define LOCATION_START LOCATION_HEAD + 4
#define PI             3.14159265359
#define CUBEMAP_SIDES  6

#ifndef MAX_JOINTS
	#define MAX_JOINTS 50
#endif

#ifndef NR_LIGHTS
	#define NR_LIGHTS 4
#endif

#ifndef MAX_LIGHTS
	#define MAX_LIGHTS 20
#endif

#ifdef LAYOUT_ALL_LIGHT_REQUIREMENTS
	#ifndef LAYOUT_FRAG_TANGENT_VIEW_POS
		#define LAYOUT_FRAG_TANGENT_VIEW_POS LAYOUT_ALL_LIGHT_REQUIREMENTS
	#endif
	#ifndef LAYOUT_FRAG_TANGENT_FRAG_POS
		#define LAYOUT_FRAG_TANGENT_FRAG_POS LAYOUT_ALL_LIGHT_REQUIREMENTS+1
	#endif
	#ifndef LAYOUT_FRAG_LIGHT_T_POS
		#define LAYOUT_FRAG_LIGHT_T_POS LAYOUT_ALL_LIGHT_REQUIREMENTS+2
	#endif
	#ifndef LAYOUT_FRAG_LIGHT_T_DIR
		#define LAYOUT_FRAG_LIGHT_T_DIR LAYOUT_ALL_LIGHT_REQUIREMENTS+6
	#endif
	#ifndef LAYOUT_FRAG_POS_LIGHT_SPACE
		#define LAYOUT_FRAG_POS_LIGHT_SPACE LAYOUT_ALL_LIGHT_REQUIREMENTS+10
	#endif
	#ifndef LAYOUT_FRAG_LIGHT_COUNT
		#define LAYOUT_FRAG_LIGHT_COUNT LAYOUT_ALL_LIGHT_REQUIREMENTS+14
	#endif
	#ifndef LAYOUT_FRAG_LIGHT_INDEXES
		#define LAYOUT_FRAG_LIGHT_INDEXES LAYOUT_ALL_LIGHT_REQUIREMENTS+15
	#endif
#endif

#ifdef FRAGMENT_SHADER
	#define FRAG_INOUT in
#elif defined(VERTEX_SHADER)
	#define FRAG_INOUT out
#else
	#error "FRAG_INOUT can only be used in vertex or fragment shaders"
#endif

#ifdef SAMPLER_COUNT
	layout(binding = 1) uniform sampler2D textures[SAMPLER_COUNT];
#endif

#ifdef SHADOW_SAMPLERS
	layout(binding = 2) uniform sampler2D shadowMap[MAX_LIGHTS];
	layout(binding = 3) uniform samplerCube shadowCubeMap[MAX_LIGHTS];
#endif

struct Light {
	mat4 matrix[CUBEMAP_SIDES];
	vec3 position;
	vec3 direction;
};

struct PointShadow {
	vec2 point; // X,Z
	float radius;
	float strength;
};

struct LightInfo {
	vec3 position;
	float intensity;
	vec3 direction;
	float cutoff;
	vec3 ambient;
	float outerCutoff;
	vec3 diffuse;
	float constant;
	vec3 specular;
	float linear;
	float quadratic;
	float nearPlane;
	float farPlane;
	int type;
};

// cameraPosition.w = [0=perspective, 1=orthographic]
layout(set = 0, binding = 0) readonly uniform UniformBufferObject {
	mat4 view;
	mat4 projection;
	mat4 uiView;
	mat4 uiProjection;
	vec4 cameraPosition;
	vec3 uiCameraPosition;
	float time;
	vec2 screenSize;
	int cascadeCount;
	float cascadePlaneDistances[5];
	Light vertLights[MAX_LIGHTS];
	LightInfo lightInfos[MAX_LIGHTS];
};

////////////////////////////////////////////////////////////////////////////////
// Vertex shader layouts
////////////////////////////////////////////////////////////////////////////////

#ifdef VERTEX_SHADER
	#ifdef SKINNING
		layout(set = 0, binding = 2) readonly buffer SkinnedSSBO {
			mat4 jointTransforms[][MAX_JOINTS];
		};
	#endif

	layout (location = 0) in vec3 Position;
	layout (location = 1) in vec3 Normal;
	layout (location = 2) in vec4 Tangent;
	layout (location = 3) in vec2 UV0;
	layout (location = 4) in vec4 Color;
	layout (location = 5) in ivec4 JointIds;
	layout (location = 6) in vec4 JointWeights;
	layout (location = 7) in vec3 MorphTarget;

	layout(location = LOCATION_HEAD) in mat4 model;
#endif

#ifdef LAYOUT_VERT_COLOR
	layout(location = LOCATION_START+LAYOUT_VERT_COLOR) in vec4 color;
#endif
#ifdef LAYOUT_VERT_METALLIC
	layout(location = LOCATION_START+LAYOUT_VERT_METALLIC) in float metallic;
#endif
#ifdef LAYOUT_VERT_ROUGHNESS
	layout(location = LOCATION_START+LAYOUT_VERT_ROUGHNESS) in float roughness;
#endif
#ifdef LAYOUT_VERT_EMISSIVE
	layout(location = LOCATION_START+LAYOUT_VERT_EMISSIVE) in float emissive;
#endif
#ifdef LAYOUT_VERT_LIGHT_IDS
	layout(location = LOCATION_START+LAYOUT_VERT_LIGHT_IDS) in int lightIds[NR_LIGHTS];
#endif
#ifdef LAYOUT_VERT_UVS
	layout(location = LOCATION_START+LAYOUT_VERT_UVS) in vec4 uvs;
#endif
#ifdef LAYOUT_VERT_FLAGS
	layout(location = LOCATION_START+LAYOUT_VERT_FLAGS) in uint flags;
#endif
#ifdef LAYOUT_VERT_FRUSTUM_PROJECTION
	layout(location = LOCATION_START+LAYOUT_VERT_FRUSTUM_PROJECTION) in mat4 frustumProjection;
#endif

#ifdef LAYOUT_FRAG_COLOR
	layout(location = LAYOUT_FRAG_COLOR) FRAG_INOUT vec4 fragColor;
#endif
#ifdef LAYOUT_FRAG_POS
	layout(location = LAYOUT_FRAG_POS) FRAG_INOUT vec3 fragPos;
#endif
#ifdef LAYOUT_FRAG_TEX_COORDS
	layout(location = LAYOUT_FRAG_TEX_COORDS) FRAG_INOUT vec2 fragTexCoords;
#endif
#ifdef LAYOUT_FRAG_VIEW_DIR
	layout(location = LAYOUT_FRAG_VIEW_DIR) FRAG_INOUT vec3 fragViewDir;
#endif
#ifdef LAYOUT_FRAG_NORMAL
	layout(location = LAYOUT_FRAG_NORMAL) FRAG_INOUT vec3 fragNormal;
#endif
#ifdef LAYOUT_FRAG_TANGENT_VIEW_POS
	layout(location = LAYOUT_FRAG_TANGENT_VIEW_POS) FRAG_INOUT vec3 fragTangentViewPos;
#endif
#ifdef LAYOUT_FRAG_TANGENT_FRAG_POS
	layout(location = LAYOUT_FRAG_TANGENT_FRAG_POS) FRAG_INOUT vec3 fragTangentFragPos;
#endif
#ifdef LAYOUT_FRAG_LIGHT_COUNT
	layout(location = LAYOUT_FRAG_LIGHT_COUNT) FRAG_INOUT flat int fragLightCount;
#endif
#ifdef LAYOUT_FRAG_LIGHT_T_POS
	layout(location = LAYOUT_FRAG_LIGHT_T_POS) FRAG_INOUT vec3 fragLightTPos[NR_LIGHTS];
#endif
#ifdef LAYOUT_FRAG_LIGHT_T_DIR
	layout(location = LAYOUT_FRAG_LIGHT_T_DIR) FRAG_INOUT vec3 fragLightTDir[NR_LIGHTS];
#endif
#ifdef LAYOUT_FRAG_POS_LIGHT_SPACE
	layout(location = LAYOUT_FRAG_POS_LIGHT_SPACE) FRAG_INOUT vec4 fragPosLightSpace[NR_LIGHTS];
#endif
#ifdef LAYOUT_FRAG_LIGHT_INDEXES
	layout(location = LAYOUT_FRAG_LIGHT_INDEXES) FRAG_INOUT flat int fragLightIndexes[NR_LIGHTS];
#endif
#ifdef LAYOUT_FRAG_METALLIC
	layout(location = LAYOUT_FRAG_METALLIC) FRAG_INOUT float fragMetallic;
#endif
#ifdef LAYOUT_FRAG_ROUGHNESS
	layout(location = LAYOUT_FRAG_ROUGHNESS) FRAG_INOUT float fragRoughness;
#endif
#ifdef LAYOUT_FRAG_EMISSIVE
	layout(location = LAYOUT_FRAG_EMISSIVE) FRAG_INOUT float fragEmissive;
#endif
#ifdef LAYOUT_FRAG_FLAGS
	layout(location = LAYOUT_FRAG_FLAGS) FRAG_INOUT flat uint fragFlags;
#endif
#ifdef LAYOUT_FRAG_WORLD_POS
	layout(location = LAYOUT_FRAG_WORLD_POS) FRAG_INOUT vec3 fragWorldPos;
#endif

#ifdef FRAGMENT_SHADER
	layout(location = 0) out vec4 outColor;
	#ifdef OIT
		layout(location = 1) out float reveal;
	#elif defined(HAS_GBUFFER)
		layout(location = 1) out vec4 outPosition;
		layout(location = 2) out vec4 outNormal;
	#endif
#endif

////////////////////////////////////////////////////////////////////////////////
// Vertex shader helper functions
////////////////////////////////////////////////////////////////////////////////

#ifdef VERTEX_SHADER
	vec3 transformPoint(mat4 mat, vec3 point) {
		vec4 pt0 = vec4(point.xyz, 1.0);
		vec4 res = mat * pt0;
		vec3 v3 = res.xyz;
		return v3 / -res.w;
	}

	mat4 skinMatrix() {
	#ifdef SKINNING
		return JointWeights.x * jointTransforms[gl_InstanceIndex][JointIds.x]
			+ JointWeights.y * jointTransforms[gl_InstanceIndex][JointIds.y]
			+ JointWeights.z * jointTransforms[gl_InstanceIndex][JointIds.z]
			+ JointWeights.w * jointTransforms[gl_InstanceIndex][JointIds.w];
	#else
		return mat4(1.0);
	#endif
	}

	vec4 skinWorldPosition() {
		return skinMatrix() * vec4(Position, 1.0);
	}

	vec4 worldPosition() {
	#ifdef SKINNING
		return skinWorldPosition();
	#else
		return model * vec4(Position, 1.0);
	#endif
	}

	void calcVertexLightInformation() {
	#if defined(LAYOUT_FRAG_LIGHT_COUNT) && defined(LAYOUT_FRAG_LIGHT_T_POS) && defined(LAYOUT_FRAG_LIGHT_T_DIR) && defined(LAYOUT_FRAG_POS_LIGHT_SPACE) && defined(LAYOUT_FRAG_LIGHT_INDEXES) && defined(LAYOUT_FRAG_LIGHT_COUNT) && defined(LAYOUT_FRAG_TANGENT_VIEW_POS) && defined(LAYOUT_FRAG_TANGENT_FRAG_POS) && defined(LAYOUT_FRAG_NORMAL)
		mat3 nmlMat = transpose(inverse(mat3(model)));
		vec3 T = normalize(nmlMat * Tangent.xyz);
		vec3 N = normalize(nmlMat * Normal);
		// re-orthogonalize T with respect to N
		T = normalize(T - dot(T, N) * N);
		// then retrieve perpendicular vector B with the cross product of T and N
		vec3 B = cross(N, T);
		mat3 TBN = transpose(mat3(T, B, N));
		fragLightCount = 0;
		for (int i = 0; i < NR_LIGHTS; ++i) {
			int idx = lightIds[i];
			if (idx < 0) {
				continue;
			}
			idx = min(idx, MAX_LIGHTS - 1);
			fragLightTPos[fragLightCount] = TBN * vertLights[idx].position;
			fragLightTDir[fragLightCount] = TBN * normalize(vertLights[idx].direction);
			fragPosLightSpace[fragLightCount] = vertLights[idx].matrix[0] * vec4(fragPos, 1.0);
			fragLightIndexes[fragLightCount] = idx;
			fragLightCount++;
		}
		fragTangentViewPos = TBN * cameraPosition.xyz;
		fragTangentFragPos = TBN * fragPos;
		fragNormal = N;
	#endif
	}

	void writeTexCoords() {
		#ifdef LAYOUT_FRAG_TEX_COORDS
			#ifdef LAYOUT_VERT_UVS
				vec2 uv = UV0;
				uv *= uvs.zw;
				uv.y += (1.0 - uvs.w) - uvs.y;
				uv.x += uvs.x;
				fragTexCoords = uv;
			#else
				fragTexCoords = UV0;
			#endif
		#endif
	}

	void writeStandardUIPosition() {
		vec4 wp = worldPosition();
		gl_Position = uiProjection * uiView * wp;
	}

	void writeStandardPosition() {
	#ifdef BILLBOARD
		vec4 centerWorld = model * vec4(0.0, 0.0, 0.0, 1.0);
		vec3 camRight = vec3(view[0].x, view[1].x, view[2].x);
		vec3 camUp    = vec3(view[0].y, view[1].y, view[2].y);
		vec3 offset = camRight * Position.x + camUp * Position.y;
		vec4 worldPos = centerWorld + vec4(offset, 0.0);
		vec4 camSpace = view * worldPos;
		#ifdef LAYOUT_FRAG_WORLD_POS
			fragWorldPos = camSpace.xyz;
		#endif
		#ifdef LAYOUT_FRAG_POS
			fragPos = camSpace.xyz;
		#endif
		gl_Position = projection * camSpace;
	#else
		vec4 wp = worldPosition();
		#ifdef LAYOUT_FRAG_WORLD_POS
			fragWorldPos = wp.xyz;
		#endif
		#ifdef LAYOUT_FRAG_POS
			fragPos = wp.xyz;
		#endif
		gl_Position = projection * view * wp;
	#endif
	}
#endif

////////////////////////////////////////////////////////////////////////////////
// Fragment shader helper functions
////////////////////////////////////////////////////////////////////////////////
#ifdef FRAGMENT_SHADER
	const vec2 poissonDisk[16] = vec2[](
		vec2(-0.94201624, -0.39906216),
		vec2(0.94558609, -0.76890725),
		vec2(-0.094184101, -0.92938870),
		vec2(0.34495938, 0.29387760),
		vec2(-0.91588581, 0.45771432),
		vec2(-0.81544232, -0.87912464),
		vec2(-0.38277543, 0.27676845),
		vec2(0.97484398, 0.75648379),
		vec2(0.44323325, -0.97511554),
		vec2(0.53742981, -0.47373420),
		vec2(-0.26496911, -0.41893023),
		vec2(0.79197514, 0.19090188),
		vec2(-0.24188840, 0.99706507),
		vec2(-0.81409955, 0.91437590),
		vec2(0.19984126, 0.78641367),
		vec2(0.14383161, -0.14100790)
	);

	// array of offset direction for sampling
	const vec3 pointSamplingDiskGrid[20] = vec3[] (
		vec3(1, 1,  1), vec3( 1, -1,  1), vec3(-1, -1,  1), vec3(-1, 1,  1),
		vec3(1, 1, -1), vec3( 1, -1, -1), vec3(-1, -1, -1), vec3(-1, 1, -1),
		vec3(1, 1,  0), vec3( 1, -1,  0), vec3(-1, -1,  0), vec3(-1, 1,  0),
		vec3(1, 0,  1), vec3(-1,  0,  1), vec3( 1,  0, -1), vec3(-1, 0, -1),
		vec3(0, 1,  1), vec3( 0, -1,  1), vec3( 0, -1, -1), vec3( 0, 1, -1)
	);

	#ifdef HAS_GBUFFER
		void processGBuffer(vec3 nml) {
		#ifndef OIT
			outPosition = vec4(fragPos, uintBitsToFloat(fragFlags));
			outNormal = vec4(nml, 0.0);
		#endif
		}
	#endif

	void processFinalColor(vec4 inColor) {
	#ifdef OIT
		reveal = inColor.a;
		if (reveal < 0.001) {
			discard;
		}
		float weight = clamp(pow(min(1.0, inColor.a * 10.0) + 0.01, 3.0) * 1e8 * pow(1.0 - gl_FragCoord.z * 0.9, 3.0), 1e-2, 3e3);
		outColor = vec4(inColor.rgb * inColor.a, inColor.a) * weight;
	#else
		if (inColor.a < (1.0 - 0.001)) {
			discard;
		}
		outColor = inColor;
	#endif
	}

	float LinearizeDepth(float depth, float near, float far) {
		float z = depth * 2.0 - 1.0; // Back to NDC 
		return (2.0 * near * far) / (far + near - z * (far - near));
	}

	vec3 fresnelSchlick(float cosTheta, vec3 F0) {
		return F0 + (1.0 - F0) * pow(max(1.0 - cosTheta, 0.0), 5.0);
	}

	float distributionGGX(vec3 N, vec3 H, float fragRoughness) {
		float a      = fragRoughness*fragRoughness;
		float a2     = a*a;
		float NdotH  = max(dot(N, H), 0.0);
		float NdotH2 = NdotH*NdotH;
		float num   = a2;
		float denom = (NdotH2 * (a2 - 1.0) + 1.0);
		denom = PI * denom * denom;
		return num / denom;
	}

	float geometrySchlickGGX(float NdotV, float fragRoughness) {
		float r = (fragRoughness + 1.0);
		float k = (r*r) / 8.0;
		float num   = NdotV;
		float denom = NdotV * (1.0 - k) + k;
		return num / denom;
	}

	float geometrySmith(vec3 N, vec3 V, vec3 L, float fragRoughness) {
		float NdotV = max(dot(N, V), 0.0);
		float NdotL = max(dot(N, L), 0.0);
		float ggx2  = geometrySchlickGGX(NdotV, fragRoughness);
		float ggx1  = geometrySchlickGGX(NdotL, fragRoughness);
		return ggx1 * ggx2;
	}

	#ifdef SHADOW_SAMPLERS
		float directShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir, int lightIdx) {
			/*
			vec4 fragPosViewSpace = view * vec4(fragPos, 1.0);
			float depthValue = abs(fragPosViewSpace.z);
				
			int layer = -1;
			for (int i = 0; i < cascadeCount; ++i)
			{
				if (depthValue < cascadePlaneDistances[i])
				{
					layer = i;
					break;
				}
			}
			if (layer == -1)
			{
				layer = cascadeCount;
			}
				
			vec4 fragPosLightSpace = lightSpaceMatrices[layer] * vec4(fragPosWorldSpace, 1.0);
			*/
			// Perform perspective divide
			vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
			// Transform to [0,1] range
			projCoords.xy = projCoords.xy * 0.5 + 0.5;
			// Get closest depth value from light's perspective
			// (using [0,1] range fragPosLight as coords)
			float closestDepth = texture(shadowMap[lightIdx], projCoords.xy).r;
			// Get depth of current fragment from light's perspective
			float currentDepth = projCoords.z;

			float bias = max(0.001 * (1.0 - dot(normal, lightDir)), 0.001);
			float slopeScale = max(0.005 * (1.0 - dot(normal, lightDir)), 0.002);
			float dzdx = dFdx(projCoords.z);
			float dzdy = dFdy(projCoords.z);
			float depthSlope = max(abs(dzdx), abs(dzdy));
			bias += slopeScale * depthSlope;
			bias = clamp(bias, 0.0001, 0.005);

			float shadow = 0.0;
			int samples = 16;
			vec2 texelSize = 1.0 / vec2(textureSize(shadowMap[lightIdx], 0));
			for(int i = 0; i < samples; ++i) {
				vec2 offset = poissonDisk[i] * texelSize * 1.5;  // Tune radius (1.0-2.0) for penumbra
				float pcfDepth = texture(shadowMap[lightIdx], projCoords.xy + offset).r;
				shadow += (currentDepth - bias) > pcfDepth ? 1.0 : 0.0;
			}
			shadow /= float(samples);
			if (projCoords.z > 1.0) {
				shadow = 0.0;
			}
			return shadow;
		}

		float spotShadowCalculation(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir, float near, float far, int lightIdx) {
			// Perform perspective divide
			vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
			// Transform to [0,1] range
			projCoords.xy = projCoords.xy * 0.5 + 0.5;

			// Get closest depth value from light's perspective
			// (using [0,1] range fragPosLight as coords)
			float closestDepth = texture(shadowMap[lightIdx], projCoords.xy).r;

			// Get depth of current fragment from light's perspective
			float currentDepth = projCoords.z;

			closestDepth = LinearizeDepth(closestDepth, near, far) / far;
			currentDepth = LinearizeDepth(currentDepth, near, far) / far;

			float bias = max(0.001 * (1.0 - dot(normal, lightDir)), 0.001);
			float slopeScale = max(0.005 * (1.0 - dot(normal, lightDir)), 0.002);
			float dzdx = dFdx(projCoords.z);
			float dzdy = dFdy(projCoords.z);
			float depthSlope = max(abs(dzdx), abs(dzdy));
			bias += slopeScale * depthSlope;
			bias = clamp(bias, 0.0001, 0.005);

			float shadow = 0.0;
			int samples = 16;
			vec2 texelSize = 1.0 / vec2(textureSize(shadowMap[lightIdx], 0));
			for(int i = 0; i < samples; ++i) {
				vec2 offset = poissonDisk[i] * texelSize * 1.5;  // Tune radius (1.0-2.0) for penumbra
				float pcfDepth = texture(shadowMap[lightIdx], projCoords.xy + offset).r;
				shadow += (currentDepth - bias) > pcfDepth ? 1.0 : 0.0;
			}
			shadow /= float(samples);
			
			if (projCoords.z > 1.0) {
				shadow = 0.0;
			}
			return shadow;
		}

		float pointShadowCalculation(vec3 fragPos, vec3 lightPos, float far, int lightIdx, vec3 normal) {
			vec3 delta = fragPos - lightPos;
			float currentDepth = length(delta);
			float shadow = 0.0;
			//float bias = 0.15;
			float bias = 0.15 + (1.0 - dot(normalize(delta), normal)) * 0.1;
			int samples = 20;
			float diskRadius = (currentDepth / far) / 25.0;
			for (int i = 0; i < samples; ++i) {
				float closestDepth = texture(shadowCubeMap[lightIdx], delta + pointSamplingDiskGrid[i] * diskRadius).r;
				closestDepth *= far;   // undo mapping [0;1]
				if ((currentDepth - bias) > closestDepth)
					shadow += 1.0;
			}
			shadow /= float(samples);
			return shadow;
		}
	#endif
#endif
