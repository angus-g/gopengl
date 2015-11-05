#version 150

in vec3 position;
in vec3 color;
in vec2 texCoord;

out vec3 vertColor;
out vec2 vertTexCoord;

uniform mat4 model;
uniform mat4 view;
uniform mat4 proj;

void main() {
    gl_Position = proj * view * model * vec4(position, 1.0);
    vertColor = color;
    vertTexCoord = texCoord;
}
