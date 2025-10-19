#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in float heightScalar;

layout(location = 0) out vec4 tescColor;
layout(location = 1) out vec2 tescTexCoord;
layout(location = 2) out vec3 tescCamPos;
layout(location = 3) out float tescScalar;
layout(location = 4) out mat4 tescView;
layout(location = 8) out mat4 tescProjection;
layout(location = 12) out mat4 tescModel;
layout(location = 16) out mat3 tescNmlModel;

void main() {
	tescColor = Color;
	tescTexCoord = UV0;
	tescModel = model;
	tescView = view;
	tescProjection = projection;
	tescCamPos = cameraPosition.xyz;
	tescScalar = heightScalar;
	tescNmlModel = transpose(inverse(mat3(model)));
	gl_Position = vec4(Position, 1.0);
}
