#version 420 core

layout (binding = 0) uniform sampler2D diffuse_tex;
layout (binding = 1) uniform sampler2D ambient_tex;
layout (binding = 2) uniform sampler2D occlusion_tex;

uniform float minWhite;
in vec2 texcoord;
out vec4 out_color;

void main()
{
    vec4 diffuse = texture(diffuse_tex, texcoord);
    vec4 ambient = texture(ambient_tex, texcoord);
    float occlusion = texture(occlusion_tex, texcoord).x;
    vec4 col = diffuse + vec4(ambient.xyz * occlusion, 1.0);
    float luma = 0.2126 * col.r + 0.7152 * col.g + 0.0722 * col.b;
	float mappedLuma = (luma * (1 + luma / minWhite)) / (1.0f + luma);
	col.xyz *= mappedLuma / luma;
    out_color = col;
}   