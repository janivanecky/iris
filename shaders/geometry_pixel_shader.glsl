#version 330 core

in vec4 position;
in vec4 normal;

layout(location = 0) out vec4 out_position;
layout(location = 1) out vec4 out_normal;

void main()
{
    out_normal.xyz = normalize(normal.xyz);
    out_normal.a = 0.0;
    out_position = position;
}