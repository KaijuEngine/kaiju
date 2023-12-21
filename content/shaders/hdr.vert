#version 300 es
precision mediump float;

layout(location = 0) in vec3 Position;
layout(location = 3) in vec2 UV0;

out vec2 TexCoords;

void main() {
    TexCoords = UV0;
	gl_Position = vec4(Position, 1.0);
}