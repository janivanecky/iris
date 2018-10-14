#version 330 core

out vec4 out_color;
in vec2 texcoord;

uniform sampler2D tex;
uniform vec4 color;

void main()
{
    out_color = texture(tex, texcoord) * color;
}