#version 460

layout(location = 0) flat in uint fragPickID;
layout(location = 0) out uint outPickID;

void main() {
	outPickID = fragPickID;
}
