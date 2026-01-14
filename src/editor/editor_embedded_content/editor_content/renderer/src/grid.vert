#version 460
#define VERTEX_SHADER

// RGB + A (alpha is the mode 0 = 2D, anything else is 3D)
#define LAYOUT_VERT_COLOR 0

#define LAYOUT_FRAG_COLOR 0
#define LAYOUT_FRAG_TEX_COORDS 1
#define LAYOUT_FRAG_WORLD_POS 2

#include "kaiju.glsl"

bool is2DMode() { return color.a - 0.00001 <= 0; }

void main() {
	fragTexCoords = UV0;
	vec4 wp = worldPosition();
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
	writeStandardPosition();
}
