#version 460

#include "inc_vertex.inl"

layout(location = LOCATION_START) in vec4 color;

layout(location = 0) out vec4 fragColor;

void main() {
	fragColor = vec4(color.rgb, 1.0);
	float outlineWidth = color.a;
	float Sx = length(vec3(model[0])); // X-axis scaling
	float Sy = length(vec3(model[1])); // Y-axis scaling
	float Sz = length(vec3(model[2])); // Z-axis scaling
	vec3 scaledPosition;
	if (cameraPosition.w == 0) {	// 0 = perspective
		vec4 pWorld = model * vec4(Position, 1.0);
		vec4 pView = view * pWorld;
		float D = -pView.z;
		float perspScale = outlineWidth * D;
		scaledPosition = vec3(
			Position.x * (1.0 + perspScale / Sx),
			Position.y * (1.0 + perspScale / Sy),
			Position.z * (1.0 + perspScale / Sz)
		);
	} else {	// 1 = orthographic
        float viewWidth = 2.0 / projection[0][0];
        float viewHeight = 2.0 / -projection[1][1];
        float orthoScaleX = outlineWidth * viewWidth / Sx;
        float orthoScaleY = outlineWidth * viewHeight / Sy;
        scaledPosition = vec3(
            Position.x * (1.0 + orthoScaleX),
            Position.y * (1.0 + orthoScaleY),
            Position.z
        );
    }
	vec4 pWorldScaled = model * vec4(scaledPosition, 1.0);
	gl_Position = projection * view * pWorldScaled;
}
