#version 330 core

out vec4 out_color;
in vec2 texcoord;

uniform sampler2D font_texture;
uniform vec4 color;

void main()
{
    out_color = color;
    out_color.a *= texture(font_texture, texcoord).r;
}