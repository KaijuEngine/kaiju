#version 460

#include "inc_vertex.inl"

// RGB + A (alpha is the mode 0 = 2D, anything else is 3D)
layout(location = LOCATION_START) in vec4 color;

layout(location = 0) out vec4 fragColor;
layout(location = 1) out vec2 fragTexCoords;
layout(location = 2) out vec3 fragWorldPos;

bool is2DMode() { return color.a - 0.00001 <= 0; }

void main() {
	fragTexCoords = UV0;
	vec4 wp = model * vec4(Position, 1.0);
	fragColor = vec4(color.rgb, 1.0);
	if (is2DMode()) {
		fragColor.a = 0.1;
		if (wp.x - 0.00001 < 0 && wp.x + 0.00001 > 0) {
			fragColor = vec4(0.3, 0.8, 0.3, 0.75);
		} else if (wp.y - 0.00001 < 0 && wp.y + 0.00001 > 0) {
			fragColor = vec4(0.8, 0.3, 0.3, 0.75);
		}
	} else {
		if (wp.x - 0.00001 < 0 && wp.x + 0.00001 > 0) {
			fragColor = vec4(0.3, 0.3, 0.8, 1.0);
		} else if (wp.z - 0.00001 < 0 && wp.z + 0.00001 > 0) {
			fragColor = vec4(0.8, 0.3, 0.3, 1.0);
		}
	}
	fragWorldPos = wp.xyz;
	gl_Position = projection * view * wp;
}
