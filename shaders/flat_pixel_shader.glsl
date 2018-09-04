#version 330 core

in vec4 position;
in vec4 normal;

layout(location = 0) out vec4 out_color;

uniform vec4 color;

void main()
{
    out_color = color;
}