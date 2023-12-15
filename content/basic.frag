#version 330 core
out vec4 FragColor;

in vec4 fragColor;
in vec2 fragTexCoords;

uniform sampler2D texSampler;

void main() {
    FragColor = texture(texSampler, fragTexCoords) * fragColor;
}
