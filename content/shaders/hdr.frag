#version 300 es
precision mediump float;

layout(location = 0) out vec4 outColor;

in vec2 TexCoords;

uniform sampler2D hdrBuffer;
uniform bool hdr;
uniform float exposure;

void main() {
	const float gamma = 2.2;
	vec2 tc = vec2(TexCoords.x, 1.0 - TexCoords.y);
	vec3 hdrColor = texture(hdrBuffer, tc).rgb;
	if(hdr) {
		// reinhard
		// vec3 result = hdrColor / (hdrColor + vec3(1.0));
		// exposure
		vec3 result = vec3(1.0) - exp(-hdrColor * exposure);
		// also gamma correct while we're at it       
		result = pow(result, vec3(1.0 / gamma));
		outColor = vec4(result, 1.0);
	} else {
		//vec3 result = pow(hdrColor, vec3(1.0 / gamma));
		//outColor = vec4(result, 1.0);
		outColor = vec4(hdrColor, 1.0);
	}
}
