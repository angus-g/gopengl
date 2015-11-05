#version 150

in vec3 vertColor;
in vec2 vertTexCoord;

out vec4 outColor;

uniform sampler2D tex;

void main() {
    vec4 texColor = texture(tex, vertTexCoord);
    outColor = texColor * vec4(vertColor, 1.0);
}
